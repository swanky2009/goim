package g

import (
	consulsd "github.com/go-kit/kit/sd/consul"
	consulapi "github.com/hashicorp/consul/api"
)

var ServiceInstancer *consulsd.Instancer

func InstanceDiscovery() error {
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = Conf.Discovery.Addr
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return err
	}

	client := consulsd.NewClient(consulClient)

	ServiceInstancer = consulsd.NewInstancer(client, Logger, "goim-comet", []string{""}, true)

	return nil
}
