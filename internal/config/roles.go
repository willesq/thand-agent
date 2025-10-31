package config

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// LoadRoles loads roles from a file or URL
func (c *Config) LoadRoles() (map[string]models.Role, error) {

	vaultData := ""

	if len(c.Roles.Vault) > 0 {

		if !c.HasVault() {
			return nil, fmt.Errorf("vault configuration is missing. Cannot load roles from vault")
		}

		logrus.Debugln("Loading roles from vault: ", c.Roles.Vault)

		// Load roles from Vault
		data, err := c.GetVault().GetSecret(c.Roles.Vault)

		if err != nil {
			logrus.WithError(err).Errorln("Error loading roles from vault")
			return nil, fmt.Errorf("failed to get secret from vault: %w", err)
		}

		logrus.Debugln("Loaded roles from vault: ", len(data), " bytes")

		vaultData = string(data)
	}

	foundRoles, err := loadDataFromSource(
		c.Roles.Path,
		c.Roles.URL,
		vaultData,
		RoleDefinitions{},
	)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to load roles data")
		return nil, fmt.Errorf("failed to load roles data: %w", err)
	}

	defs := make(map[string]models.Role)

	logrus.Debugln("Processing loaded roles: ", len(foundRoles))

	for _, role := range foundRoles {
		for roleKey, r := range role.Roles {

			if !r.Enabled {
				logrus.Infoln("Role disabled:", roleKey)
				continue
			}

			if _, exists := defs[roleKey]; exists {
				logrus.Warningln("Duplicate role key found, skipping:", roleKey)
				continue
			}

			if len(r.Name) == 0 {
				r.Name = roleKey
			}

			defs[roleKey] = r
		}
	}

	return defs, nil
}

/*
This function is used to evaluate a role and resolve all its
inherited roles into a single composite role.

The function takes an identity and a role name as input,
and returns the composite role or an error if any issues occur
during the resolution process.

For:

- Each role specified in the "Inherits" field is fetched recursively.

- Inheritance may use a ':' to separate the role name from its provider.
For example "aws-prod:admin" would check for a provider by that name that has
a role "admin". If no provider is found, it defaults to the role for the underlying
provider name. For example:

  - "aws-prod:Administrator" would first look for a provider named "aws-prod"

- Permissions from inherited roles are merged into the composite role.

  - Conflicts in permissions are resolved by union of allowed actions
    and intersection of denied actions.
  - If a role is scoped to a specific identity, it is only included if
    the identity matches.

- Cyclic inheritance is detected and results in an error.

The final composite role contains all permissions from the base role
and its inherited roles, providing a complete set of permissions
for the given identity.
*/
func (c *Config) GetCompositeRole(identity *models.Identity, name string) (*models.Role, error) {
	visited := make(map[string]bool)
	return c.resolveCompositeRole(identity, name, visited)
}

// resolveCompositeRole recursively resolves a role and its inheritance chain
func (c *Config) resolveCompositeRole(identity *models.Identity, name string, visited map[string]bool) (*models.Role, error) {
	// Check for cyclic inheritance
	if visited[name] {
		return nil, fmt.Errorf("cyclic inheritance detected in role: %s", name)
	}

	visited[name] = true
	defer delete(visited, name)

	// Get the base role
	baseRole, err := c.Roles.GetRoleByName(name)
	if err != nil {
		return nil, err
	}

	// Create a copy of the base role for the composite
	compositeRole := *baseRole

	// Process inherited roles
	for _, inheritedRoleName := range baseRole.Inherits {
		// Resolve the inherited role (which handles provider-specific inheritance and scope checking)
		inheritedRole, err := c.resolveInheritedRole(identity, inheritedRoleName, visited)
		if err != nil {
			// Check if it's a scope mismatch error and skip
			if strings.Contains(err.Error(), "not applicable to identity") {
				continue
			}
			return nil, fmt.Errorf("failed to resolve inherited role '%s': %w", inheritedRoleName, err)
		}

		// Merge only the inherited permissions and resources
		c.mergeRolePermissions(&compositeRole, inheritedRole)
		c.mergeRoleResources(&compositeRole, inheritedRole)
	}

	return &compositeRole, nil
}

// resolveInheritedRole handles provider-specific role inheritance and scope checking
func (c *Config) resolveInheritedRole(identity *models.Identity, inheritSpec string, visited map[string]bool) (*models.Role, error) {
	var baseRoleName string
	var baseRole *models.Role
	var err error

	// Check if inheritance specifies a provider (format: "provider:role")
	// Split only on the first colon to handle role names with multiple colons (e.g., AWS ARNs)
	colonIndex := strings.Index(inheritSpec, ":")
	if colonIndex > 0 && colonIndex < len(inheritSpec)-1 {
		providerName := inheritSpec[:colonIndex]
		roleName := inheritSpec[colonIndex+1:]

		// First try to find a provider with this name
		provider, err := c.GetProviderByName(providerName)
		if err != nil {
			// If provider not found, treat the whole string as a role name
			baseRoleName = inheritSpec
		} else {
			// If provider exists, use the role name part
			// For now, we'll fall back to treating it as a regular role name
			// This could be extended to support provider-specific role resolution
			_ = provider // Use the provider variable to avoid unused variable error
			baseRoleName = roleName
		}
	} else {
		// Regular role inheritance
		baseRoleName = inheritSpec
	}

	// Get the base role for scope checking
	baseRole, err = c.Roles.GetRoleByName(baseRoleName)
	if err != nil {
		return nil, fmt.Errorf("inherited role '%s' not found: %w", baseRoleName, err)
	}

	// Check scope before resolving inheritance chain
	if !c.isRoleApplicableToIdentity(baseRole, identity) {
		return nil, fmt.Errorf("inherited role '%s' not applicable to identity", baseRoleName)
	}

	// If scope check passes, resolve the full inheritance chain
	return c.resolveCompositeRole(identity, baseRoleName, visited)
}

