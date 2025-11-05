package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	swctx "github.com/serverlessworkflow/sdk-go/v3/impl/ctx"
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	models "github.com/thand-io/agent/internal/models"
	thandModel "github.com/thand-io/agent/internal/workflows/tasks/model"
	thandTask "github.com/thand-io/agent/internal/workflows/tasks/providers/thand"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func (m *WorkflowManager) registerWorkflows() error {
	if !m.config.GetServices().HasTemporal() {
		return fmt.Errorf("temporal service not configured")
	}

	temporalService := m.config.GetServices().GetTemporal()
	if temporalService == nil {
		return fmt.Errorf("temporal service not available")
	}

	if !temporalService.HasWorker() {
		return fmt.Errorf("temporal worker not configured")
	}

	worker := temporalService.GetWorker()

	// Register the primary workflow with Pinned versioning behavior
	//
	// This ensures that existing workflows continue to run with the code they started with
	// preventing non-determinism errors when code changes occur
	worker.RegisterWorkflowWithOptions(
		m.createPrimaryWorkflowHandler(),
		workflow.RegisterOptions{
			Name:               models.TemporalExecuteElevationWorkflowName,
			VersioningBehavior: workflow.VersioningBehaviorPinned,
		},
	)

	return nil
}

// createPrimaryWorkflowHandler creates the main workflow handler function
func (m *WorkflowManager) createPrimaryWorkflowHandler() func(workflow.Context, *models.WorkflowTask) (*models.WorkflowTask, error) {
	return func(rootCtx workflow.Context, workflowTask *models.WorkflowTask) (outputTask *models.WorkflowTask, outputError error) {

		// Get workflow info including the BuildID set by the worker
		workflowInfo := workflow.GetInfo(rootCtx)
		logrus.WithFields(logrus.Fields{
			"WorkflowID": workflowTask.WorkflowID,
			"TaskName":   workflowTask.WorkflowName,
			"StartTime":  workflow.Now(rootCtx),
			"BuildID":    workflowInfo.GetCurrentBuildID(),
		}).Info("Primary workflow started.")

		cancelCtx, cancelHandler := workflow.WithCancel(rootCtx)

		// Variable to store termination request, accessible to both goroutine and defer
		var terminationRequest *models.TemporalTerminationRequest

		// Setup cleanup handler
		defer func() {

			// Handle workflow panic
			if r := recover(); r != nil {
				outputError = fmt.Errorf("workflow failed: %s", r)
				return
			}

			cleanupErr := m.runCleanup(rootCtx, workflowTask, terminationRequest)

			outputTask = workflowTask

			if cleanupErr != nil {
				logrus.WithError(cleanupErr).Error("Cleanup activity failed")
				outputError = cleanupErr
			} else if cancelCtx.Err() != nil && (errors.Is(cancelCtx.Err(), context.Canceled) || temporal.IsCanceledError(cancelCtx.Err())) {
				// Suppress cancellation errors - workflow completed normally
				outputError = nil
			}
			logrus.Info("Workflow cleanup completed.")

		}()

		// Setup query handler
		if err := m.setupIsApprovedQueryHandler(cancelCtx, workflowTask); err != nil {
			logrus.WithError(err).Error("Failed to set query handler")
			return nil, err
		}

		// Setup get workflow task query handler
		if err := m.setupGetWorkflowTaskQueryHandler(cancelCtx, workflowTask); err != nil {
			logrus.WithError(err).Error("Failed to set get workflow task query handler")
			return nil, err
		}
		// Setup signal channels and handlers
		resumeSignal, terminateSignal := m.setupSignalChannels(cancelCtx)
		m.setupTerminationHandler(rootCtx, terminateSignal, cancelHandler, &terminationRequest)

		// Setup workflow selector
		workflowSelector := m.setupWorkflowSelector(
			cancelCtx, resumeSignal, workflowTask)
		workflowSelector.Select(cancelCtx)

		logrus.Info("Starting main workflow execution loop")

		// Execute main workflow loop
		return m.executeWorkflowLoop(cancelCtx, workflowSelector, workflowTask)
	}
}

