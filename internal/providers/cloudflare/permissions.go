//go:build exclude

// Cloudflare only supports roles when assigning accounts for access
// I am leaving this here for the future in case cloudflare adds support
// for more granular permissions in the future - aside from roles

package cloudflare

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// LoadPermissions loads Cloudflare permission reference data
func (p *cloudflareProvider) LoadPermissions() error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed Cloudflare permission references in %s", elapsed)
	}()

	if len(p.roles) == 0 {
		return fmt.Errorf("roles must be loaded before permissions")
	}

	var permissions []models.ProviderPermission
	permissionsMap := make(map[string]*models.ProviderPermission)

	// Convert to slice and create fast lookup map
	for _, providerRole := range p.roles {

		role, ok := providerRole.Role.(cloudflare.AccountRole) // Safe type assertion
		if !ok {
			return fmt.Errorf("providerRole.Role is not of type cloudflare.AccountRole")
		}

		rolePermissions := role.Permissions

		for name := range rolePermissions {
			// Check if permission already exists
			lowerName := strings.ToLower(name)

			readName := fmt.Sprintf("%s:read", lowerName)
			readPerm := models.ProviderPermission{
				Name:        readName,
				Title:       fmt.Sprintf("%s Read", name),
				Description: fmt.Sprintf("Read access for %s", name),
			}
			permissions = append(permissions, readPerm)
			permissionsMap[readName] = &permissions[len(permissions)-1]

			editName := fmt.Sprintf("%s:edit", lowerName)
			editPerm := models.ProviderPermission{
				Name:        editName,
				Title:       fmt.Sprintf("%s Edit", name),
				Description: fmt.Sprintf("Edit access for %s", name),
			}
			permissions = append(permissions, editPerm)
			permissionsMap[editName] = &permissions[len(permissions)-1]
		}
	}

	p.permissions = permissions
	p.permissionsMap = permissionsMap

	logrus.WithFields(logrus.Fields{
		"permissions": len(permissions),
	}).Debug("Loaded Cloudflare permissions, building search index in background")

	return nil
}

// GetPermission retrieves a specific permission by name
func (p *cloudflareProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	// Remove any legacy '#' prefix to support old references
	permission = strings.TrimPrefix(strings.ToLower(permission), "#")
	// Fast map lookup
	if perm, exists := p.permissionsMap[permission]; exists {
		return perm, nil
	}
	return nil, fmt.Errorf("permission not found")
}

// ListPermissions lists all permissions, optionally filtered
func (p *cloudflareProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {
	// If no filters, return all permissions
	if len(filters) == 0 {
		return p.permissions, nil
	}

	// Check if search index is ready
	p.indexMu.RLock()
	permissionsIndex := p.permissionsIndex
	p.indexMu.RUnlock()

	if permissionsIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, permissionsIndex, func(a *search.DocumentMatch, b models.ProviderPermission) bool {
			return strings.Compare(a.ID, b.Name) == 0
		}, p.permissions, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []models.ProviderPermission
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, perm := range p.permissions {
		// Check if any filter matches the permission name or description
		if strings.Contains(strings.ToLower(perm.Name), filterText) ||
			strings.Contains(strings.ToLower(perm.Description), filterText) {
			filtered = append(filtered, perm)
		}
	}

	return filtered, nil
}

// buildMembershipFromRole creates Cloudflare policies from role definition
// If only inherits are provided, assigns those roles directly
// If permissions are provided, builds granular policies based on inherited role permissions
func (p *cloudflareProvider) buildMembershipFromRole(
	ctx context.Context,
	params *cloudflare.CreateAccountMemberParams,
	role *models.Role,
) error {

	// Check if we have any permissions specified
	hasPermissions := len(role.Permissions.Allow) > 0 || len(role.Permissions.Deny) > 0

	// If only inherits provided (no permissions), just assign the roles directly
	if len(role.Inherits) > 0 && !hasPermissions {
		roleIDs, err := p.getRoleIDsFromInherits(role.Inherits)
		if err != nil {
			return fmt.Errorf("failed to get role IDs from inherits: %w", err)
		}
		params.Roles = roleIDs

		logrus.WithFields(logrus.Fields{
			"role":       role.Name,
			"role_ids":   roleIDs,
			"role_count": len(roleIDs),
		}).Debug("Assigned roles directly from inherits")

		return nil
	}

	// If permissions are provided, build granular policies
	if hasPermissions {
		// Build a map of all available permissions from inherited roles
		availablePermissions := make(map[string]bool)

		if len(role.Inherits) > 0 {
			for _, roleName := range role.Inherits {
				cfRole, ok := p.cfRolesMap[strings.ToLower(roleName)]
				if !ok {
					logrus.WithFields(logrus.Fields{
						"role":           role.Name,
						"inherited_role": roleName,
					}).Warn("Inherited role not found in cache, skipping")
					continue
				}

				// Add all permissions from this role to available permissions
				for permKey, permValue := range cfRole.Permissions {

					editName := fmt.Sprintf("%s:edit", permKey)
					readName := fmt.Sprintf("%s:read", permKey)

					availablePermissions[editName] = permValue.Edit
					availablePermissions[readName] = permValue.Read
				}
			}

			logrus.WithFields(logrus.Fields{
				"role":                  role.Name,
				"available_permissions": len(availablePermissions),
			}).Debug("Built permission map from inherited roles")
		}

		// Process allow permissions
		for _, permName := range role.Permissions.Allow {
			availablePermissions[strings.ToLower(permName)] = true
		}

		// Process deny permissions
		for _, permName := range role.Permissions.Deny {
			availablePermissions[strings.ToLower(permName)] = false
		}

		var allowPermissionGroups []cloudflare.Permission
		var denyPermissionGroups []cloudflare.Permission

		// Only include permissions explicitly specified in Allow and Deny
		for permName, permVal := range availablePermissions {
			if permVal {
				allowPermissionGroups = append(allowPermissionGroups, cloudflare.Permission{
					ID: fmt.Sprintf("#%s", permName),
				})
			} else {
				denyPermissionGroups = append(denyPermissionGroups, cloudflare.Permission{
					ID: fmt.Sprintf("#%s", permName),
				})
			}
		}

		resourceGroups, err := p.buildResourceGroups(ctx, role.Resources.Allow)
		if err != nil {
			return fmt.Errorf("failed to build resource groups: %w", err)
		}

		// Create policies for each resource group
		var policies []cloudflare.Policy
		for _, resourceGroup := range resourceGroups {
			if len(allowPermissionGroups) > 0 {
				policy := cloudflare.Policy{
					PermissionGroups: []cloudflare.PermissionGroup{{
						ID:          common.ConvertToSnakeCase(role.Name),
						Name:        role.Name,
						Permissions: allowPermissionGroups,
					}},
					ResourceGroups: []cloudflare.ResourceGroup{
						resourceGroup,
					},
					Access: CloudflareAllow,
				}

				policies = append(policies, policy)
			}
		}

		if len(policies) == 0 {
			return fmt.Errorf("no policies could be built from the provided permissions and resources")
		}

		params.Policies = policies

		logrus.WithFields(logrus.Fields{
			"role":         role.Name,
			"policy_count": len(policies),
			"allow_perms":  len(allowPermissionGroups),
			"deny_perms":   len(denyPermissionGroups),
		}).Debug("Built granular policies from permissions")

		return nil
	}

	return fmt.Errorf("role must specify either inherits or permissions")
}
