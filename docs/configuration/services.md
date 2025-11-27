---
layout: default
title: Services Configuration
parent: Configuration
nav_order: 2
description: "Detailed configuration reference for Thand Agent services"
---

# Services Configuration
{: .no_toc }

Complete reference for configuring Thand Agent's backend services including encryption, vault, scheduler, LLM, and Temporal.
{: .fs-6 .fw-300 }

## Table of Contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

Thand Agent uses pluggable services for core functionality like encryption, secrets management, scheduling, and workflow orchestration. Each service supports multiple providers (AWS, GCP, Azure, Local) allowing you to choose the best fit for your infrastructure.

Services are configured under the `services` key in your configuration file:

```yaml
services:
  encryption:
    provider: local
    password: "your-secure-password"
    salt: "your-unique-salt"
  vault:
    provider: aws
  scheduler:
    provider: local
  temporal:
    host: localhost
    port: 7233
  llm:
    provider: gemini
    api_key: "your-api-key"
```

---

## Configuration Inheritance

All service configuration options can be defined in `environment.config`, and each service will automatically inherit these values as the base configuration. You can then override specific options for individual services as needed.

This is particularly useful when:
- Running on a cloud platform where most services should use the same credentials
- You want to mix providers (e.g., use GCP Secret Manager but local encryption)
- You need different configuration for specific services

### How It Works

1. **Base Configuration**: Values defined in `environment.config` are available to all services
2. **Service Override**: Values defined in `services.<service_name>` override the base configuration
3. **Provider Selection**: Each service independently selects its provider

### Example: Mixed Provider Configuration

In this example, the agent runs on GCP and uses GCP Secret Manager for the vault, but uses local encryption for performance:

```yaml
environment:
  platform: gcp
  config:
    # These values are inherited by ALL services as defaults
    project_id: my-gcp-project
    location: us-central1
    key_ring: thand-keyring
    key_name: thand-key

services:
  # Vault uses GCP - inherits project_id, location from environment.config
  vault:
    provider: gcp
    # No need to specify project_id - inherited from environment.config
  
  # Encryption uses local provider - ignores GCP config, uses its own
  encryption:
    provider: local
    password: "my-secure-password"
    salt: "my-unique-salt"
  
  # If you needed GCP encryption, it would inherit the key_ring and key_name
  # encryption:
  #   provider: gcp
  #   # project_id, location, key_ring, key_name all inherited
```

### Example: AWS with Service-Specific Overrides

```yaml
environment:
  platform: aws
  config:
    region: us-east-1
    profile: production
    kms_arn: "arn:aws:kms:us-east-1:123456789012:key/default-key"

services:
  # Uses all inherited values from environment.config
  encryption:
    provider: aws
  
  # Overrides region for vault (maybe secrets are in a different region)
  vault:
    provider: aws
    region: us-west-2
  
  # Uses local scheduler (doesn't need AWS config)
  scheduler:
    provider: local
```

### Configuration Priority

Configuration values are resolved in this order (later values override earlier ones):

1. Default values (hardcoded in the application)
2. `environment.config.*` values
3. `services.<service_name>.*` values

{: .note }
> The `provider` option must always be specified at the service level. It is not inherited from `environment.platform`.

---

## Encryption Service

The encryption service handles encrypting and decrypting sensitive data within the agent. This is used for session data, credentials, and other sensitive information.

### Provider Selection

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `services.encryption.provider` | string | `local` | Encryption provider: `aws`, `gcp`, `azure`, `local` |

### Local Provider

The local provider uses AES-256-GCM encryption with PBKDF2 key derivation. This is suitable for development and single-instance deployments.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `services.encryption.password` | string | `changeme` | Master password for key derivation |
| `services.encryption.salt` | string | `changeme` | Salt for key derivation |

{: .warning }
> **Security Warning**: Always change the default `password` and `salt` values in production. Using default values will trigger a warning in the logs.

**Example Configuration:**

```yaml
services:
  encryption:
    provider: local
    password: "my-secure-master-password"
    salt: "unique-environment-identifier"
```

**Environment Variables:**

```bash
THAND_SERVICES_ENCRYPTION_PROVIDER=local
THAND_SERVICES_ENCRYPTION_PASSWORD=my-secure-master-password
THAND_SERVICES_ENCRYPTION_SALT=unique-environment-identifier
```

### AWS Provider (KMS)

