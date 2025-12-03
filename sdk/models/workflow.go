package models

import internal "github.com/thand-io/agent/internal/models"

// WorkflowTask is an alias for the internal WorkflowTask type.
// It represents a task within a workflow execution context.
// See internal/models.WorkflowTask for full documentation.
type WorkflowTask = internal.WorkflowTask

// WorkflowExecutionInfo is an alias for the internal WorkflowExecutionInfo type.
// It contains metadata and status information about a workflow execution,
// including timing, approval status, and associated identities.
// See internal/models.WorkflowExecutionInfo for full documentation.
type WorkflowExecutionInfo = internal.WorkflowExecutionInfo
