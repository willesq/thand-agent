package thand

import (
	"fmt"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/thand-io/agent/internal/models"
	"go.temporal.io/sdk/workflow"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	thandFunction "github.com/thand-io/agent/internal/workflows/functions/providers/thand"
	taskModel "github.com/thand-io/agent/internal/workflows/tasks/model"
)

const ThandAuthorizeTask = "authorize"

type ThandAuthorizeCallTask struct {
	Revocation string `json:"revocation"` // This is the state to request the revocation
}

func (t *ThandAuthorizeCallTask) HasRevocation() bool {
	return len(t.Revocation) > 0
}

func (t *thandTask) executeAuthorizeTask(
	workflowTask *models.WorkflowTask,
	taskName string,
	call *taskModel.ThandTask) (any, error) {

	elevateRequest, err := workflowTask.GetContextAsElevationRequest()

	if err != nil {
		return nil, err
	}

	isApproved := workflowTask.IsApproved()

	if isApproved != nil && *isApproved {
		modelOutput := t.buildBasicModelOutput(elevateRequest)
		return &modelOutput, nil
	}

	return t.executeAuthorization(workflowTask, taskName, call, elevateRequest)
}

// buildBasicModelOutput creates the basic model output with timestamps
func (t *thandTask) buildBasicModelOutput(elevateRequest *models.ElevateRequestInternal) map[string]any {
	duration, _ := elevateRequest.AsDuration()
	authorizedAt := time.Now().UTC()
	revocationDate := authorizedAt.Add(duration)

	return map[string]any{
		"authorized_at": authorizedAt.Format(time.RFC3339),
		"revocation_at": revocationDate.Format(time.RFC3339),
	}
}

// authResult holds the result of an authorization operation
type authResult struct {
	Identity string
	Output   any
	Error    error
}

// authTask represents an authorization task with all necessary context
type authTask struct {
	ProviderName string
	Identity     string
	AuthReq      models.AuthorizeRoleRequest
	ThandAuthReq thandFunction.ThandAuthorizeRequest
}

// temporalAuthResult represents the result of an authorization operation for temporal communication
type temporalAuthResult struct {
	Index    int
	Identity string
	Output   any
	Err      error
}

// executeAuthorization performs the main authorization workflow
func (t *thandTask) executeAuthorization(
	workflowTask *models.WorkflowTask,
	taskName string,
	call *taskModel.ThandTask,
	elevateRequest *models.ElevateRequestInternal,
) (any, error) {

	duration, err := elevateRequest.AsDuration()

	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	authorizedAt := time.Now().UTC()
	revocationDate := authorizedAt.Add(duration)

	modelOutput := map[string]any{
		"authorized_at": authorizedAt.Format(time.RFC3339),
		"revocation_at": revocationDate.Format(time.RFC3339),
	}

	// Collect all authorization tasks
	var authTasks []authTask

	for _, providerName := range elevateRequest.Providers {

		providerCall, err := t.config.GetProviderByName(providerName)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider: %w", err)
		}

		validateOutput, err := t.validateRoleAndBuildOutput(providerCall, *elevateRequest)
		if err != nil {
			return nil, err
		}

		maps.Copy(modelOutput, validateOutput)

		for _, identity := range elevateRequest.Identities {

			authReq := models.AuthorizeRoleRequest{
				RoleRequest: &models.RoleRequest{
					User: &models.User{
						Email:  identity,
						Source: "thand",
					},
					Role:     elevateRequest.Role,
					Duration: &duration,
				},
			}

			thandAuthReq := thandFunction.ThandAuthorizeRequest{
				AuthorizeRoleRequest: authReq,
				Provider:             providerName,
			}

			authTasks = append(authTasks, authTask{
				ProviderName: providerName,
				Identity:     identity,
				AuthReq:      authReq,
				ThandAuthReq: thandAuthReq,
			})

			logrus.WithFields(logrus.Fields{
				"user":     authReq.User.GetIdentity(),
				"role":     authReq.Role.GetName(),
				"provider": providerName,
				"duration": duration,
			}).Info("Preparing authorization logic")
		}
	}

	var authResults []authResult

	if workflowTask.HasTemporalContext() {
		authResults, err = t.executeTemporalParallel(workflowTask, taskName, call, authTasks)
	} else {
		authResults, err = t.executeGoParallel(workflowTask, authTasks)
	}

	if err != nil {

		logrus.WithError(err).Error("Failed to execute authorization tasks")
		return nil, err

	}

	// Process results
	authorizations := make(map[string]any)
	hasErrors := false

	for _, result := range authResults {
		if result.Error != nil {
			logrus.WithError(result.Error).WithField("identity", result.Identity).Error("Authorization failed")
			hasErrors = true
			continue
		}
		authorizations[result.Identity] = result.Output
	}

	if hasErrors && len(authorizations) == 0 {
		return nil, fmt.Errorf("all authorization requests failed")
	}

	// Schedule revocation if revocation state provided
	if err := t.scheduleRevocation(workflowTask, "", revocationDate); err != nil {
		logrus.WithError(err).Error("Failed to schedule revocation")
		return nil, fmt.Errorf("failed to schedule revocation: %w", err)
	}

	workflowTask.SetContextKeyValue(models.VarsContextApproved, true)
	workflowTask.SetContextKeyValue("authorizations", authorizations)

	return modelOutput, nil
}

