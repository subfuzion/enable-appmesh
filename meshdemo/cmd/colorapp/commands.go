package main

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"github.com/subfuzion/meshdemo/internal/awscloud"
	"github.com/subfuzion/meshdemo/pkg/io"
)

func Client(options *awscloud.SimpleClientOptions) *awscloud.SimpleClient {
	client, err := awscloud.NewClient(options)
	if err != nil {
		io.Failed("Unable to load AWS config: %s", err)
		os.Exit(1)
	}
	return client
}

func getSimpleClientOptions(cmd *cobra.Command) *awscloud.SimpleClientOptions {
	// wait is either a flag that is available on either this command
	// or a persistent flag on the parent command for commands that
	// support blocking
	wait, _ := cmd.Flags().GetBool("wait")

	return &awscloud.SimpleClientOptions{
		LoadDefaultConfig: true,
		Wait: wait,
	}
}

func CreateStackHandler(cmd *cobra.Command, args []string) {
	CreateStack(getSimpleClientOptions(cmd), &awscloud.CreateStackOptions{
		Name:         "demo",
		TemplatePath: "demo.yaml",
		Parameters:   nil,
	})
}

func CreateStack(clientOptions *awscloud.SimpleClientOptions, options *awscloud.CreateStackOptions) {
	client := Client(clientOptions)
	stackName := options.Name
	templateBody := tpl.Read(options.TemplatePath)

	io.Step("Creating stack (%s)...", stackName)

	resp, err := client.CreateStack(stackName, templateBody)
	if err != nil {
		io.Failed("Unable to create stack (%s): %s", stackName, err)
		os.Exit(1)
	}
	if client.Options.Wait {
		io.Success("Created stack (%s): %s", stackName, aws.StringValue(resp.StackId))
	}
}

func DeleteStackHandler(cmd *cobra.Command, args []string) {
	DeleteStack(getSimpleClientOptions(cmd), &awscloud.DeleteStackOptions{
		Name:         "demo",
	})
}

func DeleteStack(clientOptions *awscloud.SimpleClientOptions, options *awscloud.DeleteStackOptions) {
	client := Client(clientOptions)
	stackName := options.Name

	io.Step("Deleteing stack (%s)...", stackName)

	_, err := client.Delete(stackName)
	if err != nil {
		io.Failed("Unable to delete stack (%s): %s", stackName, err)
		os.Exit(1)
	}
	if client.Options.Wait {
		io.Success("Deleted stack (%s)", stackName)
	}
}

