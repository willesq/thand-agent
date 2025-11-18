# Thand Provider

The Thand provider enables authentication through Thand's federated OIDC service.

## Configuration

Add the following to your `config.yaml`:

```yaml
providers:
  thand:
    name: Thand Dev
    provider: thand
    description: Thand Development Provider
    enabled: true
    config:
      endpoint: "https://auth.thand.io"  # Optional, defaults to https://auth.thand.io
```

### Configuration Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `endpoint` | No | `https://auth.thand.io` | The Thand authentication endpoint URL |

### Examples

**Production:**
```yaml
providers:
  thand:
    name: Thand Production
    provider: thand
    enabled: true
    config:
      endpoint: "https://auth.thand.io"
```

**Local Development:**
```yaml
providers:
  thand:
    name: Thand Local
    provider: thand
    enabled: true
    config:
      endpoint: "http://localhost:3000"
```

## How it Works

The Thand provider:
1. Redirects users to the Thand authentication endpoint
2. Receives an authorization code after successful authentication
3. Exchanges the code for user information including email, name, and groups
4. Creates a session with 1-hour expiry

User information retrieved includes:
- User ID (sub)
- Email address
- Username
- Full name
- Groups/roles (if available)
