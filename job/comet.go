package job

import (
	"context"
	"sync/atomic"
	"time"

	pb "github.com/swanky2009/goim/grpc/comet"
	"github.com/swanky2009/goim/job/g"
	"github.com/swanky2009/goim/job/g/conf"
	"google.golang.org/grpc"
)

const (
	syncCometServersDelay = 10 * time.Minute
)

// Comet is a comet.
type Comet struct {
	serverID      string
	client        pb.CometClient
	pushChan      []chan *pb.PushMsgReq
	roomChan      []chan *pb.BroadcastRoomReq
	broadcastChan chan *pb.BroadcastReq
	pushChanNum   uint64
	roomChanNum   uint64
	routineSize   uint64

	ctx    context.Context
	cancel context.CancelFunc
}

// NewComet new a comet.
func NewComet(c *conf.Comet, addr string) *Comet {
	cmt := &Comet{
		serverID:      addr,
		client:        newCometClient(c, addr),
		pushChan:      make([]chan *pb.PushMsgReq, c.RoutineSize),
		roomChan:      make([]chan *pb.BroadcastRoomReq, c.RoutineSize),
		broadcastChan: make(chan *pb.BroadcastReq, c.RoutineSize),
		routineSize:   uint64(c.RoutineSize),
	}
	cmt.ctx, cmt.cancel = context.WithCancel(context.Background())

	for i := 0; i < c.RoutineSize; i++ {
		cmt.pushChan[i] = make(chan *pb.PushMsgReq, c.RoutineChan)
		cmt.roomChan[i] = make(chan *pb.BroadcastRoomReq, c.RoutineChan)
		go cmt.process(cmt.pushChan[i], cmt.roomChan[i], cmt.broadcastChan)
	}
	return cmt
}

func newCometClient(c *conf.Comet, target string) pb.CometClient {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(time.Duration(c.Timeout)),
		//grpc.WithCompressor(grpc.NewGZIPCompressor()),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		panic(err)
	}
	return pb.NewCometClient(conn)
}

// Push push a user message.
func (c *Comet) Push(arg *pb.PushMsgReq) (err error) {
	idx := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	c.pushChan[idx] <- arg
	return
}

// BroadcastRoom broadcast a room message.
func (c *Comet) BroadcastRoom(arg *pb.BroadcastRoomReq) (err error) {
	idx := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	c.roomChan[idx] <- arg
	return
}

// Broadcast broadcast a message.
func (c *Comet) Broadcast(arg *pb.BroadcastReq) (err error) {
	c.broadcastChan <- arg
	return
}

func (c *Comet) process(pushChan chan *pb.PushMsgReq, roomChan chan *pb.BroadcastRoomReq, broadcastChan chan *pb.BroadcastReq) {
	var err error
	for {
		select {
		case broadcastArg := <-broadcastChan:
			_, err = c.client.Broadcast(context.Background(), &pb.BroadcastReq{
				Proto:    broadcastArg.Proto,
				ProtoOp:  broadcastArg.ProtoOp,
				Speed:    broadcastArg.Speed,
				Platform: broadcastArg.Platform,
			})
			if err != nil {
				g.Logger.Errorf("c.client.Broadcast(%v, reply) serverId:%s error(%v)", broadcastArg, c.serverID, err)
			}
			g.Logger.Infof("c.client.Broadcast(%v, reply) serverId:%s", broadcastArg, c.serverID)
		case roomArg := <-roomChan:
			_, err = c.client.BroadcastRoom(context.Background(), &pb.BroadcastRoomReq{
				RoomID: roomArg.RoomID,
				Proto:  roomArg.Proto,
			})
			if err != nil {
				g.Logger.Errorf("c.client.BroadcastRoom(%v, reply) serverId:%s error(%v)", roomArg, c.serverID, err)
			}
			g.Logger.Infof("c.client.BroadcastRoom(%v, reply) serverId:%s", roomArg, c.serverID)
		case pushArg := <-pushChan:
			_, err = c.client.PushMsg(context.Background(), &pb.PushMsgReq{
				Keys:    pushArg.Keys,
				Proto:   pushArg.Proto,
				ProtoOp: pushArg.ProtoOp,
			})
			if err != nil {
				g.Logger.Errorf("c.client.PushMsg(%v, reply) serverId:%s error(%v)", pushArg, c.serverID, err)
			}
			g.Logger.Infof("c.client.PushMsg(%v, reply) serverId:%s", pushArg, c.serverID)
		case <-c.ctx.Done():
			return
		}
	}
}

