package gsuite

import "github.com/thand-io/agent/internal/models"

func (b *gsuiteProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
