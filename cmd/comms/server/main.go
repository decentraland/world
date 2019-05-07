package main

import (
	"fmt"

	"github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/webrtc-broker/pkg/commserver"
	configuration "github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
)

type commServerConfig struct {
	CoordinatorURL  string `overwrite-flag:"coordinatorURL" flag-usage:"coordinator url" validate:"required"`
	Version         string `overwrite-flag:"version"`
	LogLevel        string `overwrite-flag:"logLevel"`
	NoopAuthEnabled bool   `overwrite-flag:"noopEnabled"`
}

func main() {
	log := logging.New()
	defer logging.LogPanic()

	var conf commServerConfig
	if err := configuration.ReadConfiguration("config/comms/config", &conf); err != nil {
		log.Fatal(err)
	}

	auth := authentication.Make()
	config := commserver.Config{
		Auth: auth,
		Log:  log,
		ICEServers: []commserver.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		CoordinatorURL: fmt.Sprintf("%s/discover", conf.CoordinatorURL),
		AuthMethod:     "noop",
	}

	if err := logging.SetLevel(log, conf.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	if conf.NoopAuthEnabled {
		auth.AddOrUpdateAuthenticator("noop", &authentication.NoopAuthenticator{})
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
