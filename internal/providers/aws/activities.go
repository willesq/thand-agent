package aws

import (
	"context"

	"github.com/thand-io/agent/internal/models"
)

func (b *awsProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}

// Aws uses static roles and permissions so we don't need to them.
// Instead we will just return these in the synchronize call.
func (p *awsProvider) Synchronize(ctx context.Context, temporalService models.TemporalImpl) error {

	// Before we kick off the synchronize lets update the static roles and permissions

	awsData, err := getSharedData()

	if err != nil {
		return err
	}

	p.SetRoles(awsData.roles)
	p.SetPermissions(awsData.permissions)

	return models.Synchronize(ctx, temporalService, p)
}
