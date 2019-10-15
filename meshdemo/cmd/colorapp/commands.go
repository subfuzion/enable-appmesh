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
// commands.go processes CLI commands initialized in cli.go.
// For each command *, *Handler functions are called by the CLI; they
// are responsible for processing command line flags, args, and the
// environment, setting up the call to the corresponding * function to
// perform the desired action.
package main

import (
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
	"github.com/spf13/cobra"

	"github.com/subfuzion/meshdemo/internal/awscloud"
	"github.com/subfuzion/meshdemo/pkg/io"
)

func newClient(cmd *cobra.Command) *awscloud.SimpleClient {
	// wait is either a flag that is available on either this command
	// or a persistent flag on the parent command for commands that
	// support blocking
	wait, _ := cmd.Flags().GetBool("wait")

	client, err := awscloud.NewClient(&awscloud.SimpleClientOptions{
		LoadDefaultConfig: true,
		Wait:              wait,
	})
	if err != nil {
		io.Failed("Unable to load AWS config: %s", err)
		os.Exit(1)
	}
	return client
}

func createStackHandler(cmd *cobra.Command, args []string) {
	createStack(newClient(cmd), &awscloud.CreateStackOptions{
		Name:         "demo",
		TemplatePath: "demo.yaml",
		Parameters:   nil,
	})
}

func createStack(client *awscloud.SimpleClient, options *awscloud.CreateStackOptions) {
	stackName := options.Name
	templateBody := tmpl.Read(options.TemplatePath)

	io.Step("Creating stack (%s)...", stackName)

	resp, err := client.CreateStack(stackName, templateBody)
	if err != nil {
		io.Failed("Unable to create stack (%s): %s", stackName, err)
		os.Exit(1)
	}

	// if the request blocked until finished and there were no errors, then we can output final status here
	if client.Options.Wait {
		outputs, err := client.GetStackOutput(stackName)
		if err != nil {
			io.Alert("Request to get stack outputs failed for stack: %s", stackName)
		}

		if val, exists := outputs["ClusterName"]; exists {
			io.Status("%s.%s = %s", stackName, "ClusterName", val)
		}

		if val, exists := outputs["URL"]; exists {
			io.Status("%s.%s = %s", stackName, "URL", val)
		}

		io.Success("Created stack (%s): %s", stackName, aws.StringValue(resp.StackId))

	}
}

func deleteStackHandler(cmd *cobra.Command, args []string) {
	deleteStack(newClient(cmd), &awscloud.DeleteStackOptions{
		Name: "demo",
	})
}

func deleteStack(client *awscloud.SimpleClient, options *awscloud.DeleteStackOptions) {
	stackName := options.Name

	io.Step("Deleting stack (%s)...", stackName)

	_, err := client.DeleteStack(stackName)
	if err != nil {
		io.Failed("Unable to delete stack (%s): %s", stackName, err)
		os.Exit(1)
	}
	if client.Options.Wait {
		io.Success("Deleted stack (%s)", stackName)
	}
}


type WeightMap map[string]int

type RollingUpdateSpec struct {
	// Increment is a value between 0-100 percent for rolling out an updated in incremental stages.
	// A value of either 0 or 100 effectively disables a rolling update, meaning that the update
	// is applied immediately.
	Increment int

	// Interval is the period to wait before applying the next update stage.
	Interval time.Duration
}

// UpdateRouteCommandOptions contains settings for updating App Mesh routes.
// NOTE: this, of course, is very specific to the Color App demo.
type UpdateRouteCommandOptions struct {
	// MeshName is the name of the App Mesh mesh to use.
	MeshName string

	// RouteName is the name of the App Mesh route to use.
	RouteName string

	// VirtualRouterName is the name of the App Mesh virtual router to use.
	VirtualRouterName string

	// Map of color to weight to apply to the color virtual nodes.
	Weights WeightMap

	// RollingUpdate affects the percentage and rate of incremental updates.
	RollingUpdate *RollingUpdateSpec
}

