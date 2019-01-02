package job

import (
	"time"

	pb "github.com/swanky2009/goim/grpc/comet"
	"github.com/swanky2009/goim/job/g"
	"github.com/swanky2009/goim/job/g/conf"
)

const (
	syncCometServersDelay = 10 * time.Minute
	syncRoomServersDelay  = 1 * time.Second
)

type Comets struct {
	cometServiceMap map[string]*Comet
	roomServersMap  map[string]map[string]struct{} // roomid:servers
}

func InitComets(c *conf.Comet) *Comets {
	comets := &Comets{
		cometServiceMap: make(map[string]*Comet),
		roomServersMap:  make(map[string]map[string]struct{}),
	}

	state := g.ServiceInstancer.GetState()

	if state.Err != nil {
		panic(state.Err)
	}

	for _, addr := range state.Instances {

		g.Logger.Debugf("rpc addr: %s", addr)

		comets.cometServiceMap[addr] = NewComet(c, addr)

		g.Logger.Infof("init comet rpc: %s", addr)
	}
	go comets.SyncComets(c)

	//room info
	comets.MergeRoomServers()

	go comets.SyncRoomServers()

	return comets
}

func (this *Comets) SyncComets(c *conf.Comet) {
	for {
		state := g.ServiceInstancer.GetState()

		if state.Err != nil {
			g.Logger.Warnf("get comet rpc services error(%v)", state.Err)
		}
		addrs := make(map[string]string)

		for _, addr := range state.Instances {
			if _, ok := this.cometServiceMap[addr]; !ok {
				this.cometServiceMap[addr] = NewComet(c, addr)
				g.Logger.Infof("init new comet rpc: %s", addr)
			}
			addrs[addr] = addr
		}
		for serverId, comet := range this.cometServiceMap {
			if _, ok := addrs[serverId]; !ok {
				comet.Close()
				delete(this.cometServiceMap, serverId)
			}
		}
		time.Sleep(syncCometServersDelay)
	}
}

// push a message to a batch of subkeys
func (this *Comets) Push(serverId string, args *pb.PushMsgReq) {

	if c, ok := this.cometServiceMap[serverId]; ok {

		if err := c.Push(args); err != nil {

			g.Logger.Errorf("c.Push(%v) serverId:%s error(%v)", args, serverId, err)

			//MetricsStat.IncrPushMsgFailed()
		}
	}
	//g.MetricsStat.IncrPushMsg()
}

// broadcast a message to all
func (this *Comets) Broadcast(args *pb.BroadcastReq) {

	for serverId, c := range this.cometServiceMap {

		if err := c.Broadcast(args); err != nil {

			g.Logger.Errorf("c.Broadcast(%v) serverId:%d error(%v)", args, serverId, err)

			//MetricsStat.IncrBroadcastMsgFailed()
		}
	}
	//g.MetricsStat.IncrBroadcastMsg()
}

// broadcast aggregation messages to room
func (this *Comets) BroadcastRoom(roomId string, args *pb.BroadcastRoomReq) {
	var (
		c        *Comet
		serverId string
		servers  map[string]struct{}
		ok       bool
		err      error
	)

	if servers, ok = this.roomServersMap[roomId]; ok {

		for serverId, _ = range servers {

			if c, ok = this.cometServiceMap[serverId]; ok {

				// push routines
				if err = c.BroadcastRoom(args); err != nil {

					g.Logger.Errorf("c.BroadcastRoom(%v) roomId:%s error(%v)", args, roomId, err)

					//MetricsStat.IncrBroadcastRoomMsgFailed()
				}
			}
		}
	}
	//g.MetricsStat.IncrBroadcastRoomMsg()
}

func (this *Comets) MergeRoomServers() {

	var (
		c *Comet

		ok bool

		roomId string

		serverId string

		roomIds map[string]bool

		servers map[string]struct{}

		roomServers = make(map[string]map[string]struct{})
	)

	// all comet nodes
	for serverId, c = range this.cometServiceMap {

		if roomIds = c.GetRooms(); roomIds != nil {

			// merge room's servers
			for roomId, _ = range roomIds {

				if servers, ok = roomServers[roomId]; !ok {

					servers = make(map[string]struct{})

					roomServers[roomId] = servers
				}
				servers[serverId] = struct{}{}
			}
		}
	}
	this.roomServersMap = roomServers
}

func (this *Comets) SyncRoomServers() {
	for {
		this.MergeRoomServers()

		time.Sleep(syncRoomServersDelay)
	}
}

func (this *Comets) Close() {
	for _, c := range this.cometServiceMap {
		c.Close()
	}
}