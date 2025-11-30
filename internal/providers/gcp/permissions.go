package gcp

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2/search"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// Get embedded permissions.json can't use data package as it uses Bleve indexing
// as it does not include the stage information or
// whether the permission is deprecated etc

//go:embed permissions.json
var gcpPermissions []byte

func GetGcpPermissions() []byte {
	return gcpPermissions
}

/*
"apiDisabled": true,
"description": "Create annotation specs",
"name": "aiplatform.annotationSpecs.create",
"stage": "GA",
"title": "Create annotation specs"
*/
type gcpPermissionMap []struct {
	ApiDisabled           bool   `json:"apiDisabled,omitempty"`
	Description           string `json:"description,omitempty"`
	Name                  string `json:"name,omitempty"`
	Stage                 string `json:"stage,omitempty"`
	Title                 string `json:"title,omitempty"`
	OnlyInPredefinedRoles bool   `json:"onlyInPredefinedRoles,omitempty"`
}

func (p *gcpProvider) LoadPermissions(stage string) error {
	data, err := getSharedData(stage)
	if err != nil {
		return err
	}
	p.permissions = data.permissions
	p.permissionsMap = data.permissionsMap
	return nil
}

func loadPermissions(stage string) ([]models.ProviderPermission, map[string]*models.ProviderPermission, error) {
	var permissionMap gcpPermissionMap

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed GCP permissions in %s", elapsed)
	}()

	// Load GCP Permissions
	if err := json.Unmarshal(GetGcpPermissions(), &permissionMap); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal GCP permissions: %w", err)
	}

	var permissions = make([]models.ProviderPermission, 0, len(permissionMap))
	permissionsMap := make(map[string]*models.ProviderPermission)

	if len(stage) == 0 {
		stage = DefaultStage
	}

	for _, perm := range permissionMap {

		if perm.OnlyInPredefinedRoles {
			continue
		}

		if strings.Compare(perm.Stage, stage) != 0 {
			continue
		}

		permission := models.ProviderPermission{
			Name:        perm.Name,
			Title:       perm.Title,
			Description: perm.Description,
		}
		permissions = append(permissions, permission)
		permissionsMap[strings.ToLower(perm.Name)] = &permissions[len(permissions)-1] // Reference to the slice element
	}

	return permissions, permissionsMap, nil
}

func (p *gcpProvider) GetPermission(ctx context.Context, permission string) (*models.ProviderPermission, error) {
	permission = strings.ToLower(permission)
	// Fast map lookup
	if perm, exists := p.permissionsMap[permission]; exists {
		return perm, nil
	}
	return nil, fmt.Errorf("GCP permission not found: %s", permission)
}

func (p *gcpProvider) ListPermissions(ctx context.Context, filters ...string) ([]models.ProviderPermission, error) {
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
		// Check if any filter matches the permission name, title or description
		if strings.Contains(strings.ToLower(perm.Name), filterText) ||
			strings.Contains(strings.ToLower(perm.Title), filterText) ||
			strings.Contains(strings.ToLower(perm.Description), filterText) {
			filtered = append(filtered, perm)
		}
	}

	return filtered, nil
}
