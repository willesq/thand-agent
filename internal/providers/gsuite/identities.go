package gsuite

import (
	"context"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// refreshIdentities fetches all users and groups from GSuite and caches them
func (p *gsuiteProvider) refreshIdentities() error {
	logrus.Info("Refreshing GSuite identities...")

	// Clear existing cache and slice
	p.identityCache = make(map[string]*models.Identity)
	p.identities = []models.Identity{}

	// Fetch users
	if err := p.fetchUsers(); err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	// Fetch groups
	if err := p.fetchGroups(); err != nil {
		return fmt.Errorf("failed to fetch groups: %w", err)
	}

	// Index identities for search
	if err := p.indexIdentities(); err != nil {
		return fmt.Errorf("failed to index identities: %w", err)
	}

	logrus.WithField("count", len(p.identityCache)).Info("GSuite identities refreshed")
	return nil
}

// fetchUsers retrieves all users from GSuite
func (p *gsuiteProvider) fetchUsers() error {
	pageToken := ""
	userCount := 0

	for {
		call := p.adminService.Users.List().
			Domain(p.domain).
			MaxResults(500).
			OrderBy("email")

		if len(pageToken) > 0 {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		for _, user := range resp.Users {
			identity := &models.Identity{
				Name: user.PrimaryEmail,
				User: &models.User{
					ID:       user.Id,
					Username: strings.Split(user.PrimaryEmail, "@")[0],
					Email:    user.PrimaryEmail,
					Name:     user.Name.FullName,
					Source:   "gsuite",
				},
			}

			p.identityCache[user.Name.DisplayName] = identity
			p.identityCache[user.PrimaryEmail] = identity  // Also cache by email
			p.identities = append(p.identities, *identity) // Add to slice for search
			userCount++
		}

		if len(resp.NextPageToken) == 0 {
			break
		}
		pageToken = resp.NextPageToken
	}

	logrus.WithField("count", userCount).Info("Fetched GSuite users")
	return nil
}

// fetchGroups retrieves all groups from GSuite
func (p *gsuiteProvider) fetchGroups() error {
	pageToken := ""
	groupCount := 0

	for {
		call := p.adminService.Groups.List().
			Domain(p.domain).
			MaxResults(200).
			OrderBy("email")

		if len(pageToken) > 0 {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("failed to list groups: %w", err)
		}

		for _, group := range resp.Groups {
			identity := &models.Identity{
				Name: group.Name,
				Group: &models.Group{
					ID:    group.Id,
					Name:  group.Name,
					Email: group.Email,
				},
			}

			p.identityCache[group.Name] = identity
			p.identityCache[group.Email] = identity        // Also cache by email
			p.identities = append(p.identities, *identity) // Add to slice for search
			groupCount++
		}

		if len(resp.NextPageToken) == 0 {
			break
		}
		pageToken = resp.NextPageToken
	}

	logrus.WithField("count", groupCount).Info("Fetched GSuite groups")
	return nil
}

// indexIdentities indexes all identities in the Bleve search index
func (p *gsuiteProvider) indexIdentities() error {
	batch := p.identitiesIndex.NewBatch()

	for _, identity := range p.identities {
		err := batch.Index(identity.Name, identity)
		if err != nil {
			return fmt.Errorf("failed to index identity %s: %w", identity.Name, err)
		}
	}

	return p.identitiesIndex.Batch(batch)
}

// GetIdentity retrieves a specific identity by ID or email
func (p *gsuiteProvider) GetIdentity(ctx context.Context, identity string) (*models.Identity, error) {
	if cachedIdentity, exists := p.identityCache[identity]; exists {
		// Return a copy to prevent modification
		identityCopy := *cachedIdentity
		if cachedIdentity.User != nil {
			userCopy := *cachedIdentity.User
			identityCopy.User = &userCopy
		}
		if cachedIdentity.Group != nil {
			groupCopy := *cachedIdentity.Group
			identityCopy.Group = &groupCopy
		}
		return &identityCopy, nil
	}

	return nil, fmt.Errorf("identity %s not found", identity)
}

// ListIdentities returns all cached identities with optional filtering
func (p *gsuiteProvider) ListIdentities(ctx context.Context, filters ...string) ([]models.Identity, error) {
	return common.BleveListSearch(ctx, p.identitiesIndex, func(a *search.DocumentMatch, b models.Identity) bool {
		return strings.Compare(a.ID, b.Name) == 0
	}, p.identities, filters...)
}
