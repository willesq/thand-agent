package thand

import (
	"fmt"
	"maps"
	"net/http"
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

	// Right we need to loop over all the role providers and identities in order to
	// authorize them all.

	authorizedAt := time.Now().UTC()
	revocationDate := authorizedAt.Add(duration)

	modelOutput := map[string]any{}

	for _, providerName := range elevateRequest.Providers {

		providerCall, err := t.config.GetProviderByName(providerName)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider: %w", err)
		}

		modelOutput, err := t.validateRoleAndBuildOutput(providerCall, *elevateRequest)
		if err != nil {
			return nil, err
		}

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

			logrus.WithFields(logrus.Fields{
				"user":     authReq.User.GetIdentity(),
				"role":     authReq.Role.GetName(),
				"provider": providerName,
				"duration": duration,
			}).Info("Executing authorization logic")

			maps.Copy(modelOutput, map[string]any{
				"authorized_at": authorizedAt.Format(time.RFC3339),
				"revocation_at": revocationDate.Format(time.RFC3339),
			})

			var authOut any

			if workflowTask.HasTemporalContext() {

				temporalContext := workflowTask.GetTemporalContext()

				serviceClient := t.config.GetServices()

				ao := workflow.ActivityOptions{
					TaskQueue:           serviceClient.GetTemporal().GetTaskQueue(),
					StartToCloseTimeout: time.Minute * 5,
				}
				aoctx := workflow.WithActivityOptions(temporalContext, ao)

				// Use Temporal activity to send notification
				err = workflow.ExecuteActivity(
					aoctx,
					thandFunction.ThandAuthorizeFunction,

					// args
					workflowTask,
					taskName,
					model.CallFunction{
						Call: thandFunction.ThandNotifyFunction,
						With: call.With,
					},
					thandAuthReq,
				).Get(temporalContext, &authOut)

				if err != nil {
					return nil, fmt.Errorf("failed to send notification: %w", err)
				}

			} else {

				authOut, err = providerCall.GetClient().AuthorizeRole(
					workflowTask.GetContext(), &authReq,
				)

				if err != nil {
					return nil, fmt.Errorf("failed to authorize user: %w", err)
				}

			}

			maps.Copy(modelOutput, map[string]any{
				"authorizations": map[string]any{
					elevateRequest.User.GetIdentity(): authOut,
				},
			})

			maps.Copy(modelOutput, map[string]any{
				models.VarsContextApproved: true,
			})

		}

	}

	// Schedule revocation if revocation state provided
	if err := t.scheduleRevocation(workflowTask, "", revocationDate); err != nil {
		logrus.WithError(err).Error("Failed to schedule revocation")
		return nil, fmt.Errorf("failed to schedule revocation: %w", err)
	}

	return modelOutput, nil
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

	// If we have a temporal client, we can use that to schedule the revocation
	if workflowTask.HasTemporalContext() {

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
			signalInput = newTask
		}

		workflow.SignalExternalWorkflow(
			workflowTask.GetTemporalContext(),
			workflowTask.WorkflowID,
			models.TemporalEmptyRunId,
			signalName,
			signalInput,
		)

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
