package thand

import (
	"fmt"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/thand-io/agent/internal/config"
	"github.com/thand-io/agent/internal/models"
	"github.com/thand-io/agent/internal/workflows/tasks"
	taskModel "github.com/thand-io/agent/internal/workflows/tasks/model"
)

type thandCollection struct {
	config *config.Config
	tasks.TaskCollection
}

func NewThandCollection(config *config.Config) *thandCollection {
	return &thandCollection{
		config: config,
	}
}

func (c *thandCollection) RegisterTasks(r *tasks.TaskRegistry) {

	// Register tasks
	r.RegisterTasks(
		NewThandTask(c.config),
	)

}

type thandTask struct {
	config *config.Config
}

func NewThandTask(config *config.Config) *thandTask {
	return &thandTask{
		config: config,
	}
}

func (f *thandTask) GetName() string {
	return taskModel.ThandTaskName
}

func (f *thandTask) GetDescription() string {
	return "This task handles approvals in the Thand workflow."
}

func (f *thandTask) GetVersion() string {
	return "1.0.0"
}

// Execute executes the Thand approvals task
func (t *thandTask) Execute(
	workflowTask *models.WorkflowTask,
	task *model.TaskItem,
	input any,
) (any, error) {

	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}

	taskName := task.Key
	thandTask, ok := task.Task.(*taskModel.ThandTask)

	if !ok {
		return nil, fmt.Errorf("invalid task type for ServerlessThandTask")
	}

	switch thandTask.Thand {
	case ThandApprovalsTask:
		return t.executeApprovalsTask(workflowTask, taskName, thandTask, input)
	case ThandAuthorizeTask:
		return t.executeAuthorizeTask(workflowTask, taskName, thandTask)
	case ThandValidateTask:
		return t.executeValidateTask(workflowTask, thandTask, input)
	case ThandNotifyTask:
		return t.executeNotifyTask(workflowTask, taskName, thandTask)
	case ThandRevokeTask:
		return t.executeRevokeTask(workflowTask, taskName, thandTask)
	case ThandMonitorTask:
		return t.executeMonitorTask(workflowTask, taskName, thandTask, input)
	default:
		return nil, fmt.Errorf("unknown thand task type: %s", thandTask.Thand)
	}

}
