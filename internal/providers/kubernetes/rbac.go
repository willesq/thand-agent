package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/thand-io/agent/internal/models"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AuthorizeRole grants access for a user to a role
func (p *kubernetesProvider) AuthorizeRole(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
) (*models.AuthorizeRoleResponse, error) {

	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to authorize kubernetes role")
	}

	user := req.GetUser()
	role := req.GetRole()

	// Determine scope based on role configuration
	namespace := p.getNamespaceFromRole(role)

	if namespace != "" {
		// Create namespaced Role and RoleBinding
		return p.authorizeNamespacedRole(ctx, user, role, namespace)
	} else {
		// Create cluster-wide ClusterRole and ClusterRoleBinding
		return p.authorizeClusterRole(ctx, user, role)
	}
}

// RevokeRole removes access for a user from a role
func (p *kubernetesProvider) RevokeRole(
	ctx context.Context,
	req *models.RevokeRoleRequest,
) (*models.RevokeRoleResponse, error) {

	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to revoke kubernetes role")
	}

	user := req.GetUser()
	role := req.GetRole()

	namespace := p.getNamespaceFromRole(role)

	if namespace != "" {
		return p.revokeNamespacedRole(ctx, user, role, namespace)
	} else {
		return p.revokeClusterRole(ctx, user, role)
	}
}

// authorizeNamespacedRole creates Role and RoleBinding for namespace-scoped access
func (p *kubernetesProvider) authorizeNamespacedRole(
	ctx context.Context,
	user *models.User,
	role *models.Role,
	namespace string,
) (*models.AuthorizeRoleResponse, error) {

	client := p.GetClient()
	roleName := role.GetSnakeCaseName()

	// Create or update Role
	k8sRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
			Labels: map[string]string{
				"thand.io/managed": "true",
				"thand.io/role":    roleName,
			},
		},
		Rules: p.convertPermissionsToRules(role.Permissions.Allow),
	}

	_, err := client.RbacV1().Roles(namespace).Create(ctx, k8sRole, metav1.CreateOptions{})
	if err != nil {
		// If role exists, update it
		if strings.Contains(err.Error(), "already exists") {
			_, err = client.RbacV1().Roles(namespace).Update(ctx, k8sRole, metav1.UpdateOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to update role: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create role: %w", err)
		}
	}

	// Create RoleBinding
	bindingName := fmt.Sprintf("%s-%s", roleName, p.sanitizeUserIdentifier(user))
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
			Labels: map[string]string{
				"thand.io/managed": "true",
				"thand.io/role":    roleName,
				"thand.io/user":    p.sanitizeUserIdentifier(user),
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "User",
				Name: p.getUserIdentifier(user),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = client.RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create role binding: %w", err)
	}

	return &models.AuthorizeRoleResponse{
		Metadata: map[string]any{
			"roleName":    roleName,
			"bindingName": bindingName,
			"namespace":   namespace,
			"scope":       "namespaced",
		},
	}, nil
}

// authorizeClusterRole creates ClusterRole and ClusterRoleBinding for cluster-wide access
func (p *kubernetesProvider) authorizeClusterRole(
	ctx context.Context,
	user *models.User,
	role *models.Role,
) (*models.AuthorizeRoleResponse, error) {

	client := p.GetClient()
	roleName := role.GetSnakeCaseName()

	// Create or update ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
			Labels: map[string]string{
				"thand.io/managed": "true",
				"thand.io/role":    roleName,
			},
		},
		Rules: p.convertPermissionsToRules(role.Permissions.Allow),
	}

	_, err := client.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			_, err = client.RbacV1().ClusterRoles().Update(ctx, clusterRole, metav1.UpdateOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to update cluster role: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create cluster role: %w", err)
		}
	}

	// Create ClusterRoleBinding
	bindingName := fmt.Sprintf("%s-%s", roleName, p.sanitizeUserIdentifier(user))
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingName,
			Labels: map[string]string{
				"thand.io/managed": "true",
				"thand.io/role":    roleName,
				"thand.io/user":    p.sanitizeUserIdentifier(user),
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "User",
				Name: p.getUserIdentifier(user),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster role binding: %w", err)
	}

	return &models.AuthorizeRoleResponse{
		Metadata: map[string]any{
			"roleName":    roleName,
			"bindingName": bindingName,
			"scope":       "cluster",
		},
	}, nil
}

