package main

import (
	"database/sql"
	"fmt"
	"github.com/decentraland/world/internal/auth"
	configuration "github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/profile"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
)

type ProfileConfig struct {
	Host      string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
	Port      int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
	ConnStr   string `overwrite-flag:"connStr"   flag-usage:"psql connection string" validate:"required"`
	SchemaDir string `overwrite-flag:"schemaDir" flag-usage:"path to the directory containing json schema files" validate:"required"`
	Auth      auth.Configuration
}

func main() {

	log := logrus.New()
	router := gin.Default()
	router.Use(ginlogrus.Logger(log), gin.Recovery())

	var conf ProfileConfig
	if err := configuration.ReadConfiguration("config/profile/config", &conf); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", conf.ConnStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := setupAuthentication(router, &conf.Auth); err != nil {
		log.Fatal(err)
	}

	config := profile.Config{
		Services:  profile.Services{Log: log, Db: db},
		SchemaDir: conf.SchemaDir,
	}
	err = profile.Register(&config, router)
	if err != nil {
		log.WithError(err).Fatal("unable to start profile service")
	}

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Fail to start server.")
	}
}

func setupAuthentication(r *gin.Engine, authConfig *auth.Configuration) error {
	authMiddleware, err := auth.NewAuthMiddleware(authConfig)
	if err != nil {
		return err
	}
	r.Use(authMiddleware)
	r.Use(auth.IdExtractorMiddleware)
	return nil
}
