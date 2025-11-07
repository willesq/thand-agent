package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

const CloudflareAllow = "allow"
const CloudflareDeny = "deny"

// AuthorizeRole grants access for a user to a role in Cloudflare
// Supports both account-wide roles and resource-scoped policies
func (p *cloudflareProvider) AuthorizeRole(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
) (*models.AuthorizeRoleResponse, error) {
	// Check for nil inputs
	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to authorize Cloudflare role")
	}

	user := req.GetUser()
	role := req.GetRole()

	logrus.WithFields(logrus.Fields{
		"user": user.Email,
		"role": role.Name,
	}).Info("Authorizing Cloudflare role")

	// Get the account resource container
	accountID := p.GetAccountID()
	accountRC := cloudflare.AccountIdentifier(accountID)

	params := cloudflare.CreateAccountMemberParams{
		EmailAddress: user.Email,
	}

	// Use policy-based RBAC for granular resource access
	// Map the role name to get permission group IDs, then build policies
	err := p.buildMembershipFromRole(ctx, &params, role)
	if err != nil {
		return nil, fmt.Errorf("failed to build policies: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user":         user.Email,
		"role":         role.Name,
		"policy_count": len(params.Policies),
		"role_count":   len(params.Roles),
	}).Info("Creating account member with resource-scoped policies")

	// Check if the member already exists
	existingMember, err := p.findAccountMember(ctx, user.Email)
	if err == nil && existingMember != nil {
		// Member exists, update instead of create
		logrus.WithFields(logrus.Fields{
			"user":      user.Email,
			"member_id": existingMember.ID,
		}).Info("Member already exists, updating instead")

		// Update the member with new roles/policies
		// Convert role IDs to AccountRole objects
		var accountRoles []cloudflare.AccountRole
		for _, roleID := range params.Roles {
			accountRoles = append(accountRoles, cloudflare.AccountRole{ID: roleID})
		}
		existingMember.Roles = accountRoles
		existingMember.Policies = params.Policies

		updatedMember, err := p.client.UpdateAccountMember(ctx, accountID, existingMember.ID, *existingMember)
		if err != nil {
			return nil, fmt.Errorf("failed to update account member: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"user":      user.Email,
			"role":      role.Name,
			"member_id": updatedMember.ID,
		}).Info("Successfully updated Cloudflare role")

		return &models.AuthorizeRoleResponse{
			Metadata: map[string]any{
				"member_id": updatedMember.ID,
				"status":    updatedMember.Status,
				"updated":   true,
			},
		}, nil
	}

	// Member doesn't exist, create new
	member, err := p.client.CreateAccountMember(ctx, accountRC, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account member: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user":      user.Email,
		"role":      role.Name,
		"member_id": member.ID,
	}).Info("Successfully authorized Cloudflare role")

	// Return metadata about the authorization
	return &models.AuthorizeRoleResponse{
		Metadata: map[string]any{
			"member_id": member.ID,
			"status":    member.Status,
		},
	}, nil
}

// RevokeRole removes access for a user from a role in Cloudflare
// Handles both account-wide roles and resource-scoped policies
func (p *cloudflareProvider) RevokeRole(
	ctx context.Context,
	req *models.RevokeRoleRequest,
) (*models.RevokeRoleResponse, error) {
	// Check for nil inputs
	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to revoke Cloudflare role")
	}

	user := req.GetUser()
	role := req.GetRole()

	logrus.WithFields(logrus.Fields{
		"user": user.Email,
		"role": role.Name,
	}).Info("Revoking Cloudflare role")

	// Get the member ID from the authorization metadata if available
	var memberID string
	if req.AuthorizeRoleResponse != nil && req.AuthorizeRoleResponse.Metadata != nil {
		if id, ok := req.AuthorizeRoleResponse.Metadata["member_id"].(string); ok {
			memberID = id
		}
	}

	// If we don't have the member ID, we need to look it up
	if len(memberID) == 0 {
		accountID := p.GetAccountID()

		// List account members to find the user
		members, _, err := p.client.AccountMembers(ctx, accountID, cloudflare.PaginationOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list account members: %w", err)
		}

		for _, member := range members {
			if member.User.Email == user.Email {
				memberID = member.ID
				break
			}
		}

		if len(memberID) == 0 {
			return nil, fmt.Errorf("user %s not found in account members", user.Email)
		}
	}

	// Note: Removing the member removes ALL their access (both roles and policies)
	// If partial revocation is needed in the future, use UpdateAccountMember instead
	accountID := p.GetAccountID()

	err := p.client.DeleteAccountMember(ctx, accountID, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete account member: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user":      user.Email,
		"role":      role.Name,
		"member_id": memberID,
	}).Info("Successfully revoked Cloudflare role")

	return &models.RevokeRoleResponse{}, nil
}

