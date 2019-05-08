package token

import (
	"crypto/ecdsa"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Generator struct {
	key      *ecdsa.PrivateKey
	version  string
	duration time.Duration
}

func New(key *ecdsa.PrivateKey, version string, duration time.Duration) *Generator {
	return &Generator{
		key:      key,
		version:  version,
		duration: duration,
	}
}

func (g *Generator) NewToken(userID, ephemeralKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"user_id":       userID,
		"ephemeral_key": ephemeralKey,
		"version":       g.version,
		"exp":           time.Now().Add(time.Second * g.duration).Unix(),
	})
	ret, err := token.SignedString(g.key)
	return ret, err
}
