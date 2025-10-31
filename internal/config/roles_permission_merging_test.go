package config

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// TestPermissionMerging tests how permissions are merged during role inheritance,
// with specific attention to condensed actions like k8s:pods:get,list,watch
func TestPermissionMerging(t *testing.T) {
	t.Run("basic string-level merging - current behavior", func(t *testing.T) {
		roles := map[string]models.Role{
			"base-reader": {
				Name: "Base Reader",
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:get,list",
						"k8s:services:get,list",
						"k8s:configmaps:get,list",
						"k8s:apps/deployments:get,list",
					},
				},
				Enabled: true,
			},
			"enhanced-dev": {
				Name:     "Enhanced Developer",
				Inherits: []string{"base-reader"},
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:create,update,delete",
						"k8s:services:create,update,delete",
						"k8s:configmaps:create,update,delete",
						"k8s:apps/deployments:create,update,patch,delete",
					},
				},
				Enabled: true,
			},
		}

		config := &Config{
			Roles: RoleConfig{
				Definitions: roles,
			},
		}

		identity := &models.Identity{
			ID: "dev-user",
			User: &models.User{
				Username: "developer",
				Email:    "dev@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "enhanced-dev")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Current behavior: permissions are merged as separate strings
		// This demonstrates the issue - we get separate permission strings
		// instead of merged condensed actions
		expectedCurrentBehavior := []string{
			"k8s:pods:get,list",                               // from base-reader
			"k8s:services:get,list",                           // from base-reader
			"k8s:configmaps:get,list",                         // from base-reader
			"k8s:apps/deployments:get,list",                   // from base-reader
			"k8s:pods:create,update,delete",                   // from enhanced-dev
			"k8s:services:create,update,delete",               // from enhanced-dev
			"k8s:configmaps:create,update,delete",             // from enhanced-dev
			"k8s:apps/deployments:create,update,patch,delete", // from enhanced-dev
		}

		// Sort both slices for comparison since map iteration order is not guaranteed
		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedCurrentBehavior)

		assert.ElementsMatch(t, expectedCurrentBehavior, result.Permissions.Allow,
			"Current behavior: permissions are merged as separate strings")
	})

	t.Run("overlapping condensed permissions - demonstrates limitation", func(t *testing.T) {
		roles := map[string]models.Role{
			"partial-access": {
				Name: "Partial Access",
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:get,list,watch",
						"k8s:services:get",
					},
				},
				Enabled: true,
			},
			"extended-access": {
				Name:     "Extended Access",
				Inherits: []string{"partial-access"},
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:create,update",                 // Overlaps with get,list,watch
						"k8s:services:list,create,update,delete", // Overlaps with get
					},
				},
				Enabled: true,
			},
		}

		config := &Config{
			Roles: RoleConfig{
				Definitions: roles,
			},
		}

		identity := &models.Identity{
			ID: "test-user",
			User: &models.User{
				Username: "testuser",
				Email:    "test@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "extended-access")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Current behavior: separate strings, no intelligent merging
		expectedCurrentBehavior := []string{
			"k8s:pods:get,list,watch",                // from partial-access
			"k8s:services:get",                       // from partial-access
			"k8s:pods:create,update",                 // from extended-access
			"k8s:services:list,create,update,delete", // from extended-access
		}

		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedCurrentBehavior)

		assert.ElementsMatch(t, expectedCurrentBehavior, result.Permissions.Allow,
			"Current behavior shows duplicate/overlapping permissions without intelligent merging")

		// What we WANT to see (this test will initially fail):
		// expectedIdealBehavior := []string{
		//     "k8s:pods:get,list,watch,create,update",      // merged and condensed
		//     "k8s:services:get,list,create,update,delete", // merged and condensed
		// }
		// This demonstrates why we need smart permission merging
	})

	t.Run("deny permissions with condensed actions", func(t *testing.T) {
		roles := map[string]models.Role{
			"base-role": {
				Name: "Base Role",
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:get,list,create,update,delete",
					},
					Deny: []string{
						"k8s:pods:delete", // Should be removed from the condensed allow
					},
				},
				Enabled: true,
			},
		}

		config := &Config{
			Roles: RoleConfig{
				Definitions: roles,
			},
		}

		identity := &models.Identity{
			ID: "test-user",
			User: &models.User{
				Username: "testuser",
				Email:    "test@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "base-role")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Current behavior: deny doesn't affect condensed allow permission
		assert.Contains(t, result.Permissions.Allow, "k8s:pods:get,list,create,update,delete")
		assert.Contains(t, result.Permissions.Deny, "k8s:pods:delete")

		// This shows the limitation: the deny doesn't intelligently remove
		// the 'delete' action from the condensed allow permission
	})

	t.Run("complex inheritance with condensed permissions", func(t *testing.T) {
		roles := map[string]models.Role{
			"viewer": {
				Name: "Viewer",
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:get,list",
						"k8s:services:get,list",
						"k8s:events:get,list,watch",
					},
				},
				Enabled: true,
			},
			"editor": {
				Name:     "Editor",
				Inherits: []string{"viewer"},
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:create,update,patch",
						"k8s:services:create,update,patch",
						"k8s:configmaps:get,list,create,update,delete",
					},
				},
				Enabled: true,
			},
			"admin": {
				Name:     "Admin",
				Inherits: []string{"editor"},
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:delete",
						"k8s:services:delete",
						"k8s:secrets:get,list,create,update,delete",
					},
					Deny: []string{
						"k8s:secrets:delete", // Admin can't delete secrets
					},
				},
				Enabled: true,
			},
		}

		config := &Config{
			Roles: RoleConfig{
				Definitions: roles,
			},
		}

		identity := &models.Identity{
			ID: "admin-user",
			User: &models.User{
				Username: "admin",
				Email:    "admin@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "admin")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Current behavior: all permissions as separate strings
		expectedCurrentBehavior := []string{
			// From viewer
			"k8s:pods:get,list",
			"k8s:services:get,list",
			"k8s:events:get,list,watch",
			// From editor
			"k8s:pods:create,update,patch",
			"k8s:services:create,update,patch",
			"k8s:configmaps:get,list,create,update,delete",
			// From admin
			"k8s:pods:delete",
			"k8s:services:delete",
			"k8s:secrets:get,list,create,update,delete",
		}

		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedCurrentBehavior)

		assert.ElementsMatch(t, expectedCurrentBehavior, result.Permissions.Allow)

		// Verify deny permissions
		assert.Contains(t, result.Permissions.Deny, "k8s:secrets:delete")

		// What we WANT to see eventually:
		// expectedIdealBehavior := []string{
		//     "k8s:pods:get,list,create,update,patch,delete",       // fully merged
		//     "k8s:services:get,list,create,update,patch,delete",   // fully merged
		//     "k8s:events:get,list,watch",                         // no overlaps
		//     "k8s:configmaps:get,list,create,update,delete",      // no overlaps
		//     "k8s:secrets:get,list,create,update",                // delete removed by deny
		// }
	})
}

