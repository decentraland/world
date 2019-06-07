package main

import (
	"fmt"
	"time"

	"github.com/decentraland/world/internal/commons/metrics"

	configuration "github.com/decentraland/world/internal/commons/config"
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

type auth0Config struct {
	BaseURL string `overwrite-flag:"auth0BaseURL"`
	Domain  string `overwrite-flag:"auth0Domain"`
}

type server struct {
	Host      string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
	Port      int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
	PublicURL string `overwrite-flag:"publicURL" validate:"required"`
}

type identityConf struct {
	LogLevel        string `overwrite-flag:"logLevel"`
	ClientsDataPath string `overwrite-flag:"clientsDataPath"`
	PrivateKeyPath  string `overwrite-flag:"privateKeyPath" validate:"required"`
	JwtDuration     time.Duration
	Auth0           auth0Config
	Server          server
	MetricsConfig   metricsConfig
}

type metricsConfig struct {
	Enabled              bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
	TraceName            string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`
	AnalyticsRateEnabled bool   `overwrite-flag:"rateEnabled" flag-usage:"metrics analytics rate"`
}

func main() {
	l := log.New()
	router := gin.Default()
	router.Use(ginlogrus.Logger(l), gin.Recovery())

	var conf identityConf
	if err := configuration.ReadConfiguration("config/identity/config", &conf); err != nil {
		log.Fatal(err)
	}

	if err := logging.SetLevel(l, conf.LogLevel); err != nil {
		log.WithError(err).Fatal("Fail to start server")
	}

	log.Info("Starting Server...")
	key, err := utils.ReadPrivateKeyFromFile(conf.PrivateKeyPath)
	if err != nil {
		log.WithError(err).Fatal("Fail to start server. Error while reading Private key")
	}

	auth0, err := data.MakeAuth0Service(data.Auth0Config{
		BaseURL: conf.Auth0.BaseURL,
		Domain:  conf.Auth0.BaseURL,
	})
	if err != nil {
		log.WithError(err).Fatal("Fail to initialize Auth0 Client")
	}

	repo, err := repository.NewClientRepository(conf.ClientsDataPath)
	if err != nil {
		log.WithError(err).Fatal("Fail to initialize Client repository")
	}

	config := api.Config{
		Auth0Service:     auth0,
		Key:              key,
		ClientRepository: repo,
		ServerURL:        conf.Server.PublicURL,
		JWTDuration:      conf.JwtDuration,
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

	if err := api.InitAPI(router, &config); err != nil {
		log.WithError(err).Fatal("Fail to initialize routes")
	}

	web.SiteContent(router, repo, conf.Server.PublicURL, conf.Auth0.Domain)

	addr := fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Fail to start server.")
	}
}
