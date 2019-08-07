package main

import (
	"fmt"
	"log"
	"math/rand"

	brokerAuth "github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/world/internal/cli"
	"github.com/decentraland/world/internal/commons/config"
)

type rootConfig struct {
	CoordinatorURL string `overwrite-flag:"coordinatorURL" validate:"required"`
}

func startRndBot(coordinatorURL string, trackStats bool) {
	radius := 3
	auth := &brokerAuth.NoopAuthenticator{}
	var checkpoints [6]cli.V3

	for i := 0; i < len(checkpoints); i++ {
		p := &checkpoints[i]

		p.X = float64(rand.Intn(10)*radius*2 - radius)
		p.Y = 1.6
		p.Z = float64(rand.Intn(10)*radius*2 - radius)
	}

	opts := cli.BotOptions{
		CoordinatorURL: coordinatorURL,
		Auth:           auth,
		Checkpoints:    checkpoints[:],
		DurationMs:     10000,
		TrackStats:     trackStats,
	}

	cli.StartBot(&opts)
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting test: ", conf.CoordinatorURL)

	for i := 0; i < 10; i++ {
		go startRndBot(conf.CoordinatorURL, false)
	}

	startRndBot(conf.CoordinatorURL, true)
}
