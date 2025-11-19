package okta

import (
	"context"
	"fmt"
	"strings"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// CustomAdminRoleResponse represents the response from creating a custom admin role in Okta
type CustomAdminRoleResponse struct {
	ID          string               `json:"id"`
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Created     string               `json:"created"`
	LastUpdated string               `json:"lastUpdated"`
	Links       CustomAdminRoleLinks `json:"_links"`
}

// CustomAdminRoleLinks represents the links in the custom admin role response
type CustomAdminRoleLinks struct {
	Permissions CustomAdminRoleLink `json:"permissions"`
	Self        CustomAdminRoleLink `json:"self"`
}

// CustomAdminRoleLink represents a single link in the custom admin role response
type CustomAdminRoleLink struct {
	Href string `json:"href"`
}

// ResourceSetAssignment represents a principal assignment to a resource set
type ResourceSetAssignment struct {
	PrincipalID     string `json:"principalId"`
	PrincipalType   string `json:"principalType"`
	PermissionSetID string `json:"permissionSetId"`
	ResourceSetID   string `json:"resourceSetId"`
}

// ResourceSetAssignmentRequest represents the request body for resource set assignments
type ResourceSetAssignmentRequest struct {
	Add    []ResourceSetAssignment `json:"add,omitempty"`
	Remove []ResourceSetAssignment `json:"remove,omitempty"`
}

const MetadataUserKey = "user"
const MetadataRolesKey = "roles"
const MetadataGroupKey = "groups"

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
	var assginedRoles []string
	var assignedGroups []string

	// Check if there are groups to assign
	if len(role.Groups.Allow) > 0 {

		for _, groupId := range role.Groups.Allow {

			// Get the okta groups and add the user to each
			identity, err := p.GetIdentity(ctx, groupId)

			if err != nil {
				return nil, fmt.Errorf("failed to get group %s: %w", groupId, err)
			}

			if identity.GetGroup() == nil {
				return nil, fmt.Errorf("group %s not found in Okta", groupId)
			}

			err = p.AddUserToGroup(ctx, identity.GetGroup().ID, oktaUser.Id)

			if err != nil {
				return nil, fmt.Errorf("failed to add user to group %s: %w", groupId, err)
			}

			logrus.WithFields(logrus.Fields{
				"user_id":    oktaUser.Id,
				"user_email": user.Email,
				"group_id":   groupId,
			}).Info("Successfully added user to group in Okta")

			assignedGroups = append(assignedGroups, groupId)
		}
	}

	if len(role.Inherits) > 0 {
		// If the role inherits from other roles, assign those roles

		for _, roleType := range role.Inherits {
			// Prepare role assignment request
			roleAssignment := okta.AssignRoleRequest{
				Type: roleType,
			}

			// Assign the role to the user
			assignedRole, _, err := p.client.User.AssignRoleToUser(ctx, oktaUser.Id, roleAssignment, nil)
			if err != nil {
				if oktaErr, ok := err.(*okta.Error); ok {
					if strings.ToUpper(oktaErr.ErrorCode) == "E0000090" {

						logrus.WithFields(logrus.Fields{
							"user_id":    oktaUser.Id,
							"user_email": user.Email,
							"role_type":  roleType,
						}).Info("User already has the role assigned in Okta, skipping assignment")

						// IF the role has already been assigned, just skip
						// we don't need to mark it for removal later
						continue
					}
				}

				return nil, fmt.Errorf("failed to assign role %s to user: %w", roleType, err)
			}

			assginedRoles = append(assginedRoles, assignedRole.Id)

			logrus.WithFields(logrus.Fields{
				"user_id":       oktaUser.Id,
				"user_email":    user.Email,
				"role_type":     roleType,
				"assignment_id": assignedRole.Id,
			}).Info("Successfully assigned role to user in Okta")
		}

	} else if len(role.Permissions.Allow) > 0 {

		// Create a custom admin role with the specified permissions
		customRoleType, err := p.createCustomAdminRole(ctx, role)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom admin role: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"role_name":        role.Name,
			"custom_role_type": customRoleType,
			"permissions":      role.Permissions.Allow,
		}).Info("Created custom admin role in Okta")

		// Assign the role to the user
		err = p.assignCustomRoleToUser(ctx, ResourceSetAssignmentRequest{
			Add: []ResourceSetAssignment{
				{
					PrincipalID:     oktaUser.Id,
					PrincipalType:   "USER",
					PermissionSetID: customRoleType.ID,
					ResourceSetID:   "", // Assuming empty for now; adjust as needed
				},
			},
		})

		if err != nil {
			return nil, fmt.Errorf("failed to assign custom role to user: %w", err)
		}

		assginedRoles = append(assginedRoles, customRoleType.ID)

	} else {

		return nil, fmt.Errorf("role %s has no inherits or permissions defined", role.Name)
	}

	// Return metadata for later revocation
	metadata := map[string]any{
		MetadataUserKey:  oktaUser.Id,
		MetadataRolesKey: assginedRoles,
		MetadataGroupKey: assignedGroups,
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

	if req.AuthorizeRoleResponse == nil {
		return nil, fmt.Errorf("no authorize role response metadata found for revocation")
	}

	// Try to get from metadata first
	if req.AuthorizeRoleResponse.Metadata == nil {
		return nil, fmt.Errorf("no authorize role response metadata found for revocation")
	}

	metadata := req.AuthorizeRoleResponse.Metadata

	if roleIds, ok := metadata[MetadataRolesKey].([]string); ok {

		// Lets revoke each role
		for _, roleId := range roleIds {

			// First check if the role is a custom role
			foundRole, err := p.GetRole(ctx, roleId)

			if err != nil {

				// This is a custom role we need to remove via resource set assignments

				err = p.assignCustomRoleToUser(ctx, ResourceSetAssignmentRequest{
					Remove: []ResourceSetAssignment{
						{
							PrincipalID:     oktaUser.Id,
							PrincipalType:   "USER",
							PermissionSetID: roleId,
							ResourceSetID:   "", // Assuming empty for now; adjust as needed
						},
					},
				})

				if err != nil {
					return nil, fmt.Errorf("failed to revoke custom role %s from user: %w", roleId, err)
				}

				logrus.WithFields(logrus.Fields{
					"user_id":    oktaUser.Id,
					"user_email": user.Email,
					"role_id":    roleId,
				}).Info("Successfully revoked custom role from user in Okta")

			} else if foundRole != nil {
				// This is a standard role, remove via Assignments API

				// Remove the role from the user
				_, err = p.client.User.RemoveRoleFromUser(ctx, oktaUser.Id, roleId)

				if err != nil {
					return nil, fmt.Errorf("failed to revoke role %s from user: %w", roleId, err)
				}

				logrus.WithFields(logrus.Fields{
					"user_id":    oktaUser.Id,
					"user_email": user.Email,
					"role_id":    roleId,
				}).Info("Successfully revoked role from user in Okta")
			}
		}

		// Right now lets remove the user from any groups assigned during authorization
		if groupIds, ok := metadata[MetadataGroupKey].([]string); ok {

			for _, groupId := range groupIds {

				err = p.RemoveUserFromGroup(ctx, groupId, oktaUser.Id)

				if err != nil {
					return nil, fmt.Errorf("failed to remove user from group %s: %w", groupId, err)
				}

				logrus.WithFields(logrus.Fields{
					"user_id":    oktaUser.Id,
					"user_email": user.Email,
					"group_id":   groupId,
				}).Info("Successfully removed user from group in Okta")
			}
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
