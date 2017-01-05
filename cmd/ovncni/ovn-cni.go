// Copyright 2014 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/containernetworking/cni/pkg/ip"
	"github.com/containernetworking/cni/pkg/ipam"
	"github.com/containernetworking/cni/pkg/ns"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/heartlock/ovn-kubernetes/pkg/exec"
	"github.com/vishvananda/netlink"
)

const defaultBrName = "cni0"

type NetConf struct {
	types.NetConf
	BrName       string `json:"bridge"`
	IsGW         bool   `json:"isGateway"`
	IsDefaultGW  bool   `json:"isDefaultGateway"`
	ForceAddress bool   `json:"forceAddress"`
	IPMasq       bool   `json:"ipMasq"`
	MTU          int    `json:"mtu"`
	HairpinMode  bool   `json:"hairpinMode"`
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func loadNetConf(bytes []byte) (*NetConf, error) {
	n := &NetConf{
		BrName: defaultBrName,
	}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}
	return n, nil
}

func setupInterface(containerId string, netns ns.NetNS, ifName string, macAddress string) (string, error) {
	var hostVethName string

	err := netns.Do(func(hostNS ns.NetNS) error {
		// create the veth pair in the container and move host end into host netns
		hostVeth, contVeth, err := ip.SetupVeth(ifName, mtu, hostNS)
		if err != nil {
			return nil, err
		}
		hw, err := net.ParseMAC(macAddress)
		if err != nil {
			return nil, err
		}
		err = netlink.LinkSetHardwareAddr(contVeth, hw)
		if err != nil {
			return nil, err
		}

		hostVethName = hostVeth.Attrs().Name

		return nil
	})
	if err != nil {
		return nil, err
	}

	// need to lookup hostVeth again as its index has changed during ns move
	hostVeth, err := netlink.LinkByName(hostVethName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup %q: %v", hostVethName, err)
	}

	// set hairpin mode
	if err = netlink.LinkSetHairpin(hostVeth, hairpinMode); err != nil {
		return fmt.Errorf("failed to setup hairpin mode for %v: %v", hostVethName, err)
	}
	if err = netlink.LinkSetName(hostVeth, containerId[:15]); err != nil {
		return nil, err
	}

	return containerId[:15], nil

}

func cmdAdd(args *skel.CmdArgs) error {
	n, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	if n.IsDefaultGW {
		n.IsGW = true
	}

	k8sApiServer, err := exec.RunCommand("--if-exists", "get", "Open_vSwitch", ".", "external_ids:k8s-api-server")
	if err != nil {
		return fmt.Errorf("get k8s apiserver fail: %v", err)
	}
	k8sApiServer = "http://" + k8sApiServer

	ArgsMap := make(map[string]string)
	for cniArg := range args.Args.Split(";") {
		k, v := cniArg.Split("=")
		cniArgsMap[k] = v
	}
	namespace, ok := cniArgsMap["K8S_POD_NAMESPACE"]
	if !ok {
		return fmt.Errorf("there is no key K8S_POD_NAMESPACE")
	}
	podName, ok := cniArgsMap["K8S_POD_NAME"]
	if !ok {
		return fmt.Errorf("there is no key K8S_POD_NAME")
	}
	containerId, ok := cniArgsMap["K8S_POD_INFRA_CONTAINER_ID"]
	if !ok {
		return fmt.Errorf("there is no key K8S_POD_INFRA_CONTAINER_ID")
	}

	counter := 30

	for ; counter > 0; counter-- {
		annotations, err := kubernetes.GetPodAnnotations(k8sApiServer, namespace, podName)
		if err != nil {
			return fmt.Errorf("failed to get pod annotation: %v", err)
		}
		if annotation {
			if ovn, ok := annotation["ovn"]; ok {
				break
			}
		}

		time.Sleep(0.1)
	}
	ovn := annotations["ovn"].(map[string]string)
	macAddress, ok := ovn["mac_address"]
	if !ok {
		return fmt.Errorf("missing key macAddress")
	}
	ipAddress, ok := ovn["ip_address"]
	if !ok {
		return fmt.Errorf("missing key ipAddress")
	}

	gatewayIp, ok := ovn["gateway_ip"]
	if !ok {
		return fmt.Errorf("missing key gatewayIp")
	}
	var result *types.Result
	ipc, ipnet, err := net.ParseCIDR(ipAddress)
	ipg, _, err := net.ParseCIDR(gatewayIp)
	ipnet.IP = ipc
	result.IP4.IP = ipnet
	result.IP4.Gateway = ipg

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	vethOutside, err := setupInterface(containerId, netns, args.IfName, macAddress)
	if err != nil {
		return err
	}

	if err := netns.Do(func(_ ns.NetNS) error {
		// set the default gateway if requested
		if n.IsDefaultGW {
			_, defaultNet, err := net.ParseCIDR("0.0.0.0/0")
			if err != nil {
				return err
			}

			for _, route := range result.IP4.Routes {
				if defaultNet.String() == route.Dst.String() {
					if route.GW != nil && !route.GW.Equal(result.IP4.Gateway) {
						return fmt.Errorf(
							"isDefaultGateway ineffective because IPAM sets default route via %q",
							route.GW,
						)
					}
				}
			}

			result.IP4.Routes = append(
				result.IP4.Routes,
				types.Route{Dst: *defaultNet, GW: result.IP4.Gateway},
			)

			// TODO: IPV6
		}

		return ipam.ConfigureIface(args.IfName, result)
	}); err != nil {
		return err
	}

	ifaceId := namespace + "_" + podName

	_, err = exec.RunCommand("ovs-vsctl", "add-port", "br-int", vethOutside, "--", "set", "interface", vethOutside, "external_ids:attached_mac="+macAddress, "external_ids:iface-id="+ifaceId, "external_ids:ip_address="+ipAddress)
	if err != nil {
		return fmt.Errorf("Unable to plug interface into OVN bridge: %v", err)
	}
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	n, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	if args.Netns == "" {
		return nil
	}

	_, err = exec.RunCommand("ovs-vsctl", "del-port", args.ContainerID[:15])
	if err != nil {
		return err
	}

	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.Legacy)
}
