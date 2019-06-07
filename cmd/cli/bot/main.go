package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/decentraland/world/internal/cli"
	"github.com/decentraland/world/internal/commons/config"
)

type rootConfig struct {
	IdentityURL    string `overwrite-flag:"authURL" validate:"required"`
	CoordinatorURL string `overwrite-flag:"coordinatorURL" validate:"required"`
	Auth0          struct {
		Domain string `overwrite-flag:"auth0Domain" validate:"required"`
	}
	Cli struct {
		Auth0ClientID     string `overwrite-flag:"auth0ClientID" validate:"required"`
		Auth0Audience     string `overwrite-flag:"auth0Audience" validate:"required"`
		Auth0ClientSecret string `overwrite-flag:"auth0ClientSecret" validate:"required"`
		Email             string `overwrite-flag:"email" validate:"required"`
		Password          string `overwrite-flag:"password" validate:"required"`
		KeyPath           string `overwrite-flag:"keyPath" validate:"required"`
	}
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	centerXP := flag.Int("centerX", 0, "")
	centerYP := flag.Int("centerY", 0, "")
	radiusP := flag.Int("radius", 3, "radius (in parcels) from the center")
	trackStats := flag.Bool("trackStats", false, "")

	fmt.Println("running random simulation")

	ephemeralKey, err := cli.ReadEphemeralKeyFromFile(conf.Cli.KeyPath)
	if err != nil {
		log.Fatalf("error loading ephemeral key: %v", err)
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

	centerX := *centerXP
	centerY := *centerYP
	radius := *radiusP

	var checkpoints [6]cli.V3

	for i := 0; i < len(checkpoints); i++ {
		p := &checkpoints[i]

		p.X = float64(centerX + rand.Intn(10)*radius*2 - radius)
		p.Y = 1.6
		p.Z = float64(centerY + rand.Intn(10)*radius*2 - radius)
	}

	opts := cli.BotOptions{
		CoordinatorURL: conf.CoordinatorURL,
		Auth:           auth,
		Checkpoints:    checkpoints[:],
		DurationMs:     10000,
		TrackStats:     *trackStats,
	}

	go cli.StartBot(&opts)

	select {}
}
