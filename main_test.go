package main

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

type testMocks struct{}

func (m *testMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	tokenName := args.TypeToken + ":" + args.Name
	recordedResources = append(recordedResources, tokenName)

	// We must ensure the Docker image resource returns an "ImageName"
	// so the Deployment doesn't crash when it tries to read it.
	outputs := args.Inputs.Copy()
	if args.TypeToken == "docker:index/image:Image" {
		// Mock the generated image name that the Deployment expects
		outputs["imageName"] = resource.NewStringProperty("localhost:32000/counter-server:latest")
	}

	return args.Name, outputs, nil
}

func (m *testMocks) Call(_ pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

var recordedResources []string

func TestCreateResources(t *testing.T) {
	recordedResources = []string{}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		return createResources(ctx)
	}, pulumi.WithMocks("project", "stack", &testMocks{}))

	assert.NoError(t, err)

	// UPDATED: Added the Docker image to the list of expected resources
	expected := []string{
		"kubernetes:core/v1:PersistentVolumeClaim:sqlite-pvc",
		"docker:index/image:Image:counter-image", // The new Build-and-Push resource
		"kubernetes:apps/v1:Deployment:counter-deployment",
		"kubernetes:core/v1:Service:counter-service",
	}

	for _, exp := range expected {
		assert.Contains(t, recordedResources, exp, "Missing resource: %s", exp)
	}
}
