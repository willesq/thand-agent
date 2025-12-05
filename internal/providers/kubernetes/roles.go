package kubernetes

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

func (p *kubernetesProvider) SynchronizeRoles(ctx context.Context, req models.SynchronizeRolesRequest) (*models.SynchronizeRolesResponse, error) {
	// Kubernetes doesn't have predefined roles like AWS IAM managed policies
	// Roles come from configuration, not from the provider itself

	logrus.Debug("Kubernetes provider: No built-in roles (roles come from configuration)")

	return &models.SynchronizeRolesResponse{
		Roles: []models.ProviderRole{},
	}, nil
}
