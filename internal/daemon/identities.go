package daemon

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	if foundUser == nil {
		s.getErrorPage(c, http.StatusUnauthorized, "Unauthorized: user not found")
		return
	}

	// Get filter parameter from query string
	filter := c.Query("q")

	identityProvidersCount := s.Config.GetProvidersByCapabilityWithUser(
		foundUser.User, models.ProviderCapabilityIdentities)
	identities, err := s.Config.GetIdentitiesWithFilter(foundUser.User, filter)
	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, "Failed to get identities", err)
		return
	}

	// Return the aggregated identities
	c.JSON(http.StatusOK, gin.H{
		"identities": identities,
		"providers":  len(identityProvidersCount),
	})
}
