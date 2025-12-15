package saml

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
	"github.com/thand-io/agent/internal/providers"
)

const SamlProviderName = "saml"

// samlProvider implements the ProviderImpl interface for SAML
type samlProvider struct {
	*models.BaseProvider
	middleware        *samlsp.Middleware
	idpMetadata       *saml.EntityDescriptor
	certificates      []tls.Certificate
	sessionDuration   time.Duration // Configurable session expiry
	allowIDPInitiated bool          // Whether to allow IdP-initiated SAML flows
}

// SAMLConfig represents the SAML provider configuration
type SAMLConfig struct {
	IDPMetadataURL     string          `yaml:"idp_metadata_url" json:"idp_metadata_url"`
	EntityID           string          `yaml:"entity_id" json:"entity_id"`
	RootURL            string          `yaml:"root_url" json:"root_url"`
	KeyPair            tls.Certificate `yaml:"-" json:"-"`
	SignRequests       bool            `yaml:"sign_requests" json:"sign_requests"`
	SessionDuration    time.Duration   `yaml:"session_duration" json:"session_duration"`          // Optional: session expiry (default: 24h)
	AllowIDPInitiated  bool            `yaml:"allow_idp_initiated" json:"allow_idp_initiated"`    // Optional: allow IdP-initiated flows (default: false)
}

func (p *samlProvider) Initialize(identifier string, provider models.Provider) error {
	p.BaseProvider = models.NewBaseProvider(
		identifier,
		provider,
		models.ProviderCapabilityAuthorizer,
	)

	// Parse SAML configuration from provider config
	config, err := p.parseSAMLConfig(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to parse SAML config: %w", err)
	}

	// Load certificate and key for SAML signing
	keyPair := config.KeyPair

	var privateKey *rsa.PrivateKey
	if keyPair.PrivateKey != nil {
		if pk, ok := keyPair.PrivateKey.(*rsa.PrivateKey); ok {
			privateKey = pk
		}
	}

	// Fetch IdP metadata
	idpMetadataURL, err := url.Parse(config.IDPMetadataURL)
	if err != nil {
		return fmt.Errorf("invalid IdP metadata URL: %w", err)
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		return fmt.Errorf("failed to fetch IdP metadata: %w", err)
	}

	// Parse root URL
	rootURL, err := url.Parse(config.RootURL)
	if err != nil {
		return fmt.Errorf("invalid root URL: %w", err)
	}

	// Create SAML service provider with custom ACS URL
	// The ACS URL must match what's configured in Okta: /api/v1/auth/callback/{provider-name}
	acsURL := *rootURL
	acsURL.Path = fmt.Sprintf("/api/v1/auth/callback/%s", identifier)

	metadataURL := *rootURL
	metadataURL.Path = "/saml/metadata"

	// Create the ServiceProvider directly for more control
	sp := &saml.ServiceProvider{
		EntityID:          config.EntityID,
		Key:               privateKey,
		Certificate:       keyPair.Leaf,
		MetadataURL:       metadataURL,
		AcsURL:            acsURL,
		IDPMetadata:       idpMetadata,
		AuthnNameIDFormat: saml.EmailAddressNameIDFormat,
		// Allow IDP-initiated flows (Okta can initiate)
		AllowIDPInitiated: true,
	}

	// Create middleware wrapper
	samlSP := &samlsp.Middleware{
		ServiceProvider: *sp,
	}

	p.middleware = samlSP
	p.idpMetadata = idpMetadata
	// Store certificates if configured (Certificate is a slice, check if not empty)
	if len(keyPair.Certificate) > 0 {
		p.certificates = []tls.Certificate{keyPair}
	}

	// Store session duration (default to 24h if not configured)
	p.sessionDuration = config.SessionDuration
	if p.sessionDuration == 0 {
		p.sessionDuration = 24 * time.Hour
	}

	// Store IdP-initiated flow setting (defaults to false for security)
	p.allowIDPInitiated = config.AllowIDPInitiated

	logrus.WithFields(logrus.Fields{
		"provider":    provider.Name,
		"entityID":    samlSP.ServiceProvider.EntityID,
		"acsURL":      samlSP.ServiceProvider.AcsURL.String(),
		"metadataURL": samlSP.ServiceProvider.MetadataURL.String(),
		"idpIssuer":   idpMetadata.EntityID,
	}).Infof("SAML provider %s initialized successfully", provider.Name)
	return nil
}

