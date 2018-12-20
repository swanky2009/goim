package logic

import (
	"context"
	"time"

	"github.com/swanky2009/goim/logic/dao"
	"github.com/swanky2009/goim/logic/g/conf"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5
)

// Logic Server struct
type Server struct {
	c   *conf.Config
	dao *dao.Dao
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
	regions    map[string]string // province -> region
}

// New server
func NewServer(c *conf.Config) (l *Server) {
	l = &Server{
		c:       c,
		dao:     dao.New(c),
		regions: make(map[string]string),
	}
	//l.initRegions()
	// l.loadOnline()
	// go l.onlineproc()
	return l
}

// Ping ping resources is ok.
func (l *Server) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Server) Close() {
	l.dao.Close()
}

func (l *Server) initRegions() {
	for region, ps := range l.c.Regions {
		for _, province := range ps {
			l.regions[province] = region
		}
	}
}

// func (l *Server) onlineproc() {
// 	for {
// 		time.Sleep(_onlineTick)
// 		if err := l.loadOnline(); err != nil {
// 			g.Logger.Errorf("onlineproc error(%v)", err)
// 		}
// 	}
// }

// func (l *Server) loadOnline() (err error) {
// 	var (
// 		roomCount = make(map[string]int32)
// 	)
// 	for _, server := range l.nodes {
// 		var online *model.Online
// 		online, err = l.dao.ServerOnline(context.Background(), server.Hostname)
// 		if err != nil {
// 			return
// 		}
// 		if time.Since(time.Unix(online.Updated, 0)) > _onlineDeadline {
// 			l.dao.DelServerOnline(context.Background(), server.Hostname)
// 			continue
// 		}
// 		for roomID, count := range online.RoomCount {
// 			roomCount[roomID] += count
// 		}
// 	}
// 	l.roomCount = roomCount
// 	return
// }
