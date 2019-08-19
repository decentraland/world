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
	SpawnObserver  bool   `overwrite-flag:"observer"`
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting test: ", conf.CoordinatorURL)

	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("client-%d", i)
		log := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Str("name", name).Logger()

		go commtest.StartBot(commtest.Options{
			CoordinatorURL: conf.CoordinatorURL,
			Topic:          "testtopic",
			Subscription:   map[string]bool{"testtopic": true},
			TrackStats:     false,
			Log:            log,
		})
	}

	if conf.SpawnObserver {
		log := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Str("name", "observer").Logger()
		commtest.StartBot(commtest.Options{
			CoordinatorURL: conf.CoordinatorURL,
			Topic:          "testtopic",
			Subscription:   map[string]bool{"testtopic": true},
			TrackStats:     true,
			Log:            log,
		})
	} else {
		select {}
	}
}
