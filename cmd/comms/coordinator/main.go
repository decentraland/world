package main

import (
	"fmt"
	"net/http"

	"github.com/decentraland/webrtc-broker/pkg/coordinator"
	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
)

type rootConfig struct {
	IdentityPubKeyURL string `overwrite-flag:"authURL" validate:"required"`

	Coordinator struct {
		Version      string `overwrite-flag:"version"`
		Host         string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
		Port         int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
		LogLevel     string `overwrite-flag:"logLevel"`
		AuthTTL      int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
		ServerSecret string `overwrite-flag:"serverSecret" validate:"required"`
	}
}

func main() {
	log := logging.New()
	defer logging.LogPanic()

	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	if err := logging.SetLevel(log, conf.Coordinator.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	coordinatorAuth, err := auth.MakeAuthenticator(&auth.AuthenticatorConfig{
		AuthURL:    conf.IdentityPubKeyURL,
		Secret:     conf.Coordinator.ServerSecret,
		RequestTTL: conf.Coordinator.AuthTTL,
	})

	if err != nil {
		log.WithError(err).Fatal("cannot build authenticator")
	}

	config := coordinator.Config{Auth: coordinatorAuth, Log: log}
	state := coordinator.MakeState(&config)

	go coordinator.Start(state)

	mux := http.NewServeMux()
	coordinator.Register(state, mux)

	addr := fmt.Sprintf("%s:%d", conf.Coordinator.Host, conf.Coordinator.Port)
	log.Infof("starting coordinator %s - version: %s", addr, conf.Coordinator.Version)
	log.Fatal(http.ListenAndServe(addr, mux))
}