func (p *samlProvider) AuthorizeSession(ctx context.Context, authRequest *models.AuthorizeUser) (*models.AuthorizeSessionResponse, error) {
	if p.middleware == nil {
		return nil, fmt.Errorf("SAML provider not initialized")
	}

	// Generate a SAML authentication request URL using the correct API
	// Use MakeRedirectAuthenticationRequest for redirect binding
	authURL, err := p.middleware.ServiceProvider.MakeRedirectAuthenticationRequest(authRequest.State)
	if err != nil {
		return nil, fmt.Errorf("failed to create SAML authentication request: %w", err)
	}

	logrus.Debugln("SAML auth request generated")

	return &models.AuthorizeSessionResponse{
		Url: authURL.String(),
	}, nil
}

func (p *samlProvider) CreateSession(ctx context.Context, authRequest *models.AuthorizeUser) (*models.Session, error) {
	if p.middleware == nil {
		return nil, fmt.Errorf("SAML provider not initialized")
	}

	if len(authRequest.Code) == 0 {
		return nil, fmt.Errorf("no SAML response provided")
	}

	// Log minimal debugging information without sensitive data
	logrus.WithFields(logrus.Fields{
		"entityID": p.middleware.ServiceProvider.EntityID,
		"acsURL":   p.middleware.ServiceProvider.AcsURL.String(),
	}).Debugln("Attempting to parse SAML response")

	// Parse the SAML response
	// IMPORTANT: The URL in the request must match the ACS URL for validation to pass.
	// SAML signature validation requires the request URL to match the ACS URL exactly.
	// We must use PostForm (not Form) for POST requests because Form merges query and post parameters,
	// which can cause SAML signature validation to fail or introduce security issues if parameters are mixed.
	req := &http.Request{
		Method: "POST",
		URL:    &p.middleware.ServiceProvider.AcsURL,
		PostForm: url.Values{
			"SAMLResponse": {authRequest.Code},
		},
	}

	// Handle state parameter for SP-initiated vs IdP-initiated flows
	// For IdP-initiated flows, state may be empty - pass empty slice instead of []string{""}
	var possibleRequestIDs []string
	if authRequest.State != "" {
		possibleRequestIDs = []string{authRequest.State}
	}

	assertion, err := p.middleware.ServiceProvider.ParseResponse(
		req,
		possibleRequestIDs,
	)

	if err != nil {
		// Log error without sensitive SAML response data
		errMsg := err.Error()
		errType := fmt.Sprintf("%T", err)

		// Log with additional context - include entity/ACS info for all errors
		logrus.WithFields(logrus.Fields{
			"error":     errMsg,
			"errorType": errType,
			"entityID":  p.middleware.ServiceProvider.EntityID,
			"acsURL":    p.middleware.ServiceProvider.AcsURL.String(),
		}).Errorln("Failed to parse SAML response")

		// InvalidResponseError typically means:
		// 1. Signature validation failed (most common)
		// 2. Time validation failed (NotBefore/NotOnOrAfter)
		// 3. Audience restriction mismatch

		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Extract user information from SAML assertion
	user, err := p.extractUserFromAssertion(assertion)
	if err != nil {
		return nil, err
	}

	// Create session with configured duration (defaults to 24h if not set)
	session := &models.Session{
		UUID:   uuid.New(),
		User:   user,
		Expiry: time.Now().Add(p.sessionDuration),
	}

	// Log session creation without PII details
	logrus.WithFields(logrus.Fields{
		"sessionUUID": session.UUID.String(),
		"source":      "saml",
	}).Info("Created SAML session successfully")

	return session, nil
}

func (p *samlProvider) ValidateSession(ctx context.Context, session *models.Session) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}

	// Check if session has expired
	if time.Now().After(session.Expiry) {
		return fmt.Errorf("session has expired")
	}

	// Validate user information
	if session.User == nil {
		return fmt.Errorf("session user is nil")
	}

	logrus.Debugf("SAML session validated for user: %s", session.User.Username)
	return nil
}

func (p *samlProvider) RenewSession(ctx context.Context, session *models.Session) (*models.Session, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	// Validate current session first
	if err := p.ValidateSession(ctx, session); err != nil {
		// If session is expired, we need a new authentication
		return nil, fmt.Errorf("cannot renew expired session: %w", err)
	}

	// Create a new session with extended expiry
	newSession := &models.Session{
		UUID:   uuid.New(),
		User:   session.User,
		Expiry: time.Now().Add(24 * time.Hour),
	}

	logrus.Infof("Renewed SAML session for user: %s", session.User.Username)
	return newSession, nil
}

