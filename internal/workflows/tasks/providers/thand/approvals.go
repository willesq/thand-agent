package thand

import (
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/sirupsen/logrus"
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

	// First lets notify the approvers
	_, err := t.executeNotifyTask(workflowTask, taskName, call)

	if err != nil {

		logrus.WithError(err).WithFields(logrus.Fields{
			"taskName": taskName,
		}).Error("Failed to notify approvers")

		return nil, err
	}

	logrus.Infof("Executing Thand monitor task: %s", taskName)

	approval, err := runner.ListenTaskHandler(workflowTask, taskName, &model.ListenTask{
		Listen: model.ListenTaskConfiguration{
			To: &model.EventConsumptionStrategy{
				Any: []*model.EventFilter{
					{
						With: &model.EventProperties{
							Type: "com.thand.approval",
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
	flowDirective, err := runner.SwitchTaskHandler(workflowTask, approval, taskName, &model.SwitchTask{
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
						Value: "[$context.approvals[] | select(.approved == true)] | length > 1",
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

	return flowDirective, nil
}
