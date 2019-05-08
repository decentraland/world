// +build integration

package profile_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"database/sql"

	"github.com/decentraland/auth-go/pkg/ephemeral"
	"github.com/decentraland/world/internal/auth"
	"github.com/decentraland/world/internal/profile"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var validProfile = `{
"id": "user1",
"email": "test@test.com",
"name": "testname",
"description": "desc",
"age": 98,
"avatar": {
	"version": "v1",
	"id": "something",
	"bodyShape": "cid-123123",
	"skinColor": { "r": 0.1, "g": 1, "b": 0 },
	"hairColor": { "r": 0.1, "g": 1, "b": 0 },
	"eyeColor": { "r": 0.1, "g": 1, "b": 0 },
	"eyes": "cid-12313",
	"eyebrow": "cid-12313",
	"mouth": "cid-12313",
	"wearables": [
		{
			"contentId": "cid-123123",
			"category": "torso",
			"mappings": [
				{
					"name": "file.png",
					"file": "cid-123213"
				}
			]
		}
	]
}
}`

func prepareDb(t *testing.T) *sql.DB {
	connStr := os.Getenv("PROFILE_TEST_DB_CONN_STR")
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	_, err = db.Exec("DROP TABLE IF EXISTS profiles")
	require.NoError(t, err)

	sqlBytes, err := ioutil.ReadFile("../profile/db.sql")
	require.NoError(t, err)

	_, err = db.Exec(string(sqlBytes))
	require.NoError(t, err)

	return db
}

func prepareEngine(t *testing.T, db *sql.DB, serverKey *ecdsa.PrivateKey) *gin.Engine {
	config := profile.Config{
		Services: profile.Services{
			Log: logrus.New(),
			Db:  db,
		},
		SchemaDir: "../../pkg/profile",
	}

	router := gin.Default()

	mw, err := auth.NewThirdPartyAuthMiddleware(&serverKey.PublicKey, 6000)
	require.NoError(t, err)

	router.Use(mw)
	router.Use(auth.IdExtractorMiddleware)
	profile.Register(&config, router)
	return router
}

func getAddressFromKey(pk *ecdsa.PublicKey) string {
	return hexutil.Encode(crypto.CompressPubkey(pk))
}

func generateAccessToken(serverKey *ecdsa.PrivateKey, ephKey string, duration time.Duration, userID string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"user_id":       userID,
		"ephemeral_key": ephKey,
		"version":       "1.0",
		"exp":           time.Now().Add(time.Second * duration).Unix(),
	})

	return claims.SignedString(serverKey)
}

func TestGetProfile(t *testing.T) {
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	credential, err := ephemeral.GenerateSimpleCredential(36000)
	require.NoError(t, err)
	addr := getAddressFromKey(&credential.EphemeralPrivateKey.PublicKey)

	db := prepareDb(t)
	router := prepareEngine(t, db, serverKey)

	t.Run("Invalid auth data should return 401", func(t *testing.T) {
		sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		accessToken, err := generateAccessToken(sk, addr, 0, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("No profile stored should return 404", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, addr, 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Profile returned", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO profiles VALUES ('user1', $1)`, validProfile)
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, addr, 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var profile map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &profile)
		require.NoError(t, err)

		require.Len(t, profile, 6)
		require.Contains(t, profile, "id")
		require.Contains(t, profile, "email")
		require.Contains(t, profile, "name")
		require.Contains(t, profile, "description")
		require.Contains(t, profile, "age")
		require.Contains(t, profile, "avatar")
	})

	t.Run("Invalid profile in db should return error", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO profiles VALUES ('user1', '{"schemaVersion": 2, "bodyShape": "alien"}')`)
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, addr, 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPostProfile(t *testing.T) {
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	credential, err := ephemeral.GenerateSimpleCredential(36000)
	require.NoError(t, err)
	addr := getAddressFromKey(&credential.EphemeralPrivateKey.PublicKey)

	db := prepareDb(t)
	router := prepareEngine(t, db, serverKey)

	t.Run("Invalid profile should return badRequest", func(t *testing.T) {
		accessToken, err := generateAccessToken(serverKey, addr, 0, "user1")
		require.NoError(t, err)

		_, err = db.Exec("DELETE FROM profiles")
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/api/v1/profile", strings.NewReader(`{"bodyShape": "alien"}`))
		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Profile stored", func(t *testing.T) {
		accessToken, err := generateAccessToken(serverKey, addr, 0, "user1")
		require.NoError(t, err)

		_, err = db.Exec("DELETE FROM profiles")
		require.NoError(t, err)

		reader := strings.NewReader(validProfile)
		req, _ := http.NewRequest("POST", "/api/v1/profile", reader)
		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		fmt.Println(string(w.Body.Bytes()))
		require.Equal(t, http.StatusNoContent, w.Code)

		userID := "user1"
		row := db.QueryRow("SELECT profile FROM profiles WHERE user_id = $1", userID)

		var jsonProfile []byte
		err = row.Scan(&jsonProfile)
		require.NoError(t, err)

		var profile map[string]interface{}
		err = json.Unmarshal(jsonProfile, &profile)
		require.NoError(t, err)

		require.Len(t, profile, 6)
		require.Contains(t, profile, "id")
		require.Contains(t, profile, "email")
		require.Contains(t, profile, "name")
		require.Contains(t, profile, "description")
		require.Contains(t, profile, "age")
		require.Contains(t, profile, "avatar")
	})
}
