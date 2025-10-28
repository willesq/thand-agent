---
layout: default
title: Email Provider
description: Email provider for notifications and communication
parent: Providers
grand_parent: Configuration
---

# Email Provider

The Email provider enables email-based notifications and communication capabilities.

## Capabilities

- **Notifications**: Send email notifications
- **SMTP Integration**: Support for various SMTP servers
- **Template Support**: Customizable email templates
- **Authentication**: SMTP authentication support

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `smtp_host` | string | Yes | SMTP server hostname |
| `smtp_port` | number | Yes | SMTP server port |
| `username` | string | No | SMTP authentication username |
| `password` | string | No | SMTP authentication password |
| `from_email` | string | Yes | Sender email address |
| `tls` | boolean | No | Use TLS encryption |

## Example Configuration

```yaml
version: "1.0"
providers:
  email:
    name: Email Notifications
    description: SMTP email provider
    provider: email
    enabled: true
    config:
      smtp_host: smtp.gmail.com
      smtp_port: 587
      username: YOUR_EMAIL_USERNAME
      password: YOUR_EMAIL_PASSWORD
      from_email: agent@your-company.com
      tls: true
```

Common SMTP configurations:
- **Gmail**: smtp.gmail.com:587 (TLS)
- **Outlook**: smtp-mail.outlook.com:587 (TLS)
- **SendGrid**: smtp.sendgrid.net:587 (TLS)