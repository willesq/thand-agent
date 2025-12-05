package models

import (
	"context"
	"fmt"
	"strings"
)

type ProviderIdentities interface {
	GetIdentity(ctx context.Context, identity string) (*Identity, error)
	ListIdentities(ctx context.Context, filters ...string) ([]Identity, error)

	// Some APIs support identities, users, groups service accoutns etc.
	SynchronizeIdentities(ctx context.Context, req SynchronizeUsersRequest) (*SynchronizeUsersResponse, error)
	// Some require more granular user synchronization
	SynchronizeUsers(ctx context.Context, req SynchronizeUsersRequest) (*SynchronizeUsersResponse, error)
	// Others require group synchronization
	SynchronizeGroups(ctx context.Context, req SynchronizeGroupsRequest) (*SynchronizeGroupsResponse, error)
}

func (p *BaseProvider) SynchronizeIdentities(ctx context.Context, req SynchronizeUsersRequest) (*SynchronizeUsersResponse, error) {
	return nil, ErrNotImplemented
}

func (p *BaseProvider) SynchronizeUsers(ctx context.Context, req SynchronizeUsersRequest) (*SynchronizeUsersResponse, error) {
	return nil, ErrNotImplemented
}

func (p *BaseProvider) SynchronizeGroups(ctx context.Context, req SynchronizeGroupsRequest) (*SynchronizeGroupsResponse, error) {
	return nil, ErrNotImplemented
}

// GetIdentity retrieves a specific identity (user or group) from GCP
func (p *BaseProvider) GetIdentity(ctx context.Context, identity string) (*Identity, error) {
	// Try to get from cache first
	p.identity.mu.RLock()
	identitiesMap := p.identity.identitiesMap
	p.identity.mu.RUnlock()

	if identitiesMap != nil {
		if id, exists := identitiesMap[strings.ToLower(identity)]; exists {
			return id, nil
		}
	}

	p.identity.mu.RLock()
	defer p.identity.mu.RUnlock()

	if id, exists := p.identity.identitiesMap[strings.ToLower(identity)]; exists {
		return id, nil
	}

	return nil, fmt.Errorf("identity not found: %s", identity)
}

// ListIdentities lists all identities (users and groups) from GCP IAM
func (p *BaseProvider) ListIdentities(ctx context.Context, filters ...string) ([]Identity, error) {

	p.identity.mu.RLock()
	identities := p.identity.identities
	p.identity.mu.RUnlock()

	// If no filters, return all identities
	if len(filters) == 0 {
		return identities, nil
	}

	// Apply filters
	var filtered []Identity
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, identity := range identities {
		// Check if any filter matches the identity
		if strings.Contains(strings.ToLower(identity.Label), filterText) ||
			strings.Contains(strings.ToLower(identity.ID), filterText) ||
			(identity.User != nil && strings.Contains(strings.ToLower(identity.User.Email), filterText)) ||
			(identity.User != nil && strings.Contains(strings.ToLower(identity.User.Name), filterText)) ||
			(identity.Group != nil && strings.Contains(strings.ToLower(identity.Group.Name), filterText)) ||
			(identity.Group != nil && strings.Contains(strings.ToLower(identity.Group.Email), filterText)) {
			filtered = append(filtered, identity)
		}
	}

	return filtered, nil
}
