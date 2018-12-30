package g

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/swanky2009/goim/job/g/conf"
)

const (
	ver   = "3.0.0"
	appid = "goim.job"
)

var (
	Conf *conf.Config

	Logger logger

	MetricsStat *Metrics
)

func init() {
	curPath := GetCurrentDir()

	SetPid(curPath)

	var err error

	if Conf, err = conf.LoadConf(curPath); err != nil {
		panic(err)
	}

	Logger = newLogger()

	Logger.Infoln(spew.Sdump(Conf))

	if err := InstanceDiscovery(); err != nil {
		Logger.Fatalf("init discovery:", err)
		os.Exit(1)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)

	rand.Seed(time.Now().UTC().UnixNano())

	Logger.Infof("goim-job [version: %s env: %+v] start", ver, Conf.Env)

	//stat
	MetricsStat = MetricsInstrumenting()
}

func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func SetPid(curPath string) int {
	runtimePath := curPath + "/runtime/"
	//log目录不存在，则创建
	_, err := os.Stat(runtimePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(runtimePath, os.ModePerm)
		}
	}
	filePath := curPath + "/runtime/pid"
	pid := os.Getpid()
	ioutil.WriteFile(
		filePath,
		[]byte(strconv.Itoa(pid)),
		os.ModePerm,
	)
	return pid
}
