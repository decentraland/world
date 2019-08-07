package main

import (
	"database/sql"
	"fmt"
	"time"

	"net/http"

	_ "net/http/pprof"

	brokerAuth "github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/version"
	"github.com/decentraland/world/internal/commserver"
	_ "github.com/lib/pq"
	logrus "github.com/sirupsen/logrus"
)

type rootConfig struct {
	LogJSONDisabled bool   `overwrite-flag:"JSONDisabled"`
	IdentityURL     string `overwrite-flag:"authURL" validate:"required"`
	CoordinatorURL  string `overwrite-flag:"coordinatorURL" flag-usage:"coordinator url" validate:"required"`
	CommServer      struct {
		LogLevel     string `overwrite-flag:"logLevel"`
		ServerSecret string `overwrite-flag:"serverSecret"`
		AuthTTL      int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
		AuthEnabled  bool   `overwrite-flag:"authEnabled"`
		Cluster      string `overwrite-flag:"cluster"`
		Metrics      struct {
			Enabled         bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
			TraceName       string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`
			StatsDBHost     string `overwrite-flag:"statsDBHost"`
			StatsDBName     string `overwrite-flag:"statsDBName"`
			StatsDBPort     int    `overwrite-flag:"statsDBPort"`
			StatsDBUser     string `overwrite-flag:"statsDBUser"`
			StatsDBPassword string `overwrite-flag:"statsDBPassword"`
		}
	}
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		logrus.Fatal(err)
	}

	loggerConfig := logging.LoggerConfig{JSONDisabled: conf.LogJSONDisabled}
	log := logging.New(&loggerConfig)

	if err := logging.SetLevel(log, conf.CommServer.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}
	defer logging.LogPanic()

	var authenticator brokerAuth.ServerAuthenticator
	var err error

	if conf.CommServer.AuthEnabled {
		authenticator, err = auth.MakeAuthenticator(&auth.AuthenticatorConfig{
			IdentityURL: conf.IdentityURL,
			Secret:      conf.CommServer.ServerSecret,
			RequestTTL:  conf.CommServer.AuthTTL,
			Log:         log,
		})
		if err != nil {
			log.WithError(err).Fatal("cannot build authenticator")
		}
	} else {
		authenticator = &brokerAuth.NoopAuthenticator{}
	}

	config := commserver.Config{
		Auth: authenticator,
		Log:  log,
		ICEServers: []commserver.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		CoordinatorURL:         conf.CoordinatorURL,
		ExitOnCoordinatorClose: true,
		ReportPeriod:           10 * time.Second,
	}

	if conf.CommServer.Metrics.Enabled {
		traceName := conf.CommServer.Metrics.TraceName

		client, err := metrics.NewClient(traceName, log)
		if err != nil {
			log.WithError(err).Fatal("cannot start metrics agent")
		}
		defer client.Close()

		psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			conf.CommServer.Metrics.StatsDBHost,
			conf.CommServer.Metrics.StatsDBPort,
			conf.CommServer.Metrics.StatsDBUser,
			conf.CommServer.Metrics.StatsDBPassword,
			conf.CommServer.Metrics.StatsDBName)
		db, err := sql.Open("postgres", psqlConn)
		if err != nil {
			log.WithError(err).Fatal("cannot open postgresql connection")
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.WithError(err).Fatal("cannot open postgresql connection")
		}

		reporter := commserver.NewReporter(&commserver.ReporterConfig{
			LongReportPeriod: 10 * time.Minute,
			Client:           client,
			DB:               db,
			TraceName:        traceName,
			Log:              log,
			Cluster:          conf.CommServer.Cluster,
		})

		config.Reporter = reporter.Report
	} else {
		config.Reporter = func(stats commserver.Stats) {
			log.WithFields(logrus.Fields{
				"log_type":    "report",
				"peer count":  len(stats.Peers),
				"topic count": stats.TopicCount,
			}).Info("report")
		}
	}

	state, err := commserver.MakeState(&config)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting communication server - version: %s", version.Version())

	go func() {
		addr := fmt.Sprintf("0.0.0.0:9081")
		log.Info("Starting profiler at ", addr)
		log.Debug(http.ListenAndServe(addr, nil))
	}()

	if err := commserver.ConnectCoordinator(state); err != nil {
		log.Fatal("connect coordinator failure ", err)
	}

	go commserver.ProcessMessagesQueue(state)
	commserver.Process(state)
}
