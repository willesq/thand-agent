package config

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

const (
	IdentityTypeUser  IdentityType = "user"
	IdentityTypeGroup IdentityType = "group"
	IdentityTypeAll   IdentityType = "all"
)

type IdentityType string

func (c *Config) GetIdentitiesWithFilter(user *models.User, identityType IdentityType, filter ...string) ([]models.Identity, error) {

	// Find providers with identity capabilities
	providerMap := c.GetProvidersByCapabilityWithUser(user, models.ProviderCapabilityIdentities)

	var identities []models.Identity

	// If no identity providers found, return just the current user
	if len(providerMap) == 0 {
		// Apply filter to current user if specified
		if len(filter) > 0 {
			userFields := []string{strings.ToLower(user.Name), strings.ToLower(user.Email)}
			matchesFilter := slices.ContainsFunc(userFields, func(field string) bool {
				return strings.Contains(field, strings.ToLower(filter[0]))
			})
			if !matchesFilter {
				// User doesn't match filter, return empty list
				identities = []models.Identity{}
			} else {
				identities = []models.Identity{
					{
						ID:    user.Email,
						Label: user.Name,
						User:  user,
					},
				}
			}
		} else {
			identities = []models.Identity{
				{
					ID:    user.Email,
					Label: user.Name,
					User:  user,
				},
			}
		}

	} else {

		// Query all identity providers in parallel
		ctx := context.Background()
		var wg sync.WaitGroup
		var mu sync.Mutex

		// Map to aggregate identities by name (to avoid duplicates across providers)
		identityMap := make(map[string]models.Identity)

		// Channel to collect errors
		errorChan := make(chan error, len(providerMap))

		for _, provider := range providerMap {
			wg.Add(1)
			go func(p models.Provider) {
				defer wg.Done()

				// Query identities from this provider with filter
				var identities []models.Identity
				var err error

				identities, err = p.GetClient().ListIdentities(ctx, filter...)

				if err != nil {
					logrus.WithError(err).WithField("provider", p.Name).Error("Failed to get identities from provider")
					errorChan <- err
					return
				}

				// Add identities to the map (thread-safe)
				mu.Lock()
				for _, identity := range identities {

					if identityType == IdentityTypeUser && identity.User == nil {
						continue
					}
					if identityType == IdentityTypeGroup && identity.Group == nil {
						continue
					}

					// Use identity ID as key to avoid duplicates
					// If the same identity exists from multiple providers, keep the first one
					if _, exists := identityMap[identity.GetId()]; !exists {
						identityMap[identity.GetId()] = identity
					}
				}
				mu.Unlock()
			}(provider)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errorChan)

		// Collect all errors
		var errors []error
		for err := range errorChan {
			if err != nil {
				errors = append(errors, err)
			}
		}

		// If there were errors, return them
		if len(errors) > 0 {
			return nil, fmt.Errorf("errors occurred while retrieving identities: %v", errors)
		}

		// Convert map to slice
		identities = make([]models.Identity, 0, len(identityMap))
		for _, identity := range identityMap {
			identities = append(identities, identity)
		}

	}

	return identities, nil

}