// runCleanup executes the cleanup activity and returns any cleanup-specific errors
func (m *WorkflowManager) runCleanup(
	rootCtx workflow.Context,
	workflowTask *models.WorkflowTask,
	terminationRequest *models.TemporalTerminationRequest,
) error {

	// Log termination request if present
	if terminationRequest != nil {
		logrus.WithFields(logrus.Fields{
			"Reason":      terminationRequest.Reason,
			"EntryPoint":  terminationRequest.EntryPoint,
			"ScheduledAt": terminationRequest.ScheduledAt,
		}).Info("Cleanup running with termination request")
	}

	if approved := workflowTask.IsApproved(); approved == nil || !*approved {
		logrus.Info("Workflow not approved, skipping cleanup activity.")
		return nil
	}

	// Check if a user or role is associated with the workflow
	elevationRequest, err := workflowTask.GetContextAsElevationRequest()
	if err != nil || !elevationRequest.IsValid() {
		logrus.Info("No valid elevation context found, skipping cleanup activity.")
		return nil
	}

	// Use a disconnected context for cleanup to ensure it runs even if workflow is cancelled
	newCtx, _ := workflow.NewDisconnectedContext(rootCtx)
	workflowTask = workflowTask.WithTemporalContext(newCtx)

	// If a termination request with entrypoint is provided then we need to use it
	if terminationRequest != nil && len(terminationRequest.EntryPoint) > 0 {

		// Resume the workflow task with the specified entrypoint
		workflowTask.SetEntrypoint(terminationRequest.EntryPoint)

		result, err := m.ResumeWorkflowTask(
			workflowTask.WithTemporalContext(newCtx),
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to resume workflow for cleanup with termination entrypoint")
			return err
		}

		logrus.WithFields(logrus.Fields{
			"Status": result.GetStatus(),
		}).Info("Workflow resumed for cleanup with termination entrypoint")

	} else {

		// Get the taskItem from the workflow spec or create a synthetic one
		revocationTask := &model.TaskItem{
			Key: "$cleanup",
			Task: &thandModel.ThandTask{
				Thand: thandTask.ThandRevokeTask,
				With:  nil,
			},
		}

		// Run the revocation task
		revokeTask, foundTask := m.tasks.GetTaskHandler(revocationTask)

		if !foundTask {
			logrus.Error("Failed to get revoke task handler for cleanup")
			return errors.New("failed to get revoke task handler for cleanup")
		}

		_, err = revokeTask.Execute(
			workflowTask,
			revocationTask,
			nil,
		)

		if err != nil {
			logrus.WithError(err).Error("Cleanup activity failed")
			return err
		}

	}

	logrus.Info("Cleanup completed successfully")
	return nil
}

// setupQueryHandler sets up the query handler for the workflow
func (m *WorkflowManager) setupIsApprovedQueryHandler(
	ctx workflow.Context, workflowTask *models.WorkflowTask) error {
	return workflow.SetQueryHandler(ctx, models.TemporalIsApprovedQueryName, func() (*bool, error) {
		logrus.WithFields(logrus.Fields{
			"WorkflowID": workflowTask.WorkflowID,
		}).Debug("IsApproved query received")
		return workflowTask.IsApproved(), nil
	})
}

func (m *WorkflowManager) setupGetWorkflowTaskQueryHandler(
	ctx workflow.Context, workflowTask *models.WorkflowTask) error {
	return workflow.SetQueryHandler(ctx, models.TemporalGetWorkflowTaskQueryName, func() (*models.WorkflowTask, error) {
		logrus.WithFields(logrus.Fields{
			"WorkflowID": workflowTask.WorkflowID,
		}).Debug("GetWorkflowTask query received")
		return workflowTask, nil
	})
}

// setupSignalChannels creates and returns the signal channels
func (m *WorkflowManager) setupSignalChannels(ctx workflow.Context) (workflow.ReceiveChannel, workflow.ReceiveChannel) {
	resumeSignal := workflow.GetSignalChannel(ctx, models.TemporalResumeSignalName)
	terminateSignal := workflow.GetSignalChannel(ctx, models.TemporalTerminateSignalName)
	return resumeSignal, terminateSignal
}

// setupTerminationHandler sets up the background termination handler
func (m *WorkflowManager) setupTerminationHandler(
	rootCtx workflow.Context,
	terminateSignal workflow.ReceiveChannel,
	cancelHandler workflow.CancelFunc,
	terminationRequest **models.TemporalTerminationRequest) {

	workflow.Go(rootCtx, func(ctx workflow.Context) {
		logrus.Info("Listening for terminate signal in background goroutine")

		terminateSelector := workflow.NewSelector(ctx)
		terminateSelector.AddReceive(terminateSignal, func(c workflow.ReceiveChannel, more bool) {
			var req models.TemporalTerminationRequest
			c.Receive(ctx, &req)
			*terminationRequest = &req
			logrus.Info("Terminate Signal Received")
		})

		terminateSelector.Select(ctx)

		if *terminationRequest != nil {
			m.handleTerminationRequest(ctx, *terminationRequest)
		}

		cancelHandler()
		logrus.Info("Workflow cancellation initiated due to terminate signal")
	})
}

