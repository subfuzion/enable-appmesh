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
package config

import (
	"github.com/spf13/cobra"
	"github.com/subfuzion/meshdemo/internal/configuration"
	"github.com/subfuzion/meshdemo/pkg/io"
)

// ConfigCmd represents the config command
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Print config file in use",
	Long: `Print the config file in use`,
	Run: func(cmd *cobra.Command, args []string) {
		c := configuration.ConfigFileUsed()
		if c == "" {
			c = "(no config found; try `config create`)"
		}
		io.Info("Current config file: %s", c)
	},
}

func init() {
	Cmd.AddCommand(createCmd)
}
