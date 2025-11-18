# Thand Provider

The Thand provider implements a federated OpenID Connect (OIDC) authentication provider for universal federation through `auth.thand.io`. This provider enables secure authentication via a centralized thand identity service that can federate with multiple identity providers.

## Overview

The Thand provider follows the OAuth2/OIDC authorization code flow:

1. **Authorization Request** - Redirects users to `auth.thand.io` with a `return_to` callback URL
2. **Token Exchange** - Securely exchanges the authorization code for access/refresh tokens
3. **User Information** - Retrieves authenticated user information from the userinfo endpoint
4. **Session Management** - Validates and renews sessions using refresh tokens

## Security Features

### Token Security
- **TLS 1.2+** - All communication uses secure HTTPS with TLS 1.2 minimum
- **Certificate Validation** - Certificate verification is always enabled (never skips)
- **Secure Token Storage** - Tokens are encrypted using the configured encryption service (AWS KMS, GCP KMS, Azure Key Vault, or local encryption)
- **Token Validation** - Access tokens are validated by calling the userinfo endpoint
- **Refresh Tokens** - Long-lived sessions using secure refresh token rotation

### Session Security
The agent's encryption service automatically handles:
- **Encryption at Rest** - Sessions are encrypted before storage using AES-256-GCM (local) or cloud KMS
- **Secure Transmission** - State parameters are encrypted/signed during OAuth flow
- **Token Expiry** - Automatic session expiration tracking
- **Session Validation** - Regular validation of active sessions

### Data Protection
- **User Information** - Returned from thand.dev and encrypted in local session storage
- **State Parameter** - Encrypted using the configured encryption service to prevent CSRF attacks
- **No Plain Text Storage** - Tokens and sensitive data are never stored in plain text

## Configuration

### Basic Configuration

```yaml
version: "1.0"
providers:
  thand:
    name: Thand Federation
    description: Federated OIDC authentication via Thand
    provider: thand
    enabled: true
    config:
      # Optional: Override default endpoints
      auth_endpoint: "https://auth.thand.io"
      token_endpoint: "https://auth.thand.io/token"
      userinfo_endpoint: "https://auth.thand.io/userinfo"
      
      # Optional: Client credentials (if required by thand.dev)
      client_id: "your-client-id"
      client_secret: "${THAND_CLIENT_SECRET}"
```

### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `auth_endpoint` | string | No | `https://auth.thand.io` | OAuth2 authorization endpoint |
| `token_endpoint` | string | No | `https://auth.thand.io/token` | Token exchange endpoint |
| `userinfo_endpoint` | string | No | `https://auth.thand.io/userinfo` | User information endpoint |
| `client_id` | string | No | - | OAuth2 client ID (if required) |
| `client_secret` | string | No | - | OAuth2 client secret (if required) |

### Encryption Configuration

The thand provider relies on the agent's encryption service for securing tokens. Configure encryption in your main config:

```yaml
services:
  encryption:
    provider: "local"  # or aws, gcp, azure
    config:
      key: "${ENCRYPTION_KEY}"
      salt: "${ENCRYPTION_SALT}"
```

For production, use cloud KMS:

```yaml
services:
  encryption:
    provider: "aws"
    config:
      kms_arn: "arn:aws:kms:us-east-1:123456789:key/abc-123"
      region: "us-east-1"
```

## How It Works

### 1. Authorization Flow

When a user initiates authentication:

```
User → Thand Agent → auth.thand.io (with return_to callback)
```

The agent constructs an authorization URL:
```
https://auth.thand.io?return_to=https://agent.local/api/v1/auth/callback/thand&state=<encrypted-state>
```

The `state` parameter is encrypted using the encryption service and contains:
- Callback URL
- Client identifier
- Provider name
- Anti-CSRF token

### 2. Token Exchange

After user authenticates at auth.thand.io, they're redirected back with an authorization code:

```
https://agent.local/api/v1/auth/callback/thand?code=<auth-code>&state=<encrypted-state>
```

The agent then:
1. Validates the encrypted state parameter
2. Exchanges the code for tokens via POST to token endpoint
3. Validates the token response
4. Fetches user information from userinfo endpoint

### 3. Session Creation

Once authenticated:
1. User information is retrieved from the userinfo endpoint
2. A session object is created with user details and tokens
3. The session is encrypted using the encryption service
4. Encrypted session is stored locally and in cookies

### 4. Session Validation

For each authenticated request:
1. Session is decrypted from storage
2. Token expiry is checked
3. Optionally, access token is validated against userinfo endpoint
4. If expired, automatic renewal using refresh token

