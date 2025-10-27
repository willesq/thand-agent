---
layout: default
title: Home
nav_order: 1
description: "Thand Agent - Open-source privilege access management (PAM) and just-in-time access (JIT) for cloud infrastructure and SaaS applications"
permalink: /
---

# Thand Agent Documentation
{: .fs-9 }

Open-source privilege access management (PAM) and just-in-time access (JIT) to cloud infrastructure, SaaS applications and local systems.
{: .fs-6 .fw-300 }

[Get Started Now](getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/thand-io/agent){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## What is Thand?

Thand eliminates standing access to critical infrastructure and SaaS apps. Instead of permanent admin rights, users request access when needed, for only as long as needed. The Thand server is extensible, customisable and easy to deploy as a standalone service.

### The Security Crisis

- **Static credentials get leaked**: API keys in repos, AWS keys in logs, service account keys shared in Slack
- **Over-privileged users**: 90% of permissions are unused, but remain active attack vectors
- **Automatic grants**: Users are often granted access without understanding the implications
- **Lack of visibility**: No clear audit trail of who accessed what, when, and why
- **Persistent threats**: Users with admin access can maintain access indefinitely

### The Thand Solution

- **Zero standing privileges**: No permanent admin access anywhere
- **No static credentials**: All access is temporary and tied to your identity
- **Just-in-time permissions**: Request the access you need, when you need it - and lose it once you're done
- **Complete audit trail**: Every access request and action logged for compliance

---

## Quick Start

Get up and running with Thand in just a few steps:

1. **[Install the Agent](getting-started#installation)** - Download and install the Thand agent
2. **[Configure Your Environment](setup/)** - Set up Thand for your infrastructure
3. **[Request Access](getting-started#requesting-access)** - Make your first access request

---

## Architecture Overview

The Thand architecture breaks down into three components:

```
Your Machine             Your Infrastructure             Thand Cloud (Optional)
─────────────            ───────────────────             ──────────────────────

Thand Agent  ──HTTPS──▶  Thand Server        ──HTTPS──▶  Thand Cloud
├─ CLI                   ├─ REST API                     ├─ Agent Management
├─ Sessions              ├─ Session Management           ├─ Role Management
├─ Local elevations      ├─ Workflow Worker              ├─ Workflow Management
├─ REST API              ├─ Audit Forwarder              ├─ Audit Dashboard
└─ Attestations          ├─ Basic Approvals              └─ etc
                         ├─ Event collection
```

- **Agent**: Runs on the user's local machine, provides session management and local callback endpoints
- **Server**: Forms a "login server" to allow CLIs and other clients to request and be granted elevations
- **Cloud**: Thand's proprietary cloud service that orchestrates all your servers and agents (optional)

---

## Documentation Sections

<div class="code-example" markdown="1">

### 🚀 Getting Started
Quick start guide and basic setup to get Thand running on your system.

### ⚙️ Setup Guides  
Platform-specific deployment guides for Google Cloud Platform, AWS, and local development.

### 📚 API Reference
Complete API documentation and examples for integrating with Thand.

### 🔧 Configuration
Detailed configuration options for providers, roles, and workflows.

</div>

---

## License

Thand is licensed under the BSL 1.1 license. See [LICENSE.md](https://github.com/thand-io/agent/blob/main/LICENSE.md) for more details.