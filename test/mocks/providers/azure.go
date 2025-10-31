package providers

import (
	coreProviders "github.com/thand-io/agent/internal/providers"
	"github.com/thand-io/agent/internal/providers/azure"
)

func init() {
	// Register mock Azure provider to override the real one for all tests
	coreProviders.Set(azure.ProviderName, azure.NewMockAzureProvider())
}
