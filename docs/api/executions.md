---
layout: default
title: Execution Management
parent: API Reference
nav_order: 8
---

# Execution Management

Manage workflow executions and their lifecycle.

## List Workflow Executions

Get all running workflow executions for the authenticated user.

**GET** `/executions`

### Response

```json
{
  "executions": [
    {
      "workflow_id": "wf_abc123",
      "run_id": "run_456789",
      "status": "running",
      "start_time": "2024-01-15T10:30:00Z",
      "execution_time": "PT5M",
      "task_queue": "thand-task-queue",
      "user": "alice@example.com",
      "workflow_type": {
        "name": "ElevateWorkflow"
      }
    }
  ]
}
```

## Create Workflow Execution

**POST** `/execution`

### Request Body

```json
{
  "workflow": "default",
  "input": {
    "role": "admin",
    "provider": "aws",
    "reason": "Emergency maintenance"
  }
}
```

## Get Execution Details

**GET** `/execution/{id}`

### Response

```json
{
  "workflow_id": "wf_abc123",
  "run_id": "run_456789",
  "status": "running",
  "start_time": "2024-01-15T10:30:00Z",
  "execution_time": "PT5M",
  "task_queue": "thand-task-queue",
  "user": "alice@example.com",
  "workflow_type": {
    "name": "ElevateWorkflow"
  },
  "history": [
    {
      "event_type": "WorkflowExecutionStarted",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ]
}
```

## Cancel Execution

**GET** `/execution/{id}/cancel`

Gracefully cancel a running workflow execution.

### Response

```json
{
  "status": "ok",
  "message": "Workflow termination signal sent"
}
```

## Terminate Execution

**GET** `/execution/{id}/terminate`

Forcefully terminate a running workflow execution.

### Response

```json
{
  "status": "ok",
  "message": "Workflow termination signal sent"
}
```