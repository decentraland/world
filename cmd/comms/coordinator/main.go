package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "net/http/pprof"

	brokerAuth "github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/webrtc-broker/pkg/coordinator"

	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/version"

	zl "github.com/rs/zerolog/log"
)

type rootConfig struct {
	IdentityURL    string `overwrite-flag:"authURL" validate:"required"`
	CoordinatorURL string `overwrite-flag:"coordinatorURL" validate:"required"`

	Coordinator struct {
		LogLevel string `overwrite-flag:"logLevel"`

		Host    string `overwrite-flag:"host"      flag-usage:"host name" validate:"required"`
		Port    int    `overwrite-flag:"port"      flag-usage:"host port" validate:"required"`
		APIPort int    `overwrite-flag:"apiPort" validate:"required"`

		AuthTTL      int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
		AuthEnabled  bool   `overwrite-flag:"authEnabled"`
		ServerSecret string `overwrite-flag:"serverSecret" validate:"required"`

		Metrics struct {
			Cluster   string `overwrite-flag:"cluster"`
			Enabled   bool   `overwrite-flag:"metrics" flag-usage:"enable metrics"`
			TraceName string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`
		}
	}
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		zl.Fatal().Err(err).Msg("cannot read config")
	}

	loggerConfig := logging.LoggerConfig{Level: conf.Coordinator.LogLevel}
	log, err := logging.New(&loggerConfig)
	if err != nil {
		log.Fatal().Msg("error setting log level")
	}

	defer logging.LogPanic(log)

	var authenticator brokerAuth.CoordinatorAuthenticator

	if conf.Coordinator.AuthEnabled {
		authenticator, err = auth.MakeAuthenticator(&auth.AuthenticatorConfig{
			IdentityURL:    conf.IdentityURL,
			CoordinatorURL: conf.CoordinatorURL,
			Secret:         conf.Coordinator.ServerSecret,
			RequestTTL:     conf.Coordinator.AuthTTL,
			Log:            log,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("cannot build authenticator")
		}
	} else {
		authenticator = &brokerAuth.NoopAuthenticator{}
	}

	config := coordinator.Config{
		Auth:         authenticator,
		Log:          &log,
		ReportPeriod: 10 * time.Second,
	}

	if conf.Coordinator.Metrics.Enabled {
		traceName := conf.Coordinator.Metrics.TraceName
		versionTag := fmt.Sprintf("version:%s", version.Version())
		clusterTag := fmt.Sprintf("cluster:%s", conf.Coordinator.Metrics.Cluster)
		tags := []string{"env:local", versionTag, clusterTag}

		metricsClient, err := metrics.NewClient(traceName, log)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot start metrics agent")
		}
		defer metricsClient.Close()

		config.Reporter = func(stats coordinator.Stats) {
			metricsClient.GaugeInt("client.count", stats.ClientCount, tags)
			metricsClient.GaugeInt("server.count", stats.ServerCount, tags)

			log.Info().Str("log_type", "report").
				Int("client_count", stats.ClientCount).
				Int("server_count", stats.ServerCount).
				Msg("report")
		}
	} else {
		config.Reporter = func(stats coordinator.Stats) {
			log.Info().Str("log_type", "report").
				Int("client_count", stats.ClientCount).
				Int("server_count", stats.ServerCount).
				Msg("report")
		}
	}

	state := coordinator.MakeState(&config)

	go func() {
		addr := fmt.Sprintf("0.0.0.0:9081")
		log.Info().Str("address", addr).Msg("Starting profiler")
		log.Error().Err(http.ListenAndServe(addr, nil))
	}()

	go coordinator.Start(state)

	go func() {
		versionResponse, err := json.Marshal(map[string]string{"version": version.Version()})
		if err != nil {
			log.Fatal().Err(err).Msg("invalid version")
			return
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(versionResponse)
		})
		addr := fmt.Sprintf("%s:%d", conf.Coordinator.Host, conf.Coordinator.APIPort)
		log.Info().Str("address", addr).Msg("Starting HTTP API")
		log.Fatal().Err(http.ListenAndServe(addr, mux)).Msg("")
	}()

	mux := http.NewServeMux()
	coordinator.Register(state, mux)

	addr := fmt.Sprintf("%s:%d", conf.Coordinator.Host, conf.Coordinator.Port)
	log.Info().Str("addr", addr).Str("version", version.Version()).Msg("starting coordinator")
	log.Fatal().Err(http.ListenAndServe(addr, mux)).Msg("")
}
