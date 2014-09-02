package main

import (
	//"fmt"
	"github.com/zhgwenming/vrouter/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/zhgwenming/vrouter/pkg/registry"
	"log"
)

var (
	daemon     bool
	gateway    bool
	hostname   string
	etcdServer string
)

func virtRouter(c *cobra.Command, args []string) {
	if daemon {
		registry.KeepAlive(hostname)
	} else {
		c.Help()
	}
}

func main() {
	routerCmd := &cobra.Command{
		Use:  "vrouter",
		Long: "vrouter is a tool for routing distributed Docker containers.\n\n",
		Run:  virtRouter,
	}

	routerCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "whether to run as daemon mode")
	routerCmd.Flags().BoolVarP(&gateway, "gateway", "g", false, "to run as dedicated gateway, will not allocate subnet on this machine")
	routerCmd.Flags().StringVarP(&hostname, "hostname", "H", "", "hostname to use in daemon mode")
	routerCmd.PersistentFlags().StringVarP(&etcdServer, "etcd_server", "e", "http://127.0.0.1:4001", "etcd registry addr")

	registry.Init(routerCmd, etcdServer)

	if err := routerCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
