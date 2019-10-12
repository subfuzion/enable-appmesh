/*
Copyright © 2019 Tony Pujals <tpujals@gmail.com>

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
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/subfuzion/meshdemo/cmd/colorapp/cmd/config"
	"github.com/subfuzion/meshdemo/cmd/colorapp/cmd/stack"
	"github.com/subfuzion/meshdemo/internal/configuration"
	"github.com/subfuzion/meshdemo/pkg/io"
)

var RootCmd = &cobra.Command{
	Use:   "colorapp",
	Short: "CLI for demonstrating App Mesh",
	Long: `colorapp is a command line tool uses the ColorApp to demonstrate AWS App Mesh.`,
	Run: func(cmd *cobra.Command, args []string) {
		io.Printf(cmd.UsageString())
	},
}

// Execute starts command processing each time the CLI is used. It's called once by main.main().
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// if errors have been silenced, then print surfaced error here before exiting
		if RootCmd.SilenceErrors {
			io.Fatal(1, err)
		}
	}
}

func init() {
	configuration.Init()

	RootCmd.PersistentFlags().StringVar(&configuration.ConfigFile, "config", "", "config file (default is $HOME/.colorapp.yaml)")
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	RootCmd.AddCommand(config.Cmd)
	RootCmd.AddCommand(stack.Cmd)

}