// findAccountMember finds an existing account member by email
func (p *cloudflareProvider) findAccountMember(ctx context.Context, email string) (*cloudflare.AccountMember, error) {
	accountID := p.GetAccountID()

	// List account members to find the user
	members, _, err := p.client.AccountMembers(ctx, accountID, cloudflare.PaginationOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list account members: %w", err)
	}

	for _, member := range members {
		if member.User.Email == email {
			return &member, nil
		}
	}

	return nil, fmt.Errorf("member not found: %s", email)
}

// GetAuthorizedAccessUrl returns the URL to access Cloudflare dashboard
func (p *cloudflareProvider) GetAuthorizedAccessUrl(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
	resp *models.AuthorizeRoleResponse,
) string {
	// Return the Cloudflare dashboard URL for the account
	accountID := p.GetAccountID()
	return fmt.Sprintf("https://dash.cloudflare.com/%s", accountID)
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
				for permKey := range cfRole.Permissions {
					availablePermissions[permKey] = true
				}
			}

			logrus.WithFields(logrus.Fields{
				"role":                  role.Name,
				"available_permissions": len(availablePermissions),
			}).Debug("Built permission map from inherited roles")
		}

		// Build permission groups for allow and deny
		var allowPermissionGroups []cloudflare.PermissionGroup
		var denyPermissionGroups []cloudflare.PermissionGroup

		// Process allow permissions
		for _, permName := range role.Permissions.Allow {
			// If we have inherited roles, verify the permission is available
			if len(role.Inherits) > 0 && !availablePermissions[permName] {
				logrus.WithFields(logrus.Fields{
					"role":       role.Name,
					"permission": permName,
				}).Warn("Allow permission not available in inherited roles, skipping")
				continue
			}

			allowPermissionGroups = append(allowPermissionGroups, cloudflare.PermissionGroup{
				ID: permName,
			})
		}

		// Process deny permissions
		for _, permName := range role.Permissions.Deny {
			// If we have inherited roles, verify the permission is available
			if len(role.Inherits) > 0 && !availablePermissions[permName] {
				logrus.WithFields(logrus.Fields{
					"role":       role.Name,
					"permission": permName,
				}).Warn("Deny permission not available in inherited roles, skipping")
				continue
			}

			denyPermissionGroups = append(denyPermissionGroups, cloudflare.PermissionGroup{
				ID: permName,
			})
		}

		// Build resource groups from the role's resource specifications
		resourceGroups, err := p.buildResourceGroups(ctx, role.Resources.Allow)
		if err != nil {
			return fmt.Errorf("failed to build resource groups: %w", err)
		}

		// Create policies for each resource group
		var policies []cloudflare.Policy
		for _, resourceGroup := range resourceGroups {
			if len(allowPermissionGroups) > 0 {
				policy := cloudflare.Policy{
					PermissionGroups: allowPermissionGroups,
					ResourceGroups: []cloudflare.ResourceGroup{
						resourceGroup,
					},
					Access: CloudflareAllow,
				}
				policies = append(policies, policy)
			}

			if len(denyPermissionGroups) > 0 {
				policy := cloudflare.Policy{
					PermissionGroups: denyPermissionGroups,
					ResourceGroups: []cloudflare.ResourceGroup{
						resourceGroup,
					},
					Access: CloudflareDeny,
				}
				policies = append(policies, policy)
			}
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

// getRoleIDsFromInherits gets role IDs from multiple Cloudflare roles
// This extracts the role IDs for direct role assignment in CreateAccountMemberParams
func (p *cloudflareProvider) getRoleIDsFromInherits(inherits []string) ([]string, error) {
	if len(inherits) == 0 {
		return nil, fmt.Errorf("no roles specified in inherits - must specify at least one Cloudflare role (e.g., 'DNS', 'Firewall', 'Workers Admin')")
	}

	var roleIDs []string
	seenIDs := make(map[string]bool) // Track to avoid duplicates

	// Get role ID from each inherited role
	for _, roleName := range inherits {
		// Look up the cached Cloudflare role
		cfRole, ok := p.cfRolesMap[strings.ToLower(roleName)]
		if !ok {
			return nil, fmt.Errorf("role '%s' is not a recognized Cloudflare role - when using resource-scoped policies, the role name must match a Cloudflare role (e.g., 'DNS', 'Firewall', 'Workers Admin')", roleName)
		}

		// Add unique role IDs
		if !seenIDs[cfRole.ID] {
			roleIDs = append(roleIDs, cfRole.ID)
			seenIDs[cfRole.ID] = true
		}
	}

	if len(roleIDs) == 0 {
		return nil, fmt.Errorf("no role IDs found for inherited roles: %v", inherits)
	}

	logrus.WithFields(logrus.Fields{
		"inherited_roles": inherits,
		"role_id_count":   len(roleIDs),
	}).Debug("Retrieved role IDs from inherited Cloudflare roles")

	return roleIDs, nil
}

// buildResourceGroups creates Cloudflare resource groups from resource specifications
func (p *cloudflareProvider) buildResourceGroups(ctx context.Context, resources []string) ([]cloudflare.ResourceGroup, error) {
	var resourceGroups []cloudflare.ResourceGroup

	accountID := p.GetAccountID()

	for _, resource := range resources {
		var rg cloudflare.ResourceGroup

		// Parse resource specification
		// Format: "zone:example.com" or "account:*" or "zone:*"
		if resource == "*" || resource == fmt.Sprintf("%s:*", resourceTypeAccount) {
			// Full account access - use cached account
			var account cloudflare.Account
			found := false

			// Find the account in cached resources
			for _, res := range p.resources {
				if res.Type == resourceTypeAccount && res.Id == accountID {
					if acc, ok := res.Resource.(cloudflare.Account); ok {
						account = acc
						found = true
						break
					}
				}
			}

			if !found {
				// Fallback to API call if not found in cache
				var err error
				account, _, err = p.client.Account(ctx, accountID)
				if err != nil {
					return nil, fmt.Errorf("failed to get account: %w", err)
				}
			}

			rg = cloudflare.NewResourceGroupForAccount(account)
		} else if len(resource) > 5 && resource[:5] == fmt.Sprintf("%s:", resourceTypeZone) {
			// Zone-specific access
			zoneName := resource[5:]
			if zoneName == "*" {
				// All zones - use cached resources
				for _, res := range p.resources {
					if res.Type == resourceTypeZone {
						// Use cached zone from Resource field (no API calls needed)
						if zone, ok := res.Resource.(cloudflare.Zone); ok {
							resourceGroups = append(resourceGroups, cloudflare.NewResourceGroupForZone(zone))
						} else {
							logrus.WithFields(logrus.Fields{
								resourceTypeZone: res.Name,
								"zoneID":         res.Id,
							}).Warn("Zone resource does not contain zone details, skipping")
							continue
						}
					}
				}
				continue
			} else {
				// Specific zone - look up from cache first
				cachedResource, err := p.GetResource(ctx, resource)
				if err != nil {
					return nil, fmt.Errorf("zone not found in cache: %s", zoneName)
				}

				if cachedResource.Type == resourceTypeZone {
					// Use cached zone from Resource field (no API calls needed)
					if zone, ok := cachedResource.Resource.(cloudflare.Zone); ok {
						rg = cloudflare.NewResourceGroupForZone(zone)
					} else {
						return nil, fmt.Errorf("zone resource does not contain zone details: %s", zoneName)
					}
				} else {
					return nil, fmt.Errorf("cached resource is not a zone: %s", zoneName)
				}
			}
		} else {
			// Custom resource group key
			rg = cloudflare.NewResourceGroup(resource)
		}

		resourceGroups = append(resourceGroups, rg)
	}

	if len(resourceGroups) == 0 {
		return nil, fmt.Errorf("no resource groups could be created from resources: %v", resources)
	}

	return resourceGroups, nil
}