// TestCondensedActionParsing tests the helper functions for parsing condensed actions
func TestCondensedActionParsing(t *testing.T) {
	t.Run("expand condensed actions", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected []string
		}{
			{
				name:     "simple condensed actions",
				input:    "k8s:pods:get,list,watch",
				expected: []string{"k8s:pods:get", "k8s:pods:list", "k8s:pods:watch"},
			},
			{
				name:     "single action",
				input:    "k8s:pods:get",
				expected: []string{"k8s:pods:get"},
			},
			{
				name:     "no actions part",
				input:    "k8s:pods",
				expected: []string{"k8s:pods"},
			},
			{
				name:     "complex resource path",
				input:    "k8s:apps/deployments:get,list,create,update,patch,delete",
				expected: []string{"k8s:apps/deployments:get", "k8s:apps/deployments:list", "k8s:apps/deployments:create", "k8s:apps/deployments:update", "k8s:apps/deployments:patch", "k8s:apps/deployments:delete"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := expandCondensedActions(tc.input)
				assert.ElementsMatch(t, tc.expected, result)
			})
		}
	})

	t.Run("condense actions", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    []string
			expected []string
		}{
			{
				name: "condense related actions",
				input: []string{
					"k8s:pods:get",
					"k8s:pods:list",
					"k8s:pods:watch",
					"k8s:services:get",
					"k8s:services:list",
				},
				expected: []string{
					"k8s:pods:get,list,watch",
					"k8s:services:get,list",
				},
			},
			{
				name: "single actions remain single",
				input: []string{
					"k8s:pods:get",
					"k8s:services:delete",
				},
				expected: []string{
					"k8s:pods:get",
					"k8s:services:delete",
				},
			},
			{
				name: "mixed condensed and single",
				input: []string{
					"k8s:pods:get",
					"k8s:pods:list",
					"k8s:services:delete",
					"k8s:configmaps:create",
				},
				expected: []string{
					"k8s:pods:get,list",
					"k8s:services:delete",
					"k8s:configmaps:create",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := condenseActions(tc.input)
				sort.Strings(result)
				sort.Strings(tc.expected)
				assert.ElementsMatch(t, tc.expected, result)
			})
		}
	})
}

// Note: Helper functions expandCondensedActions and condenseActions 
// are now implemented in roles.go