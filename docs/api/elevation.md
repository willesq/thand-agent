---
layout: default
title: Elevation (Access Requests)
parent: API Reference
nav_order: 4
---

# Elevation (Access Request) Endpoints

Request elevated access to resources through approval workflows.

## Static Elevation Request

Request elevated access using predefined roles.

**GET** `/elevate`

### Query Parameters

- `role` - Role name (required)
- `provider` - Target provider (required)
- `workflow` - Workflow name (optional, uses role default if not specified)
- `reason` - Justification for access (required)
- `duration` - Access duration in ISO 8601 format (optional)
- `identities` - Comma-separated list of identities to elevate (optional)
- `session` - Encoded session token (optional)

### Response

```json
{
  "status": "pending",
  "output": {
    "workflow_id": "wf_abc123",
    "execution_id": "exec_456789"
  }
}
```

## Dynamic Elevation Request

Submit elevation request via JSON or form data.

**POST** `/elevate`

### Request Body (JSON - Static Request)

```json
{
  "role": {
    "name": "admin",
    "description": "Administrative access",
    "providers": ["aws", "gcp"],
    "workflows": ["default"],
    "enabled": true
  },
  "providers": ["aws"],
  "authenticator": "aws",
  "workflow": "default",
  "reason": "Emergency maintenance required",
  "duration": "PT2H",
  "identities": ["alice@example.com"],
  "session": {
    "version": 1,
    "expiry": "2024-01-15T12:30:00Z",
    "session": "encrypted_token"
  }
}
```

### Request Body (Form Data - Dynamic Request)

```
authenticator=aws
workflow=default
reason=Emergency maintenance
duration=PT2H
identities=alice@example.com
providers=aws,gcp
permissions=ec2:*,s3:GetObject
resources=arn:aws:ec2:*:*:instance/*
groups=admins
users=alice@example.com
```

### Response

```json
{
  "status": "pending",
  "output": {
    "workflow_id": "wf_abc123",
    "execution_id": "exec_456789"
  }
}
```

## LLM-Assisted Elevation

Request access using natural language description.

**GET** `/elevate/llm`

Returns HTML form for LLM-assisted elevation.

**POST** `/elevate/llm`

### Request Body

```json
{
  "reason": "I need to check the EC2 instances in production to investigate high CPU usage alerts"
}
```

### Response

```json
{
  "status": "pending",
  "output": {
    "suggested_role": "ec2-read-only",
    "suggested_duration": "PT1H",
    "suggested_providers": ["aws"],
    "workflow_id": "wf_abc123"
  }
}
```

## Resume Elevation Workflow

Resume a paused elevation workflow.

**GET** `/elevate/resume`

Returns workflow resumption interface.

**POST** `/elevate/resume`

### Request Body

```json
{
  "workflow_id": "wf_abc123",
  "task_token": "task_token_xyz",
  "user_input": {
    "approved": true,
    "additional_context": "Approved for emergency maintenance"
  }
}
```