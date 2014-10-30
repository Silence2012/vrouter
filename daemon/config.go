package daemon

import (
	"github.com/zhgwenming/vrouter/Godeps/_workspace/src/github.com/coreos/go-etcd/etcd"
	"github.com/zhgwenming/vrouter/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/zhgwenming/vrouter/netinfo"
	//"github.com/zhgwenming/vrouter/registry"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

type Command struct {
	etcdServers *string
	hostip      string

	// command switches
	daemonMode  bool
	gatewayMode bool

	// tls authentication related
	CaFile   string
	CertFile string
	KeyFile  string

	// vrouter daemon
	Daemon *Daemon
}

func NewCommand() *Command {
	return &Command{}
}

func (cmd *Command) InitCmd(servers *string) *cobra.Command {

	vrouter := NewDaemon()
	cmd.Daemon = vrouter

	cmd.etcdServers = servers

	routerCmd := &cobra.Command{
		Use:  "vrouter",
		Long: "vrouter is a tool for routing distributed Docker containers.\n\n",
		Run:  cmd.Run,
	}

	var ipnet *net.IPNet
	ipnetlist := netinfo.ListIPNet(true)
	if len(ipnetlist) > 0 {
		ipnet = ipnetlist[0]
	}

	// vrouter flags
	cmdflags := routerCmd.Flags()

	vrouter.Hostname, _ = os.Hostname()

	cmdflags.BoolVarP(&cmd.daemonMode, "daemon", "d", false, "whether to run as daemon mode")
	cmdflags.BoolVarP(&cmd.gatewayMode, "gateway", "g", false, "to run as dedicated gateway, will not allocate subnet on this machine")

	// need to convert to IPNet form
	cmdflags.StringVarP(&cmd.hostip, "hostip", "i", ipnet.String(), "use specified ip/mask instead auto detected ip address")

	// vrouter information
	cmdflags.StringVarP(&vrouter.Hostname, "hostname", "n", vrouter.Hostname, "hostname to use in daemon mode")
	cmdflags.StringVarP(&vrouter.bridgeName, "bridge", "b", "docker0", "bridge name to setup")

	return routerCmd
}

func (cmd *Command) Run(c *cobra.Command, args []string) {
	if cmd.daemonMode {
		//daemon := cmd.Daemon
		// -peer-addr 127.0.0.1:7001 -addr 127.0.0.1:4001 -data-dir machines/machine1 -name machine1
		//go registry.StartEtcd("-peer-addr", "127.0.0.1:7001", "-addr", "127.0.0.1:4001", "-data-dir", "machines/"+daemon.Hostname, "-name", daemon.Hostname)

		servers := strings.Split(*cmd.etcdServers, ",")
		vrouter := cmd.Daemon

		// start keepalive first
		//if cmd.CaFile != "" && cmd.CertFile != "" && cmd.KeyFile != "" {
		log.Printf("%v", servers)
		if strings.HasPrefix(servers[0], "https://") {
			eclient, err := etcd.NewTLSClient(servers, cmd.CertFile, cmd.KeyFile, cmd.CaFile)
			if err != nil {
				log.Fatalf("error to create tls client: %s", err)
			}

			log.Printf("established tls connection.")
			vrouter.etcdClient = eclient
		} else {
			vrouter.etcdClient = etcd.NewClient(servers)
			log.Printf("established plain text connection.")
		}
		err := vrouter.KeepAlive()
		if err != nil {
			log.Fatalf("error to keepalive: %s, other instance running?", err)
		}

		// bind and get a bridge IPNet with our iface ip
		// create the routing table entry in registry
		bridgeIPNet, err := vrouter.BindBridgeIPNet(cmd.hostip)
		if err != nil {
			log.Fatal("Failed to bind router interface: ", err)
		} else {
			log.Printf("Requested bridge ip - %v\n", bridgeIPNet)
		}

		// create bridge if we're running under linux
		// to debug on Mac OS X
		if runtime.GOOS == "linux" {
			err = vrouter.CreateBridge(bridgeIPNet.String())
			if err != nil {
				log.Fatal(err)
			}
		}

		// monitor the routing table change
		err = vrouter.ManageRoute()
		if err != nil {
			log.Fatal(err)
		}

	} else {
		c.Help()
	}
}
