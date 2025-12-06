package cloudflare

import "github.com/thand-io/agent/internal/models"

func (b *cloudflareProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
