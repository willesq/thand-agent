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

		// Updated behavior: permissions are now intelligently condensed
		expectedCondensedBehavior := []string{
			"k8s:apps/deployments:create,delete,get,list,patch,update", // merged and condensed
			"k8s:configmaps:create,delete,get,list,update",             // merged and condensed
			"k8s:pods:create,delete,get,list,update",                   // merged and condensed
			"k8s:services:create,delete,get,list,update",               // merged and condensed
		}

		// Sort both slices for comparison since map iteration order is not guaranteed
		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedCondensedBehavior)

		assert.ElementsMatch(t, expectedCondensedBehavior, result.Permissions.Allow,
			"Current behavior: permissions are merged and intelligently condensed")
	})

	t.Run("overlapping condensed permissions - intelligent merging", func(t *testing.T) {
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

		// Updated behavior: intelligent merging is now implemented
		expectedImprovedBehavior := []string{
			"k8s:pods:create,get,list,update,watch",      // merged and condensed
			"k8s:services:create,delete,get,list,update", // merged and condensed
		}

		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedImprovedBehavior)

		assert.ElementsMatch(t, expectedImprovedBehavior, result.Permissions.Allow,
			"New behavior shows intelligent merging of overlapping permissions")
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

		// Updated behavior: deny properly removes actions from condensed allow permissions
		// The "delete" action should be removed from the allow list due to the deny
		assert.Contains(t, result.Permissions.Allow, "k8s:pods:create,get,list,update")
		assert.Empty(t, result.Permissions.Deny, "Deny permissions should be empty after conflict resolution")

		// This shows the improved behavior: deny intelligently removes
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

		// Updated behavior: intelligent merging across inheritance levels
		expectedImprovedBehavior := []string{
			"k8s:pods:create,delete,get,list,patch,update",     // fully merged
			"k8s:services:create,delete,get,list,patch,update", // fully merged
			"k8s:events:get,list,watch",                        // no overlaps
			"k8s:configmaps:create,delete,get,list,update",     // no overlaps
			"k8s:secrets:create,get,list,update",               // delete removed by deny
		}

		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedImprovedBehavior)

		assert.ElementsMatch(t, expectedImprovedBehavior, result.Permissions.Allow)

		// Verify deny permissions remain to enforce security policy
		assert.ElementsMatch(t, []string{"k8s:secrets:delete"}, result.Permissions.Deny, "Explicit deny should remain to enforce security policy")
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
			{
				name: "wildcard overrides specific actions",
				input: []string{
					"ec2:describeInstances",
					"ec2:*",
					"ec2:startInstances",
					"rds:*",
					"rds:describeDBInstances",
					"s3:listBuckets",
					"s3:getBucketLocation",
				},
				expected: []string{
					"ec2:*",
					"rds:*",
					"s3:getBucketLocation,listBuckets",
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