Uses AWS Key Management Service (KMS) for encryption operations.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `services.encryption.kms_arn` | string | **Yes** | ARN of the KMS key to use |
| `services.encryption.region` | string | No | AWS region (defaults to environment config) |
| `services.encryption.profile` | string | No | AWS profile name |
| `services.encryption.access_key_id` | string | No | AWS access key ID |
| `services.encryption.secret_access_key` | string | No | AWS secret access key |

**Example Configuration:**

```yaml
services:
  encryption:
    provider: aws
    kms_arn: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
    region: us-east-1
```

**Environment Variables:**

```bash
THAND_SERVICES_ENCRYPTION_PROVIDER=aws
THAND_SERVICES_ENCRYPTION_KMS_ARN=arn:aws:kms:us-east-1:123456789012:key/...
AWS_REGION=us-east-1
AWS_PROFILE=my-profile
```

### GCP Provider (Cloud KMS)

Uses Google Cloud Key Management Service for encryption operations.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `services.encryption.project_id` | string | **Yes** | GCP project ID |
| `services.encryption.location` | string | No | KMS location (default: `global`) |
| `services.encryption.key_ring` | string | **Yes** | Cloud KMS key ring name |
| `services.encryption.key_name` | string | **Yes** | Cloud KMS key name |

**Example Configuration:**

```yaml
services:
  encryption:
    provider: gcp
    project_id: my-gcp-project # This will be auto-detected on GCE
    location: us-central1
    key_ring: thand-keyring
    key_name: thand-encryption-key
```

**Environment Variables:**

```bash
THAND_SERVICES_ENCRYPTION_PROVIDER=gcp
THAND_SERVICES_ENCRYPTION_PROJECT_ID=my-gcp-project
THAND_SERVICES_ENCRYPTION_LOCATION=us-central1
THAND_SERVICES_ENCRYPTION_KEY_RING=thand-keyring
THAND_SERVICES_ENCRYPTION_KEY_NAME=thand-encryption-key
```

### Azure Provider (Key Vault)

Uses Azure Key Vault for encryption operations with RSA-OAEP algorithm.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `services.encryption.vault_url` | string | **Yes** | Azure Key Vault URL |
| `services.encryption.key_name` | string | **Yes** | Key name in the vault |

**Example Configuration:**

```yaml
services:
  encryption:
    provider: azure
    vault_url: "https://my-keyvault.vault.azure.net"
    key_name: thand-encryption-key
```

**Environment Variables:**

```bash
THAND_SERVICES_ENCRYPTION_PROVIDER=azure
THAND_SERVICES_ENCRYPTION_VAULT_URL=https://my-keyvault.vault.azure.net
THAND_SERVICES_ENCRYPTION_KEY_NAME=thand-encryption-key
```

---

## Vault Service

The vault service handles secure storage and retrieval of secrets like API keys, credentials, and tokens.

### Provider Selection

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `services.vault.provider` | string | `local` | Vault provider: `aws`, `gcp`, `azure`, `hashicorp`, `local` |

### Local Provider

The local vault provider is a placeholder for development. It does not actually store secrets persistently.

{: .note }
> The local vault provider is not recommended for production use. Use a cloud provider or HashiCorp Vault for secure secret storage.

### AWS Provider (Secrets Manager)

Uses AWS Secrets Manager for secure secret storage.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `services.vault.region` | string | No | AWS region (defaults to environment config) |
| `services.vault.profile` | string | No | AWS profile name |
| `services.vault.access_key_id` | string | No | AWS access key ID |
| `services.vault.secret_access_key` | string | No | AWS secret access key |

**Example Configuration:**

```yaml
services:
  vault:
    provider: aws
    region: us-east-1
```

**Environment Variables:**

```bash
THAND_SERVICES_VAULT_PROVIDER=aws
AWS_REGION=us-east-1
```

### GCP Provider (Secret Manager)

Uses Google Cloud Secret Manager for secure secret storage.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `services.vault.project_id` | string | No | GCP project ID (auto-detected on GCE) |

**Example Configuration:**

```yaml
services:
  vault:
    provider: gcp
    project_id: my-gcp-project
```

**Environment Variables:**

```bash
THAND_SERVICES_VAULT_PROVIDER=gcp
THAND_SERVICES_VAULT_PROJECT_ID=my-gcp-project
```

### Azure Provider (Key Vault Secrets)

Uses Azure Key Vault for secure secret storage.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `services.vault.vault_url` | string | **Yes** | Azure Key Vault URL |

**Example Configuration:**

