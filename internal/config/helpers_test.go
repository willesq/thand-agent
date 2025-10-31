package config

import (
	"testing"

	"github.com/thand-io/agent/internal/models"
	_ "github.com/thand-io/agent/test/mocks/providers" // Register mock providers
)

// newTestConfig creates a Config with mock providers initialized
// The mock providers are registered via the test/mocks package import
// This prevents actual cloud provider client initialization during tests
// but still loads roles and permissions from embedded data
func newTestConfig(t *testing.T, roles map[string]models.Role, providers map[string]models.Provider) *Config {
	t.Helper()

	config := &Config{
		mode: "server", // Set mode to server so getProviderImplementation works
		Roles: RoleConfig{
			Definitions: roles,
		},
		Providers: ProviderConfig{
			Definitions: providers,
		},
	}

	// Initialize providers - will use mock implementations registered in test/mocks
	if len(providers) > 0 {
		loadedProviders, err := config.InitializeProviders(providers)
		if err != nil {
			t.Fatalf("Failed to initialize mock providers: %v", err)
		}
		config.Providers.Definitions = loadedProviders
	}

	return config
}
