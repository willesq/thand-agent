package gcp

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/api/cloudresourcemanager/v1"
	iam "google.golang.org/api/iam/v1"
)

// newThandCondition creates a new IAM condition used to tag bindings managed by thand
// We create a fresh copy each time to avoid shared state mutation
func newThandCondition() *cloudresourcemanager.Expr {
	return &cloudresourcemanager.Expr{
		Title:       "managed-by-thand",
		Description: "This binding is managed by thand",
		Expression:  "true", // Always evaluates to true, used as a tag
	}
}

// Authorize grants access for a user to a role
func (p *gcpProvider) AuthorizeRole(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
) (*models.AuthorizeRoleResponse, error) {

	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to authorize gcp role")
	}

	user := req.GetUser()
	role := req.GetRole()

	if len(role.Inherits) == 0 && len(role.Permissions.Allow) == 0 {
		return nil, fmt.Errorf("role %s has no inherits or permissions defined", role.Name)
	}

	config := p.GetConfig()
	projectId := p.GetProjectId()
	stage := config.GetStringWithDefault("stage", "GA")

	var assignedRoles []string

	// If inherits is specified, validate and bind predefined GCP roles
	if len(role.Inherits) > 0 {
		for _, inheritedRole := range role.Inherits {
			// Validate that the role is a valid GCP predefined role
			predefinedRole, err := p.GetRole(ctx, inheritedRole)
			if err != nil {
				return nil, fmt.Errorf("invalid GCP role '%s': %w", inheritedRole, err)
			}

			// Bind the user to the predefined role via IAM policy
			err = p.bindUserToPredefinedRole(projectId, user, predefinedRole.Name)
			if err != nil {
				return nil, temporal.NewApplicationErrorWithOptions(
					fmt.Sprintf("failed to bind user to role %s: %v", predefinedRole.Name, err),
					"GcpRoleBindingError",
					temporal.ApplicationErrorOptions{
						NextRetryDelay: 3 * time.Second,
						Cause:          err,
					},
				)
			}

			logrus.WithFields(logrus.Fields{
				"user_email": user.Email,
				"role":       predefinedRole.Name,
				"project_id": projectId,
			}).Info("Successfully bound user to predefined GCP role")

			assignedRoles = append(assignedRoles, predefinedRole.Name)
		}
	}

	// If permissions are specified, create a custom role with those permissions
	if len(role.Permissions.Allow) > 0 {
		// Check if the custom role already exists
		customRoleName := role.GetSnakeCaseName()
		existingRole, err := p.getRole(projectId, customRoleName)
		if err != nil {
			// If role doesn't exist, create it
			existingRole, err = p.createRole(
				projectId,
				customRoleName,
				role.GetName(),
				role.GetDescription(),
				stage,
				role.Permissions.Allow,
			)
			if err != nil {
				return nil, temporal.NewApplicationErrorWithOptions(
					fmt.Sprintf("failed to create custom role %s: %v", customRoleName, err),
					"GcpCustomRoleCreationError",
					temporal.ApplicationErrorOptions{
						NextRetryDelay: 3 * time.Second,
						Cause:          err,
					},
				)
			}

			logrus.WithFields(logrus.Fields{
				"role_name":   customRoleName,
				"project_id":  projectId,
				"permissions": role.Permissions.Allow,
			}).Info("Created custom GCP role")
		}

		// Bind the user to the custom role via IAM policy
		err = p.bindUserToRole(projectId, user, existingRole)
		if err != nil {
			return nil, temporal.NewApplicationErrorWithOptions(
				fmt.Sprintf("failed to bind user to custom role %s: %v", existingRole.Name, err),
				"GcpCustomRoleBindingError",
				temporal.ApplicationErrorOptions{
					NextRetryDelay: 3 * time.Second,
					Cause:          err,
				},
			)
		}

		logrus.WithFields(logrus.Fields{
			"user_email": user.Email,
			"role":       existingRole.Name,
			"project_id": projectId,
		}).Info("Successfully bound user to custom GCP role")

		assignedRoles = append(assignedRoles, existingRole.Name)
	}

	return &models.AuthorizeRoleResponse{
		UserId: user.Email,
		Roles:  assignedRoles,
	}, nil
}

