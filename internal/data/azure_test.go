package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetParsedAzureRoles(t *testing.T) {
	roles, err := GetParsedAzureRoles()
	require.NoError(t, err)
	assert.NotEmpty(t, roles, "Azure built-in roles should not be empty")

	// Validate first role contains data
	firstRole := roles[0]
	assert.NotEmpty(t, firstRole.Name, "Role name should not be empty")
	assert.NotEmpty(t, firstRole.Description, "Role description should not be empty")
}

func TestGetParsedAzurePermissions(t *testing.T) {
	permissions, err := GetParsedAzurePermissions()
	require.NoError(t, err)
	assert.NotEmpty(t, permissions, "Azure permissions should not be empty")

	// Validate first permission contains data
	firstPermission := permissions[0]
	assert.NotEmpty(t, firstPermission.Name, "Permission name should not be empty")
}
