package cmd

import {

	"github.com/spf13/cobra"
}

func InitMaster() &cobra.Command {

	var MasterCmd = &cobra.Command{
        Use:   "master [no options!]",
        Short: "init ovn master",
        PreRun: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Inside subCmd PreRun with args: %v\n", args)
        },
        Run: func(cmd *cobra.Command, args []string) {
            if err := initMaster(cmd, args); err != nil {
            	fmt.Fprint(os.Stderr, err)
            	os.exit(1)
            }
        },
        PostRun: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Inside subCmd PostRun with args: %v\n", args)
        },
        PersistentPostRun: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Inside subCmd PersistentPostRun with args: %v\n", args)
        },
    }

    MasterCmd.Flags().StringP("cluster-ip-subnet", "", "", "The cluster wide larger subnet of private ip addresses.")
    MasterCmd.Flags().StringP("master-switch-subnet", "", "", "The smaller subnet just for master.")
    MasterCmd.Flags().StringP("node-name", "", "", "A unique node name.")

    return MasterCmd
}

func initMaster(cmd *cobra.Command, args []string) error {

}

