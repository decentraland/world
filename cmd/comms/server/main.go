package main

import (
	"fmt"
	"github.com/decentraland/world/internal/commons/version"

	"github.com/decentraland/webrtc-broker/pkg/commserver"
	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
)

type rootConfig struct {
	IdentityURL    string `overwrite-flag:"authURL" validate:"required"`
	CoordinatorURL string `overwrite-flag:"coordinatorURL" flag-usage:"coordinator url" validate:"required"`
	CommServer     struct {
		LogLevel     string `overwrite-flag:"logLevel"`
		ServerSecret string `overwrite-flag:"serverSecret"`
		AuthTTL      int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
	}
}

func main() {
	log := logging.New()
	defer logging.LogPanic()

	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	serverAuth, err := auth.MakeAuthenticator(&auth.AuthenticatorConfig{
		IdentityURL: conf.IdentityURL,
		Secret:      conf.CommServer.ServerSecret,
		RequestTTL:  conf.CommServer.AuthTTL,
	})

	if err != nil {
		log.WithError(err).Fatal("cannot build authenticator")
	}

	config := commserver.Config{
		Auth: serverAuth,
		Log:  log,
		ICEServers: []commserver.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		CoordinatorURL: fmt.Sprintf("%s/discover", conf.CoordinatorURL),
	}

	if err := logging.SetLevel(log, conf.CommServer.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	state, err := commserver.MakeState(&config)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting communication server - version: %s", version.Version())

	if err := commserver.ConnectCoordinator(state); err != nil {
		log.Fatal("connect coordinator failure ", err)
	}

	go commserver.ProcessMessagesQueue(state)
	commserver.Process(state)
}