// Revoke removes access for a user from a role
func (p *gcpProvider) RevokeRole(
	ctx context.Context,
	req *models.RevokeRoleRequest,
) (*models.RevokeRoleResponse, error) {

	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to revoke gcp role")
	}

	user := req.GetUser()
	projectId := p.GetProjectId()

	if req.AuthorizeRoleResponse == nil {
		return nil, fmt.Errorf("no authorize role response found for revocation")
	}

	// Get the roles that were assigned during authorization
	metadata := req.AuthorizeRoleResponse

	if len(metadata.Roles) == 0 {
		return nil, fmt.Errorf("no roles found in authorization response for revocation")
	}

	// Revoke each role that was assigned
	for _, roleName := range metadata.Roles {
		// Check if this is a predefined role (starts with "roles/") or custom role (starts with "projects/")
		if strings.HasPrefix(roleName, "roles/") {
			// Predefined role - unbind directly by role name
			err := p.unbindUserFromPredefinedRole(projectId, user, roleName)
			if err != nil {
				return nil, temporal.NewApplicationErrorWithOptions(
					fmt.Sprintf("failed to unbind user from predefined role %s: %v", roleName, err),
					"GcpRoleUnbindingError",
					temporal.ApplicationErrorOptions{
						NextRetryDelay: 3 * time.Second,
						Cause:          err,
					},
				)
			}

			logrus.WithFields(logrus.Fields{
				"user_email": user.Email,
				"role":       roleName,
				"project_id": projectId,
			}).Info("Successfully unbound user from predefined GCP role")
		} else {
			// Custom role - get the role object and unbind
			// Extract the role name from the full path (projects/{project}/roles/{roleName})
			parts := strings.Split(roleName, "/")
			if len(parts) < 1 || parts[len(parts)-1] == "" {
				return nil, fmt.Errorf("invalid custom role name format: %q", roleName)
			}
			customRoleName := parts[len(parts)-1]

			existingRole, err := p.getRole(projectId, customRoleName)
			if err != nil {
				return nil, temporal.NewApplicationErrorWithOptions(
					fmt.Sprintf("failed to get custom role %s: %v", customRoleName, err),
					"GcpGetRoleError",
					temporal.ApplicationErrorOptions{
						NextRetryDelay: 3 * time.Second,
						Cause:          err,
					},
				)
			}

			err = p.unbindUserFromRole(projectId, user, existingRole)
			if err != nil {
				return nil, temporal.NewApplicationErrorWithOptions(
					fmt.Sprintf("failed to unbind user from custom role %s: %v", roleName, err),
					"GcpCustomRoleUnbindingError",
					temporal.ApplicationErrorOptions{
						NextRetryDelay: 3 * time.Second,
						Cause:          err,
					},
				)
			}

			logrus.WithFields(logrus.Fields{
				"user_email": user.Email,
				"role":       roleName,
				"project_id": projectId,
			}).Info("Successfully unbound user from custom GCP role")
		}
	}

	return &models.RevokeRoleResponse{}, nil
}

func (p *gcpProvider) GetAuthorizedAccessUrl(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
	resp *models.AuthorizeRoleResponse,
) string {

	return p.GetConfig().GetStringWithDefault(
		"sso_start_url", "https://console.cloud.google.com/")
}

// createRole creates a custom role.
func (p *gcpProvider) createRole(projectID, name, title, description, stage string, permissions []string) (*iam.Role, error) {

	service := p.GetIamClient()

	request := &iam.CreateRoleRequest{
		Role: &iam.Role{
			Title:               title,
			Description:         description,
			IncludedPermissions: permissions,
			Stage:               stage,
		},
		RoleId: name,
	}
	role, err := service.Projects.Roles.Create("projects/"+projectID, request).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.Roles.Create: %w", err)
	}
	return role, nil
}

func (p *gcpProvider) getRole(projectID, roleName string) (*iam.Role, error) {
	service := p.GetIamClient()

	role, err := service.Projects.Roles.Get("projects/" + projectID + "/roles/" + roleName).Do()
	if err != nil {
		// Return nil role and error if role doesn't exist
		return nil, err
	}
	return role, nil
}

// bindUserToPredefinedRole binds a user to a predefined GCP role (e.g., roles/viewer)
func (p *gcpProvider) bindUserToPredefinedRole(projectID string, user *models.User, roleName string) error {
	crmService := p.crmClient

	// Get current IAM policy - request version 3 to support conditions
	policy, err := crmService.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{
		Options: &cloudresourcemanager.GetPolicyOptions{
			RequestedPolicyVersion: 3,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %w", err)
	}

	// Ensure policy version is 3 for conditions support
	policy.Version = 3

	// Create member string based on user type
	var member string
	if len(user.Email) > 0 {
		member = "user:" + user.Email
	} else {
		return fmt.Errorf("user email is required for GCP IAM binding")
	}

	// Check if binding already exists with our thand condition
	bindingExists := false
	for _, binding := range policy.Bindings {
		if binding.Role == roleName && isThandManagedBinding(binding) {
			if slices.Contains(binding.Members, member) {
				bindingExists = true
			}
			if !bindingExists {
				// Add member to existing thand-managed binding
				binding.Members = append(binding.Members, member)
				bindingExists = true
			}
			break
		}
	}

	// If no binding exists for this role with our condition, create a new one
	if !bindingExists {
		newBinding := &cloudresourcemanager.Binding{
			Role:      roleName,
			Members:   []string{member},
			Condition: newThandCondition(),
		}
		policy.Bindings = append(policy.Bindings, newBinding)
	}

	// Set the updated IAM policy
	_, err = crmService.Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %w", err)
	}

	return nil
}

