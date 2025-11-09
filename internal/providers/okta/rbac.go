package okta

import (
	"context"
	"fmt"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// ValidateRole validates if a role can be assigned to an identity
func (p *oktaProvider) ValidateRole(
	ctx context.Context,
	identity *models.Identity,
	role *models.Role,
) (map[string]any, error) {
	if identity == nil || role == nil {
		return nil, fmt.Errorf("identity and role must be provided")
	}

	// Check if the role exists in our predefined roles
	_, err := p.GetRole(ctx, role.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid role: %w", err)
	}

	// Get the user from the identity
	user := identity.GetUser()
	if user == nil {
		return nil, fmt.Errorf("identity must have a user")
	}

	// Try to find the user in Okta
	oktaUser, _, err := p.client.User.GetUser(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found in Okta: %w", err)
	}

	metadata := map[string]any{
		"user_id":    oktaUser.Id,
		"user_email": user.Email,
		"role_type":  role.Name,
	}

	return metadata, nil
}

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

	// Prepare role assignment request
	roleAssignment := okta.AssignRoleRequest{
		Type: role.Name,
	}

	// Assign the role to the user
	assignedRole, _, err := p.client.User.AssignRoleToUser(ctx, oktaUser.Id, roleAssignment, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to assign role to user: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       oktaUser.Id,
		"user_email":    user.Email,
		"role_type":     role.Name,
		"assignment_id": assignedRole.Id,
	}).Info("Successfully assigned role to user in Okta")

	// Return metadata for later revocation
	metadata := map[string]any{
		"user_id":       oktaUser.Id,
		"role_id":       assignedRole.Id,
		"role_type":     assignedRole.Type,
		"assignment_id": assignedRole.Id,
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

	// If we have the role ID from the authorization metadata, use it directly
	var roleId string
	if req.AuthorizeRoleResponse != nil && req.AuthorizeRoleResponse.Metadata != nil {
		if id, ok := req.AuthorizeRoleResponse.Metadata["role_id"].(string); ok {
			roleId = id
		}
	}

	// If we don't have the role ID, we need to find it by listing the user's roles
	if roleId == "" {
		roles, _, err := p.client.User.ListAssignedRolesForUser(ctx, oktaUser.Id, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list user roles: %w", err)
		}

		// Find the role by type
		for _, role := range roles {
			if role.Type == req.GetRole().Name {
				roleId = role.Id
				break
			}
		}

		if roleId == "" {
			return nil, fmt.Errorf("role assignment not found for user")
		}
	}

	// Remove the role from the user
	_, err = p.client.User.RemoveRoleFromUser(ctx, oktaUser.Id, roleId)
	if err != nil {
		return nil, fmt.Errorf("failed to remove role from user: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    oktaUser.Id,
		"user_email": user.Email,
		"role_id":    roleId,
	}).Info("Successfully revoked role from user in Okta")

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

// GetResource returns information about an Okta resource
func (p *oktaProvider) GetResource(ctx context.Context, resource string) (*models.ProviderResource, error) {
	// Okta resources could be users, groups, apps, etc.
	// For now, return a basic implementation
	return &models.ProviderResource{
		Name:        resource,
		Description: fmt.Sprintf("Okta resource: %s", resource),
	}, nil
}

// ListResources lists available resources in Okta
func (p *oktaProvider) ListResources(ctx context.Context, filters ...string) ([]models.ProviderResource, error) {
	var resources []models.ProviderResource

	// List some common Okta resource types
	resourceTypes := []string{"users", "groups", "applications", "policies"}

	for _, rt := range resourceTypes {
		resources = append(resources, models.ProviderResource{
			Name:        rt,
			Description: fmt.Sprintf("Okta %s", rt),
		})
	}

	return resources, nil
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

// GetIdentity retrieves an identity (user) from Okta
func (p *oktaProvider) GetIdentity(ctx context.Context, identity string) (*models.Identity, error) {
	user, _, err := p.client.User.GetUser(ctx, identity)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Okta: %w", err)
	}

	email := ""
	name := ""
	if user.Profile != nil {
		if emailVal, ok := (*user.Profile)["email"].(string); ok {
			email = emailVal
		}
		if nameVal, ok := (*user.Profile)["firstName"].(string); ok {
			name = nameVal
		}
		if lastNameVal, ok := (*user.Profile)["lastName"].(string); ok {
			if name != "" {
				name = name + " " + lastNameVal
			} else {
				name = lastNameVal
			}
		}
	}

	return &models.Identity{
		ID:    user.Id,
		Label: email,
		User: &models.User{
			ID:     user.Id,
			Email:  email,
			Name:   name,
			Source: "okta",
		},
	}, nil
}

// ListIdentities lists all identities (users) from Okta
func (p *oktaProvider) ListIdentities(ctx context.Context, filters ...string) ([]models.Identity, error) {
	var identities []models.Identity

	// Build query parameters based on filters
	qp := query.NewQueryParams()
	if len(filters) > 0 {
		// Use the first filter as a search term
		qp = query.NewQueryParams(query.WithSearch(filters[0]))
	}

	users, _, err := p.client.User.ListUsers(ctx, qp)
	if err != nil {
		return nil, fmt.Errorf("failed to list users from Okta: %w", err)
	}

	for _, user := range users {
		email := ""
		name := ""
		if user.Profile != nil {
			if emailVal, ok := (*user.Profile)["email"].(string); ok {
				email = emailVal
			}
			if nameVal, ok := (*user.Profile)["firstName"].(string); ok {
				name = nameVal
			}
			if lastNameVal, ok := (*user.Profile)["lastName"].(string); ok {
				if name != "" {
					name = name + " " + lastNameVal
				} else {
					name = lastNameVal
				}
			}
		}

		identities = append(identities, models.Identity{
			ID:    user.Id,
			Label: email,
			User: &models.User{
				ID:     user.Id,
				Email:  email,
				Name:   name,
				Source: "okta",
			},
		})
	}

	return identities, nil
}

// RefreshIdentities refreshes the identity cache (if applicable)
func (p *oktaProvider) RefreshIdentities(ctx context.Context) error {
	// The Okta SDK has caching built in, so we don't need to do anything special here
	logrus.Info("Refreshing Okta identities cache")
	return nil
}
