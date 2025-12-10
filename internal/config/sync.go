package config

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/models"
)

type configDiff struct {
	RoleConfig     *RoleConfig     `json:"roles,omitempty"`
	WorkflowConfig *WorkflowConfig `json:"workflows,omitempty"`
	ProviderConfig *ProviderConfig `json:"providers,omitempty"`
}

func (c *Config) MergeConfiguration(config *RegistrationResponse) error {

	incoming := configDiff{
		RoleConfig:     config.Roles,
		WorkflowConfig: config.Workflows,
		ProviderConfig: config.Providers,
	}

	incomingData, err := json.Marshal(incoming)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to marshal incoming configuration for diffing")
		return err
	}

	roles := c.GetRoles()
	workflows := c.GetWorkflows()
	providers := c.GetProviders()

	existing := configDiff{
		RoleConfig:     &roles,
		WorkflowConfig: &workflows,
		ProviderConfig: &providers,
	}

	existingData, err := json.Marshal(existing)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to marshal existing configuration for diffing")
		return err
	}

	// Apply the incoming changes over the existing configurations
	newData, err := jsonpatch.MergePatch(existingData, incomingData)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to create merge patch for configuration diffing")
		return err
	}

	// Create a patch to see the differences between existing and new
	incomingPatch, err := jsonpatch.CreateMergePatch(existingData, newData)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to create merge patch for configuration diffing")
		return err
	}

	// Convert patches back to structs
	var incomingDiff configDiff
	err = json.Unmarshal(incomingPatch, &incomingDiff)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to unmarshal incoming patch")
		return err
	}

	err = c.applyPatch(incomingDiff)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to apply incoming configuration patch")
		return err
	}

	// The incoming patch is what needs to be applied to the existing configuration

	// However, if we have outgoing changes then we need to update the remove server
	// with these changes.

	return nil

}

func (c *Config) applyPatch(diff configDiff) error {
	// Apply role changes
	if diff.RoleConfig != nil {
		err := c.updateRoles(diff.RoleConfig)
		if err != nil {
			logrus.WithError(err).Errorln("Failed to apply role configuration patch")
			return err
		}
	}

	// Apply workflow changes
	if diff.WorkflowConfig != nil {
		err := c.updateWorkflows(diff.WorkflowConfig)
		if err != nil {
			logrus.WithError(err).Errorln("Failed to apply workflow configuration patch")
			return err
		}
	}

	// Apply provider changes
	if diff.ProviderConfig != nil {
		err := c.updateProviders(diff.ProviderConfig)
		if err != nil {
			logrus.WithError(err).Errorln("Failed to apply provider configuration patch")
			return err
		}
	}

	return nil
}

func (c *Config) updateRoles(roleConfig *RoleConfig) error {
	_, err := c.ApplyRoles([]*models.RoleDefinitions{{
		Roles: roleConfig.Definitions,
	}})
	return err
}

func (c *Config) updateWorkflows(workflowConfig *WorkflowConfig) error {
	_, err := c.ApplyWorkflows([]*models.WorkflowDefinitions{{
		Workflows: workflowConfig.Definitions,
	}})
	return err
}

func (c *Config) updateProviders(providerConfig *ProviderConfig) error {
	_, err := c.ApplyProviders([]*models.ProviderDefinitions{{
		Providers: providerConfig.Definitions,
	}})
	return err
}