// executeTemporalParallel executes authorization tasks in parallel using Temporal
func (t *thandTask) executeTemporalParallel(
	workflowTask *models.WorkflowTask,
	taskName string,
	call *taskModel.ThandTask,
	authTasks []authTask,
) ([]authResult, error) {

	temporalContext := workflowTask.GetTemporalContext()
	serviceClient := t.config.GetServices()

	ao := workflow.ActivityOptions{
		TaskQueue:           serviceClient.GetTemporal().GetTaskQueue(),
		StartToCloseTimeout: time.Minute * 5,
	}
	aoctx := workflow.WithActivityOptions(temporalContext, ao)

	// Create channel and results slice
	results := make([]authResult, len(authTasks))
	resultCh := workflow.NewChannel(temporalContext)

	// Start all tasks in parallel using workflow.Go
	for i, task := range authTasks {
		taskIndex := i
		authTask := task

		workflow.Go(temporalContext, func(ctx workflow.Context) {
			var authOut any
			err := workflow.ExecuteActivity(
				aoctx,
				thandFunction.ThandAuthorizeFunction,
				workflowTask,
				taskName,
				model.CallFunction{
					Call: thandFunction.ThandNotifyFunction,
					With: call.With.AsMap(),
				},
				authTask.ThandAuthReq,
			).Get(ctx, &authOut)

			// Send result through channel
			resultCh.Send(ctx, temporalAuthResult{
				Index:    taskIndex,
				Identity: authTask.Identity,
				Output:   authOut,
				Err:      err,
			})
		})
	}

	// Collect all results
	for range authTasks {
		var result temporalAuthResult
		resultCh.Receive(temporalContext, &result)
		results[result.Index] = authResult{
			Identity: result.Identity,
			Output:   result.Output,
			Error:    result.Err,
		}
	}

	return results, nil
}

// executeGoParallel executes authorization tasks in parallel using Go routines and WaitGroup
func (t *thandTask) executeGoParallel(
	workflowTask *models.WorkflowTask,
	authTasks []authTask,
) ([]authResult, error) {

	results := make([]authResult, len(authTasks))
	var wg sync.WaitGroup

	for i, task := range authTasks {
		wg.Add(1)
		go func(index int, authTask authTask) {
			defer wg.Done()

			providerCall, err := t.config.GetProviderByName(authTask.ProviderName)
			if err != nil {
				results[index] = authResult{
					Identity: authTask.Identity,
					Output:   nil,
					Error:    fmt.Errorf("failed to get provider: %w", err),
				}
				return
			}

			authOut, err := providerCall.GetClient().AuthorizeRole(
				workflowTask.GetContext(), &authTask.AuthReq,
			)

			results[index] = authResult{
				Identity: authTask.Identity,
				Output:   authOut,
				Error:    err,
			}
		}(i, task)
	}

	wg.Wait()

	return results, nil
}

