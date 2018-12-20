package g

import (
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

var ServiceResolver naming.Resolver

func InstanceDiscovery() error {
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = Conf.Discovery.Addr
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return err
	}
	ServiceResolver = NewConsulResolver(consulClient, "goim-comet")

	return nil
}
