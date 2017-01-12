package cmd

import (
	"fmt"

	"github.com/heartlock/ovn-kubernetes/pkg/exec"
)

var CNI_CONF_PATH = "/etc/cni/net.d"
var CNI_LINK_PATH = "/opt/cni/bin"
var CNI_PLUGIN = "ovn-k8s-cni-overlay"
var OVN_NB string
var K8S_API_SERVER string
var K8S_CLUSTER_ROUTER string
var K8S_CLUSTER_LB_TCP string
var K8S_CLUSTER_LB_UDP string
var K8S_NS_LB_TCP string
var K8S_NS_LB_UDP string
var OVN_MODE string

func fetchOVNNB() (string, error) {
	OVN_NB, err := exec.RunCommand("ovs-vsctl", "--if-exists", "get", "Open_vSwitch", ".", "external_ids:ovn-nb")
	if err != nil || OVN_NB == "" {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("OVN central database's ip address not set: %v", err)
	}
	return OVN_NB, nil

}
