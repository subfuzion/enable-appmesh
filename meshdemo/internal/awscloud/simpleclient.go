package awscloud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type SimpleClientOptions struct {
	Wait bool
}

type SimpleClient struct {
	AWSConfig            aws.Config
	CloudFormationClient *cloudformation.Client
	Options              SimpleClientOptions
}

func New(options SimpleClientOptions) *SimpleClient {
	return &SimpleClient{Options: options}
}

func DefaultClient(options SimpleClientOptions) (*SimpleClient, error) {
	client := New(options)
	err := client.LoadDefaultConfig()
	return client, err
}

func (c *SimpleClient) LoadDefaultConfig() error {
	config, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}
	c.AWSConfig = config
	return nil
}

func (c *SimpleClient) Deploy(name string, templateBody string) (*cloudformation.CreateStackResponse, error) {
	cf := cloudformation.New(c.AWSConfig)
	c.CloudFormationClient = cf

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

func (c *SimpleClient) Delete(name string) (*cloudformation.DeleteStackResponse, error) {
	cf := cloudformation.New(c.AWSConfig)
	c.CloudFormationClient = cf

	req := cf.DeleteStackRequest(&cloudformation.DeleteStackInput{
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
		cf.WaitUntilStackDeleteComplete(context.TODO(), &cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		})
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}
