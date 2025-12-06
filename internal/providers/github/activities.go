package github

import "github.com/thand-io/agent/internal/models"

func (b *githubProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
