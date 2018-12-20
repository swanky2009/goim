package comet

import (
	"context"
	"github.com/swanky2009/goim/comet/g"
	model "github.com/swanky2009/goim/grpc/comet"
	logic "github.com/swanky2009/goim/grpc/logic"
	"github.com/swanky2009/goim/pkg/strings"
)

// Connect .
func (s *Server) Connect(p *model.Proto, cookie string) (mid int64, key, rid, platform string, accepts []int32, err error) {
	var (
		reply *logic.ConnectReply
	)
	if reply, err = s.rpcClient.Connect(context.Background(), &logic.ConnectReq{
		Server:    s.serverID,
		ServerKey: s.NextKey(),
		Cookie:    cookie,
		Token:     p.Body,
	}); err != nil {
		return
	}
	return reply.Mid, reply.Key, reply.RoomID, reply.Platform, reply.Accepts, nil
}

// Disconnect .
func (s *Server) Disconnect(mid int64, key string) (err error) {
	_, err = s.rpcClient.Disconnect(context.Background(), &logic.DisconnectReq{
		Mid:    mid,
		Server: s.serverID,
		Key:    key,
	})
	return
}

// Heartbeat .
func (s *Server) Heartbeat(mid int64, key string) (err error) {
	_, err = s.rpcClient.Heartbeat(context.Background(), &logic.HeartbeatReq{
		Mid:    mid,
		Server: s.serverID,
		Key:    key,
	})
	return
}

// RenewOnline .
func (s *Server) RenewOnline(serverID string, rommCount map[string]int32) (allRoom map[string]int32, err error) {
	var (
		reply *logic.OnlineReply
	)
	if reply, err = s.rpcClient.RenewOnline(context.Background(), &logic.OnlineReq{
		Server:    s.serverID,
		RoomCount: rommCount,
	}); err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

// Report message to logic.
func (s *Server) Report(mid int64, proto *model.Proto) (rp *model.Proto, err error) {
	if _, err = s.rpcClient.Receive(context.Background(), &logic.ReceiveReq{
		Mid: mid,
	}); err != nil {
		return
	}
	return nil, nil
}

// Operate .
func (s *Server) Operate(p *model.Proto, ch *Channel, b *Bucket) (err error) {
	switch {
	case p.Op >= model.MinBusinessOp && p.Op <= model.MaxBusinessOp:
		// TODO report a message
		_, err = s.Report(ch.Mid, p)
		if err != nil {
			return
		}
		p.Body = nil
	case p.Op == model.OpChangeRoom:
		err = b.ChangeRoom(string(p.Body), ch)
		p.Op = model.OpChangeRoomReply
	case p.Op == model.OpRegister:
		ops, err := strings.SplitInt32s(string(p.Body), ",")
		if err == nil {
			ch.Watch(ops...)
		}
		p.Op = model.OpRegisterReply
	case p.Op == model.OpUnregister:
		ops, err := strings.SplitInt32s(string(p.Body), ",")
		if err == nil {
			ch.UnWatch(ops...)
		}
		p.Op = model.OpUnregisterReply
	default:
		err = g.ErrOperation
	}
	if err != nil {
		g.Logger.Errorf("operate(%+v) error(%v)", p, err)
	}
	return
}
