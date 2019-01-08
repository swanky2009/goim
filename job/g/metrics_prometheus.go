package g

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	// messages
	AllMsg           metrics.Counter
	PushMsg          metrics.Counter
	BroadcastMsg     metrics.Counter
	BroadcastRoomMsg metrics.Counter

	// miss
	PushMsgFailed          metrics.Counter
	BroadcastMsgFailed     metrics.Counter
	BroadcastRoomMsgFailed metrics.Counter

	// room
	ActiveRoomCount metrics.Gauge

	// nodes
	CometNodes metrics.Gauge
}

func MetricsInstrumenting() *Metrics {
	namespace, subsystem := "goim", "job"
	fieldKeys := []string{}

	AllMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "allmsg",
		Help:      "Number of messages received.",
	}, fieldKeys)
	PushMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pushmsg",
		Help:      "Number of push messages received.",
	}, fieldKeys)
	BroadcastMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "broadcastmsg",
		Help:      "Number of broadcast messages received.",
	}, fieldKeys)
	BroadcastRoomMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "broadcastroommsg",
		Help:      "Number of broadcastroom messages received.",
	}, fieldKeys)

	PushMsgFailed := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pushmsg_failed",
		Help:      "Number of push messages failed.",
	}, fieldKeys)
	BroadcastMsgFailed := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "broadcastmsg_failed",
		Help:      "Number of broadcast messages failed.",
	}, fieldKeys)
	BroadcastRoomMsgFailed := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "broadcastroommsg_failed",
		Help:      "Number of broadcastroom messages failed.",
	}, fieldKeys)

	ActiveRoomCount := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "activeroomcount",
		Help:      "Number of rooms actived.",
	}, fieldKeys)

	CometNodes := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cometnodes",
		Help:      "Number of comet nodes.",
	}, fieldKeys)

	return &Metrics{
		AllMsg,
		PushMsg,
		BroadcastMsg,
		BroadcastRoomMsg,
		PushMsgFailed,
		BroadcastMsgFailed,
		BroadcastRoomMsgFailed,
		ActiveRoomCount,
		CometNodes,
	}
}

func (s *Metrics) IncrPushMsg() {
	s.PushMsg.Add(1)
	s.AllMsg.Add(1)
}

func (s *Metrics) IncrBroadcastMsg() {
	s.BroadcastMsg.Add(1)
	s.AllMsg.Add(1)
}

func (s *Metrics) IncrBroadcastRoomMsg() {
	s.BroadcastRoomMsg.Add(1)
	s.AllMsg.Add(1)
}

func (s *Metrics) IncrPushMsgFailed() {
	s.PushMsgFailed.Add(1)
}

func (s *Metrics) IncrBroadcastMsgFailed() {
	s.BroadcastMsgFailed.Add(1)
}

func (s *Metrics) IncrBroadcastRoomMsgFailed() {
	s.BroadcastRoomMsgFailed.Add(1)
}

func (s *Metrics) SetActiveRoomCount(n float64) {
	s.ActiveRoomCount.Set(n)
}

func (s *Metrics) IncrCometNodes() {
	s.CometNodes.Add(1)
}

func (s *Metrics) DecrCometNodes() {
	s.CometNodes.Add(-1)
}
