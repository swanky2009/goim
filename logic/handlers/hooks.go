package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/swanky2009/goim/logic"
	"github.com/swanky2009/goim/logic/g"
	"google.golang.org/grpc"
)

func InterruptHandler(srv *logic.Server, rpcSrv *grpc.Server, httpSrv *http.Server, errc chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	terminateError := fmt.Errorf("%s", <-c)

	//Place whatever shutdown handling you want here
	g.ServiceRegistrar.Deregister() //注销服务
	httpSrv.Shutdown(context.TODO())
	rpcSrv.GracefulStop()
	srv.Close()

	errc <- terminateError
}