## User Information

The provider retrieves and validates the following user information:

```go
type UserInfoResponse struct {
    Sub               string   // Unique user ID (required)
    Email             string   // User's email address
    EmailVerified     bool     // Whether email is verified
    Name              string   // Full name
    PreferredUsername string   // Username
    Groups            []string // User groups/roles
}
```

## API Methods

### AuthorizeSession

Initiates the OAuth2 authorization flow by redirecting to auth.thand.io.

```go
resp, err := provider.AuthorizeSession(ctx, &models.AuthorizeUser{
    RedirectUri: "https://agent.local/api/v1/auth/callback/thand",
    State:       encryptedState,
    Scopes:      []string{"openid", "profile", "email"},
})
```

### CreateSession

Exchanges authorization code for tokens and creates an authenticated session.

```go
session, err := provider.CreateSession(ctx, &models.AuthorizeUser{
    Code:        authCode,
    RedirectUri: callbackUrl,
})
```

### ValidateSession

Validates an existing session and checks token validity.

```go
err := provider.ValidateSession(ctx, session)
```

### RenewSession

Renews an expired session using the refresh token.

```go
newSession, err := provider.RenewSession(ctx, session)
```

## Security Best Practices

### 1. Use HTTPS Only
Always use HTTPS for all endpoints. The provider enforces TLS 1.2+ and certificate validation.

### 2. Secure Client Credentials
Store client secrets in environment variables or secure vaults:
```yaml
config:
  client_secret: "${THAND_CLIENT_SECRET}"
```

### 3. Enable Encryption Service
Configure a robust encryption service for production:
- **AWS**: Use AWS KMS with proper IAM policies
- **GCP**: Use Cloud KMS with service account permissions
- **Azure**: Use Azure Key Vault with managed identities
- **Local**: Use strong encryption keys (not recommended for production)

### 4. Token Rotation
The provider automatically handles token refresh. Ensure refresh tokens are stored securely.

### 5. Session Expiry
Configure appropriate session timeouts:
```yaml
server:
  session:
    max_age: "24h"
    cookie_secure: true
    cookie_httponly: true
```

## Troubleshooting

### Token Exchange Failed

**Error**: `failed to exchange code for token`

**Solutions**:
- Verify the token endpoint URL is correct
- Check client_id and client_secret if required
- Ensure the authorization code hasn't expired
- Verify network connectivity to auth.thand.io

### Invalid User Info

**Error**: `invalid userinfo: missing subject (sub)`

**Solutions**:
- Verify the userinfo endpoint URL
- Check that the access token is valid
- Ensure proper scopes were requested (openid, profile, email)

### Session Validation Failed

**Error**: `token validation failed`

**Solutions**:
- Check if the access token has expired
- Try renewing the session using refresh token
- Re-authenticate if refresh token is also expired

### Certificate Verification Failed

**Error**: `x509: certificate verification failed`

**Solutions**:
- Ensure system CA certificates are up to date
- Verify auth.thand.io has a valid SSL certificate
- Check for network proxies interfering with TLS

## Example Workflow Configuration

```yaml
version: "1.0"
workflows:
  secure_access:
    description: Secure resource access with Thand authentication
    authentication: thand  # Use thand provider for auth
    enabled: true
    workflow:
      document:
        dsl: "1.0.0-alpha5"
        namespace: "thand"
        name: "secure-access"
        version: "1.0.0"
      do:
        - validate:
            thand: validate
            with:
              validator: static
            then: authorize
        - authorize:
            thand: authorize
            with:
              provider: aws
              role: admin
```

## Architecture

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│             │         │              │         │             │
│    User     ├────────►│ Thand Agent  ├────────►│ auth.thand  │
│             │         │              │         │    .io      │
└─────────────┘         └──────┬───────┘         └─────────────┘
                               │
                               │ Encrypted
                               │ Session
                               ▼
                        ┌──────────────┐
                        │  Encryption  │
                        │   Service    │
                        │ (KMS/Local)  │
                        └──────────────┘
```

## Compliance

The Thand provider is designed to support compliance with:
- **OAuth 2.0** (RFC 6749)
- **OpenID Connect Core 1.0**
- **GDPR** - Minimal data retention, encrypted storage
- **SOC 2** - Audit logging, secure credential handling
- **ISO 27001** - Information security controls

## Related Documentation

- [OAuth2 Provider](../oauth2/README.md)
- [Google OAuth2 Provider](../oauth2.google/README.md)
- [Encryption Service](../../config/services/encrypt/README.md)
- [Session Management](../../sessions/README.md)