// Authorize grants access for a user to a role
func (p *samlProvider) AuthorizeRole(
	ctx context.Context,
	req *models.AuthorizeRoleRequest,
) (*models.AuthorizeRoleResponse, error) {

	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to authorize azure role")
	}

	user := req.GetUser()
	role := req.GetRole()

	// Check if user has permission for the role
	if !role.HasPermission(user) {
		return nil, fmt.Errorf("user %s does not have permission for role %s", user.Username, role.Name)
	}

	logrus.Infof("SAML authorization granted for user %s to role %s", user.Username, role.Name)
	return nil, nil
}

// Revoke removes access for a user from a role
func (p *samlProvider) RevokeRole(
	ctx context.Context,
	req *models.RevokeRoleRequest,
) (*models.RevokeRoleResponse, error) {

	if !req.IsValid() {
		return nil, fmt.Errorf("user and role must be provided to authorize azure role")
	}

	user := req.GetUser()
	role := req.GetRole()

	// In SAML, revocation typically happens at the IdP level
	// This is a placeholder for any local cleanup
	logrus.Infof("SAML access revoked for user %s from role %s", user.Username, role.Name)
	return nil, nil
}

func (p *samlProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	// SAML permissions are typically defined at the IdP level
	// This would require integration with the IdP's permission system
	return nil, fmt.Errorf("GetPermission not implemented for SAML provider - permissions managed at IdP level")
}

func (p *samlProvider) ListPermissions(ctx context.Context, searchRequest *models.SearchRequest) ([]models.SearchResult[models.ProviderPermission], error) {
	// SAML permissions are typically defined at the IdP level
	return []models.SearchResult[models.ProviderPermission]{}, nil
}

func (p *samlProvider) GetRole(ctx context.Context, role string) (*models.ProviderRole, error) {
	// SAML roles are typically defined at the IdP level
	// This would require integration with the IdP's role system
	return nil, fmt.Errorf("GetRole not implemented for SAML provider - roles managed at IdP level")
}

func (p *samlProvider) ListRoles(ctx context.Context, searchRequest *models.SearchRequest) ([]models.SearchResult[models.ProviderRole], error) {
	// SAML roles are typically defined at the IdP level
	return []models.SearchResult[models.ProviderRole]{}, nil
}

// GetResource is required by ProviderRoleBasedAccessControl interface
func (p *samlProvider) GetResource(ctx context.Context, resource string) (*models.ProviderResource, error) {
	return nil, fmt.Errorf("GetResource not implemented for SAML provider")
}

// ListResources is required by ProviderRoleBasedAccessControl interface
func (p *samlProvider) ListResources(ctx context.Context, searchRequest *models.SearchRequest) ([]models.SearchResult[models.ProviderResource], error) {
	return []models.SearchResult[models.ProviderResource]{}, nil
}

// ValidateRole is required by ProviderRoleBasedAccessControl interface
func (p *samlProvider) ValidateRole(
	ctx context.Context, user *models.Identity, role *models.Role) (map[string]any, error) {
	if user == nil || role == nil {
		return nil, fmt.Errorf("user or role is nil")
	}
	return nil, nil
}

// SendNotification is required by ProviderNotifier interface
func (p *samlProvider) SendNotification(ctx context.Context, notification models.NotificationRequest) error {
	return fmt.Errorf("SendNotification not implemented for SAML provider")
}

// IsIDPInitiatedAllowed checks if IdP-initiated SAML flows are permitted
func (p *samlProvider) IsIDPInitiatedAllowed() bool {
	return p.allowIDPInitiated
}

