package logic

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/swanky2009/goim/grpc/comet"
	"github.com/swanky2009/goim/logic/g"
	xstr "github.com/swanky2009/goim/pkg/strings"
)

// Connect connected a conn.
func (l *Server) Connect(c context.Context, server, serverKey, cookie string, token []byte) (mid int64, key, roomID string, paltform string, accepts []int32, err error) {
	// TODO test example: mid|key|roomid|platform|accepts
	params := strings.Split(string(token), "|")
	if len(params) != 5 {
		err = fmt.Errorf("invalid token:%s", token)
		return
	}
	if mid, err = strconv.ParseInt(params[0], 10, 64); err != nil {
		return
	}
	key = params[1]
	roomID = params[2]
	paltform = params[3]
	if accepts, err = xstr.SplitInt32s(params[4], ","); err != nil {
		return
	}
	if err = l.dao.AddMapping(c, mid, key, server); err != nil {
		g.Logger.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if err = l.dao.IncrServerScore(c, server); err != nil {
		g.Logger.Errorf("l.dao.IncrServerScore(%s) error(%v)", server, err)
		return
	}
	g.Logger.Infof("conn connected key:%s server:%s mid:%d token:%s", key, server, mid, token)
	return
}

// Disconnect disconnect a conn.
func (l *Server) Disconnect(c context.Context, mid int64, key, server string) (has bool, err error) {
	if has, err = l.dao.DelMapping(c, mid, key, server); err != nil {
		g.Logger.Errorf("l.dao.DelMapping(%d,%s) error(%v)", mid, key, server)
		return
	}
	if err = l.dao.DecrServerScore(c, server); err != nil {
		g.Logger.Errorf("l.dao.DecrServerScore(%s) error(%v)", server, err)
		return
	}
	g.Logger.Infof("conn disconnected key:%s server:%s mid:%d", key, server, mid)
	return
}

// Heartbeat heartbeat a conn.
func (l *Server) Heartbeat(c context.Context, mid int64, key, server string) (err error) {
	has, err := l.dao.ExpireMapping(c, mid, key)
	if err != nil {
		g.Logger.Errorf("l.dao.ExpireMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if !has {
		if err = l.dao.AddMapping(c, mid, key, server); err != nil {
			g.Logger.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	g.Logger.Infof("conn heartbeat key:%s server:%s mid:%d", key, server, mid)
	return
}

// RenewOnline renew a server online.
func (l *Server) RenewOnline(c context.Context, server string, roomCount map[string]int32) (allRoomCount map[string]int32, err error) {

	if err = l.dao.UpdateRoomCount(c, server, roomCount); err != nil {
		g.Logger.Errorf("l.dao.UpdateRoomCount(%s) error(%v)", server, err)
		return
	}

	g.Logger.Debugf("RenewOnline server:%s roomCount(%v)", server, roomCount)

	if err = l.dao.AddServerScore(c, server); err != nil {
		g.Logger.Errorf("l.dao.AddServerScore(%s) error(%v)", server, err)
		return
	}
	allRoomCount, err = l.dao.GetAllRoomCount(c)

	g.Logger.Debugf("RenewOnline server:%s allRoomCount(%v)", server, allRoomCount)

	return
}

// Receive receive a message.
func (l *Server) Receive(c context.Context, mid int64, room string, op int32, msg []byte) (err error) {
	// TODO upstream message
	g.Logger.Debugf("conn receive a message mid:%d room:%s msg:%s", mid, room, string(msg))

	if op == pb.OpSendMsg {
		err = l.PushRoom(c, pb.OpSendMsgReply, room, msg)
		if err != nil {
			g.Logger.Warningf("push room mid:%d room:%s error(%v)", mid, room, err)
		}
	}
	return
}
