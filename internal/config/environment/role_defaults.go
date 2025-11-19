package environment

import (
	_ "embed"

	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// This file contains default environment setup for roles

//go:embed aws/roles.yaml
var defaultAWSRolesYAML []byte

//go:embed gcp/roles.yaml
var defaultGCPRolesYAML []byte

//go:embed kubernetes/roles.yaml
var defaultKubernetesRolesYAML []byte

//go:embed local/roles.yaml
var defaultLocalRolesYAML []byte

// GetDefaultRoles returns default roles based on the environment platform
func GetDefaultRoles(platform models.EnvironmentPlatform) ([]*models.RoleDefinitions, error) {
	defaults, err := defaultRoles(platform)

	if err != nil {
		return nil, err
	}

	return []*models.RoleDefinitions{
		defaults,
	}, nil
}

func defaultRoles(platform models.EnvironmentPlatform) (*models.RoleDefinitions, error) {
	exampleRoles := models.RoleDefinitions{}

	switch platform {
	case models.AWS:
		return common.ReadDataToInterface(
			defaultAWSRolesYAML, exampleRoles)
	case models.GCP:
		return common.ReadDataToInterface(
			defaultGCPRolesYAML, exampleRoles)
	case models.Kubernetes:
		return common.ReadDataToInterface(
			defaultKubernetesRolesYAML, exampleRoles)
	case models.Local:
		fallthrough
	default:
		return common.ReadDataToInterface(
			defaultLocalRolesYAML, exampleRoles)
	}
}
