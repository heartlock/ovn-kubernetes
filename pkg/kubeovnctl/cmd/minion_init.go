package cmd

import (
	"fmt"
	"os"

	"github.com/heartlock/ovn-kubernetes/pkg/exec"
	"github.com/spf13/cobra"
)

func InitMinion() *cobra.Command {

	var MinionCmd = &cobra.Command{
		Use:   "minion [no options!]",
		Short: "init ovn minion",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initMinion(); err != nil {
				return fmt.Errorf("failed init minion: %v", err)
			}
		},
	}

	MinionCmd.Flags().StringP("cluster-ip-subnet", "", "", "The cluster wide larger subnet of private ip addresses.")
	MinionCmd.Flags().StringP("--minion-switch-subnet", "", "", "The smaller subnet just for this master.")
	MasterCmd.Flags().StringP("node-name", "", "", "A unique node name.")

	return MinionCmd
}

func initMinion(cmd *cobra.Command, args []string) error {

	_, err := fetchOVNNB()
	if err != nil {
		return err
	}

	minionSwitchSubnet := cmd.Flags().Lookup("minion-switch-subnet").Value.String()
	if masterSwitchSubnet == "" {
		return fmt.Errorf("failed get minion-switch-subnet")
	}

	clusterIpSubnet := cmd.Flags().Lookup("cluster-ip-subnet").Value.String()
	if clusterIpSubnet == "" {
		return fmt.Errorf("failed get cluster-ip-subnet")
	}

	nodeName := cmd.Flags().Lookup("node-name").Value.String()
	if nodNname == "" {
		return fmt.Errorf("failed get node-name")
	}

	cniPluginPath, err := exec.LookPath(CNI_PLUGIN)
	if err != nil {
		return err
	}

	_, err = os.Stat(CNI_LINK_PATH)
	if err != nil && !os.IsExist(err) {
		err = os.MkdirAll(CNI_LINK_PATH, os.ModeDir)
		if err != nil {
			return err
		}
	}
	cniFile := CNI_LINK_PATH + "/ovn_cni"
	_, err = os.Stat(cniFile)
	if err != nil && !os.IsExist(err) {
		_, err = exec.RunCommand("ln", "-s", cni_plugin_path, cni_file)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(CNI_CONF_PATH)
	if err != nil && !os.IsExist(err) {
		err = os.MkdirAll(CNI_CONF_PATH, os.ModeDir)
		if err != nil {
			return err
		}
	}

	// Create the CNI config
	cniConf := CNI_CONF_PATH + "/10-net.conf"
	_, err = os.Stat(cniConf)
	if err != nil && !os.IsExist(err) {
		// TODO:verify if it is needed to set config file in 10-net.conf
		/*data := &main.NetConf{
					"name": "net",
		            "type": "ovn_cni",
		            "bridge": "br-int",
		            "isGateway": "true",
		            "ipMasq": "false",
		            "ipam": types.NetConf.IPAM{
		                         "type": "host-local",
		                         "subnet": minionSwitchSubnet,
							},

					}
		*/
	}

	err = createManagementPort(nodeName, minionSwitchSubnet, clusterIpSubnet)
	if err != nil {
		return err
	}

}
