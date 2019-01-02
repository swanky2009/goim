package g

import (
	"github.com/go-kit/kit/metrics"
	//kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	//stdprometheus "github.com/prometheus/client_golang/prometheus"
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

	// speed
	SpeedMsgSecond       metrics.Gauge
	SpeedRoomBatchSecond metrics.Gauge

	// room
	ActiveRoomCount metrics.Gauge

	// nodes
	CometNodes metrics.Gauge
}

func MetricsInstrumenting() *Metrics {
	//namespace, subsystem := "goim", "job"
	//fieldKeys := []string{"count"}

	// AllMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Subsystem: subsystem,
	// 	Name:      "allmsg",
	// 	Help:      "Number of messages received.",
	// }, fieldKeys)
	// PushMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Subsystem: subsystem,
	// 	Name:      "pushmsg",
	// 	Help:      "Number of push messages received.",
	// }, fieldKeys)
	// BroadcastMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Subsystem: subsystem,
	// 	Name:      "broadcastmsg",
	// 	Help:      "Number of broadcast messages received.",
	// }, fieldKeys)
	// BroadcastRoomMsg := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
	// 	Namespace: namespace,
	// 	Subsystem: subsystem,
	// 	Name:      "broadcastroommsg",
	// 	Help:      "Number of broadcastroom messages received.",
	// }, fieldKeys)
	// SpeedMsgSecond := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
	// 	Namespace: namespace,
	// 	Subsystem: subsystem,
	// 	Name:      "speedmsgsecond",
	// 	Help:      "Total count of messages received in seconds.",
	// }, fieldKeys)
	// CometNodes := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
	// 	Namespace: namespace,
	// 	Subsystem: subsystem,
	// 	Name:      "cometnodes",
	// 	Help:      "Number of comet nodes.",
	// }, fieldKeys)

	return &Metrics{
		// AllMsg,
		// PushMsg,
		// BroadcastMsg,
		// BroadcastRoomMsg,
		// SpeedMsgSecond,
		// CometNodes,
	}
}

// func (s *Metrics) IncrTcpOnline() {
// 	lvs := []string{"count", "/v1/tcp_online"}
// 	s.TcpOnline.With(lvs...).Add(1)

// 	lvs = []string{"count", "/v1/online"}
// 	s.Online.With(lvs...).Add(1)
// }

// func (s *Metrics) DecrTcpOnline() {
// 	lvs := []string{"count", "/v1/tcp_online"}
// 	s.TcpOnline.With(lvs...).Add(-1)

// 	lvs = []string{"count", "/v1/online"}
// 	s.Online.With(lvs...).Add(-1)
// }

// func (s *Metrics) IncrWsOnline() {
// 	lvs := []string{"count", "/v1/ws_online"}
// 	s.WsOnline.With(lvs...).Add(1)

// 	lvs = []string{"count", "/v1/online"}
// 	s.Online.With(lvs...).Add(1)
// }

// func (s *Metrics) DecrWsOnline() {
// 	lvs := []string{"count", "/v1/ws_online"}
// 	s.WsOnline.With(lvs...).Add(-1)

// 	lvs = []string{"count", "/v1/online"}
// 	s.Online.With(lvs...).Add(-1)
// }

func (s *Metrics) IncrPushMsg() {
	lvs := []string{"count", "/v1/pushmsg"}
	s.PushMsg.With(lvs...).Add(1)

	lvs = []string{"count", "/v1/allmsg"}
	s.AllMsg.With(lvs...).Add(1)
}

func (s *Metrics) IncrBroadcastMsg() {
	lvs := []string{"count", "/v1/broadcastmsg"}
	s.BroadcastMsg.With(lvs...).Add(1)

	lvs = []string{"count", "/v1/allmsg"}
	s.AllMsg.With(lvs...).Add(1)
}

func (s *Metrics) IncrBroadcastRoomMsg() {
	s.BroadcastRoomMsg.Add(1)
	s.AllMsg.Add(1)
}

// func (s *Metrics) SetBucketChannels() {
// 	lvs = []string{"count", "/v1/bucketchannels"}
// 	s.BucketChannels.With(lvs...).Observe()
// }

// func (s *Metrics) SetBucketChannels() {
// 	lvs = []string{"count", "/v1/bucketrooms"}
// 	s.BucketRooms.With(lvs...).Observe()
// }
