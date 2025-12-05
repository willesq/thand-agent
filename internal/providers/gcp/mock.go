package gcp

import (
	"context"

	"github.com/thand-io/agent/internal/models"
)

// gcpProviderMock is a mock implementation of gcpProvider for testing
type gcpProviderMock struct {
	*gcpProvider
}

// NewMockGcpProvider creates a new mock GCP provider for testing
func NewMockGcpProvider() models.ProviderImpl {
	return &gcpProviderMock{
		gcpProvider: &gcpProvider{},
	}
}

// Initialize loads permissions and roles without connecting to GCP
func (p *gcpProviderMock) Initialize(identifier string, provider models.Provider) error {
	// Initialize the embedded gcpProvider struct
	p.gcpProvider = &gcpProvider{}

	// Set the provider to the base provider
	p.BaseProvider = models.NewBaseProvider(
		identifier,
		provider,
		models.ProviderCapabilityRBAC,
	)

	// Load GCP Permissions and Roles from shared singleton
	if err := p.Synchronize(context.Background(), nil); err != nil {
		return err
	}

	return nil
}
