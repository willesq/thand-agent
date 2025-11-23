package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// Test GetCompositeRole functionality
func TestGetCompositeRole(t *testing.T) {
	tests := []struct {
		name          string
		roles         map[string]models.Role
		providers     map[string]models.Provider
		identity      *models.Identity
		roleName      string
		expected      *models.Role
		expectError   bool
		errorContains string
	}{
		{
			name: "simple role without inheritance",
			roles: map[string]models.Role{
				"basic": {
					Name:        "basic",
					Description: "Basic role",
					Permissions: models.Permissions{
						Allow: []string{"read"},
						Deny:  []string{"delete"},
					},
					Enabled: true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Email:    "test@example.com",
				},
			},
			roleName: "basic",
			expected: &models.Role{
				Name:        "basic",
				Description: "Basic role",
				Permissions: models.Permissions{
					Allow: []string{"read"},
					Deny:  []string{"delete"},
				},
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "role with single inheritance",
			roles: map[string]models.Role{
				"base": {
					Name:        "base",
					Description: "Base role",
					Permissions: models.Permissions{
						Allow: []string{"read"},
					},
					Resources: models.Resources{
						Allow: []string{"resource1"},
					},
					Enabled: true,
				},
				"extended": {
					Name:        "extended",
					Description: "Extended role",
					Inherits:    []string{"base"},
					Permissions: models.Permissions{
						Allow: []string{"write"},
						Deny:  []string{"admin"},
					},
					Resources: models.Resources{
						Allow: []string{"resource2"},
					},
					Enabled: true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
				},
			},
			roleName: "extended",
			expected: &models.Role{
				Name:        "extended",
				Description: "Extended role",
				Inherits:    []string{"base"},
				Permissions: models.Permissions{
					Allow: []string{"write", "read"},
					Deny:  []string{"admin"},
				},
				Resources: models.Resources{
					Allow: []string{"resource2", "resource1"},
				},
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "role with multiple inheritance",
			roles: map[string]models.Role{
				"read-role": {
					Name:        "read-role",
					Description: "Read role",
					Permissions: models.Permissions{
						Allow: []string{"read"},
					},
					Workflows: []string{"read-workflow"},
					Enabled:   true,
				},
				"write-role": {
					Name:        "write-role",
					Description: "Write role",
					Permissions: models.Permissions{
						Allow: []string{"write"},
					},
					Workflows: []string{"write-workflow"},
					Enabled:   true,
				},
				"admin": {
					Name:        "admin",
					Description: "Admin role",
					Inherits:    []string{"read-role", "write-role"},
					Permissions: models.Permissions{
						Allow: []string{"admin"},
					},
					Workflows: []string{"admin-workflow"},
					Enabled:   true,
				},
			},
			identity: &models.Identity{
				ID: "admin1",
				User: &models.User{
					Username: "admin",
				},
			},
			roleName: "admin",
			expected: &models.Role{
				Name:        "admin",
				Description: "Admin role",
				Inherits:    []string{"read-role", "write-role"},
				Permissions: models.Permissions{
					Allow: []string{"admin", "read", "write"},
				},
				Workflows: []string{"admin-workflow"}, // Only the role's own workflows, not inherited
				Enabled:   true,
			},
			expectError: false,
		},
		{
			name: "role with scopes - user allowed",
			roles: map[string]models.Role{
				"scoped": {
					Name:        "scoped",
					Description: "Scoped role",
					Permissions: models.Permissions{
						Allow: []string{"special"},
					},
					Scopes: &models.RoleScopes{
						Users: []string{"test@example.com"},
					},
					Enabled: true,
				},
				"public": {
					Name:        "public",
					Description: "Public role",
					Inherits:    []string{"scoped"},
					Permissions: models.Permissions{
						Allow: []string{"read"},
					},
					Enabled: true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Email:    "test@example.com",
				},
			},
			roleName: "public",
			expected: &models.Role{
				Name:        "public",
				Description: "Public role",
				Inherits:    []string{"scoped"},
				Permissions: models.Permissions{
					Allow: []string{"read", "special"},
				},
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "role with scopes - user not allowed",
			roles: map[string]models.Role{
				"scoped": {
					Name:        "scoped",
					Description: "Scoped role",
					Permissions: models.Permissions{
						Allow: []string{"special"},
					},
					Scopes: &models.RoleScopes{
						Users: []string{"other@example.com"},
					},
					Enabled: true,
				},
				"public": {
					Name:        "public",
					Description: "Public role",
					Inherits:    []string{"scoped"},
					Permissions: models.Permissions{
						Allow: []string{"read"},
					},
					Enabled: true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Email:    "test@example.com",
				},
			},
			roleName: "public",
			expected: &models.Role{
				Name:        "public",
				Description: "Public role",
				Inherits:    []string{"scoped"},
				Permissions: models.Permissions{
					Allow: []string{"read"},
				},
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "cyclic inheritance",
			roles: map[string]models.Role{
				"role1": {
					Name:     "role1",
					Inherits: []string{"role2"},
					Enabled:  true,
				},
				"role2": {
					Name:     "role2",
					Inherits: []string{"role1"},
					Enabled:  true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
				},
			},
			roleName:      "role1",
			expectError:   true,
			errorContains: "cyclic inheritance detected",
		},
		{
			name: "nonexistent inherited role",
			roles: map[string]models.Role{
				"parent": {
					Name:     "parent",
					Inherits: []string{"nonexistent"},
					Enabled:  true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
				},
			},
			roleName:      "parent",
			expectError:   true,
			errorContains: "role not found: nonexistent",
		},
		{
			name: "nonexistent base role",
			roles: map[string]models.Role{
				"other": {
					Name:    "other",
					Enabled: true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
				},
			},
			roleName:      "nonexistent",
			expectError:   true,
			errorContains: "role not found: nonexistent",
		},
		{
			name: "group scope inheritance",
			roles: map[string]models.Role{
				"group-role": {
					Name:        "group-role",
					Description: "Group specific role",
					Permissions: models.Permissions{
						Allow: []string{"group-action"},
					},
					Scopes: &models.RoleScopes{
						Groups: []string{"developers"},
					},
					Enabled: true,
				},
				"user-role": {
					Name:        "user-role",
					Description: "User role inheriting group role",
					Inherits:    []string{"group-role"},
					Permissions: models.Permissions{
						Allow: []string{"user-action"},
					},
					Enabled: true,
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "developer1",
					Groups:   []string{"developers", "users"},
				},
			},
			roleName: "user-role",
			expected: &models.Role{
				Name:        "user-role",
				Description: "User role inheriting group role",
				Inherits:    []string{"group-role"},
				Permissions: models.Permissions{
					Allow: []string{"user-action", "group-action"},
				},
				Enabled: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a config with test data
			config := &Config{
				Roles: RoleConfig{
					Definitions: tt.roles,
				},
			}

			if tt.providers != nil {
				config.Providers = ProviderConfig{
					Definitions: tt.providers,
				}
			}

			// Call GetCompositeRole
			result, err := config.GetCompositeRoleByName(tt.identity, tt.roleName)

			// Check error expectations
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			// Check success case
			require.NoError(t, err)
			require.NotNil(t, result)

			// Compare the results
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Description, result.Description)
			assert.Equal(t, tt.expected.Enabled, result.Enabled)

			// Compare permissions (order doesn't matter)
			assert.ElementsMatch(t, tt.expected.Permissions.Allow, result.Permissions.Allow)
			assert.ElementsMatch(t, tt.expected.Permissions.Deny, result.Permissions.Deny)

			// Compare resources (order doesn't matter)
			assert.ElementsMatch(t, tt.expected.Resources.Allow, result.Resources.Allow)
			assert.ElementsMatch(t, tt.expected.Resources.Deny, result.Resources.Deny)

			// Compare workflows (order doesn't matter)
			assert.ElementsMatch(t, tt.expected.Workflows, result.Workflows)

			// Compare providers (order doesn't matter)
			assert.ElementsMatch(t, tt.expected.Providers, result.Providers)
		})
	}
}

func TestGetCompositeRole_ProviderSpecificInheritance(t *testing.T) {
	roles := map[string]models.Role{
		"admin": {
			Name:        "admin",
			Description: "Admin role in AWS",
			Permissions: models.Permissions{
				Allow: []string{"aws:admin"},
			},
			Enabled: true,
		},
		"base": {
			Name:        "base",
			Description: "Base role",
			Inherits:    []string{"aws-prod:admin"},
			Permissions: models.Permissions{
				Allow: []string{"base:read"},
			},
			Enabled: true,
		},
	}

	providers := map[string]models.Provider{
		"aws-prod": {
			Name:        "aws-prod",
			Description: "AWS Production",
			Provider:    "aws",
		},
	}

	config := &Config{
		Roles: RoleConfig{
			Definitions: roles,
		},
		Providers: ProviderConfig{
			Definitions: providers,
		},
	}

	identity := &models.Identity{
		ID: "user1",
		User: &models.User{
			Username: "testuser",
		},
	}

	result, err := config.GetCompositeRoleByName(identity, "base")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should inherit from the 'admin' role since aws-prod provider exists
	assert.Equal(t, "base", result.Name)
	assert.ElementsMatch(t, []string{"base:read"}, result.Permissions.Allow)
}

func TestMergeStringSlices(t *testing.T) {
	config := &Config{}

	tests := []struct {
		name     string
		slice1   []string
		slice2   []string
		expected []string
	}{
		{
			name:     "empty slices",
			slice1:   []string{},
			slice2:   []string{},
			expected: []string{},
		},
		{
			name:     "one empty slice",
			slice1:   []string{"a", "b"},
			slice2:   []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "no duplicates",
			slice1:   []string{"a", "b"},
			slice2:   []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "with duplicates",
			slice1:   []string{"a", "b", "c"},
			slice2:   []string{"b", "c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "all duplicates",
			slice1:   []string{"a", "b"},
			slice2:   []string{"a", "b"},
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.mergeStringSlices(tt.slice1, tt.slice2)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestIsRoleApplicableToIdentity(t *testing.T) {
	config := &Config{}

	tests := []struct {
		name     string
		role     *models.Role
		identity *models.Identity
		expected bool
	}{
		{
			name: "no scopes - always applicable",
			role: &models.Role{
				Name: "test",
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
				},
			},
			expected: true,
		},
		{
			name: "user scope - email match",
			role: &models.Role{
				Name: "test",
				Scopes: &models.RoleScopes{
					Users: []string{"test@example.com"},
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Email:    "test@example.com",
				},
			},
			expected: true,
		},
		{
			name: "user scope - username match",
			role: &models.Role{
				Name: "test",
				Scopes: &models.RoleScopes{
					Users: []string{"testuser"},
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Email:    "test@example.com",
				},
			},
			expected: true,
		},
		{
			name: "user scope - no match",
			role: &models.Role{
				Name: "test",
				Scopes: &models.RoleScopes{
					Users: []string{"other@example.com"},
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Email:    "test@example.com",
				},
			},
			expected: false,
		},
		{
			name: "group scope - user in group",
			role: &models.Role{
				Name: "test",
				Scopes: &models.RoleScopes{
					Groups: []string{"developers"},
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Groups:   []string{"developers", "users"},
				},
			},
			expected: true,
		},
		{
			name: "group scope - user not in group",
			role: &models.Role{
				Name: "test",
				Scopes: &models.RoleScopes{
					Groups: []string{"admins"},
				},
			},
			identity: &models.Identity{
				ID: "user1",
				User: &models.User{
					Username: "testuser",
					Groups:   []string{"developers", "users"},
				},
			},
			expected: false,
		},
		{
			name: "group identity - group match",
			role: &models.Role{
				Name: "test",
				Scopes: &models.RoleScopes{
					Groups: []string{"developers"},
				},
			},
			identity: &models.Identity{
				ID: "group1",
				Group: &models.Group{
					Name: "developers",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.isRoleApplicableToIdentity(tt.role, tt.identity)
			assert.Equal(t, tt.expected, result)
		})
	}
}
