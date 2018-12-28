package comet

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/swanky2009/goim/comet/g"
	grpc "github.com/swanky2009/goim/grpc/comet"
	"github.com/swanky2009/goim/pkg/bufio"
	"github.com/swanky2009/goim/pkg/bytes"
	xtime "github.com/swanky2009/goim/pkg/time"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(server *Server, addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			g.Logger.Errorf("net.ResolveTCPAddr(tcp4, %s) error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			g.Logger.Errorf("net.ListenTCP(tcp4, %s) error(%v)", bind, err)
			return
		}
		g.Logger.Infof("start tcp server listen: %s", bind)
		// split N core accept
		for i := 0; i < accept; i++ {
			go acceptTCP(server, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			g.Logger.Errorf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(server.c.TCP.Keepalive); err != nil {
			g.Logger.Errorf("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(server.c.TCP.Rcvbuf); err != nil {
			g.Logger.Errorf("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(server.c.TCP.Sndbuf); err != nil {
			g.Logger.Errorf("conn.SetWriteBuffer() error(%v)", err)
			return
		}
		go serveTCP(server, conn, r)
		if r++; r == _maxInt {
			r = 0
		}
	}
}

func serveTCP(s *Server, conn *net.TCPConn, r int) {
	var (
		// timer
		tr = s.round.Timer(r)
		rp = s.round.Reader(r)
		wp = s.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)

	g.Logger.Debugf("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)

	s.ServeTCP(conn, rp, wp, tr)
}

// ServeTCP .
func (s *Server) ServeTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *xtime.Timer) {
	var (
		err     error
		rid     string
		accepts []int32
		p       *grpc.Proto
		b       *Bucket
		trd     *xtime.TimerData
		lastHb  = time.Now()
		rb      = rp.Get()
		wb      = wp.Get()
		ch      = NewChannel(s.c.ProtoSection.CliProto, s.c.ProtoSection.SvrProto)
		rr      = &ch.Reader
		wr      = &ch.Writer
	)
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	// handshake
	step := 0
	trd = tr.Add(time.Duration(s.c.ProtoSection.HandshakeTimeout), func() {
		conn.Close()
		g.Logger.Errorf("key: %s remoteIP: %s step: %d tcp handshake timeout", ch.Key, conn.RemoteAddr().String(), step)
	})
	ch.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	// must not setadv, only used in auth
	step = 1
	if p, err = ch.CliProto.Set(); err == nil {
		if ch.Mid, ch.Key, rid, ch.Platform, accepts, err = s.authTCP(rr, wr, p); err == nil {
			ch.Watch(accepts...)
			b = s.Bucket(ch.Key)
			err = b.Put(rid, ch)

			g.Logger.Debugf("tcp connnected key:%s mid:%d proto:%+v", ch.Key, ch.Mid, p)
		}
	}
	step = 2
	if err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		g.Logger.Errorf("key: %s handshake failed error(%v)", ch.Key, err)
		return
	}
	trd.Key = ch.Key
	tr.Set(trd, _clientHeartbeat)
	step = 3
	// increase tcp stat
	g.StatMetrics.IncrTcpOnline()
	// hanshake ok start dispatch goroutine
	go s.dispatchTCP(conn, wr, wp, wb, ch)
	serverHeartbeat := s.RandServerHearbeat()
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			g.Logger.Errorf("key: %s tcp channel ring set error(%v)", ch.Key, err)
			break
		}
		g.Logger.Debugf("key: %s tcp start read proto", ch.Key)

		if err = p.ReadTCP(rr); err != nil {
			g.Logger.Errorf("key: %s tcp read proto error(%v)", ch.Key, err)
			break
		}
		g.Logger.Debugf("key: %s tcp end read proto:%v", ch.Key, p)

		if p.Op == grpc.OpHeartbeat {
			tr.Set(trd, _clientHeartbeat)
			p.Body = nil
			p.Op = grpc.OpHeartbeatReply
			// last server heartbeat
			if now := time.Now(); now.Sub(lastHb) > serverHeartbeat {
				if err = s.Heartbeat(ch.Mid, ch.Key); err == nil {
					lastHb = now
				} else {
					err = nil
				}
			}
			g.Logger.Debugf("tcp heartbeat receive key:%s, mid:%d", ch.Key, ch.Mid)
			step++
		} else {
			if err = s.Operate(p, ch, b); err != nil {
				g.Logger.Errorf("key: %s tcp operate error(%v)", ch.Key, err)
				break
			}
		}

		g.Logger.Debugf("key: %s process proto:%v", ch.Key, p)

		ch.CliProto.SetAdv()
		ch.Signal()

		g.Logger.Debugf("key: %s signal", ch.Key)
	}

	g.Logger.Debugf("key: %s server tcp error(%v)", ch.Key, err)

	if err != nil && err != io.EOF && !strings.Contains(err.Error(), "closed") {
		g.Logger.Errorf("key: %s server tcp failed error(%v)", ch.Key, err)
	}
	b.Del(ch)
	tr.Del(trd)
	rp.Put(rb)
	conn.Close()
	ch.Close()
	if err = s.Disconnect(ch.Mid, ch.Key); err != nil {
		g.Logger.Errorf("key: %s operator do disconnect error(%v)", ch.Key, err)
	}
	g.Logger.Debugf("tcp disconnected key: %s mid:%d", ch.Key, ch.Mid)
	// decrease tcp stat
	g.StatMetrics.DecrTcpOnline()
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (s *Server) dispatchTCP(conn *net.TCPConn, wr *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
		online int32
	)
	g.Logger.Debugf("key: %s start dispatch tcp goroutine", ch.Key)

	for {
		var p = ch.Ready()

		g.Logger.Debugf("key: %s dispatch msg: %v", ch.Key, *p)

		switch p {
		case grpc.ProtoFinish:
			g.Logger.Debugf("key: %s wakeup exit dispatch goroutine", ch.Key)
			finish = true
			goto failed
		case grpc.ProtoReady:
			// fetch message from svrbox(client send)
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					err = nil // must be empty error
					break
				}
				g.Logger.Debugf("key: %s start write client proto(%v)", ch.Key, p)

				if p.Op == grpc.OpHeartbeatReply {
					if ch.Room != nil {
						online = ch.Room.OnlineNum()
					}
					if err = p.WriteTCPHeart(wr, online); err != nil {
						goto failed
					}
				} else {
					if err = p.WriteTCP(wr); err != nil {
						goto failed
					}
				}
				g.Logger.Debugf("key: %s write client proto%v", ch.Key, p)
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			// server send
			if err = p.WriteTCP(wr); err != nil {
				goto failed
			}
			g.Logger.Debugf("tcp sent a message key:%s mid:%d proto(%v)", ch.Key, ch.Mid, p)
		}
		g.Logger.Debugf("key: %s tcp write start flush", ch.Key)
		// only hungry flush response
		if err = wr.Flush(); err != nil {
			g.Logger.Errorf("key: %s tcp write flush error(%v)", ch.Key, err)
			break
		}
		g.Logger.Debugf("key: %s tcp write end flush", ch.Key)
	}
failed:
	if err != nil {
		g.Logger.Errorf("key: %s dispatch tcp error(%v)", ch.Key, err)
	}
	conn.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (ch.Ready() == grpc.ProtoFinish)
	}
	g.Logger.Debugf("key: %s dispatch goroutine exit", ch.Key)
}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, p *grpc.Proto) (mid int64, key string, rid string, platform string, accepts []int32, err error) {
	for {
		if err = p.ReadTCP(rr); err != nil {
			return
		}
		if p.Op == grpc.OpAuth {
			break
		} else {
			g.Logger.Errorf("tcp request operation(%d) not auth", p.Op)
		}
	}
	if mid, key, rid, platform, accepts, err = s.Connect(p, ""); err != nil {
		g.Logger.Errorf("authTCP.Connect(key:%v).err(%v)", key, err)
		return
	}
	p.Op = grpc.OpAuthReply
	p.Body = nil
	if err = p.WriteTCP(wr); err != nil {
		g.Logger.Errorf("authTCP.WriteTCP(key:%v).err(%v)", key, err)
		return
	}
	err = wr.Flush()
	return
}
