package providers

import (
	coreProviders "github.com/thand-io/agent/internal/providers"
	"github.com/thand-io/agent/internal/providers/kubernetes"
)

func init() {
	// Register mock Kubernetes provider to override the real one for all tests
	coreProviders.Set(kubernetes.ProviderName, kubernetes.NewMockKubernetesProvider())
}
