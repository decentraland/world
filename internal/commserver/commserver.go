package commserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	logrus "github.com/sirupsen/logrus"

	pq "github.com/lib/pq"

	"github.com/decentraland/webrtc-broker/pkg/commserver"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/version"
)

type Config = commserver.Config
type ICEServer = commserver.ICEServer
type Stats = commserver.Stats

var MakeState = commserver.MakeState
var ConnectCoordinator = commserver.ConnectCoordinator
var ProcessMessagesQueue = commserver.ProcessMessagesQueue
var Process = commserver.Process

type ReporterConfig struct {
	LongReportPeriod time.Duration
	DB               *sql.DB
	Client           *metrics.Client
	Cluster          string
	TraceName        string
	Log              *logging.Logger
}

type Reporter struct {
	longReportPeriod  time.Duration
	lastLongReport    time.Time
	db                *sql.DB
	client            *metrics.Client
	tags              []string
	log               *logging.Logger
	statusCheckMetric string
}

func NewReporter(config *ReporterConfig) *Reporter {
	versionTag := fmt.Sprintf("version:%s", version.Version())
	clusterTag := fmt.Sprintf("cluster:%s", config.Cluster)
	tags := []string{"env:local", versionTag, clusterTag}

	return &Reporter{
		longReportPeriod:  config.LongReportPeriod,
		lastLongReport:    time.Now(),
		db:                config.DB,
		client:            config.Client,
		tags:              tags,
		log:               config.Log,
		statusCheckMetric: fmt.Sprintf("%s-statsOk", config.TraceName),
	}
}

func (r *Reporter) Report(stats Stats) {
	if time.Since(r.lastLongReport) > r.longReportPeriod {
		defer func() {
			r.lastLongReport = time.Now()
		}()

		go r.reportDB(r.db, stats)
	}

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
		stateTags := append([]string{peerTag}, r.tags...)

		if peerStats.TopicCount > 0 {
			r.client.GaugeUint32("peer.topicCount", peerStats.TopicCount, stateTags)
		}

		stateCount[peerStats.State]++

		if peerStats.Nomination {
			localCandidateTypeCount[peerStats.LocalCandidateType]++
			remoteCandidateTypeCount[peerStats.LocalCandidateType]++
		}
	}

	r.client.GaugeInt("connection.count", len(stats.Peers), r.tags)
	r.client.GaugeInt("topic.count", stats.TopicCount, r.tags)
	r.client.GaugeUint64("totalBytesReceived", bytesReceived, r.tags)
	r.client.GaugeUint64("totalBytesSent", bytesSent, r.tags)
	r.client.GaugeUint64("totalMessagesSent", messagesSent, r.tags)
	r.client.GaugeUint64("totalMessagesReceived", messagesReceived, r.tags)

	for connState, count := range stateCount {
		stateTag := fmt.Sprintf("state:%s", connState.String())
		stateTags := append([]string{stateTag}, r.tags...)
		r.client.GaugeUint32("connection.stateCount", count, stateTags)
	}

	for localCandidateType, count := range localCandidateTypeCount {
		candidateTypeTag := fmt.Sprintf("candidateType:%s", localCandidateType.String())
		candidateTypeTags := append([]string{candidateTypeTag}, r.tags...)
		r.client.GaugeUint32("connection.localCandidateTypeCount", count, candidateTypeTags)
	}

	for remoteCandidateType, count := range remoteCandidateTypeCount {
		candidateTypeTag := fmt.Sprintf("candidateType:%s", remoteCandidateType.String())
		candidateTypeTags := append([]string{candidateTypeTag}, r.tags...)
		r.client.GaugeUint32("connection.remoteCandidateTypeCount", count, candidateTypeTags)
	}

	r.client.ServiceUp(r.statusCheckMetric)

	r.log.WithFields(logrus.Fields{
		"log_type":    "report",
		"peer count":  len(stats.Peers),
		"topic count": stats.TopicCount,
	}).Info("report")
}

func (r *Reporter) reportDB(db *sql.DB, stats Stats) {
	if len(stats.Peers) == 0 {
		return
	}

	txn, err := db.Begin()
	if err != nil {
		r.log.WithError(err).Error("cannot start tx")
		return
	}

	stmt, err := txn.Prepare(pq.CopyIn("stats", "peer_alias", "user_id", "version", "stats"))
	if err != nil {
		r.log.WithError(err).Error("cannot prepare statement")
		return
	}

	for _, peerStats := range stats.Peers {
		encodedStats, err := json.Marshal(peerStats)
		if err != nil {
			r.log.WithError(err).Error("cannot encode stats as json")
			continue
		}

		_, err = stmt.Exec(peerStats.Alias, string(peerStats.Identity), version.Version(), encodedStats)
		if err != nil {
			r.log.WithError(err).Error("cannot exec statement")
			continue
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		r.log.WithError(err).Error("cannot finalize statement")
		return
	}

	err = stmt.Close()
	if err != nil {
		r.log.WithError(err).Error("cannot close statement")
		return
	}

	err = txn.Commit()
	if err != nil {
		r.log.WithError(err).Error("cannot commit tx")
		return
	}
}
