package main

import (
	"log"
	"os"

	"bufio"
	"strings"

	"github.com/decentraland/world/internal/cli"
	"github.com/decentraland/world/internal/commons/config"
)

type rootConfig struct {
	IdentityURL string `overwrite-flag:"authURL" validate:"required"`
	ProfileURL  string `overwrite-flag:"profileURL" validate:"required"`
	Auth0       struct {
		Domain string `overwrite-flag:"auth0Domain" validate:"required"`
	}
	Cli struct {
		Store             bool   `overwrite-flag:"store"`
		Retrieve          bool   `overwrite-flag:"retrieve"`
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

	if conf.Cli.Store == conf.Cli.Retrieve {
		log.Fatal("please specify --store or --retrieve")
	}

	ephemeralKey, err := cli.ReadEphemeralKeyFromFile(conf.Cli.KeyPath)
	if err != nil {
		log.Fatalf("error loading ephemeral key: %v", err)
	}

	auth0 := cli.Auth0{
		Domain:       conf.Auth0.Domain,
		ClientID:     conf.Cli.Auth0ClientID,
		ClientSecret: conf.Cli.Auth0ClientSecret,
		Audience:     conf.Cli.Auth0Audience,
		Email:        conf.Cli.Email,
		Password:     conf.Cli.Password,
	}

	auth := cli.Auth{
		IdentityURL: conf.IdentityURL,
		PubKey:      cli.EncodePublicKey(ephemeralKey),
	}

	accessToken, err := cli.ExecuteAuthFlow(&auth0, &auth)
	if err != nil {
		log.Fatalf("auth failure: %v", err)
	}

	client := cli.ProfileClient{
		ProfileURL:   conf.ProfileURL,
		EphemeralKey: ephemeralKey,
	}

	if conf.Cli.Retrieve {
		profile, err := client.RetrieveProfile(accessToken)
		if err != nil {
			log.Fatalf("error retrieving profile: %v", err)
		}
		log.Println(profile)
	} else {
		builder := strings.Builder{}
		reader := bufio.NewReader(os.Stdin)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			builder.WriteString(scanner.Text())
		}

		if err = scanner.Err(); err != nil {
			log.Fatalf("error reading from stdin: %v", err)
		}

		profile := strings.NewReader(builder.String())
		if err = client.StoreProfile(accessToken, profile); err != nil {
			log.Fatalf("error storing profile: %v", err)
		}
	}
}
