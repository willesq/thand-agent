package config

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/config/environment"
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
		models.RoleDefinitions{},
	)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to load roles data")
		return nil, fmt.Errorf("failed to load roles data: %w", err)
	}

	if len(foundRoles) == 0 {
		logrus.Warningln("No roles found from any source, loading defaults")
		foundRoles, err = environment.GetDefaultRoles(c.Environment.Platform)
		if err != nil {
			return nil, fmt.Errorf("failed to load default roles: %w", err)
		}
		logrus.Infoln("Loaded default roles:", len(foundRoles))
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

func (c *Config) GetRoleByName(name string) (*models.Role, error) {
	return c.Roles.GetRoleByName(name)
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

func (c *Config) GetCompositeRole(identity *models.Identity, baseRole *models.Role) (*models.Role, error) {
	visited := make(map[string]bool)
	return c.resolveCompositeRole(identity, baseRole, visited)
}

func (c *Config) GetCompositeRoleByName(identity *models.Identity, roleName string) (*models.Role, error) {

	// Get the base role
	baseRole, err := c.GetRoleByName(roleName)
	if err != nil {
		return nil, err
	}

	return c.GetCompositeRole(identity, baseRole)
}

func (c *Config) resolveCompositeRoleByName(identity *models.Identity, roleName string, visited map[string]bool) (*models.Role, error) {

	// Get the base role
	baseRole, err := c.GetRoleByName(roleName)
	if err != nil {
		return nil, err
	}

	return c.resolveCompositeRole(identity, baseRole, visited)

}

// resolveCompositeRole recursively resolves a role and its inheritance chain
func (c *Config) resolveCompositeRole(identity *models.Identity, baseRole *models.Role, visited map[string]bool) (*models.Role, error) {
	// Check for cyclic inheritance
	if visited[baseRole.Name] {
		return nil, fmt.Errorf("cyclic inheritance detected in role: %s", baseRole.Name)
	}

	visited[baseRole.Name] = true
	defer delete(visited, baseRole.Name)

	// Create a copy of the base role for the composite
	compositeRole := *baseRole

	// Process inherited roles and track which ones to keep vs remove
	var remainingInherits []string
	for _, inheritedRoleName := range baseRole.Inherits {

		// First check to see if the rolename exists as-is for
		// one of the providers
		providerRole := c.GetProviderRole(inheritedRoleName, baseRole.Providers...)

		if providerRole != nil {
			// Keep provider roles in the inherits list
			remainingInherits = append(remainingInherits, inheritedRoleName)
			continue
		}

		// Try to resolve the inherited role (which handles provider-specific inheritance and scope checking)
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

		// Only provider roles are kept in the inherits list; resolved inherited roles are excluded
	}

	// Update the inherits list to only contain provider roles that weren't resolved
	compositeRole.Inherits = remainingInherits

	// Even for roles without inheritance, we need to resolve Allow/Deny conflicts
	// within the same role and handle condensed actions properly
	c.resolveInternalPermissionConflicts(&compositeRole)

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
		_, err := c.GetProviderByName(providerName)
		if err != nil {
			// If provider not found, treat the whole string as a role name
			baseRoleName = inheritSpec
		} else {
			// If provider exists, use the role name part
			// For now, we'll fall back to treating it as a regular role name
			// This could be extended to support provider-specific role resolution
			baseRoleName = roleName
		}
	} else {
		// Regular role inheritance
		baseRoleName = inheritSpec
	}

	// Get the base role for scope checking
	baseRole, err = c.GetRoleByName(baseRoleName)
	if err != nil {
		return nil, fmt.Errorf("inherited role '%s' not found: %w", baseRoleName, err)
	}

	// Check scope before resolving inheritance chain
	if !c.isRoleApplicableToIdentity(baseRole, identity) {
		return nil, fmt.Errorf("inherited role '%s' not applicable to identity", baseRoleName)
	}

	// If scope check passes, resolve the full inheritance chain
	return c.resolveCompositeRoleByName(identity, baseRoleName, visited)
}

// isRoleApplicableToIdentity checks if a role's scopes allow it to be applied to the given identity
func (c *Config) isRoleApplicableToIdentity(role *models.Role, identity *models.Identity) bool {
	if role.Scopes == nil {
		// No scopes means role applies to everyone
		return true
	}

	if identity == nil {
		// No identity provided, cannot match scoped role
		return false
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
			if slices.Contains(role.Scopes.Groups, userGroup) {
				return true
			}
		}
	}

	// If scopes are defined but no match found, role doesn't apply
	if len(role.Scopes.Users) > 0 || len(role.Scopes.Groups) > 0 {
		return false
	}

	return true
}

// expandPermissionsToSet is a generic helper that expands a slice of permissions
// into a set (map[string]bool) by expanding any condensed actions
func expandPermissionsToSet(permissions []string) map[string]bool {
	permSet := make(map[string]bool)
	for _, perm := range permissions {
		expandedPerms := expandCondensedActions(perm)
		for _, expandedPerm := range expandedPerms {
			permSet[expandedPerm] = true
		}
	}
	return permSet
}

// setToSlice converts a permission set back to a sorted slice
func setToSlice(permSet map[string]bool) []string {
	result := make([]string, 0, len(permSet))
	for perm := range permSet {
		result = append(result, perm)
	}
	sort.Strings(result)
	return result
}

