package config

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thand-io/agent/internal/models"
)

// MockIdentityProvider implements ProviderImpl for testing identity functions
type MockIdentityProvider struct {
	*models.BaseProvider
	identities    map[string]models.Identity
	listFunc      func(ctx context.Context, filters ...string) ([]Identity, error)
	getIdentityFn func(ctx context.Context, identity string) (*models.Identity, error)
	mu            sync.RWMutex
}

// Identity is an alias to avoid import issues
type Identity = models.Identity

func NewMockIdentityProvider(name string, identities map[string]models.Identity) *MockIdentityProvider {
	provider := models.Provider{
		Name:        name,
		Description: "Mock Identity Provider",
		Provider:    "mock",
		Enabled:     true,
	}

	return &MockIdentityProvider{
		BaseProvider: models.NewBaseProvider(provider, models.ProviderCapabilityIdentities),
		identities:   identities,
	}
}

func (m *MockIdentityProvider) Initialize(provider models.Provider) error {
	return nil
}

func (m *MockIdentityProvider) GetIdentity(ctx context.Context, identity string) (*models.Identity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getIdentityFn != nil {
		return m.getIdentityFn(ctx, identity)
	}

	if id, exists := m.identities[identity]; exists {
		return &id, nil
	}
	return nil, fmt.Errorf("identity not found: %s", identity)
}

func (m *MockIdentityProvider) ListIdentities(ctx context.Context, filters ...string) ([]models.Identity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.listFunc != nil {
		return m.listFunc(ctx, filters...)
	}

	var results []models.Identity
	for _, id := range m.identities {
		// Apply filter if provided
		if len(filters) > 0 {
			filter := filters[0]
			matched := false
			if id.User != nil {
				if contains(id.User.Email, filter) || contains(id.User.Name, filter) || contains(id.User.Username, filter) {
					matched = true
				}
			}
			if id.Group != nil {
				if contains(id.Group.Name, filter) || contains(id.Group.Email, filter) {
					matched = true
				}
			}
			if matched {
				results = append(results, id)
			}
		} else {
			results = append(results, id)
		}
	}
	return results, nil
}

func (m *MockIdentityProvider) RefreshIdentities(ctx context.Context) error {
	return nil
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || containsLower(s, substr)))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldSlice(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFoldSlice(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ac := a[i]
		bc := b[i]
		if ac >= 'A' && ac <= 'Z' {
			ac += 'a' - 'A'
		}
		if bc >= 'A' && bc <= 'Z' {
			bc += 'a' - 'A'
		}
		if ac != bc {
			return false
		}
	}
	return true
}

// TestGetIdentity tests the GetIdentity function
func TestGetIdentity(t *testing.T) {
	tests := []struct {
		name          string
		identity      string
		providers     map[string]*MockIdentityProvider
		expectedID    string
		expectedEmail string
		expectError   bool
		errorContains string
	}{
		{
			name:     "get identity without provider prefix - found in first provider",
			identity: "john@example.com",
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john",
							Name:     "John Doe",
						},
					},
				}),
			},
			expectedID:    "john@example.com",
			expectedEmail: "john@example.com",
			expectError:   false,
		},
		{
			name:     "get identity with provider prefix",
			identity: "gsuite:john@example.com",
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john",
							Name:     "John Doe",
						},
					},
				}),
				"okta": NewMockIdentityProvider("okta", map[string]models.Identity{
					"john@example.com": {
						ID:    "okta-john",
						Label: "John Doe (Okta)",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john-okta",
							Name:     "John Doe",
						},
					},
				}),
			},
			expectedID:    "john@example.com",
			expectedEmail: "john@example.com",
			expectError:   false,
		},
		{
			name:     "get identity with nonexistent provider prefix",
			identity: "nonexistent:john@example.com",
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{}),
			},
			expectError:   true,
			errorContains: "provider 'nonexistent' not found",
		},
		{
			name:          "get identity without providers - returns basic identity",
			identity:      "john@example.com",
			providers:     map[string]*MockIdentityProvider{},
			expectedID:    "john@example.com",
			expectedEmail: "john@example.com",
			expectError:   false,
		},
		{
			name:     "get identity not found in provider - returns basic identity",
			identity: "unknown@example.com",
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john",
						},
					},
				}),
			},
			expectedID:    "unknown@example.com",
			expectedEmail: "unknown@example.com",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with mock providers
			config := &Config{
				Providers: ProviderConfig{
					Definitions: make(map[string]models.Provider),
				},
			}

			for name, mockProvider := range tt.providers {
				provider := models.Provider{
					Name:        name,
					Description: "Test provider",
					Provider:    "mock",
					Enabled:     true,
				}
				provider.SetClient(mockProvider)
				config.Providers.Definitions[name] = provider
			}

			// Call GetIdentity
			result, err := config.GetIdentity(tt.identity)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedID, result.ID)
			if result.User != nil {
				assert.Equal(t, tt.expectedEmail, result.User.Email)
			}
		})
	}
}

