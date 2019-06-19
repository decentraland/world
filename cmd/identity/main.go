package main

import (
	"fmt"
	"time"

	"github.com/decentraland/world/internal/commons/metrics"

	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/utils"
	"github.com/decentraland/world/internal/identity/api"
	"github.com/decentraland/world/internal/identity/data"
	"github.com/decentraland/world/internal/identity/repository"
	"github.com/decentraland/world/internal/identity/web"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

type rootConfig struct {
	Auth0 struct {
		Domain string `overwrite-flag:"auth0Domain"`
	}
	Identity struct {
		PublicURL       string `overwrite-flag:"publicURL" validate:"required"`
		Host            string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
		Port            int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
		LogLevel        string `overwrite-flag:"logLevel"`
		ClientsDataPath string `overwrite-flag:"clientsDataPath"`
		PrivateKeyPath  string `overwrite-flag:"privateKeyPath" validate:"required"`
		JwtDuration     time.Duration
		Metrics         struct {
			Enabled   bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
			TraceName string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`

			AnalyticsRateEnabled bool `overwrite-flag:"rateEnabled" flag-usage:"metrics analytics rate"`
		}
	}
}

func main() {
	l := logging.New()
	router := gin.Default()
	router.Use(ginlogrus.Logger(l), gin.Recovery())

	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		log.Fatal(err)
	}

	if err := logging.SetLevel(l, conf.Identity.LogLevel); err != nil {
		log.WithError(err).Fatal("Fail to start server")
	}

	log.Info("Starting Server...")
	key, err := utils.ReadPrivateKeyFromFile(conf.Identity.PrivateKeyPath)
	if err != nil {
		log.WithError(err).Fatal("Fail to start server. Error while reading Private key")
	}

	auth0, err := data.MakeAuth0Service(data.Auth0Config{
		Domain: conf.Auth0.Domain,
	})
	if err != nil {
		log.WithError(err).Fatal("Fail to initialize Auth0 Client")
	}

	repo, err := repository.NewClientRepository(conf.Identity.ClientsDataPath)
	if err != nil {
		log.WithError(err).Fatal("Fail to initialize Client repository")
	}

	config := api.Config{
		Auth0Service:     auth0,
		Key:              key,
		ClientRepository: repo,
		ServerURL:        conf.Identity.PublicURL,
		JWTDuration:      conf.Identity.JwtDuration,
	}

	if conf.Identity.Metrics.Enabled {
		metricsConfig := &metrics.HttpMetricsConfig{
			TraceName:            conf.Identity.Metrics.TraceName,
			AnalyticsRateEnabled: conf.Identity.Metrics.AnalyticsRateEnabled,
		}
		if err := metrics.EnableRouterMetrics(metricsConfig, router); err != nil {
			log.WithError(err).Fatal("Unable to start metrics")
		}
		defer metrics.StopMetrics()
	}

	if err := api.InitAPI(router, &config); err != nil {
		log.WithError(err).Fatal("Fail to initialize routes")
	}

	web.SiteContent(router, repo, conf.Identity.PublicURL, conf.Auth0.Domain)

	addr := fmt.Sprintf("%s:%d", conf.Identity.Host, conf.Identity.Port)
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Fail to start server.")
	}
}
