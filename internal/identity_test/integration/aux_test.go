package integration

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/decentraland/world/internal/identity/data"
	"github.com/dgrijalva/jwt-go"
)

type U struct {
	Token        string
	Email        string
	ID           string
	EphemeralKey string
}

func (u U) gen() *testUser {
	setProp := func(prop *string, value string) {
		if value != "" {
			*prop = value
		}
		if value == "empty" {
			*prop = ""
		}
	}

	user := validUser()
	setProp(&user.Token, u.Token)
	setProp(&user.ID, u.ID)
	setProp(&user.Email, u.Email)
	setProp(&user.EphemeralKey, u.EphemeralKey)
	return &user
}

type testUser struct {
	Token        string
	Email        string
	ID           string
	EphemeralKey string
	SessionJWT   string
}

func (u *testUser) User() data.User {
	return data.User{Email: u.Email, UserID: u.ID}
}

func validUser() testUser {
	return testUser{
		Token:        "valid_user_token",
		Email:        "valid_user_name",
		ID:           "valid_user_id",
		EphemeralKey: "0304f7febb7837fa61ba07ccadcc470997063d5a6962f6d9651ece233f9acf6655",
	}
}

func getJSONValue(data []byte, path string) string {
	var j interface{}
	_ = json.Unmarshal(data, &j)
	m := j.(map[string]interface{})
	s := m[path]
	if s != nil {
		return s.(string)
	} else {
		return ""
	}
}

func getKeyJWT(x *big.Int, y *big.Int) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("worng signing method")
		}
		return &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		}, nil
	}
}