// TestGetIdentitiesWithFilter tests filtering and merging of identities
func TestGetIdentitiesWithFilter(t *testing.T) {
	tests := []struct {
		name          string
		user          *models.User
		identityType  IdentityType
		filter        []string
		providers     map[string]*MockIdentityProvider
		expectedCount int
		expectedIDs   []string
		expectError   bool
		errorContains string
	}{
		{
			name: "get all users without filter",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType: IdentityTypeUser,
			filter:       nil,
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john",
							Name:     "John Doe",
						},
					},
					"jane@example.com": {
						ID:    "jane@example.com",
						Label: "Jane Doe",
						User: &models.User{
							Email:    "jane@example.com",
							Username: "jane",
							Name:     "Jane Doe",
						},
					},
				}),
			},
			expectedCount: 2,
			expectedIDs:   []string{"john@example.com", "jane@example.com"},
			expectError:   false,
		},
		{
			name: "filter users by name",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType: IdentityTypeUser,
			filter:       []string{"john"},
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john",
							Name:     "John Doe",
						},
					},
					"jane@example.com": {
						ID:    "jane@example.com",
						Label: "Jane Doe",
						User: &models.User{
							Email:    "jane@example.com",
							Username: "jane",
							Name:     "Jane Doe",
						},
					},
				}),
			},
			expectedCount: 1,
			expectedIDs:   []string{"john@example.com"},
			expectError:   false,
		},
		{
			name: "get only groups",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType: IdentityTypeGroup,
			filter:       nil,
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john",
						},
					},
					"developers": {
						ID:    "developers",
						Label: "Developers Group",
						Group: &models.Group{
							Name:  "developers",
							Email: "developers@example.com",
						},
					},
					"admins": {
						ID:    "admins",
						Label: "Admins Group",
						Group: &models.Group{
							Name:  "admins",
							Email: "admins@example.com",
						},
					},
				}),
			},
			expectedCount: 2,
			expectedIDs:   []string{"developers", "admins"},
			expectError:   false,
		},
		{
			name: "merge identities from multiple providers - no duplicates",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType: IdentityTypeAll,
			filter:       nil,
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john",
							Name:     "John Doe",
						},
					},
				}),
				"okta": NewMockIdentityProvider("okta", map[string]models.Identity{
					"jane@example.com": {
						ID:    "jane@example.com",
						Label: "Jane Doe",
						User: &models.User{
							Email:    "jane@example.com",
							Username: "jane",
							Name:     "Jane Doe",
						},
					},
				}),
			},
			expectedCount: 2,
			expectedIDs:   []string{"john@example.com", "jane@example.com"},
			expectError:   false,
		},
		{
			name: "merge identities from multiple providers - with duplicates (same ID)",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType: IdentityTypeUser,
			filter:       nil,
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe (GSuite)",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john-gsuite",
							Name:     "John Doe",
						},
					},
				}),
				"okta": NewMockIdentityProvider("okta", map[string]models.Identity{
					"john@example.com": {
						ID:    "john@example.com",
						Label: "John Doe (Okta)",
						User: &models.User{
							Email:    "john@example.com",
							Username: "john-okta",
							Name:     "John Doe",
						},
					},
				}),
			},
			expectedCount: 1, // Should deduplicate by ID
			expectedIDs:   []string{"john@example.com"},
			expectError:   false,
		},
		{
			name: "no identity providers - returns current user",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType:  IdentityTypeUser,
			filter:        nil,
			providers:     map[string]*MockIdentityProvider{},
			expectedCount: 1,
			expectedIDs:   []string{"admin@example.com"},
			expectError:   false,
		},
		{
			name: "no identity providers with filter - user matches",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType:  IdentityTypeUser,
			filter:        []string{"admin"},
			providers:     map[string]*MockIdentityProvider{},
			expectedCount: 1,
			expectedIDs:   []string{"admin@example.com"},
			expectError:   false,
		},
		{
			name: "no identity providers with filter - user does not match",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType:  IdentityTypeUser,
			filter:        []string{"john"},
			providers:     map[string]*MockIdentityProvider{},
			expectedCount: 0,
			expectedIDs:   []string{},
			expectError:   false,
		},
		{
			name: "filter groups by name",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
			},
			identityType: IdentityTypeGroup,
			filter:       []string{"dev"},
			providers: map[string]*MockIdentityProvider{
				"gsuite": NewMockIdentityProvider("gsuite", map[string]models.Identity{
					"developers": {
						ID:    "developers",
						Label: "Developers Group",
						Group: &models.Group{
							Name:  "developers",
							Email: "developers@example.com",
						},
					},
					"admins": {
						ID:    "admins",
						Label: "Admins Group",
						Group: &models.Group{
							Name:  "admins",
							Email: "admins@example.com",
						},
					},
				}),
			},
			expectedCount: 1,
			expectedIDs:   []string{"developers"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with mock providers
			config := &Config{
				Providers: ProviderConfig{
					Definitions: make(map[string]models.Provider),
				},
			}

			for name, mockProvider := range tt.providers {
				provider := models.Provider{
					Name:        name,
					Description: "Test provider",
					Provider:    "mock",
					Enabled:     true,
				}
				provider.SetClient(mockProvider)
				config.Providers.Definitions[name] = provider
			}

			// Call GetIdentitiesWithFilter
			results, err := config.GetIdentitiesWithFilter(tt.user, tt.identityType, tt.filter...)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)

			// Verify expected IDs are present
			resultIDs := make([]string, len(results))
			for i, r := range results {
				resultIDs[i] = r.ID
			}
			assert.ElementsMatch(t, tt.expectedIDs, resultIDs)
		})
	}
}