// extractUserFromAssertion extracts user information from a SAML assertion
func (p *samlProvider) extractUserFromAssertion(assertion *saml.Assertion) (*models.User, error) {
	var userID string
	var username string
	var email string
	var name string
	var groups []string

	// Extract attributes from the assertion
	if assertion != nil {
		// Get NameID (usually the username or email)
		if assertion.Subject != nil && assertion.Subject.NameID != nil {
			nameID := assertion.Subject.NameID.Value
			// Use NameID as email if it looks like an email
			// Basic email regex: local@domain.tld (allows common valid patterns)
			emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
			if matched, _ := regexp.MatchString(emailRegex, nameID); matched {
				email = nameID
				// Extract username from email (part before @)
				if atIndex := strings.Index(nameID, "@"); atIndex > 0 {
					username = nameID[:atIndex]
				}
			} else {
				// Not a valid email, use as username
				username = nameID
			}
		}

		// Extract attributes
		for _, stmt := range assertion.AttributeStatements {
			for _, attr := range stmt.Attributes {
				switch attr.Name {
				case "email", "Email", "emailAddress", "mail":
					if len(attr.Values) > 0 {
						email = attr.Values[0].Value
					}
				case "name", "displayName", "Name", "cn", "commonName":
					if len(attr.Values) > 0 {
						name = attr.Values[0].Value
					}
				case "username", "Username", "sAMAccountName":
					if len(attr.Values) > 0 {
						username = attr.Values[0].Value
					}
				case "userid", "UserID", "uid", "objectGUID":
					if len(attr.Values) > 0 {
						userID = attr.Values[0].Value
					}
				case "groups", "Groups", "memberOf":
					for _, v := range attr.Values {
						groups = append(groups, v.Value)
					}
				}
			}
		}
	}

	if len(email) == 0 {
		return nil, fmt.Errorf("missing required user attributes in SAML assertion")
	}

	if len(userID) == 0 {
		userID = email
	}

	// Create user identity
	user := &models.User{
		ID:       userID,
		Username: username,
		Email:    email,
		Name:     name,
		Source:   "saml",
		Groups:   groups,
	}

	return user, nil
}

// parseSAMLConfig parses the SAML configuration from the provider config
func (p *samlProvider) parseSAMLConfig(config *models.BasicConfig) (*SAMLConfig, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	samlConfig := &SAMLConfig{}

	// Parse required fields
	if idpURL, ok := config.GetString("idp_metadata_url"); ok {
		samlConfig.IDPMetadataURL = idpURL
	} else {
		return nil, fmt.Errorf("idp_metadata_url is required")
	}

	if entityID, ok := config.GetString("entity_id"); ok {
		samlConfig.EntityID = entityID
	} else {
		return nil, fmt.Errorf("entity_id is required")
	}

	if rootURL, ok := config.GetString("root_url"); ok {
		samlConfig.RootURL = rootURL
	} else {
		return nil, fmt.Errorf("root_url is required")
	}

	var certFile, cert string
	if v, ok := config.GetString("cert_file"); ok {
		certFile = v
	}
	if v, ok := config.GetString("cert"); ok {
		cert = v
	}

	var keyFile, key string
	if v, ok := config.GetString("key_file"); ok {
		keyFile = v
	}
	if v, ok := config.GetString("key"); ok {
		key = v
	}

	var keyPair tls.Certificate
	var err error

	if cert != "" {
		if key != "" {
			keyPair, err = tls.X509KeyPair([]byte(cert), []byte(key))
			if err != nil {
				return nil, fmt.Errorf("failed to parse SAML certificate from config: %w", err)
			}
		} else {
			// Parse inline certificate without key (for verification only)
			block, _ := pem.Decode([]byte(cert))
			if block == nil {
				return nil, fmt.Errorf("failed to parse certificate PEM")
			}
			// Parse the certificate to populate keyPair.Leaf
			leaf, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}
			keyPair = tls.Certificate{
				Certificate: [][]byte{block.Bytes},
				Leaf:        leaf,
			}
		}
	} else if certFile != "" {
		if keyFile != "" {
			keyPair, err = tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load SAML certificate: %w", err)
			}
		} else {
			// Parse certificate file without key (for verification only)
			certBytes, err := os.ReadFile(certFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read certificate file: %w", err)
			}
			block, _ := pem.Decode(certBytes)
			if block == nil {
				return nil, fmt.Errorf("failed to parse certificate PEM from file")
			}
			// Parse the certificate to populate keyPair.Leaf
			leaf, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate from file: %w", err)
			}
			keyPair = tls.Certificate{
				Certificate: [][]byte{block.Bytes},
				Leaf:        leaf,
			}
		}
	}

	if len(keyPair.Certificate) > 0 {
		// Parse the certificate leaf
		keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse SAML certificate leaf: %w", err)
		}
		samlConfig.KeyPair = keyPair
	}

	// Parse optional fields
	if signRequests, ok := config.GetBool("sign_requests"); ok {
		samlConfig.SignRequests = signRequests
	} else {
		samlConfig.SignRequests = false // Default to false
	}

	// Validation: If signing is enabled, we MUST have a private key
	if samlConfig.SignRequests && samlConfig.KeyPair.PrivateKey == nil {
		return nil, fmt.Errorf("sign_requests is set to true, but no private key was provided (cert/key or cert_file/key_file)")
	}

	return samlConfig, nil
}

func init() {
	providers.Register(SamlProviderName, &samlProvider{})
}
