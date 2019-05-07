package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/decentraland/world/internal/commons/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatalf("You need to provide the output directory")
	}

	outputDir := os.Args[1]

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.WithError(err).Fatalf("Fail to generate random ECDSA key")
	}

	strPvKey, err := utils.PemEncodePrivateKey(key)
	if err != nil {
		log.WithError(err).Fatalf("Error formatting private key")
	}

	strPubKey, err := utils.PemEncodePublicKey(&key.PublicKey)
	if err != nil {
		log.WithError(err).Fatalf("Error formatting public key")
	}

	if err := writeKey(strPvKey, "generated.key", outputDir); err != nil {
		log.WithError(err).Fatalf("Fail to persist private key")
	}

	if err := writeKey(strPubKey, "generated.pem", outputDir); err != nil {
		log.WithError(err).Fatalf("Fail to persist public key")
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
