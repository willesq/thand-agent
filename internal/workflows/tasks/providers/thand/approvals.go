package thand

import (
	"errors"
	"fmt"
	"slices"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	runner "github.com/thand-io/agent/internal/workflows/runner"
	taskModel "github.com/thand-io/agent/internal/workflows/tasks/model"
)

var ThandApprovalsTask = "approvals"

func (t *thandTask) executeApprovalsTask(
	workflowTask *models.WorkflowTask,
	taskName string,
	call *taskModel.ThandTask,
	input any) (any, error) {

	elevationRequest, err := workflowTask.GetContextAsElevationRequest()

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"taskName": taskName,
		}).Error("Failed to get elevation request from context")

		return nil, err
	}

	var notifyReq NotifyRequest
	err = common.ConvertInterfaceToInterface(call.With, &notifyReq)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"taskName": taskName,
		}).Error("Failed to parse notification request")

		return nil, err
	}

	if !notifyReq.IsValid() {
		return nil, errors.New("invalid notification request")
	}

	if common.IsNilOrZero(input) {

		notifyReq.Entrypoint = taskName

		logrus.Infof("Starting Thand approvals task: %s", taskName)

		newConfig := &models.BasicConfig{}
		newConfig.Update(notifyReq.AsMap())

		call.With = newConfig

		// Set the context for the notification
		approvalNotifier := NewApprovalsNotifier(
			t.config,
			workflowTask,
			elevationRequest,
			&notifyReq,
		)

		// First lets notify the approvers
		_, err := t.executeNotify(
			workflowTask,
			fmt.Sprintf("%s.notify", taskName),
			approvalNotifier,
		)

		if err != nil {

			logrus.WithError(err).WithFields(logrus.Fields{
				"taskName": taskName,
			}).Error("Failed to notify approvers")

			return nil, err
		}

	} else {
		logrus.Infof("Resuming Thand approvals task: %s", taskName)
	}

	logrus.Infof("Executing Thand monitor task: %s", taskName)

	approval, err := runner.ListenTaskHandler(
		workflowTask, fmt.Sprintf("%s.listen", taskName), &model.ListenTask{
			Listen: model.ListenTaskConfiguration{
				To: &model.EventConsumptionStrategy{
					Any: []*model.EventFilter{
						{
							With: &model.EventProperties{
								Type: ThandApprovalEventType,
							},
						},
					},
				},
			},
		}, input)

	if err != nil {

		logrus.WithError(err).WithFields(logrus.Fields{
			"taskName": taskName,
		}).Error("Failed to listen for approval event")

		return nil, err
	}

	defaultFlowState := model.FlowDirective{
		Value: taskName, // loop back to await more approvals
	}

	// Set the context to hold all the approvals
	/*
		output:
			# Simply convert the output to a list of approvals
			as: '${ { "approvals": [{"approved": .data.approved}] } }'
		export:
			# Next we need to map the existing approvals to the new
			# list of approvals in the context as export handles
			# context access
			as: '${ $context + { "approvals": ($context.approvals // []) + .approvals } }'
		then: check_approval
	*/

	workflowContext := workflowTask.GetContextAsMap()

	approvals, ok := workflowContext["approvals"].(map[string]any)

	if !ok {
		approvals = map[string]any{}
	}

	var approvalData map[string]any

	if approvalEvent, ok := approval.(*cloudevents.Event); ok {

		approvalEvent.DataAs(&approvalData)
		extensions := approvalEvent.Extensions()

		userIdentity, userExists := extensions[models.VarsContextUser].(string)

		if !userExists {
			logrus.Warn("Approval event missing user extension")
			return &defaultFlowState, nil
		}

		// Check if self-approval is disabled and the approver is the requester or one of the elevated identities
		if !notifyReq.SelfApprove {
			requesterIdentity := elevationRequest.User.GetIdentity()

			// Check if approver is the requester
			if userIdentity == requesterIdentity {
				logrus.WithFields(logrus.Fields{
					"taskName":          taskName,
					"userIdentity":      userIdentity,
					"requesterIdentity": requesterIdentity,
				}).Warn("Self-approval is disabled; ignoring approval from requester")

				// Return to the default flow state to await more approvals
				return &defaultFlowState, nil
			}

			// Check if approver is one of the identities being elevated
			if slices.Contains(elevationRequest.Identities, userIdentity) {
				logrus.WithFields(logrus.Fields{
					"taskName":     taskName,
					"userIdentity": userIdentity,
				}).Warn("Self-approval is disabled; ignoring approval from identity being elevated")

				// Return to the default flow state to await more approvals
				return &defaultFlowState, nil
			}
		}

		approvedVal, exists := approvalData["approved"]

		if exists {

			approved, ok := approvedVal.(bool)

			if !ok {
				logrus.WithFields(logrus.Fields{
					"taskName":     taskName,
					"userIdentity": userIdentity,
				}).Warn("Approval value is not a boolean; ignoring this approval")
				return &defaultFlowState, nil
			}

			approvals[userIdentity] = map[string]any{
				"approved":  approved,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}

			// If the approval was denied then mark the approval as denied
			if !approved {

				logrus.WithFields(logrus.Fields{
					"taskName":     taskName,
					"userIdentity": userIdentity,
				}).Info("Approval denied by user")

				workflowTask.SetContextKeyValue(models.VarsContextApproved, false)
			}
		}
	}

	workflowTask.SetContextKeyValue("approvals", approvals)

	/*
		# If anyone rejects then reject the entire request
		# otherwise if the required number of approvals is met then authorize
		# Approvals are stored as a map[identity]approval_data structure
		- case1:
			when: any($context.approvals | to_entries[]; .value.approved == false)
			then: denied
		- case2:
			when: '[$context.approvals | to_entries[] | select(.value.approved == true)] | length >= N'
			then: authorize
		- default:
			then: loop back to task to await more approvals
	*/

	approvedState, foundApprovedState := call.On.GetString("approved")
	deniedState, foundDeniedState := call.On.GetString("denied")

	if !foundApprovedState || !foundDeniedState {
		return nil, errors.New("both approved and denied states must be specified in the on block")
	}

	// Create the switch task to handle approval or rejection
	flowDirective, err := t.evaluateApprovalSwitch(
		workflowTask,
		taskName,
		approvals,
		notifyReq.Approvals,
		approvedState,
		deniedState,
	)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"taskName": taskName,
		}).Error("Failed to execute switch task for approval logic")

		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"taskName":      taskName,
		"flowDirective": flowDirective.Value,
	}).Info("Completed Thand approvals task")

	return flowDirective, nil
}

