---
layout: default
title: Email SMTP
description: SMTP email provider configuration
parent: Providers
grand_parent: Configuration
---

# Email SMTP Provider

Standard SMTP integration for email delivery through any SMTP-compatible email server.

## Overview

The SMTP email provider (`email.smtp`) enables email notifications through traditional SMTP servers. This is the default platform for the email provider and works with any SMTP-compatible service including Gmail, Outlook, SendGrid, and self-hosted mail servers.

## Capabilities

- **Universal Compatibility**: Works with any SMTP server
- **TLS/SSL Support**: Secure email transmission
- **Authentication**: Username/password authentication
- **HTML & Plain Text**: Support for both email formats
- **Multiple Recipients**: Send to multiple addresses simultaneously

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `platform` | string | No | Set to `smtp` (default if omitted) |
| `host` | string | Yes | SMTP server hostname |
| `port` | number | Yes | SMTP server port (typically 587 for TLS, 465 for SSL) |
| `user` | string | Yes | SMTP authentication username |
| `pass` | string | Yes | SMTP authentication password |
| `from` | string | Yes | Default sender email address |
| `tls_skip_verify` | boolean | No | Skip TLS certificate verification (default: false) |

## Example Configuration

### Basic SMTP Configuration

```yaml
version: "1.0"
providers:
  email:
    name: Email Notifications
    description: SMTP email provider
    provider: email  # can use 'email' with platform: smtp, or use 'email.smtp' directly
    enabled: true
    config:
      platform: smtp
      host: smtp.example.com
      port: 587
      user: <your-smtp-username>
      pass: <your-smtp-password>
      from: noreply@example.com
```

**Alternative** (using direct provider name, no platform needed):

```yaml
version: "1.0"
providers:
  email:
    name: Email Notifications
    description: SMTP email provider
    provider: email.smtp  # directly specify email.smtp
    enabled: true
    config:
      host: smtp.example.com
      port: 587
      user: <your-smtp-username>
      pass: <your-smtp-password>
      from: noreply@example.com
```

### Gmail Configuration

```yaml
version: "1.0"
providers:
  email-gmail:
    name: Gmail SMTP
    description: Gmail SMTP email provider
    provider: email
    enabled: true
    config:
      platform: smtp
      host: smtp.gmail.com
      port: 587
      user: <your-gmail-address>
      pass: <your-app-password>
      from: <your-gmail-address>
```

### Outlook/Office 365 Configuration

```yaml
version: "1.0"
providers:
  email-outlook:
    name: Outlook SMTP
    description: Outlook SMTP email provider
    provider: email
    enabled: true
    config:
      platform: smtp
      host: smtp-mail.outlook.com
      port: 587
      user: <your-outlook-email>
      pass: <your-password>
      from: <your-outlook-email>
```

### SendGrid Configuration

```yaml
version: "1.0"
providers:
  email-sendgrid:
    name: SendGrid SMTP
    description: SendGrid SMTP email provider
    provider: email
    enabled: true
    config:
      platform: smtp
      host: smtp.sendgrid.net
      port: 587
      user: apikey
      pass: <your-sendgrid-api-key>
      from: noreply@example.com
```

### Self-Hosted SMTP with TLS Skip

```yaml
version: "1.0"
providers:
  email-internal:
    name: Internal SMTP
    description: Internal SMTP server
    provider: email
    enabled: true
    config:
      platform: smtp
      host: mail.internal.company.com
      port: 587
      user: <smtp-user>
      pass: <smtp-password>
      from: agent@company.com
      tls_skip_verify: true
```

## Common SMTP Server Settings

