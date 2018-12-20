package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swanky2009/goim/job"
	"github.com/swanky2009/goim/job/g"
	"github.com/swanky2009/goim/job/handlers"
	"github.com/swanky2009/goim/pkg/waitgroup"
)

func main() {
	var (
		wg  waitgroup.WaitGroupWrapper
		srv *job.Job
	)
	// Mechanical domain.
	errc := make(chan error)

	// new job server
	srv = job.New(g.Conf)
	wg.Wrap(func() {
		srv.Consume()
	})

	//prometheus mertics
	wg.Wrap(func() {
		mux := http.NewServeMux()

		mux.Handle("/metrics", promhttp.Handler())

		g.Logger.Infof("start metrics server of prometheus listen: %s", g.Conf.MetricsServer.Addr)

		errc <- http.ListenAndServe(g.Conf.MetricsServer.Addr, mux)
	})

	// Interrupt handler.
	go handlers.InterruptHandler(srv, errc)

	// Run!
	g.Logger.Infof("goim-job exit", <-errc)
}
