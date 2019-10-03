package cli

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"time"

	"github.com/decentraland/auth-go/pkg/ephemeral"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog"

	"github.com/decentraland/webrtc-broker/pkg/authentication"
	broker "github.com/decentraland/webrtc-broker/pkg/protocol"
	"github.com/decentraland/webrtc-broker/pkg/simulation"
	"github.com/decentraland/world/pkg/protocol"
	pion "github.com/pion/webrtc/v2"
	"github.com/segmentio/ksuid"
)

const (
	parcelSize = 16
	maxParcel  = 150
	minParcel  = -150
)

func nowMs() float64 {
	return float64(time.Now().UnixNano() / int64(time.Millisecond))
}

func max(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

type V3 struct {
	X float64
	Y float64
	Z float64
}

func (v V3) Length() float64 {
	r := math.Sqrt(math.Pow(float64(v.X), 2) + math.Pow(float64(v.Y), 2) + math.Pow(float64(v.Z), 2))
	return r
}

func (v V3) Sub(a V3) V3 {
	return V3{v.X - a.X, v.Y - a.Y, v.Z - a.Z}
}

func (v V3) Add(a V3) V3 {
	return V3{v.X + a.X, v.Y + a.Y, v.Z + a.Z}
}

func (v V3) ScalarProd(n float64) V3 {
	return V3{v.X * n, v.Y * n, v.Z * n}
}

func (v V3) Normalize() V3 {
	len := v.Length()
	return v.ScalarProd(1 / len)
}

func EncodeTopicMessage(topic string, data proto.Message) ([]byte, error) {
	body, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}

	msg := &broker.TopicMessage{
		Type:  broker.MessageType_TOPIC,
		Topic: topic,
		Body:  body,
	}

	bytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func EncodeTopicIdentityMessage(topic string, data proto.Message) ([]byte, error) {
	body, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}

	msg := &broker.TopicIdentityMessage{
		Type:  broker.MessageType_TOPIC_IDENTITY,
		Topic: topic,
		Body:  body,
	}

	bytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

type ClientAuthenticator struct {
	IdentityURL  string
	EphemeralKey *ephemeral.EphemeralKey

	Email    string
	Password string

	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	Auth0Audience     string

	accessToken string
}

func (a *ClientAuthenticator) getAccessToken() (string, error) {
	if a.accessToken != "" {
		return a.accessToken, nil
	}

	auth0 := Auth0{
		Domain:       a.Auth0Domain,
		ClientID:     a.Auth0ClientID,
		ClientSecret: a.Auth0ClientSecret,
		Audience:     a.Auth0Audience,
		Email:        a.Email,
		Password:     a.Password,
	}

	auth := Auth{
		IdentityURL: a.IdentityURL,
		PubKey:      EncodePublicKey(a.EphemeralKey),
	}

	accessToken, err := ExecuteAuthFlow(&auth0, &auth)
	if err != nil {
		return accessToken, err
	}

	a.accessToken = accessToken

	return a.accessToken, nil
}

func (a *ClientAuthenticator) GenerateClientAuthMessage() (*broker.AuthMessage, error) {
	accessToken, err := a.getAccessToken()
	if err != nil {
		return nil, err
	}

	msg := []byte{}

	fields, err := a.EphemeralKey.MakeCredentials(msg, accessToken)
	if err != nil {
		return nil, err
	}

	data := protocol.AuthData{
		Signature:   fields["x-signature"],
		Identity:    fields["x-identity"],
		Timestamp:   fields["x-timestamp"],
		AccessToken: fields["x-access-token"],
	}

	encodedData, err := proto.Marshal(&data)
	if err != nil {
		return nil, err
	}

	m := &broker.AuthMessage{
		Type: broker.MessageType_AUTH,
		Role: broker.Role_CLIENT,
		Body: encodedData,
	}

	return m, nil
}

func (a *ClientAuthenticator) GenerateClientConnectURL(coordinatorURL string) (string, error) {
	u, err := url.Parse(coordinatorURL)
	u.Path = path.Join(u.Path, "/connect")
	if err != nil {
		return "", nil
	}

	accessToken, err := a.getAccessToken()
	if err != nil {
		return "", nil
	}

	msg := fmt.Sprintf("GET:%s", u.String())
	fields, err := a.EphemeralKey.MakeCredentials([]byte(msg), accessToken)
	if err != nil {
		return "", err
	}

	v := url.Values{}
	v.Set("signature", fields["x-signature"])
	v.Set("identity", fields["x-identity"])
	v.Set("timestamp", fields["x-timestamp"])
	v.Set("access-token", fields["x-access-token"])

	u.RawQuery = v.Encode()
	return u.String(), nil
}

type BotOptions struct {
	Auth           authentication.ClientAuthenticator
	CoordinatorURL string
	Checkpoints    []V3
	DurationMs     uint
	Log            zerolog.Logger
	TrackStats     bool
}

func StartBot(options *BotOptions) {
	log := options.Log

	if len(options.Checkpoints) < 2 {
		log.Fatal().Msg("invalid path, need at least two checkpoints")
	}

	config := simulation.Config{
		Auth:           options.Auth,
		CoordinatorURL: options.CoordinatorURL,
		ICEServers: []pion.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		Log: log,
	}

	if options.TrackStats {
		trackCh := make(chan []byte, 256)
		config.OnMessageReceived = func(reliable bool, msgType broker.MessageType, raw []byte) {
			if !reliable && msgType == broker.MessageType_TOPIC_FW {
				trackCh <- raw
			}
		}

		go func() {
			peers := make(map[uint64]*simulation.Stats)
			topicFwMessage := broker.TopicFWMessage{}
			dataHeader := protocol.DataHeader{}

			onMessage := func(rawMsg []byte) {
				if err := proto.Unmarshal(rawMsg, &topicFwMessage); err != nil {
					log.Error().Err(err).Msg("error unmarshalling data message")
					return
				}

				if err := proto.Unmarshal(topicFwMessage.Body, &dataHeader); err != nil {
					log.Error().Err(err).Msg("error unmarshalling data header")
					return
				}

				if dataHeader.Category != protocol.Category_POSITION {
					return
				}

				alias := topicFwMessage.FromAlias
				stats := peers[alias]

				if stats == nil {
					stats = &simulation.Stats{}
					peers[alias] = stats
				}

				stats.Seen(time.Now())
			}

			reportTicker := time.NewTicker(30 * time.Second)
			defer reportTicker.Stop()

			for {
				select {
				case rawMsg := <-trackCh:
					onMessage(rawMsg)

					n := len(trackCh)
					for i := 0; i < n; i++ {
						rawMsg = <-trackCh
						onMessage(rawMsg)
					}
				case <-reportTicker.C:
					fmt.Println("Avg duration between position messages")
					for alias, stats := range peers {
						fmt.Printf("%d: %f ms (%d messages)\n", alias, stats.Avg(), stats.Samples())

						if time.Since(stats.LastSeen).Seconds() > 1 {
							delete(peers, alias)
						}
					}
				}

			}
		}()
	}

	client := simulation.Start(&config)
	checkpoints := options.Checkpoints

	totalDistance := 0.0
	for i := 1; i < len(checkpoints); i++ {
		totalDistance += checkpoints[i].Sub(checkpoints[i-1]).Length()
	}

	// NOTE: velocity in ms
	vMs := totalDistance / float64(options.DurationMs)

	p := checkpoints[0]
	nextCheckpointIndex := 1
	lastPositionMsg := time.Now()

	profileTicker := time.NewTicker(1 * time.Second)
	positionTicker := time.NewTicker(100 * time.Millisecond)
	chatTicker := time.NewTicker(10 * time.Second)
	defer profileTicker.Stop()
	defer positionTicker.Stop()
	defer chatTicker.Stop()

	hashLocation := func() string {
		parcelX := (int(p.X/parcelSize) + maxParcel) >> 2
		parcelZ := (int(p.Z/parcelSize) + maxParcel) >> 2
		hash := fmt.Sprintf("%d:%d", parcelX, parcelZ)
		return hash
	}

	for {
		select {
		case <-profileTicker.C:
			ms := nowMs()
			bytes, err := EncodeTopicIdentityMessage(hashLocation(), &protocol.ProfileData{
				Category:       protocol.Category_PROFILE,
				Time:           ms,
				ProfileVersion: "1",
			})
			if err != nil {
				log.Fatal().Err(err).Msg("encode profile failed")
			}
			client.SendReliable <- bytes
		case <-chatTicker.C:
			ms := nowMs()
			bytes, err := EncodeTopicMessage(hashLocation(), &protocol.ChatData{
				Category:  protocol.Category_CHAT,
				Time:      ms,
				MessageId: ksuid.New().String(),
				Text:      "hi",
			})
			if err != nil {
				log.Fatal().Err(err).Msg("encode chat failed")
			}
			client.SendReliable <- bytes
		case <-positionTicker.C:
			nextCheckpoint := checkpoints[nextCheckpointIndex]
			v := nextCheckpoint.Sub(p)
			tMax := float64(v.Length()) / vMs
			dt := float64(time.Since(lastPositionMsg).Nanoseconds() / int64(time.Millisecond))

			if dt < tMax {
				dir := v.Normalize()
				p = p.Add(dir.ScalarProd(dt * vMs))
			} else {
				if nextCheckpointIndex == len(checkpoints)-1 {
					nextCheckpointIndex = 0
				} else {
					nextCheckpointIndex++
				}
				p = nextCheckpoint
			}

			topics := make(map[string]bool)
			radius := 4
			parcelX := int(p.X / parcelSize)
			parcelZ := int(p.Z / parcelSize)

			minX := ((max(minParcel, parcelX-radius) + maxParcel) >> 2) << 2
			maxX := ((min(maxParcel, parcelX+radius) + maxParcel) >> 2) << 2
			minZ := ((max(minParcel, parcelZ-radius) + maxParcel) >> 2) << 2
			maxZ := ((min(maxParcel, parcelZ+radius) + maxParcel) >> 2) << 2

			newTopics := make(map[string]bool)
			topicsChanged := false

			for x := minX; x <= maxX; x += 4 {
				for z := minZ; z <= maxZ; z += 4 {
					hash := fmt.Sprintf("%d:%d", x>>2, z>>2)
					newTopics[hash] = true
					if !topics[hash] {
						topicsChanged = true
					}
				}
			}

			if topicsChanged {
				topics = newTopics
				client.SendTopicSubscriptionMessage(newTopics)
			}

			ms := nowMs()
			bytes, err := EncodeTopicMessage(hashLocation(), &protocol.PositionData{
				Category:  protocol.Category_POSITION,
				Time:      ms,
				PositionX: float32(p.X),
				PositionY: float32(p.Y),
				PositionZ: float32(p.Z),
				RotationX: 0,
				RotationY: 0,
				RotationZ: 0,
				RotationW: 0,
			})
			if err != nil {
				log.Fatal().Err(err).Msg("encode position failed")
			}

			client.SendUnreliable <- bytes
			lastPositionMsg = time.Now()
		}
	}
}
