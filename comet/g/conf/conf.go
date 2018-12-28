package conf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	xtime "github.com/swanky2009/goim/pkg/time"
	"gopkg.in/yaml.v2"
)

const (
	RUN_MODE_LOCAL     = "local"
	RUN_MODE_CONTAINER = "container"
	RUN_MODE_K8S       = "k8s"
)

// Config is comet config.
type Config struct {
	ServiceName   string `yaml:"service_name"`
	RunMode       string // dev test prod
	MaxProc       int
	ServerTick    xtime.Duration
	OnlineTick    xtime.Duration
	LogPath       string
	Env           *Env
	Discovery     *DiscoveryConf
	TCP           *TCP       `yaml:"tcp_server"`
	WebSocket     *WebSocket `yaml:"websocket_server"`
	Timer         *Timer
	ProtoSection  *ProtoSection
	Bucket        *Bucket
	RPCClient     *RPCClient `yaml:"rpc_lient"`
	RPCServer     *RPCServer `yaml:"rpc_server"`
	Zipkin        *zipkinConf
	MetricsServer struct {
		Addr string
	} `yaml:"metrics_server"`
}

// Env is env config.
type Env struct {
	Region string
	Zone   string
	Host   string
	Weight int64
}

type DiscoveryConf struct {
	Addr string
}

type zipkinConf struct {
	Url         string
	ServiceName string `yaml:"service_name"`
	Reporter    struct {
		Timeout       int
		BatchSize     int `yaml:"batch_size"`
		BatchInterval int `yaml:"batch_interval"`
		MaxBacklog    int `yaml:"max_backlog"`
	}
}

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	Keepalive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// WebSocket is websocket config.
type WebSocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Timer is timer config.
type Timer struct {
	Timer     int
	TimerSize int
}

// ProtoSection is proto section.
type ProtoSection struct {
	HandshakeTimeout xtime.Duration
	WriteTimeout     xtime.Duration
	SvrProto         int
	CliProto         int
}

// Whitelist is white list config.
type Whitelist struct {
	Whitelist []int64
	WhiteLog  string
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
}

func LoadConf(curPath string) (*Config, error) {
	conf := new(Config)
	if err := godotenv.Load(curPath + "/.env"); err != nil {
		return nil, errors.New("Error loading .env file")
	}

	runMode := os.Getenv("RUN_MODE")
	fmt.Println("run mode:", runMode)

	var confFile string
	switch runMode {
	case RUN_MODE_LOCAL:
		confFile = curPath + "/conf/local.yaml"
	case RUN_MODE_CONTAINER:
		confFile = curPath + "/conf/container.yaml"
	case RUN_MODE_K8S:
		confFile = curPath + "/conf/k8s.yaml"
	default:
		return nil, errors.New("unsuppoer run mode! supports:[local,container,k8s]")
	}

	confb, _ := ioutil.ReadFile(confFile)
	if err := yaml.Unmarshal(confb, &conf); err != nil {
		return nil, errors.New("conf load failed: " + err.Error())
	}

	//LogPath Fix
	if conf.LogPath == "" {
		conf.LogPath = curPath + "/logs/"
	}

	conf.fix()

	return conf, nil
}

func (c *Config) fix() {
	if c.Env == nil {
		c.Env = new(Env)
	}
	c.Env.fix()

	if c.MaxProc <= 0 {
		c.MaxProc = 32
	}
	if c.ServerTick <= 0 {
		c.ServerTick = xtime.Duration(1 * time.Second)
	}
	if c.OnlineTick <= 0 {
		c.OnlineTick = xtime.Duration(10 * time.Second)
	}
	if c.RPCClient == nil {
		c.RPCClient = new(RPCClient)
	}
	c.RPCClient.fix()
	if c.RPCServer == nil {
		c.RPCServer = new(RPCServer)
	}
	c.RPCServer.fix()
	c.TCP.fix()
	c.ProtoSection.fix()
}

func (e *Env) fix() {
	if e.Region == "" {
		e.Region = os.Getenv("REGION")
	}
	if e.Zone == "" {
		e.Zone = os.Getenv("ZONE")
	}
	if e.Host == "" {
		e.Host, _ = os.Hostname()
	}
	if e.Weight <= 0 {
		e.Weight, _ = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
	}
}

func (r *RPCClient) fix() {
	if r.Dial <= 0 {
		r.Dial = xtime.Duration(time.Second)
	}
	if r.Timeout <= 0 {
		r.Timeout = xtime.Duration(time.Second)
	}
}

func (r *RPCServer) fix() {
	if r.Network == "" {
		r.Network = "tcp"
	}
	if r.Timeout <= 0 {
		r.Timeout = xtime.Duration(time.Second)
	}
	if r.IdleTimeout <= 0 {
		r.IdleTimeout = xtime.Duration(time.Second * 60)
	}
	if r.MaxLifeTime <= 0 {
		r.MaxLifeTime = xtime.Duration(time.Hour * 2)
	}
	if r.ForceCloseWait <= 0 {
		r.ForceCloseWait = xtime.Duration(time.Second * 20)
	}
	if r.KeepAliveInterval <= 0 {
		r.KeepAliveInterval = xtime.Duration(time.Second * 10)
	}
	if r.KeepAliveTimeout <= 0 {
		r.KeepAliveTimeout = xtime.Duration(time.Second * 10)
	}
}

func (r *TCP) fix() {
	if r.Sndbuf <= 0 {
		r.Sndbuf = 4096
	}
	if r.Rcvbuf <= 0 {
		r.Rcvbuf = 4096
	}
	if r.Reader <= 0 {
		r.Reader = 32
	}
	if r.ReadBuf <= 0 {
		r.ReadBuf = 1024
	}
	if r.ReadBufSize <= 0 {
		r.ReadBufSize = 2048
	}
	if r.Writer <= 0 {
		r.Writer = 32
	}
	if r.WriteBuf <= 0 {
		r.WriteBuf = 1024
	}
	if r.WriteBufSize <= 0 {
		r.WriteBufSize = 2048
	}
}

func (r *ProtoSection) fix() {
	if r.SvrProto <= 0 {
		r.SvrProto = 10
	}
	if r.CliProto <= 0 {
		r.CliProto = 5
	}
}
