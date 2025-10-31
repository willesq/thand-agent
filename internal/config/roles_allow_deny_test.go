package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

func TestAllowDenyConflictResolution(t *testing.T) {
	t.Run("parent allow overrides child deny", func(t *testing.T) {
		roles := map[string]models.Role{
			"child": {
				Name: "Child Role",
				Permissions: models.Permissions{
					Allow: []string{"read", "list"},
					Deny:  []string{"write", "delete"},
				},
				Resources: models.Resources{
					Allow: []string{"bucket1"},
					Deny:  []string{"bucket2", "bucket3"},
				},
				Enabled: true,
			},
			"parent": {
				Name:     "Parent Role",
				Inherits: []string{"child"},
				Permissions: models.Permissions{
					Allow: []string{"write"}, // This should override child's deny for "write"
					Deny:  []string{"read"},  // This should override child's allow for "read"
				},
				Resources: models.Resources{
					Allow: []string{"bucket2"}, // This should override child's deny for "bucket2"
					Deny:  []string{"bucket1"}, // This should override child's allow for "bucket1"
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
			ID: "test1",
			User: &models.User{
				Username: "test1",
			},
		}

		result, err := config.GetCompositeRole(identity, "parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Expected: parent permissions override child permissions in conflicts
		// - "read" should be denied (parent deny overrides child allow)
		// - "write" should be allowed (parent allow overrides child deny)
		// - "list" should be allowed (no conflict, inherited from child)
		// - "delete" should be denied (no conflict, inherited from child)
		expectedAllowPerms := []string{"write", "list"}
		expectedDenyPerms := []string{"read", "delete"}

		assert.ElementsMatch(t, expectedAllowPerms, result.Permissions.Allow)
		assert.ElementsMatch(t, expectedDenyPerms, result.Permissions.Deny)

		// Expected: parent resources override child resources in conflicts
		// - "bucket1" should be denied (parent deny overrides child allow)
		// - "bucket2" should be allowed (parent allow overrides child deny)
		// - "bucket3" should be denied (no conflict, inherited from child)
		expectedAllowResources := []string{"bucket2"}
		expectedDenyResources := []string{"bucket1", "bucket3"}

		assert.ElementsMatch(t, expectedAllowResources, result.Resources.Allow)
		assert.ElementsMatch(t, expectedDenyResources, result.Resources.Deny)
	})

	t.Run("multi-level inheritance with conflicts", func(t *testing.T) {
		roles := map[string]models.Role{
			"grandchild": {
				Name: "Grandchild Role",
				Permissions: models.Permissions{
					Allow: []string{"read", "list"},
					Deny:  []string{"write"},
				},
				Enabled: true,
			},
			"child": {
				Name:     "Child Role",
				Inherits: []string{"grandchild"},
				Permissions: models.Permissions{
					Allow: []string{"write"}, // Overrides grandchild's deny
					Deny:  []string{"list"},  // Overrides grandchild's allow
				},
				Enabled: true,
			},
			"parent": {
				Name:     "Parent Role",
				Inherits: []string{"child"},
				Permissions: models.Permissions{
					Allow: []string{"delete", "list"}, // "list" overrides child's deny
					Deny:  []string{"read"},           // Overrides grandchild's allow (inherited through child)
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
			ID: "test1",
			User: &models.User{
				Username: "test1",
			},
		}

		result, err := config.GetCompositeRole(identity, "parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Expected final state after all inheritance and conflict resolution:
		// - "read": denied by parent (overrides grandchild's allow)
		// - "list": allowed by parent (overrides child's deny)
		// - "write": allowed by child (overrides grandchild's deny)
		// - "delete": allowed by parent (new permission)
		expectedAllowPerms := []string{"list", "write", "delete"}
		expectedDenyPerms := []string{"read"}

		assert.ElementsMatch(t, expectedAllowPerms, result.Permissions.Allow)
		assert.ElementsMatch(t, expectedDenyPerms, result.Permissions.Deny)
	})

	t.Run("no conflicts - simple merge", func(t *testing.T) {
		roles := map[string]models.Role{
			"child": {
				Name: "Child Role",
				Permissions: models.Permissions{
					Allow: []string{"read"},
					Deny:  []string{"delete"},
				},
				Enabled: true,
			},
			"parent": {
				Name:     "Parent Role",
				Inherits: []string{"child"},
				Permissions: models.Permissions{
					Allow: []string{"write"},
					Deny:  []string{"admin"},
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
			ID: "test1",
			User: &models.User{
				Username: "test1",
			},
		}

		result, err := config.GetCompositeRole(identity, "parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// No conflicts, should be simple union
		expectedAllowPerms := []string{"read", "write"}
		expectedDenyPerms := []string{"delete", "admin"}

		assert.ElementsMatch(t, expectedAllowPerms, result.Permissions.Allow)
		assert.ElementsMatch(t, expectedDenyPerms, result.Permissions.Deny)
	})

	t.Run("parent deny overrides child allow", func(t *testing.T) {
		roles := map[string]models.Role{
			"permissive-child": {
				Name: "Permissive Child",
				Permissions: models.Permissions{
					Allow: []string{"read", "write", "delete"},
				},
				Enabled: true,
			},
			"restrictive-parent": {
				Name:     "Restrictive Parent",
				Inherits: []string{"permissive-child"},
				Permissions: models.Permissions{
					Deny: []string{"delete", "write"}, // Parent denies what child allows
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
			ID: "test1",
			User: &models.User{
				Username: "test1",
			},
		}

		result, err := config.GetCompositeRole(identity, "restrictive-parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Parent deny should override child allow
		expectedAllowPerms := []string{"read"}           // Only read survives
		expectedDenyPerms := []string{"delete", "write"} // Parent denies override child allows

		assert.ElementsMatch(t, expectedAllowPerms, result.Permissions.Allow)
		assert.ElementsMatch(t, expectedDenyPerms, result.Permissions.Deny)
	})
}
