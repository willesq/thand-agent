package environment

import (
	_ "embed"

	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// This file contains default environment setup for cloud providers

//go:embed aws/providers.yaml
var defaultAWSProvidersYAML []byte

//go:embed gcp/providers.yaml
var defaultGCPProvidersYAML []byte

//go:embed kubernetes/providers.yaml
var defaultKubernetesProvidersYAML []byte

//go:embed local/providers.yaml
var defaultLocalProvidersYAML []byte

// GetDefaultProviders returns default providers based on the environment platform
func GetDefaultProviders(platform models.EnvironmentPlatform) ([]*models.ProviderDefinitions, error) {
	defaults, err := defaultProviders(platform)

	if err != nil {
		return nil, err
	}

	return []*models.ProviderDefinitions{
		defaults,
	}, nil
}

func defaultProviders(platform models.EnvironmentPlatform) (*models.ProviderDefinitions, error) {
	exampleProviders := models.ProviderDefinitions{}

	switch platform {
	case models.AWS:
		return common.ReadDataToInterface(
			defaultAWSProvidersYAML, exampleProviders)
	case models.GCP:
		return common.ReadDataToInterface(
			defaultGCPProvidersYAML, exampleProviders)
	case models.Kubernetes:
		return common.ReadDataToInterface(
			defaultKubernetesProvidersYAML, exampleProviders)
	case models.Local:
		fallthrough
	default:
		return common.ReadDataToInterface(
			defaultLocalProvidersYAML, exampleProviders)
	}
}
