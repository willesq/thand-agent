package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/thand-io/agent/internal/common"
)

type ProviderIdentitiesResponse struct {
	Version    string     `json:"version"`
	Provider   string     `json:"provider"`
	Identities []Identity `json:"identities"`
}

type ProviderRolesResponse struct {
	Version  string         `json:"version"`
	Provider string         `json:"provider"`
	Roles    []ProviderRole `json:"roles"`
}

type ProviderRole struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// Store the underlying provider-specific role object if needed
	Role any `json:"-"`
}

func (p *BaseProvider) SynchronizeRoles(
	ctx context.Context,
	req SynchronizeRolesRequest,
) (*SynchronizeRolesResponse, error) {
	return nil, ErrNotImplemented
}

func (p *BaseProvider) GetRole(ctx context.Context, role string) (*ProviderRole, error) {

	// If the role is a policy arn: arn:aws:iam::aws:policy/AdministratorAccess
	// Then parse the role and extract the policy name and convert it to a role
	role = strings.TrimPrefix(role, "arn:aws:iam::aws:policy/")
	role = strings.ToLower(role)

	// Fast map lookup
	if r, exists := p.rbac.rolesMap[role]; exists {
		return r, nil
	}

	return nil, fmt.Errorf("role not found")
}

func (p *BaseProvider) ListRoles(ctx context.Context, filters ...string) ([]ProviderRole, error) {
	// If no filters, return all roles
	if len(filters) == 0 {
		return p.rbac.roles, nil
	}

	// Check if search index is ready
	p.rbac.mu.RLock()
	rolesIndex := p.rbac.rolesIndex
	p.rbac.mu.RUnlock()

	if rolesIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, rolesIndex, func(a *search.DocumentMatch, b ProviderRole) bool {
			return strings.Compare(a.ID, b.Name) == 0
		}, p.rbac.roles, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []ProviderRole
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, role := range p.rbac.roles {
		// Check if any filter matches the role name
		if strings.Contains(strings.ToLower(role.Name), filterText) {
			filtered = append(filtered, role)
		}
	}

	return filtered, nil
}
