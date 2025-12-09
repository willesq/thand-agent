package config

import (
	"fmt"

	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/workflow"
)

func (c *Config) setupTemporalServices() error {

	if !c.IsServer() {
		return fmt.Errorf("temporal services can only be set up in server mode")
	}

	// Register workflows
	err := c.registerTemporalWorkflows()
	if err != nil {
		return fmt.Errorf("registering temporal workflows: %w", err)
	}

	// Register activities
	err = c.registerTemporalActivities()
	if err != nil {
		return fmt.Errorf("registering temporal activities: %w", err)
	}

	return nil
}

// Register temporal workflows and activities
func (c *Config) registerTemporalWorkflows() error {

	if c.GetServices() == nil || c.GetServices().GetTemporal() == nil {
		return fmt.Errorf("temporal service is not initialized")
	}

	temporalWorker := c.servicesClient.GetTemporal().GetWorker()

	if temporalWorker == nil {
		return fmt.Errorf("temporal worker is not initialized")
	}

	if c.HasThandService() {

		// Register system sync workflow
		temporalWorker.RegisterWorkflowWithOptions(
			ThandSyncWorkflow,
			workflow.RegisterOptions{
				Name: models.ThandSyncWorkflowName,
			},
		)
	}

	return nil

}

func (c *Config) registerTemporalActivities() error {

	if c.GetServices() == nil || c.GetServices().GetTemporal() == nil {
		return fmt.Errorf("temporal service is not initialized")
	}

	temporalWorker := c.servicesClient.GetTemporal().GetWorker()

	if temporalWorker == nil {
		return fmt.Errorf("temporal worker is not initialized")
	}

	if c.HasThandService() {

		// Register system activities
		systemActivities := &ThandActivities{
			Config: c,
		}

		temporalWorker.RegisterActivity(systemActivities.GetLocalConfigurationChunk)
		temporalWorker.RegisterActivity(systemActivities.SynchronizeThandStart)
		temporalWorker.RegisterActivity(systemActivities.SynchronizeThandChunk)
		temporalWorker.RegisterActivity(systemActivities.SynchronizeThandCommit)

	}

	return nil

}
