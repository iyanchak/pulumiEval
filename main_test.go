package main

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// 1. Corrected the struct definition
type testMocks struct{}

// 2. FIXED: Correct signature for NewResource
// Returns: (id string, state resource.PropertyMap, err error)
func (m *testMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	// Record the resource token and name for verification
	tokenName := args.TypeToken + ":" + args.Name
	recordedResources = append(recordedResources, tokenName)

	// Pulumi expects (ID, State, Error)
	// For mocks, we usually just return the name as the ID and the inputs as the state
	return args.Name, args.Inputs, nil
}

// 3. FIXED: Correct signature for Call
func (m *testMocks) Call(_ pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

var recordedResources []string

func TestCreateResources(t *testing.T) {
	// Reset the slice for each test run
	recordedResources = []string{}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Ensure createResources(ctx) is defined in your main.go
		return createResources(ctx)
	}, pulumi.WithMocks("project", "stack", &testMocks{}))

	assert.NoError(t, err)

	expected := []string{
		"kubernetes:core/v1:PersistentVolumeClaim:sqlite-pvc",
		"kubernetes:apps/v1:Deployment:counter-deployment",
		"kubernetes:core/v1:Service:counter-service",
	}

	for _, exp := range expected {
		assert.Contains(t, recordedResources, exp, "Missing resource: %s", exp)
	}
}
