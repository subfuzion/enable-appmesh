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
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"github.com/subfuzion/meshdemo/internal/awscloud"
	"github.com/subfuzion/meshdemo/internal/configuration"
	"github.com/subfuzion/meshdemo/pkg/io"
)

func init() {
	configuration.Init()

	cmd.PersistentFlags().StringVar(&configuration.ConfigFile, "config", "", "config file (default is $HOME/.colorapp.yaml)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")


	cmd.AddCommand(configCmd)
	cmd.AddCommand(createCmd)

	// $ colorapp deploy
	cmd.AddCommand(deployCmd)
	deployCmd.AddCommand(deployStackCmd)

	// $ colorapp delete
	cmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteStackCmd)
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
		io.Info("configcalled")
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		io.Info("create called")
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy AWS resources",
}

var deployStackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Deploy CloudFormation stack",
	Run: func(cmd *cobra.Command, args []string) {
		templateBody := tpl.Read("demo.yaml")

		client, err := awscloud.DefaultClient()
		if err != nil {
			io.Failed("Unable to load AWS config: %s", err)
			os.Exit(1)
		}

		stackName := "demo"
		resp, err := client.Deploy(stackName, templateBody)
		if err != nil {
			io.Failed("Unable to deploy stack (%s): %s", stackName, err)
		}
		io.Success("Deploying stack (%s): %s", stackName, aws.StringValue(resp.StackId))
	},
}


var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete AWS resources",
}

var deleteStackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Deploy CloudFormation stack",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := awscloud.DefaultClient()
		if err != nil {
			io.Failed("Unable to load AWS config: %s", err)
			os.Exit(1)
		}

		stackName := "demo"
		resp, err := client.Delete(stackName)
		if err != nil {
			io.Failed("Unable to delete stack (%s): %s", stackName, err)
			os.Exit(1)
		}
		io.Success("Deleting stack (%s): %s", stackName, resp.String())
	},
}


