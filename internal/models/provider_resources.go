package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/thand-io/agent/internal/common"
)

func (p *BaseProvider) SynchronizeResources(ctx context.Context, req SynchronizeResourcesRequest) (*SynchronizeResourcesResponse, error) {
	return nil, ErrNotImplemented
}

func (p *BaseProvider) GetResource(ctx context.Context, resource string) (*ProviderResource, error) {

	// If the role is a policy arn: arn:aws:iam::aws:policy/AdministratorAccess
	// Then parse the role and extract the policy name and convert it to a role
	resource = strings.ToLower(resource)

	// Fast map lookup
	if r, exists := p.rbac.resourcesMap[resource]; exists {
		return r, nil
	}

	return nil, fmt.Errorf("resource not found")
}

func (p *BaseProvider) ListResources(ctx context.Context, filters ...string) ([]ProviderResource, error) {
	// If no filters, return all roles
	if len(filters) == 0 {
		return p.rbac.resources, nil
	}

	// Check if search index is ready
	p.rbac.mu.RLock()
	resourcesIndex := p.rbac.resourcesIndex
	p.rbac.mu.RUnlock()

	if resourcesIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, resourcesIndex, func(a *search.DocumentMatch, b ProviderResource) bool {
			return strings.Compare(a.ID, b.Name) == 0
		}, p.rbac.resources, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []ProviderResource
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, resource := range p.rbac.resources {
		// Check if any filter matches the resource name
		if strings.Contains(strings.ToLower(resource.Name), filterText) {
			filtered = append(filtered, resource)
		}
	}

	return filtered, nil
}
