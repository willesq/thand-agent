package azure

import (
	"github.com/thand-io/agent/internal/models"
)

// azureProviderMock is a mock implementation of azureProvider for testing
type azureProviderMock struct {
	*azureProvider
}

// NewMockAzureProvider creates a new mock Azure provider for testing
func NewMockAzureProvider() models.ProviderImpl {
	return &azureProviderMock{
		azureProvider: &azureProvider{},
	}
}

// Initialize loads permissions and roles without connecting to Azure
func (p *azureProviderMock) Initialize(provider models.Provider) error {
	// Initialize the embedded azureProvider struct
	p.azureProvider = &azureProvider{}

	// Set the provider to the base provider
	p.BaseProvider = models.NewBaseProvider(
		provider,
		models.ProviderCapabilityRBAC,
	)

	// Load Azure Permissions and Roles from shared singleton
	data, err := getSharedData()
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

	// Skip initializing Azure clients and credentials
	// This prevents actual Azure API connections during tests
	// We don't require subscription_id or other config for mocks

	return nil
}
