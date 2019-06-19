package main

import (
	"database/sql"
	"fmt"

	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/utils"
	"github.com/decentraland/world/internal/profile"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	ginlogrus "github.com/toorop/gin-logrus"
)

type rootConfig struct {
	IdentityURL string `overwrite-flag:"authURL" validate:"required"`
	Profile     struct {
		PublicURL string `overwrite-flag:"publicURL" flag-usage:"Example: http://yourDomain.com" validate:"required"`
		Host      string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
		Port      int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
		LogLevel  string `overwrite-flag:"logLevel"`
		ConnStr   string `overwrite-flag:"connStr"   flag-usage:"psql connection string" validate:"required"`
		SchemaDir string `overwrite-flag:"schemaDir" flag-usage:"path to the directory containing json schema files" validate:"required"`
		AuthTTL   int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
		Metrics   struct {
			Enabled   bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
			TraceName string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`

			AnalyticsRateEnabled bool `overwrite-flag:"rateEnabled" flag-usage:"metrics analytics rate"`
		}
	}
}

func main() {
	log := logging.New()
	router := gin.Default()
	router.Use(ginlogrus.Logger(log), gin.Recovery())

	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	if err := logging.SetLevel(log, conf.Profile.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	db, err := sql.Open("postgres", conf.Profile.ConnStr)
	if err != nil {
		log.Fatal(err)
	}

	if conf.Profile.Metrics.Enabled {
		metricsConfig := &metrics.HttpMetricsConfig{
			TraceName:            conf.Profile.Metrics.TraceName,
			AnalyticsRateEnabled: conf.Profile.Metrics.AnalyticsRateEnabled,
		}
		if err := metrics.EnableRouterMetrics(metricsConfig, router); err != nil {
			log.WithError(err).Fatal("Unable to start metrics")
		}
		defer metrics.StopMetrics()
	}

	router.Use(utils.CorsMiddleware())

	authMiddleware, err := auth.NewAuthMiddleware(&auth.MiddlewareConfiguration{
		PublicURL:   conf.Profile.PublicURL,
		IdentityURL: conf.IdentityURL,
		RequestTTL:  conf.Profile.AuthTTL,
		Log:         log,
	})
	if err != nil {
		log.WithError(err).Fatal("error creating auth middleware")
	}

	config := profile.Config{
		Services:       profile.Services{Log: log, Db: db},
		SchemaDir:      conf.Profile.SchemaDir,
		AuthMiddleware: authMiddleware,
		IdentityURL:    conf.IdentityURL,
	}

	if err = profile.Register(&config, router); err != nil {
		log.WithError(err).Fatal("unable to start profile service")
	}

	addr := fmt.Sprintf("%s:%d", conf.Profile.Host, conf.Profile.Port)
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Failed to start server.")
	}
}
