package azure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/data"
	"github.com/thand-io/agent/internal/models"
)

// GetPermission retrieves a specific permission by name
func (p *azureProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	// Loop over permissions and match by name
	for _, perm := range p.permissions {
		if strings.EqualFold(perm.Name, permission) {
			return &perm, nil
		}
	}
	return nil, fmt.Errorf("permission '%s' not found", permission)
}

// ListPermissions returns all available permissions
func (p *azureProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {

	return common.BleveListSearch(ctx, p.permissionsIndex, func(a *search.DocumentMatch, b models.ProviderPermission) bool {
		return strings.Compare(a.ID, b.Name) == 0
	}, p.permissions, filters...)

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

	// Create in-memory Bleve index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return fmt.Errorf("failed to create search index: %w", err)
	}

	for _, operation := range azureOperations {
		permissions = append(permissions, models.ProviderPermission{
			Name:        operation.Name,
			Description: operation.Description,
		})
	}

	p.permissions = permissions
	p.permissionsIndex = index

	logrus.WithFields(logrus.Fields{
		"permissions": len(permissions),
	}).Debug("Loaded Azure permissions")

	return nil
}
