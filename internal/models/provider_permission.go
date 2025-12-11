package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/thand-io/agent/internal/common"
)

type ProviderPermissionsResponse struct {
	Version     string               `json:"version"`
	Provider    string               `json:"provider"`
	Permissions []ProviderPermission `json:"permissions"`
}

type ProviderPermission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// Store the underlying provider-specific permission object if needed
	Permission any `json:"-"`
}

func (p *BaseProvider) GetPermission(ctx context.Context, permission string) (*ProviderPermission, error) {

	permission = strings.ToLower(permission)
	// Fast map lookup
	if perm, exists := p.rbac.permissionsMap[permission]; exists {
		return perm, nil
	}
	return nil, fmt.Errorf("permission not found")
}

func (p *BaseProvider) ListPermissions(ctx context.Context, filters ...string) ([]ProviderPermission, error) {
	// If no filters, return all permissions
	if len(filters) == 0 {
		return p.rbac.permissions, nil
	}

	// Check if search index is ready
	p.rbac.mu.RLock()
	permissionsIndex := p.rbac.permissionsIndex
	p.rbac.mu.RUnlock()

	if permissionsIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, permissionsIndex, func(a *search.DocumentMatch, b ProviderPermission) bool {
			return strings.EqualFold(a.ID, b.Name)
		}, p.rbac.permissions, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []ProviderPermission
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, perm := range p.rbac.permissions {
		// Check if any filter matches the permission name or description
		if strings.Contains(strings.ToLower(perm.Name), filterText) ||
			strings.Contains(strings.ToLower(perm.Description), filterText) {
			filtered = append(filtered, perm)
		}
	}

	return filtered, nil
}

func (p *BaseProvider) SynchronizePermissions(
	ctx context.Context,
	req SynchronizePermissionsRequest,
) (*SynchronizePermissionsResponse, error) {
	return nil, ErrNotImplemented
}
