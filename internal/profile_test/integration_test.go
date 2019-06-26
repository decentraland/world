// +build integration

package profile_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
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
	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/profile"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var validProfile = `{
	"name": "name",
	"description": "description",
	"created_at": 11111111,
	"avatar": {
	  "version": 2,
  
	  "skin": {
		"color": {
		  "r": 0.1,
		  "g": 0.1,
		  "b": 0.1
		}
	  },
  
	  "hair": {
		"color": {
		  "r": 0.1,
		  "g": 0.1,
		  "b": 0.1
		}
	  },
  
	  "eyes": {
		"color": {
		  "r": 0.1,
		  "g": 0.1,
		  "b": 0.1
		}
	  },
  
	  "bodyShape": "dcl://base-avatars/BaseMale",
  
	  "wearables": [
		"dcl://base-avatars/m_blue_jacket",
		"dcl://base-avatars/black_sun_glasses"
	  ]
	}
  }
  `

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

	mw, err := auth.NewThirdPartyAuthMiddleware(&serverKey.PublicKey, &auth.MiddlewareConfiguration{
		Log: logrus.New(),
	})
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

	config := ephemeral.EphemeralKeyConfig{}
	ephemeralKey, err := ephemeral.NewEphemeralKey(&config)
	require.NoError(t, err)
	addr := getAddressFromKey(ephemeralKey.PublicKey())

	db := prepareDb(t)
	router := prepareEngine(t, db, serverKey)

	t.Run("Invalid auth data should return 401", func(t *testing.T) {
		sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		accessToken, err := generateAccessToken(sk, addr, 0, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)

		accessToken, err = generateAccessToken(sk, addr, 0, "user1")
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/api/v1/profile/user1", nil)

		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("No profile stored should return 404", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, addr, 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)

		accessToken, err = generateAccessToken(serverKey, addr, 6000, "user1")
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/api/v1/profile/user2", nil)

		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

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

		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var profileResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &profileResponse)
		require.NoError(t, err)

		require.NotNil(t, profileResponse["version"])
		require.Equal(t, "user1", profileResponse["user_id"])

		profile := profileResponse["profile"]

		require.Len(t, profile, 4)
		require.Contains(t, profile, "name")
		require.Contains(t, profile, "description")
		require.Contains(t, profile, "created_at")
		require.Contains(t, profile, "avatar")
	})

	t.Run("Other Users Profile returned", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO profiles VALUES ('user1', $1)`, validProfile)
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, addr, 6000, "user2")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile/user1", nil)

		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var profileResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &profileResponse)
		require.NoError(t, err)

		require.NotNil(t, profileResponse["version"])
		require.Equal(t, "user1", profileResponse["user_id"])

		profile := profileResponse["profile"]

		require.Len(t, profile, 4)
		require.Contains(t, profile, "name")
		require.Contains(t, profile, "description")
		require.Contains(t, profile, "created_at")
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

		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPostProfile(t *testing.T) {
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	config := ephemeral.EphemeralKeyConfig{}
	ephemeralKey, err := ephemeral.NewEphemeralKey(&config)
	require.NoError(t, err)
	addr := getAddressFromKey(ephemeralKey.PublicKey())

	db := prepareDb(t)
	router := prepareEngine(t, db, serverKey)

	t.Run("Invalid profile should return badRequest", func(t *testing.T) {
		accessToken, err := generateAccessToken(serverKey, addr, 0, "user1")
		require.NoError(t, err)

		_, err = db.Exec("DELETE FROM profiles")
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/api/v1/profile", strings.NewReader(`{"bodyShape": "alien"}`))
		err = ephemeralKey.AddRequestHeaders(req, accessToken)
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
		err = ephemeralKey.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		userID := "user1"
		row := db.QueryRow("SELECT profile, version FROM profiles WHERE user_id = $1", userID)

		var jsonProfile []byte
		var version int64
		err = row.Scan(&jsonProfile, &version)
		require.NoError(t, err)

		var profile map[string]interface{}
		err = json.Unmarshal(jsonProfile, &profile)
		require.NoError(t, err)

		var responseVersion float64 = response["version"].(float64)

		require.Equal(t, version, int64(responseVersion))
		require.Equal(t, userID, response["user_id"])
		require.Len(t, profile, 4)
		require.Contains(t, profile, "name")
		require.Contains(t, profile, "description")
		require.Contains(t, profile, "created_at")
		require.Contains(t, profile, "avatar")

		reader = strings.NewReader(validProfile)
		updateReq, _ := http.NewRequest("POST", "/api/v1/profile", reader)
		accessToken, err = generateAccessToken(serverKey, addr, 0, "user1")
		require.NoError(t, err)

		err = ephemeralKey.AddRequestHeaders(updateReq, accessToken)
		require.NoError(t, err)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, updateReq)
		require.Equal(t, http.StatusOK, w.Code)

		var updateResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
		require.NoError(t, err)

		var updateVersion float64 = updateResponse["version"].(float64)
		require.True(t, updateVersion > responseVersion)
	})
}
