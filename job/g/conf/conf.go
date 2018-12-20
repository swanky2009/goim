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

// Config is job config.
type Config struct {
	ServiceName   string `yaml:"service_name"`
	RunMode       string // dev test prod
	MaxProc       int
	ServerTick    xtime.Duration
	OnlineTick    xtime.Duration
	LogPath       string
	Env           *Env
	Kafka         *Kafka
	Discovery     *DiscoveryConf
	Comet         *Comet
	Zipkin        *zipkinConf
	MetricsServer struct {
		Addr string
	} `yaml:"metrics_server"`
}

// Comet is comet config.
type Comet struct {
	Dial        xtime.Duration
	Timeout     xtime.Duration
	RoutineChan int
	RoutineSize int
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
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
	if c.Comet != nil {
		c.Comet.fix()
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

func (c *Comet) fix() {
	if c.Dial <= 0 {
		c.Dial = xtime.Duration(time.Second)
	}
	if c.Timeout <= 0 {
		c.Timeout = xtime.Duration(time.Second)
	}
	if c.RoutineChan <= 0 {
		c.RoutineChan = 10
	}
	if c.RoutineSize <= 0 {
		c.RoutineSize = 10
	}
}
