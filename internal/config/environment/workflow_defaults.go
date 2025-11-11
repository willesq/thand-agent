package environment

import (
	_ "embed"

	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
)

// This file contains default environment setup for workflows

//go:embed aws/workflows.yaml
var defaultAWSWorkflowsYAML []byte

//go:embed gcp/workflows.yaml
var defaultGCPWorkflowsYAML []byte

//go:embed kubernetes/workflows.yaml
var defaultKubernetesWorkflowsYAML []byte

//go:embed local/workflows.yaml
var defaultLocalWorkflowsYAML []byte

// GetDefaultWorkflows returns default workflows based on the environment platform
func GetDefaultWorkflows(platform models.EnvironmentPlatform) ([]*models.WorkflowDefinitions, error) {
	defaults, err := defaultWorkflows(platform)

	if err != nil {
		return nil, err
	}

	return []*models.WorkflowDefinitions{
		defaults,
	}, nil
}

func defaultWorkflows(platform models.EnvironmentPlatform) (*models.WorkflowDefinitions, error) {
	exampleWorkflows := models.WorkflowDefinitions{}

	switch platform {
	case models.AWS:
		return common.ReadDataToInterface(
			defaultAWSWorkflowsYAML, exampleWorkflows)
	case models.GCP:
		return common.ReadDataToInterface(
			defaultGCPWorkflowsYAML, exampleWorkflows)
	case models.Kubernetes:
		return common.ReadDataToInterface(
			defaultKubernetesWorkflowsYAML, exampleWorkflows)
	case models.Local:
		fallthrough
	default:
		return common.ReadDataToInterface(
			defaultLocalWorkflowsYAML, exampleWorkflows)
	}
}
