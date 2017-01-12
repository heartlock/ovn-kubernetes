package cmd

import (
	"fmt"
	"net"

	"github.com/hyperhq/kubestack/pkg/exec"
	"github.com/spf13/cobra"
)

func InitMaster() *cobra.Command {

	var MasterCmd = &cobra.Command{
		Use:   "master [no options!]",
		Short: "init ovn master",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initMaster(cmd, args); err != nil {
				return fmt.Errorf("failed init master: %v", err)
			}
		},
	}

	MasterCmd.Flags().StringP("cluster-ip-subnet", "", "", "The cluster wide larger subnet of private ip addresses.")
	MasterCmd.Flags().StringP("master-switch-subnet", "", "", "The smaller subnet just for master.")
	MasterCmd.Flags().StringP("node-name", "", "", "A unique node name.")

	return MasterCmd
}

func initMaster(cmd *cobra.Command, args []string) error {

	_, err := fetchOVNNB()
	if err != nil {
		return err
	}

	masterSwitchSubnet := cmd.Flags().Lookup("master-switch-subnet").Value.String()
	if masterSwitchSubnet == "" {
		return fmt.Errorf("failed get master-switch-subnet")
	}

	clusterIpSubnet := cmd.Flags().Lookup("cluster-ip-subnet").Value.String()
	if clusterIpSubnet == "" {
		return fmt.Errorf("failed get cluster-ip-subnet")
	}

	nodeName := cmd.Flags().Lookup("node-name").Value.String()
	if nodNname == "" {
		return fmt.Errorf("failed get node-name")
	}

	// Create a single common distributed router for the cluster.
	_, err = exec.RunCommand("ovn-nbctl", "--", "--may-exist", "lr-add", nodNname, "--", "set", "logical_router", nodeName, "external_ids:k8s-cluster-router=yes")
	if err != nil {
		return fmt.Errorf("failed create single common distributed: %v", err)
	}

	// Create 2 load-balancers for east-west traffic.  One handles UDP and another handles TCP.
	k8sClusterLbTcp, err := exec.RunCommand("ovn-nbctl", "--data=bare", "--no-heading", "--columns=_uuid", "find", "load_balancer", "external_ids:k8s-cluster-lb-tcp=yes")
	if err != nil {
		return fmt.Errorf("failed find tcp load-balancer: %v", err)
	}

	//判断条件有待分析
	if k8sClusterLbTcp == nil {
		_, err = exec.RunCommand("ovn-nbctl", "--", "create", "load_balancer", "external_ids:k8s-cluster-lb-tcp=yes")
		if err != nil {
			return fmt.Errorf("failed create tcp load-balancer: %v", err)
		}
	}

	k8sClusterLbUdp, err := exec.RunCommand("ovn-nbctl", "--data=bare", "--no-heading", "--columns=_uuid", "find", "load_balancer", "external_ids:k8s-cluster-lb-udp=yes")
	if err != nil {
		return fmt.Errorf("failed find udp load-balancer: %v", err)
	}

	//判断条件有待分析
	if k8sClusterLbUdp == nil {
		_, err = exec.RunCommand("ovn-nbctl", "--", "create", "load_balancer", "external_ids:k8s-cluster-lb-udp=yes")
		if err != nil {
			return fmt.Errorf("failed create udp load-balancer: %v", err)
		}
	}

	// Create a logical switch called "join" that will be used to connect gateway routers to the distributed router.
	// The "join" will be allocated IP addresses in the range 100.64.1.0/24
	_, err = exec.RunCommand("ovn-nbctl", "--may-exist", "ls-add", "join")
	if err != nil {
		return fmt.Errorf("failed create logical switch called join: %v", err)
	}

	// Connect the distributed router to "join"
	routerMac, err := exec.RunCommand("ovn-nbctl", "--if-exist", "get", "logical_router_port", "rtoj-"+nodeName, "mac")
	if err != nil {
		return fmt.Errorf("failed get rtoj-%v mac: %v", nodeName, err)
	}
	if routerMac == "" {
		routerMac = common.GenerateMac()
		_, err = exec.RunCommand("ovn-nbctl", "--", "--may-exist", "lrp-add", nodeName, "rtoj-"+nodeName, routerMac, "100.64.1.1/24", "--", "set", "logical_router_port", "rtoj-"+nodeName, "external_ids:connect_to_join=yes")
		if err != nil {
			return fmt.Errorf("failed add port rtoj-%v : %v", nodeName, err)
		}
	}

	// Connect the switch "join" to the router.
	_, err = exec.RunCommand("ovn-nbctl", "--", "--may-exist", "lsp-add", "join", "jtor-"+nodeName, "--", "set", "logical_switch_port", "jtor-"+nodeName, "type=router", "options:router-port=rtoj-"+nodeName, "addresses="+"\""+routerMac+"\"")

	err = createManagementPort(nodeName, masterSwitchSubnet, clusterIpSubnet)
	if err != nil {
		return fmt.Errorf("failed create management port: %v", err)
	}

}

