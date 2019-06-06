package main

import (
	"database/sql"
	"fmt"
	"github.com/decentraland/world/internal/commons/metrics"

	"github.com/decentraland/world/internal/commons/auth"
	configuration "github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/gindcl"
	"github.com/decentraland/world/internal/profile"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
)

type authConfiguration struct {
	AuthServerURL string `overwrite-flag:"authURL" flag-usage:"path to the file containing the auth-service public key"`
	RequestTTL    int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
}

type profileConfig struct {
	Host          string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
	Port          int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
	ConnStr       string `overwrite-flag:"connStr"   flag-usage:"psql connection string" validate:"required"`
	SchemaDir     string `overwrite-flag:"schemaDir" flag-usage:"path to the directory containing json schema files" validate:"required"`
	PublicURL     string `overwrite-flag:"publicURL" flag-usage:"Example: http://yourDomain.com" validate:"required"`
	Auth          authConfiguration
	MetricsConfig metricsConfig
}

type metricsConfig struct {
	Enabled              bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
	TraceName            string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`
	AnalyticsRateEnabled bool   `overwrite-flag:"rateEnabled" flag-usage:"metrics analytics rate"`
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

	if conf.MetricsConfig.Enabled {
		metricsConfig := &metrics.HttpMetricsConfig{
			TraceName:            conf.MetricsConfig.TraceName,
			AnalyticsRateEnabled: conf.MetricsConfig.AnalyticsRateEnabled,
		}
		if err := metrics.EnableRouterMetrics(metricsConfig, router); err != nil {
			log.WithError(err).Fatal("Unable to start metrics")
		}
		defer metrics.StopMetrics()
	}

	router.Use(gindcl.CorsMiddleware())

	authMiddleware, err := auth.NewAuthMiddleware(&auth.MiddlewareConfiguration{
		AuthServerURL: conf.Auth.AuthServerURL,
		RequestTTL:    conf.Auth.RequestTTL,
	}, conf.PublicURL)
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

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Failed to start server.")
	}
}
