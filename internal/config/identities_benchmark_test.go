package config

import (
	"fmt"
	"testing"

	"github.com/thand-io/agent/internal/models"
)

func setupBenchmarkConfig(b *testing.B, numProviders int, identitiesPerProvider int) *Config {
	c := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	for i := range numProviders {
		name := fmt.Sprintf("provider-%d", i)
		identities := make(map[string]models.Identity)
		for j := range identitiesPerProvider {
			id := fmt.Sprintf("user-%d-%d@example.com", i, j)
			identities[id] = models.Identity{
				ID:    id,
				Label: fmt.Sprintf("User %d %d", i, j),
				User: &models.User{
					Email: id,
					Name:  fmt.Sprintf("User %d %d", i, j),
				},
			}
		}

		mockProvider := NewMockIdentityProvider(name, identities)

		// Create a provider model manually
		providerModel := models.Provider{
			Name:        name,
			Description: "Mock Identity Provider",
			Provider:    "mock",
			Enabled:     true,
		}
		providerModel.SetClient(mockProvider)

		c.Providers.Definitions[name] = providerModel
	}
	return c
}

func BenchmarkGetIdentitiesWithFilter_NoProviders(b *testing.B) {
	c := setupBenchmarkConfig(b, 0, 0)
	user := &models.User{Email: "test@example.com", Name: "Test User"}

	for b.Loop() {
		_, _ = c.GetIdentitiesWithFilter(user, IdentityTypeUser)
	}
}

func BenchmarkGetIdentitiesWithFilter_SingleProvider_Small(b *testing.B) {
	c := setupBenchmarkConfig(b, 1, 10)
	user := &models.User{Email: "test@example.com", Name: "Test User"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.GetIdentitiesWithFilter(user, IdentityTypeUser)
	}
}

func BenchmarkGetIdentitiesWithFilter_SingleProvider_Large(b *testing.B) {
	c := setupBenchmarkConfig(b, 1, 1000)
	user := &models.User{Email: "test@example.com", Name: "Test User"}

	for b.Loop() {
		_, _ = c.GetIdentitiesWithFilter(user, IdentityTypeUser)
	}
}

func BenchmarkGetIdentitiesWithFilter_ManyProviders_Small(b *testing.B) {
	c := setupBenchmarkConfig(b, 10, 10)
	user := &models.User{Email: "test@example.com", Name: "Test User"}

	for b.Loop() {
		_, _ = c.GetIdentitiesWithFilter(user, IdentityTypeUser)
	}
}

func BenchmarkGetIdentitiesWithFilter_ManyProviders_Large(b *testing.B) {
	c := setupBenchmarkConfig(b, 10, 100)
	user := &models.User{Email: "test@example.com", Name: "Test User"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.GetIdentitiesWithFilter(user, IdentityTypeUser)
	}
}

func BenchmarkGetIdentitiesWithFilter_WithFilter(b *testing.B) {
	c := setupBenchmarkConfig(b, 5, 100)
	user := &models.User{Email: "test@example.com", Name: "Test User"}
	filter := "user-0-50" // Should match one user in provider-0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.GetIdentitiesWithFilter(user, IdentityTypeUser, filter)
	}
}