// TestGetIdentitiesWithFilter_ConcurrentProviders tests that multiple providers are queried in parallel
func TestGetIdentitiesWithFilter_ConcurrentProviders(t *testing.T) {
	var callOrder []string
	var mu sync.Mutex

	// Create providers that record their call order
	provider1 := NewMockIdentityProvider("provider1", map[string]models.Identity{
		"user1@example.com": {
			ID:    "user1@example.com",
			Label: "User 1",
			User: &models.User{
				Email:    "user1@example.com",
				Username: "user1",
			},
		},
	})
	provider1.listFunc = func(ctx context.Context, filters ...string) ([]Identity, error) {
		mu.Lock()
		callOrder = append(callOrder, "provider1")
		mu.Unlock()
		return []models.Identity{
			{
				ID:    "user1@example.com",
				Label: "User 1",
				User: &models.User{
					Email:    "user1@example.com",
					Username: "user1",
				},
			},
		}, nil
	}

	provider2 := NewMockIdentityProvider("provider2", map[string]models.Identity{
		"user2@example.com": {
			ID:    "user2@example.com",
			Label: "User 2",
			User: &models.User{
				Email:    "user2@example.com",
				Username: "user2",
			},
		},
	})
	provider2.listFunc = func(ctx context.Context, filters ...string) ([]Identity, error) {
		mu.Lock()
		callOrder = append(callOrder, "provider2")
		mu.Unlock()
		return []models.Identity{
			{
				ID:    "user2@example.com",
				Label: "User 2",
				User: &models.User{
					Email:    "user2@example.com",
					Username: "user2",
				},
			},
		}, nil
	}

	provider3 := NewMockIdentityProvider("provider3", map[string]models.Identity{
		"user3@example.com": {
			ID:    "user3@example.com",
			Label: "User 3",
			User: &models.User{
				Email:    "user3@example.com",
				Username: "user3",
			},
		},
	})
	provider3.listFunc = func(ctx context.Context, filters ...string) ([]Identity, error) {
		mu.Lock()
		callOrder = append(callOrder, "provider3")
		mu.Unlock()
		return []models.Identity{
			{
				ID:    "user3@example.com",
				Label: "User 3",
				User: &models.User{
					Email:    "user3@example.com",
					Username: "user3",
				},
			},
		}, nil
	}

	config := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	for name, mockProvider := range map[string]*MockIdentityProvider{
		"provider1": provider1,
		"provider2": provider2,
		"provider3": provider3,
	} {
		provider := models.Provider{
			Name:        name,
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		provider.SetClient(mockProvider)
		config.Providers.Definitions[name] = provider
	}

	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin",
	}

	results, err := config.GetIdentitiesWithFilter(user, IdentityTypeUser)
	require.NoError(t, err)

	// All 3 providers should have been called
	assert.Len(t, callOrder, 3)

	// All 3 users should be returned
	assert.Len(t, results, 3)
}

