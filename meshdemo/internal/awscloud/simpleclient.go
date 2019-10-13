package awscloud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type SimpleClient struct {
	AWSConfig            aws.Config
	CloudFormationClient *cloudformation.Client
}

func New() *SimpleClient {
	return &SimpleClient{}
}

func DefaultClient() (*SimpleClient, error) {
	client := &SimpleClient{}
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

	return req.Send(context.TODO())
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

	return req.Send(context.TODO())
}
