package models

import (
	"go.temporal.io/sdk/worker"
)

// BaseProvider provides a base implementation of the ProviderImpl interface
func (b *BaseProvider) RegisterWorkflows(worker worker.Worker) error {

	if worker == nil {
		return ErrNotImplemented
	}

	if b.provider.GetClient() == nil {
		return ErrNotImplemented
	}

	// Register the Synchronize workflow. This updates roles, permissions,
	// resources and identities for RBAC
	worker.RegisterWorkflow(Synchronize)

	return nil

}

// RegisterActivities registers provider-specific activities with the Temporal worker
func (b *BaseProvider) RegisterActivities(worker worker.Worker) error {

	if worker == nil {
		return ErrNotImplemented
	}

	if b.provider.GetClient() == nil {
		return ErrNotImplemented
	}

	providerActivities := providerActivities{
		provider: b.provider.GetClient(),
	}

	// Register provider activities
	worker.RegisterActivity(providerActivities)

	return nil
}
