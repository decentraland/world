package profile

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/decentraland/world/internal/commons/version"
	"net/http"
	"path"
	"strings"

	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/utils"

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

// StatusResponse contains all the service dependencies status
type statusResponse struct {
	Ok     bool              `json:"ok"`
	Errors map[string]string `json:"errors"`
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
	profile.GET("", func(c *gin.Context) {
		userID := c.GetString("userId")

		row := db.QueryRow("SELECT profile FROM profiles WHERE user_id = $1", userID)

		var jsonProfile []byte
		err := row.Scan(&jsonProfile)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		if err != nil {
			log.WithError(err).Error("query profile failed")
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		profile := make(map[string]interface{})

		// var profile definitions.Profile
		if err := json.Unmarshal(jsonProfile, &profile); err != nil {
			log.WithError(err).Error("json unmarshalling failed")
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		documentLoader := gojsonschema.NewGoLoader(profile)
		result, err := schema.Validate(documentLoader)
		if err != nil {
			log.WithError(err).Error("json validation failed")
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		if !result.Valid() {
			errors := make([]string, 0, len(result.Errors()))
			for _, desc := range result.Errors() {
				errors = append(errors, desc.String())
			}
			log.Errorf("invalid response from get profile: %s", strings.Join(errors, ", "))
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		c.JSON(200, profile)
	})

	profile.POST("", func(c *gin.Context) {
		userID := c.GetString("userId")

		data, err := c.GetRawData()
		if err != nil {
			log.WithError(err).Error("get raw data failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
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

		_, err = db.Exec(`
INSERT INTO profiles (user_id, profile) VALUES($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET profile = $2`,
			userID, data)

		if err != nil {
			log.WithError(err).Error("insert failed")
			c.JSON(http.StatusInternalServerError, internalError)
			return
		}

		c.JSON(http.StatusNoContent, gin.H{})
	})

	v1.OPTIONS("/profile", utils.PrefligthChecksMiddleware("POST, GET", "*"))
	
	v1.GET("/status", func(c *gin.Context) {
		errors := map[string]string{}
		pingError := db.Ping()
		if pingError != nil {
			log.WithError(pingError).Error("failed to connect with db")
			errors["database"] = "failed to reach db"
		}

		statusResponse := &statusResponse{Ok: len(errors) == 0, Errors: errors}

		if statusResponse.Ok {
			c.JSON(http.StatusOK, statusResponse)
		} else {
			c.JSON(http.StatusServiceUnavailable, statusResponse)
		}

	})

	return nil
}
