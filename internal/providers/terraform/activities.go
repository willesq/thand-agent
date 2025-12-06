package terraform

import "github.com/thand-io/agent/internal/models"

func (b *terraformProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
