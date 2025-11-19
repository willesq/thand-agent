package okta

import (
	"context"
	"fmt"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// AuthorizeRole assigns a role to a user in Okta
func (p *oktaProvider) AuthorizeRole(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
) (*models.AuthorizeRoleResponse, error) {
	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to authorize Okta role")
	}

	user := req.GetUser()
	role := req.GetRole()

	// Get the Okta user
	oktaUser, _, err := p.client.User.GetUser(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user in Okta: %w", err)
	}

	// Determine which Okta roles to assign
	var rolesToAssign []string
	if len(role.Inherits) > 0 {
		// If the role inherits from other roles, assign those roles
		rolesToAssign = role.Inherits
	} else if len(role.Permissions.Allow) > 0 {
		// TODO: If permissions are provided, we should ideally create a custom role
		// or apply permissions directly. For now, we'll fall back to using the role name
		// assuming it corresponds to a pre-created custom role in Okta.
		logrus.Warnf("Role %s has permissions defined but custom role creation is not fully implemented. Using role name as Okta role type.", role.Name)
		rolesToAssign = []string{role.Name}
	} else {
		// Default to using the role name
		rolesToAssign = []string{role.Name}
	}

	var assignedRoleIds []string

	for _, roleType := range rolesToAssign {
		// Prepare role assignment request
		roleAssignment := okta.AssignRoleRequest{
			Type: roleType,
		}

		// Assign the role to the user
		assignedRole, _, err := p.client.User.AssignRoleToUser(ctx, oktaUser.Id, roleAssignment, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to assign role %s to user: %w", roleType, err)
		}

		assignedRoleIds = append(assignedRoleIds, assignedRole.Id)

		logrus.WithFields(logrus.Fields{
			"user_id":       oktaUser.Id,
			"user_email":    user.Email,
			"role_type":     roleType,
			"assignment_id": assignedRole.Id,
		}).Info("Successfully assigned role to user in Okta")
	}

	// Return metadata for later revocation
	metadata := map[string]any{
		"user_id":  oktaUser.Id,
		"role_ids": assignedRoleIds,
	}

	return &models.AuthorizeRoleResponse{
		Metadata: metadata,
	}, nil
}

// RevokeRole removes a role from a user in Okta
func (p *oktaProvider) RevokeRole(
	ctx context.Context,
	req *models.RevokeRoleRequest,
) (*models.RevokeRoleResponse, error) {
	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to revoke Okta role")
	}

	user := req.GetUser()

	// Get the Okta user
	oktaUser, _, err := p.client.User.GetUser(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user in Okta: %w", err)
	}

	// Get role IDs to revoke
	var roleIds []string

	// Try to get from metadata first
	if req.AuthorizeRoleResponse != nil && req.AuthorizeRoleResponse.Metadata != nil {
		if ids, ok := req.AuthorizeRoleResponse.Metadata["role_ids"].([]string); ok {
			roleIds = ids
		} else if ids, ok := req.AuthorizeRoleResponse.Metadata["role_ids"].([]interface{}); ok {
			// Handle JSON unmarshaling
			for _, id := range ids {
				if strId, ok := id.(string); ok {
					roleIds = append(roleIds, strId)
				}
			}
		} else if id, ok := req.AuthorizeRoleResponse.Metadata["role_id"].(string); ok {
			// Backward compatibility
			roleIds = []string{id}
		}
	}

	// If we don't have the role IDs, we need to find them by listing the user's roles
	if len(roleIds) == 0 {
		roles, _, err := p.client.User.ListAssignedRolesForUser(ctx, oktaUser.Id, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list user roles: %w", err)
		}

		// Determine which role types we are looking for
		var targetRoleTypes []string
		if len(req.GetRole().Inherits) > 0 {
			targetRoleTypes = req.GetRole().Inherits
		} else {
			targetRoleTypes = []string{req.GetRole().Name}
		}

		// Find the roles by type
		for _, targetType := range targetRoleTypes {
			for _, role := range roles {
				if role.Type == targetType {
					roleIds = append(roleIds, role.Id)
					break
				}
			}
		}

		if len(roleIds) == 0 {
			return nil, fmt.Errorf("role assignment not found for user")
		}
	}

	// Remove the roles from the user
	for _, roleId := range roleIds {
		_, err = p.client.User.RemoveRoleFromUser(ctx, oktaUser.Id, roleId)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to remove role %s from user", roleId)
			// We continue trying to remove other roles even if one fails
		} else {
			logrus.WithFields(logrus.Fields{
				"user_id":    oktaUser.Id,
				"user_email": user.Email,
				"role_id":    roleId,
			}).Info("Successfully revoked role from user in Okta")
		}
	}

	return &models.RevokeRoleResponse{}, nil
}

// GetAuthorizedAccessUrl returns the URL where the user can access their Okta dashboard
func (p *oktaProvider) GetAuthorizedAccessUrl(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
	resp *models.AuthorizeRoleResponse,
) string {
	// Return the Okta organization URL where users can log in
	return p.orgUrl
}

// Helper function to list groups in Okta
func (p *oktaProvider) ListGroups(ctx context.Context) ([]*okta.Group, error) {
	groups, _, err := p.client.Group.ListGroups(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	return groups, nil
}

// Helper function to get a specific group
func (p *oktaProvider) GetGroup(ctx context.Context, groupId string) (*okta.Group, error) {
	group, _, err := p.client.Group.GetGroup(ctx, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return group, nil
}

// Helper function to list users in Okta
func (p *oktaProvider) ListUsers(ctx context.Context) ([]*okta.User, error) {
	users, _, err := p.client.User.ListUsers(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// Helper function to get a specific user
func (p *oktaProvider) GetUser(ctx context.Context, userId string) (*okta.User, error) {
	user, _, err := p.client.User.GetUser(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// Helper function to list applications in Okta
func (p *oktaProvider) ListApplications(ctx context.Context) ([]okta.App, error) {
	apps, _, err := p.client.Application.ListApplications(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	return apps, nil
}

// Helper function to add user to a group
func (p *oktaProvider) AddUserToGroup(ctx context.Context, groupId string, userId string) error {
	_, err := p.client.Group.AddUserToGroup(ctx, groupId, userId)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}
	return nil
}

// Helper function to remove user from a group
func (p *oktaProvider) RemoveUserFromGroup(ctx context.Context, groupId string, userId string) error {
	_, err := p.client.Group.RemoveUserFromGroup(ctx, groupId, userId)
	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}
	return nil
}
