package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"time"

	"github.com/rs/zerolog"

	brokerAuth "github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/world/internal/cli"
	"github.com/decentraland/world/internal/commons/config"
)

type rootConfig struct {
	CoordinatorURL string `overwrite-flag:"coordinatorURL" validate:"required"`
	RealisticTest  struct {
		NBots         int  `overwrite-flag:"n"`
		SpawnObserver bool `overwrite-flag:"observer"`
		Duration      int  `overwrite-flag:"duration" flag-usage:"duration in seconds"`

		CenterX int `overwrite-flag:"centerX"`
		CenterY int `overwrite-flag:"centerY"`
		Radius  int `overwrite-flag:"radius" flag-usage:"radius in parcels"`
	}
}

func newLogger(name string) zerolog.Logger {
	return zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Str("name", name).Logger()
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting test: ", conf.CoordinatorURL)

	auth := &brokerAuth.NoopAuthenticator{}

	for i := 0; i < conf.RealisticTest.NBots; i++ {
		log := newLogger(fmt.Sprintf("client-%d", i))

		var checkpoints [6]cli.V3

		for i := 0; i < len(checkpoints); i++ {
			p := &checkpoints[i]

			p.X = float64(conf.RealisticTest.CenterX + rand.Intn(10)*conf.RealisticTest.Radius*2 - conf.RealisticTest.Radius)
			p.Y = 1.6
			p.Z = float64(conf.RealisticTest.CenterY + rand.Intn(10)*conf.RealisticTest.Radius*2 - conf.RealisticTest.Radius)
		}

		opts := cli.BotOptions{
			CoordinatorURL: conf.CoordinatorURL,
			Auth:           auth,
			Checkpoints:    checkpoints[:],
			DurationMs:     10000,
			TrackStats:     false,
			Log:            log,
		}

		go cli.StartBot(&opts)
	}

	if conf.RealisticTest.Duration > 0 {
		go func() {
			time.Sleep(time.Duration(conf.RealisticTest.Duration) * time.Second)
			os.Exit(0)
		}()
	}

	if conf.RealisticTest.SpawnObserver {
		log := newLogger("observer")

		var checkpoints [6]cli.V3
		for i := 0; i < len(checkpoints); i++ {
			p := &checkpoints[i]
			p.X = float64(conf.RealisticTest.CenterX)
			p.Y = 1.6
			p.Z = float64(conf.RealisticTest.CenterY)
		}

		opts := cli.BotOptions{
			CoordinatorURL: conf.CoordinatorURL,
			Auth:           auth,
			Checkpoints:    checkpoints[:],
			DurationMs:     10000,
			TrackStats:     true,
			Log:            log,
		}

		cli.StartBot(&opts)
	} else {
		select {}
	}
}
