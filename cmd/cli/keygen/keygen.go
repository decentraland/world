package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/decentraland/world/internal/commons/utils"
	log "github.com/sirupsen/logrus"
)

func main() {
	curveFmt := flag.String("curve", "", "s256 (used for ephemeral keys) or p256 (used by identity service)")
	outputDir := flag.String("outputDir", "", "")
	flag.Parse()

	if *curveFmt == "" {
		log.Fatal("--curve is required")
	}

	if *outputDir == "" {
		log.Fatal("--outputDir is required")
	}

	if *curveFmt == "s256" {
		key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			log.WithError(err).Fatal("Fail to generate random ECDSA key")
		}

		strPrivKey := hex.EncodeToString(crypto.FromECDSA(key))
		if err := writeKey(strPrivKey, "client.key", *outputDir); err != nil {
			log.WithError(err).Fatal("Fail to persist private key")
		}
	} else if *curveFmt == "p256" {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.WithError(err).Fatal("Fail to generate random ECDSA key")
		}

		strPvKey, err := utils.PemEncodePrivateKey(key)
		if err != nil {
			log.WithError(err).Fatal("Error formatting private key")
		}

		strPubKey, err := utils.PemEncodePublicKey(&key.PublicKey)
		if err != nil {
			log.WithError(err).Fatal("Error formatting public key")
		}

		if err := writeKey(strPvKey, "identity.key", *outputDir); err != nil {
			log.WithError(err).Fatal("Fail to persist private key")
		}

		if err := writeKey(strPubKey, "identity.pem", *outputDir); err != nil {
			log.WithError(err).Fatal("Fail to persist public key")
		}
	} else {
		log.Fatalf("Invalid curve %s", *curveFmt)
	}
}

func writeKey(key, fileName, rootDir string) error {
	path := filepath.Join(rootDir, fileName)
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.WriteString(out, key)
	if err != nil {
		return err
	}
	return nil
}
