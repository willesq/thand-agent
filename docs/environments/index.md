---
layout: default
title: Environments
nav_order: 5
description: "Platform-specific deployment guides for Thand Agent"
has_children: true
---

# Setup Guides
{: .no_toc }

Platform-specific deployment and configuration guides for Thand Agent.
{: .fs-6 .fw-300 }

## Overview

Thand Agent can be deployed across various platforms and environments. Choose the guide that matches your infrastructure:

### Cloud Platforms

- **[Google Cloud Platform](gcp)** - Deploy Thand on GCP with IAM integration
- **[Amazon Web Services](aws)** - Set up Thand with AWS IAM roles and policies
- **[Microsoft Azure](azure)** - Configure Thand for Azure Active Directory

### Container Platforms

- **[Kubernetes](kubernetes)** - Deploy Thand agent and server on Kubernetes
- **[Docker](docker)** - Run Thand components in Docker containers

### Local Development

- **[Local Setup](local)** - Set up Thand for local development and testing
- **[Development Environment](development)** - Configure development workflow

---

## Prerequisites

Before starting any setup guide, ensure you have:

1. **Administrative access** to your target platform
2. **Temporal server** (optional): For more advanced setups, having a temporal server can help manage workflow state and retries. For temporal its recommended to use Temporal Cloud and run Thand continually to receive temporal tasks.