```yaml
services:
  vault:
    provider: azure
    vault_url: "https://my-keyvault.vault.azure.net"
```

**Environment Variables:**

```bash
THAND_SERVICES_VAULT_PROVIDER=azure
THAND_SERVICES_VAULT_VAULT_URL=https://my-keyvault.vault.azure.net
```

### HashiCorp Vault Provider

Uses HashiCorp Vault for enterprise-grade secret management.

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `services.vault.vault_url` | string | **Yes** | - | HashiCorp Vault server URL |
| `services.vault.token` | string | No | - | Vault authentication token |
| `services.vault.mount_path` | string | No | `secret` | KV secrets engine mount path |
| `services.vault.secret_path` | string | No | `data` | Path within the mount (use `data` for KV v2) |
| `services.vault.timeout` | duration | No | - | Request timeout |

**Example Configuration:**

```yaml
services:
  vault:
    provider: hashicorp
    vault_url: "https://vault.example.com:8200"
    token: "hvs.your-vault-token"
    mount_path: secret
    secret_path: data
```

**Environment Variables:**

```bash
THAND_SERVICES_VAULT_PROVIDER=hashicorp
THAND_SERVICES_VAULT_VAULT_URL=https://vault.example.com:8200
VAULT_TOKEN=hvs.your-vault-token
```

{: .note }
> The HashiCorp Vault provider will also check the `VAULT_TOKEN` environment variable and `~/.vault-token` file for authentication.

---

## Scheduler Service

The scheduler service handles job scheduling for time-based operations like session expiration and cleanup tasks.

### Provider Selection

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `services.scheduler.provider` | string | `local` | Scheduler provider: `aws`, `gcp`, `azure`, `local` |

### Local Provider

Uses an in-memory scheduler based on gocron. Suitable for single-instance deployments.

**Example Configuration:**

```yaml
services:
  scheduler:
    provider: local
```

{: .note }
> The local scheduler runs in-process and does not persist scheduled jobs across restarts. For production deployments requiring persistence, use Temporal workflows instead.

### Cloud Providers (AWS, GCP, Azure)

{: .warning }
> Cloud scheduler providers (AWS EventBridge Scheduler, GCP Cloud Scheduler, Azure Logic Apps) are not yet implemented. Use the local scheduler or Temporal for production scheduling.

---

## Temporal Service

Temporal provides durable workflow orchestration for access request workflows, approvals, and complex multi-step operations.

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `services.temporal.host` | string | `localhost` | Temporal server hostname |
| `services.temporal.port` | integer | `7233` | Temporal server port |
| `services.temporal.namespace` | string | `default` | Temporal namespace |
| `services.temporal.api_key` | string | - | API key for Temporal Cloud |
| `services.temporal.mtls_cert` | string | - | mTLS certificate (PEM content) |
| `services.temporal.mtls_cert_path` | string | - | Path to mTLS certificate file |

### Local Development

For local development, you can run Temporal using Docker:

```bash
docker run -d --name temporal \
  -p 7233:7233 \
  temporalio/auto-setup:latest
```

**Configuration:**

```yaml
services:
  temporal:
    host: localhost
    port: 7233
    namespace: default
```

### Temporal Cloud

For production deployments, use Temporal Cloud with API key authentication:

```yaml
services:
  temporal:
    host: my-namespace.tmprl.cloud
    port: 7233
    namespace: my-namespace.my-account
    api_key: "your-temporal-cloud-api-key"
```

**Environment Variables:**

```bash
THAND_SERVICES_TEMPORAL_HOST=my-namespace.tmprl.cloud
THAND_SERVICES_TEMPORAL_PORT=7233
THAND_SERVICES_TEMPORAL_NAMESPACE=my-namespace.my-account
THAND_SERVICES_TEMPORAL_API_KEY=your-temporal-cloud-api-key
```

### Self-Hosted with mTLS

For self-hosted Temporal with mTLS authentication:

```yaml
services:
  temporal:
    host: temporal.internal.example.com
    port: 7233
    namespace: production
    mtls_cert_path: /etc/thand/temporal-cert.pem
```

---

## LLM Service (Large Language Model)

The LLM service provides AI capabilities for intelligent access request processing and natural language understanding.

### Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `services.llm.provider` | string | **Yes** | LLM provider: `gemini`, `openai`, `anthropic` |
| `services.llm.api_key` | string | **Yes** | API key for the LLM provider |
| `services.llm.model` | string | No | Model name (provider-specific default) |
| `services.llm.base_url` | string | No | Custom API base URL |

