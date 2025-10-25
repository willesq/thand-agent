package daemon

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

func (s *Server) getIdentities(c *gin.Context) {
	// Get user information
	if !s.Config.IsServer() {
		s.getErrorPage(c, http.StatusForbidden, "Identities endpoint is only available in server mode")
		return
	}

	_, foundUser, err := s.getUser(c)
	if err != nil {
		s.getErrorPage(c, http.StatusUnauthorized, "Unauthorized: unable to get user for list of available roles", err)
		return
	}

	// Get filter parameter from query string
	filter := c.Query("q")

	// Find providers with identity capabilities
	providerMap := s.Config.GetProvidersByCapability(models.ProviderCapabilityIdentities)

	identityProviders := []models.Provider{}
	for _, provider := range providerMap {
		if provider.Enabled {
			// Check if user has permission to access this provider
			if provider.HasPermission(foundUser.User) {
				identityProviders = append(identityProviders, provider)
			}
		}
	}

	var identities []models.Identity

	// If no identity providers found, return just the current user
	if len(identityProviders) == 0 {
		// Apply filter to current user if specified
		if filter != "" && !strings.Contains(strings.ToLower(foundUser.User.Name), strings.ToLower(filter)) &&
			!strings.Contains(strings.ToLower(foundUser.User.Email), strings.ToLower(filter)) {
			// User doesn't match filter, return empty list
			identities = []models.Identity{}
		} else {
			identities = []models.Identity{
				{
					Name: foundUser.User.Name,
					User: foundUser.User,
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
		errorChan := make(chan error, len(identityProviders))

		for _, provider := range identityProviders {
			wg.Add(1)
			go func(p models.Provider) {
				defer wg.Done()

				// Query identities from this provider with filter
				var identities []models.Identity
				var err error

				identities, err = p.GetClient().ListIdentities(ctx, filter)

				if err != nil {
					logrus.WithError(err).WithField("provider", p.Name).Error("Failed to get identities from provider")
					errorChan <- err
					return
				}

				// Add identities to the map (thread-safe)
				mu.Lock()
				for _, identity := range identities {
					// Use identity name as key to avoid duplicates
					// If the same identity exists from multiple providers, keep the first one
					if _, exists := identityMap[identity.Name]; !exists {
						identityMap[identity.Name] = identity
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

		// If there were errors, use getErrorPage
		if len(errors) > 0 {
			s.getErrorPage(c, http.StatusInternalServerError, "Failed to retrieve identities from some providers", errors...)
			return
		}

		// Convert map to slice
		identities = make([]models.Identity, 0, len(identityMap))
		for _, identity := range identityMap {
			identities = append(identities, identity)
		}

	}

	// Return the aggregated identities
	c.JSON(http.StatusOK, gin.H{
		"identities": identities,
		"providers":  len(identityProviders),
	})
}
