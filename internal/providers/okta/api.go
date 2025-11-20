package okta

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// https://developer.okta.com/docs/api/openapi/okta-management/management/tag/RoleECustom/
// getCustomAdminRole retrieves a custom admin role by label from Okta
// Returns nil if the role doesn't exist
func (p *oktaProvider) getCustomAdminRole(ctx context.Context, roleLabel string) (*CustomAdminRoleResponse, error) {
	// Use the request executor to make a direct API call
	// GET /api/v1/iam/roles/{roleIdOrLabel}
	reqExecutor := p.client.CloneRequestExecutor()

	var role CustomAdminRoleResponse
	req, err := reqExecutor.NewRequest("GET", fmt.Sprintf("/api/v1/iam/roles/%s", roleLabel), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get request: %w", err)
	}

	_, err = reqExecutor.Do(ctx, req, &role)
	if err != nil {
		// Role doesn't exist or error occurred
		return nil, err
	}

	if len(role.ID) == 0 {
		return nil, fmt.Errorf("role not found")
	}

	logrus.WithFields(logrus.Fields{
		"role_label": roleLabel,
		"role_id":    role.ID,
	}).Debug("Retrieved custom admin role from Okta")

	return &role, nil
}

// createCustomAdminRole creates a custom admin role in Okta with the specified permissions
// If a role with the same label already exists, it returns the existing role
func (p *oktaProvider) createCustomAdminRole(ctx context.Context, role *models.Role) (*CustomAdminRoleResponse, error) {
	// Generate a unique role label based on the role name
	roleLabel := fmt.Sprintf("thand-%s", role.Name)

	// First, check if the role already exists
	existingRole, err := p.getCustomAdminRole(ctx, roleLabel)
	if err == nil && existingRole != nil {
		// Role already exists, return it
		logrus.WithFields(logrus.Fields{
			"role_label": roleLabel,
			"role_id":    existingRole.ID,
		}).Debug("Custom admin role already exists, reusing existing role")
		return existingRole, nil
	}

	// Role doesn't exist, create it
	// Prepare the custom role request payload
	customRolePayload := map[string]any{
		"label":       roleLabel,
		"description": role.Description,
		"permissions": role.Permissions.Allow,
	}

	// Use the request executor to make a direct API call
	// POST /api/v1/iam/roles
	reqExecutor := p.client.CloneRequestExecutor()

	var createdRole CustomAdminRoleResponse

	req, err := reqExecutor.NewRequest("POST", "/api/v1/iam/roles", customRolePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	_, err = reqExecutor.Do(ctx, req, &createdRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create custom admin role via API: %w", err)
	}

	// Extract the role ID from the response
	if createdRole.ID == "" {
		return nil, fmt.Errorf("failed to extract role ID from response")
	}

	logrus.WithFields(logrus.Fields{
		"role_label": roleLabel,
		"role_id":    createdRole.ID,
	}).Debug("Custom admin role created successfully")

	return &createdRole, nil
}

// deleteCustomAdminRole deletes a custom admin role from Okta
func (p *oktaProvider) deleteCustomAdminRole(ctx context.Context, roleId string) error {
	// Use the request executor to make a direct API call
	// DELETE /api/v1/iam/roles/{roleId}
	reqExecutor := p.client.CloneRequestExecutor()

	req, err := reqExecutor.NewRequest("DELETE", fmt.Sprintf("/api/v1/iam/roles/%s", roleId), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	_, err = reqExecutor.Do(ctx, req, nil)
	if err != nil {
		logrus.WithError(err).Warnf("Failed to delete custom admin role %s", roleId)
		return err
	}

	logrus.WithField("role_id", roleId).Debug("Custom admin role deleted successfully")
	return nil
}
