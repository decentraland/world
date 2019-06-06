package main

import (
	"fmt"
	"net/http"

	"github.com/decentraland/webrtc-broker/pkg/coordinator"
	"github.com/decentraland/world/internal/commons/auth"
	configuration "github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
)

type authConfiguration struct {
	ServerSecret string `overwrite-flag:"serverSecret" validate:"required"`
	AuthURL      string `overwrite-flag:"authURL" flag-usage:"identity service public key url"`
	RequestTTL   int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
}

type coordinatorConfig struct {
	CoordinatorHost string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
	CoordinatorPort int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
	Version         string `overwrite-flag:"version"`
	LogLevel        string `overwrite-flag:"logLevel"`
	Auth            authConfiguration
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

	coordinatorAuth, err := auth.MakeAuthenticator(&auth.AuthenticatorConfig{
		Secret:     conf.Auth.ServerSecret,
		AuthURL:    conf.Auth.AuthURL,
		RequestTTL: conf.Auth.RequestTTL,
	})

	if err != nil {
		log.WithError(err).Fatal("cannot build authenticator")
	}

	config := coordinator.Config{Auth: coordinatorAuth, Log: log}
	state := coordinator.MakeState(&config)

	go coordinator.Start(state)

	mux := http.NewServeMux()
	coordinator.Register(state, mux)

	addr := fmt.Sprintf("%s:%d", conf.CoordinatorHost, conf.CoordinatorPort)
	log.Infof("starting coordinator %s - version: %s", addr, conf.Version)
	log.Fatal(http.ListenAndServe(addr, mux))
}
