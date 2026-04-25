package infra

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

	outputs := args.Inputs.Copy()

	return args.Name, outputs, nil
}

func (m *testMocks) Call(_ pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

var recordedResources []string

func TestCreateResources(t *testing.T) {
	recordedResources = []string{}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		return CreateResources(ctx)
	}, pulumi.WithMocks("project", "stack", &testMocks{}))

	assert.NoError(t, err)

	expected := []string{
		"kubernetes:core/v1:PersistentVolumeClaim:sqlite-pvc",
		"command:local:Command:build-counter",  // The new pack build command
		"command:local:Command:import-counter", // Import to microk8s
		"kubernetes:apps/v1:Deployment:counter-deployment",
		"kubernetes:core/v1:Service:counter-service",
	}

	for _, exp := range expected {
		assert.Contains(t, recordedResources, exp, "Missing resource: %s", exp)
	}
}
