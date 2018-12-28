package main

import (
	"net/http"
	"runtime"

	//"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swanky2009/goim/comet"
	"github.com/swanky2009/goim/comet/g"
	cometgrpc "github.com/swanky2009/goim/comet/grpc"
	"github.com/swanky2009/goim/comet/handlers"
	"github.com/swanky2009/goim/pkg/waitgroup"
	"google.golang.org/grpc"
)

func main() {
	var (
		wg     waitgroup.WaitGroupWrapper
		srv    *comet.Server
		rpcSrv *grpc.Server
	)
	// Mechanical domain.
	errc := make(chan error)

	// new comet server
	srv = comet.NewServer(g.Conf)

	wg.Wrap(func() {
		if err := comet.InitTCP(srv, g.Conf.TCP.Bind, g.Conf.MaxProc); err != nil {
			errc <- err
		}
	})

	wg.Wrap(func() {
		if err := comet.InitWebsocket(srv, g.Conf.WebSocket.Bind, g.Conf.MaxProc); err != nil {
			errc <- err
		}
	})

	if g.Conf.WebSocket.TLSOpen {
		wg.Wrap(func() {
			if err := comet.InitWebsocketWithTLS(srv, g.Conf.WebSocket.TLSBind, g.Conf.WebSocket.CertFile, g.Conf.WebSocket.PrivateFile, runtime.NumCPU()); err != nil {
				errc <- err
			}
		})
	}

	// new grpc server
	rpcSrv = cometgrpc.New(g.Conf.RPCServer, srv)

	wg.Wrap(func() {
		cometgrpc.Start(g.Conf.RPCServer, rpcSrv, errc)
	})

	//prometheus mertics
	wg.Wrap(func() {
		mux := http.NewServeMux()

		mux.Handle("/metrics", promhttp.Handler())

		mux.HandleFunc("/check", healthCheck)

		g.Logger.Infof("start metrics server of prometheus listen: %s", g.Conf.MetricsServer.Addr)

		errc <- http.ListenAndServe(g.Conf.MetricsServer.Addr, mux)
	})

	// Interrupt handler.
	go handlers.InterruptHandler(srv, rpcSrv, errc)

	// Run!
	g.Logger.Infof("goim-comet exit", <-errc)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
