package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swanky2009/goim/logic"
	"github.com/swanky2009/goim/logic/g"
	logicgrpc "github.com/swanky2009/goim/logic/grpc"
	"github.com/swanky2009/goim/logic/handlers"
	logichttp "github.com/swanky2009/goim/logic/http"
	"github.com/swanky2009/goim/pkg/waitgroup"
	"google.golang.org/grpc"
)

const (
	ver   = "3.0.0"
	appid = "goim.logic"
)

func main() {
	var (
		wg      waitgroup.WaitGroupWrapper
		srv     *logic.Server
		rpcSrv  *grpc.Server
		httpSrv *http.Server
	)
	// Mechanical domain.
	errc := make(chan error)

	// new logic server
	srv = logic.NewServer(g.Conf)

	// new grpc server
	rpcSrv = logicgrpc.New(g.Conf.RPCServer, srv)
	wg.Wrap(func() {
		logicgrpc.Start(g.Conf.RPCServer, rpcSrv, errc)
	})

	// new http server
	httpSrv = logichttp.New(g.Conf.HTTPServer, srv)
	wg.Wrap(func() {
		logichttp.Start(httpSrv, errc)
	})

	//prometheus mertics
	wg.Wrap(func() {
		mux := http.NewServeMux()

		mux.Handle("/metrics", promhttp.Handler())

		g.Logger.Infof("start metrics server of prometheus listen: %s", g.Conf.MetricsServer.Addr)

		errc <- http.ListenAndServe(g.Conf.MetricsServer.Addr, mux)
	})

	// Interrupt handler.
	go handlers.InterruptHandler(srv, rpcSrv, httpSrv, errc)

	// Run!
	g.Logger.Infof("goim-logic exit", <-errc)
}
