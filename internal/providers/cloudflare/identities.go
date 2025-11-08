package cloudflare

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// RefreshIdentities fetches and caches user identities from Cloudflare
func (p *cloudflareProvider) RefreshIdentities(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Refreshed Cloudflare identities in %s", elapsed)
	}()

	accountID := p.GetAccountID()

	// List all account members
	members, _, err := p.client.AccountMembers(ctx, accountID, cloudflare.PaginationOptions{})
	if err != nil {
		return fmt.Errorf("failed to list account members: %w", err)
	}

	var identities []models.Identity
	identitiesMap := make(map[string]*models.Identity)

	for _, member := range members {
		identity := models.Identity{
			ID:    member.ID,
			Label: member.User.Email,
			User: &models.User{
				ID:    member.User.ID,
				Email: member.User.Email,
				Name:  fmt.Sprintf("%s %s", member.User.FirstName, member.User.LastName),
			},
		}

		identities = append(identities, identity)
		identitiesMap[strings.ToLower(member.ID)] = &identities[len(identities)-1]
		identitiesMap[strings.ToLower(member.User.Email)] = &identities[len(identities)-1]
	}

	// Update the provider's identities cache
	p.indexMu.Lock()
	p.identities = identities
	p.identitiesMap = identitiesMap
	p.indexMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"identities": len(identities),
	}).Debug("Refreshed Cloudflare identities")

	return nil
}

// GetIdentity retrieves a specific identity by ID or email
func (p *cloudflareProvider) GetIdentity(ctx context.Context, identity string) (*models.Identity, error) {
	// Try to get from cache first
	p.indexMu.RLock()
	identitiesMap := p.identitiesMap
	p.indexMu.RUnlock()

	if identitiesMap != nil {
		if id, exists := identitiesMap[strings.ToLower(identity)]; exists {
			return id, nil
		}
	}

	// If not in cache, refresh and try again
	err := p.RefreshIdentities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh identities: %w", err)
	}

	p.indexMu.RLock()
	identitiesMap = p.identitiesMap
	p.indexMu.RUnlock()

	if id, exists := identitiesMap[strings.ToLower(identity)]; exists {
		return id, nil
	}

	return nil, fmt.Errorf("identity not found: %s", identity)
}

// ListIdentities lists all identities, optionally filtered
func (p *cloudflareProvider) ListIdentities(ctx context.Context, filters ...string) ([]models.Identity, error) {
	// Ensure we have fresh data
	err := p.RefreshIdentities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh identities: %w", err)
	}

	p.indexMu.RLock()
	identities := p.identities
	p.indexMu.RUnlock()

	// If no filters, return all identities
	if len(filters) == 0 {
		return identities, nil
	}

	// Apply filters
	var filtered []models.Identity
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, identity := range identities {
		// Check if any filter matches the identity
		if strings.Contains(strings.ToLower(identity.Label), filterText) ||
			strings.Contains(strings.ToLower(identity.ID), filterText) ||
			(identity.User != nil && strings.Contains(strings.ToLower(identity.User.Email), filterText)) ||
			(identity.User != nil && strings.Contains(strings.ToLower(identity.User.Name), filterText)) {
			filtered = append(filtered, identity)
		}
	}

	return filtered, nil
}
