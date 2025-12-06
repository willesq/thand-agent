package aws

import "github.com/thand-io/agent/internal/models"

func (b *awsProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
