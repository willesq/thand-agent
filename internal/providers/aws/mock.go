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

	// Load AWS Permissions. This loads from internal/data/iam-dataset/aws/docs.json
	// this is an embedded resource
	err := p.awsProvider.LoadPermissions()
	if err != nil {
		return fmt.Errorf("failed to load permissions: %w", err)
	}

	err = p.awsProvider.LoadRoles()
	if err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	// Start background indexing
	go p.awsProvider.buildSearchIndexAsync()

	return nil
}
