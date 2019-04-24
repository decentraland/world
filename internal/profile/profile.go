package profile

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	definitions "github.com/decentraland/world/pkg/profile"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

type ProfileV1 = definitions.ProfileV1

type Services struct {
	Log *logrus.Logger
	Db  *sql.DB
}

func Register(services Services, router gin.IRouter) error {
	log := services.Log
	db := services.Db

	sl := gojsonschema.NewSchemaLoader()

	path, err := filepath.Abs("pkg/profile/v1-profile.json")
	if err != nil {
		log.WithError(err).Error("error getting profile schema abspath")
		return err
	}

	schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", path))
	schema, err := sl.Compile(schemaLoader)
	if err != nil {
		log.WithError(err).Error("error loading schema")
		return err
	}

	internalError := gin.H{"error": "internal error, please retry later"}
	router.GET("/profile", func(c *gin.Context) {
		// TODO: validate access token first
		userId := "user1"

		row := db.QueryRow("SELECT profile FROM profiles WHERE user_id = $1", userId)

		var jsonProfile []byte
		err := row.Scan(&jsonProfile)
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{})
			return
		}

		if err != nil {
			log.WithError(err).Error("query profile failed")
			c.JSON(500, internalError)
			return
		}

		var profile ProfileV1
		if err := json.Unmarshal(jsonProfile, &profile); err != nil {
			log.WithError(err).Error("json unmarshalling failed")
			c.JSON(500, internalError)
			return
		}

		documentLoader := gojsonschema.NewGoLoader(profile)
		result, err := schema.Validate(documentLoader)
		if err != nil {
			log.WithError(err).Error("json validation failed")
			c.JSON(500, internalError)
			return
		}

		if !result.Valid() {
			errors := make([]string, 0, len(result.Errors()))
			for _, desc := range result.Errors() {
				errors = append(errors, desc.String())
			}
			log.Errorf("invalid response from get profile: %s", strings.Join(errors, ", "))
			c.JSON(500, internalError)
			return
		}

		c.JSON(200, profile)
	})

	router.POST("/profile", func(c *gin.Context) {
		// TODO: validate access token first
		userId := "user1"

		data, err := c.GetRawData()
		if err != nil {
			log.WithError(err).Error("get raw data failed")
			c.JSON(500, gin.H{"error": err})
			return
		}

		documentLoader := gojsonschema.NewBytesLoader(data)
		result, err := schema.Validate(documentLoader)
		if err != nil {
			log.WithError(err).Error("json validation failed")
			c.JSON(500, internalError)
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

		var profile ProfileV1
		if err := json.Unmarshal(data, &profile); err != nil {
			log.WithError(err).Error("json unmarshalling failed")
			c.JSON(500, internalError)
			return
		}

		_, err = db.Exec(`
INSERT INTO profiles (user_id, schema_version, profile) VALUES($1, $2, $3)
ON CONFLICT (user_id)
DO UPDATE SET schema_version = $2, profile = $3`,
			userId, profile.SchemaVersion, data)

		if err != nil {
			log.WithError(err).Error("insert failed")
			c.JSON(500, internalError)
			return
		}

		c.JSON(204, gin.H{})
	})

	return nil
}
