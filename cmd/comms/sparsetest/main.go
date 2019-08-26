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

	SparseTest struct {
		NTopics  int `overwrite-flag:"n"`
		Duration int `overwrite-flag:"duration" flag-usage:"duration in seconds"`
	}
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting test: ", conf.CoordinatorURL)

	for i := 0; i < conf.SparseTest.NTopics; i++ {
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

	if conf.SparseTest.Duration > 0 {
		time.Sleep(time.Duration(conf.SparseTest.Duration) * time.Second)
		os.Exit(0)
	} else {
		select {}
	}
}
