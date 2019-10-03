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
	lastStats        map[uint64]commserver.PeerStats
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
		lastStats:        make(map[uint64]commserver.PeerStats),
	}
}

func getBytesSentByDC(s commserver.PeerStats) uint64 {
	return s.ReliableBytesSent + s.UnreliableBytesSent
}

func getBytesReceivedByDC(s commserver.PeerStats) uint64 {
	return s.ReliableBytesReceived + s.UnreliableBytesReceived
}

func getMessagesSentByDC(s commserver.PeerStats) uint64 {
	return uint64(s.ReliableMessagesSent) + uint64(s.UnreliableMessagesSent)
}

func getMessagesReceivedByDC(s commserver.PeerStats) uint64 {
	return uint64(s.ReliableMessagesReceived) + uint64(s.UnreliableMessagesReceived)
}

func (r *Reporter) Report(stats Stats) {
	if r.db != nil && time.Since(r.lastLongReport) > r.longReportPeriod {
		r.lastLongReport = time.Now()

		go r.reportDB(r.db, stats)
	}

	stateCount := make(map[commserver.ICEConnectionState]uint32)
	localCandidateTypeCount := make(map[commserver.ICECandidateType]uint32)
	remoteCandidateTypeCount := make(map[commserver.ICECandidateType]uint32)

	var messagesSentByDC, bytesSentByDC, bytesSentByICE, bytesSentBySCTP uint64
	var messagesReceivedByDC, bytesReceivedByDC, bytesReceivedByICE, bytesReceivedBySCTP uint64

	for _, pStats := range stats.Peers {
		pLastStats := r.lastStats[pStats.Alias]

		peerTag := fmt.Sprintf("peer:%d", pStats.Alias)
		stateTags := append([]string{peerTag}, r.tags...)

		if r.ddClient != nil && pStats.TopicCount > 0 {
			r.ddClient.GaugeUint32("peer.topicCount", pStats.TopicCount, stateTags)
		}

		stateCount[pStats.State]++

		if pStats.Nomination {
			localCandidateTypeCount[pStats.LocalCandidateType]++
			remoteCandidateTypeCount[pStats.LocalCandidateType]++
		}

		messagesSentByDC += getMessagesSentByDC(pStats) - getMessagesSentByDC(pLastStats)
		bytesSentByDC += getBytesSentByDC(pStats) - getBytesSentByDC(pLastStats)
		bytesSentByICE += pStats.ICETransportBytesSent - pLastStats.ICETransportBytesSent
		bytesSentBySCTP += pStats.SCTPTransportBytesSent - pLastStats.SCTPTransportBytesSent

		messagesReceivedByDC += getMessagesReceivedByDC(pStats) - getMessagesReceivedByDC(pLastStats)
		bytesReceivedByDC += getBytesReceivedByDC(pStats) - getBytesReceivedByDC(pLastStats)
		bytesReceivedByICE += pStats.ICETransportBytesReceived - pLastStats.ICETransportBytesReceived
		bytesReceivedBySCTP += pStats.SCTPTransportBytesReceived - pLastStats.SCTPTransportBytesReceived

		r.lastStats[pStats.Alias] = pStats
	}

	seconds := uint64(10)

	messagesSent := messagesSentByDC / seconds
	bytesSent := bytesSentByDC / seconds
	iceBytesSent := bytesSentByICE / seconds
	sctpBytesSent := bytesSentBySCTP / seconds

	messagesReceived := messagesReceivedByDC / seconds
	bytesReceived := bytesReceivedByDC / seconds
	iceBytesReceived := bytesReceivedByICE / seconds
	sctpBytesReceived := bytesReceivedBySCTP / seconds

	if r.ddClient != nil {
		r.ddClient.GaugeInt("topicCh.size", stats.TopicChSize, r.tags)
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
		r.log.Info().Str("log_type", "report").
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