// validateRoleAndBuildOutput validates the role and builds the initial model output
func (t *thandTask) validateRoleAndBuildOutput(
	providerCall *models.Provider,
	elevateRequest models.ElevateRequestInternal,
) (map[string]any, error) {
	modelOutput := map[string]any{}

	validateOut, err := models.ValidateRole(providerCall.GetClient(), elevateRequest)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"role":  elevateRequest.Role,
		}).Error("Failed to validate role")
		return nil, err
	}

	if len(validateOut) > 0 {
		maps.Copy(modelOutput, validateOut)
	}

	return modelOutput, nil
}

func (t *thandTask) GetExport() *model.Export {
	return &model.Export{
		As: model.NewObjectOrRuntimeExpr(
			model.RuntimeExpression{
				Value: "${ $context + . }",
			},
		),
	}
}

// Add to your function
func (t *thandTask) scheduleRevocation(
	workflowTask *models.WorkflowTask,
	revocationTask string,
	revocationAt time.Time,
) error {

	newTask := workflowTask.Clone().(*models.WorkflowTask)
	newTask.SetEntrypoint(revocationTask)

	serviceClient := t.config.GetServices()

	// If we have a temporal client, we can use that to schedule the revocation
	if serviceClient.HasTemporal() && serviceClient.GetTemporal().HasClient() {

		signalName := models.TemporalResumeSignalName
		var signalInput any

		// If the user has not provided a revocation task, we just terminate
		if len(revocationTask) == 0 {
			signalName = models.TemporalTerminateSignalName
			signalInput = models.TemporalTerminationRequest{
				Reason:      "No revocation state provided",
				ScheduledAt: revocationAt,
			}
		} else {
			// Otherwise send the new task as the signal input to resume the workflow
			// and set an execution timeout
			// TODO: Fiigure out how to delay the signal until the revocation time
			signalInput = newTask
		}

		temporalClient := serviceClient.GetTemporal().GetClient()

		err := temporalClient.SignalWorkflow(
			workflowTask.GetContext(),
			workflowTask.WorkflowID,
			models.TemporalEmptyRunId,
			signalName,
			signalInput,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to signal workflow for revocation")
			return fmt.Errorf("failed to signal workflow: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"task": newTask.GetTaskName(),
			"url":  t.config.GetResumeCallbackUrl(newTask),
		}).Info("Scheduled revocation via Temporal")

	} else if t.config.GetServices().HasScheduler() {

		err := t.config.GetServices().GetScheduler().AddJob(
			models.NewAtJob(
				revocationAt,
				func() {

					// Make call to revoke the user
					callingUrl := t.config.GetResumeCallbackUrl(newTask)

					logrus.WithFields(logrus.Fields{
						"task": newTask.GetTaskName(),
						"url":  callingUrl,
					}).Info("Executing scheduled revocation")

					response, err := common.InvokeHttpRequest(&model.HTTPArguments{
						Method: http.MethodGet,
						Endpoint: &model.Endpoint{
							URITemplate: &model.LiteralUri{
								Value: callingUrl,
							},
						},
					})

					if err != nil {
						logrus.WithError(err).Error("Failed to call revoke endpoint")
						return
					}

					if response.StatusCode() != http.StatusOK {
						logrus.WithFields(logrus.Fields{
							"status_code": response.StatusCode(),
							"body":        response.Body(),
						}).Error("Revoke endpoint returned non-200 status")
						return
					}

					logrus.WithFields(logrus.Fields{
						"revocation_task": newTask.GetTaskName(),
						"workflow":        workflowTask,
					}).Info("Scheduled revocation")

				},
			),
		)

		if err != nil {
			return fmt.Errorf("failed to schedule revocation: %w", err)
		}

	} else {

		logrus.Error("No scheduler available to schedule revocation")
		return fmt.Errorf("no scheduler available to schedule revocation")

	}

	return nil

}
