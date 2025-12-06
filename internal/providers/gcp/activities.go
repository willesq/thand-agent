package gcp

import "github.com/thand-io/agent/internal/models"

func (b *gcpProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