| Provider | Host | Port | Notes |
|----------|------|------|-------|
| Gmail | smtp.gmail.com | 587 | Requires App Password if 2FA enabled |
| Outlook/Office 365 | smtp-mail.outlook.com | 587 | Standard authentication |
| Yahoo | smtp.mail.yahoo.com | 587 | Requires App Password |
| SendGrid | smtp.sendgrid.net | 587 | Username is always "apikey" |
| Mailgun | smtp.mailgun.org | 587 | Use domain-specific credentials |
| AWS SES (SMTP) | email-smtp.us-east-1.amazonaws.com | 587 | Use SMTP credentials from SES |

## Usage in Workflows

```yaml
workflows:
  - name: access-approval
    steps:
      - task: notify
        type: email
        config:
          provider: email
          subject: "Access Request Approval Required"
          to:
            - approver@example.com
          body:
            text: "A new access request requires your approval."
            html: "<h2>Access Request</h2><p>A new access request requires your approval.</p>"
```

## Best Practices

### Security

1. **Use App Passwords**: For Gmail and other services with 2FA, generate app-specific passwords
2. **Secure Credentials**: Store SMTP credentials in environment variables or secrets management
3. **TLS Verification**: Keep `tls_skip_verify: false` unless working with trusted internal servers
4. **Sender Authentication**: Ensure sender address is authorized on the SMTP server

### Performance

1. **Connection Pooling**: The provider reuses connections for multiple emails
2. **Rate Limits**: Be aware of provider-specific rate limits (e.g., Gmail: 500/day for free accounts)
3. **Timeout Handling**: SMTP operations have built-in timeout handling

### Reliability

1. **Delivery Confirmation**: Monitor SMTP response codes
2. **Retry Logic**: Implement retry logic for transient failures
3. **Fallback Providers**: Configure multiple email providers for redundancy

## Troubleshooting

### Authentication Failures

**Problem**: `535 Authentication failed` or similar errors

**Solutions**:
- Verify username and password are correct
- For Gmail, generate and use an App Password
- Check if "less secure app access" needs to be enabled (not recommended)
- Verify account is not locked or suspended

### Connection Errors

**Problem**: `connection refused` or timeout errors

**Solutions**:
- Verify host and port are correct
- Check firewall rules allow outbound connections on SMTP ports
- Test connectivity: `telnet smtp.example.com 587`
- Verify SMTP server is accessible from your network

### TLS/SSL Issues

**Problem**: Certificate verification or TLS handshake failures

**Solutions**:
- Update system CA certificates
- Verify SMTP server certificate is valid
- Use `tls_skip_verify: true` only for internal/trusted servers
- Check if STARTTLS is required vs direct SSL

### Sending Failures

**Problem**: Emails not being delivered or rejected

**Solutions**:
- Verify sender address is authorized
- Check SPF, DKIM, and DMARC records
- Ensure recipient addresses are valid
- Review SMTP server logs for rejection reasons
- Check if server requires verified sender domain

## Security Considerations

### Credential Management

- Never commit SMTP credentials to version control
- Use environment variables: `pass: ${SMTP_PASSWORD}`
- Consider using secrets management solutions
- Rotate credentials regularly

### Email Security

- Configure SPF records for your sending domain
- Set up DKIM signing if supported by SMTP provider
- Implement DMARC policies
- Monitor for unauthorized usage

### Network Security

- Use TLS/SSL for all connections
- Restrict SMTP access to authorized networks
- Monitor outbound SMTP traffic
- Implement rate limiting

## Migration Guide

### From Direct SMTP to Email Provider

If you're currently using direct SMTP libraries, migrating to the email provider offers:
- Unified configuration management
- Built-in retry and error handling
- Platform flexibility (easy switch to SES or ACS)
- Integration with workflow engine

### From Other Email Services

When migrating from other email services:
1. Update configuration to use SMTP provider
2. Verify sender address authorization
3. Test email delivery in non-production environment
4. Update workflow configurations
5. Monitor delivery success rates

## Related Documentation

- [Email Provider Overview](../email/)
- [AWS SES Provider](../email.ses/)
- [Azure Communication Services Provider](../email.acs/)
- [Workflow Configuration](/docs/configuration/workflows/)
