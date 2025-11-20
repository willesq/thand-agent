package okta

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

// RefreshIdentities fetches and caches user and group identities from Okta
func (p *oktaProvider) RefreshIdentities(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Refreshed Okta identities in %s", elapsed)
	}()

	var identities []models.Identity
	identitiesMap := make(map[string]*models.Identity)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	var userCount, groupCount int

	// Fetch users in parallel with groups
	wg.Add(2)

	// Fetch users
	go func() {
		defer wg.Done()
		userIdentities, err := p.fetchUserIdentities(ctx)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to fetch users: %w", err))
			mu.Unlock()
			return
		}

		mu.Lock()
		identities = append(identities, userIdentities...)
		userCount = len(userIdentities)
		mu.Unlock()
	}()

	// Fetch groups
	go func() {
		defer wg.Done()
		groupIdentities, err := p.fetchGroupIdentities(ctx)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to fetch groups: %w", err))
			mu.Unlock()
			return
		}

		mu.Lock()
		identities = append(identities, groupIdentities...)
		groupCount = len(groupIdentities)
		mu.Unlock()
	}()

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("errors refreshing identities: %v", errs)
	}

	// Build the identities map
	for i := range identities {
		identity := &identities[i]
		// Map by ID (lowercase)
		identitiesMap[strings.ToLower(identity.ID)] = identity
		// Map by label (lowercase)
		identitiesMap[strings.ToLower(identity.Label)] = identity

		// For users, also map by email
		if identity.User != nil && identity.User.Email != "" {
			identitiesMap[strings.ToLower(identity.User.Email)] = identity
		}
		// For groups, also map by name
		if identity.Group != nil && identity.Group.Name != "" {
			identitiesMap[strings.ToLower(identity.Group.Name)] = identity
		}
	}

	p.indexMu.Lock()
	p.identities = identities
	p.identitiesMap = identitiesMap
	p.indexMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"users":  userCount,
		"groups": groupCount,
		"total":  len(identities),
	}).Debug("Refreshed Okta identities")

	return nil
}

// fetchUserIdentities retrieves all users from Okta
func (p *oktaProvider) fetchUserIdentities(ctx context.Context) ([]models.Identity, error) {
	users, _, err := p.client.User.ListUsers(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var identities []models.Identity
	for _, user := range users {
		email := ""
		name := ""
		if user.Profile != nil {
			if emailVal, ok := (*user.Profile)["email"].(string); ok {
				email = emailVal
			}
			if nameVal, ok := (*user.Profile)["firstName"].(string); ok {
				name = nameVal
			}
			if lastNameVal, ok := (*user.Profile)["lastName"].(string); ok {
				if name != "" {
					name = name + " " + lastNameVal
				} else {
					name = lastNameVal
				}
			}
		}

		identity := models.Identity{
			ID:    email,
			Label: name,
			User: &models.User{
				ID:     user.Id,
				Email:  email,
				Name:   name,
				Source: "okta",
			},
		}

		identities = append(identities, identity)
	}

	return identities, nil
}

// fetchGroupIdentities retrieves all groups from Okta
func (p *oktaProvider) fetchGroupIdentities(ctx context.Context) ([]models.Identity, error) {
	groups, _, err := p.client.Group.ListGroups(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	var identities []models.Identity
	for _, group := range groups {
		identity := models.Identity{
			ID:    group.Id,
			Label: group.Profile.Name,
			Group: &models.Group{
				ID:   group.Id,
				Name: group.Profile.Name,
			},
		}

		identities = append(identities, identity)
	}

	return identities, nil
}

// GetIdentity retrieves a specific identity (user or group) from Okta
func (p *oktaProvider) GetIdentity(ctx context.Context, identity string) (*models.Identity, error) {
	// Try to get from cache first
	p.indexMu.RLock()
	identitiesMap := p.identitiesMap
	p.indexMu.RUnlock()

	if identitiesMap != nil {
		if id, exists := identitiesMap[strings.ToLower(identity)]; exists {
			return id, nil
		}
	}

	// If not in cache, try to fetch directly from API
	// First try as user
	user, _, err := p.client.User.GetUser(ctx, identity)
	if err == nil && user != nil {
		email := ""
		name := ""
		if user.Profile != nil {
			if emailVal, ok := (*user.Profile)["email"].(string); ok {
				email = emailVal
			}
			if nameVal, ok := (*user.Profile)["firstName"].(string); ok {
				name = nameVal
			}
			if lastNameVal, ok := (*user.Profile)["lastName"].(string); ok {
				if name != "" {
					name = name + " " + lastNameVal
				} else {
					name = lastNameVal
				}
			}
		}

		return &models.Identity{
			ID:    user.Id,
			Label: email,
			User: &models.User{
				ID:     user.Id,
				Email:  email,
				Name:   name,
				Source: "okta",
			},
		}, nil
	}

	// Try as group
	group, _, err := p.client.Group.GetGroup(ctx, identity)
	if err == nil && group != nil {
		return &models.Identity{
			ID:    group.Id,
			Label: group.Profile.Name,
			Group: &models.Group{
				ID:   group.Id,
				Name: group.Profile.Name,
			},
		}, nil
	}

	return nil, fmt.Errorf("identity not found: %s", identity)
}

// ListIdentities lists all identities (users and groups) from Okta
func (p *oktaProvider) ListIdentities(ctx context.Context, filters ...string) ([]models.Identity, error) {
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
			(identity.User != nil && strings.Contains(strings.ToLower(identity.User.Name), filterText)) ||
			(identity.Group != nil && strings.Contains(strings.ToLower(identity.Group.Name), filterText)) {
			filtered = append(filtered, identity)
		}
	}

	return filtered, nil
}
