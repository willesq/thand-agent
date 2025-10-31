package config

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// TestIntelligentPermissionMerging tests the improved permission merging logic
// that properly handles condensed actions like k8s:pods:get,list,watch
func TestIntelligentPermissionMerging(t *testing.T) {
	t.Run("intelligent merging of condensed permissions", func(t *testing.T) {
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

		// With intelligent merging, we should get condensed permissions that merge actions
		expectedMergedBehavior := []string{
			"k8s:apps/deployments:create,delete,get,list,patch,update", // merged and condensed
			"k8s:configmaps:create,delete,get,list,update",             // merged and condensed
			"k8s:pods:create,delete,get,list,update",                   // merged and condensed
			"k8s:services:create,delete,get,list,update",               // merged and condensed
		}

		// Sort both for comparison
		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedMergedBehavior)

		assert.ElementsMatch(t, expectedMergedBehavior, result.Permissions.Allow,
			"Intelligent merging should combine overlapping actions into condensed permissions")
	})

	t.Run("deny permissions properly remove actions from condensed allows", func(t *testing.T) {
		roles := map[string]models.Role{
			"base-role": {
				Name: "Base Role",
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:get,list,create,update,delete",
					},
					Deny: []string{
						"k8s:pods:delete", // Should remove delete from the condensed allow
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

		// Should have allow permissions with delete action removed
		expectedAllow := []string{"k8s:pods:create,get,list,update"}
		assert.ElementsMatch(t, expectedAllow, result.Permissions.Allow)

		// Should have empty deny permissions since the conflict was resolved by removing
		// the conflicting action from allow and keeping only non-conflicting denies
		assert.Empty(t, result.Permissions.Deny,
			"Deny permissions should be empty when all deny actions conflict with allow actions")
	})

	t.Run("complex inheritance with intelligent merging", func(t *testing.T) {
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

		// Should have intelligently merged permissions
		expectedAllow := []string{
			"k8s:configmaps:create,delete,get,list,update",     // from editor
			"k8s:events:get,list,watch",                        // from viewer (no overlaps)
			"k8s:pods:create,delete,get,list,patch,update",     // merged from all levels
			"k8s:secrets:create,get,list,update",               // delete removed by deny
			"k8s:services:create,delete,get,list,patch,update", // merged from all levels
		}

		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedAllow)

		assert.ElementsMatch(t, expectedAllow, result.Permissions.Allow)

		// Should have deny permissions to enforce security policy
		assert.ElementsMatch(t, []string{"k8s:secrets:delete"}, result.Permissions.Deny)
	})

	t.Run("mixed condensed and individual permissions", func(t *testing.T) {
		roles := map[string]models.Role{
			"mixed-role": {
				Name: "Mixed Role",
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:get,list,watch",
						"k8s:services:get", // individual
						"k8s:events:watch", // individual
						"k8s:configmaps:create,update",
					},
				},
				Enabled: true,
			},
			"extended-mixed": {
				Name:     "Extended Mixed",
				Inherits: []string{"mixed-role"},
				Permissions: models.Permissions{
					Allow: []string{
						"k8s:pods:create",          // should merge with get,list,watch
						"k8s:services:list,create", // should merge with get
						"k8s:events:get,list",      // should merge with watch
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
			ID: "mixed-user",
			User: &models.User{
				Username: "mixeduser",
				Email:    "mixed@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "extended-mixed")
		require.NoError(t, err)
		require.NotNil(t, result)

		expectedAllow := []string{
			"k8s:configmaps:create,update",   // no overlaps
			"k8s:events:get,list,watch",      // merged
			"k8s:pods:create,get,list,watch", // merged
			"k8s:services:create,get,list",   // merged
		}

		sort.Strings(result.Permissions.Allow)
		sort.Strings(expectedAllow)

		assert.ElementsMatch(t, expectedAllow, result.Permissions.Allow)
	})
}
