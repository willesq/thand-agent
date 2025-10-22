package thand

import (
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/thand-io/agent/internal/models"
	runner "github.com/thand-io/agent/internal/workflows/runner"
	taskModel "github.com/thand-io/agent/internal/workflows/tasks/model"
)

const ThandMonitorTask = "monitor"

// ThandMonitorTask represents a custom task for Thand monitoring
func (t *thandTask) executeMonitorTask(
	workflowTask *models.WorkflowTask,
	taskName string,
	call *taskModel.ThandTask,
	input any) (any, error) {

	// Placeholder for monitoring logic
	// Implement the actual monitoring logic here

	return runner.ListenTaskHandler(workflowTask, taskName, &model.ListenTask{
		Listen: model.ListenTaskConfiguration{
			To: &model.EventConsumptionStrategy{
				Any: []*model.EventFilter{
					{
						With: &model.EventProperties{
							Type: "com.thand.alert",
						},
					},
				},
			},
		},
	}, input)
}
