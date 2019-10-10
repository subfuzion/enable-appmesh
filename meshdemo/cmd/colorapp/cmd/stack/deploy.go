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
	"github.com/subfuzion/meshdemo/internal/template"
	"github.com/subfuzion/meshdemo/pkg/io"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Long: `A fuller description of your command`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")
		deploy()
	},
}

func init() {
//	stackCmd.AddCommand(deployCmd)
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func deploy() {
	s := template.Read("./demo.yaml")
	io.Printf(s)
}