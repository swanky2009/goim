package handlers

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	zipkingo "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"google.golang.org/grpc/metadata"
	"time"

	"github.com/swanky2009/goim/logic/g"
)

func ZipkinEndpointMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			zipkinTracer, err := g.NewZipkinTracer()
			if err != nil {
				g.Logger.Error("zipkinTracer create failed,", err)
				return next(ctx, request)
			}

			var sc model.SpanContext
			md, _ := metadata.FromIncomingContext(ctx)
			sc = zipkinTracer.Extract(b3.ExtractGRPC(&md))

			span := zipkinTracer.StartSpan("in service goim-logic", zipkingo.Parent(sc))
			span.Annotate(time.Now(), "in endpoint")

			defer func() {
				span.Annotate(time.Now(), "out endpoint")
				span.Finish()
			}()

			ctx = zipkingo.NewContext(ctx, span)
			return next(ctx, request)
		}
	}
}
