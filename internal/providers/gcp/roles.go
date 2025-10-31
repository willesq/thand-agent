package gcp

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

func (p *gcpProvider) LoadRoles(stage string) error {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed GCP roles in %s", elapsed)
	}()

	// Get pre-parsed GCP roles from data package
	predefinedRoles, err := data.GetParsedGcpRoles()
	if err != nil {
		return fmt.Errorf("failed to get parsed GCP roles: %w", err)
	}

	var roles = make([]models.ProviderRole, 0, len(predefinedRoles))
	rolesMap := make(map[string]*models.ProviderRole)

	if len(stage) == 0 {
		stage = DefaultStage
	}

	for _, gcpRole := range predefinedRoles {

		if strings.Compare(gcpRole.Stage, stage) != 0 {
			continue
		}

		role := models.ProviderRole{
			Name:        gcpRole.Name,
			Title:       gcpRole.Title,
			Description: gcpRole.Description,
		}
		roles = append(roles, role)
		rolesMap[strings.ToLower(gcpRole.Name)] = &roles[len(roles)-1] // Reference to the slice element
	}

	p.roles = roles
	p.rolesMap = rolesMap

	logrus.WithFields(logrus.Fields{
		"roles": len(roles),
	}).Debug("Loaded GCP roles, building search index in background")

	return nil
}

func (p *gcpProvider) GetRole(ctx context.Context, role string) (*models.ProviderRole, error) {
	role = strings.ToLower(role)
	// Fast map lookup
	if r, exists := p.rolesMap[role]; exists {
		return r, nil
	}
	return nil, fmt.Errorf("role not found")
}

func (p *gcpProvider) ListRoles(ctx context.Context, filters ...string) ([]models.ProviderRole, error) {
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
