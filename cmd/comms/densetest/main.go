package main

import (
	"fmt"
	"log"
	"os"

	"time"

	"github.com/rs/zerolog"

	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commtest"
)

type rootConfig struct {
	CoordinatorURL string `overwrite-flag:"coordinatorURL" validate:"required"`
	DenseTest      struct {
		NBots         int  `overwrite-flag:"n"`
		SpawnObserver bool `overwrite-flag:"observer"`
		Duration      int  `overwrite-flag:"duration" flag-usage:"duration in seconds"`
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

	for i := 0; i < conf.DenseTest.NBots; i++ {
		log := newLogger(fmt.Sprintf("client-%d", i))

		go commtest.StartBot(commtest.Options{
			CoordinatorURL: conf.CoordinatorURL,
			Topic:          "testtopic",
			Subscription:   map[string]bool{"testtopic": true},
			TrackStats:     false,
			Log:            log,
		})
	}

	if conf.DenseTest.Duration > 0 {
		go func() {
			time.Sleep(time.Duration(conf.DenseTest.Duration) * time.Second)
			os.Exit(0)
		}()
	}

	if conf.DenseTest.SpawnObserver {
		log := newLogger("observer")
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
