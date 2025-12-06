package azure

import (
	"context"

	"github.com/thand-io/agent/internal/models"
)

func (b *azureProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}

// Azure uses static roles and permissions so we don't need to them.
// Instead we will just return these in the synchronize call.
func (p *azureProvider) Synchronize(ctx context.Context, temporalService models.TemporalImpl) error {

	// Before we kick off the synchronize lets update the static roles and permissions

	azureData, err := getSharedData()

	if err != nil {
		return err
	}

	p.SetRoles(azureData.roles)
	p.SetPermissions(azureData.permissions)

	return models.Synchronize(ctx, temporalService, p)
}
