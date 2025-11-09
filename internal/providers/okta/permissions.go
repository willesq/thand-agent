package okta

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// Okta permissions map - based on Okta's OAuth 2.0 scopes and admin permissions
// Reference: https://developer.okta.com/docs/reference/api/oidc/#scopes
var oktaPermissions = map[string]string{
	// User permissions
	"okta.users.read":                       "Read information about users",
	"okta.users.manage":                     "Create, update, and delete users",
	"okta.users.read.self":                  "Read information about the current user",
	"okta.users.manage.self":                "Update the current user's profile",
	"okta.users.appLinks.read":              "Read user's app links",
	"okta.users.credentials.manage":         "Manage user credentials including passwords and factors",
	"okta.users.credentials.resetFactors":   "Reset user authentication factors",
	"okta.users.credentials.resetPassword":  "Reset user passwords",
	"okta.users.credentials.expirePassword": "Expire user passwords",

	// Group permissions
	"okta.groups.read":   "Read information about groups",
	"okta.groups.manage": "Create, update, and delete groups",

	// App permissions
	"okta.apps.read":   "Read information about applications",
	"okta.apps.manage": "Create, update, and delete applications",

	// Authorization Server permissions
	"okta.authorizationServers.read":   "Read authorization server information",
	"okta.authorizationServers.manage": "Manage authorization servers",

	// Client permissions
	"okta.clients.read":     "Read OAuth client information",
	"okta.clients.manage":   "Manage OAuth clients",
	"okta.clients.register": "Register new OAuth clients",

	// Event permissions
	"okta.eventHooks.read":   "Read event hook configurations",
	"okta.eventHooks.manage": "Manage event hooks",
	"okta.events.read":       "Read system log events",

	// Factor permissions
	"okta.factors.read":   "Read information about factors",
	"okta.factors.manage": "Manage factor configurations",

	// Identity Provider permissions
	"okta.idps.read":   "Read identity provider information",
	"okta.idps.manage": "Manage identity providers",

	// Inline Hook permissions
	"okta.inlineHooks.read":   "Read inline hook configurations",
	"okta.inlineHooks.manage": "Manage inline hooks",

	// Linked Object permissions
	"okta.linkedObjects.read":   "Read linked object definitions",
	"okta.linkedObjects.manage": "Manage linked objects",

	// Log Stream permissions
	"okta.logStreams.read":   "Read log stream configurations",
	"okta.logStreams.manage": "Manage log streams",

	// Policy permissions
	"okta.policies.read":   "Read policy configurations",
	"okta.policies.manage": "Manage policies",

	// Profile Mapping permissions
	"okta.profileMappings.read":   "Read profile mapping configurations",
	"okta.profileMappings.manage": "Manage profile mappings",

	// Role permissions
	"okta.roles.read":   "Read role assignments",
	"okta.roles.manage": "Manage role assignments",

	// Schema permissions
	"okta.schemas.read":   "Read schema definitions",
	"okta.schemas.manage": "Manage schemas",

	// Session permissions
	"okta.sessions.read":   "Read session information",
	"okta.sessions.manage": "Manage user sessions",

	// Trusted Origin permissions
	"okta.trustedOrigins.read":   "Read trusted origin configurations",
	"okta.trustedOrigins.manage": "Manage trusted origins",

	// Template permissions
	"okta.templates.read":   "Read email and SMS templates",
	"okta.templates.manage": "Manage email and SMS templates",

	// Device permissions
	"okta.devices.read":   "Read device information",
	"okta.devices.manage": "Manage devices",

	// Domain permissions
	"okta.domains.read":   "Read domain configurations",
	"okta.domains.manage": "Manage domains",

	// Network Zone permissions
	"okta.networkZones.read":   "Read network zone configurations",
	"okta.networkZones.manage": "Manage network zones",

	// Org permissions
	"okta.orgs.read":   "Read organization information",
	"okta.orgs.manage": "Manage organization settings",

	// Rate Limit permissions
	"okta.rateLimitSettings.read":   "Read rate limit settings",
	"okta.rateLimitSettings.manage": "Manage rate limit settings",

	// Subscription permissions
	"okta.subscriptions.read":   "Read subscription information",
	"okta.subscriptions.manage": "Manage subscriptions",

	// Threat Insight permissions
	"okta.threatInsights.read":   "Read threat insight data",
	"okta.threatInsights.manage": "Manage threat insight settings",

	// API Token permissions
	"okta.apiTokens.read":   "Read API token information",
	"okta.apiTokens.manage": "Manage API tokens",

	// Standard OAuth/OIDC scopes
	"openid":         "OpenID Connect authentication",
	"profile":        "Access to user's profile information",
	"email":          "Access to user's email address",
	"address":        "Access to user's address information",
	"phone":          "Access to user's phone number",
	"offline_access": "Request refresh tokens for offline access",
}

func (p *oktaProvider) LoadPermissions() error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Loaded Okta permissions in %s", elapsed)
	}()

	var permissions []models.ProviderPermission
	permissionsMap := make(map[string]*models.ProviderPermission, len(oktaPermissions))

	// Convert to slice and create fast lookup map
	for name, description := range oktaPermissions {
		perm := models.ProviderPermission{
			Name:        name,
			Description: description,
		}
		permissions = append(permissions, perm)
		permissionsMap[strings.ToLower(name)] = &permissions[len(permissions)-1]
	}

	p.permissions = permissions
	p.permissionsMap = permissionsMap

	logrus.WithFields(logrus.Fields{
		"permissions": len(permissions),
	}).Debug("Loaded Okta permissions, building search index in background")

	return nil
}

func (p *oktaProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	permission = strings.ToLower(permission)
	// Fast map lookup
	if perm, exists := p.permissionsMap[permission]; exists {
		return perm, nil
	}
	return nil, fmt.Errorf("permission not found: %s", permission)
}

func (p *oktaProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {
	// If no filters, return all permissions
	if len(filters) == 0 {
		return p.permissions, nil
	}

	// Check if search index is ready
	p.indexMu.RLock()
	permissionsIndex := p.permissionsIndex
	p.indexMu.RUnlock()

	if permissionsIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, permissionsIndex, func(a *search.DocumentMatch, b models.ProviderPermission) bool {
			return strings.Compare(a.ID, b.Name) == 0
		}, p.permissions, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []models.ProviderPermission
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, perm := range p.permissions {
		// Check if any filter matches the permission name or description
		if strings.Contains(strings.ToLower(perm.Name), filterText) ||
			strings.Contains(strings.ToLower(perm.Description), filterText) {
			filtered = append(filtered, perm)
		}
	}

	return filtered, nil
}
