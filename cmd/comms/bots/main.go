package main

import (
	"flag"
	"log"
	"math/rand"

	"github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/world/internal/comms/bots"
)

func main() {
	coordinatorURL := flag.String("coordinatorURL", "ws://localhost:9000/connect", "")
	centerXP := flag.Int("centerX", 0, "")
	centerYP := flag.Int("centerY", 0, "")
	radiusP := flag.Int("radius", 3, "radius (in parcels) from the center")
	subscribeP := flag.Bool("subscribe", false, "subscribe to the position and profile topics of the comm area")
	nBotsP := flag.Int("n", 5, "number of bots")
	authMethodP := flag.String("authMethod", "noop", "")
	trackStats := flag.Bool("trackStats", false, "")

	flag.Parse()

	log.Println("running random simulation")

	auth := authentication.Make()
	auth.AddOrUpdateAuthenticator("noop", &authentication.NoopAuthenticator{})

	centerX := *centerXP
	centerY := *centerYP
	radius := *radiusP
	subscribe := *subscribeP
	authMethod := *authMethodP

	for i := 0; i < *nBotsP; i++ {
		var checkpoints [6]bots.V3

		for i := 0; i < len(checkpoints); i++ {
			p := &checkpoints[i]

			p.X = float64(centerX + rand.Intn(10)*radius*2 - radius)
			p.Y = 1.6
			p.Z = float64(centerY + rand.Intn(10)*radius*2 - radius)
		}

		opts := bots.BotOptions{
			CoordinatorURL:            *coordinatorURL,
			Auth:                      auth,
			AuthMethod:                authMethod,
			Checkpoints:               checkpoints[:],
			DurationMs:                10000,
			SubscribeToPositionTopics: subscribe,
			TrackStats:                *trackStats,
		}

		go bots.Start(&opts)
	}

	select {}
}
