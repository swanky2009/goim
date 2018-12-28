package g

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	consulsd "github.com/go-kit/kit/sd/consul"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/swanky2009/goim/pkg/ip"
	"google.golang.org/grpc/naming"
)

var ServiceRegistrar *consulsd.Registrar
var ServiceResolver naming.Resolver

//var peo People = &Stduent{}

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
		Kind:    consulapi.ServiceKindTypical,
		ID:      fmt.Sprintf("%s-%s-%s-%s", Conf.ServiceName, Conf.Env.Region, Conf.Env.Zone, Conf.Env.Host),
		Name:    Conf.ServiceName,
		Port:    portInt,
		Tags:    []string{"v1"},
		Address: host,
	}

	//增加check consul 0.7以上版本才支持 grpc health check
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

	ServiceResolver = NewConsulResolver(consulClient, "goim-logic")

	return nil
}