// Close close the resouces.
func (c *Comet) Close() {
	finish := make(chan bool)
	go func() {
		for {
			n := len(c.broadcastChan)
			for _, ch := range c.pushChan {
				n += len(ch)
			}
			for _, ch := range c.roomChan {
				n += len(ch)
			}
			if n == 0 {
				finish <- true
				return
			}
			time.Sleep(time.Second)
		}
	}()
	select {
	case <-finish:
		g.Logger.Info("close comet finish")
	case <-time.After(5 * time.Second):
		g.Logger.Errorf("close comet(server:%s push:%d room:%d broadcast:%d) timeout", c.serverID, len(c.pushChan), len(c.roomChan), len(c.broadcastChan))
	}
	c.cancel()
	return
}

type Comets struct {
	cometServiceMap map[string]*Comet
}

func (this *Comets) InitComets(c *conf.Comet) {

	this.cometServiceMap = make(map[int32]*Comet)

	state := g.ServiceInstancer.GetState()

	if state.Err != nil {
		if err != nil {
			panic(err)
		}
	}

	for _, addr := range state.Instances {

		g.Logger.Debugf("rpc addr: %s", addr)

		this.cometServiceMap[addr] = NewComet(c, addr)

		g.Logger.Infof("init comet rpc: %s", addr)
	}
	go this.SyncComets(c)

	//room info
	this.MergeRoomServers()

	go this.SyncRoomServers()

	return
}

func (this *Comets) SyncComets(c *conf.Comet) {
	for {
		state := ServiceInstancer.GetState()

		if state.Err != nil {
			if err != nil {
				panic(err)
			}
		}
		addrs := make(map[string]string)

		for _, addr := range state.Instances {
			if _, ok = cometServiceMap[addr]; !ok {
				this.cometServiceMap[addr] = NewComet(c, addr)
				g.Logger.Infof("init new comet rpc: %s", addr)
			}
			addrs[addr] = addr
		}
		for serverId, comet := range cometServiceMap {
			if _, ok = addrs[serverId]; !ok {
				comet.Close()
				delete(cometServiceMap, comet)
			}
		}
		time.Sleep(syncCometServersDelay)
	}
}

// push a message to a batch of subkeys
func (this *Comets) Push(serverId string, args *pb.PushMsgReq) {

	if c, ok := cometServiceMap[serverId]; ok {

		if err := c.Push(&args); err != nil {

			log.Error("c.Push(%v) serverId:%s error(%v)", args, serverId, err)

			//MetricsStat.IncrPushMsgFailed()
		}
	}
	MetricsStat.IncrPushMsg()
}

// broadcast a message to all
func (this *Comets) Broadcast(args *pb.BroadcastReq) {

	for serverId, c := range cometServiceMap {

		if err := c.Broadcast(&args); err != nil {

			log.Error("c.Broadcast(%v) serverId:%d error(%v)", args, serverId, err)

			//MetricsStat.IncrBroadcastMsgFailed()
		}
	}
	MetricsStat.IncrBroadcastMsg()
}

// broadcast aggregation messages to room
func (this *Comets) BroadcastRoom(args *pb.BroadcastRoomReq) {
	var (
		c        *Comet
		serverId int32
		servers  map[int32]struct{}
		ok       bool
		err      error
	)

	// if servers, ok = RoomServersMap[roomId]; ok {

	// 	for serverId, _ = range servers {

	// 		if c, ok = cometServiceMap[serverId]; ok {

	// 			// push routines
	// 			if err = c.BroadcastRoom(&args); err != nil {

	// 				log.Error("c.BroadcastRoom(%v) roomId:%d error(%v)", args, roomId, err)

	// 				//MetricsStat.IncrBroadcastRoomMsgFailed()
	// 			}
	// 		}
	// 	}
	//}
	MetricsStat.IncrBroadcastRoomMsg()
}
