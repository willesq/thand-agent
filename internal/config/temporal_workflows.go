package config

import (
	"time"

	"github.com/google/uuid"
	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	// Thresholds for ContinueAsNew
	MaxHistoryEvents = 1000
)

type SystemSyncRequest struct {
	AgentIdentifier uuid.UUID
}

// SystemSyncState represents the current state of the sync workflow
type SystemSyncState struct {
	LastSyncTime time.Time
	Status       string
}

// SystemSyncWorkflow connects to thand.io upstream service to synchronize system configuration
// including roles, workflows, and providers. It handles local updates via signals
// and pushes them upstream, while also applying remote updates received from upstream.
func ThandSyncWorkflow(ctx workflow.Context, req SystemSyncRequest) error {

	// Upon starting the workflow, will onboard with the upstream service and
	// synchronize initial state and check for updates and syncronize if singaled.

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting System Sync Workflow", "AgentIdentifier", req.AgentIdentifier)

	lao := workflow.LocalActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    100 * time.Second,
			MaximumAttempts:    10,
		},
	}

	ctx = workflow.WithLocalActivityOptions(ctx, lao)

	// Phase 1: Initial Sync (Chunked)
	// Paginate through local configuration to avoid Temporal history limits (2MB)
	// We send each chunk upstream immediately instead of buffering the entire initial state.

	// Start Sync Session
	var startResp models.SynchronizeStartResponse
	err := workflow.ExecuteLocalActivity(ctx, "SynchronizeThandStart", req.AgentIdentifier.String()).Get(ctx, &startResp)
	if err != nil {
		logger.Error("Failed to start sync session", "error", err)
		return err
	}

	var initialSyncCursor *models.ConfigurationCursor
	for {
		var result struct {
			Chunk      models.SystemChunk
			NextCursor *models.ConfigurationCursor
		}

		// Fetch chunk
		// Note: "GetLocalConfigurationChunk" must be registered in your worker
		err := workflow.ExecuteLocalActivity(ctx, "GetLocalConfigurationChunk", initialSyncCursor).Get(ctx, &result)
		if err != nil {
			logger.Error("Failed to get local configuration chunk", "error", err)
			break
		}

		// Sync chunk if not empty
		if hasUpdates(result.Chunk) {
			err = workflow.ExecuteLocalActivity(ctx, "SynchronizeThandChunk", req.AgentIdentifier.String(), startResp.WorkflowID, result.Chunk).Get(ctx, nil)
			if err != nil {
				logger.Error("Failed to sync initial chunk", "error", err)
			}
		}

		if result.NextCursor == nil {
			break
		}
		initialSyncCursor = result.NextCursor
	}

	// Commit Sync Session
	err = workflow.ExecuteLocalActivity(ctx, "SynchronizeThandCommit", req.AgentIdentifier.String(), startResp.WorkflowID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to commit sync session", "error", err)
	}
	return nil
}

func hasUpdates(chunk models.SystemChunk) bool {
	return len(chunk.Roles) > 0 ||
		len(chunk.Workflows) > 0 ||
		len(chunk.Providers) > 0 ||
		len(chunk.ProviderData) > 0
}
