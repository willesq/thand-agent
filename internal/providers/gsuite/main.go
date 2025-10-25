package gsuite

import (
	"context"
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/thand-io/agent/internal/models"
	"github.com/thand-io/agent/internal/providers"
)

// gsuiteProvider implements the ProviderImpl interface for Google Workspace (GSuite)
type gsuiteProvider struct {
	*models.BaseProvider

	adminService    *admin.Service
	domain          string
	adminEmail      string
	identityCache   map[string]*models.Identity
	identities      []models.Identity
	identitiesIndex bleve.Index
}

func (p *gsuiteProvider) Initialize(provider models.Provider) error {
	p.BaseProvider = models.NewBaseProvider(
		provider,
		models.ProviderCapabilityIdentities,
	)

	// Get configuration
	config := p.GetConfig()

	// Get required configuration values
	serviceAccountKeyPath, foundKeyPath := config.GetString("service_account_key_path")
	if !foundKeyPath {
		return fmt.Errorf("service_account_key_path is required for GSuite provider")
	}

	domain, foundDomain := config.GetString("domain")
	if !foundDomain {
		return fmt.Errorf("domain is required for GSuite provider")
	}

	adminEmail, foundAdminEmail := config.GetString("admin_email")
	if !foundAdminEmail {
		return fmt.Errorf("admin_email is required for GSuite provider")
	}

	p.domain = domain
	p.adminEmail = adminEmail
	p.identityCache = make(map[string]*models.Identity)

	// Create admin service with domain-wide delegation
	ctx := context.Background()

	// Read service account credentials
	credentialsData, err := os.ReadFile(serviceAccountKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read service account key file: %w", err)
	}

	// Create JWT config for domain-wide delegation
	conf, err := google.JWTConfigFromJSON(credentialsData,
		admin.AdminDirectoryUserReadonlyScope,
		admin.AdminDirectoryGroupReadonlyScope,
	)
	if err != nil {
		return fmt.Errorf("failed to create JWT config: %w", err)
	}

	// Set the subject (admin email) for domain-wide delegation
	conf.Subject = adminEmail

	// Create the admin service with the JWT config
	adminService, err := admin.NewService(ctx, option.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		return fmt.Errorf("failed to create admin service: %w", err)
	}

	p.adminService = adminService

	// Initialize Bleve index for identities
	if err := p.initializeIdentitiesIndex(); err != nil {
		return fmt.Errorf("failed to initialize identities index: %w", err)
	}

	// Initialize identity cache
	if err := p.refreshIdentities(); err != nil {
		return fmt.Errorf("failed to initialize identity cache: %w", err)
	}

	return nil
}

// initializeIdentitiesIndex creates a new Bleve index for identities
func (p *gsuiteProvider) initializeIdentitiesIndex() error {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return fmt.Errorf("failed to create identities index: %w", err)
	}
	p.identitiesIndex = index
	return nil
}

// Refresh updates the cached identities by re-fetching from GSuite
func (p *gsuiteProvider) Refresh() error {
	return p.refreshIdentities()
}

func init() {
	providers.Register("gsuite", &gsuiteProvider{})
}
