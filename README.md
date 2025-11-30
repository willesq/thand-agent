<div align="center">

# Thand Agent
## Distributed Just-in-Time Access Management

**[Introduction](#introduction) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Getting Started](#getting-started) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Architecture](#thand-architecture) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Documentation](#documentation) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Status](#status) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Thand Docs](https://docs.thand.io/) &nbsp;&nbsp;**

[![Watch the Thand demo](https://github.com/thand-io/agent/blob/main/docs/assets/images/youtube_demo.png?raw=true)](https://youtu.be/WLJ1Ab0zeno)

[![Go Report Card](https://goreportcard.com/badge/github.com/thand-io/agent)](https://goreportcard.com/report/github.com/thand-io/agent)
[![Build and Release](https://github.com/thand-io/agent/actions/workflows/test-and-build.yml/badge.svg)](https://github.com/thand-io/agent/actions/workflows/test-and-build.yml)
[![Slack Community](https://img.shields.io/badge/Slack-4A154B?style=plastic&logo=slack&logoColor=white)](https://join.slack.com/t/thand-community/shared_invite/zt-3hegenhb7-w~q7JG7WYIyfefGz9NrSeA)

</div>

## Introduction

Thand is a distributed open-source agent for privileged access management (PAM) and just-in-time access (JIT) to cloud infrastructure, SaaS applications and local systems. It uses [Serverless Workflows](https://serverlessworkflow.io/) and [Temporal](https://www.temporal.io) to orchestrate and guarantee robust deterministic workflow execution and revocation, of permissions across cloud/on-prem environments and systems. It tasks “agents” to grant access where it needs to be rather than centralising permission stores. Run it locally for sudo, UAC. Or in the cloud for IAM or for individual applications. Connect to Thand Cloud for enterprise features.

**We're keen to understand different use cases in this space**. If it looks like you could make use of the Thand agent and would like some help getting it setup and configured for your environment, [let us know](https://forms.gle/dkYzfu1Nrs33HmNM9) and we'll setup some time to work with you.

## 🎯 What is Thand?

Thand eliminates standing access to critical infrastructure and SaaS apps. Instead of permanent admin rights, users request access when needed, for only as long as needed. The Thand server is extensible, customisable and easy to deploy as a standalone service.

**The Security Crisis**:

- **Static credentials get leaked**: API keys in repos, AWS keys in logs, service account keys shared in Slack
- **Over-privileged users**: 90% of permissions are unused, but remain active attack vectors. Broad workflows and roles lead to excessive permissions. Review and revocation is often a time consuming, manual process.
- **Automatic grants**: Users are often granted access without understanding the implications, leading to accidental misuse.
- **Lack of visibility**: No clear audit trail of who accessed what, when, and why.
- **Persistent threats**: Users with admin access can maintain access indefinitely, even after leaving the company.

**The Thand Solution**:

- **Zero standing privileges**: No permanent admin access anywhere
- **No static credentials**: All access is temporary and tied to your identity
- **Just-in-time permissions**: Request the access you need, when you need it - and lose it once you're done
- **Complete audit trail**: Every access request and action logged for compliance. Access is automatically reviewed during usage and revoked if the user moves off-task.

Thand is licensed under the BSL 1.1 license. See [LICENSE.md](LICENSE.md) for more details.


## Getting Started

### Option 1:  Use Thand Cloud (SaaS)

Firstly, install the Thand Agent on your local machine. You can do this via the install script:

```bash
# Install agent (https://github.com/thand-io/agent/blob/main/scripts/install.sh). Trust but verify!
curl -sSL https://get.thand.io | sh

```

Or via Homebrew on macOS / Linux:

```bash
brew tap thand-io/tap
brew install thand
```

```bash
# Connect to cloud for AI features
thand login

# Request with natural language
thand request "I need to debug customer data issue"

```

### Option 2: Self-Host Everything (Open Source)

Thand can be deployed quickly on your infrastructure. The server requires no persistent storage and can be run ephemerally. You can deploy via Docker, Kubernetes or even as an AWS Lambda function or GCP Cloud Function.

A full guide to self-hosting can be found in the [docs](https://docs.thand.io/environments/). Or you can quickly deploy Thand using the following one-click deploy options:

- [![Launch Stack](https://img.shields.io/badge/Deploy%20with-AWS%20CloudFormation-FF9900?logo=amazon-aws)](https://console.aws.amazon.com/cloudformation/home#/stacks/new?stackName=thand-agent&templateURL=https://raw.githubusercontent.com/thand-io/agent/refs/heads/main/deploy/aws/cloudformation.yaml)

To get started quickly you can run the server locally via Docker. This will start your server with the default configuration defined in the examples directory. For production usage you should provide your own configuration file. See the [docs](https://docs.thand.io/configuration/) for more details.

```bash

# Run the server locally via Docker
docker run -p 8080:8080 ghcr.io/thand-io/agent:latest server

# Or build and run locally
git clone https://github.com/thand-io/agent.git
cd agent

docker build -t thand-dev/agent:latest .
docker run -p 8080:8080 thand-dev/agent:latest server

```

You can then connect the agent to your server.

```bash
# Install agent
curl -sSL https://get.thand.io | sh

# Connect to your server
thand login --login-server http://localhost:8080

```

## Thand Architecture

The Thand architecture breaks down into three components. Both the agent and server
are contained within this repository. All access keys are stored on your infrastructure.
The Thand server can be deployed ephemerally without any persistent storage, providing a
low maintenance, high security solution. Temporal.io is used to orchestrate all workflows
and ensure just-in-time access is granted and revoked correctly and guarantees state maintenance.

- **Agent**: Runs on the user's local machine, provides session management and local callback endpoints to attest to the user's authenticity.
- **Server**: This can run anywhere you need to provide access. This forms a "login server" to allow CLIs and other clients to request and be granted elevations.
- **Cloud**: This is Thand's proprietary cloud service that orchestrates all your servers, agents and centralizes management and remote revocations. The cloud also provides additional features such as AI-driven insights and analytics. See thand.io for all the capabilities.

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
                         └─ Platform Connectors

```

## Documentation

Full documentation can be found at [docs](https://docs.thand.io).

- [Quick Start](http://docs.thand.io/getting-started.html)
- [Self-Hosting](http://docs.thand.io/environments/)
- [Roles](http://docs.thand.io/configuration/roles/)
- [Providers](http://docs.thand.io/configuration/providers/)
- [Workflows](http://docs.thand.io/configuration/workflows/)
- [CLI Reference](http://docs.thand.io/configuration/cli.html)
- [API Reference](http://docs.thand.io/api/)
- [FAQ](http://docs.thand.io/faq.html)


## Status

Thand is released & we consider it stable; we follow [semver](https://semver.org/) for releases, so major versions indicate potentially breaking changes, command line or other behaviour. We try to minimise this where possible.

We're very happy to accept pull requests, feature requests, and bugs if it's not working for you. Thand is under active development.

Please see the [contributing guide](CONTRIBUTING.md) for more details.
