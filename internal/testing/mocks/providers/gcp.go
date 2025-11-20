package providers

import (
	coreProviders "github.com/thand-io/agent/internal/providers"
	"github.com/thand-io/agent/internal/providers/gcp"
)

func init() {
	// Register mock GCP provider to override the real one for all tests
	coreProviders.Set(gcp.GcpProviderName, gcp.NewMockGcpProvider())
}
