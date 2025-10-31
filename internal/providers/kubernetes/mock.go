package kubernetes

import (
	"github.com/thand-io/agent/internal/models"
)

// kubernetesProviderMock is a mock implementation of kubernetesProvider for testing
type kubernetesProviderMock struct {
	*kubernetesProvider
}

// NewMockKubernetesProvider creates a new mock Kubernetes provider for testing
func NewMockKubernetesProvider() models.ProviderImpl {
	return &kubernetesProviderMock{
		kubernetesProvider: &kubernetesProvider{},
	}
}

// Initialize loads permissions and roles without connecting to Kubernetes
func (p *kubernetesProviderMock) Initialize(provider models.Provider) error {
	// Initialize the embedded kubernetesProvider struct
	p.kubernetesProvider = &kubernetesProvider{}

	// Set the provider to the base provider
	p.BaseProvider = models.NewBaseProvider(
		provider,
		models.ProviderCapabilityRBAC,
	)

	// Load Kubernetes Permissions and Roles from embedded resources (no cluster connection)
	err := p.LoadPermissions()
	if err != nil {
		return err
	}

	err = p.LoadRoles()
	if err != nil {
		return err
	}

	// Skip initializing Kubernetes client
	// This prevents actual Kubernetes API connections during tests

	return nil
}
