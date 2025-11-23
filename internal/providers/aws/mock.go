package aws

import (
	"fmt"

	"github.com/thand-io/agent/internal/models"
)

type awsProviderMock struct {
	*awsProvider
}

// NewMockAwsProvider creates a new mock AWS provider
func NewMockAwsProvider() *awsProviderMock {
	return &awsProviderMock{
		awsProvider: &awsProvider{},
	}
}

func (p *awsProviderMock) Initialize(provider models.Provider) error {
	// Initialize the embedded awsProvider struct first
	p.awsProvider = &awsProvider{}
	p.awsProvider.BaseProvider = models.NewBaseProvider(
		provider,
		models.ProviderCapabilityRBAC,
	)

	// Load AWS Permissions and Roles from shared singleton
	data, err := getSharedData()
	if err != nil {
		return fmt.Errorf("failed to load shared AWS data: %w", err)
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

	// TODO: Implement AWS client initialization if mock interactions with AWS services are needed.

	return nil
}
