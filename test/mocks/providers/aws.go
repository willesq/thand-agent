package providers

import (
	coreProviders "github.com/thand-io/agent/internal/providers"
	"github.com/thand-io/agent/internal/providers/aws"
)

func init() {
	// Register mock AWS provider to override the real one for all tests
	coreProviders.Set(aws.ProviderName, aws.NewMockAwsProvider())
}
