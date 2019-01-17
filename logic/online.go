package logic

import (
	"context"
	"strings"

	"github.com/swanky2009/goim/logic/g"
)

// OnlineTop get the top online address.
func (l *Server) OnlineTop(c context.Context, typeStr string, n int64) (addrs []string, err error) {
	var (
		sids  []string
		metas map[string]map[string]string
		binds string
		ports []string
		addr  string
	)
	if sids, err = l.dao.ServersRank(c, n); err != nil {
		g.Logger.Errorf("ServersRank error(%v)", err)
		return
	}
	//get consul metas
	if metas, err = g.GetCometServiceMetas(); err != nil {
		g.Logger.Errorf("GetCometService error(%v)", err)
		return
	}

	g.Logger.Debugf("GetCometServiceMetas (%v)", metas)

	for _, sid := range sids {
		if meta, ok := metas[sid]; ok {
			if typeStr == "tcp" {
				binds = meta["tcp"]
			} else if typeStr == "ws" {
				binds = meta["ws"]
			} else if typeStr == "wstls" {
				binds = meta["wstls"]
			}
			ports = strings.Split(binds, ",")
			for _, addr = range ports {
				if len(addr) > 0 {
					addrs = append(addrs, addr)
				}
			}
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