// handleTerminationRequest processes the termination request
func (m *WorkflowManager) handleTerminationRequest(
	ctx workflow.Context,
	terminationRequest *models.TemporalTerminationRequest,
) {

	if terminationRequest == nil {
		logrus.Info("No termination request provided, skipping termination handling")
		return
	}

	logrus.WithFields(logrus.Fields{
		"reason":       terminationRequest.Reason,
		"EntryPoint":   terminationRequest.EntryPoint,
		"scheduled_at": terminationRequest.ScheduledAt,
	}).Info("Processing termination request")

	var timerDuration time.Duration
	if terminationRequest.ScheduledAt != nil && !terminationRequest.ScheduledAt.IsZero() {
		// Use workflow.Now() instead of time.Now() for deterministic time
		now := workflow.Now(ctx)
		delay := terminationRequest.ScheduledAt.Sub(now)
		timerDuration = max(delay, 0)
	}

	// New behavior: always create timer, but with minimum duration
	if timerDuration <= 0 {
		timerDuration = time.Nanosecond // Minimum timer duration
	}
	timer := workflow.NewTimer(ctx, timerDuration)
	timer.Get(ctx, nil)
	logrus.WithField("Duration", timerDuration).Info("Termination timer completed")

}

// setupWorkflowSelector creates and configures the workflow selector
func (m *WorkflowManager) setupWorkflowSelector(
	ctx workflow.Context,
	resumeSignal workflow.ReceiveChannel,
	workflowTask *models.WorkflowTask,
) workflow.Selector {
	workflowSelector := workflow.NewSelector(ctx)

	workflowSelector.AddReceive(resumeSignal, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &workflowTask)
		logrus.Info("Resume Signal Received")
	})

	workflowSelector.AddFuture(workflow.NewTimer(ctx, 0), func(f workflow.Future) {
		logrus.Debug("Timer triggered for context cancellation check")
		if ctx.Err() != nil {
			logrus.Debug("Context cancellation detected via timer")
		}
	})

	return workflowSelector
}

// shouldContinueAsNew checks if the workflow should perform Continue-As-New
// This allows upgrading to new worker versions and prevents event history size issues
func (m *WorkflowManager) shouldContinueAsNew(ctx workflow.Context) bool {
	// Check Temporal's built-in suggestion for Continue-As-New
	// This is triggered when event history approaches size limits
	if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
		logrus.Info("Continue-As-New suggested by Temporal (event history size)")
		return true
	}

	// TODO(hugh): Add custom Continue-As-New triggers here

	return false
}

// executeWorkflowLoop executes the main workflow execution loop
func (m *WorkflowManager) executeWorkflowLoop(
	cancelCtx workflow.Context,
	workflowSelector workflow.Selector,
	workflowTask *models.WorkflowTask,
) (*models.WorkflowTask, error) {

	for {

		logrus.Info("Waiting for signal...")

		// Check if we should Continue-As-New before waiting for signal
		// This allows upgrading to new worker versions at safe checkpoints
		if m.shouldContinueAsNew(cancelCtx) {
			currentBuildID := workflow.GetInfo(cancelCtx).GetCurrentBuildID()
			logrus.WithFields(logrus.Fields{
				"WorkflowID":     workflowTask.WorkflowID,
				"CurrentBuildID": currentBuildID,
			}).Info("Continue-As-New suggested, upgrading workflow to latest version")

			return workflowTask, workflow.NewContinueAsNewError(
				cancelCtx,
				models.TemporalExecuteElevationWorkflowName,
				workflowTask,
			)
		}

		if err := m.waitForSignal(cancelCtx, workflowSelector); err != nil {
			return nil, err
		}

		if cancelCtx.Err() != nil {
			if errors.Is(cancelCtx.Err(), context.Canceled) {
				logrus.Info("Workflow context cancelled, exiting main loop")
				break
			}
			logrus.WithError(cancelCtx.Err()).Error("Error while waiting for signal")
			return nil, cancelCtx.Err()
		}

		workflowSelector.Select(cancelCtx)

		if workflowTask == nil {
			continue
		}

		logrus.WithFields(logrus.Fields{
			"WorkflowID": workflowTask.WorkflowID,
			"Status":     workflowTask.GetStatus(),
		}).Info("Resuming ...")

		// Execute workflow step
		result, err := m.executeWorkflowStep(cancelCtx, workflowTask)

		// Check if the context was cancelled during execution
		if cancelCtx.Err() != nil {
			if errors.Is(cancelCtx.Err(), context.Canceled) {
				if result != nil {
					result.SetStatus(swctx.CancelledStatus)
				}
				logrus.Info("Workflow context cancelled during execution, exiting main loop")
				return result, nil
			}
			logrus.WithError(cancelCtx.Err()).Error("Error while executing workflow step")
			return result, cancelCtx.Err()
		}

		// If execution completed or failed, return the result
		if err != nil || (result != nil && result.GetStatus() != swctx.RunningStatus) {
			return result, err
		}

		// Continue loop for running workflows
		workflowTask = result
	}

	// Loop exited due to cancellation
	return workflowTask, nil
}

