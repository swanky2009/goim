package logic

import (
	"context"
	"strings"

	"github.com/swanky2009/goim/logic/g"
)

// OnlineTop get the top online address.
func (l *Server) OnlineTop(c context.Context, typeStr string, n int64) (addrs []string, err error) {
	var (
		sids    []string
		servers map[string]string
	)
	if sids, err = l.dao.ServersRank(c, n); err != nil {
		g.Logger.Errorf("ServersRank error(%v)", err)
		return
	}
	//get consul addrs
	if servers, err = g.GetCometService(); err != nil {
		g.Logger.Errorf("GetCometService error(%v)", err)
		return
	}
	for _, sid := range sids {
		if addr, ok := servers[sid]; ok {
			if typeStr == "tcp" {
				addr = strings.Replace(addr, ":8009", ":8001", 1)
			} else {
				addr = strings.Replace(addr, ":8009", ":8002", 1)
			}
			addrs = append(addrs, addr)
		}
	}
	return
}

// OnlineRoom get rooms online.
func (l *Server) OnlineRoom(c context.Context, rooms []string) (res map[string]int32, err error) {

	if res, err = l.dao.GetAllRoomCount(c); err != nil {
		g.Logger.Errorf("GetAllRoomCount error(%v)", err)
		return
	}
	// if len(rooms) > 0 {
	// 	res = make(map[string]int32, len(rooms))
	// 	for _, roomID := range rooms {
	// 		res[roomID] = roomCount[roomID]
	// 	}
	// } else {
	// 	res = roomCount
	// }
	return
}
