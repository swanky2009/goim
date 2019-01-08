package logic

import (
	"context"
	"time"

	"github.com/swanky2009/goim/logic/dao"
	"github.com/swanky2009/goim/logic/g"
	"github.com/swanky2009/goim/logic/g/conf"
)

const (
	_onlineTick = time.Second * 10
)

// Logic Server struct
type Server struct {
	c   *conf.Config
	dao *dao.Dao
}

// New server
func NewServer(c *conf.Config) (l *Server) {
	l = &Server{
		c:   c,
		dao: dao.New(c),
	}
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

func (l *Server) onlineproc() {
	for {
		time.Sleep(_onlineTick)
		if err := l.loadOnline(); err != nil {
			g.Logger.Errorf("onlineproc error(%v)", err)
		}
	}
}

func (l *Server) loadOnline() (err error) {
	var (
		sids    []string
		servers map[string]string
	)
	if sids, err = l.dao.ServersRank(context.TODO(), -1); err != nil {
		g.Logger.Errorf("ServersRank error(%v)", err)
		return
	}
	//get consul addrs
	if servers, err = g.GetCometService(); err != nil {
		g.Logger.Errorf("GetCometService error(%v)", err)
		return
	}
	for _, sid := range sids {
		if _, ok := servers[sid]; !ok {
			l.dao.DelServerScore(context.TODO(), sid)
		}
	}
	return
}
