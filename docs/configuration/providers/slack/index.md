---
layout: default
title: Slack
description: Slack provider for team communication and notifications
parent: Providers
grand_parent: Configuration
---

# Slack Provider

The Slack provider enables integration with Slack for notifications and team communication.

## Capabilities

- **Notifications**: Send notifications to Slack channels
- **Team Integration**: Access to Slack workspace and user information
- **Bot Integration**: Support for Slack bot tokens and app integration

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `token` | string | Yes | Slack bot token or OAuth token |
| `channel` | string | No | Default channel for notifications |
| `webhook_url` | string | No | Slack webhook URL (alternative to token) |

## Example Configuration

```yaml
version: "1.0"
providers:
  slack:
    name: Slack
    description: Slack notifications
    provider: slack
    enabled: true
    config:
      token: YOUR_SLACK_BOT_TOKEN
      channel: "#general"
```

For detailed setup instructions, refer to the [Slack API documentation](https://api.slack.com/).