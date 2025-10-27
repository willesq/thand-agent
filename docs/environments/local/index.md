---
layout: default
title: Local
parent: Environments
nav_order: 4
description: "Set up Thand Server for local development and testing"
---

# Local Development Setup
{: .no_toc }

Set up Thand Agent for local development and testing.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Prerequisites

- Go 1.21 or later
- Git
- Docker (optional, for containerized setup)
- Make (optional, for build automation)

---

## Install Thand

Firstly, follow the [Installation Guide](../../getting-started/#install-via-script) to install the Thand Agent.

Once Thand has been installed you can simply run the server locally.

### Start the Server

```bash
# Terminal 1: Start server
./bin/thand server
```

The server should start without issue. However, you will need to configure; roles, providers and workflow before you can really use it.

- [Providers guide](../../configuration/providers/)
- [Workflows guide](../../configuration/workflows/) 
- [Roles guide](../../configuration/roles/)

Then review the [Getting Started](../../getting-started/) guide to authenticate and request access.

---

Once you've got the Thand server running locally, you can connect the Thand Agent to it by specifying the server URL during login:

```bash
thand login --server http://localhost:9090
```

Then proceed to login and review the roles you have access to.

```bash
thand roles
```

## Next Steps

- **[GCP Setup](gcp)** - Deploy to Google Cloud
- **[Configuration](../configuration/)** - Advanced configuration
- **[API Reference](../api)** - REST API documentation