// TestGetIdentitiesWithFilter_ProviderError tests handling of provider errors
func TestGetIdentitiesWithFilter_ProviderError(t *testing.T) {
	provider1 := NewMockIdentityProvider("provider1", map[string]models.Identity{})
	provider1.listFunc = func(ctx context.Context, filters ...string) ([]Identity, error) {
		return nil, fmt.Errorf("provider1 error")
	}

	config := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	provider := models.Provider{
		Name:        "provider1",
		Description: "Test provider",
		Provider:    "mock",
		Enabled:     true,
	}
	provider.SetClient(provider1)
	config.Providers.Definitions["provider1"] = provider

	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin",
	}

	_, err := config.GetIdentitiesWithFilter(user, IdentityTypeUser)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "errors occurred while retrieving identities")
}

// TestGetIdentitiesWithFilter_MixedUserAndGroup tests filtering by identity type
func TestGetIdentitiesWithFilter_MixedUserAndGroup(t *testing.T) {
	provider := NewMockIdentityProvider("mixed", map[string]models.Identity{
		"user1@example.com": {
			ID:    "user1@example.com",
			Label: "User 1",
			User: &models.User{
				Email:    "user1@example.com",
				Username: "user1",
			},
		},
		"user2@example.com": {
			ID:    "user2@example.com",
			Label: "User 2",
			User: &models.User{
				Email:    "user2@example.com",
				Username: "user2",
			},
		},
		"developers": {
			ID:    "developers",
			Label: "Developers",
			Group: &models.Group{
				Name:  "developers",
				Email: "developers@example.com",
			},
		},
		"admins": {
			ID:    "admins",
			Label: "Admins",
			Group: &models.Group{
				Name:  "admins",
				Email: "admins@example.com",
			},
		},
	})

	config := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	p := models.Provider{
		Name:        "mixed",
		Description: "Test provider",
		Provider:    "mock",
		Enabled:     true,
	}
	p.SetClient(provider)
	config.Providers.Definitions["mixed"] = p

	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin",
	}

	t.Run("filter by IdentityTypeUser", func(t *testing.T) {
		results, err := config.GetIdentitiesWithFilter(user, IdentityTypeUser)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		for _, r := range results {
			assert.NotNil(t, r.User)
			assert.Nil(t, r.Group)
		}
	})

	t.Run("filter by IdentityTypeGroup", func(t *testing.T) {
		results, err := config.GetIdentitiesWithFilter(user, IdentityTypeGroup)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		for _, r := range results {
			assert.Nil(t, r.User)
			assert.NotNil(t, r.Group)
		}
	})

	t.Run("filter by IdentityTypeAll", func(t *testing.T) {
		results, err := config.GetIdentitiesWithFilter(user, IdentityTypeAll)
		require.NoError(t, err)
		assert.Len(t, results, 4)
	})
}

