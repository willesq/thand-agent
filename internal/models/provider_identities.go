package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus"
)

type ProviderIdentities interface {
	GetIdentity(ctx context.Context, identity string) (*Identity, error)
	ListIdentities(ctx context.Context, filters ...string) ([]Identity, error)

	SetIdentities(identities []Identity)

	// Some APIs support identities, users, groups service accoutns etc.
	SynchronizeIdentities(ctx context.Context, req SynchronizeIdentitiesRequest) (*SynchronizeIdentitiesResponse, error)
	// Some require more granular user synchronization
	SynchronizeUsers(ctx context.Context, req SynchronizeUsersRequest) (*SynchronizeUsersResponse, error)
	// Others require group synchronization
	SynchronizeGroups(ctx context.Context, req SynchronizeGroupsRequest) (*SynchronizeGroupsResponse, error)
}

func (p *BaseProvider) SynchronizeIdentities(ctx context.Context, req SynchronizeIdentitiesRequest) (*SynchronizeIdentitiesResponse, error) {
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

	if p.identity == nil || !p.HasCapability(
		ProviderCapabilityIdentities,
	) {
		logrus.Warningln("provider has no identities")
		return nil, fmt.Errorf("provider has no identities")
	}

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

	if p.identity == nil || !p.HasCapability(
		ProviderCapabilityIdentities,
	) {
		logrus.Warningln("provider has no identities")
		return nil, fmt.Errorf("provider has no identities")
	}

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

func (p *BaseProvider) buildIdentitiyIndices() error {
	// Placeholder for building indices
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Built identity search indices in %s", elapsed)
	}()

	identityMapping := bleve.NewIndexMapping()
	identityIndex, err := bleve.NewMemOnly(identityMapping)
	if err != nil {
		return fmt.Errorf("failed to create identity search index: %v", err)
	}

	// Index identities
	for _, identity := range p.identity.identities {
		if err := identityIndex.Index(identity.ID, identity); err != nil {
			return fmt.Errorf("failed to index identity %s: %v", identity.ID, err)
		}
	}

	p.identity.mu.Lock()
	p.identity.identitiesIndex = identityIndex
	p.identity.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"identities": len(p.identity.identities),
	}).Debug("Identity search indices ready")

	return nil
}
