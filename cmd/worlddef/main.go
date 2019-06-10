package main

import (
	"fmt"

	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/utils"
	"github.com/decentraland/world/internal/worlddef"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

type rootConfig struct {
	IdentityURL     string `overwrite-flag:"identityURL" validate:"required"`
	CoordinatorURL  string `overwrite-flag:"coordinatorURL" validate:"required"`
	ProfileURL      string `overwrite-flag:"profileURL" validate:"required"`
	WorldDefinition struct {
		WorldName        string `overwrite-flag:"name"      flag-usage:"host name" validate:"required"`
		WorldDescription string `overwrite-flag:"description"      flag-usage:"host name" validate:"required"`
		Host             string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
		Port             int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
		LogLevel         string `overwrite-flag:"logLevel"`
		Metrics          struct {
			Enabled   bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
			TraceName string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`

			AnalyticsRateEnabled bool `overwrite-flag:"rateEnabled" flag-usage:"metrics analytics rate"`
		}
	}
}

func main() {
	log := logrus.New()
	router := gin.Default()
	router.Use(ginlogrus.Logger(log), gin.Recovery())

	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	if err := logging.SetLevel(log, conf.WorldDefinition.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	if conf.WorldDefinition.Metrics.Enabled {
		metricsConfig := &metrics.HttpMetricsConfig{
			TraceName:            conf.WorldDefinition.Metrics.TraceName,
			AnalyticsRateEnabled: conf.WorldDefinition.Metrics.AnalyticsRateEnabled,
		}
		if err := metrics.EnableRouterMetrics(metricsConfig, router); err != nil {
			log.WithError(err).Fatal("Unable to start metrics")
		}
		defer metrics.StopMetrics()
	}

	router.Use(utils.CorsMiddleware())

	config := worlddef.Config{
		WorldName:        conf.WorldDefinition.WorldName,
		WorldDescription: conf.WorldDefinition.WorldDescription,
		CoordinatorURL:   conf.CoordinatorURL,
		IdentityURL:      conf.IdentityURL,
		ProfileURL:       conf.ProfileURL,
		Log:              log,
	}

	if err := worlddef.Register(&config, router); err != nil {
		log.WithError(err).Fatal("unable to start world definition service")
	}

	addr := fmt.Sprintf("%s:%d", conf.WorldDefinition.Host, conf.WorldDefinition.Port)
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Failed to start server.")
	}
}
