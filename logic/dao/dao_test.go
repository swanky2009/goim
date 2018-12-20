package dao

import (
	"flag"
	"testing"

	"github.com/swanky2009/goim/logic/g"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../../cmd/logic/logic-example.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
}
