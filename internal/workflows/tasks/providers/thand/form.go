package thand

import (
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	taskModel "github.com/thand-io/agent/internal/workflows/tasks/model"
)

const ThandFormTask = "form"

type FormTask struct {
}

// executeFormTask handles the "form" Thand task
func (t *thandTask) executeFormTask(
	workflowTask *models.WorkflowTask,
	taskName string,
	call *taskModel.ThandTask,
) (any, error) {

	log := workflowTask.GetLogger()

	_, err := workflowTask.GetContextAsElevationRequest()

	if err != nil {
		return nil, err
	}

	// Parse the form task configuration
	var revokeCallTask RevokeTask
	err = common.ConvertInterfaceToInterface(call.With, &revokeCallTask)
	if err != nil {
		log.WithError(err).Error("Failed to parse form task configuration")
		// Continue without notifiers if parsing fails
	}

	return nil, nil
}
