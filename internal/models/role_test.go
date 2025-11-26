package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		user     *User
		expected bool
	}{
		{
			name: "nil user returns false",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
			},
			user:     nil,
			expected: false,
		},
		{
			name: "no scopes defined allows access",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes:      nil,
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "empty scopes allows access",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes:      &RoleScopes{},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "user in allowed users by username",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users: []string{"testuser", "otheruser"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "user in allowed users by ID",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users: []string{"user1", "user2"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "user not in allowed users",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users: []string{"otheruser", "anotheruser"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: false,
		},
		{
			name: "user in allowed group",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Groups: []string{"admins", "developers"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   []string{"developers", "testers"},
			},
			expected: true,
		},
		{
			name: "user not in allowed groups",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Groups: []string{"admins", "superusers"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   []string{"developers", "testers"},
			},
			expected: false,
		},
		{
			name: "user in allowed domain",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Domains: []string{"example.com", "company.org"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "user not in allowed domains but no users or groups scopes",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Domains: []string{"company.org", "internal.net"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			// Domain scopes properly deny access when user's domain doesn't match
			expected: false,
		},
		{
			name: "user matches via group when users scope is also defined",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users:  []string{"otheruser"},
					Groups: []string{"developers"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   []string{"developers"},
			},
			expected: true,
		},
		{
			name: "user matches via domain when users and groups are defined",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users:   []string{"otheruser"},
					Groups:  []string{"admins"},
					Domains: []string{"example.com"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   []string{"developers"},
			},
			expected: true,
		},
		{
			name: "user with no groups and group scope defined",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Groups: []string{"admins"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   nil,
			},
			expected: false,
		},
		{
			name: "user with empty groups and group scope defined",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Groups: []string{"admins"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   []string{},
			},
			expected: false,
		},
		{
			name: "user with no email and domain scope defined",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Domains: []string{"example.com"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "",
			},
			// Domain scopes deny access when user has no email (empty domain won't match)
			expected: false,
		},
		{
			name: "multiple users in scope - first match",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users: []string{"testuser", "user2", "user3"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "multiple users in scope - last match",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users: []string{"user1", "user2", "testuser"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "differentuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "user in multiple groups - one matches",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Groups: []string{"admins"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   []string{"developers", "admins", "testers"},
			},
			expected: true,
		},
		{
			name: "only domains scope defined - user not matching is denied",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Domains: []string{"company.org"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			// Domain scopes properly deny access when user's domain doesn't match
			expected: false,
		},
		{
			name: "domains with users scope - user domain not matching and user not in list",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users:   []string{"otheruser"},
					Domains: []string{"company.org"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: false,
		},
		// Case-insensitive matching tests
		{
			name: "user matches by username case-insensitive",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users: []string{"TestUser"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "user matches by email case-insensitive",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Users: []string{"Test@Example.COM"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
		{
			name: "user matches group case-insensitive",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Groups: []string{"ADMINS", "Developers"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
				Groups:   []string{"admins"},
			},
			expected: true,
		},
		{
			name: "user matches domain case-insensitive",
			role: Role{
				Name:        "admin",
				Description: "Admin role",
				Scopes: &RoleScopes{
					Domains: []string{"Example.COM"},
				},
			},
			user: &User{
				ID:       "user1",
				Username: "testuser",
				Email:    "test@example.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.HasPermission(tt.user)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{
			name: "valid role with name and description",
			role: Role{
				Name:        "admin",
				Description: "Administrator role",
			},
			expected: true,
		},
		{
			name: "invalid role - empty name",
			role: Role{
				Name:        "",
				Description: "Administrator role",
			},
			expected: false,
		},
		{
			name: "invalid role - empty description",
			role: Role{
				Name:        "admin",
				Description: "",
			},
			expected: false,
		},
		{
			name: "invalid role - both empty",
			role: Role{
				Name:        "",
				Description: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRole_GetName(t *testing.T) {
	role := Role{Name: "admin"}
	assert.Equal(t, "admin", role.GetName())
}

func TestRole_GetSnakeCaseName(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		expected string
	}{
		{
			name:     "simple name",
			roleName: "admin",
			expected: "admin",
		},
		{
			name:     "camelCase name",
			roleName: "superAdmin",
			expected: "superadmin",
		},
		{
			name:     "name with spaces",
			roleName: "Super Admin",
			expected: "super_admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := Role{Name: tt.roleName}
			result := role.GetSnakeCaseName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRole_GetDescription(t *testing.T) {
	role := Role{Description: "Test description"}
	assert.Equal(t, "Test description", role.GetDescription())
}

func TestRole_AsMap(t *testing.T) {
	role := Role{
		Name:        "admin",
		Description: "Administrator role",
		Enabled:     true,
		Providers:   []string{"aws", "gcp"},
	}

	result := role.AsMap()
	assert.NotNil(t, result)
	assert.Equal(t, "admin", result["name"])
	assert.Equal(t, "Administrator role", result["description"])
	assert.Equal(t, true, result["enabled"])
}
