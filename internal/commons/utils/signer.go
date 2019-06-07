package utils

import (
	"crypto/ecdsa"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenGenerator struct {
	key      *ecdsa.PrivateKey
	version  string
	duration time.Duration
}

func NewTokenGenerator(key *ecdsa.PrivateKey, version string, duration time.Duration) *TokenGenerator {
	return &TokenGenerator{
		key:      key,
		version:  version,
		duration: duration,
	}
}

func (g *TokenGenerator) NewToken(userID, ephemeralKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"user_id":       userID,
		"ephemeral_key": ephemeralKey,
		"version":       g.version,
		"exp":           time.Now().Add(time.Second * g.duration).Unix(),
	})
	ret, err := token.SignedString(g.key)
	return ret, err
}