### Google Gemini

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `services.llm.model` | string | `gemini-2.5-flash` | Gemini model to use |

**Example Configuration:**

```yaml
services:
  llm:
    provider: gemini
    api_key: "your-google-api-key"
    model: gemini-2.5-flash
```

**Environment Variables:**

```bash
THAND_SERVICES_LLM_PROVIDER=gemini
THAND_SERVICES_LLM_API_KEY=your-google-api-key
THAND_SERVICES_LLM_MODEL=gemini-2.5-flash
```

### OpenAI

{: .note }
> OpenAI provider integration is planned but not yet implemented.

**Example Configuration:**

```yaml
services:
  llm:
    provider: openai
    api_key: "sk-your-openai-api-key"
    model: gpt-4
    base_url: https://api.openai.com/v1
```

### Anthropic

{: .note }
> Anthropic Claude provider integration is planned but not yet implemented.

**Example Configuration:**

```yaml
services:
  llm:
    provider: anthropic
    api_key: "your-anthropic-api-key"
    model: claude-3-opus-20240229
```

---

## Complete Example Configuration

Here's a complete example showing all services configured:

### Local Development

```yaml
# config.yaml - Local Development
services:
  encryption:
    provider: local
    password: "dev-password-change-in-prod"
    salt: "dev-salt-change-in-prod"
  
  vault:
    provider: local
  
  scheduler:
    provider: local
  
  temporal:
    host: localhost
    port: 7233
    namespace: default
  
  llm:
    provider: gemini
    api_key: "your-gemini-api-key"
```

### AWS Production

```yaml
# config.yaml - AWS Production
environment:
  platform: aws
  config:
    region: us-east-1
    profile: production

services:
  encryption:
    provider: aws
    kms_arn: "arn:aws:kms:us-east-1:123456789012:key/..."
  
  vault:
    provider: aws
  
  scheduler:
    provider: local
  
  temporal:
    host: production.tmprl.cloud
    port: 7233
    namespace: production.my-org
    api_key: "${TEMPORAL_API_KEY}"
  
  llm:
    provider: gemini
    api_key: "${GEMINI_API_KEY}"
```

### GCP Production

```yaml
# config.yaml - GCP Production
environment:
  platform: gcp
  config:
    project_id: my-gcp-project
    location: us-central1

services:
  encryption:
    provider: gcp
    key_ring: thand-keyring
    key_name: thand-key
  
  vault:
    provider: gcp
  
  scheduler:
    provider: local
  
  temporal:
    host: production.tmprl.cloud
    port: 7233
    namespace: production.my-org
    api_key: "${TEMPORAL_API_KEY}"
  
  llm:
    provider: gemini
    api_key: "${GEMINI_API_KEY}"
```

### Azure Production

```yaml
# config.yaml - Azure Production
environment:
  platform: azure
  config:
    vault_url: "https://my-keyvault.vault.azure.net"

services:
  encryption:
    provider: azure
    vault_url: "https://my-keyvault.vault.azure.net"
    key_name: thand-encryption-key
  
  vault:
    provider: azure
    vault_url: "https://my-keyvault.vault.azure.net"
  
  scheduler:
    provider: local
  
  temporal:
    host: production.tmprl.cloud
    port: 7233
    namespace: production.my-org
    api_key: "${TEMPORAL_API_KEY}"
  
  llm:
    provider: gemini
    api_key: "${GEMINI_API_KEY}"
```

---

## Troubleshooting

### Encryption Service

**Problem:** Warning about default secrets  
**Solution:** Set custom `password` and `salt` values in your configuration.

```yaml
services:
  encryption:
    password: "your-secure-password"
    salt: "your-unique-salt"
```

**Problem:** AWS KMS encryption failing  
**Solution:** Verify the `kms_arn` is correct and your IAM role has `kms:Encrypt` and `kms:Decrypt` permissions.

### Vault Service

**Problem:** HashiCorp Vault connection failing  
**Solution:** Verify the `vault_url` is accessible and check your token authentication.

```bash
# Test vault connection
curl -H "X-Vault-Token: $VAULT_TOKEN" https://vault.example.com:8200/v1/sys/health
```

### Temporal Service

**Problem:** Cannot connect to Temporal  
**Solution:** Verify the host, port, and namespace are correct. For Temporal Cloud, ensure your API key is valid.

```bash
# Test Temporal connection
temporal operator namespace describe --address your-host:7233 --namespace your-namespace
```
