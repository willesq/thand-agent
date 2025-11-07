package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

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
// Uses the role's Permissions field to set allow/deny permission groups
// Optionally uses Inherits field for backward compatibility with Cloudflare role IDs
func (p *cloudflareProvider) buildMembershipFromRole(
	ctx context.Context,
	params *cloudflare.CreateAccountMemberParams,
	role *models.Role,
) error {

	var policies []cloudflare.Policy
	var roleIDs []string

	// Get role IDs from inherited Cloudflare roles if specified (optional, uses cached data)
	if len(role.Inherits) > 0 {
		ids, err := p.getRoleIDsFromInherits(role.Inherits)
		if err != nil {
			// Log warning but don't fail - inherits is optional when using permissions directly
			logrus.WithFields(logrus.Fields{
				"role":     role.Name,
				"inherits": role.Inherits,
				"error":    err,
			}).Warn("Failed to get role IDs from inherits, continuing with permission-based policies")
		} else {
			roleIDs = ids
		}
	}

	// Get permission groups from role's permissions (allow and deny)
	allowPermissionGroups, err := p.getPermissionGroupsFromPermissions(role.Permissions.Allow)
	if err != nil {
		return fmt.Errorf("failed to get allow permission groups: %w", err)
	}

	denyPermissionGroups, err := p.getPermissionGroupsFromPermissions(role.Permissions.Deny)
	if err != nil {
		return fmt.Errorf("failed to get deny permission groups: %w", err)
	}

	// Build resource groups from the role's resource specifications
	resourceGroups, err := p.buildResourceGroups(ctx, role.Resources.Allow)
	if err != nil {
		return fmt.Errorf("failed to build resource groups: %w", err)
	}

	// Create allow policies for each resource group with the specified permission groups
	for _, resourceGroup := range resourceGroups {
		if len(allowPermissionGroups) > 0 {
			policy := cloudflare.Policy{
				PermissionGroups: allowPermissionGroups,
				ResourceGroups: []cloudflare.ResourceGroup{
					resourceGroup,
				},
				Access: "allow",
			}
			policies = append(policies, policy)
		}

		// Create deny policies for each resource group
		if len(denyPermissionGroups) > 0 {
			policy := cloudflare.Policy{
				PermissionGroups: denyPermissionGroups,
				ResourceGroups: []cloudflare.ResourceGroup{
					resourceGroup,
				},
				Access: "deny",
			}
			policies = append(policies, policy)
		}
	}

	// Set both role IDs and policies
	params.Roles = roleIDs
	params.Policies = policies

	return nil
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

// getPermissionGroupsFromPermissions creates permission groups from permission names
// Maps permission names (analytics, dns, billing, etc.) to Cloudflare PermissionGroup objects
func (p *cloudflareProvider) getPermissionGroupsFromPermissions(permissions []string) ([]cloudflare.PermissionGroup, error) {
	if len(permissions) == 0 {
		// Empty permissions is valid - return empty array
		return []cloudflare.PermissionGroup{}, nil
	}

	var permissionGroups []cloudflare.PermissionGroup
	seenGroups := make(map[string]bool) // Track to avoid duplicates

	// Create a permission group for each permission name
	for _, permissionName := range permissions {
		// Avoid duplicates
		if seenGroups[permissionName] {
			continue
		}

		// Create a PermissionGroup with the permission name as the ID
		// Cloudflare uses permission keys like: analytics, dns, billing, etc.
		permissionGroups = append(permissionGroups, cloudflare.PermissionGroup{
			ID: permissionName,
		})
		seenGroups[permissionName] = true
	}

	logrus.WithFields(logrus.Fields{
		"permissions":            permissions,
		"permission_group_count": len(permissionGroups),
	}).Debug("Created permission groups from permission names")

	return permissionGroups, nil
}

// getPermissionGroupsFromInherits gets permission groups from multiple Cloudflare roles
// DEPRECATED: Use getPermissionGroupsFromPermissions instead
// This function is kept for backward compatibility but should not be used
// buildResourceGroups creates Cloudflare resource groups from resource specifications
func (p *cloudflareProvider) buildResourceGroups(ctx context.Context, resources []string) ([]cloudflare.ResourceGroup, error) {
	var resourceGroups []cloudflare.ResourceGroup

	accountID := p.GetAccountID()

	for _, resource := range resources {
		var rg cloudflare.ResourceGroup

		// Parse resource specification
		// Format: "zone:example.com" or "account:*" or "zone:*"
		if resource == "*" || resource == fmt.Sprintf("%s:*", resourceTypeAccount) {
			// Full account access
			account, _, err := p.client.Account(ctx, accountID)
			if err != nil {
				return nil, fmt.Errorf("failed to get account: %w", err)
			}
			rg = cloudflare.NewResourceGroupForAccount(account)
		} else if len(resource) > 5 && resource[:5] == fmt.Sprintf("%s:", resourceTypeZone) {
			// Zone-specific access
			zoneName := resource[5:]
			if zoneName == "*" {
				// All zones - use cached resources
				for _, res := range p.resources {
					if res.Type == resourceTypeZone {
						// Look up zone details from cache using ID
						zoneID, err := p.client.ZoneIDByName(res.Name)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								resourceTypeZone: res.Name,
								"error":          err,
							}).Warn("Failed to get zone ID for cached zone, skipping")
							continue
						}
						zone, err := p.client.ZoneDetails(ctx, zoneID)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								resourceTypeZone: res.Name,
								"error":          err,
							}).Warn("Failed to get zone details for cached zone, skipping")
							continue
						}
						resourceGroups = append(resourceGroups, cloudflare.NewResourceGroupForZone(zone))
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

					// Get zone details using the cached zone ID
					zone, err := p.client.ZoneDetails(ctx, cachedResource.Id)
					if err != nil {
						return nil, fmt.Errorf("failed to get zone details for %s: %w", zoneName, err)
					}
					rg = cloudflare.NewResourceGroupForZone(zone)
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
