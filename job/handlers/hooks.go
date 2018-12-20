package handlers

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/swanky2009/goim/job"
)

func InterruptHandler(srv *job.Job, errc chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	terminateError := fmt.Errorf("%s", <-c)

	//Place whatever shutdown handling you want here
	srv.Close()

	errc <- terminateError
}
