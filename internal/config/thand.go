package config

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func (c *Config) KickStartThandSync(registration *RegistrationResponse) error {

	if !c.HasThandService() {
		return fmt.Errorf("thand service is not configured")
	}

	if c.GetServices() == nil || c.GetServices().GetTemporal() == nil {
		return fmt.Errorf("temporal service is not initialized")
	}

	temporalClient := c.servicesClient.GetTemporal()

	if temporalClient == nil || !temporalClient.HasClient() {
		return fmt.Errorf("temporal client is not initialized")
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        models.CreateTemporalWorkflowIdentifier("sync"),
		TaskQueue: temporalClient.GetTaskQueue(),
	}

	if !temporalClient.IsVersioningDisabled() {
		workflowOptions.VersioningOverride = &client.PinnedVersioningOverride{
			Version: worker.WorkerDeploymentVersion{
				DeploymentName: models.TemporalDeploymentName,
				BuildID:        common.GetBuildIdentifier(),
			},
		}
	}

	we, err := temporalClient.GetClient().ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		ThandSyncWorkflow,
		SystemSyncRequest{
			AgentIdentifier: common.GetClientIdentifier(),
		},
	)
	if err != nil {
		return fmt.Errorf("starting thand sync workflow: %w", err)
	}

	logrus.Infof("Started Thand sync workflow. WorkflowID: %s, RunID: %s", we.GetID(), we.GetRunID())

	return nil

}