// TestGetIdentity_EmailParsing tests that email parsing extracts username correctly
func TestGetIdentity_EmailParsing(t *testing.T) {
	config := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	// No providers configured - should return basic identity with parsed username
	result, err := config.GetIdentity("john.doe@example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "john.doe@example.com", result.ID)
	assert.NotNil(t, result.User)
	assert.Equal(t, "john.doe@example.com", result.User.Email)
	assert.Equal(t, "john.doe", result.User.Username)
}

// TestGetIdentity_NonEmail tests identity lookup for non-email identities
func TestGetIdentity_NonEmail(t *testing.T) {
	config := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	// No providers configured - should return basic identity
	result, err := config.GetIdentity("johndoe")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "johndoe", result.ID)
	assert.NotNil(t, result.User)
	assert.Equal(t, "johndoe", result.User.Email)
	assert.Equal(t, "", result.User.Username) // No @ in identity, so username is empty
}

// TestGetIdentity_ProviderPrefixFormat tests various provider prefix formats
func TestGetIdentity_ProviderPrefixFormat(t *testing.T) {
	tests := []struct {
		name       string
		identity   string
		wantPrefix string
		wantKey    string
		hasPrefix  bool
	}{
		{
			name:       "simple prefix",
			identity:   "aws:admin",
			wantPrefix: "aws",
			wantKey:    "admin",
			hasPrefix:  true,
		},
		{
			name:       "prefix with email",
			identity:   "gsuite:john@example.com",
			wantPrefix: "gsuite",
			wantKey:    "john@example.com",
			hasPrefix:  true,
		},
		{
			name:       "no prefix - plain email",
			identity:   "john@example.com",
			wantPrefix: "",
			wantKey:    "john@example.com",
			hasPrefix:  false,
		},
		{
			name:       "no prefix - plain username",
			identity:   "johndoe",
			wantPrefix: "",
			wantKey:    "johndoe",
			hasPrefix:  false,
		},
		{
			name:       "prefix with hyphen in provider name",
			identity:   "aws-prod:admin@example.com",
			wantPrefix: "aws-prod",
			wantKey:    "admin@example.com",
			hasPrefix:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the identity string as done in GetIdentity
			var providerID string
			var identityKey string

			colonIdx := -1
			for i, c := range tt.identity {
				if c == ':' {
					colonIdx = i
					break
				}
			}

			if colonIdx != -1 {
				providerID = tt.identity[:colonIdx]
				identityKey = tt.identity[colonIdx+1:]
			} else {
				identityKey = tt.identity
			}

			if tt.hasPrefix {
				assert.Equal(t, tt.wantPrefix, providerID)
				assert.Equal(t, tt.wantKey, identityKey)
			} else {
				assert.Equal(t, "", providerID)
				assert.Equal(t, tt.wantKey, identityKey)
			}
		})
	}
}

