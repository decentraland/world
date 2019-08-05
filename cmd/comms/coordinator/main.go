package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/decentraland/webrtc-broker/pkg/coordinator"

	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/version"

	logrus "github.com/sirupsen/logrus"
)

type rootConfig struct {
	LogJSONDisabled bool   `overwrite-flag:"JSONDisabled"`
	IdentityURL     string `overwrite-flag:"authURL" validate:"required"`
	CoordinatorURL  string `overwrite-flag:"coordinatorURL" validate:"required"`

	Coordinator struct {
		Host            string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
		Port            int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
		HealthCheckPort int    `overwrite-flag:"healthCheckPort" validate:"required"`
		LogLevel        string `overwrite-flag:"logLevel"`
		AuthTTL         int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
		Cluster         string `overwrite-flag:"cluster"`
		ServerSecret    string `overwrite-flag:"serverSecret" validate:"required"`
		Metrics         struct {
			Enabled   bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
			TraceName string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`
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
	defer logging.LogPanic()

	if err := logging.SetLevel(log, conf.Coordinator.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}

	coordinatorAuth, err := auth.MakeAuthenticator(&auth.AuthenticatorConfig{
		IdentityURL:    conf.IdentityURL,
		CoordinatorURL: conf.CoordinatorURL,
		Secret:         conf.Coordinator.ServerSecret,
		RequestTTL:     conf.Coordinator.AuthTTL,
		Log:            log,
	})

	if err != nil {
		log.WithError(err).Fatal("cannot build authenticator")
	}

	config := coordinator.Config{
		Auth:         coordinatorAuth,
		Log:          log,
		ReportPeriod: 10 * time.Second,
	}

	if conf.Coordinator.Metrics.Enabled {
		traceName := conf.Coordinator.Metrics.TraceName
		statusCheckMetric := fmt.Sprintf("%s-statsOk", traceName)
		versionTag := fmt.Sprintf("version:%s", version.Version())
		clusterTag := fmt.Sprintf("cluster:%s", conf.Coordinator.Cluster)
		tags := []string{"env:local", versionTag, clusterTag}

		metricsClient, err := metrics.NewClient(traceName, log)
		if err != nil {
			log.WithError(err).Fatal("cannot start metrics agent")
		}
		defer metricsClient.Close()

		config.Reporter = func(stats coordinator.Stats) {
			metricsClient.GaugeInt("client.count", stats.ClientCount, tags)
			metricsClient.GaugeInt("server.count", stats.ServerCount, tags)
			metricsClient.ServiceUp(statusCheckMetric)

			log.WithFields(logrus.Fields{
				"log_type":     "report",
				"client count": stats.ClientCount,
				"server count": stats.ServerCount,
			}).Info("report")
		}
	} else {
		config.Reporter = func(stats coordinator.Stats) {
			log.WithFields(logrus.Fields{
				"log_type":     "report",
				"client count": stats.ClientCount,
				"server count": stats.ServerCount,
			}).Info("report")
		}
	}

	state := coordinator.MakeState(&config)

	go coordinator.Start(state)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		addr := fmt.Sprintf("%s:%d", conf.Coordinator.Host, conf.Coordinator.HealthCheckPort)
		log.Fatal(http.ListenAndServe(addr, mux))
	}()

	mux := http.NewServeMux()
	coordinator.Register(state, mux)

	addr := fmt.Sprintf("%s:%d", conf.Coordinator.Host, conf.Coordinator.Port)
	log.Infof("starting coordinator %s - version: %s", addr, version.Version())
	log.Fatal(http.ListenAndServe(addr, mux))
}