// convertPermissionsToRules converts thand permissions to Kubernetes RBAC rules
func (p *kubernetesProvider) convertPermissionsToRules(permissions []string) []rbacv1.PolicyRule {
	var rules []rbacv1.PolicyRule

	// Group permissions by API group and resource
	ruleMap := make(map[string]*rbacv1.PolicyRule)

	for _, permission := range permissions {
		rule := p.parsePermission(permission)
		if rule != nil {
			key := fmt.Sprintf("%s:%s", strings.Join(rule.APIGroups, ","), strings.Join(rule.Resources, ","))
			if existingRule, exists := ruleMap[key]; exists {
				// Merge verbs
				existingRule.Verbs = append(existingRule.Verbs, rule.Verbs...)
				existingRule.Verbs = p.deduplicateSlice(existingRule.Verbs)
			} else {
				ruleMap[key] = rule
			}
		}
	}

	// Convert map back to slice
	for _, rule := range ruleMap {
		rules = append(rules, *rule)
	}

	return rules
}

// parsePermission converts a permission string to PolicyRule
func (p *kubernetesProvider) parsePermission(permission string) *rbacv1.PolicyRule {
	// Expected formats:
	// "k8s:pods:get" -> get pods in core API group
	// "k8s:apps/deployments:list,watch" -> list,watch deployments in apps API group
	// "k8s:*/secrets:get,create" -> get,create secrets in all namespaces

	parts := strings.Split(permission, ":")
	if len(parts) != 3 {
		return nil // Invalid format
	}

	apiGroup := ""
	resource := parts[1]
	verbs := strings.Split(parts[2], ",")

	// Parse API group and resource
	if strings.Contains(resource, "/") {
		groupResource := strings.Split(resource, "/")
		if len(groupResource) == 2 {
			apiGroup = groupResource[0]
			resource = groupResource[1]
		}
	}

	rule := &rbacv1.PolicyRule{
		APIGroups: []string{apiGroup},
		Resources: []string{resource},
		Verbs:     verbs,
	}

	return rule
}

// Security helper functions
func (p *kubernetesProvider) getUserIdentifier(user *models.User) string {
	// Prefer email for OIDC integration, fallback to username
	if user.Email != "" {
		return user.Email
	}
	return user.Username
}

func (p *kubernetesProvider) sanitizeUserIdentifier(user *models.User) string {
	identifier := p.getUserIdentifier(user)
	// Replace invalid characters for Kubernetes resource names
	identifier = strings.ReplaceAll(identifier, "@", "-at-")
	identifier = strings.ReplaceAll(identifier, ".", "-")
	identifier = strings.ToLower(identifier)
	return identifier
}

func (p *kubernetesProvider) getNamespaceFromRole(role *models.Role) string {
	// Check if role has namespace-specific resources
	for _, resource := range role.Resources.Allow {
		if strings.Contains(resource, "namespace:") {
			parts := strings.Split(resource, ":")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "" // Empty string means cluster-wide
}

func (p *kubernetesProvider) deduplicateSlice(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Revocation functions
func (p *kubernetesProvider) revokeNamespacedRole(
	ctx context.Context,
	user *models.User,
	role *models.Role,
	namespace string,
) (*models.RevokeRoleResponse, error) {

	client := p.GetClient()
	bindingName := fmt.Sprintf("%s-%s", role.GetSnakeCaseName(), p.sanitizeUserIdentifier(user))

	// Check if RoleBinding exists before attempting to delete
	_, err := client.RbacV1().RoleBindings(namespace).Get(ctx, bindingName, metav1.GetOptions{})
	if err != nil {
		// If the binding doesn't exist, consider it already revoked
		if strings.Contains(err.Error(), "not found") {
			return &models.RevokeRoleResponse{}, nil
		}
		return nil, fmt.Errorf("failed to check role binding existence: %w", err)
	}

	// Delete RoleBinding
	err = client.RbacV1().RoleBindings(namespace).Delete(ctx, bindingName, metav1.DeleteOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to delete role binding: %w", err)
	}

	return &models.RevokeRoleResponse{}, nil
}

func (p *kubernetesProvider) revokeClusterRole(
	ctx context.Context,
	user *models.User,
	role *models.Role,
) (*models.RevokeRoleResponse, error) {

	client := p.GetClient()
	bindingName := fmt.Sprintf("%s-%s", role.GetSnakeCaseName(), p.sanitizeUserIdentifier(user))

	// Delete ClusterRoleBinding
	err := client.RbacV1().ClusterRoleBindings().Delete(ctx, bindingName, metav1.DeleteOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to delete cluster role binding: %w", err)
	}

	return &models.RevokeRoleResponse{}, nil
}
