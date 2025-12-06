package okta

import "github.com/thand-io/agent/internal/models"

func (b *oktaProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
