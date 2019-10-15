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
// Package awscloud is a simple wrapper around a few specific packages
// (cloudformation, appmesh) from the AWS Go SDK.
package awscloud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type SimpleClient struct {
	// AWWConfig holds standard AWS configuration, such as region and credentials.
	AWSConfig aws.Config

	// Options to set client behavior.
	Options *SimpleClientOptions

	// CloudFormationClient for stack operations.
	cloudFormationClient *cloudformation.Client

	// AppMeshClient for App Mesh operations.
	appMeshClient *appmesh.Client
}

type SimpleClientOptions struct {
	LoadDefaultConfig bool
	Wait              bool
}

// StackParameters are used to set CloudFormation stack template parameters.
type StackParameters map[string]string

// CreateStackOptions contains settings for CloudFormation stack creation.
type CreateStackOptions struct {
	// Name is the name of the stack.
	Name string

	// TemplatePath is a relative path to a file inside the (embedded) _templates directory.
	TemplatePath string

	// Parameters are CloudFormation template parameters used when deploying a stack.
	Parameters StackParameters
}

// DeleteStackOptions contains settings for CloudFormation stack deletion.
type DeleteStackOptions struct {
	Name string
}

// NewClient returns a SimpleClient instance.
// If options.LoadDefaultConfig is set and there is an error loading
// the user's AWS config, then it returns an error.
func NewClient(options *SimpleClientOptions) (*SimpleClient, error) {
	client := &SimpleClient{Options: options}
	if options.LoadDefaultConfig {
		err := client.LoadDefaultConfig()
		return client, err
	}
	return client, nil
}

// LoadDefaultConfig loads AWS configuration from standard sources.
// The default configuration sources are:
// * Environment Variables
// * Shared Configuration and Shared Credentials files (`$HOME/.aws/`).
func (c *SimpleClient) LoadDefaultConfig() error {
	config, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}
	c.AWSConfig = config
	return nil
}

// CloudFormationClient gets client for stack operations.
func (c *SimpleClient) CloudFormationClient() *cloudformation.Client {
	if c.cloudFormationClient == nil {
		c.cloudFormationClient = cloudformation.New(c.AWSConfig)
	}
	return c.cloudFormationClient
}

// CreateStack creates a CloudFormation stack using a template.
// For blocking behavior, set `Wait` on the client.
func (c *SimpleClient) CreateStack(name string, templateBody string) (*cloudformation.CreateStackResponse, error) {
	cf := c.CloudFormationClient()

	req := cf.CreateStackRequest(&cloudformation.CreateStackInput{
		Capabilities:                []cloudformation.Capability{cloudformation.CapabilityCapabilityIam},
		ClientRequestToken:          nil,
		DisableRollback:             nil,
		EnableTerminationProtection: nil,
		NotificationARNs:            nil,
		OnFailure:                   "ROLLBACK",
		Parameters:                  nil,
		ResourceTypes:               nil,
		RoleARN:                     nil,
		RollbackConfiguration:       nil,
		StackName:                   aws.String(name),
		StackPolicyBody:             nil,
		StackPolicyURL:              nil,
		Tags:                        nil,
		TemplateBody:                aws.String(templateBody),
		TemplateURL:                 nil,
		TimeoutInMinutes:            nil,
	})

	resp, err := req.Send(context.TODO())
	if err != nil {
		return nil, err
	}

	if c.Options.Wait {
		err := cf.WaitUntilStackCreateComplete(context.TODO(), &cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		})
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}

// DeleteStack creates a CloudFormation stack using a template.
// For blocking behavior, set `Wait` on the client.
func (c *SimpleClient) DeleteStack(name string) (*cloudformation.DeleteStackResponse, error) {
	client := c.CloudFormationClient()

	req := client.DeleteStackRequest(&cloudformation.DeleteStackInput{
		ClientRequestToken: nil,
		RetainResources:    nil,
		RoleARN:            nil,
		StackName:          aws.String(name),
	})

	resp, err := req.Send(context.TODO())
	if err != nil {
		return nil, err
	}

	if c.Options.Wait {
		client.WaitUntilStackDeleteComplete(context.TODO(), &cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		})
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}

// AppMeshClient gets client for App Mesh operations.
func (c *SimpleClient) AppMeshClient() *appmesh.Client {
	if c.appMeshClient == nil {
		c.appMeshClient = appmesh.New(c.AWSConfig)
	}
	return c.appMeshClient
}

// UpdateRoute updates the App Mesh route.
func (c *SimpleClient) UpdateRoute(input *appmesh.UpdateRouteInput) (*appmesh.UpdateRouteResponse, error){
	client := c.AppMeshClient()

	//req := client.UpdateRouteRequest(&appmesh.UpdateRouteInput{
	//	ClientToken: nil,
	//	MeshName:    nil,
	//	RouteName:   nil,
	//	Spec: &appmesh.RouteSpec{
	//		HttpRoute: &appmesh.HttpRoute{
	//			Action: &appmesh.HttpRouteAction{
	//				WeightedTargets: []appmesh.WeightedTarget{
	//					weightedTarget,
	//				},
	//			},
	//			Match: &appmesh.HttpRouteMatch{
	//				Headers: nil,
	//				Method:  "",
	//				Prefix:  nil,
	//				Scheme:  "",
	//			},
	//			RetryPolicy: &appmesh.HttpRetryPolicy{
	//				HttpRetryEvents: nil,
	//				MaxRetries:      nil,
	//				PerRetryTimeout: nil,
	//				TcpRetryEvents:  nil,
	//			},
	//		},
	//		Priority: nil,
	//		TcpRoute: nil,
	//	},
	//	VirtualRouterName: nil,
	//})

	req := client.UpdateRouteRequest(input)
	return req.Send(context.TODO())
}

func (c *SimpleClient) GetStackOutput(name string, keys ...string) (map[string]string, error) {
	client := c.CloudFormationClient()

	req := client.DescribeStacksRequest(&cloudformation.DescribeStacksInput{
		NextToken: nil,
		StackName: aws.String(name),
	})


	resp, err := req.Send(context.TODO())
	if err != nil {
		return nil, err
	}

	outputs := resp.Stacks[0].Outputs

	if len(keys) > 0 {
		// filter will be a set that we can use to test if a key is present
		filter := map[string]struct{}{}

		// filtered will be the final outputs array
		filtered := []cloudformation.Output{}

		// add the requested keys to the filter set
		for _, k := range keys {
			filter[k] = struct{}{}
		}

		// filter the outputs array
		for _, o := range outputs {
			key := aws.StringValue(o.OutputKey)
			if _, exists := filter[key]; exists {
				filtered = append(filtered, o)
			}
		}

		outputs = filtered
	}

	results := map[string]string{}
	for _, o := range outputs {
		k := aws.StringValue(o.OutputKey)
		v := aws.StringValue(o.OutputValue)
		results[k] = v
	}

	return results, nil

}