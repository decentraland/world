package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/decentraland/world/internal/cli"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
)

type rootConfig struct {
	IdentityURL    string `overwrite-flag:"authURL" validate:"required"`
	CoordinatorURL string `overwrite-flag:"coordinatorURL" validate:"required"`
	Auth0          struct {
		Domain string `overwrite-flag:"auth0Domain" validate:"required"`
	}
	Cli struct {
		LogLevel          string `overwrite-flag:"logLevel"`
		Auth0ClientID     string `overwrite-flag:"auth0ClientID" validate:"required"`
		Auth0Audience     string `overwrite-flag:"auth0Audience" validate:"required"`
		Auth0ClientSecret string `overwrite-flag:"auth0ClientSecret" validate:"required"`
		Email             string `overwrite-flag:"email" validate:"required"`
		Password          string `overwrite-flag:"password" validate:"required"`
		KeyPath           string `overwrite-flag:"keyPath" validate:"required"`
		CenterX           int    `overwrite-flag:"centerX"`
		CenterY           int    `overwrite-flag:"centerY"`
		Radius            int    `overwrite-flag:"radius" flag-usage:"radius in parcels"`
		TrackStats        bool   `overwrite-flag:"trackStats"`
	}
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	fmt.Println("running random simulation")

	log, err := logging.New(&logging.LoggerConfig{Level: conf.Cli.LogLevel})
	if err != nil {
		log.Fatal().Msg("error setting log level")
	}
	defer logging.LogPanic(log)

	ephemeralKey, err := cli.ReadEphemeralKeyFromFile(conf.Cli.KeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading ephemeral key")
	}

	auth := &cli.ClientAuthenticator{
		IdentityURL:       conf.IdentityURL,
		EphemeralKey:      ephemeralKey,
		Email:             conf.Cli.Email,
		Password:          conf.Cli.Password,
		Auth0Domain:       conf.Auth0.Domain,
		Auth0ClientID:     conf.Cli.Auth0ClientID,
		Auth0ClientSecret: conf.Cli.Auth0ClientSecret,
		Auth0Audience:     conf.Cli.Auth0Audience,
	}

	var checkpoints [6]cli.V3

	for i := 0; i < len(checkpoints); i++ {
		p := &checkpoints[i]

		p.X = float64(conf.Cli.CenterX + rand.Intn(10)*conf.Cli.Radius*2 - conf.Cli.Radius)
		p.Y = 1.6
		p.Z = float64(conf.Cli.CenterY + rand.Intn(10)*conf.Cli.Radius*2 - conf.Cli.Radius)
	}

	opts := cli.BotOptions{
		CoordinatorURL: conf.CoordinatorURL,
		Auth:           auth,
		Checkpoints:    checkpoints[:],
		DurationMs:     10000,
		TrackStats:     conf.Cli.TrackStats,
		Log:            log,
	}

	cli.StartBot(&opts)
}
