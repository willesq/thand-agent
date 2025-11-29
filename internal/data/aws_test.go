package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetParsedAwsDocs(t *testing.T) {
	docs, err := GetParsedAwsDocs()
	require.NoError(t, err)
	assert.NotEmpty(t, docs, "AWS docs should not be empty")

	// Validate a record contains data
	for name, description := range docs {
		assert.NotEmpty(t, name, "Permission name should not be empty")
		assert.NotEmpty(t, description, "Permission description should not be empty")
		break
	}
}

func TestGetParsedAwsRoles(t *testing.T) {
	roles, err := GetParsedAwsRoles()
	require.NoError(t, err)
	assert.NotEmpty(t, roles.Policies, "AWS managed policies should not be empty")

	// Validate first policy contains data
	firstPolicy := roles.Policies[0]
	assert.NotEmpty(t, firstPolicy.Name, "Policy name should not be empty")
}
