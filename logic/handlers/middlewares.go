package handlers

import (
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

// WrapEndpoints accepts the service's entire collection of endpoints, so that a
// set of middlewares can be wrapped around every middleware (e.g., access
// logging and instrumentation), and others wrapped selectively around some
// endpoints and not others (e.g., endpoints requiring authenticated access).
// Note that the final middleware wrapped will be the outermost middleware
// (i.e. applied first)
func WrapEndpoints(in endpoint.Endpoint) endpoint.Endpoint {
	//全链路追踪
	in = ZipkinEndpointMiddleware()(in)

	//限频
	in = ratelimit.NewErroringLimiter(rate.NewLimiter(
		1000,  //每秒产生的令牌数
		10000, //令牌池最大值
	))(in)

	//熔断
	in = circuitbreaker.Gobreaker(
		gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "logic-rpc",
			MaxRequests: 5000,
			Interval:    1 * time.Second,
		}))(in)

	return in
}
