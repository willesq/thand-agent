package okta

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

const resourceTypeApplication = "application"

// Okta has custom admin resources that can span multuiple resources

// OktaApplication is an interface that represents the common methods
// available on all Okta application types returned by the SDK
// applications can be assigned to non-administrator users
type OktaApplication interface {
	GetId() string
	GetLabel() string
	GetName() string
	GetStatus() string
}

// LoadResources loads Okta resources (applications) from the API
func (p *oktaProvider) LoadResources(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Loaded Okta resources in %s", elapsed)
	}()

	var resourcesData []models.ProviderResource
	resourcesMap := make(map[string]*models.ProviderResource)

	// Load applications
	appResources, err := p.loadApplicationResources(ctx)
	if err != nil {
		return fmt.Errorf("failed to load application resources: %w", err)
	}
	resourcesData = append(resourcesData, appResources...)

	// Build the resource map
	for i := range resourcesData {
		resource := &resourcesData[i]
		// Map by type:name (lowercase)
		resourcesMap[fmt.Sprintf("%s:%s", resource.Type, strings.ToLower(resource.Name))] = resource
		// Map by type:id (lowercase)
		resourcesMap[fmt.Sprintf("%s:%s", resource.Type, strings.ToLower(resource.Id))] = resource
		// Also map by name alone
		resourcesMap[strings.ToLower(resource.Name)] = resource
		// And by ID alone
		resourcesMap[strings.ToLower(resource.Id)] = resource
	}

	p.resources = resourcesData
	p.resourcesMap = resourcesMap

	logrus.WithFields(logrus.Fields{
		"resources":    len(resourcesData),
		"applications": len(appResources),
	}).Debug("Loaded Okta resources")

	return nil
}

// loadApplicationResources loads application resources from the Okta API
// and stores full application details in the Resource field for later use
func (p *oktaProvider) loadApplicationResources(ctx context.Context) ([]models.ProviderResource, error) {
	apps, _, err := p.client.Application.ListApplications(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	var resources []models.ProviderResource
	for _, app := range apps {
		// Type assertion to get the underlying Application struct
		// The App interface wraps different application types
		if appImpl, ok := app.(OktaApplication); ok {
			resource := models.ProviderResource{
				Id:          appImpl.GetId(),
				Type:        resourceTypeApplication,
				Name:        appImpl.GetLabel(),
				Description: fmt.Sprintf("Application: %s (Type: %s, Status: %s)", appImpl.GetLabel(), appImpl.GetName(), appImpl.GetStatus()),
				Resource:    app, // In-memory only: stores the full app object to avoid GetApplication API calls later; not persisted due to json:"-" tag
			}
			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// GetResource retrieves a specific resource by name or ID
func (p *oktaProvider) GetResource(ctx context.Context, resource string) (*models.ProviderResource, error) {
	resourceKey := strings.ToLower(resource)

	// Try lookup by name or ID
	if r, exists := p.resourcesMap[resourceKey]; exists {
		return r, nil
	}

	return nil, fmt.Errorf("resource not found: %s", resource)
}

// ListResources lists all resources, optionally filtered by search terms
func (p *oktaProvider) ListResources(ctx context.Context, filters ...string) ([]models.ProviderResource, error) {
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
