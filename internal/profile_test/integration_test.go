// +build integration

package profile_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"database/sql"

	"github.com/decentraland/world/internal/profile"
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

func prepareEngine(t *testing.T, db *sql.DB) *gin.Engine {
	config := profile.Config{
		Services: profile.Services{
			Log: logrus.New(),
			Db:  db,
		},
		SchemaDir: "../../pkg/profile",
	}

	router := gin.Default()
	profile.Register(&config, router)
	return router
}

func TestGetProfile(t *testing.T) {
	db := prepareDb(t)
	router := prepareEngine(t, db)

	t.Run("No profile stored should return 404", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Profile returned", func(t *testing.T) {
		_, err := db.Exec("DELETE FROM profiles")
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO profiles VALUES ('user1', 1, '{"schemaVersion": 1, "bodyShape": "girl"}')`)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/profile", nil)
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

		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
