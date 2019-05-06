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
	"github.com/decentraland/world/internal/auth"
	"github.com/decentraland/world/internal/profile"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

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

func generateAccessToken(serverKey *ecdsa.PrivateKey, ephKey string, duration time.Duration, userId string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"user_id":       userId,
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
		_, err = db.Exec(`INSERT INTO profiles VALUES ('user1', 1, '{"schemaVersion": 1, "bodyShape": "girl"}')`)
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, addr, 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/v1/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Len(t, response, 2)
		require.Contains(t, response, "schemaVersion")
		require.Contains(t, response, "bodyShape")
		require.Equal(t, response["schemaVersion"], 1.0)
		require.Equal(t, response["bodyShape"], "girl")
	})

	t.Run("Invalid profile in db should return error", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO profiles VALUES ('user1', 1, '{"schemaVersion": 2, "bodyShape": "alien"}')`)
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

		reader := strings.NewReader(`{"schemaVersion": 1, "bodyShape": "girl"}`)
		req, _ := http.NewRequest("POST", "/api/v1/profile", reader)
		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)

		userId := "user1"
		row := db.QueryRow("SELECT profile FROM profiles WHERE user_id = $1", userId)

		var jsonProfile []byte
		err = row.Scan(&jsonProfile)
		require.NoError(t, err)

		var profile map[string]interface{}
		err = json.Unmarshal(jsonProfile, &profile)
		require.NoError(t, err)

		require.Len(t, profile, 2)
		require.Contains(t, profile, "schemaVersion")
		require.Contains(t, profile, "bodyShape")
		require.Equal(t, profile["schemaVersion"], 1.0)
		require.Equal(t, profile["bodyShape"], "girl")
	})
}