// mergeRolePermissions merges permissions from inherited role into composite role
// with proper Allow/Deny conflict resolution and intelligent condensed action handling
// Parent (composite) permissions override child (inherited) permissions in conflicts
func (c *Config) mergeRolePermissions(composite *models.Role, inherited *models.Role) {
	// Expand all condensed actions to individual permissions for merging
	childAllowSet := expandPermissionsToSet(inherited.Permissions.Allow)
	childDenySet := expandPermissionsToSet(inherited.Permissions.Deny)
	parentAllowSet := expandPermissionsToSet(composite.Permissions.Allow)
	parentDenySet := expandPermissionsToSet(composite.Permissions.Deny)

	// Merge with inheritance override logic:
	// 1. Start with child permissions
	// 2. Parent Allow overrides child Deny (remove from child deny, add to final allow)
	// 3. Parent Deny overrides child Allow (remove from child allow, add to final deny)
	// 4. Add remaining parent permissions that don't conflict

	finalAllowSet := make(map[string]bool)
	finalDenySet := make(map[string]bool)

	// Start with child permissions
	for perm := range childAllowSet {
		finalAllowSet[perm] = true
	}
	for perm := range childDenySet {
		finalDenySet[perm] = true
	}

	// Handle parent overrides: parent Allow overrides child Deny
	for perm := range parentAllowSet {
		// Remove from final deny if child denied it
		delete(finalDenySet, perm)
		// Add to final allow
		finalAllowSet[perm] = true
	}

	// Handle parent overrides: parent Deny overrides child Allow
	for perm := range parentDenySet {
		// Remove from final allow if child allowed it
		delete(finalAllowSet, perm)
		// Add to final deny
		finalDenySet[perm] = true
	}

	// Convert expanded sets back to slices and condense actions back for cleaner output
	composite.Permissions.Allow = condenseActions(setToSlice(finalAllowSet))
	composite.Permissions.Deny = condenseActions(setToSlice(finalDenySet))
}

// resolveInternalPermissionConflicts resolves Allow/Deny conflicts within a single role
// and handles condensed actions properly
func (c *Config) resolveInternalPermissionConflicts(role *models.Role) {
	// Expand all condensed actions to individual permissions
	expandedAllowSet := expandPermissionsToSet(role.Permissions.Allow)
	expandedDenySet := expandPermissionsToSet(role.Permissions.Deny)

	// Resolve conflicts: remove permissions that are both allowed and denied
	// Within a single role, deny takes precedence by removing from allow
	for denyPerm := range expandedDenySet {
		if expandedAllowSet[denyPerm] {
			// Remove from allow set (deny wins the conflict)
			delete(expandedAllowSet, denyPerm)
			// Remove from deny set too since there's no point denying something not allowed
			delete(expandedDenySet, denyPerm)
		}
	}

	// Convert expanded sets back to slices and condense actions back for cleaner output
	role.Permissions.Allow = condenseActions(setToSlice(expandedAllowSet))
	role.Permissions.Deny = condenseActions(setToSlice(expandedDenySet))
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

// expandCondensedActions expands a condensed permission like "k8s:pods:get,list,watch"
// into individual permissions ["k8s:pods:get", "k8s:pods:list", "k8s:pods:watch"]
func expandCondensedActions(permission string) []string {
	// Find the last colon
	idx := strings.LastIndex(permission, ":")
	if idx == -1 {
		return []string{permission}
	}

	resource := permission[:idx]
	actions := permission[idx+1:]

	// Check if actions contains comma
	if !strings.Contains(actions, ",") {
		return []string{permission}
	}

	actionParts := strings.Split(actions, ",")
	result := make([]string, 0, len(actionParts))

	for i, action := range actionParts {
		action = strings.TrimSpace(action)
		if action != "" {
			result = append(result, resource+":"+action)
		} else {
			logrus.Warnf("Empty action detected in permission string '%s' at position %d. This may indicate a configuration error.", permission, i)
		}
	}

	return result
}

// condenseActions takes individual permissions and groups them by resource,
// condensing actions where possible
func condenseActions(permissions []string) []string {
	resourceActions := make(map[string][]string)

	for _, perm := range permissions {
		idx := strings.LastIndex(perm, ":")
		if idx == -1 {
			resourceActions[perm] = []string{""}
			continue
		}

		resource := perm[:idx]
		action := perm[idx+1:]

		resourceActions[resource] = append(resourceActions[resource], action)
	}

	result := make([]string, 0, len(resourceActions))
	for resource, actions := range resourceActions {
		if len(actions) == 1 && len(actions[0]) == 0 {
			// This was a permission without actions
			result = append(result, resource)
		} else if len(actions) == 1 {
			// Single action
			result = append(result, resource+":"+actions[0])
		} else {
			// Multiple actions - check for wildcard
			hasWildcard := slices.Contains(actions, "*")

			if hasWildcard {
				// Wildcard overrides all other actions
				result = append(result, resource+":*")
			} else {
				// Multiple specific actions - condense them
				sort.Strings(actions)
				result = append(result, resource+":"+strings.Join(actions, ","))
			}
		}
	}

	sort.Strings(result)
	return result
}
