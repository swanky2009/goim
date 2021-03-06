package main

// Start Commond eg: ./client 1 5000 localhost:8080
// first parameter：beginning userId
// second parameter: amount of clients
// third parameter: comet server ip

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	grpc "github.com/swanky2009/goim/grpc/comet"
	"github.com/swanky2009/goim/pkg/bufio"
)

const (
	OP_HANDSHARE        = int32(0)
	OP_HANDSHARE_REPLY  = int32(1)
	OP_HEARTBEAT        = int32(2)
	OP_HEARTBEAT_REPLY  = int32(3)
	OP_SEND_SMS         = int32(4)
	OP_SEND_SMS_REPLY   = int32(5)
	OP_DISCONNECT_REPLY = int32(6)
	OP_AUTH             = int32(7)
	OP_AUTH_REPLY       = int32(8)
	OP_RAW_MSG          = int32(9)
	OP_TEST             = int32(254)
	OP_TEST_REPLY       = int32(255)
)

const (
	rawHeaderLen = uint16(16)
	heart        = 20 * time.Second //s
	msg          = 30 * time.Second //s
	roomid       = 1
	platform     = "pc"
	accepts      = "0,1,2,3,4,5,6,7,8,9,254,255"
)

var (
	countDown int64
	rpcserver string
)

type ServerBody struct {
	Code int      `json:"code"`
	Data []string `json:"data"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("http://%s/online/top?type=tcp", os.Args[3])

	body, err := httpGet(url)
	if err != nil {
		panic(err)
		return
	}

	log.Print(string(body))

	var servers ServerBody

	err = json.Unmarshal(body, &servers)
	if err != nil {
		panic(err)
		return
	}
	log.Print(servers)

	if len(servers.Data) == 0 {
		log.Println("no comet server")
		return
	}

	rpcserver = servers.Data[0]

	go result()

	for i := begin; i < begin+num; i++ {
		go startClient(fmt.Sprintf("%d", i))
	}

	var exit chan bool
	<-exit
}

func result() {
	var (
		lastTimes int64
		diff      int64
		nowCount  int64
		timer     = int64(30)
	)

	for {
		nowCount = atomic.LoadInt64(&countDown)
		diff = nowCount - lastTimes
		lastTimes = nowCount
		fmt.Println(fmt.Sprintf("%s down:%d down/s:%d", time.Now().Format("2006-01-02 15:04:05"), nowCount, diff/timer))
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func client(key string) {
	for {
		startClient(key)
		time.Sleep(3 * time.Second)
	}
}

func startClient(key string) {

	time.Sleep(time.Duration(mrand.Intn(1000)) * time.Millisecond)

	quit := make(chan bool, 1)

	conn, err := net.Dial("tcp", rpcserver)
	if err != nil {
		log.Printf("net.Dial(\"%s\") error(%v)", rpcserver, err)
		return
	}
	conn.SetReadDeadline(time.Now().Add(heart + 60*time.Second))

	wr := bufio.NewWriterSize(conn, 256)
	seqId := int32(0)
	proto := new(grpc.Proto)
	proto.Ver = 1
	// auth
	// test handshake timeout
	// time.Sleep(time.Second * 31)
	proto.Op = OP_AUTH
	proto.Seq = seqId

	// TODO test example: mid|key|roomid|platform|accepts
	mid := key
	body := fmt.Sprintf("%s|%s|%d|%s|%s", mid, key, roomid, platform, accepts)
	proto.Body = []byte(body)

	if err = proto.WriteTCP(wr); err != nil {
		log.Printf("WriteTCP Auth error(%v)", err)
		return
	}
	if err = wr.Flush(); err != nil {
		log.Printf("WriteTCP Auth error(%v)", err)
		return
	}

	seqId++
	// writer
	go func() {
		// heartbeat
		proto1 := new(grpc.Proto)
		proto1.Op = OP_HEARTBEAT
		proto1.Body = nil

		//msg
		proto2 := new(grpc.Proto)
		proto2.Op = OP_SEND_SMS
		proto2.Body = []byte("hello,everyone! I am " + mid)

		ticker := time.NewTicker(heart)
		ticker_msg := time.NewTicker(msg)
		for {
			select {
			case <-ticker.C:
				proto1.Seq = seqId
				if err = proto1.WriteTCPHeart(wr, roomid); err != nil {
					log.Printf("key:%s WriteTCPHeart() error(%v)", key, err)
					return
				}
				if err = wr.Flush(); err != nil {
					log.Printf("key:%s WriteTCPHeart() error(%v)", key, err)
					return
				}
				log.Printf("key:%s Write heartbeat", key)
				seqId++
			case <-ticker_msg.C:
				proto2.Seq = seqId
				if err = proto2.WriteTCP(wr); err != nil {
					log.Printf("key:%s WriteTCP() error(%v)", key, err)
					return
				}
				if err = wr.Flush(); err != nil {
					log.Printf("key:%s WriteTCP() error(%v)", key, err)
					return
				}
				log.Printf("key:%s Write send msg(%v)", key, proto2)
				seqId++
			default:
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	// reader
	go func() {
		rd := bufio.NewReaderSize(conn, 256)
		for {
			if err = proto.ReadTCP(rd); err != nil {
				if err == io.EOF {
					time.Sleep(200 * time.Millisecond)
					continue
				}
				log.Printf("key:%s tcpReadProto error(%v)", key, err)
				quit <- true
				return
			}

			log.Printf("key:%s proto Operation: %d", key, proto.Op)

			if proto.Op == OP_AUTH_REPLY {
				log.Printf("key:%s auth ok, proto: %v", key, proto)
			} else if proto.Op == OP_HEARTBEAT_REPLY {
				log.Printf("key:%s receive heartbeat", key)
				if err = conn.SetReadDeadline(time.Now().Add(heart + 60*time.Second)); err != nil {
					log.Printf("conn.SetReadDeadline() error(%v)", err)
					quit <- true
					return
				}
			} else if proto.Op == OP_TEST_REPLY {
				log.Printf("key:%s reply msg: %s", key, string(proto.Body))
				atomic.AddInt64(&countDown, 1)
			} else if proto.Op == OP_SEND_SMS_REPLY {
				log.Printf("key:%s reply msg: %s", key, string(proto.Body))
				atomic.AddInt64(&countDown, 1)
				if err = conn.SetReadDeadline(time.Now().Add(heart + 60*time.Second)); err != nil {
					log.Printf("conn.SetReadDeadline() error(%v)", err)
					quit <- true
					return
				}
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	<-quit
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