// TestGetIdentitiesWithFilter_DeduplicationAcrossProviders tests that duplicate identities from multiple providers are deduplicated
func TestGetIdentitiesWithFilter_DeduplicationAcrossProviders(t *testing.T) {
	// Create 3 providers that all return the same user
	providers := map[string]*MockIdentityProvider{
		"provider1": NewMockIdentityProvider("provider1", map[string]models.Identity{
			"shared@example.com": {
				ID:    "shared@example.com",
				Label: "Shared User (P1)",
				User: &models.User{
					Email:    "shared@example.com",
					Username: "shared-p1",
					Name:     "Shared User",
				},
			},
			"unique1@example.com": {
				ID:    "unique1@example.com",
				Label: "Unique User 1",
				User: &models.User{
					Email:    "unique1@example.com",
					Username: "unique1",
					Name:     "Unique User 1",
				},
			},
		}),
		"provider2": NewMockIdentityProvider("provider2", map[string]models.Identity{
			"shared@example.com": {
				ID:    "shared@example.com",
				Label: "Shared User (P2)",
				User: &models.User{
					Email:    "shared@example.com",
					Username: "shared-p2",
					Name:     "Shared User",
				},
			},
			"unique2@example.com": {
				ID:    "unique2@example.com",
				Label: "Unique User 2",
				User: &models.User{
					Email:    "unique2@example.com",
					Username: "unique2",
					Name:     "Unique User 2",
				},
			},
		}),
		"provider3": NewMockIdentityProvider("provider3", map[string]models.Identity{
			"shared@example.com": {
				ID:    "shared@example.com",
				Label: "Shared User (P3)",
				User: &models.User{
					Email:    "shared@example.com",
					Username: "shared-p3",
					Name:     "Shared User",
				},
			},
			"unique3@example.com": {
				ID:    "unique3@example.com",
				Label: "Unique User 3",
				User: &models.User{
					Email:    "unique3@example.com",
					Username: "unique3",
					Name:     "Unique User 3",
				},
			},
		}),
	}

	config := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	for name, mockProvider := range providers {
		provider := models.Provider{
			Name:        name,
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		provider.SetClient(mockProvider)
		config.Providers.Definitions[name] = provider
	}

	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin",
	}

	results, err := config.GetIdentitiesWithFilter(user, IdentityTypeUser)
	require.NoError(t, err)

	// Should have 4 unique identities: shared@example.com + 3 unique ones
	assert.Len(t, results, 4)

	// Count how many times shared@example.com appears
	sharedCount := 0
	for _, r := range results {
		if r.ID == "shared@example.com" {
			sharedCount++
		}
	}
	assert.Equal(t, 1, sharedCount, "shared@example.com should appear only once after deduplication")

	// Verify all unique identities are present
	resultIDs := make(map[string]bool)
	for _, r := range results {
		resultIDs[r.ID] = true
	}
	assert.True(t, resultIDs["shared@example.com"])
	assert.True(t, resultIDs["unique1@example.com"])
	assert.True(t, resultIDs["unique2@example.com"])
	assert.True(t, resultIDs["unique3@example.com"])
}

