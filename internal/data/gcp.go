package data

import (
	_ "embed"
	"sync"

	"github.com/thand-io/agent/internal/data/iam-dataset/generated/gcp"
)

//go:embed iam-dataset/gcp/predefined_roles.fb
var gcpRolesFb []byte

type GcpPredefinedRole struct {
	Name        string
	Title       string
	Description string
	Stage       string
}

var (
	parsedGcpRoles []GcpPredefinedRole
	gcpRolesOnce   sync.Once
	gcpRolesErr    error
)

// GetParsedGcpRoles returns the pre-parsed GCP roles slice from FlatBuffer
func GetParsedGcpRoles() ([]GcpPredefinedRole, error) {
	gcpRolesOnce.Do(func() {
		// Parse FlatBuffer
		predefinedRolesList := gcp.GetRootAsPredefinedRolesList(gcpRolesFb, 0)

		// Extract roles - including Stage field needed for filtering
		for i := 0; i < predefinedRolesList.RolesLength(); i++ {
			var role gcp.PredefinedRole
			if predefinedRolesList.Roles(&role, i) {
				parsedGcpRoles = append(parsedGcpRoles, GcpPredefinedRole{
					Name:        string(role.Name()),
					Title:       string(role.Title()),
					Description: string(role.Description()),
					Stage:       string(role.Stage()),
				})
			}
		}
	})
	return parsedGcpRoles, gcpRolesErr
}
