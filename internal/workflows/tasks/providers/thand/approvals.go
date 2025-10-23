package thand

import (
	"errors"
	"fmt"

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

	if common.IsNilOrZero(input) {

		var notifyReq NotifyRequest
		common.ConvertInterfaceToInterface(call.With, &notifyReq)

		if !notifyReq.IsValid() {
			return nil, errors.New("invalid notification request")
		}

		// Force approvals to be true
		notifyReq.Notifier.Approvals = true
		notifyReq.Notifier.Entrypoint = taskName

		call.With.SetKeyWithValue("notifier", notifyReq.Notifier.AsMap())

		// First lets notify the approvers
		_, err := t.executeNotifyTask(
			workflowTask, fmt.Sprintf("%s.notify", taskName), call)

		if err != nil {

			logrus.WithError(err).WithFields(logrus.Fields{
				"taskName": taskName,
			}).Error("Failed to notify approvers")

			return nil, err
		}

	} else {
		logrus.Infof("Resuming Thand approvals task: %s", taskName)
	}

	threshold := call.With.GetIntWithDefault("threshold", 1)

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

	approvals, ok := workflowContext["approvals"].([]any)

	if !ok {
		approvals = []any{}
	}

	var approvalData map[string]any

	if approvalEvent, ok := approval.(*cloudevents.Event); ok {

		approvalEvent.DataAs(&approvalData)

		if approved, exists := approvalData["approved"]; exists {
			approvals = append(approvals, map[string]any{
				"approved": approved,
			})
		}
	}

	workflowTask.SetContextKeyValue("approvals", approvals)

	/*
		# If anyone rejects then reject the entire request
		# otherwise if there is more than one approval then
		# authorize
		- case1:
			when: any($context.approvals[]; .approved == false)
			then: denied
		- case2:
			when: '[$context.approvals[] | select(.approved == true)] | length > 1'
			then: authorize
	*/

	// Create the switch task to handle approval or rejection
	flowDirective, err := runner.SwitchTaskHandler(
		workflowTask,
		map[string]any{
			"approvals": approvals,
		},
		fmt.Sprintf("%s.switch", taskName),
		&model.SwitchTask{
			Switch: []model.SwitchItem{
				{
					"case1": model.SwitchCase{
						When: &model.RuntimeExpression{
							Value: "any($context.approvals[]; .approved == false)",
						},
						Then: &model.FlowDirective{
							Value: "denied",
						},
					},
				},
				{
					"case2": model.SwitchCase{
						When: &model.RuntimeExpression{
							Value: fmt.Sprintf("[$context.approvals[] | select(.approved == true)] | length >= %d", threshold),
						},
						Then: &model.FlowDirective{
							Value: "authorize",
						},
					},
				},
				{
					"default": model.SwitchCase{
						// No When condition = default case (return to await more approvals)
						Then: &model.FlowDirective{
							Value: "approvals",
						},
					},
				},
			},
		})

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
