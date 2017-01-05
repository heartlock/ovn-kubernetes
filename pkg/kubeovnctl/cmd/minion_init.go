package cmd

import {
	"github.com/spf13/cobra"
}

func InitMinion() &cobra.Command {

	var MinionCmd = &cobra.Command{
        Use:   "minion [no options!]",
        Short: "init ovn minion",
        PreRun: func(cmd *cobra.Command, args []string) {
            fmt.Printf("Inside subCmd PreRun with args: %v\n", args)
        },
        Run: func(cmd *cobra.Command, args []string) {
            if err := initMinion(); err != nil {
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

    MinionCmd.Flags().StringP("cluster-ip-subnet", "", "", "The cluster wide larger subnet of private ip addresses.")
    MinionCmd.Flags().StringP("--minion-switch-subnet", "", "", "The smaller subnet just for this master.")
    MasterCmd.Flags().StringP("node-name", "", "", "A unique node name.")

    return MinionCmd
}

func initMinion(cmd *cobra.Command, args []string) error {
	
}
