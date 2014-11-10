package controller

import (
	"fmt"
	"github.com/zhgwenming/vrouter/Godeps/_workspace/src/github.com/coreos/go-etcd/etcd"
	"github.com/zhgwenming/vrouter/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/zhgwenming/vrouter/registry"
	"github.com/zhgwenming/vrouter/service"
	"log"
)

type ServiceManager struct {
	service.Service
	config     *Config
	etcdClient *etcd.Client
}

func (srv *ServiceManager) Run(cmd *cobra.Command, args []string) {
	var action string

	srv.etcdClient = registry.NewClient(srv.config.etcdConfig)

	if len(args) > 0 {
		var err error
		action = args[0]
		// all the actions
		switch action {
		case "add":
			err = srv.Add()
		case "delete":
			err = srv.Delete()
		case "list":
			err = srv.List()
		default:
			cmd.Usage()
		}

		if err != nil {
			log.Fatalf("Error to add service: %s\n", err)
		}
		fmt.Printf("%#v\n", action)
	} else {
		// list all exist service?
	}

}

func (srv *ServiceManager) Add() error {
	if srv.Name == "" {
		return fmt.Errorf("No service name specified")
	}

	key := registry.RouterServicesPrefix() + "/" + srv.Name
	value := string(srv.Marshal())
	if _, err := srv.etcdClient.Create(key, value, uint64(0)); err != nil {
		return err
	}
	fmt.Printf("service %s added\n", srv.Name)

	return nil
}

func (srv *ServiceManager) Delete() error {
	if srv.Name == "" {
		return fmt.Errorf("No service name specified")
	}

	key := registry.RouterServicesPrefix() + "/" + srv.Name
	if _, err := srv.etcdClient.Delete(key, true); err != nil {
		return err
	}

	fmt.Printf("service %s deleted\n", srv.Name)
	return nil
}

func (srv *ServiceManager) List() error {
	fmt.Printf("List services\n")

	return nil
}