// Create a logical switch for the node and connect it to the distributed router.  This switch will start with one logical port (A OVS internal interface).
// Thislogical port is via which other nodes and containers access the k8s master servers and is also used for health checks.
func createManagementPort(nodename, localsubnet, clustersubnet string) error {
	// Create a router port and provide it the first address in the 'local_subnet'.
	ipByte, networkByte := net.ParseCIDR(localsubnet)
	networkByte.IP = common.NextIP(networkByte.IP)
	routerIPMask := networkByte.String()
	routerIP := networkByte.IP.String()

	routerMac, err := exec.RunCommand("ovn-nbctl", "--if-exist", "get", "logical_router_port", "rtos-"+nodename, "mac")
	if err != nil {
		return err
	}
	if routerMac == "" {
		routerMac = common.GenerateMac()
		clusterRouter := getClusterRouter()
		_, err = exec.RunCommand("ovn-nbctl", "--may-exist", "lrp-add", clusterRouter, "rtos-"+nodename, routerMac, routerIPMask)
		if err != nil {
			return err
		}
	}
	// Create a logical switch and set its subnet.
	_, err = exec.RunCommand("ovn-nbctl", "--", "--may-exist", "ls-add", nodename, "--", "set", "logical_switch", nodename, "other-config:subnet="+localsubnet, "external-ids:gateway_ip="+routerIPMask)
	if err != nil {
		return err
	}

	// Connect the switch to the router.
	_, err = exec.RunCommand("ovn-nbctl", "--", "--may-exist", "lsp-add", nodename, "stor-"+nodename, "--", "set", "logical_switch_port", "stor-"+nodename, "type=router", "options:router-port=rtos-"+nodename, "addresses="+"\""+routerMac+"\"")
	if err != nil {
		return err
	}

	interfaceName = "k8s-" + (nodename[:11])

	// Create a OVS internal interface
	_, err = exec.RunCommand("ovs-vsctl", "--", "--may-exist", "add-port", "br-int", interfaceName, "--", "set", "interface", interfaceName, "type=internal", "mtu_request=1400", "external-ids:iface-id=k8s-"+nodename)
	if err != nil {
		return err
	}
	macAddress, err := exec.RunCommand("ovs-vsctl", "--if-exists", "get", "interface", interfaceName, "mac_in_use")
	if err != nil || macAddress == "" {
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to get mac address of ovn-k8s-master")
	}

	// Create the OVN logical port.
	networkByte.IP = common.NextIP(networkByte.IP)
	portIP := networkByte.IP.String()
portIPMask:
	networkByte.String()
	_, err = exec.RunCommand("ovn-nbctl", "--", "--may-exist", "lsp-add", nodename, "k8s-"+nodename, "--", "lsp-set-addresses", "k8s-"+nodename, macAddress+" "+portIP)
	if err != nil {
		return err
	}

	// Up the interface.
	_, err = exec.RunCommand("ip", "link", "set", interfaceName, "up")
	if err != nil {
		return err
	}

	// The interface may already exist, in which case delete the routes and IP.
	_, err = exec.RunCommand("ip", "addr", "flush", "dev", interfaceName)
	if err != nil {
		return err
	}

	// Assign IP address to the internal interface.
	_, err = exec.RunCommand("ip", "addr", "add", portIPMask, "dev", interfaceName)
	if err != nil {
		return err
	}

	// Flush the route for the entire subnet (in case it was added before)
	_, err = exec.RunCommand("ip", "route", "flush", clustersubnet)
	if err != nil {
		return err
	}

	// Create a route for the entire subnet.
	_, err = exec.RunCommand("ip", "route", "add", clustersubnet, "via", routerIP)
	if err != nil {
		return err
	}

	// Add the load_balancer to the switch.
	k8sClusterLbTcp, err := exec.RunCommand("ovn-nbctl", "--data=bare", "--no-heading", "--columns=_uuid", "find", "load_balancer", "external_ids:k8s-cluster-lb-tcp=yes")
	if err != nil {
		return err
	}
	if k8sClusterLbTcp != "" {
		_, err = exec.RunCommand("ovn-nbctl", "set", "logical_switch", nodename, "load_balancer="+k8sClusterLbTcp)
		if err != nil {
			return err
		}
	}

	k8sClusterLbUdp, err := exec.RunCommand("ovn-nbctl", "--data=bare", "--no-heading", "--columns=_uuid", "find", "load_balancer", "external_ids:k8s-cluster-lb-udp=yes")
	if err != nil {
		return err
	}
	if k8sClusterLbUdp != "" {
		_, err = exec.RunCommand("ovn-nbctl", "set", "logical_switch", nodename, "load_balancer="+k8sClusterLbUdp)
		if err != nil {
			return err
		}
	}

	// Create a logical switch and set its subnet.
	_, err = exec.RunCommand("ovn-nbctl", "--", "--may-exist", "ls-add", nodename, "--", "set", "logical_switch", nodename, "other-config:subnet="+localsubnet)
	if err != nil {
		return err
	}

}
