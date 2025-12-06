package azure

import "github.com/thand-io/agent/internal/models"

func (b *azureProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
