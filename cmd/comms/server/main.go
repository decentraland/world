package main

import (
	"fmt"
	"time"

	"github.com/decentraland/webrtc-broker/pkg/commserver"
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
	CoordinatorURL  string `overwrite-flag:"coordinatorURL" flag-usage:"coordinator url" validate:"required"`
	CommServer      struct {
		LogLevel     string `overwrite-flag:"logLevel"`
		ServerSecret string `overwrite-flag:"serverSecret"`
		AuthTTL      int64  `overwrite-flag:"authTTL" flag-usage:"request time to live"`
		Cluster      string `overwrite-flag:"cluster"`
		Metrics      struct {
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
	if err := logging.SetLevel(log, conf.CommServer.LogLevel); err != nil {
		log.Fatal("error setting log level")
	}
	defer logging.LogPanic()

	serverAuth, err := auth.MakeAuthenticator(&auth.AuthenticatorConfig{
		IdentityURL: conf.IdentityURL,
		Secret:      conf.CommServer.ServerSecret,
		RequestTTL:  conf.CommServer.AuthTTL,
		Log:         log,
	})
	if err != nil {
		log.WithError(err).Fatal("cannot build authenticator")
	}

	config := commserver.Config{
		Auth: serverAuth,
		Log:  log,
		ICEServers: []commserver.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		CoordinatorURL:         fmt.Sprintf("%s/discover", conf.CoordinatorURL),
		ExitOnCoordinatorClose: true,
		ReportPeriod:           10 * time.Second,
	}

	if conf.CommServer.Metrics.Enabled {
		traceName := conf.CommServer.Metrics.TraceName
		statusCheckMetric := fmt.Sprintf("%s-statsOk", traceName)
		versionTag := fmt.Sprintf("version:%s", version.Version())
		clusterTag := fmt.Sprintf("cluster:%s", conf.CommServer.Cluster)
		tags := []string{"env:local", versionTag, clusterTag}

		metricsClient, err := metrics.NewMetricsClient(traceName, log)
		if err != nil {
			log.WithError(err).Fatal("cannot start metrics agent")
		}
		defer metricsClient.Close()

		config.Reporter = func(stats commserver.Stats) {
			var bytesReceived, bytesSent, messagesSent, messagesReceived uint64
			stateCount := make(map[commserver.ICEConnectionState]uint32)
			localCandidateTypeCount := make(map[commserver.ICECandidateType]uint32)
			remoteCandidateTypeCount := make(map[commserver.ICECandidateType]uint32)

			for _, peerStats := range stats.Peers {
				bytesReceived += peerStats.ReliableBytesReceived + peerStats.UnreliableBytesReceived
				bytesSent += peerStats.ReliableBytesSent + peerStats.UnreliableBytesSent

				messagesReceived += uint64(peerStats.ReliableMessagesReceived +
					peerStats.UnreliableMessagesReceived)
				messagesSent += uint64(peerStats.ReliableMessagesSent +
					peerStats.UnreliableMessagesSent)

				peerTag := fmt.Sprintf("peer:%d", peerStats.Alias)
				stateTags := append([]string{peerTag}, tags...)
				metricsClient.GaugeUint32("peer.topicCount", peerStats.TopicCount, stateTags)

				stateCount[peerStats.State]++

				if peerStats.Nomination {
					localCandidateTypeCount[peerStats.LocalCandidateType]++
					remoteCandidateTypeCount[peerStats.LocalCandidateType]++
				}
			}

			metricsClient.GaugeInt("connection.count", len(stats.Peers), tags)
			metricsClient.GaugeInt("topic.count", stats.TopicCount, tags)
			metricsClient.GaugeUint64("totalBytesReceived", bytesReceived, tags)
			metricsClient.GaugeUint64("totalBytesSent", bytesSent, tags)
			metricsClient.GaugeUint64("totalMessagesSent", messagesSent, tags)
			metricsClient.GaugeUint64("totalMessagesReceived", messagesReceived, tags)

			for connState, count := range stateCount {
				stateTag := fmt.Sprintf("state:%s", connState.String())
				stateTags := append([]string{stateTag}, tags...)
				metricsClient.GaugeUint32("connection.stateCount", count, stateTags)
			}

			for localCandidateType, count := range localCandidateTypeCount {
				candidateTypeTag := fmt.Sprintf("candidateType:%s", localCandidateType.String())
				candidateTypeTags := append([]string{candidateTypeTag}, tags...)
				metricsClient.GaugeUint32("connection.localCandidateTypeCount", count, candidateTypeTags)
			}

			for remoteCandidateType, count := range remoteCandidateTypeCount {
				candidateTypeTag := fmt.Sprintf("candidateType:%s", remoteCandidateType.String())
				candidateTypeTags := append([]string{candidateTypeTag}, tags...)
				metricsClient.GaugeUint32("connection.remoteCandidateTypeCount", count, candidateTypeTags)
			}

			metricsClient.ServiceUp(statusCheckMetric)
		}
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

	if err := commserver.ConnectCoordinator(state); err != nil {
		log.Fatal("connect coordinator failure ", err)
	}

	go commserver.ProcessMessagesQueue(state)
	commserver.Process(state)
}
