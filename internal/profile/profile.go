package profile

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/utils"
	"github.com/decentraland/world/internal/commons/version"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

// Services represents profile service dependencies
type Services struct {
	Log *logrus.Logger
	Db  *sql.DB
}

// Config represents profile service config
type Config struct {
	SchemaDir      string
	Services       Services
	AuthMiddleware func(ctx *gin.Context)
	IdentityURL    string
}

type createProfileResponse struct {
	Version int64  `json:"version"`
	UserID  string `json:"user_id"`
}

type getProfileResponse struct {
	Version int64                  `json:"version"`
	Profile map[string]interface{} `json:"profile"`
	UserID  string                 `json:"user_id"`
}

// Register register api routes
func Register(config *Config, router gin.IRouter) error {
	services := config.Services
	log := services.Log
	db := services.Db

	sl := gojsonschema.NewSchemaLoader()
	profileSchemaPath := path.Join(config.SchemaDir, "v1-profile.json")
	schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", profileSchemaPath))
	schema, err := sl.Compile(schemaLoader)
	if err != nil {
		log.WithError(err).Error("error loading schema")
		return err
	}

	api := router.Group("/api")
	v1 := api.Group("/v1")

	version.RegisterVersionEndpoint(v1)

	profile := v1.Group("/profile")

	if config.AuthMiddleware != nil {
		profile.Use(config.AuthMiddleware)
	}

	profile.Use(auth.IdExtractorMiddleware)

	internalError := gin.H{"error": "Internal error, please retry later"}

	profile.GET("", getUser(db, log, schema, func(c *gin.Context) string {
		return c.GetString("userId")
	}))

	profile.GET("/:id", getUser(db, log, schema, func(c *gin.Context) string {
		return c.Param("id")
	}))

	profile.POST("", func(c *gin.Context) {
		userID := c.GetString("userId")

		data, err := c.GetRawData()
		if err != nil {
			log.WithError(err).Error("get raw data failed")
			c.Error(err)
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		documentLoader := gojsonschema.NewBytesLoader(data)
		result, err := schema.Validate(documentLoader)
		if err != nil {
			log.WithError(err).Error("json validation failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if !result.Valid() {
			errors := make([]string, 0, len(result.Errors()))
			for _, desc := range result.Errors() {
				errors = append(errors, desc.String())
			}

			c.JSON(400, gin.H{"errors": errors})
			return
		}

		var version int64
		err = db.QueryRow(`
INSERT INTO profiles (user_id, profile) VALUES($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET profile = $2
RETURNING version`,
			userID, data).Scan(&version)

		if err != nil {
			log.WithError(err).Error("insert failed")
			c.Error(err)
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		c.JSON(http.StatusOK, &createProfileResponse{UserID: userID, Version: version})
	})

	v1.OPTIONS("/profile", utils.PrefligthChecksMiddleware("POST, GET", utils.AllHeaders))
	v1.OPTIONS("/profile/:id", utils.PrefligthChecksMiddleware("GET", utils.AllHeaders))

	v1.GET("/status", utils.ServiceStatusHandler(func() map[string]string {
		errors := map[string]string{}
		pingError := db.Ping()
		if pingError != nil {
			log.WithError(pingError).Error("failed to connect with db")
			errors["database"] = "failed to reach db"
		}

		return errors
	}))

	return nil
}

func getUser(db *sql.DB, log *logrus.Logger, schema *gojsonschema.Schema, userIdProvider func(c *gin.Context) string) func(c *gin.Context) {
	return func(c *gin.Context) {
		internalError := gin.H{"error": "Internal error, please retry later"}

		userID := userIdProvider(c)

		row := db.QueryRow("SELECT profile, version FROM profiles WHERE user_id = $1", userID)

		var jsonProfile []byte
		var version int64
		err := row.Scan(&jsonProfile, &version)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		if err != nil {
			log.WithError(err).Error("query profile failed")
			c.Error(fmt.Errorf("query profile failed: %s", err.Error()))
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		profile := make(map[string]interface{})

		// var profile definitions.Profile
		if err := json.Unmarshal(jsonProfile, &profile); err != nil {
			log.WithError(err).Error("json unmarshalling failed")
			c.Error(err)
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		documentLoader := gojsonschema.NewGoLoader(profile)
		result, err := schema.Validate(documentLoader)
		if err != nil {
			log.WithError(err).Error("json validation failed")
			c.Error(err)
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		if !result.Valid() {
			errs := make([]string, 0, len(result.Errors()))
			for _, desc := range result.Errors() {
				errs = append(errs, desc.String())
			}
			logErr := fmt.Errorf("invalid response from get profile: %s", strings.Join(errs, ", "))
			log.WithError(logErr)
			c.Error(logErr)
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		c.JSON(200, &getProfileResponse{Version: version, Profile: profile, UserID: userID})
	}
}
