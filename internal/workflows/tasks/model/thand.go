package model

import (
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/thand-io/agent/internal/models"
)

const ThandTaskName = "thand"

// ThandTask defines a custom Thand task
type ThandTask struct {
	model.TaskBase `json:",inline"`    // Inline TaskBase fields
	Thand          string              `json:"thand" validate:"required"`
	With           *models.BasicConfig `json:"with,omitempty"`
}

func (f *ThandTask) GetBase() *model.TaskBase {
	return &f.TaskBase
}

func init() {

	// and register with serverlessworkflows
	err := model.RegisterTask(ThandTaskName, model.TaskConstructor(func() model.Task {
		return &ThandTask{} // Create a new instance for each task
	}))

	if err != nil {
		panic("failed to register task type with serverlessworkflow SDK: " + err.Error())
	}
}
