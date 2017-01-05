package cmd

import {

 	"os"

	"fmt"
	"github.com/spf13/cobra"
}

func InitGateway() &cobra.Command {

	var GatewayCmd = &cobra.Command{
        Use:   "gateway [no options!]",
        Short: "init ovn gateway",
        PreRun: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Inside subCmd PreRun with args: %v\n", args)
        },
        Run: func(cmd *cobra.Command, args []string) {
        	if err := initGateway(); err != nil {
        		fmt.Fprint(os.Stderr, err)
        		os.Exit(1)
        	}
            fmt.Printf("Inside subCmd Run with args: %v\n", args)
        },
        PostRun: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Inside subCmd PostRun with args: %v\n", args)
        },
        PersistentPostRun: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Inside subCmd PersistentPostRun with args: %v\n", args)
        },
    }

    GatewayCmd.Flags().StringP("cluster-ip-subnet", "", "", "The cluster wide larger subnet of private ip addresses.")
    GatewayCmd.Flags().StringP("physical-interface", "", "", "The physical interface via which external connectivity is provided.")
    GatewayCmd.Flags().StringP("bridge-interface", "", "", "The OVS bridge interface via which external connectivity is provided.")
    GatewayCmd.Flags().StringP("physical-ip", "", "", "The ip address of the physical interface or bridge interface via which external connectivity is provided. This should be of the form IP/MASK.")
    GatewayCmd.Flags().StringP("node-name", "", "", "A unique node name.")
    GatewayCmd.Flags().StringP("default-gw", "", "", "The next hop IP address for your physical interface.")
    GatewayCmd.Flags().StringP("rampout-ip-subnets", "", "", "Uses this gateway to rampout traffic originating from the specified comma separated ip subnets.  Used to distribute outgoing traffic via multiple gateways.")

    return GatewayCmd
}

func initGateway(cmd *cobra.Command, args []string) error {

}

