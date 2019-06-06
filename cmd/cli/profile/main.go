package main

import (
	"flag"
	"log"
	"os"

	"bufio"
	"strings"

	"github.com/decentraland/world/internal/cli"
)

func main() {
	authURL := flag.String("authURL", "http://localhost:9001", "")
	profileURL := flag.String("profileURL", "http://localhost:9002", "")

	email := flag.String("email", "", "")
	password := flag.String("password", "", "")
	auth0Domain := flag.String("auth0Domain", "", "")
	keyPath := flag.String("keyPath", "", "")
	auth0ClientID := flag.String("auth0ClientID", "", "")
	auth0ClientSecret := flag.String("auth0ClientSecret", "", "")
	auth0Audience := flag.String("auth0Audience", "", "")

	store := flag.Bool("store", false, "")
	retrieve := flag.Bool("retrieve", false, "")
	flag.Parse()

	// TODO: command line validations
	// if *email == "" || *password == "" || *auth0Domain == "" {
	// 	log.Fatal("--email, --password, --auth0Domain is required")
	// }

	if *store == *retrieve {
		log.Fatal("please specify --store or --retrieve")
	}

	ephemeralKey, err := cli.ReadEphemeralKeyFromFile(*keyPath)
	if err != nil {
		log.Fatalf("error loading ephemeral key: %v", err)
	}

	auth0 := cli.Auth0{
		Domain:       *auth0Domain,
		ClientID:     *auth0ClientID,
		ClientSecret: *auth0ClientSecret,
		Audience:     *auth0Audience,
		Email:        *email,
		Password:     *password,
	}

	auth := cli.Auth{
		AuthURL: *authURL,
		PubKey:  cli.EncodePublicKey(ephemeralKey),
	}

	accessToken, err := cli.ExecuteAuthFlow(&auth0, &auth)
	if err != nil {
		log.Fatalf("auth failure: %v", err)
	}

	client := cli.ProfileClient{
		ProfileURL:   *profileURL,
		EphemeralKey: ephemeralKey,
	}

	if *retrieve {
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
