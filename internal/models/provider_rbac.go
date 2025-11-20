package models

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
)

type AuthorizeRoleRequest struct {
	*RoleRequest
}

type AuthorizeRoleResponse struct {
	Metadata map[string]any `json:"metadata,omitempty"` // Any metadata returned from the provider
}

type RevokeRoleRequest struct {
	*RoleRequest
	AuthorizeRoleResponse *AuthorizeRoleResponse `json:"response,omitempty"`
}

type RevokeRoleResponse struct {
}

type ProviderRoleBasedAccessControl interface {

	// Role
	GetRole(ctx context.Context, role string) (*ProviderRole, error)
	ListRoles(ctx context.Context, filters ...string) ([]ProviderRole, error)

	// Permissions are individual accesses. Used as part of a role
	GetPermission(ctx context.Context, permission string) (*ProviderPermission, error)
	ListPermissions(ctx context.Context, filters ...string) ([]ProviderPermission, error)

	// Resources are things that permissions can be applied to
	GetResource(ctx context.Context, resource string) (*ProviderResource, error)
	ListResources(ctx context.Context, filters ...string) ([]ProviderResource, error)

	// Validate a role for a user
	ValidateRole(ctx context.Context, identity *Identity, role *Role) (map[string]any, error)

	// Authorize a role for a user (Bind a user to a role)
	AuthorizeRole(
		ctx context.Context,
		req *AuthorizeRoleRequest,
	) (
		*AuthorizeRoleResponse, // Return any custom metadata the provider wants to store
		error,
	)

	// Revoke a role from a user
	RevokeRole(
		ctx context.Context,
		req *RevokeRoleRequest, // Any metadata returned from AuthorizeRole
	) (*RevokeRoleResponse, error)

	// When applicable, get the URL to redirect the user to after post-authorize
	GetAuthorizedAccessUrl(
		ctx context.Context,
		req *AuthorizeRoleRequest,
		resp *AuthorizeRoleResponse,
	) string
}

/* Default implementations for role-based access control */

func (p *BaseProvider) GetRole(ctx context.Context, role string) (*ProviderRole, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement GetRole", p.GetProvider())
}

func (p *BaseProvider) ListRoles(ctx context.Context, filters ...string) ([]ProviderRole, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement ListRoles", p.GetProvider())
}

func (p *BaseProvider) GetPermission(ctx context.Context, permission string) (*ProviderPermission, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement GetPermission", p.GetProvider())
}

func (p *BaseProvider) ListPermissions(ctx context.Context, filters ...string) ([]ProviderPermission, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement ListPermissions", p.GetProvider())
}

func (p *BaseProvider) GetResource(ctx context.Context, resource string) (*ProviderResource, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement GetResource", p.GetProvider())
}

func (p *BaseProvider) ListResources(ctx context.Context, filters ...string) ([]ProviderResource, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement ListResources", p.GetProvider())
}

func (p *BaseProvider) AuthorizeRole(
	ctx context.Context,
	req *AuthorizeRoleRequest,
) (*AuthorizeRoleResponse, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement AuthorizeRole", p.GetProvider())
}

func (p *BaseProvider) RevokeRole(
	ctx context.Context,
	req *RevokeRoleRequest,
) (*RevokeRoleResponse, error) {
	// Default implementation does nothing
	return nil, fmt.Errorf("the provider '%s' does not implement RevokeRole", p.GetProvider())
}

func (p *BaseProvider) GetAuthorizedAccessUrl(
	ctx context.Context,
	req *AuthorizeRoleRequest,
	resp *AuthorizeRoleResponse,
) string {
	// Default implementation does nothing
	return ""
}

/* Default implementation for ValidateRole */
func (p *BaseProvider) ValidateRole(ctx context.Context, user *Identity, role *Role) (map[string]any, error) {
	// TODO this won't work as its the base provider. needs to call the actual provider
	// to validate the role
	return nil, ErrNotImplemented
}

// executeUserValidation triggers user approval workflow
func ValidateRole(
	providerCall ProviderImpl,
	elevateRequest ElevateRequestInternal,
) (map[string]any, error) {
	// Check the user has access to the required scopes etc

	identity := Identity{
		ID:    elevateRequest.User.GetIdentity(),
		Label: elevateRequest.User.GetName(),
		User:  elevateRequest.User,
	}

	res, err := providerCall.ValidateRole(
		context.Background(),
		&identity,
		elevateRequest.Role,
	)

	if err != nil {

		if !errors.Is(err, ErrNotImplemented) {
			return nil, fmt.Errorf("failed to validate role: %w", err)
		}

		logrus.Warn("Provider does not implement role validation, using default")
		err = validateRole(providerCall, &identity, elevateRequest.Role)

		if err != nil {

			logrus.WithError(err).Warn("Role validation failed")
			return nil, err

		}
	}

	return res, nil
}

func validateRole(provider ProviderImpl, _ *Identity, role *Role) error {
	if err := validateRoleNotEmpty(role); err != nil {
		return err
	}

	if err := validateRoleInheritance(provider, role); err != nil {
		return err
	}

	return validateRolePermissions(provider, role)
}

