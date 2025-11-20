package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// TestGroupsInheritance tests the inheritance of groups' allow and deny rules
// Groups define which groups a user should be added to when assuming a role
func TestGroupsInheritance(t *testing.T) {
	t.Run("simple role without inheritance", func(t *testing.T) {
		roles := map[string]models.Role{
			"basic": {
				Name:        "basic",
				Description: "Basic role with groups",
				Groups: models.Groups{
					Allow: []string{"developers", "users"},
					Deny:  []string{"admins"},
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
			ID: "user1",
			User: &models.User{
				Username: "testuser",
				Email:    "test@example.com",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "basic")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have the same groups as defined
		assert.ElementsMatch(t, []string{"developers", "users"}, result.Groups.Allow)
		assert.ElementsMatch(t, []string{"admins"}, result.Groups.Deny)
	})

	t.Run("role with single inheritance - no conflicts", func(t *testing.T) {
		roles := map[string]models.Role{
			"base": {
				Name:        "base",
				Description: "Base role",
				Groups: models.Groups{
					Allow: []string{"users", "readers"},
					Deny:  []string{"external"},
				},
				Enabled: true,
			},
			"extended": {
				Name:        "extended",
				Description: "Extended role",
				Inherits:    []string{"base"},
				Groups: models.Groups{
					Allow: []string{"developers", "contributors"},
					Deny:  []string{"guests"},
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
			ID: "user1",
			User: &models.User{
				Username: "testuser",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "extended")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should merge groups from both roles - no conflicts
		expectedAllow := []string{"developers", "contributors", "users", "readers"}
		expectedDeny := []string{"guests", "external"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})

	t.Run("parent allow overrides child deny", func(t *testing.T) {
		roles := map[string]models.Role{
			"child": {
				Name: "Child Role",
				Groups: models.Groups{
					Allow: []string{"users", "readers"},
					Deny:  []string{"developers", "admins"},
				},
				Enabled: true,
			},
			"parent": {
				Name:     "Parent Role",
				Inherits: []string{"child"},
				Groups: models.Groups{
					Allow: []string{"developers"}, // Parent allows what child denies
					Deny:  []string{"users"},      // Parent denies what child allows
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

		result, err := config.GetCompositeRoleByName(identity, "parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Parent permissions override child permissions in conflicts
		// - "developers" should be allowed (parent allow overrides child deny)
		// - "users" should be denied (parent deny overrides child allow)
		// - "readers" should be allowed (no conflict, inherited from child)
		// - "admins" should be denied (no conflict, inherited from child)
		expectedAllow := []string{"developers", "readers"}
		expectedDeny := []string{"users", "admins"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})

	t.Run("multi-level inheritance with conflicts", func(t *testing.T) {
		roles := map[string]models.Role{
			"grandchild": {
				Name: "Grandchild Role",
				Groups: models.Groups{
					Allow: []string{"users", "readers"},
					Deny:  []string{"developers"},
				},
				Enabled: true,
			},
			"child": {
				Name:     "Child Role",
				Inherits: []string{"grandchild"},
				Groups: models.Groups{
					Allow: []string{"developers", "contributors"}, // Overrides grandchild's deny
					Deny:  []string{"readers"},                    // Overrides grandchild's allow
				},
				Enabled: true,
			},
			"parent": {
				Name:     "Parent Role",
				Inherits: []string{"child"},
				Groups: models.Groups{
					Allow: []string{"admins", "readers"}, // "readers" overrides child's deny
					Deny:  []string{"users"},             // Overrides grandchild's allow (inherited through child)
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

		result, err := config.GetCompositeRoleByName(identity, "parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Expected final state after all inheritance and conflict resolution:
		// - "users": denied by parent (overrides grandchild's allow)
		// - "readers": allowed by parent (overrides child's deny)
		// - "developers": allowed by child (overrides grandchild's deny)
		// - "contributors": allowed by child (new group)
		// - "admins": allowed by parent (new group)
		expectedAllow := []string{"readers", "developers", "contributors", "admins"}
		expectedDeny := []string{"users"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})

	t.Run("multiple inheritance with conflicts", func(t *testing.T) {
		roles := map[string]models.Role{
			"read-role": {
				Name:        "read-role",
				Description: "Read role",
				Groups: models.Groups{
					Allow: []string{"readers", "users"},
					Deny:  []string{"editors"},
				},
				Enabled: true,
			},
			"write-role": {
				Name:        "write-role",
				Description: "Write role",
				Groups: models.Groups{
					Allow: []string{"editors", "contributors"},
					Deny:  []string{"readers"},
				},
				Enabled: true,
			},
			"admin": {
				Name:        "admin",
				Description: "Admin role",
				Inherits:    []string{"read-role", "write-role"},
				Groups: models.Groups{
					Allow: []string{"admins"},
					Deny:  []string{"users"},
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
			ID: "admin1",
			User: &models.User{
				Username: "admin",
			},
		}

		result, err := config.GetCompositeRoleByName(identity, "admin")
		require.NoError(t, err)
		require.NotNil(t, result)

		// With multiple inheritance, the base role (admin) is the parent
		// and inherited roles are children processed in order:
		// 1. Start with admin: Allow["admins"], Deny["users"]
		// 2. Merge read-role: Allow["readers", "users"], Deny["editors"]
		//    - admin's Deny["users"] overrides read-role's Allow["users"]
		//    - Result: Allow["admins", "readers"], Deny["users", "editors"]
		// 3. Merge write-role: Allow["editors", "contributors"], Deny["readers"]
		//    - Current composite Allow["readers"] overrides write-role's Deny["readers"]
		//    - Current composite Deny["editors"] overrides write-role's Allow["editors"]
		//    - Result: Allow["admins", "readers", "contributors"], Deny["users", "editors"]
		expectedAllow := []string{"admins", "readers", "contributors"}
		expectedDeny := []string{"users", "editors"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})

	t.Run("parent deny overrides child allow", func(t *testing.T) {
		roles := map[string]models.Role{
			"permissive-child": {
				Name: "Permissive Child",
				Groups: models.Groups{
					Allow: []string{"developers", "admins", "editors", "users"},
				},
				Enabled: true,
			},
			"restrictive-parent": {
				Name:     "Restrictive Parent",
				Inherits: []string{"permissive-child"},
				Groups: models.Groups{
					Deny: []string{"admins", "editors"}, // Parent denies what child allows
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

		result, err := config.GetCompositeRoleByName(identity, "restrictive-parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Parent deny overrides child allow
		expectedAllow := []string{"developers", "users"}
		expectedDeny := []string{"admins", "editors"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})

	t.Run("complex inheritance with groups and permissions", func(t *testing.T) {
		roles := map[string]models.Role{
			"viewer": {
				Name: "Viewer",
				Groups: models.Groups{
					Allow: []string{"viewers", "users"},
				},
				Permissions: models.Permissions{
					Allow: []string{"read"},
				},
				Enabled: true,
			},
			"editor": {
				Name:     "Editor",
				Inherits: []string{"viewer"},
				Groups: models.Groups{
					Allow: []string{"editors"},
					Deny:  []string{"viewers"}, // Editors don't need viewer group
				},
				Permissions: models.Permissions{
					Allow: []string{"write"},
				},
				Enabled: true,
			},
			"admin": {
				Name:     "Admin",
				Inherits: []string{"editor"},
				Groups: models.Groups{
					Allow: []string{"admins"},
					Deny:  []string{"users"}, // Admins don't need users group
				},
				Permissions: models.Permissions{
					Allow: []string{"admin"},
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

		// Groups should be properly merged with parent overrides
		expectedAllowGroups := []string{"editors", "admins"}
		expectedDenyGroups := []string{"viewers", "users"}

		assert.ElementsMatch(t, expectedAllowGroups, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDenyGroups, result.Groups.Deny)

		// Permissions should also be properly merged
		expectedAllowPerms := []string{"read", "write", "admin"}
		assert.ElementsMatch(t, expectedAllowPerms, result.Permissions.Allow)
	})

	t.Run("no conflicts - simple merge", func(t *testing.T) {
		roles := map[string]models.Role{
			"child": {
				Name: "Child Role",
				Groups: models.Groups{
					Allow: []string{"users"},
					Deny:  []string{"external"},
				},
				Enabled: true,
			},
			"parent": {
				Name:     "Parent Role",
				Inherits: []string{"child"},
				Groups: models.Groups{
					Allow: []string{"developers"},
					Deny:  []string{"guests"},
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

		result, err := config.GetCompositeRoleByName(identity, "parent")
		require.NoError(t, err)
		require.NotNil(t, result)

		// No conflicts, should be simple union
		expectedAllow := []string{"users", "developers"}
		expectedDeny := []string{"external", "guests"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})

	t.Run("empty groups inheritance", func(t *testing.T) {
		roles := map[string]models.Role{
			"base": {
				Name: "Base Role",
				Groups: models.Groups{
					Allow: []string{"users", "developers"},
				},
				Enabled: true,
			},
			"derived": {
				Name:     "Derived Role",
				Inherits: []string{"base"},
				// No groups defined, should inherit from base
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

		result, err := config.GetCompositeRoleByName(identity, "derived")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should inherit groups from base
		expectedAllow := []string{"users", "developers"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.Empty(t, result.Groups.Deny)
	})

	t.Run("duplicate groups in allow and deny lists", func(t *testing.T) {
		roles := map[string]models.Role{
			"base": {
				Name: "Base Role",
				Groups: models.Groups{
					Allow: []string{"users", "developers"},
					Deny:  []string{"admins"},
				},
				Enabled: true,
			},
			"override": {
				Name:     "Override Role",
				Inherits: []string{"base"},
				Groups: models.Groups{
					Allow: []string{"developers", "admins"}, // developers is duplicate, admins conflicts
					Deny:  []string{"users"},                // users conflicts
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

		result, err := config.GetCompositeRoleByName(identity, "override")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Parent overrides child in conflicts, duplicates removed
		// - "developers": allowed (no conflict, appears in both)
		// - "admins": allowed (parent allow overrides child deny)
		// - "users": denied (parent deny overrides child allow)
		expectedAllow := []string{"developers", "admins"}
		expectedDeny := []string{"users"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})

	t.Run("deep inheritance chain", func(t *testing.T) {
		roles := map[string]models.Role{
			"level1": {
				Name: "Level 1",
				Groups: models.Groups{
					Allow: []string{"group1", "group2"},
					Deny:  []string{"group3"},
				},
				Enabled: true,
			},
			"level2": {
				Name:     "Level 2",
				Inherits: []string{"level1"},
				Groups: models.Groups{
					Allow: []string{"group3", "group4"}, // group3 conflicts with level1 deny
					Deny:  []string{"group2"},           // group2 conflicts with level1 allow
				},
				Enabled: true,
			},
			"level3": {
				Name:     "Level 3",
				Inherits: []string{"level2"},
				Groups: models.Groups{
					Allow: []string{"group5"},
					Deny:  []string{"group4"}, // group4 conflicts with level2 allow
				},
				Enabled: true,
			},
			"level4": {
				Name:     "Level 4",
				Inherits: []string{"level3"},
				Groups: models.Groups{
					Allow: []string{"group1", "group2"}, // group2 conflicts with level2 deny
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

		result, err := config.GetCompositeRoleByName(identity, "level4")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Deep chain resolution with parent overrides:
		// - group1: allowed (level1, level4 - no conflicts)
		// - group2: allowed (level4 overrides level2's deny)
		// - group3: allowed (level2 overrides level1's deny)
		// - group4: denied (level3 overrides level2's allow)
		// - group5: allowed (level3 - no conflicts)
		expectedAllow := []string{"group1", "group2", "group3", "group5"}
		expectedDeny := []string{"group4"}

		assert.ElementsMatch(t, expectedAllow, result.Groups.Allow)
		assert.ElementsMatch(t, expectedDeny, result.Groups.Deny)
	})
}
