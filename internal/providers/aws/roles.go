package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/data"
	"github.com/thand-io/agent/internal/models"
)

func (p *awsProvider) LoadRoles() error {
	data, err := getSharedData()
	if err != nil {
		return err
	}
	p.roles = data.roles
	p.rolesMap = data.rolesMap
	return nil
}

func loadRoles() ([]models.ProviderRole, map[string]*models.ProviderRole, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed AWS roles in %s", elapsed)
	}()

	// Get pre-parsed AWS roles from data package
	docs, err := data.GetParsedAwsRoles()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get parsed AWS roles: %w", err)
	}

	var roles []models.ProviderRole
	rolesMap := make(map[string]*models.ProviderRole, len(docs.Policies))

	// Convert to slice and create fast lookup map
	for _, policy := range docs.Policies {
		role := models.ProviderRole{
			Name: policy.Name,
		}
		roles = append(roles, role)
		rolesMap[strings.ToLower(policy.Name)] = &roles[len(roles)-1] // Reference to the slice element
	}

	return roles, rolesMap, nil
}

func (p *awsProvider) GetRole(ctx context.Context, role string) (*models.ProviderRole, error) {

	// If the role is a policy arn: arn:aws:iam::aws:policy/AdministratorAccess
	// Then parse the role and extract the policy name and convert it to a role
	role = strings.TrimPrefix(role, "arn:aws:iam::aws:policy/")
	role = strings.ToLower(role)

	// Fast map lookup
	if r, exists := p.rolesMap[role]; exists {
		return r, nil
	}
	return nil, fmt.Errorf("role not found")
}

func (p *awsProvider) ListRoles(ctx context.Context, filters ...string) ([]models.ProviderRole, error) {
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
		// Check if any filter matches the role name
		if strings.Contains(strings.ToLower(role.Name), filterText) {
			filtered = append(filtered, role)
		}
	}

	return filtered, nil
}
