package aws

import (
	"context"

	"github.com/thand-io/agent/internal/models"
)

type awsProviderMock struct {
	*awsProvider
}

// NewMockAwsProvider creates a new mock AWS provider
func NewMockAwsProvider() *awsProviderMock {

	// Start by getting a copy of the base awsProvider
	return &awsProviderMock{
		awsProvider: &awsProvider{},
	}
}

func (p *awsProviderMock) Initialize(identifier string, provider models.Provider) error {
	// Initialize the embedded awsProvider struct first
	p.awsProvider = &awsProvider{}
	p.awsProvider.BaseProvider = models.NewBaseProvider(
		identifier,
		provider,
		models.ProviderCapabilityRBAC,
	)

	// Load AWS Permissions and Roles from shared singleton
	if err := p.Synchronize(context.Background(), nil); err != nil {
		return err
	}

	return nil
}

func (p *awsProviderMock) Synchronize(ctx context.Context, temporalService models.TemporalImpl) error {
	return PreSynchronizeActivities(ctx, temporalService, p)
}
