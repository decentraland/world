package main

import (
	"strconv"
	"time"

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

type identityConf struct {
	Auth0           data.Auth0Config
	LogLevel        string        `overwrite-env:"LOG_LEVEL"`
	JwtDuration     time.Duration `overwrite-env:"JWT_DURATION"`
	ClientsDataPath string        `overwrite-env:"CLIENTS_DATA_PATH"`
	Server          Server
	PrivateKeyPath  string `overwrite-env:"" validate:"required"`
}

type Server struct {
	Port int    `overwrite-env:"PORT" validate:"required"`
	URL  string `overwrite-env:"URL" validate:"required"`
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

	auth0, err := data.MakeAuth0Service(conf.Auth0)
	if err != nil {
		log.WithError(err).Fatal("Fail to initialize Auth0 Client")
	}

	repo, err := repository.NewClientRepository(conf.ClientsDataPath)
	if err != nil {
		log.WithError(err).Fatal("Fail to initialize Client repository")
	}

	if err := api.InitApi(auth0, key, router, repo, conf.Server.URL, conf.JwtDuration); err != nil {
		log.WithError(err).Fatal("Fail to initialize routes")
	}

	web.SiteContent(router, repo, conf.Server.URL, conf.Auth0.Domain)

	if err := router.Run(":" + strconv.Itoa(conf.Server.Port)); err != nil {
		log.WithError(err).Fatal("Fail to start server.")
	}
}
