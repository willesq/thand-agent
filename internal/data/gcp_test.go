package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetParsedGcpRoles(t *testing.T) {
	roles, err := GetParsedGcpRoles()
	require.NoError(t, err)
	assert.NotEmpty(t, roles, "GCP predefined roles should not be empty")

	// Validate first role contains data
	firstRole := roles[0]
	assert.NotEmpty(t, firstRole.Name, "Role name should not be empty")
	assert.NotEmpty(t, firstRole.Title, "Role title should not be empty")
}
