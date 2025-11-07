package cloudflare

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

// LoadPermissions loads Cloudflare permission reference data
// NOTE: Cloudflare doesn't have a "permissions" API concept like AWS IAM.
// These are reference mappings for documentation and role validation purposes only.
// Actual access control in Cloudflare is done through:
// 1. Account Roles (predefined by Cloudflare with associated permission groups)
// 2. Policies (custom combinations of Permission Groups + Resource Groups)
func (p *cloudflareProvider) LoadPermissions() error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed Cloudflare permission references in %s", elapsed)
	}()

	// Define Cloudflare permission groups for reference
	// These correspond to Cloudflare's Permission Groups that can be used in Policies
	// https://developers.cloudflare.com/fundamentals/api/reference/permissions/
	permissionsData := map[string]string{
		// Account-level permission groups
		"account:read":           "Read account settings",
		"account:edit":           "Edit account settings",
		"account_analytics:read": "Read account analytics",
		"account_settings:read":  "Read account settings",
		"account_settings:edit":  "Edit account settings",
		"account_rulesets:read":  "Read account rulesets",
		"account_rulesets:edit":  "Edit account rulesets",

		// Access permission groups
		"access:read": "Read Access applications and policies",
		"access:edit": "Edit Access applications and policies",

		// API Gateway permission groups
		"api_gateway:read": "Read API Gateway settings",
		"api_gateway:edit": "Edit API Gateway settings",

		// Analytics permission groups
		"analytics:read": "Read analytics data",

		// Billing permission groups
		"billing:read": "Read billing information",
		"billing:edit": "Edit billing information",

		// Cache permissions
		"cache_purge:edit": "Purge cache",

		// DNS permissions
		"dns_records:read": "Read DNS records",
		"dns_records:edit": "Edit DNS records",

		// Firewall permissions
		"waf:read":               "Read WAF rules",
		"waf:edit":               "Edit WAF rules",
		"firewall_services:read": "Read firewall services",
		"firewall_services:edit": "Edit firewall services",

		// Load Balancing permissions
		"lb:read": "Read load balancers",
		"lb:edit": "Edit load balancers",

		// Logs permissions
		"logs:read": "Read logs",
		"logs:edit": "Edit logs configuration",

		// Member permissions
		"member:read": "Read account members",
		"member:edit": "Edit account members",

		// Organization permissions
		"organization:read": "Read organization settings",
		"organization:edit": "Edit organization settings",

		// Page Rules permissions
		"page_rules:read": "Read page rules",
		"page_rules:edit": "Edit page rules",

		// SSL/TLS permissions
		"ssl:read": "Read SSL/TLS settings",
		"ssl:edit": "Edit SSL/TLS settings",

		// Stream permissions
		"stream:read": "Read Cloudflare Stream",
		"stream:edit": "Edit Cloudflare Stream",

		// Workers permissions
		"workers:read":            "Read Workers scripts",
		"workers:edit":            "Edit Workers scripts",
		"workers_kv_storage:read": "Read Workers KV storage",
		"workers_kv_storage:edit": "Edit Workers KV storage",
		"workers_r2:read":         "Read R2 storage",
		"workers_r2:edit":         "Edit R2 storage",

		// Zone permissions
		"zone:read":          "Read zone settings",
		"zone:edit":          "Edit zone settings",
		"zone_settings:read": "Read zone settings",
		"zone_settings:edit": "Edit zone settings",

		// Images permissions
		"images:read": "Read Cloudflare Images",
		"images:edit": "Edit Cloudflare Images",

		// Magic Transit permissions
		"magic_transit:read": "Read Magic Transit settings",
		"magic_transit:edit": "Edit Magic Transit settings",

		// Zero Trust permissions
		"zero_trust:read": "Read Zero Trust settings",
		"zero_trust:edit": "Edit Zero Trust settings",

		// DDoS permissions
		"ddos:read": "Read DDoS settings",
		"ddos:edit": "Edit DDoS settings",
	}

	var permissions []models.ProviderPermission
	permissionsMap := make(map[string]*models.ProviderPermission, len(permissionsData))

	// Convert to slice and create fast lookup map
	for name, description := range permissionsData {
		perm := models.ProviderPermission{
			Name:        name,
			Description: description,
		}
		permissions = append(permissions, perm)
		// Ensure all lookups use the new format (without '#' prefix)
		permissionsMap[strings.ToLower(name)] = &permissions[len(permissions)-1]
	}

	p.permissions = permissions
	p.permissionsMap = permissionsMap

	logrus.WithFields(logrus.Fields{
		"permissions": len(permissions),
	}).Debug("Loaded Cloudflare permissions, building search index in background")

	return nil
}

// GetPermission retrieves a specific permission by name
func (p *cloudflareProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	// Remove any legacy '#' prefix to support old references
	permission = strings.TrimPrefix(strings.ToLower(permission), "#")
	// Fast map lookup
	if perm, exists := p.permissionsMap[permission]; exists {
		return perm, nil
	}
	return nil, fmt.Errorf("permission not found")
}

// ListPermissions lists all permissions, optionally filtered
func (p *cloudflareProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {
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
