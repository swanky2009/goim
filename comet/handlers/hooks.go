package handlers

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/swanky2009/goim/comet"
	"github.com/swanky2009/goim/comet/g"
	"google.golang.org/grpc"
)

func InterruptHandler(srv *comet.Server, rpcSrv *grpc.Server, errc chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	terminateError := fmt.Errorf("%s", <-c)

	//Place whatever shutdown handling you want here
	g.ServiceRegistrar.Deregister() //注销服务
	rpcSrv.GracefulStop()
	srv.Close()

	errc <- terminateError
}
