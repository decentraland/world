package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rs/zerolog"

	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commtest"
)

type rootConfig struct {
	CoordinatorURL string `overwrite-flag:"coordinatorURL" validate:"required"`
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting test: ", conf.CoordinatorURL)

	for i := 0; i < 25; i++ {
		for j := 0; j < 2; j++ {
			topic := fmt.Sprintf("topic-%d", i)
			subscription := make(map[string]bool)
			subscription[topic] = true

			opts := commtest.Options{
				CoordinatorURL: conf.CoordinatorURL,
				Subscription:   subscription,
				Topic:          topic,
			}

			if j == 0 {
				name := fmt.Sprintf("client-%d", i)
				opts.Log = zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Str("name", name).Logger()
			} else {
				name := fmt.Sprintf("client-%d-observer", i)
				opts.Log = zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Str("name", name).Logger()
				opts.TrackStats = true
			}

			go commtest.StartBot(opts)
		}
	}

	select {}
}
