/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"github.com/heartlock/ovn-kubernetes/pkg/cmd"
	"github.com/spf13/cobra"
)

/*
WARNING: this logic is duplicated, with minor changes, in cmd/hyperkube/kubectl.go
Any salient changes here will need to be manually reflected in that file.
*/
func Run() error {
	rootCmd := &cobra.Command{
		Use:   "kube-ovnctl [string to echo]",
		Short: "run kube-ovnctl",
		Long:  `run kube-ovnctl to init master, minion, gateway`,
	}

	materCmd := cmd.InitMaster()
	minionCmd := cmd.InitMinion()
	gatewayCmd := cmd.InitGateway()

	rootCmd.AddCommand(masterCmd)
	rootCmd.AddCommand(minionCmd)
	rootCmd.AddCommand(gatewayCmd)

	return rootCmd.Execute()
}
