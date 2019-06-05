package main

import (
	"fmt"
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

type Auth0Config struct {
	BaseURL string `overwrite-flag:"auth0BaseURL"`
	Domain  string `overwrite-flag:"auth0Domain"`
}

type identityConf struct {
	Auth0           Auth0Config
	LogLevel        string `overwrite-flag:"logLevel"`
	JwtDuration     time.Duration
	ClientsDataPath string `overwrite-flag:"clientsDataPath"`
	Server          Server
	PrivateKeyPath  string `overwrite-flag:"privateKeyPath" validate:"required"`
}

type Server struct {
	Host      string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
	Port      int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
	PublicURL string `overwrite-flag:"publicURL" validate:"required"`
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

	if err := api.InitApi(router, &config); err != nil {
		log.WithError(err).Fatal("Fail to initialize routes")
	}

	web.SiteContent(router, repo, conf.Server.PublicURL, conf.Auth0.Domain)

	addr := fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)
	if err := router.Run(addr); err != nil {
		log.WithError(err).Fatal("Fail to start server.")
	}
}
