package commserver

import (
	"fmt"

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

type Reporter struct {
	client            *metrics.Client
	tags              []string
	log               *logging.Logger
	statusCheckMetric string
}

type ReporterConfig struct {
	Client    *metrics.Client
	Cluster   string
	TraceName string
	Log       *logging.Logger
}

func NewReporter(config *ReporterConfig) *Reporter {
	versionTag := fmt.Sprintf("version:%s", version.Version())
	clusterTag := fmt.Sprintf("cluster:%s", config.Cluster)
	tags := []string{"env:local", versionTag, clusterTag}

	return &Reporter{
		client:            config.Client,
		tags:              tags,
		log:               config.Log,
		statusCheckMetric: fmt.Sprintf("%s-statsOk", config.TraceName),
	}
}

func (r *Reporter) Report(stats Stats) {
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
		r.client.GaugeUint32("peer.topicCount", peerStats.TopicCount, stateTags)

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
}