// unbindUserFromPredefinedRole removes a user from a predefined GCP role
func (p *gcpProvider) unbindUserFromPredefinedRole(projectID string, user *models.User, roleName string) error {
	crmService := p.crmClient

	// Get current IAM policy - request version 3 to support conditions
	policy, err := crmService.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{
		Options: &cloudresourcemanager.GetPolicyOptions{
			RequestedPolicyVersion: 3,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %w", err)
	}

	// Ensure policy version is 3 for conditions support
	policy.Version = 3

	// Create member string based on user type
	var member string
	if len(user.Email) > 0 {
		member = "user:" + user.Email
	} else {
		return fmt.Errorf("user email is required for GCP IAM binding")
	}

	// Find and remove the user from the thand-managed role binding
	bindingFound := false
	for i, binding := range policy.Bindings {
		if binding.Role == roleName && isThandManagedBinding(binding) {
			bindingFound = true
			// Find and remove the member from this binding
			for j, bindingMember := range binding.Members {
				if bindingMember == member {
					// Remove the member from the slice
					binding.Members = append(binding.Members[:j], binding.Members[j+1:]...)
					break
				}
			}
			// If the binding has no members left, remove the entire binding
			if len(binding.Members) == 0 {
				policy.Bindings = append(policy.Bindings[:i], policy.Bindings[i+1:]...)
			}
			break
		}
	}

	// If no binding was found for this role, the user wasn't bound to it
	if !bindingFound {
		return fmt.Errorf("thand-managed role binding not found for role %s", roleName)
	}

	// Set the updated IAM policy
	_, err = crmService.Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %w", err)
	}

	return nil
}

// isThandManagedBinding checks if a binding has the thand condition tag
func isThandManagedBinding(binding *cloudresourcemanager.Binding) bool {
	return binding.Condition != nil && binding.Condition.Title == "managed-by-thand"
}

func (p *gcpProvider) bindUserToRole(projectID string, user *models.User, iamRole *iam.Role) error {
	crmService := p.crmClient

	// Get current IAM policy - request version 3 to support conditions
	policy, err := crmService.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{
		Options: &cloudresourcemanager.GetPolicyOptions{
			RequestedPolicyVersion: 3,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %w", err)
	}

	// Ensure policy version is 3 for conditions support
	policy.Version = 3

	// Create member string based on user type
	var member string
	if len(user.Email) > 0 {
		member = "user:" + user.Email
	} else {
		return fmt.Errorf("user email is required for GCP IAM binding")
	}

	// Check if binding already exists with our thand condition
	bindingExists := false
	for _, binding := range policy.Bindings {
		if binding.Role == iamRole.Name && isThandManagedBinding(binding) {
			if slices.Contains(binding.Members, member) {
				bindingExists = true
			}
			if !bindingExists {
				// Add member to existing thand-managed binding
				binding.Members = append(binding.Members, member)
				bindingExists = true
			}
			break
		}
	}

	// If no binding exists for this role with our condition, create a new one
	if !bindingExists {
		newBinding := &cloudresourcemanager.Binding{
			Role:      iamRole.Name,
			Members:   []string{member},
			Condition: newThandCondition(),
		}
		policy.Bindings = append(policy.Bindings, newBinding)
	}

	// Set the updated IAM policy
	_, err = crmService.Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %w", err)
	}

	return nil
}

func (p *gcpProvider) unbindUserFromRole(projectID string, user *models.User, iamRole *iam.Role) error {
	crmService := p.crmClient

	// Get current IAM policy - request version 3 to support conditions
	policy, err := crmService.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{
		Options: &cloudresourcemanager.GetPolicyOptions{
			RequestedPolicyVersion: 3,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to get IAM policy: %w", err)
	}

	// Ensure policy version is 3 for conditions support
	policy.Version = 3

	// Create member string based on user type
	var member string
	if len(user.Email) > 0 {
		member = "user:" + user.Email
	} else {
		return fmt.Errorf("user email is required for GCP IAM binding")
	}

	// Find and remove the user from the thand-managed role binding
	bindingFound := false
	for i, binding := range policy.Bindings {
		if binding.Role == iamRole.Name && isThandManagedBinding(binding) {
			bindingFound = true
			// Find and remove the member from this binding
			for j, bindingMember := range binding.Members {
				if bindingMember == member {
					// Remove the member from the slice
					binding.Members = append(binding.Members[:j], binding.Members[j+1:]...)
					break
				}
			}
			// If the binding has no members left, remove the entire binding
			if len(binding.Members) == 0 {
				policy.Bindings = append(policy.Bindings[:i], policy.Bindings[i+1:]...)
			}
			break
		}
	}

	// If no binding was found for this role, the user wasn't bound to it
	if !bindingFound {
		return fmt.Errorf("thand-managed role binding not found for role %s", iamRole.Name)
	}

	// Set the updated IAM policy
	_, err = crmService.Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %w", err)
	}

	return nil
}
