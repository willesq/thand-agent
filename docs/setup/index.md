---
layout: default
title: Setup Guides
nav_order: 3
description: "Platform-specific setup guides for Thand Agent deployment"
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

### SaaS Integrations

- **[Salesforce](salesforce)** - Integrate with Salesforce for user management
- **[Slack](slack)** - Set up Slack-based approvals and notifications
- **[OAuth Providers](oauth)** - Configure OAuth2/OIDC authentication

---

## Prerequisites

Before starting any setup guide, ensure you have:

1. **Administrative access** to your target platform
2. **Network connectivity** between agent and server components  
3. **DNS resolution** for the Thand server endpoint
4. **TLS certificates** (recommended for production)

---

## Architecture Considerations

When planning your deployment, consider:

- **High Availability**: Deploy multiple server instances
- **Network Security**: Use VPNs or private networks when possible
- **Audit Logging**: Configure centralized log collection
- **Backup Strategy**: Regular backup of configuration and audit data