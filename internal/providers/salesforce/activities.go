package salesforce

import "github.com/thand-io/agent/internal/models"

func (b *salesForceProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
