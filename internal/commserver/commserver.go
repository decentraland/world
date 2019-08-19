package commserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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

	lastBytesReceivedByDC uint64
	lastBytesSentByDC     uint64

	lastBytesReceivedBySCTP uint64
	lastBytesSentBySCTP     uint64

	lastBytesReceivedByICE uint64
	lastBytesSentByICE     uint64

	lastMessagesSentByDC     uint64
	lastMessagesReceivedByDC uint64
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

func (r *Reporter) Report(stats Stats) {
	if r.db != nil && time.Since(r.lastLongReport) > r.longReportPeriod {
		r.lastLongReport = time.Now()

		go r.reportDB(r.db, stats)
	}

	var bytesReceivedReliable, bytesSentReliable, messagesSentReliable, messagesReceivedReliable uint64
	var bytesReceivedUnreliable, bytesSentUnreliable, messagesSentUnreliable, messagesReceivedUnreliable uint64

	var sctpTransportBytesSent, sctpTransportBytesReceived uint64
	var iceTransportBytesSent, iceTransportBytesReceived uint64

	stateCount := make(map[commserver.ICEConnectionState]uint32)
	localCandidateTypeCount := make(map[commserver.ICECandidateType]uint32)
	remoteCandidateTypeCount := make(map[commserver.ICECandidateType]uint32)

	for _, peerStats := range stats.Peers {
		bytesSentReliable += peerStats.ReliableBytesSent
		bytesReceivedReliable += peerStats.ReliableBytesReceived
		messagesSentReliable += uint64(peerStats.ReliableMessagesSent)
		messagesReceivedReliable += uint64(peerStats.ReliableMessagesReceived)

		bytesSentUnreliable += peerStats.UnreliableBytesSent
		bytesReceivedUnreliable += peerStats.UnreliableBytesReceived
		messagesSentUnreliable += uint64(peerStats.UnreliableMessagesSent)
		messagesReceivedUnreliable += uint64(peerStats.UnreliableMessagesReceived)

		sctpTransportBytesSent += peerStats.SCTPTransportBytesSent
		sctpTransportBytesReceived += peerStats.SCTPTransportBytesReceived

		iceTransportBytesSent += peerStats.ICETransportBytesSent
		iceTransportBytesReceived += peerStats.ICETransportBytesReceived

		peerTag := fmt.Sprintf("peer:%d", peerStats.Alias)
		stateTags := append([]string{peerTag}, r.tags...)

		if r.ddClient != nil && peerStats.TopicCount > 0 {
			r.ddClient.GaugeUint32("peer.topicCount", peerStats.TopicCount, stateTags)
		}

		stateCount[peerStats.State]++

		if peerStats.Nomination {
			localCandidateTypeCount[peerStats.LocalCandidateType]++
			remoteCandidateTypeCount[peerStats.LocalCandidateType]++
		}
	}

	bytesReceivedByDC := bytesReceivedReliable + bytesReceivedUnreliable
	bytesSentByDC := bytesSentReliable + bytesSentUnreliable

	messagesReceivedByDC := messagesReceivedReliable + messagesReceivedUnreliable
	messagesSentByDC := messagesSentReliable + messagesSentUnreliable

	if r.ddClient != nil {
		r.ddClient.GaugeInt("topicCh.size", stats.TopicChSize, r.tags)
		r.ddClient.GaugeInt("connectCh.size", stats.ConnectChSize, r.tags)
		r.ddClient.GaugeInt("webrtcControlCh.size", stats.WebRtcControlChSize, r.tags)
		r.ddClient.GaugeInt("messagesCh.size", stats.MessagesChSize, r.tags)
		r.ddClient.GaugeInt("unregisterCh.size", stats.UnregisterChSize, r.tags)

		r.ddClient.GaugeInt("connection.count", len(stats.Peers), r.tags)
		r.ddClient.GaugeInt("topic.count", stats.TopicCount, r.tags)

		r.ddClient.GaugeUint64("totalBytesReceived", bytesReceivedByDC, r.tags)
		r.ddClient.GaugeUint64("totalBytesSent", bytesSentByDC, r.tags)

		r.ddClient.GaugeUint64("totalMessagesSent", messagesSentByDC, r.tags)
		r.ddClient.GaugeUint64("totalMessagesReceived", messagesReceivedByDC, r.tags)

		r.ddClient.GaugeUint64("totalBytesReceivedSCTP", sctpTransportBytesReceived, r.tags)
		r.ddClient.GaugeUint64("totalBytesSentSCTP", sctpTransportBytesSent, r.tags)

		r.ddClient.GaugeUint64("totalBytesReceivedICE", iceTransportBytesReceived, r.tags)
		r.ddClient.GaugeUint64("totalBytesSentICE", iceTransportBytesSent, r.tags)

		for connState, count := range stateCount {
			stateTag := fmt.Sprintf("state:%s", connState.String())
			stateTags := append([]string{stateTag}, r.tags...)
			r.ddClient.GaugeUint32("connection.stateCount", count, stateTags)
		}

		for localCandidateType, count := range localCandidateTypeCount {
			candidateTypeTag := fmt.Sprintf("candidateType:%s", localCandidateType.String())
			candidateTypeTags := append([]string{candidateTypeTag}, r.tags...)
			r.ddClient.GaugeUint32("connection.localCandidateTypeCount", count, candidateTypeTags)
		}

		for remoteCandidateType, count := range remoteCandidateTypeCount {
			candidateTypeTag := fmt.Sprintf("candidateType:%s", remoteCandidateType.String())
			candidateTypeTags := append([]string{candidateTypeTag}, r.tags...)
			r.ddClient.GaugeUint32("connection.remoteCandidateTypeCount", count, candidateTypeTags)
		}
	}

	if r.debugModeEnabled {
		r.log.Info().Str("log_type", "sent_stats").
			Uint64("avg bytes sent per second [DC]", (bytesSentByDC-r.lastBytesSentByDC)/10).
			Uint64("avg messages sent per second [DC]", (messagesSentByDC-r.lastMessagesSentByDC)/10).
			Uint64("avg bytes sent per second [ICE]", (iceTransportBytesSent-r.lastBytesSentByICE)/10).
			Uint64("avg bytes sent per second [SCTP]", (sctpTransportBytesSent-r.lastBytesSentBySCTP)/10).
			Msg("")

		r.log.Info().Str("log_type", "recv_stats").
			Uint64("avg bytes received per second [DC]", (bytesReceivedByDC-r.lastBytesReceivedByDC)/10).
			Uint64("avg messages received per second [DC]", (messagesReceivedByDC-r.lastMessagesReceivedByDC)/10).
			Uint64("avg bytes received per second [ICE]", (iceTransportBytesReceived-r.lastBytesReceivedByICE)/10).
			Uint64("avg bytes received per second [SCTP]", (sctpTransportBytesReceived-r.lastBytesReceivedBySCTP)/10).
			Msg("")

		r.lastBytesReceivedByDC = bytesReceivedByDC
		r.lastBytesSentByDC = bytesSentByDC
		r.lastBytesReceivedBySCTP = sctpTransportBytesReceived
		r.lastBytesSentBySCTP = sctpTransportBytesSent
		r.lastBytesReceivedByICE = iceTransportBytesReceived
		r.lastBytesSentByICE = iceTransportBytesSent
		r.lastMessagesSentByDC = messagesSentByDC
		r.lastMessagesReceivedByDC = messagesReceivedByDC

		r.log.Info().Str("log_type", "report").
			Int("peer_count", len(stats.Peers)).
			Int("topic_count", stats.TopicCount).
			Msg("report")
	}
}

func (r *Reporter) reportDB(db *sql.DB, stats Stats) {
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

	for _, peerStats := range stats.Peers {
		encodedStats, err := json.Marshal(peerStats)
		if err != nil {
			r.log.Error().Err(err).Msg("cannot encode stats as json")
			continue
		}

		_, err = stmt.Exec(peerStats.Alias, string(peerStats.Identity), version.Version(), encodedStats)
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
