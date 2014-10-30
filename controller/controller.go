package controller

import (
	"fmt"
	"github.com/zhgwenming/vrouter/Godeps/_workspace/src/github.com/coreos/go-etcd/etcd"
	"github.com/zhgwenming/vrouter/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/zhgwenming/vrouter/netinfo"
	"github.com/zhgwenming/vrouter/registry"
	"log"
	"net"
	"strings"
)

var (
	machines      string
	hostNames     []string
	cellSubnet    string
	overlaySubnet string
	etcdConfig    *registry.ClientConfig
	etcdClient    *etcd.Client
)

// create etcd client
// register cobra subcommand
func InitCmd(parent *cobra.Command, client *registry.ClientConfig) {

	etcdConfig = client

	// register new subcommand
	initCmd := &cobra.Command{
		Use:   "init <machine1,machine2,..>",
		Short: "init the machine registry",
		Long:  "init the machine registry with specific ip network information",
		Run:   registryInit,
	}

	initCmd.Flags().StringVarP(&cellSubnet, "cellnet", "c", registry.DEFAULT_SUBNET, "cell cidr subnet ip address")
	initCmd.Flags().StringVarP(&overlaySubnet, "overlay", "o", registry.DEFAULT_SUBNET, "the whole overlay subnet ip address")

	parent.AddCommand(initCmd)
}

func registryInit(cmd *cobra.Command, args []string) {

	etcdClient = registry.NewClient(etcdConfig)

	if len(args) > 0 {
		machines = args[0]
		// just make slice if machines is not empty
		if len(machines) > 0 {
			hostNames = strings.Split(machines, ",")
		} else {
			cmd.Help()
			log.Fatal("Empty machine list specified")
		}
	} else {
		cmd.Help()
		log.Fatal("No machine list specified")
	}

	_, ipnet, err := net.ParseCIDR(cellSubnet)

	if err != nil {
		log.Fatal(err)
	}

	nets := netinfo.GetAllSubnet(ipnet, 8)

	fmt.Printf("vrouter init %s, %v\n", cellSubnet, ipnet)
	//fmt.Printf("%v\n", nets)

	//fmt.Printf("hostnames %d, %v\n", len(hostNames), hostNames)
	for i, node := range hostNames {
		key := registry.BridgeInfoPath(node)
		log.Printf("initialize config for host %s\n", node)
		if _, err := etcdClient.Create(key, nets[i].String(), 0); err != nil {
			log.Printf("Error to create node: %s", err)
		}

	}

	// create the overlay network ip info
	key := registry.RouterOverlayPath()
	log.Printf("initialize overlay network ip information with %s\n", overlaySubnet)
	if _, err := etcdClient.Create(key, overlaySubnet, 0); err != nil {
		log.Printf("Error to create node: %s", err)
	}
}
