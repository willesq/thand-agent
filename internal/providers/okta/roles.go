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

// Okta predefined administrator roles
// Reference: https://help.okta.com/en-us/content/topics/security/administrators-admin-comparison.htm
var oktaPredefinedRoles = map[string]models.ProviderRole{
	"SUPER_ADMIN": {
		Id:          "SUPER_ADMIN",
		Name:        "Super Administrator",
		Description: "Full administrative access to the Okta organization. Can perform all administrative tasks including managing other administrators.",
	},
	"ORG_ADMIN": {
		Id:          "ORG_ADMIN",
		Name:        "Organization Administrator",
		Description: "Full administrative access except for managing super administrators. Can manage users, groups, apps, and most org settings.",
	},
	"APP_ADMIN": {
		Id:          "APP_ADMIN",
		Name:        "Application Administrator",
		Description: "Can create and manage applications and their assignments. Cannot manage users or groups unless they are assigned to apps.",
	},
	"USER_ADMIN": {
		Id:          "USER_ADMIN",
		Name:        "User Administrator",
		Description: "Can create and manage users and groups. Cannot manage applications or advanced settings.",
	},
	"GROUP_MEMBERSHIP_ADMIN": {
		Id:          "GROUP_MEMBERSHIP_ADMIN",
		Name:        "Group Membership Administrator",
		Description: "Can manage group membership but cannot create or delete groups.",
	},
	"HELP_DESK_ADMIN": {
		Id:          "HELP_DESK_ADMIN",
		Name:        "Help Desk Administrator",
		Description: "Can reset passwords and MFA factors for users. Limited administrative capabilities for support purposes.",
	},
	"READ_ONLY_ADMIN": {
		Id:          "READ_ONLY_ADMIN",
		Name:        "Read-Only Administrator",
		Description: "Can view all aspects of the Okta organization but cannot make changes.",
	},
	"MOBILE_ADMIN": {
		Id:          "MOBILE_ADMIN",
		Name:        "Mobile Administrator",
		Description: "Can manage mobile device management settings and policies.",
	},
	"API_ACCESS_MANAGEMENT_ADMIN": {
		Id:          "API_ACCESS_MANAGEMENT_ADMIN",
		Name:        "API Access Management Administrator",
		Description: "Can manage authorization servers, scopes, and claims for API access management.",
	},
	"REPORT_ADMIN": {
		Id:          "REPORT_ADMIN",
		Name:        "Report Administrator",
		Description: "Can create and view reports about the Okta organization.",
	},
	"GROUP_ADMIN": {
		Id:          "GROUP_ADMIN",
		Name:        "Group Administrator",
		Description: "Can create, manage, and delete groups. Can manage group membership.",
	},
}

// Also load in user groups as these can have roles assigned too

func (p *oktaProvider) LoadRoles() error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Loaded Okta roles in %s", elapsed)
	}()

	var roles []models.ProviderRole
	rolesMap := make(map[string]*models.ProviderRole)

	// Load predefined standard roles
	// These are Okta's built-in administrator roles that are consistent across all Okta orgs
	// Reference: https://help.okta.com/en-us/content/topics/security/administrators-admin-comparison.htm
	for _, role := range oktaPredefinedRoles {
		roles = append(roles, role)
		rolesMap[strings.ToLower(role.Name)] = &roles[len(roles)-1]
	}

	p.roles = roles
	p.rolesMap = rolesMap

	logrus.WithFields(logrus.Fields{
		"roles": len(roles),
	}).Debug("Loaded Okta standard roles, building search index in background")

	return nil
}

func (p *oktaProvider) GetRole(ctx context.Context, role string) (*models.ProviderRole, error) {
	role = strings.ToLower(role)
	// Fast map lookup
	if r, exists := p.rolesMap[role]; exists {
		return r, nil
	}
	return nil, fmt.Errorf("role not found: %s", role)
}

func (p *oktaProvider) ListRoles(ctx context.Context, filters ...string) ([]models.ProviderRole, error) {
	// If no filters, return all roles
	if len(filters) == 0 {
		return p.roles, nil
	}

	// Check if search index is ready
	p.indexMu.RLock()
	rolesIndex := p.rolesIndex
	p.indexMu.RUnlock()

	if rolesIndex != nil {
		// Use Bleve search for better search capabilities
		return common.BleveListSearch(ctx, rolesIndex, func(a *search.DocumentMatch, b models.ProviderRole) bool {
			return strings.Compare(a.ID, b.Name) == 0
		}, p.roles, filters...)
	}

	// Fallback to simple substring filtering while index is being built
	var filtered []models.ProviderRole
	filterText := strings.ToLower(strings.Join(filters, " "))

	for _, role := range p.roles {
		// Check if any filter matches the role name, title or description
		if strings.Contains(strings.ToLower(role.Name), filterText) ||
			strings.Contains(strings.ToLower(role.Title), filterText) ||
			strings.Contains(strings.ToLower(role.Description), filterText) {
			filtered = append(filtered, role)
		}
	}

	return filtered, nil
}
