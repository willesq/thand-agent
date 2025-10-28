---
layout: default
title: Workflow Management
parent: Agent
grand_parent: API Reference
nav_order: 7
---

# Workflow Management

Manage approval workflows and their execution.

## List Workflows

Get all available workflows.

**GET** `/workflows`

### Response

```json
{
  "version": "1.0",
  "workflows": {
    "default": {
      "name": "default",
      "description": "Standard approval workflow",
      "enabled": true
    },
    "emergency": {
      "name": "emergency",
      "description": "Emergency bypass workflow",
      "enabled": true
    }
  }
}
```

## Get Workflow Details

**GET** `/workflow/{name}`

### Response

```json
{
  "name": "default",
  "description": "Standard approval workflow",
  "enabled": true,
  "document": {
    "dsl": "1.0.0",
    "namespace": "thand",
    "name": "default",
    "version": "1.0.0",
    "do": [
      {
        "request": {
          "call": "http",
          "with": {
            "method": "post",
            "uri": "https://api.example.com/approve"
          }
        }
      }
    ]
  }
}
```