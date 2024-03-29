package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "net/http/pprof"

	brokerAuth "github.com/decentraland/webrtc-broker/pkg/authentication"
	"github.com/decentraland/webrtc-broker/pkg/broker"
	"github.com/decentraland/webrtc-broker/pkg/protocol"
	"github.com/decentraland/world/internal/commons/auth"
	"github.com/decentraland/world/internal/commons/config"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/version"
	"github.com/decentraland/world/internal/commserver"
	_ "github.com/lib/pq"
	pion "github.com/pion/webrtc/v2"
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"
)

type rootConfig struct {
	IdentityURL    string `overwrite-flag:"authURL" validate:"required"`
	CoordinatorURL string `overwrite-flag:"coordinatorURL" flag-usage:"coordinator url" validate:"required"`

	CommServer struct {
		LogLevel string `overwrite-flag:"logLevel"`

		APIHost string `overwrite-flag:"apiHost" validate:"required"`
		APIPort int    `overwrite-flag:"apiPort" validate:"required"`

		AuthEnabled  bool   `overwrite-flag:"authEnabled"`
		AuthTTL      int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
		ServerSecret string `overwrite-flag:"serverSecret"`

		MaxPeers int `overwrite-flag:"maxPeers"`

		Metrics struct {
			Cluster string `overwrite-flag:"cluster"`

			DDEnabled bool   `overwrite-flag:"ddMetrics" flag-usage:"enable dd metrics"`
			TraceName string `overwrite-flag:"traceName" flag-usage:"metrics identifier" validate:"required"`

			DBEnabled       bool   `overwrite-flag:"dbMetrics" flag-usage:"enable db metrics"`
			StatsDBHost     string `overwrite-flag:"statsDBHost"`
			StatsDBName     string `overwrite-flag:"statsDBName"`
			StatsDBPort     int    `overwrite-flag:"statsDBPort"`
			StatsDBUser     string `overwrite-flag:"statsDBUser"`
			StatsDBPassword string `overwrite-flag:"statsDBPassword"`

			DebugEnabled bool `overwrite-flag:"debugMetrics" flag-usage:"enable debug metrics"`
		}
	}
}

func main() {
	var conf rootConfig
	if err := config.ReadConfiguration("config/config", &conf); err != nil {
		zl.Fatal().Err(err).Msg("invalid config")
	}

	log, err := logging.New(&logging.LoggerConfig{Level: conf.CommServer.LogLevel})
	if err != nil {
		log.Fatal().Msg("error setting log level")
	}
	defer logging.LogPanic(log)

	var authenticator brokerAuth.ServerAuthenticator

	if conf.CommServer.AuthEnabled {
		authenticator, err = auth.MakeAuthenticator(&auth.AuthenticatorConfig{
			IdentityURL: conf.IdentityURL,
			Secret:      conf.CommServer.ServerSecret,
			RequestTTL:  conf.CommServer.AuthTTL,
			Log:         log,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("cannot build authenticator")
		}
	} else {
		authenticator = &brokerAuth.NoopAuthenticator{}
	}

	config := broker.Config{
		Role: protocol.Role_COMMUNICATION_SERVER,
		Auth: authenticator,
		Log:  &log,
		ICEServers: []pion.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		CoordinatorURL:         conf.CoordinatorURL,
		MaxPeers:               uint16(conf.CommServer.MaxPeers),
		ExitOnCoordinatorClose: true,
		WebRtcLogLevel:         zerolog.WarnLevel,
	}

	reportConfig := commserver.ReporterConfig{
		LongReportPeriod: 10 * time.Minute,
		Log:              log,
		Cluster:          conf.CommServer.Metrics.Cluster,
		DebugModeEnabled: conf.CommServer.Metrics.DebugEnabled,
	}

	if conf.CommServer.Metrics.DDEnabled {
		client, err := metrics.NewClient(conf.CommServer.Metrics.TraceName, log)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot start metrics agent")
		}
		defer client.Close()

		reportConfig.DDClient = client
	}

	if conf.CommServer.Metrics.DBEnabled {
		psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			conf.CommServer.Metrics.StatsDBHost,
			conf.CommServer.Metrics.StatsDBPort,
			conf.CommServer.Metrics.StatsDBUser,
			conf.CommServer.Metrics.StatsDBPassword,
			conf.CommServer.Metrics.StatsDBName)
		db, err := sql.Open("postgres", psqlConn)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot open postgresql connection")
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.Fatal().Err(err).Msg("cannot open postgresql connection")
		}

		reportConfig.DB = db
	}

	b, err := broker.NewBroker(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create a new broker")
	}

	log.Info().Str("version", version.Version()).Msg("starting communication server")

	go func() {
		addr := fmt.Sprintf("0.0.0.0:9081")
		log.Info().Str("address", addr).Msg("Starting profiler")
		log.Error().Err(http.ListenAndServe(addr, nil))
	}()

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
		addr := fmt.Sprintf("%s:%d", conf.CommServer.APIHost, conf.CommServer.APIPort)
		log.Info().Str("address", addr).Msg("Starting HTTP API")
		log.Fatal().Err(http.ListenAndServe(addr, mux)).Msg("")
	}()

	if err := b.Connect(); err != nil {
		log.Fatal().Err(err).Msg("connect coordinator failure")
	}

	go b.ProcessSubscriptionChannel()

	go b.ProcessMessagesChannel()

	go b.ProcessControlMessages()

	reporter := commserver.NewReporter(&reportConfig)
	seconds := uint64(10)
	reportTicker := time.NewTicker(time.Duration(seconds) * time.Second)

	for {
		<-reportTicker.C

		stats := b.GetBrokerStats()
		reporter.Report(stats)
	}
}