// evaluateApprovalSwitch evaluates the approval logic using a switch task
// to determine if the request should be approved, denied, or loop back for more approvals
func (t *thandTask) evaluateApprovalSwitch(
	workflowTask *models.WorkflowTask,
	taskName string,
	approvals map[string]any,
	requiredApprovals int,
	approvedState string,
	deniedState string,
) (*model.FlowDirective, error) {

	return runner.SwitchTaskHandler(
		workflowTask,
		map[string]any{
			"approvals": approvals,
		},
		fmt.Sprintf("%s.switch", taskName),
		&model.SwitchTask{
			Switch: []model.SwitchItem{{
				"case1": model.SwitchCase{
					When: &model.RuntimeExpression{
						Value: "any($context.approvals | to_entries[]; .value.approved == false)",
					},
					Then: &model.FlowDirective{
						Value: deniedState, // go to denied state
					},
				},
			}, {
				"case2": model.SwitchCase{
					When: &model.RuntimeExpression{
						Value: fmt.Sprintf("[$context.approvals | to_entries[] | select(.value.approved == true)] | length >= %d", requiredApprovals),
					},
					Then: &model.FlowDirective{
						Value: approvedState, // proceed to the next state
					},
				},
			}, {
				"default": model.SwitchCase{
					// No When condition = default case (return to await more approvals)
					Then: &model.FlowDirective{
						Value: taskName, // loop back to await more approvals
					},
				},
			}},
		})
}
