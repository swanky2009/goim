package grpc

import (
	"context"
	"net"
	"time"

	pb "github.com/swanky2009/goim/grpc/logic"
	"github.com/swanky2009/goim/logic"
	"github.com/swanky2009/goim/logic/g"
	"github.com/swanky2009/goim/logic/g/conf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// New logic grpc server
func New(c *conf.RPCServer, s *logic.Server) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)

	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())

	//初始化注入点
	endpoints = NewEndpoints(s)

	pb.RegisterLogicServer(srv, &server{s})

	return srv
}

func Start(c *conf.RPCServer, s *grpc.Server, errc chan error) {
	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		errc <- err
		return
	}
	go func() {
		errc <- s.Serve(lis)
	}()
	g.Logger.Infof("start rpc server listen: %s", c.Addr)
	//注册服务
	g.ServiceRegistrar.Register()

	return
}

type server struct {
	srv *logic.Server
}

var _ pb.LogicServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingReply, error) {
	return &pb.PingReply{}, nil
}

// Close Service
func (s *server) Close(ctx context.Context, req *pb.CloseReq) (*pb.CloseReply, error) {
	return &pb.CloseReply{}, nil
}

// Connect connect a conn.
func (s *server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {

	//call middleware func add Zipkin,ratelimit,circuitbreaker
	resp, err := endpoints.ConnectEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ConnectReply), nil

	// mid, key, room, platform, accepts, err := s.srv.Connect(ctx, req.Server, req.ServerKey, req.Cookie, req.Token)
	// if err != nil {
	// 	return &pb.ConnectReply{}, err
	// }
	// return &pb.ConnectReply{Mid: mid, Key: key, RoomID: room, Accepts: accepts, Platform: platform}, nil
}

// Disconnect disconnect a conn.
func (s *server) Disconnect(ctx context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.srv.Disconnect(ctx, req.Mid, req.Key, req.Server)
	if err != nil {
		return &pb.DisconnectReply{}, err
	}
	return &pb.DisconnectReply{Has: has}, nil
}

// Heartbeat beartbeat a conn.
func (s *server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.srv.Heartbeat(ctx, req.Mid, req.Key, req.Server); err != nil {
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

// RenewOnline renew server online.
func (s *server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.srv.RenewOnline(ctx, req.Server, req.RoomCount)
	if err != nil {
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}

// Receive receive a message.
func (s *server) Receive(ctx context.Context, req *pb.ReceiveReq) (*pb.ReceiveReply, error) {
	if err := s.srv.Receive(ctx, req.Mid, req.Room, req.Op, req.Msg); err != nil {
		return &pb.ReceiveReply{}, err
	}
	return &pb.ReceiveReply{}, nil
}
