package okta

import (
	"context"
	"fmt"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/sirupsen/logrus"

	"github.com/thand-io/agent/internal/models"
	"github.com/thand-io/agent/internal/providers"
)

const OktaProviderName = "okta"

// oktaProvider implements the ProviderImpl interface for Okta
type oktaProvider struct {
	*models.BaseProvider

	client           *okta.Client
	orgUrl           string
	apiToken         string
	permissions      []models.ProviderPermission
	permissionsMap   map[string]*models.ProviderPermission
	permissionsIndex bleve.Index
	roles            []models.ProviderRole
	rolesMap         map[string]*models.ProviderRole
	rolesIndex       bleve.Index
	resources        []models.ProviderResource
	resourcesMap     map[string]*models.ProviderResource
	identities       []models.Identity
	identitiesMap    map[string]*models.Identity

	indexMu sync.RWMutex
}

func (p *oktaProvider) Initialize(provider models.Provider) error {
	p.BaseProvider = models.NewBaseProvider(
		provider,
		models.ProviderCapabilityRBAC,
		models.ProviderCapabilityIdentities,
	)

	// Get Okta configuration
	oktaConfig := p.GetConfig()

	// Create Okta client
	client, err := CreateOktaClient(oktaConfig)
	if err != nil {
		return fmt.Errorf("failed to create Okta client: %w", err)
	}

	p.client = client

	// Store configuration values
	orgUrl, found := oktaConfig.GetString("org_url")
	if !found {
		return fmt.Errorf("org_url is required for Okta provider")
	}
	p.orgUrl = orgUrl

	apiToken, found := oktaConfig.GetString("api_token")
	if !found {
		return fmt.Errorf("api_token is required for Okta provider")
	}
	p.apiToken = apiToken

	// Load Okta Permissions
	err = p.LoadPermissions()
	if err != nil {
		return fmt.Errorf("failed to load permissions: %w", err)
	}

	// Load Okta Roles
	err = p.LoadRoles()
	if err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	// Load Okta Resources (applications)
	err = p.LoadResources(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}

	// Load Okta Identities (users and groups)
	err = p.RefreshIdentities(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load identities: %w", err)
	}

	// Start background indexing
	go p.buildSearchIndex()

	logrus.WithField("org_url", p.orgUrl).Info("Initialized Okta provider")
	return nil
}

// CreateOktaClient creates and configures an Okta API client
func CreateOktaClient(oktaConfig *models.BasicConfig) (*okta.Client, error) {
	ctx := context.Background()

	// Get required configuration
	orgUrl, foundOrgUrl := oktaConfig.GetString("org_url")
	apiToken, foundApiToken := oktaConfig.GetString("api_token")

	if !foundOrgUrl {
		return nil, fmt.Errorf("org_url is required for Okta provider")
	}

	if !foundApiToken {
		return nil, fmt.Errorf("api_token is required for Okta provider")
	}

	// Configure Okta client
	_, client, err := okta.NewClient(
		ctx,
		okta.WithOrgUrl(orgUrl),
		okta.WithToken(apiToken),
		okta.WithCache(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize Okta client: %w", err)
	}

	logrus.WithField("org_url", orgUrl).Info("Created Okta client")
	return client, nil
}

// GetClient returns the Okta API client
func (p *oktaProvider) GetClient() *okta.Client {
	return p.client
}

// GetOrgUrl returns the Okta organization URL
func (p *oktaProvider) GetOrgUrl() string {
	return p.orgUrl
}

// GetApiToken returns the Okta API token (use with caution)
func (p *oktaProvider) GetApiToken() string {
	return p.apiToken
}

func init() {
	providers.Register(OktaProviderName, &oktaProvider{})
}
