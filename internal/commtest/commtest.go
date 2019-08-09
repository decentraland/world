package commtest

import (
	"time"

	"github.com/rs/zerolog"

	brokerAuth "github.com/decentraland/webrtc-broker/pkg/authentication"
	broker "github.com/decentraland/webrtc-broker/pkg/protocol"
	"github.com/decentraland/webrtc-broker/pkg/simulation"
	"github.com/decentraland/world/internal/cli"
	"github.com/decentraland/world/pkg/protocol"
	"github.com/golang/protobuf/proto"

	pion "github.com/pion/webrtc/v2"
)

type Options struct {
	CoordinatorURL string
	Topic          string
	Subscription   map[string]bool
	TrackStats     bool
	Log            zerolog.Logger
}

func StartBot(opts Options) {
	log := opts.Log
	config := simulation.Config{
		Auth:           &brokerAuth.NoopAuthenticator{},
		CoordinatorURL: opts.CoordinatorURL,
		ICEServers: []pion.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		Log: log,
	}

	if opts.TrackStats {
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
					for alias, stats := range peers {
						log.Info().Msgf("%d: %f ms (%d messages)", alias, stats.Avg(), stats.Samples())

						if time.Since(stats.LastSeen).Seconds() > 1 {
							delete(peers, alias)
						}
					}
				}

			}
		}()
	}

	client := simulation.Start(&config)
	client.SendTopicSubscriptionMessage(opts.Subscription)

	if opts.Topic != "" {
		positionTicker := time.NewTicker(100 * time.Millisecond)
		defer positionTicker.Stop()

		for {
			<-positionTicker.C
			bytes, err := cli.EncodeTopicMessage(opts.Topic, &protocol.PositionData{
				Category: protocol.Category_POSITION,
			})
			if err != nil {
				log.Fatal().Err(err).Msg("encode position failed")
			}

			client.SendUnreliable <- bytes
		}
	}
}
