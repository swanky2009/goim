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

// Config .
type Config struct {
	ServiceName   string `yaml:"service_name"`
	RunMode       string // dev test prod
	MaxProc       int
	ServerTick    xtime.Duration
	OnlineTick    xtime.Duration
	LogPath       string
	Env           *Env
	Discovery     *DiscoveryConf
	RPCServer     *RPCServer
	HTTPServer    *HTTPServer
	Kafka         *Kafka
	Redis         *Redis
	Regions       map[string][]string
	Zipkin        *zipkinConf
	MetricsServer struct {
		Addr string
	} `yaml:"metrics_server"`
	//Backoff     *Backoff
}

// Env is env config.
type Env struct {
	Region string
	Zone   string
	Env    string
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

// Backoff .
// type Backoff struct {
// 	MaxDelay  int32
// 	BaseDelay int32
// 	Factor    float32
// 	Jitter    float32
// }

// Redis .
type Redis struct {
	Addrs        []string
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	PoolSize     int
	PoolTimeout  xtime.Duration
	IdleTimeout  xtime.Duration
	Expire       xtime.Duration
}

// Kafka .
type Kafka struct {
	Topic   string
	Brokers []string
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

// HTTPServer is http server config.
type HTTPServer struct {
	Network      string
	Addr         string
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
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
	if c.RPCServer != nil {
		c.RPCServer.fix()
	}
	if c.Redis != nil {
		c.Redis.fix()
	}
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

func (r *Redis) fix() {
	if r.DialTimeout <= 0 {
		r.DialTimeout = xtime.Duration(time.Millisecond * 100)
	}
	if r.ReadTimeout <= 0 {
		r.ReadTimeout = xtime.Duration(time.Second)
	}
	if r.WriteTimeout <= 0 {
		r.WriteTimeout = xtime.Duration(time.Second)
	}
	if r.PoolSize <= 0 {
		r.PoolSize = 10
	}
	if r.PoolTimeout <= 0 {
		r.PoolTimeout = xtime.Duration(time.Millisecond * 100)
	}
	if r.IdleTimeout <= 0 {
		r.IdleTimeout = xtime.Duration(time.Second * 30)
	}
}
