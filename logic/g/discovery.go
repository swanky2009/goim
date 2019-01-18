package g

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	consulsd "github.com/go-kit/kit/sd/consul"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/swanky2009/goim/pkg/hash"
	"github.com/swanky2009/goim/pkg/ip"
)

var (
	ServiceRegistrar *consulsd.Registrar
	ServiceInstancer *consulsd.Instancer
)

func InstanceDiscovery() error {
	//创建一个新服务
	host, port, err := net.SplitHostPort(Conf.RPCServer.Addr)
	if err != nil {
		return errors.New("rpc server addr error")
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("rpc server addr error")
	}
	if host == "" {
		host = ip.InternalIP()
	}
	registration := &consulapi.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%s-%s", Conf.ServiceName, Conf.Env.Region, Conf.Env.Zone, Conf.Env.Host),
		Name:    Conf.ServiceName,
		Port:    portInt,
		Tags:    []string{"v1"},
		Address: host,
	}

	//增加check  consul 0.7以上版本才支持 grpc health check
	// check := &consulapi.AgentServiceCheck{
	// 	Interval:                       time.Duration(Conf.RPCServer.KeepAliveInterval).String(),
	// 	Timeout:                        time.Duration(Conf.RPCServer.Timeout).String(),
	// 	DeregisterCriticalServiceAfter: time.Duration(Conf.RPCServer.IdleTimeout).String(),
	// 	GRPC:                           fmt.Sprintf("%v:%v/%v", host, portInt, Conf.ServiceName),
	// }
	//增加check consul 0.7以下版本用 http health check
	mhost, mport, err := net.SplitHostPort(Conf.MetricsServer.Addr)
	if err != nil {
		return errors.New("metrics server addr error")
	}
	if mhost == "" {
		mhost = host
	}
	check := &consulapi.AgentServiceCheck{
		Interval:                       time.Duration(Conf.RPCServer.KeepAliveInterval).String(),
		Timeout:                        time.Duration(Conf.RPCServer.Timeout).String(),
		DeregisterCriticalServiceAfter: time.Duration(Conf.RPCServer.IdleTimeout).String(),
		HTTP:                           fmt.Sprintf("http://%s:%s/check", mhost, mport),
	}
	registration.Check = check

	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = Conf.Discovery.Addr
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return err
	}

	client := consulsd.NewClient(consulClient)

	ServiceRegistrar = consulsd.NewRegistrar(client, registration, Logger)

	ServiceInstancer = consulsd.NewInstancer(client, Logger, "goim-comet", []string{""}, true)

	return nil
}

func GetCometService() (addrs map[string]string, err error) {
	var (
		addr string
	)
	state := ServiceInstancer.GetServiceState()

	if state.Err != nil {
		err = state.Err
		return
	}
	addrs = make(map[string]string, len(state.Instances))

	for _, addr = range state.Instances {
		addrs[hash.Sha1s(addr)] = addr
	}
	return
}

func GetCometServiceMetas() (metas map[string]map[string]string, err error) {
	var (
		entry   *consulapi.ServiceEntry
		entries []*consulapi.ServiceEntry
		addr    string
	)
	entries, err = ServiceInstancer.GetServiceEntrys()

	if err != nil {
		return
	}
	metas = make(map[string]map[string]string, len(entries))

	for _, entry = range entries {
		addr = fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port)
		metas[hash.Sha1s(addr)] = entry.Service.Meta
	}
	return
}
