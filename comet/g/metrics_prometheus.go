package g

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	// online
	Online    metrics.Gauge
	TcpOnline metrics.Gauge
	WsOnline  metrics.Gauge
	// messages
	AllMsg           metrics.Counter
	PushMsg          metrics.Counter
	BroadcastMsg     metrics.Counter
	BroadcastRoomMsg metrics.Counter
	//speed
	SpeedMsgSecond metrics.Gauge
	// buckets
	BucketChannels metrics.Histogram
	BucketRooms    metrics.Histogram
}

func MetricsInstrumenting() *Metrics {
	namespace, subsystem := "goim", "comet"
	fieldKeys := []string{"count"}

	Online := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "online",
		Help:      "Number of online user.",
	}, fieldKeys)
	TcpOnline := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "tcponline",
		Help:      "Number of tcp online user.",
	}, fieldKeys)
	WsOnline := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "wsonline",
		Help:      "Number of websocket online user.",
	}, fieldKeys)
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
	SpeedMsgSecond := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "speedmsgsecond",
		Help:      "Total count of messages received in seconds.",
	}, fieldKeys)

	buckets := []float64{}
	for i := 0; i < Conf.Bucket.Size; i++ {
		buckets = append(buckets, float64(i))
	}

	BucketChannels := kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "bucketchannels",
		Help:      "This is the help string for the histogram.",
		Buckets:   buckets,
	}, fieldKeys)

	BucketRooms := kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "bucketrooms",
		Help:      "This is the help string for the histogram.",
		Buckets:   buckets,
	}, fieldKeys)

	return &Metrics{
		Online,
		TcpOnline,
		WsOnline,
		AllMsg,
		PushMsg,
		BroadcastMsg,
		BroadcastRoomMsg,
		SpeedMsgSecond,
		BucketChannels,
		BucketRooms,
	}
}

func (s *Metrics) IncrTcpOnline() {
	lvs := []string{"count", "/v1/tcp_online"}
	s.TcpOnline.With(lvs...).Add(1)

	lvs = []string{"count", "/v1/online"}
	s.Online.With(lvs...).Add(1)
}

func (s *Metrics) DecrTcpOnline() {
	lvs := []string{"count", "/v1/tcp_online"}
	s.TcpOnline.With(lvs...).Add(-1)

	lvs = []string{"count", "/v1/online"}
	s.Online.With(lvs...).Add(-1)
}

func (s *Metrics) IncrWsOnline() {
	lvs := []string{"count", "/v1/ws_online"}
	s.WsOnline.With(lvs...).Add(1)

	lvs = []string{"count", "/v1/online"}
	s.Online.With(lvs...).Add(1)
}

func (s *Metrics) DecrWsOnline() {
	lvs := []string{"count", "/v1/ws_online"}
	s.WsOnline.With(lvs...).Add(-1)

	lvs = []string{"count", "/v1/online"}
	s.Online.With(lvs...).Add(-1)
}

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
	lvs := []string{"count", "/v1/broadcastroommsg"}
	s.BroadcastRoomMsg.With(lvs...).Add(1)

	lvs = []string{"count", "/v1/allmsg"}
	s.AllMsg.With(lvs...).Add(1)
}

// func (s *Metrics) SetBucketChannels() {
// 	lvs = []string{"count", "/v1/bucketchannels"}
// 	s.BucketChannels.With(lvs...).Observe()
// }

// func (s *Metrics) SetBucketChannels() {
// 	lvs = []string{"count", "/v1/bucketrooms"}
// 	s.BucketRooms.With(lvs...).Observe()
// }