// validateRoleNotEmpty checks if the role has any permissions or inherits from other roles
func validateRoleNotEmpty(role *Role) error {
	if len(role.Permissions.Allow) == 0 &&
		len(role.Permissions.Deny) == 0 &&
		len(role.Resources.Allow) == 0 &&
		len(role.Resources.Deny) == 0 &&
		len(role.Groups.Allow) == 0 &&
		len(role.Groups.Deny) == 0 &&
		len(role.Inherits) == 0 {
		return fmt.Errorf("role %s has no permissions, inherits, groups or resources", role.Name)
	}
	return nil
}

// validateRoleInheritance validates that all inherited roles exist in the provider
func validateRoleInheritance(provider ProviderImpl, role *Role) error {
	if len(role.Inherits) == 0 {
		return nil
	}

	providerRoles, err := provider.ListRoles(context.TODO())
	if err != nil {
		return err
	}

	if len(providerRoles) == 0 {
		logrus.Warning("No roles found in provider")
		return nil
	}

	return validateInheritedRolesExist(provider, role, providerRoles)
}

// validateInheritedRolesExist checks that all inherited roles exist in the provider
func validateInheritedRolesExist(provider ProviderImpl, role *Role, providerRoles []ProviderRole) error {
	for _, inherit := range role.Inherits {
		if !strings.HasPrefix(inherit, fmt.Sprintf("%s:", provider.GetName())) {
			// This is a local role, skip validation
			continue
		}

		roleExists := slices.ContainsFunc(providerRoles, func(r ProviderRole) bool {
			return strings.Compare(r.Name, inherit) == 0
		})

		if !roleExists {
			return fmt.Errorf("role %s inherits from non-existent role %s", role.Name, inherit)
		}
	}
	return nil
}

// validateRolePermissions validates that role permissions exist in the provider
func validateRolePermissions(provider ProviderImpl, role *Role) error {
	if len(role.Permissions.Allow) == 0 && len(role.Permissions.Deny) == 0 {
		return nil
	}

	providerPermissions, err := provider.ListPermissions(context.TODO())
	if err != nil {
		return err
	}

	if len(providerPermissions) == 0 {
		logrus.Warning("No permissions found in provider")
		return nil
	}

	return validateRolePermissionLists(role, providerPermissions)
}

// validateRolePermissionLists validates both allow and deny permission lists
func validateRolePermissionLists(role *Role, providerPermissions []ProviderPermission) error {
	var err error

	role.Permissions.Allow, err = validatePermissions(providerPermissions, role.Permissions.Allow)
	if err != nil {
		return err
	}

	role.Permissions.Deny, err = validatePermissions(providerPermissions, role.Permissions.Deny)
	if err != nil {
		return err
	}

	return nil
}

func validatePermissions(providerPermissions []ProviderPermission, permissions []string) ([]string, error) {

	validatedPermissions := []string{}

	// Now lets check are remove permissions that don't exist
	for _, perm := range permissions {

		if strings.HasSuffix(perm, ":*") || strings.HasSuffix(perm, ".*") {
			// Permission ends with a wildcard. Lets expand this
			// out to include all permissions. As some IAMs do not
			// support wildcarding.
			validatedPermissions = append(validatedPermissions,
				expandPermissionsWildcard(providerPermissions, perm)...)

		} else if permission := getCondensedActions(perm); permission != nil {

			// If the last part is delimited by comma, e.g., k8s:pods:get,list,watch
			// lets use a more complex parsing with regex and then expand those
			// into individual permissions
			validatedPermissions = append(validatedPermissions, permission...)
			// We have a match, now expand it

		} else if !slices.ContainsFunc(providerPermissions, func(p ProviderPermission) bool {
			found := strings.Compare(p.Name, perm) == 0
			if found {
				validatedPermissions = append(validatedPermissions, p.Name)
			}
			return found
		}) {
			return nil, fmt.Errorf("the requested permission: %s was not found", perm)
		}
	}

	return validatedPermissions, nil
}

func expandPermissionsWildcard(providerPermissions []ProviderPermission, permission string) []string {

	if strings.HasSuffix(permission, ":*") {
		permission = strings.TrimSuffix(permission, ":*")
	} else if strings.HasSuffix(permission, ".*") {
		permission = strings.TrimSuffix(permission, ".*")
	}

	expandedPermissions := []string{}

	for _, providerPerm := range providerPermissions {
		if strings.HasPrefix(providerPerm.Name, permission) {
			expandedPermissions = append(expandedPermissions, providerPerm.Name)
		}
	}

	return expandedPermissions
}

/*
k8s:pods:get,list,watch
*/
func getCondensedActions(permission string) []string {

	// split on the last colon
	idx := strings.LastIndex(permission, ":")

	if idx == -1 {
		return nil
	}

	resource := permission[:idx]
	actions := permission[idx+1:]

	// Check if the second part contains a comma
	actionParts := strings.Split(actions, ",")

	if len(actionParts) == 0 {
		return nil
	}

	permissions := []string{}

	for _, action := range actionParts {
		permissions = append(permissions, fmt.Sprintf("%s:%s", resource, action))
	}

	return permissions
}
