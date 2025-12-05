package gcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/data"
	"github.com/thand-io/agent/internal/models"
)

//go:embed permissions.json
var gcpPermissions []byte

func GetGcpPermissions() []byte {
	return gcpPermissions
}

type gcpData struct {
	permissions []models.ProviderPermission
	roles       []models.ProviderRole
}

type gcpSingleton struct {
	data *gcpData
	err  error
	once sync.Once
}

var (
	sharedDataMap = make(map[string]*gcpSingleton)
	sharedDataMu  sync.Mutex
)

func getSharedData(stage string) (*gcpData, error) {
	sharedDataMu.Lock()
	singleton, ok := sharedDataMap[stage]
	if !ok {
		singleton = &gcpSingleton{}
		sharedDataMap[stage] = singleton
	}
	sharedDataMu.Unlock()

	singleton.once.Do(func() {
		data := &gcpData{}
		var err error

		data.permissions, err = loadPermissions(stage)
		if err != nil {
			singleton.err = err
			return
		}

		data.roles, err = loadRoles(stage)
		if err != nil {
			singleton.err = err
			return
		}

		singleton.data = data
	})

	return singleton.data, singleton.err
}

type gcpPermissionMap []struct {
	ApiDisabled           bool   `json:"apiDisabled,omitempty"`
	Description           string `json:"description,omitempty"`
	Name                  string `json:"name,omitempty"`
	Stage                 string `json:"stage,omitempty"`
	Title                 string `json:"title,omitempty"`
	OnlyInPredefinedRoles bool   `json:"onlyInPredefinedRoles,omitempty"`
}

func loadPermissions(stage string) ([]models.ProviderPermission, error) {
	var permissionMap gcpPermissionMap

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed GCP permissions in %s", elapsed)
	}()

	// Load GCP Permissions
	if err := json.Unmarshal(GetGcpPermissions(), &permissionMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GCP permissions: %w", err)
	}

	var permissions = make([]models.ProviderPermission, 0, len(permissionMap))

	if len(stage) == 0 {
		stage = DefaultStage
	}

	for _, perm := range permissionMap {

		if perm.OnlyInPredefinedRoles {
			continue
		}

		if !strings.EqualFold(perm.Stage, stage) {
			continue
		}

		permission := models.ProviderPermission{
			Name:        perm.Name,
			Title:       perm.Title,
			Description: perm.Description,
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func loadRoles(stage string) ([]models.ProviderRole, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logrus.Debugf("Parsed GCP roles in %s", elapsed)
	}()

	// Get pre-parsed GCP roles from data package
	predefinedRoles, err := data.GetParsedGcpRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get parsed GCP roles: %w", err)
	}

	var roles = make([]models.ProviderRole, 0, len(predefinedRoles))

	if len(stage) == 0 {
		stage = DefaultStage
	}

	for _, gcpRole := range predefinedRoles {

		if !strings.EqualFold(gcpRole.Stage, stage) {
			continue
		}

		role := models.ProviderRole{
			Name:        gcpRole.Name,
			Title:       gcpRole.Title,
			Description: gcpRole.Description,
		}
		roles = append(roles, role)
	}

	return roles, nil
}
