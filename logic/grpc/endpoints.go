package grpc

import (
	"context"
	//"fmt"

	"github.com/go-kit/kit/endpoint"

	pb "github.com/swanky2009/goim/grpc/logic"
	"github.com/swanky2009/goim/logic"
	"github.com/swanky2009/goim/logic/handlers"
)

var (
	endpoints Endpoints
)

type Endpoints struct {
	ConnectEndpoint endpoint.Endpoint
}

func NewEndpoints(s *logic.Server) Endpoints {

	connectEndpoint := MakeConnectEndpoint(s)

	connectEndpoint = handlers.WrapEndpoints(connectEndpoint)

	return Endpoints{
		ConnectEndpoint: connectEndpoint,
	}
}

// Make Endpoints
func MakeConnectEndpoint(s *logic.Server) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.ConnectReq)
		mid, key, room, platform, accepts, err := s.Connect(ctx, req.Server, req.ServerKey, req.Cookie, req.Token)
		if err != nil {
			return &pb.ConnectReply{}, err
		}
		return &pb.ConnectReply{Mid: mid, Key: key, RoomID: room, Accepts: accepts, Platform: platform}, nil
	}
}