// waitForSignal waits for any signals to be available
func (m *WorkflowManager) waitForSignal(cancelCtx workflow.Context, workflowSelector workflow.Selector) error {
	return workflow.Await(cancelCtx, func() bool {
		if cancelCtx.Err() != nil {
			logrus.WithError(cancelCtx.Err()).Info("Context error")
			if errors.Is(cancelCtx.Err(), context.Canceled) {
				logrus.Info("Context was cancelled")
			}
			return true
		}

		pending := workflowSelector.HasPending()
		logrus.WithField("Pending", pending).Info("Signal pending")
		return pending
	})
}

// executeWorkflowStep executes a single workflow step and handles the result
func (m *WorkflowManager) executeWorkflowStep(
	ctx workflow.Context, workflowTask *models.WorkflowTask) (*models.WorkflowTask, error) {
	logrus.Info("Starting workflow execution")

	f := m.StartWorkflow(ctx, workflowTask)
	err := f.Get(ctx, &workflowTask)

	if err != nil {
		logrus.WithError(err).Error("Workflow execution failed")
		workflowTask.SetStatus(swctx.FaultedStatus)
		return workflowTask, err
	}

	logrus.WithField("Status", workflowTask.GetStatus()).Info("Workflow execution step completed")

	return m.handleWorkflowStatus(workflowTask)
}

// handleWorkflowStatus handles different workflow status cases
func (m *WorkflowManager) handleWorkflowStatus(workflowTask *models.WorkflowTask) (*models.WorkflowTask, error) {
	switch workflowTask.GetStatus() {
	case swctx.RunningStatus:
		logrus.WithFields(logrus.Fields{
			"workflow_id": workflowTask.WorkflowID,
			"task_name":   workflowTask.GetTaskName(),
		}).Info("Workflow is still running")
		return workflowTask, nil // Continue loop

	case swctx.CompletedStatus:
		logrus.Info("Workflow completed successfully.")
		return workflowTask, nil

	case swctx.FaultedStatus:
		logrus.Error("Workflow failed.")
		return workflowTask, fmt.Errorf("workflow failed")

	case swctx.WaitingStatus:
		logrus.Info("Workflow is waiting, pausing execution.")
		return workflowTask, nil

	case swctx.PendingStatus:
		logrus.Info("Workflow is pending, pausing execution.")
		return workflowTask, nil

	default:
		logrus.WithField("status", workflowTask.GetStatus()).Error("Workflow ended in unknown state")
		return workflowTask, fmt.Errorf("workflow ended in unknown state: %s", workflowTask.GetStatus())
	}
}

func (m *WorkflowManager) StartWorkflow(ctx workflow.Context, workflowTask *models.WorkflowTask) workflow.Future {

	logrus.WithField("WorkflowID", workflowTask.WorkflowID).Info("Starting workflow execution loop")

	future, settable := workflow.NewFuture(ctx)

	workflow.Go(ctx, func(ctx workflow.Context) {

		// Continue to resume the workflow until it is completed, faulted, or waiting
		// This loop allows us to handle the workflow execution in a single Temporal workflow run
		// and manage its state transitions effectively

		// Resume the workflow task
		result, err := m.ResumeWorkflowTask(
			workflowTask.WithTemporalContext(ctx),
		)

		settable.Set(result, err)

		logrus.WithField("Status", workflowTask.GetStatus()).Info("Workflow resumed")

	})

	return future

}
