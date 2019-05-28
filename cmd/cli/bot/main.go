package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/decentraland/auth-go/pkg/ephemeral"
	"github.com/decentraland/world/internal/cli"
)

func main() {
	coordinatorURL := flag.String("coordinatorURL", "ws://localhost:9000/connect", "")

	centerXP := flag.Int("centerX", 0, "")
	centerYP := flag.Int("centerY", 0, "")
	radiusP := flag.Int("radius", 3, "radius (in parcels) from the center")
	trackStats := flag.Bool("trackStats", false, "")

	authURL := flag.String("authURL", "http://localhost:9001", "")

	email := flag.String("email", "", "")
	password := flag.String("password", "", "")
	auth0Domain := flag.String("auth0Domain", "", "")
	keyPath := flag.String("keyPath", "", "")
	auth0ClientID := flag.String("auth0ClientID", "", "")
	auth0ClientSecret := flag.String("auth0ClientSecret", "", "")
	auth0Audience := flag.String("auth0Audience", "", "")

	flag.Parse()

	fmt.Println("running random simulation")

	// TODO: command line validations
	// if *email == "" || *password == "" || *auth0Domain == "" {
	// 	log.Fatal("--email, --password, --auth0Domain is required")
	// }

	privateKey, err := cli.ReadKey(*keyPath)
	if err != nil {
		log.Fatalf("cannot load private key from file %s: %v", *keyPath, err)
	}

	config := ephemeral.EphemeralKeyConfig{
		PrivateKey: privateKey,
	}
	ephemeralKey, err := ephemeral.NewEphemeralKey(&config)
	if err != nil {
		log.Fatalf("cannot create ephemeral key: %v", err)
	}

	auth := &cli.ClientAuthenticator{
		AuthURL:           *authURL,
		EphemeralKey:      ephemeralKey,
		Email:             *email,
		Password:          *password,
		Auth0Domain:       *auth0Domain,
		Auth0ClientID:     *auth0ClientID,
		Auth0ClientSecret: *auth0ClientSecret,
		Auth0Audience:     *auth0Audience,
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
		CoordinatorURL: *coordinatorURL,
		Auth:           auth,
		Checkpoints:    checkpoints[:],
		DurationMs:     10000,
		TrackStats:     *trackStats,
	}

	go cli.StartBot(&opts)

	select {}
}
