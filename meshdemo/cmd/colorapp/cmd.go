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
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/subfuzion/meshdemo/internal/configuration"
	"github.com/subfuzion/meshdemo/internal/template"
	"github.com/subfuzion/meshdemo/pkg/fs"
	"github.com/subfuzion/meshdemo/pkg/io"
)

func init() {
	configuration.Init()

	cmd.PersistentFlags().StringVar(&configuration.ConfigFile, "config", "", "config file (default is $HOME/.colorapp.yaml)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cmd.AddCommand(configCmd)
	cmd.AddCommand(createCmd)
	cmd.AddCommand(deployCmd)

}

// Execute starts command processing each time the CLI is used. It's called once by main.main().
func Execute() {
	if err := cmd.Execute(); err != nil {
		// if errors have been silenced, then print surfaced error here before exiting
		if cmd.SilenceErrors {
			io.Fatal(1, err)
		}
	}
}

var cmd = &cobra.Command{
	Use:   "colorapp",
	Short: "CLI for demonstrating App Mesh",
	Run: func(cmd *cobra.Command, args []string) {
		io.Printf(cmd.UsageString())
	},
}


var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print config file in use",
	Run: func(cmd *cobra.Command, args []string) {
		c := configuration.ConfigFileUsed()
		if c == "" {
			c = "(no config found; try `config create`)"
		}
		io.Info("Current config file: %s", c)
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a config file",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f := args[0]
		err := fs.EnsureFile(f)
		if err != nil {
			io.Fatal(1, "can't create file: %s", err)
		}
		io.Step("creating config file: %s", f)
		io.Step("creating config file: %s", f)
		io.Step("creating config file: %s", f)
		io.Status("created config file: %s", f)
		io.Success("created config file: %s", f)
		io.Warn("created config file: %s", f)
		io.Alert("created config file: %s", f)
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")
		s := template.Read("./demo.yaml")
		io.Printf(s)
	},
}

