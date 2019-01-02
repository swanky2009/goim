package comet

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/swanky2009/goim/comet/g"
	"github.com/swanky2009/goim/comet/g/conf"
	"github.com/swanky2009/goim/grpc/logic"
	"github.com/swanky2009/goim/pkg/hash"
	"github.com/swanky2009/goim/pkg/ip"
	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

var (
	_maxInt = 1<<31 - 1
)

const (
	_clientHeartbeat       = time.Second * 90
	_minSrvHeartbeatSecond = 600  // 10m
	_maxSrvHeartbeatSecond = 1200 // 20m
)

// Server .
type Server struct {
	c         *conf.Config
	round     *Round    // accept round store
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32

	serverID  string
	rpcClient logic.LogicClient
}

func newLogicClient(c *conf.RPCClient) logic.LogicClient {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBalancer(grpc.RoundRobin(g.ServiceResolver)),
		grpc.WithTimeout(time.Duration(c.Timeout)),
		//grpc.WithCompressor(grpc.NewGZIPCompressor()),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		panic(err)
	}
	return logic.NewLogicClient(conn)
}

// serverID sha1(host:port)
func getServerID(c *conf.RPCServer) string {
	host, port, err := net.SplitHostPort(c.Addr)
	if err != nil {
		panic(err)
	}
	if host == "" {
		host = ip.InternalIP()
	}
	return hash.Sha1s(fmt.Sprintf("%s:%s", host, port))
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:         c,
		round:     NewRound(c),
		rpcClient: newLogicClient(c.RPCClient),
		serverID:  getServerID(c.RPCServer),
	}

	// init bucket
	s.buckets = make([]*Bucket, c.Bucket.Size)

	s.bucketIdx = uint32(c.Bucket.Size)

	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}

	go s.onlineproc()
	return s
}

// Buckets return all buckets.
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// Bucket get the bucket by subkey.
func (s *Server) Bucket(subKey string) *Bucket {

	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx

	g.Logger.Debugf("%s hit channel bucket index: %d use cityhash", subKey, idx)

	return s.buckets[idx]
}

// NextKey generate a server key.
func (s *Server) NextKey() string {
	return fmt.Sprintf("%s-%d", s.serverID, time.Now().UnixNano())
}

// RandServerHearbeat rand server heartbeat.
func (s *Server) RandServerHearbeat() time.Duration {
	return time.Duration(_minSrvHeartbeatSecond+rand.Intn(_maxSrvHeartbeatSecond-_minSrvHeartbeatSecond)) * time.Second
}

// Close close the server.
func (s *Server) Close() (err error) {
	return
}

func (s *Server) onlineproc() {
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}
		if allRoomsCount, err = s.RenewOnline(s.serverID, roomCount); err != nil {
			time.Sleep(time.Duration(s.c.OnlineTick))
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Duration(s.c.OnlineTick))
	}
}
