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
	"github.com/spf13/cobra"

	"github.com/subfuzion/meshdemo/internal/configuration"
	"github.com/subfuzion/meshdemo/pkg/io"
)

// CLI sets up the entire CLI command structure.
// It returns the root command.
func CLI() *cobra.Command {
	// root command
	var cmd = &cobra.Command{
		Use:   "colorapp",
		Short: "CLI for demonstrating App Mesh",
		Run: func(cmd *cobra.Command, args []string) {
			io.Printf(cmd.UsageString())
		},
	}
	cmd.PersistentFlags().StringVar(&configuration.ConfigFile, "config", "", "config file (default is $HOME/.colorapp.yaml)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// root command subcommands
	cmd.AddCommand(newConfigCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newDeleteCommand())

	return cmd
}

// cmd config
func newConfigCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Print config file in use",
		Run: func(cmd *cobra.Command, args []string) {
			io.Info("config called")
		},
	}
	// TODO: config specific flags
	cmd.AddCommand(newConfigCreateCommand())
	return cmd
}

// cmd config create
func newConfigCreateCommand() *cobra.Command {
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
}

// cmd create
func newCreateCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create AWS resource",
	}
	cmd.PersistentFlags().BoolP("wait", "w", false, "if set, command blocks until operation completes")
	cmd.AddCommand(newCreateStackCommand())
	return cmd
}

// cmd create stack
func newCreateStackCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "stack",
		Short: "Create CloudFormation stack",
		Run:   createStackHandler,
	}
	// TODO: map deploy flags to stack template property overrides
	return cmd
}

// cmd delete
func newDeleteCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete AWS resource",
	}
	cmd.PersistentFlags().BoolP("wait", "w", false, "if set, command blocks until operation completes")
	cmd.AddCommand(newDeleteStackCommand())
	return cmd
}

// cmd delete stack
func newDeleteStackCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "stack",
		Short: "Delete CloudFormation stack",
		Run:   deleteStackHandler,
	}
	// TODO: delete specific flags
	return cmd
}

// cmd update
func newUpdateCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Update AWS resource",
	}
	// TODO: update specific flags
	cmd.AddCommand(newUpdateRouteCommand())
	return cmd
}

// cmd update route
func newUpdateRouteCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "route",
		Short: "Update App Mesh route",
		Run:   updateRouteHandler,
	}
	cmd.Flags().IntP("blue", "b", 0, "set the weight for the blue virtual node")
	cmd.Flags().IntP("green", "g", 0, "set the weight for the green virtual node")
	cmd.Flags().IntP("red", "r", 0, "set the weight for the red virtual node")
	cmd.Flags().Int("rolling", 0, "set increment (as a percentage) for rolling update (either 0 or 100 disables")
	cmd.Flags().Int("interval", 0, "set interval (in seconds) between each rolling update")
	return cmd
}
