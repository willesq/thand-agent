package azure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/data"
	"github.com/thand-io/agent/internal/models"
)

// GetPermission retrieves a specific permission by name
func (p *azureProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	permission = strings.ToLower(permission)
	// Fast map lookup
	if perm, exists := p.permissionsMap[permission]; exists {
		return perm, nil
	}
	return nil, fmt.Errorf("permission '%s' not found", permission)
}

// ListPermissions returns all available permissions
func (p *azureProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {
	// If no filters, return all permissions
	if len(filters) == 0 {
		return p.permissions, nil
	}

	// Check if search index is ready
	p.indexMu.RLock()
	permissionsIndex := p.permissionsIndex
	p.indexMu.RUnlock()

	if permissionsIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, permissionsIndex, func(a *search.DocumentMatch, b models.ProviderPermission) bool {
			return strings.Compare(a.ID, b.Name) == 0
		}, p.permissions, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []models.ProviderPermission
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, perm := range p.permissions {
		// Check if any filter matches the permission name or description
		if strings.Contains(strings.ToLower(perm.Name), filterText) ||
			strings.Contains(strings.ToLower(perm.Description), filterText) {
			filtered = append(filtered, perm)
		}
	}

	return filtered, nil
}

// LoadPermissions loads Azure permissions from the embedded provider operations data
func (p *azureProvider) LoadPermissions() error {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed Azure permissions in %s", elapsed)
	}()

	// Get pre-parsed Azure permissions from data package
	azureOperations, err := data.GetParsedAzurePermissions()
	if err != nil {
		return fmt.Errorf("failed to get parsed Azure permissions: %w", err)
	}

	var permissions []models.ProviderPermission
	permissionsMap := make(map[string]*models.ProviderPermission, len(azureOperations))

	for _, operation := range azureOperations {
		permission := models.ProviderPermission{
			Name:        operation.Name,
			Description: operation.Description,
		}
		permissions = append(permissions, permission)
		permissionsMap[strings.ToLower(operation.Name)] = &permissions[len(permissions)-1] // Reference to the slice element
	}

	p.permissions = permissions
	p.permissionsMap = permissionsMap

	logrus.WithFields(logrus.Fields{
		"permissions": len(permissions),
	}).Debug("Loaded Azure permissions, building search index in background")

	return nil
}
