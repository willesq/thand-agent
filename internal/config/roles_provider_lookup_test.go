package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// TestProviderRoleLookup tests the ability to specify provider:role to get a specific role lookup from a provider
func TestProviderRoleLookup(t *testing.T) {
	// Test case 1: Lookup by explicit provider name (e.g., "aws-prod:AdministratorAccess")
	// This verifies that we can inherit a role from a specific provider instance.
	t.Run("lookup by provider name", func(t *testing.T) {
		roles := map[string]models.Role{
			"custom-admin": {
				Name: "Custom Admin Role",
				// Inherit from specific provider "aws-prod"
				Inherits: []string{"aws-prod:AdministratorAccess"},
				// Base role must allow this provider
				Providers: []string{"aws-prod"},
				Permissions: models.Permissions{
					Allow: []string{"internal:check"},
				},
				Enabled: true,
			},
		}

		providers := map[string]models.Provider{
			"aws-prod": {
				Name:        "AWS Production",
				Description: "Production AWS Account",
				Provider:    "aws",
				Enabled:     true,
			},
		}

		config := newTestConfig(t, roles, providers)

		identity := &models.Identity{
			ID: "user1",
			User: &models.User{
				Username: "testuser",
				Email:    "testuser@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "custom-admin")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify that the role was resolved. The provider role name "AdministratorAccess" should be in Inherits.
		assert.Contains(t, result.Inherits, "AdministratorAccess")
		assert.Contains(t, result.Permissions.Allow, "internal:check")
	})

	// Test case 2: Lookup by provider type as name (e.g., "aws:AdministratorAccess")
	// This verifies that if a provider is named "aws" (matching its type), we can look it up.
	t.Run("lookup by provider type as name", func(t *testing.T) {
		roles := map[string]models.Role{
			"aws-admin": {
				Name:      "AWS Admin",
				Inherits:  []string{"aws:AdministratorAccess"},
				Providers: []string{"aws"},
				Enabled:   true,
			},
		}

		providers := map[string]models.Provider{
			"aws": {
				Name:     "AWS Default",
				Provider: "aws",
				Enabled:  true,
			},
		}

		config := newTestConfig(t, roles, providers)

		identity := &models.Identity{
			ID: "user1",
			User: &models.User{
				Username: "testuser",
				Email:    "testuser@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "aws-admin")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Contains(t, result.Inherits, "AdministratorAccess")
	})

	// Test case 3: Lookup by provider type with base role restriction
	// This verifies that "aws:AdministratorAccess" works when the base role uses a provider "aws-prod" (of type "aws").
	// The lookup "aws:..." implies "any provider of type aws that is allowed by the base role".
	t.Run("lookup by provider type with base role restriction", func(t *testing.T) {
		roles := map[string]models.Role{
			"aws-admin": {
				Name:      "AWS Admin",
				Inherits:  []string{"aws:AdministratorAccess"},
				Providers: []string{"aws-prod"}, // Restricted to aws-prod
				Enabled:   true,
			},
		}

		providers := map[string]models.Provider{
			"aws-prod": {
				Name:     "AWS Production",
				Provider: "aws",
				Enabled:  true,
			},
		}

		config := newTestConfig(t, roles, providers)

		identity := &models.Identity{
			ID: "user1",
			User: &models.User{
				Username: "testuser",
				Email:    "testuser@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "aws-admin")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Contains(t, result.Inherits, "AdministratorAccess")
	})

	// Test case 4: Lookup by provider name mismatch with base role
	// This verifies that "aws-dev:AdministratorAccess" is SKIPPED if the base role only allows "aws-prod".
	// Inherited roles must belong to one of the base role's providers.
	t.Run("lookup by provider name mismatch with base role", func(t *testing.T) {
		roles := map[string]models.Role{
			"aws-admin": {
				Name:      "AWS Admin",
				Inherits:  []string{"aws-dev:AdministratorAccess"},
				Providers: []string{"aws-prod"}, // Restricted to aws-prod
				Enabled:   true,
			},
		}

		providers := map[string]models.Provider{
			"aws-prod": {
				Name:     "AWS Production",
				Provider: "aws",
				Enabled:  true,
			},
			"aws-dev": {
				Name:     "AWS Development",
				Provider: "aws",
				Enabled:  true,
			},
		}

		config := newTestConfig(t, roles, providers)

		identity := &models.Identity{
			ID: "user1",
			User: &models.User{
				Username: "testuser",
				Email:    "testuser@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "aws-admin")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should NOT contain "AdministratorAccess" because it was skipped due to provider mismatch
		assert.NotContains(t, result.Inherits, "AdministratorAccess")
		assert.NotContains(t, result.Inherits, "aws-dev:AdministratorAccess")
	})
}
