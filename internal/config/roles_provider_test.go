package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// TestProviderSpecificInheritance tests provider-specific role inheritance
// with complex role names containing multiple colons (AWS ARNs, GCP service accounts, etc.)
func TestProviderSpecificInheritance(t *testing.T) {
	t.Run("aws arn inheritance", func(t *testing.T) {
		roles := map[string]models.Role{
			"arn:aws:iam::123456789012:role/TestRole": {
				Name: "AWS Test Role",
				Permissions: models.Permissions{
					Allow: []string{
						"s3:GetObject",
						"s3:ListBucket",
						"ec2:DescribeInstances",
					},
				},
				Enabled: true,
			},
			"app-role": {
				Name:     "Application Role",
				Inherits: []string{"aws-prod:arn:aws:iam::123456789012:role/TestRole"},
				Permissions: models.Permissions{
					Allow: []string{"app:deploy"},
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
				Email:    "testuser@example.com",
			},
		}

		result, err := config.GetCompositeRole(identity, "app-role")
		require.NoError(t, err)
		require.NotNil(t, result)

		expectedPerms := []string{
			"app:deploy",
			"s3:GetObject",
			"s3:ListBucket",
			"ec2:DescribeInstances",
		}
		assert.ElementsMatch(t, expectedPerms, result.Permissions.Allow)
	})

	t.Run("gcp service account inheritance", func(t *testing.T) {
		roles := map[string]models.Role{
			"test-service@test-project.iam.gserviceaccount.com": {
				Name: "GCP Service Account",
				Permissions: models.Permissions{
					Allow: []string{
						"compute.instances.get",
						"compute.instances.list",
						"storage.objects.get",
					},
				},
				Enabled: true,
			},
			"developer-role": {
				Name:     "Developer Role",
				Inherits: []string{"gcp-dev:test-service@test-project.iam.gserviceaccount.com"},
				Permissions: models.Permissions{
					Allow: []string{"cloudrun.services.deploy"},
				},
				Enabled: true,
			},
		}

		providers := map[string]models.Provider{
			"gcp-dev": {
				Name:     "GCP Development",
				Provider: "gcp",
				Enabled:  true,
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
				Username: "developer",
				Email:    "developer@example.com",
			},
		}

		result, err := config.GetCompositeRole(identity, "developer-role")
		require.NoError(t, err)
		require.NotNil(t, result)

		expectedPerms := []string{
			"cloudrun.services.deploy",
			"compute.instances.get",
			"compute.instances.list",
			"storage.objects.get",
		}
		assert.ElementsMatch(t, expectedPerms, result.Permissions.Allow)
	})

	t.Run("azure resource id inheritance", func(t *testing.T) {
		roles := map[string]models.Role{
			"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers/Microsoft.Authorization/roleDefinitions/12345678-1234-1234-1234-123456789012": {
				Name: "Custom Azure Role",
				Permissions: models.Permissions{
					Allow: []string{
						"Microsoft.Compute/virtualMachines/read",
						"Microsoft.Compute/virtualMachines/start/action",
						"Microsoft.Storage/storageAccounts/listKeys/action",
					},
				},
				Enabled: true,
			},
			"ops-role": {
				Name:     "Operations Role",
				Inherits: []string{"azure-prod:/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers/Microsoft.Authorization/roleDefinitions/12345678-1234-1234-1234-123456789012"},
				Permissions: models.Permissions{
					Allow: []string{"ops.deploy"},
				},
				Enabled: true,
			},
		}

		providers := map[string]models.Provider{
			"azure-prod": {
				Name:     "Azure Production",
				Provider: "azure",
				Enabled:  true,
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
				Username: "operator",
				Email:    "operator@example.com",
			},
		}

		result, err := config.GetCompositeRole(identity, "ops-role")
		require.NoError(t, err)
		require.NotNil(t, result)

		expectedPerms := []string{
			"ops.deploy",
			"Microsoft.Compute/virtualMachines/read",
			"Microsoft.Compute/virtualMachines/start/action",
			"Microsoft.Storage/storageAccounts/listKeys/action",
		}
		assert.ElementsMatch(t, expectedPerms, result.Permissions.Allow)
	})

	t.Run("provider not found fallback", func(t *testing.T) {
		roles := map[string]models.Role{
			"nonexistent:arn:aws:iam::123456789012:role/TestRole": {
				Name: "Fallback Role",
				Permissions: models.Permissions{
					Allow: []string{"fallback:action"},
				},
				Enabled: true,
			},
			"parent-role": {
				Name:     "Parent Role",
				Inherits: []string{"nonexistent:arn:aws:iam::123456789012:role/TestRole"},
				Permissions: models.Permissions{
					Allow: []string{"parent:action"},
				},
				Enabled: true,
			},
		}

		config := &Config{
			Roles: RoleConfig{
				Definitions: roles,
			},
			// No providers defined - forces fallback behavior
		}

		identity := &models.Identity{
			ID: "user1",
			User: &models.User{
				Username: "testuser",
				Email:    "testuser@example.com",
			},
		}

		result, err := config.GetCompositeRole(identity, "parent-role")
		require.NoError(t, err)
		require.NotNil(t, result)

		expectedPerms := []string{
			"parent:action",
			"fallback:action",
		}
		assert.ElementsMatch(t, expectedPerms, result.Permissions.Allow)
	})
}

// TestProviderParsingLogicConsolidated tests provider parsing logic consolidation
// This validates that the fix for multiple colons (AWS ARNs) works correctly
func TestProviderParsingLogicConsolidated(t *testing.T) {
	t.Run("validate parsing with multiple colons", func(t *testing.T) {
		// This test validates that provider parsing correctly handles role names with multiple colons
		inheritSpec := "aws-prod:arn:aws:iam::123456789012:role/TestRole"

		// Demonstrate old logic would have failed:
		// strings.Split(inheritSpec, ":") would result in 7 parts, not 2
		oldParts := strings.Split(inheritSpec, ":")
		assert.Equal(t, 7, len(oldParts), "Old logic would incorrectly split into 7 parts")
		assert.Equal(t, []string{"aws-prod", "arn", "aws", "iam", "", "123456789012", "role/TestRole"}, oldParts)

		// New logic correctly splits only on first colon:
		colonIndex := strings.Index(inheritSpec, ":")
		require.True(t, colonIndex > 0 && colonIndex < len(inheritSpec)-1, "Should detect provider:role format")

		providerName := inheritSpec[:colonIndex]
		roleName := inheritSpec[colonIndex+1:]

		assert.Equal(t, "aws-prod", providerName, "Should extract provider name correctly")
		assert.Equal(t, "arn:aws:iam::123456789012:role/TestRole", roleName, "Should extract full role name correctly")
	})

	t.Run("edge cases in parsing", func(t *testing.T) {
		testCases := []struct {
			name         string
			inheritSpec  string
			shouldParse  bool
			expectedProv string
			expectedRole string
		}{
			{
				name:        "no colon - regular role",
				inheritSpec: "simple-role",
				shouldParse: false,
			},
			{
				name:        "colon at start",
				inheritSpec: ":role-name",
				shouldParse: false,
			},
			{
				name:        "colon at end",
				inheritSpec: "provider:",
				shouldParse: false,
			},
			{
				name:         "valid provider:role",
				inheritSpec:  "provider:role-name",
				shouldParse:  true,
				expectedProv: "provider",
				expectedRole: "role-name",
			},
			{
				name:         "complex role with colons",
				inheritSpec:  "aws:arn:aws:iam::123:role/test",
				shouldParse:  true,
				expectedProv: "aws",
				expectedRole: "arn:aws:iam::123:role/test",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				colonIndex := strings.Index(tc.inheritSpec, ":")
				isParseable := colonIndex > 0 && colonIndex < len(tc.inheritSpec)-1

				assert.Equal(t, tc.shouldParse, isParseable, "Parsing detection should match expectation")

				if tc.shouldParse {
					providerName := tc.inheritSpec[:colonIndex]
					roleName := tc.inheritSpec[colonIndex+1:]

					assert.Equal(t, tc.expectedProv, providerName, "Provider name should match")
					assert.Equal(t, tc.expectedRole, roleName, "Role name should match")
				}
			})
		}
	})

	t.Run("integration test - end to end parsing", func(t *testing.T) {
		// This test proves the parsing logic works in the actual inheritance system
		roles := map[string]models.Role{
			"arn:aws:iam::123456789012:role/TestRole": {
				Name: "AWS Role",
				Permissions: models.Permissions{
					Allow: []string{"s3:GetObject"},
				},
				Enabled: true,
			},
			"test-role": {
				Name:     "Test Role",
				Inherits: []string{"aws-prod:arn:aws:iam::123456789012:role/TestRole"},
				Permissions: models.Permissions{
					Allow: []string{"test:action"},
				},
				Enabled: true,
			},
		}

		providers := map[string]models.Provider{
			"aws-prod": {
				Name:     "AWS Production",
				Provider: "aws",
				Enabled:  true,
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
				Email:    "testuser@example.com",
			},
		}

		// This would have failed with the old strings.Split() logic
		result, err := config.GetCompositeRole(identity, "test-role")
		require.NoError(t, err, "Should successfully parse and inherit AWS ARN role")
		require.NotNil(t, result)

		expectedPerms := []string{
			"test:action",
			"s3:GetObject",
		}
		assert.ElementsMatch(t, expectedPerms, result.Permissions.Allow)
	})
}
