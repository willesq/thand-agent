---
layout: default
title: Temporal Configuration
parent: Configuration
nav_order: 2
description: "Instructions for setting up Temporal Search Attributes for the Agent"
---

# Temporal Configuration

The Agent is a core component of this system responsible for managing and orchestrating workflows. It relies on Temporal for workflow orchestration. To function correctly, specific **Typed Search Attributes** must be configured in your Temporal Namespace. These search attributes enable the Agent to track and manage workflow executions effectively.

Search attributes are checked at runtime. If any of the required search attributes are missing, the Agent will fail to start and display an error message indicating which attributes are missing. Ensure all required search attributes are configured before starting the Agent.

## Required Search Attributes

The following search attributes must be created in your Temporal Namespace before starting the Agent.

| Name | Type | Description |
|------|------|-------------|
| `status` | Keyword | The status of the workflow execution |
| `task` | Keyword | The current task being executed |
| `user` | Keyword | The user who initiated the workflow |
| `role` | Keyword | The role associated with the workflow |
| `workflow` | Keyword | The name of the workflow |
| `providers` | KeywordList | List of providers involved in the workflow |
| `reason` | Text | Description or reason for the workflow |
| `duration` | Int | Duration of the workflow or request |
| `identities` | KeywordList | Identities associated with the user |
| `approved` | Bool | Whether the request has been approved |


## Temporal Cloud Setup (Recommended)

For Temporal Cloud, you can manage search attributes via the UI by following these steps:

1. Navigate to your Namespace in the [Temporal Cloud UI](https://cloud.temporal.io/).
2. Scroll all the way down to the **Custom Search Attributes** section.
3. Click **Edit**.
4. Add each attribute listed above with the corresponding type.

## Local Temporal Setup

If you are running Temporal locally (e.g., using the Temporal CLI), you can create these attributes using the following commands:

```bash
temporal operator search-attribute create --namespace default --name status --type Keyword
temporal operator search-attribute create --namespace default --name task --type Keyword
temporal operator search-attribute create --namespace default --name user --type Keyword
temporal operator search-attribute create --namespace default --name role --type Keyword
temporal operator search-attribute create --namespace default --name workflow --type Keyword
temporal operator search-attribute create --namespace default --name providers --type KeywordList
temporal operator search-attribute create --namespace default --name reason --type Text
temporal operator search-attribute create --namespace default --name duration --type Int
temporal operator search-attribute create --namespace default --name identities --type KeywordList
temporal operator search-attribute create --namespace default --name approved --type Bool
```
