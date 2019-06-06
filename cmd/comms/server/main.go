package main

import (
	"fmt"

	"github.com/decentraland/webrtc-broker/pkg/commserver"
	"github.com/decentraland/world/internal/commons/auth"
	configuration "github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
)

type authConfiguration struct {
	ServerSecret string `overwrite-flag:"serverSecret"`
	AuthURL      string `overwrite-flag:"authURL" flag-usage:"identity service public key url"`
	RequestTTL   int64  `overwrite-flag:"authTtl" flag-usage:"request time to live"`
}

type commServerConfig struct {
	CoordinatorURL string `overwrite-flag:"coordinatorURL" flag-usage:"coordinator url" validate:"required"`
	Version        string `overwrite-flag:"version"`
	LogLevel       string `overwrite-flag:"logLevel"`
	Auth           authConfiguration
}

func main() {
	log := logging.New()
	defer logging.LogPanic()

	var conf commServerConfig
	if err := configuration.ReadConfiguration("config/comms/config", &conf); err != nil {
		log.Fatal(err)
	}

	serverAuth, err := auth.MakeAuthenticator(&auth.AuthenticatorConfig{
		Secret:     conf.Auth.ServerSecret,
		AuthURL:    conf.Auth.AuthURL,
		RequestTTL: conf.Auth.RequestTTL,
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

	if err := logging.SetLevel(log, conf.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	state, err := commserver.MakeState(&config)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting communication server - version: %s", conf.Version)

	if err := commserver.ConnectCoordinator(state); err != nil {
		log.Fatal("connect coordinator failure ", err)
	}

	go commserver.ProcessMessagesQueue(state)
	commserver.Process(state)
}
