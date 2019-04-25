// +build integration

package profile_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"

	"database/sql"

	"github.com/decentraland/world/internal/profile"
	"github.com/decentraland/auth-go/pkg/keys"
	"github.com/decentraland/world/internal/auth"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/decentraland/auth-go/pkg/ephemeral"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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
	pubk, err := keys.PemEncodePublicKey(&serverKey.PublicKey)
	require.NoError(t, err)

	authConfig := &auth.Configuration{Mode: auth.AuthThirdParty, AuthKey: pubk, RequestTTL: 6000}
	mw, err := auth.NewAuthMiddleware(authConfig)
	require.NoError(t, err)

	router.Use(mw)
	router.Use(auth.IdExtractorMiddleware)
	profile.Register(&config, router)
	return router
}

func TestGetProfile(t *testing.T) {
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	credential, err := ephemeral.GenerateSimpleCredential(36000)
	require.NoError(t, err)

	db := prepareDb(t)
	router := prepareEngine(t, db, serverKey)

	t.Run("Invalid auth data should return 401", func(t *testing.T) {
		sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		accessToken, err := generateAccessToken(sk, getAddressFromKey(&credential.EphemeralPrivateKey.PublicKey), 0, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("No profile stored should return 404", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, getAddressFromKey(&credential.EphemeralPrivateKey.PublicKey), 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/profile", nil)

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

		accessToken, err := generateAccessToken(serverKey, getAddressFromKey(&credential.EphemeralPrivateKey.PublicKey), 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		log.Println(response)
		require.Len(t, response, 2)
		require.Contains(t, response, "schemaVersion")
		require.Contains(t, response, "bodyShape")
		require.Equal(t, response["schemaVersion"], 1.0)
		require.Equal(t, response["bodyShape"], "girl")
	})

	t.Run("Invalid profile in db should return error", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO profiles VALUES ('user1', 1, '{"schemaVersion": 2, "bodyType": "girl"}')`)
		require.NoError(t, err)

		accessToken, err := generateAccessToken(serverKey, getAddressFromKey(&credential.EphemeralPrivateKey.PublicKey), 6000, "user1")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/profile", nil)

		err = credential.AddRequestHeaders(req, accessToken)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
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