func updateRouteHandler(cmd *cobra.Command, args []string) {
	meshName := "demo"
	routeName := "color-route"
	virtualRouterName := "colorteller-vr"
	weights := WeightMap{}

	blue, err := cmd.Flags().GetInt("blue")
	if err != nil {
		io.Fatal(1, "bad value for --blue option: %s", err)
	}
	weights["blue-vn"] = blue

	green, err := cmd.Flags().GetInt("green")
	if err != nil {
		io.Fatal(1, "bad value for --green option: %s", err)
	}
	weights["green-vn"] = green

	red, err := cmd.Flags().GetInt("red")
	if err != nil {
		io.Fatal(1, "bad value for --red option: %s", err)
	}
	weights["red-vn"] = red

	rolling, err := cmd.Flags().GetInt("rolling")
	if err != nil {
		io.Fatal(1, "bad value for --rolling option: %s", err)
	}
	interval, err := cmd.Flags().GetInt("interval")
	if err != nil {
		io.Fatal(1, "bad value for --interval option: %s", err)
	}

	updateRoute(newClient(cmd), &UpdateRouteCommandOptions{
		MeshName:        meshName,
		RouteName:       routeName,
		VirtualRouterName: virtualRouterName,
		Weights: weights,
		RollingUpdate: &RollingUpdateSpec{
			Increment: rolling,
			Interval:  time.Duration(interval) * time.Second,
		},
	})
}

func updateRoute(client *awscloud.SimpleClient, options *UpdateRouteCommandOptions) {
	routeSpec := BuildRouteSpec(options)

	input := &appmesh.UpdateRouteInput{
		ClientToken: nil,
		MeshName:    aws.String(options.MeshName),
		RouteName:   aws.String(options.RouteName),
		Spec: routeSpec,
		VirtualRouterName: aws.String(options.VirtualRouterName),
	}

	io.Step("Updating route...")

	resp, err := client.UpdateRoute(input)
	if err != nil {
		io.Failed("Unable to update route: %s", err)
		os.Exit(1)
	}
	io.Success("Updated route: %s\n%s",
		options.RouteName,
		formatUpdateRouteResponse(resp))
}

func BuildRouteSpec(options *UpdateRouteCommandOptions) *appmesh.RouteSpec {
	if len(options.Weights) == 0 {
		io.Fatal(1, "must supply at least one weighted target (blue|green|red)")
	}

	weightedTargets := []appmesh.WeightedTarget{}
	for node, weight := range options.Weights {
		if weight > 0 {
			weightedTargets = append(weightedTargets, appmesh.WeightedTarget{
				VirtualNode: aws.String(node),
				Weight:      aws.Int64(int64(weight)),
			})
		}
	}

	routeMatch := &appmesh.HttpRouteMatch{
		Prefix:  aws.String("/"),
	}

	spec := &appmesh.RouteSpec{
		HttpRoute: &appmesh.HttpRoute{
			Action: &appmesh.HttpRouteAction{
				WeightedTargets: weightedTargets,
			},
			Match: routeMatch,
		},
	}

	return spec
}

func formatUpdateRouteResponse(resp *appmesh.UpdateRouteResponse) string {
	sb := &strings.Builder{}
	t := tmpl.Parse("update_route_response.tmpl")
	t.Execute(sb, resp.Route)
	return sb.String()
}

func getStackUrlHandler(cmd *cobra.Command, args []string) {
	getStackUrl(newClient(cmd), "demo")
}

func getStackUrl(client *awscloud.SimpleClient, stackName string) {
	output, err := client.GetStackOutput(stackName, "URL")
	if err != nil {
		io.Alert("Request to get stack output (%s) failed for stack: %s", "URL", stackName)
	}
	if len(output) == 0 {
		io.Failed("missing URL for stack: %s", stackName)
	}

	io.Info("%s.%s = %s", stackName, "URL", output["URL"])
}
