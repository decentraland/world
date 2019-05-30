package main

import (
	"fmt"
	"net/http"

	"github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/webrtc-broker/pkg/coordinator"
	configuration "github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type coordinatorConfig struct {
	CoordinatorHost string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
	CoordinatorPort int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
	Version         string `overwrite-flag:"version"`
	LogLevel        string `overwrite-flag:"logLevel"`
	NoopAuthEnabled bool   `overwrite-flag:"noopEnabled"`
}

func main() {
	log := logging.New()
	defer logging.LogPanic()

	var conf coordinatorConfig
	if err := configuration.ReadConfiguration("config/comms/config", &conf); err != nil {
		log.Fatal(err)
	}

	if err := logging.SetLevel(log, conf.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	auth := authentication.Make()
	if conf.NoopAuthEnabled {
		auth.AddOrUpdateAuthenticator("noop", &authentication.NoopAuthenticator{})
	}

	config := coordinator.Config{
		ServerSelector: coordinator.MakeRandomServerSelector(),
		Auth:           auth,
		Log:            log,
	}
	state := coordinator.MakeState(&config)

	go coordinator.Process(state)

	mux := http.NewServeMux()
	coordinator.Register(state, mux)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	addr := fmt.Sprintf("%s:%d", conf.CoordinatorHost, conf.CoordinatorPort)
	log.Infof("starting coordinator %s - version: %s", addr, conf.Version)
	log.Fatal(http.ListenAndServe(addr, mux))
}
