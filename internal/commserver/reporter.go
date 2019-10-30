package commserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	pq "github.com/lib/pq"

	"github.com/decentraland/webrtc-broker/pkg/broker"
	"github.com/decentraland/world/internal/commons/logging"
	"github.com/decentraland/world/internal/commons/metrics"
	"github.com/decentraland/world/internal/commons/version"
)

type ReporterConfig struct {
	LongReportPeriod time.Duration
	DB               *sql.DB
	DDClient         *metrics.Client
	Cluster          string
	Log              logging.Logger
	DebugModeEnabled bool
}

type Reporter struct {
	longReportPeriod time.Duration
	lastLongReport   time.Time
	db               *sql.DB
	ddClient         *metrics.Client
	tags             []string
	log              logging.Logger
	debugModeEnabled bool
}

func NewReporter(config *ReporterConfig) *Reporter {
	versionTag := fmt.Sprintf("version:%s", version.Version())
	clusterTag := fmt.Sprintf("cluster:%s", config.Cluster)
	tags := []string{"env:local", versionTag, clusterTag}

	return &Reporter{
		longReportPeriod: config.LongReportPeriod,
		lastLongReport:   time.Now(),
		db:               config.DB,
		ddClient:         config.DDClient,
		tags:             tags,
		log:              config.Log,
		debugModeEnabled: config.DebugModeEnabled,
	}
}

func (r *Reporter) Report(stats broker.Stats) {
	if r.db != nil && time.Since(r.lastLongReport) > r.longReportPeriod {
		r.lastLongReport = time.Now()

		go r.reportDB(r.db, stats)
	}

	seconds := uint64(10)
	summaryGenerator := broker.NewStatsSummaryGenerator()

	summary := summaryGenerator.Generate(stats)

	messagesSent := summary.MessagesSentByDC / seconds
	bytesSent := summary.BytesSentByDC / seconds
	iceBytesSent := summary.BytesSentByICE / seconds
	sctpBytesSent := summary.BytesSentBySCTP / seconds

	messagesReceived := summary.MessagesReceivedByDC / seconds
	bytesReceived := summary.BytesReceivedByDC / seconds
	iceBytesReceived := summary.BytesReceivedByICE / seconds
	sctpBytesReceived := summary.BytesReceivedBySCTP / seconds

	if r.ddClient != nil {
		// r.ddClient.GaugeInt("topicCh.size", stats.TopicChSize, r.tags)
		r.ddClient.GaugeInt("connectCh.size", stats.ConnectChSize, r.tags)
		r.ddClient.GaugeInt("webrtcControlCh.size", stats.WebRtcControlChSize, r.tags)
		r.ddClient.GaugeInt("messagesCh.size", stats.MessagesChSize, r.tags)
		r.ddClient.GaugeInt("unregisterCh.size", stats.UnregisterChSize, r.tags)

		r.ddClient.GaugeInt("connection.count", len(stats.Peers), r.tags)
		r.ddClient.GaugeInt("topic.count", stats.TopicCount, r.tags)

		r.ddClient.GaugeUint64("messagesSent", messagesSent, r.tags)
		r.ddClient.GaugeUint64("bytesSent", bytesSent, r.tags)
		r.ddClient.GaugeUint64("bytesSentICE", iceBytesSent, r.tags)
		r.ddClient.GaugeUint64("bytesSentSCTP", sctpBytesSent, r.tags)

		r.ddClient.GaugeUint64("messagesReceived", messagesReceived, r.tags)
		r.ddClient.GaugeUint64("bytesReceived", bytesReceived, r.tags)
		r.ddClient.GaugeUint64("bytesReceivedICE", iceBytesReceived, r.tags)
		r.ddClient.GaugeUint64("bytesReceivedSCTP", sctpBytesReceived, r.tags)

		for connState, count := range summary.StateCount {
			stateTag := fmt.Sprintf("state:%s", connState.String())
			stateTags := append([]string{stateTag}, r.tags...)
			r.ddClient.GaugeUint32("connection.stateCount", count, stateTags)
		}

		for localCandidateType, count := range summary.LocalCandidateTypeCount {
			candidateTypeTag := fmt.Sprintf("candidateType:%s", localCandidateType.String())
			candidateTypeTags := append([]string{candidateTypeTag}, r.tags...)
			r.ddClient.GaugeUint32("connection.localCandidateTypeCount", count, candidateTypeTags)
		}

		for remoteCandidateType, count := range summary.RemoteCandidateTypeCount {
			candidateTypeTag := fmt.Sprintf("candidateType:%s", remoteCandidateType.String())
			candidateTypeTags := append([]string{candidateTypeTag}, r.tags...)
			r.ddClient.GaugeUint32("connection.remoteCandidateTypeCount", count, candidateTypeTags)
		}
	}

	if r.debugModeEnabled {
		r.log.Info().
			Uint64("messages sent per second [DC]", messagesSent).
			Uint64("bytes sent per second [DC]", bytesSent).
			Uint64("bytes sent per second [ICE]", iceBytesSent).
			Uint64("bytes sent per second [SCTP]", sctpBytesSent).
			Uint64("messages received per second [DC]", messagesReceived).
			Uint64("bytes received per second [DC]", bytesReceived).
			Uint64("bytes received per second [ICE]", iceBytesReceived).
			Uint64("bytes received per second [SCTP]", sctpBytesReceived).
			Int("peer_count", len(stats.Peers)).
			Int("topic_count", stats.TopicCount).
			Msg("")
	}
}

func (r *Reporter) reportDB(db *sql.DB, stats broker.Stats) {
	if len(stats.Peers) == 0 {
		return
	}

	txn, err := db.Begin()
	if err != nil {
		r.log.Error().Err(err).Msg("cannot start tx")
		return
	}

	stmt, err := txn.Prepare(pq.CopyIn("stats", "peer_alias", "user_id", "version", "stats"))
	if err != nil {
		r.log.Error().Err(err).Msg("cannot prepare statement")
		return
	}

	for _, pStats := range stats.Peers {
		encodedStats, err := json.Marshal(pStats)
		if err != nil {
			r.log.Error().Err(err).Msg("cannot encode stats as json")
			continue
		}

		_, err = stmt.Exec(pStats.Alias, string(pStats.Identity), version.Version(), encodedStats)
		if err != nil {
			r.log.Error().Err(err).Msg("cannot exec statement")
			continue
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		r.log.Error().Err(err).Msg("cannot finalize statement")
		return
	}

	err = stmt.Close()
	if err != nil {
		r.log.Error().Err(err).Msg("cannot close statement")
		return
	}

	err = txn.Commit()
	if err != nil {
		r.log.Error().Err(err).Msg("cannot commit tx")
		return
	}
}
