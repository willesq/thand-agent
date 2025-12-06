package okta

import (
	"context"
	"fmt"
	"time"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

const resourceTypeApplication = "application"

// Okta has custom admin resources that can span multiple resources

// OktaApplication is an interface that represents the common methods
// available on all Okta application types returned by the SDK
// applications can be assigned to non-administrator users
type OktaApplication interface {
	GetId() string
	GetLabel() string
	GetName() string
	GetStatus() string
}

// SynchronizeResources loads Okta resources (applications) from the API
func (p *oktaProvider) SynchronizeResources(ctx context.Context, req models.SynchronizeResourcesRequest) (*models.SynchronizeResourcesResponse, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Loaded Okta resources in %s", elapsed)
	}()

	if req.Pagination == nil {
		req.Pagination = &models.PaginationOptions{
			PageSize: 100,
		}
	}

	// Load applications
	appResources, nextPageToken, err := p.loadApplicationResources(ctx, req.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to load application resources: %w", err)
	}

	response := models.SynchronizeResourcesResponse{
		Resources: appResources,
	}

	if len(nextPageToken) != 0 {
		response.Pagination = &models.PaginationOptions{
			Token:    nextPageToken,
			PageSize: req.Pagination.PageSize,
		}
	}

	logrus.WithFields(logrus.Fields{
		"resources": len(appResources),
	}).Debug("Loaded Okta resources")

	return &response, nil
}

// loadApplicationResources loads application resources from the Okta API
// and stores full application details in the Resource field for later use
func (p *oktaProvider) loadApplicationResources(ctx context.Context, pagination *models.PaginationOptions) ([]models.ProviderResource, string, error) {

	queryParams := &query.Params{
		Limit: int64(pagination.PageSize),
	}

	if len(pagination.Token) != 0 {
		queryParams.After = pagination.Token
	}

	apps, resp, err := p.client.Application.ListApplications(ctx, queryParams)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list applications: %w", err)
	}

	var resources []models.ProviderResource
	for _, app := range apps {
		// Type assertion to get the underlying Application struct
		// The App interface wraps different application types
		if appImpl, ok := app.(*okta.Application); ok {
			resource := models.ProviderResource{
				Id:       appImpl.Id,
				Type:     resourceTypeApplication,
				Name:     appImpl.Label,
				Resource: app, // In-memory only: stores the full app object to avoid GetApplication API calls later; not persisted due to json:"-" tag
			}
			resources = append(resources, resource)
		}
	}

	return resources, resp.NextPage, nil
}