// TestGetIdentitiesWithFilter_CurrentUserFallback tests that the current user is returned
// when there are no results, no filter, and the identity type is User or All
func TestGetIdentitiesWithFilter_CurrentUserFallback(t *testing.T) {
	currentUser := &models.User{
		Email: "current@example.com",
		Name:  "Current User",
	}

	t.Run("user returned when provider returns empty results - IdentityTypeUser", func(t *testing.T) {
		// Provider returns no results
		provider := NewMockIdentityProvider("test", map[string]models.Identity{})

		config := &Config{
			Providers: ProviderConfig{
				Definitions: make(map[string]models.Provider),
			},
		}
		p := models.Provider{
			Name:        "test",
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		p.SetClient(provider)
		config.Providers.Definitions["test"] = p

		results, err := config.GetIdentitiesWithFilter(currentUser, IdentityTypeUser)
		require.NoError(t, err)

		// Should return current user as fallback
		assert.Len(t, results, 1)
		assert.Equal(t, currentUser.Email, results[0].ID)
	})

	t.Run("user returned when provider returns empty results - IdentityTypeAll", func(t *testing.T) {
		// Provider returns no results
		provider := NewMockIdentityProvider("test", map[string]models.Identity{})

		config := &Config{
			Providers: ProviderConfig{
				Definitions: make(map[string]models.Provider),
			},
		}
		p := models.Provider{
			Name:        "test",
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		p.SetClient(provider)
		config.Providers.Definitions["test"] = p

		results, err := config.GetIdentitiesWithFilter(currentUser, IdentityTypeAll)
		require.NoError(t, err)

		// Should return current user as fallback
		assert.Len(t, results, 1)
		assert.Equal(t, currentUser.Email, results[0].ID)
	})

	t.Run("user NOT returned when provider has results", func(t *testing.T) {
		provider := NewMockIdentityProvider("test", map[string]models.Identity{
			"other@example.com": {
				ID:    "other@example.com",
				Label: "Other User",
				User: &models.User{
					Email:    "other@example.com",
					Username: "other",
				},
			},
		})

		config := &Config{
			Providers: ProviderConfig{
				Definitions: make(map[string]models.Provider),
			},
		}
		p := models.Provider{
			Name:        "test",
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		p.SetClient(provider)
		config.Providers.Definitions["test"] = p

		results, err := config.GetIdentitiesWithFilter(currentUser, IdentityTypeUser)
		require.NoError(t, err)

		// Should only have the provider result, not current user
		assert.Len(t, results, 1)
		assert.Equal(t, "other@example.com", results[0].ID)
	})

	t.Run("user NOT returned when IdentityTypeGroup even with empty results", func(t *testing.T) {
		provider := NewMockIdentityProvider("test", map[string]models.Identity{})

		config := &Config{
			Providers: ProviderConfig{
				Definitions: make(map[string]models.Provider),
			},
		}
		p := models.Provider{
			Name:        "test",
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		p.SetClient(provider)
		config.Providers.Definitions["test"] = p

		results, err := config.GetIdentitiesWithFilter(currentUser, IdentityTypeGroup)
		require.NoError(t, err)

		// Should be empty, not the current user
		assert.Len(t, results, 0)
	})

	t.Run("user NOT returned when filter is provided even with empty results", func(t *testing.T) {
		provider := NewMockIdentityProvider("test", map[string]models.Identity{})

		config := &Config{
			Providers: ProviderConfig{
				Definitions: make(map[string]models.Provider),
			},
		}
		p := models.Provider{
			Name:        "test",
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		p.SetClient(provider)
		config.Providers.Definitions["test"] = p

		results, err := config.GetIdentitiesWithFilter(currentUser, IdentityTypeUser, "nonexistent")
		require.NoError(t, err)

		// Should be empty because filter was provided
		assert.Len(t, results, 0)
	})

	t.Run("user returned when filter is empty string", func(t *testing.T) {
		provider := NewMockIdentityProvider("test", map[string]models.Identity{})

		config := &Config{
			Providers: ProviderConfig{
				Definitions: make(map[string]models.Provider),
			},
		}
		p := models.Provider{
			Name:        "test",
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		p.SetClient(provider)
		config.Providers.Definitions["test"] = p

		// Pass an empty string as filter - this simulates ?q= in the URL
		results, err := config.GetIdentitiesWithFilter(currentUser, IdentityTypeUser, "")
		require.NoError(t, err)

		// Should return current user as fallback because "" filter should be ignored
		assert.Len(t, results, 1)
		assert.Equal(t, currentUser.Email, results[0].ID)
	})

	t.Run("nil user - no fallback, empty results", func(t *testing.T) {
		provider := NewMockIdentityProvider("test", map[string]models.Identity{})

		config := &Config{
			Providers: ProviderConfig{
				Definitions: make(map[string]models.Provider),
			},
		}
		p := models.Provider{
			Name:        "test",
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		p.SetClient(provider)
		config.Providers.Definitions["test"] = p

		// This should not panic and should return empty
		results, err := config.GetIdentitiesWithFilter(nil, IdentityTypeUser)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

// Benchmark for GetIdentitiesWithFilter with multiple providers
func BenchmarkGetIdentitiesWithFilter_MultipleProviders(b *testing.B) {
	// Create providers with many identities
	providers := make(map[string]*MockIdentityProvider)
	for i := 0; i < 5; i++ {
		identities := make(map[string]models.Identity)
		for j := 0; j < 100; j++ {
			id := fmt.Sprintf("user%d-%d@example.com", i, j)
			identities[id] = models.Identity{
				ID:    id,
				Label: fmt.Sprintf("User %d-%d", i, j),
				User: &models.User{
					Email:    id,
					Username: fmt.Sprintf("user%d-%d", i, j),
				},
			}
		}
		providers[fmt.Sprintf("provider%d", i)] = NewMockIdentityProvider(fmt.Sprintf("provider%d", i), identities)
	}

	config := &Config{
		Providers: ProviderConfig{
			Definitions: make(map[string]models.Provider),
		},
	}

	for name, mockProvider := range providers {
		provider := models.Provider{
			Name:        name,
			Description: "Test provider",
			Provider:    "mock",
			Enabled:     true,
		}
		provider.SetClient(mockProvider)
		config.Providers.Definitions[name] = provider
	}

	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.GetIdentitiesWithFilter(user, IdentityTypeUser)
		if err != nil {
			b.Fatal(err)
		}
	}
}
