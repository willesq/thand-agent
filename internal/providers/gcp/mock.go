package gcp

import (
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
func (p *gcpProviderMock) Initialize(provider models.Provider) error {
	// Initialize the embedded gcpProvider struct
	p.gcpProvider = &gcpProvider{}

	// Set the provider to the base provider
	p.BaseProvider = models.NewBaseProvider(
		provider,
		models.ProviderCapabilityRBAC,
	)

	// Load GCP Permissions and Roles from embedded resources (no cloud connection)
	// Use default stage for testing
	err := p.LoadPermissions(DefaultStage)
	if err != nil {
		return err
	}

	err = p.LoadRoles(DefaultStage)
	if err != nil {
		return err
	}

	// Skip initializing GCP clients (iamClient, crmClient)
	// This prevents actual GCP API connections during tests

	return nil
}
