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
package stack

import (
	"fmt"

	"github.com/spf13/cobra"
)

// StackCmd represents the stack command
var Cmd = &cobra.Command{
	Use:   "stack",
	Short: "A brief description of your command",
	Long: `A fuller description of your command`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stack called")
	},
}

func init() {
//	cmd.RootCmd.AddCommand(stackCmd)
	// stackCmd.PersistentFlags().String("foo", "", "A help for foo")
	// stackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	Cmd.AddCommand(deployCmd)
}
