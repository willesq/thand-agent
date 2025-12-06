package kubernetes

import (
	"github.com/thand-io/agent/internal/models"
)

func (b *kubernetesProvider) RegisterActivities(temporalClient models.TemporalImpl) error {
	return models.RegisterActivities(temporalClient, models.NewProviderActivities(b))
}
