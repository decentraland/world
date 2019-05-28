package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/decentraland/world/internal/auth"
	configuration "github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/gindcl"
	"github.com/decentraland/world/internal/profile"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

type profileConfig struct {
	Host      string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
	Port      int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
	ConnStr   string `overwrite-flag:"connStr"   flag-usage:"psql connection string" validate:"required"`
	SchemaDir string `overwrite-flag:"schemaDir" flag-usage:"path to the directory containing json schema files" validate:"required"`
	PublicURL string `overwrite-flag:"publicURL" flag-usage:"Example: http://yourDomain.com" validate:"required"`
	Auth      auth.Configuration
}

func main() {
	log := logrus.New()
	router := gin.Default()
	router.Use(ginlogrus.Logger(log), gin.Recovery())

	var conf profileConfig
	if err := configuration.ReadConfiguration("config/profile/config", &conf); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", conf.ConnStr)
	if err != nil {
		log.Fatal(err)
	}

	router.Use(gindcl.CorsMiddleware())

	authMiddleware, err := auth.NewAuthMiddleware(&conf.Auth, conf.PublicURL)
	if err != nil {
		log.WithError(err).Fatal("error creating auth middleware")
	}

	if authMiddleware != nil {
		router.Use(authMiddleware)
	}

	config := profile.Config{
		Services:  profile.Services{Log: log, Db: db},
		SchemaDir: conf.SchemaDir,
	}

	if err = profile.Register(&config, router); err != nil {
		log.WithError(err).Fatal("unable to start profile service")
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Failed to start server.")
	}
}
