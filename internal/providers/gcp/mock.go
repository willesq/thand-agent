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

	// Load GCP Permissions and Roles from shared singleton
	// Use default stage for testing
	data, err := getSharedData(DefaultStage)
	if err != nil {
		return err
	}

	p.permissions = data.permissions
	p.permissionsMap = data.permissionsMap
	p.roles = data.roles
	p.rolesMap = data.rolesMap

	// Wait for indices to be ready for mock
	<-data.indexReady
	p.indexMu.Lock()
	p.permissionsIndex = data.permissionsIndex
	p.rolesIndex = data.rolesIndex
	p.indexMu.Unlock()

	// Skip initializing GCP clients (iamClient, crmClient)
	// This prevents actual GCP API connections during tests

	return nil
}
