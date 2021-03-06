package logic

import (
	"context"

	"github.com/swanky2009/goim/logic/g"
)

// PushKeys push a message by keys.
func (l *Server) PushKeys(c context.Context, op int32, keys []string, msg []byte) (err error) {
	servers, err := l.dao.ServersByKeys(c, keys)
	if err != nil {
		g.Logger.Errorf("dao.ServersByKeys error(%v)", err)
		return
	}

	g.Logger.Debugf("dao.ServersByKeys servers(%v)", servers)

	pushKeys := make(map[string][]string)
	for i, key := range keys {
		server := servers[i]
		if server != "" && key != "" {
			pushKeys[server] = append(pushKeys[server], key)
		}
	}
	for server := range pushKeys {
		if err = l.dao.PushMsg(c, op, server, pushKeys[server], msg); err != nil {
			g.Logger.Errorf("dao.PushMsg error(%v)", err)
			return
		}
	}
	return
}

// PushMids push a message by mid.
func (l *Server) PushMids(c context.Context, op int32, mids []int64, msg []byte) (err error) {
	keyServers, _, err := l.dao.KeysByMids(c, mids)
	if err != nil {
		return
	}
	keys := make(map[string][]string)
	for key, server := range keyServers {
		if key == "" || server == "" {
			g.Logger.Warningf("push key:%s server:%s is empty", key, server)
			continue
		}
		keys[server] = append(keys[server], key)
	}
	for server, keys := range keys {
		if err = l.dao.PushMsg(c, op, server, keys, msg); err != nil {
			return
		}
	}
	return
}

// PushRoom push a message by room.
func (l *Server) PushRoom(c context.Context, op int32, room string, msg []byte) (err error) {
	return l.dao.BroadcastRoomMsg(c, op, room, msg)
}

// PushAll push a message to all.
func (l *Server) PushAll(c context.Context, op, speed int32, platform string, msg []byte) (err error) {
	return l.dao.BroadcastMsg(c, op, speed, platform, msg)
}
