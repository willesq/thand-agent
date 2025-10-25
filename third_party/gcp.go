package third_party

import (
	"encoding/json"
	"sync"
)

type GcpPredefinedRole struct {
	Description string `json:"description"`
	Etag        string `json:"etag"`
	Name        string `json:"name"`
	Stage       string `json:"stage"`
	Title       string `json:"title"`
}

type GcpPermission struct {
	ApiDisabled           bool   `json:"apiDisabled,omitempty"`
	Description           string `json:"description,omitempty"`
	Name                  string `json:"name,omitempty"`
	Stage                 string `json:"stage,omitempty"`
	Title                 string `json:"title,omitempty"`
	OnlyInPredefinedRoles bool   `json:"onlyInPredefinedRoles,omitempty"`
}

type GcpPermissionMap []GcpPermission

var (
	parsedGcpRoles []GcpPredefinedRole
	gcpRolesOnce   sync.Once
	gcpRolesErr    error
)

var (
	parsedGcpPermissions GcpPermissionMap
	gcpPermissionsOnce   sync.Once
	gcpPermissionsErr    error
)

// GetParsedGcpRoles returns the pre-parsed GCP roles slice, parsing it once on first call
func GetParsedGcpRoles() ([]GcpPredefinedRole, error) {
	gcpRolesOnce.Do(func() {
		gcpRolesErr = json.Unmarshal(gcpRoles, &parsedGcpRoles)
	})
	return parsedGcpRoles, gcpRolesErr
}

// GetParsedGcpPermissions returns the pre-parsed GCP permissions map, parsing it once on first call
func GetParsedGcpPermissions() (GcpPermissionMap, error) {
	gcpPermissionsOnce.Do(func() {
		gcpPermissionsErr = json.Unmarshal(gcpPermissions, &parsedGcpPermissions)
	})
	return parsedGcpPermissions, gcpPermissionsErr
}
