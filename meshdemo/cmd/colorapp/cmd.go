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

// Command sets up the entire CLI command structure.
// It returns the root command.
func Command() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "colorapp",
		Short: "CLI for demonstrating App Mesh",
		Run: func(cmd *cobra.Command, args []string) {
			io.Printf(cmd.UsageString())
		},
	}
	cmd.PersistentFlags().StringVar(&configuration.ConfigFile, "config", "", "config file (default is $HOME/.colorapp.yaml)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// cmd config
	cmd.AddCommand((func() *cobra.Command {
		var cmd = &cobra.Command{
			Use:   "config",
			Short: "Print config file in use",
			Run: func(cmd *cobra.Command, args []string) {
				io.Info("config called")
			},
		}
		// TODO: config specific flags
		return cmd
	})())

	// cmd create
	cmd.AddCommand((func() *cobra.Command {
		var cmd = &cobra.Command{
			Use:   "create",
			Short: "Create a config file",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				io.Info("create called")
			},
		}
		// TODO: create specific flags
		return cmd
	})())

	// init AWS client commands
	func() {
		var stackName = "demo"
		var wait = false
		var init = func() *awscloud.SimpleClient {
			client, err := awscloud.DefaultClient(awscloud.SimpleClientOptions{Wait: wait})
			if err != nil {
				io.Failed("Unable to load AWS config: %s", err)
				os.Exit(1)
			}
			return client
		}

		// cmd deploy
		cmd.AddCommand((func() *cobra.Command {
			var cmd = &cobra.Command{
				Use:   "deploy",
				Short: "Deploy AWS resources",
			}
			cmd.PersistentFlags().BoolVarP(&wait, "wait", "w", false, "if set, command blocks until operation completes")

			// cmd deploy stack
			cmd.AddCommand((func() *cobra.Command {
				var cmd = &cobra.Command{
					Use:   "stack",
					Short: "Deploy CloudFormation stack",
					Run: func(cmd *cobra.Command, args []string) {
						templateBody := tpl.Read("demo.yaml")

						var client = init()
						io.Step("Deploying stack (%s)...", stackName)
						resp, err := client.Deploy(stackName, templateBody)
						if err != nil {
							io.Failed("Unable to deploy stack (%s): %s", stackName, err)
							os.Exit(1)
						}
						if client.Options.Wait {
							io.Success("Deployed stack (%s): %s", stackName, aws.StringValue(resp.StackId))
						}
					},
				}
				// TODO: map deploy flags to stack template property overrides
				return cmd
			})())

			return cmd
		})())

		// cmd delete
		cmd.AddCommand((func() *cobra.Command {
			var cmd = &cobra.Command{
				Use:   "delete",
				Short: "Delete AWS resources",
			}
			cmd.PersistentFlags().BoolVarP(&wait, "wait", "w", false, "if set, command blocks until operation completes")

			// cmd delete stack
			cmd.AddCommand((func() *cobra.Command {
				var cmd = &cobra.Command{
					Use:   "stack",
					Short: "Deploy CloudFormation stack",
					Run: func(cmd *cobra.Command, args []string) {
						var client = init()
						io.Step("Deleting stack (%s)...", stackName)
						_, err := client.Delete(stackName)
						if err != nil {
							io.Failed("Unable to delete stack (%s): %s", stackName, err)
							os.Exit(1)
						}
						if client.Options.Wait {
							io.Success("Deleted stack (%s)", stackName)
						}
					},
				}
				// TODO: delete specific flags
				return cmd
			})())

			return cmd
		})())

	}() // aws client commands

	return cmd
}
