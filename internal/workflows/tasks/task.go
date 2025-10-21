package tasks

import (
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/thand-io/agent/internal/models"
)

type TaskCollection interface {
	RegisterTasks(registry *TaskRegistry)
}

// Function defines the interface that all Thand Functions must implement
type Task interface {
	// GetName returns the unique name/identifier for this Task
	GetName() string

	// GetDescription returns a human-readable description of what this Task does
	GetDescription() string

	// GetVersion returns the version of this Task implementation
	GetVersion() string

	// Execute performs the main Function logic
	// Security: All inputs should be pre-validated by ValidateRequest
	Execute(
		workflowTask *models.WorkflowTask,
		task *model.TaskItem,
		input any,
	) (any, error)
}