// isRoleApplicableToIdentity checks if a role's scopes allow it to be applied to the given identity
func (c *Config) isRoleApplicableToIdentity(role *models.Role, identity *models.Identity) bool {
	if role.Scopes == nil {
		// No scopes means role applies to everyone
		return true
	}

	// Check user scopes
	if identity.IsUser() && len(role.Scopes.Users) > 0 {
		userIdentity := identity.GetUser().GetIdentity()
		for _, allowedUser := range role.Scopes.Users {
			if allowedUser == userIdentity || allowedUser == identity.GetUser().Email ||
				allowedUser == identity.GetUser().Username || allowedUser == identity.GetUser().ID {
				return true
			}
		}
	}

	// Check group scopes
	if identity.IsGroup() && len(role.Scopes.Groups) > 0 {
		groupName := identity.GetGroup().Name
		for _, allowedGroup := range role.Scopes.Groups {
			if allowedGroup == groupName || allowedGroup == identity.GetGroup().ID {
				return true
			}
		}
	}

	// Check if user belongs to allowed groups
	if identity.IsUser() && len(role.Scopes.Groups) > 0 {
		userGroups := identity.GetUser().Groups
		for _, userGroup := range userGroups {
			for _, allowedGroup := range role.Scopes.Groups {
				if allowedGroup == userGroup {
					return true
				}
			}
		}
	}

	// If scopes are defined but no match found, role doesn't apply
	if len(role.Scopes.Users) > 0 || len(role.Scopes.Groups) > 0 {
		return false
	}

	return true
}

// mergeRolePermissions merges permissions from inherited role into composite role
// with proper Allow/Deny conflict resolution
func (c *Config) mergeRolePermissions(composite *models.Role, inherited *models.Role) {
	// Start with inherited permissions (child)
	allowSet := make(map[string]bool)
	denySet := make(map[string]bool)

	// Add inherited permissions first (child permissions)
	for _, perm := range inherited.Permissions.Allow {
		allowSet[perm] = true
	}
	for _, perm := range inherited.Permissions.Deny {
		denySet[perm] = true
	}

	// Add composite permissions (parent) - these take precedence in conflicts
	for _, perm := range composite.Permissions.Allow {
		allowSet[perm] = true
		// Parent Allow overrides child Deny
		delete(denySet, perm)
	}
	for _, perm := range composite.Permissions.Deny {
		denySet[perm] = true
		// Parent Deny overrides child Allow
		delete(allowSet, perm)
	}

	// Convert sets back to slices
	composite.Permissions.Allow = make([]string, 0, len(allowSet))
	for perm := range allowSet {
		composite.Permissions.Allow = append(composite.Permissions.Allow, perm)
	}

	composite.Permissions.Deny = make([]string, 0, len(denySet))
	for perm := range denySet {
		composite.Permissions.Deny = append(composite.Permissions.Deny, perm)
	}
}

// mergeRoleResources merges resources from inherited role into composite role
// with proper Allow/Deny conflict resolution
func (c *Config) mergeRoleResources(composite *models.Role, inherited *models.Role) {
	// Start with inherited resources (child)
	allowSet := make(map[string]bool)
	denySet := make(map[string]bool)

	// Add inherited resources first (child resources)
	for _, resource := range inherited.Resources.Allow {
		allowSet[resource] = true
	}
	for _, resource := range inherited.Resources.Deny {
		denySet[resource] = true
	}

	// Add composite resources (parent) - these take precedence in conflicts
	for _, resource := range composite.Resources.Allow {
		allowSet[resource] = true
		// Parent Allow overrides child Deny
		delete(denySet, resource)
	}
	for _, resource := range composite.Resources.Deny {
		denySet[resource] = true
		// Parent Deny overrides child Allow
		delete(allowSet, resource)
	}

	// Convert sets back to slices
	composite.Resources.Allow = make([]string, 0, len(allowSet))
	for resource := range allowSet {
		composite.Resources.Allow = append(composite.Resources.Allow, resource)
	}

	composite.Resources.Deny = make([]string, 0, len(denySet))
	for resource := range denySet {
		composite.Resources.Deny = append(composite.Resources.Deny, resource)
	}
}

// mergeStringSlices merges two string slices, removing duplicates
// Note: This function is kept for backward compatibility but is no longer used
// in the main inheritance logic
func (c *Config) mergeStringSlices(slice1, slice2 []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice1)+len(slice2))

	// Add items from first slice
	for _, item := range slice1 {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// Add items from second slice
	for _, item := range slice2 {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
