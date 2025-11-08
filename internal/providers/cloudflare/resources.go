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

const resourceTypeZone = "zone"
const resourceTypeAccount = "account"

// LoadResources loads Cloudflare resources (zones, accounts) from the API
func (p *cloudflareProvider) LoadResources(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Loaded Cloudflare resources in %s", elapsed)
	}()

	var resourcesData []models.ProviderResource
	resourcesMap := make(map[string]*models.ProviderResource)

	// Load zones
	zoneResources, err := p.loadZoneResources(ctx)
	if err != nil {
		return fmt.Errorf("failed to load zone resources: %w", err)
	}
	resourcesData = append(resourcesData, zoneResources...)

	// Load accounts
	accountResources, err := p.loadAccountResources(ctx)
	if err != nil {
		return fmt.Errorf("failed to load account resources: %w", err)
	}
	resourcesData = append(resourcesData, accountResources...)

	// Build the resource map
	for i := range resourcesData {
		resource := &resourcesData[i]
		// Map by type:name (lowercase)
		resourcesMap[fmt.Sprintf("%s:%s", resource.Type, strings.ToLower(resource.Name))] = resource
		// Map by type:id (lowercase)
		resourcesMap[fmt.Sprintf("%s:%s", resource.Type, strings.ToLower(resource.Id))] = resource
	}

	p.resources = resourcesData
	p.resourcesMap = resourcesMap

	logrus.WithFields(logrus.Fields{
		"resources": len(resourcesData),
		"zones":     len(zoneResources),
		"accounts":  len(accountResources),
	}).Debug("Loaded Cloudflare resources")

	return nil
}

// loadZoneResources loads zone resources from the Cloudflare API
// and stores full zone details in the Resource field for later use
func (p *cloudflareProvider) loadZoneResources(ctx context.Context) ([]models.ProviderResource, error) {
	zones, err := p.client.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	var resources []models.ProviderResource
	for _, zone := range zones {
		resource := models.ProviderResource{
			Id:          zone.ID,
			Type:        resourceTypeZone,
			Name:        zone.Name,
			Description: fmt.Sprintf("Zone: %s (Status: %s)", zone.Name, zone.Status),
			Resource:    zone, // In-memory only: stores the full zone object to avoid ZoneDetails API calls later; not persisted due to json:"-" tag
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// loadAccountResources loads account resources from the Cloudflare API
// and stores full account details in the Resource field for later use
func (p *cloudflareProvider) loadAccountResources(ctx context.Context) ([]models.ProviderResource, error) {
	accounts, _, err := p.client.Accounts(ctx, cloudflare.AccountsListParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	var resources []models.ProviderResource
	for _, account := range accounts {
		resource := models.ProviderResource{
			Id:          account.ID,
			Type:        resourceTypeAccount,
			Name:        account.Name,
			Description: fmt.Sprintf("Account: %s (Type: %s)", account.Name, account.Type),
			Resource:    account, // Store the full account object to avoid Account API calls later
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// GetResource retrieves a specific resource by name or ID
func (p *cloudflareProvider) GetResource(ctx context.Context, resource string) (*models.ProviderResource, error) {
	resourceKey := strings.ToLower(resource)

	// Try lookup by name or ID
	if r, exists := p.resourcesMap[resourceKey]; exists {
		return r, nil
	}

	return nil, fmt.Errorf("resource not found: %s", resource)
}

// ListResources lists all resources, optionally filtered by search terms
func (p *cloudflareProvider) ListResources(ctx context.Context, filters ...string) ([]models.ProviderResource, error) {
	// If no filters, return all resources
	if len(filters) == 0 {
		return p.resources, nil
	}

	// Fallback to simple substring filtering
	var filtered []models.ProviderResource
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, resource := range p.resources {
		// Check if any filter matches the resource name, type, or description
		if strings.Contains(strings.ToLower(resource.Name), filterText) ||
			strings.Contains(strings.ToLower(resource.Type), filterText) ||
			strings.Contains(strings.ToLower(resource.Description), filterText) {
			filtered = append(filtered, resource)
		}
	}

	return filtered, nil
}
