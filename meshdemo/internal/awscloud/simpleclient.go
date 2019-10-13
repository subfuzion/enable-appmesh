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
	Options              *SimpleClientOptions
}

type SimpleClientOptions struct {
	LoadDefaultConfig bool
	Wait bool
}

type CreateStackOptions struct {
	// Name is the name of the stack
	Name string

	// TemplatePath is a relative path to a file inside the (embedded) _templates directory
	TemplatePath string

	// Parameters are CloudFormation template parameters used when deploying a stack
	Parameters StackParameters
}

type StackParameters map[string]string

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

func (c *SimpleClient) LoadDefaultConfig() error {
	config, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}
	c.AWSConfig = config
	return nil
}

func (c *SimpleClient) CreateStack(name string, templateBody string) (*cloudformation.CreateStackResponse, error) {
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
