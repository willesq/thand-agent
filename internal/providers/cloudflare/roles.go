package cloudflare

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// LoadRoles loads Cloudflare roles from the API
func (p *cloudflareProvider) LoadRoles(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Loaded Cloudflare roles in %s", elapsed)
	}()

	accountRC := cloudflare.AccountIdentifier(p.accountID)
	//roles, err := p.client.ListAccountRoles(ctx, accountRC, cloudflare.ListAccountRolesParams{})
	roles, err := p.client.ListPermissionGroups(ctx, accountRC, cloudflare.ListPermissionGroupParams{})
	if err != nil {
		return fmt.Errorf("failed to list account roles: %w", err)
	}

	var rolesData []models.ProviderRole
	rolesMap := make(map[string]*models.ProviderRole, len(roles))

	// Convert to slice and create fast lookup map
	for _, role := range roles {
		newRole := models.ProviderRole{
			Id:   role.ID,
			Name: role.Name,
			Role: role, // Store the full Cloudflare role object for later use
		}
		rolesData = append(rolesData, newRole)
		rolesMap[strings.ToLower(role.Name)] = &rolesData[len(rolesData)-1]

		logrus.WithFields(logrus.Fields{
			"role":    role.Name,
			"role_id": role.ID,
		}).Debug("Loaded role")
	}

	p.roles = rolesData
	p.rolesMap = rolesMap

	logrus.WithFields(logrus.Fields{
		"roles": len(rolesData),
	}).Debug("Loaded Cloudflare roles, building search index in background")

	return nil
}

// GetRole retrieves a specific role by name
func (p *cloudflareProvider) GetRole(ctx context.Context, role string) (*models.ProviderRole, error) {
	role = strings.ToLower(role)
	// Fast map lookup
	if r, exists := p.rolesMap[role]; exists {
		return r, nil
	}
	return nil, fmt.Errorf("role not found")
}

// ListRoles lists all roles, optionally filtered
func (p *cloudflareProvider) ListRoles(ctx context.Context, filters ...string) ([]models.ProviderRole, error) {
	// If no filters, return all roles
	if len(filters) == 0 {
		return p.roles, nil
	}

	// Check if search index is ready
	p.indexMu.RLock()
	rolesIndex := p.rolesIndex
	p.indexMu.RUnlock()

	if rolesIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, rolesIndex, func(a *search.DocumentMatch, b models.ProviderRole) bool {
			return strings.Compare(a.ID, b.Name) == 0
		}, p.roles, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []models.ProviderRole
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, role := range p.roles {
		// Check if any filter matches the role name or description
		if strings.Contains(strings.ToLower(role.Name), filterText) ||
			strings.Contains(strings.ToLower(role.Description), filterText) {
			filtered = append(filtered, role)
		}
	}

	return filtered, nil
}
