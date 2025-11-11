---
layout: default
title: Email SES
description: AWS SES email provider configuration
parent: Providers
grand_parent: Configuration
---

# Email SES Provider

Amazon Simple Email Service (SES) integration for scalable, cost-effective email delivery.

## Overview

The AWS SES email provider (`email.ses`) enables email notifications through Amazon's Simple Email Service. SES is a cloud-based email sending service designed for high-volume email delivery with excellent deliverability and competitive pricing.

## Capabilities

- **High Deliverability**: Amazon's infrastructure and reputation
- **Scalability**: Handle millions of emails per day
- **Cost-Effective**: Pay only for what you use
- **AWS Integration**: Native integration with AWS services
- **HTML & Plain Text**: Support for both email formats
- **Multiple Recipients**: Send to multiple addresses simultaneously

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `platform` | string | Yes | Set to `ses` |
| `from` | string | Yes | Default sender email address (must be verified in SES) |
| `region` | string | No | AWS region (default: us-east-1) |
| `profile` | string | No | AWS profile name for authentication |
| `access_key_id` | string | No | AWS access key ID for static credentials |
| `secret_access_key` | string | No | AWS secret access key for static credentials |
| `endpoint` | string | No | Custom endpoint URL (e.g., for LocalStack testing) |
| `imds_disable` | boolean | No | Disable EC2 IMDSv2 for credential retrieval |

## Authentication Methods

The SES provider supports multiple AWS authentication methods in order of precedence:

1. **Static Credentials**: `access_key_id` and `secret_access_key`
2. **AWS Profile**: `profile` configuration
3. **IAM Role**: EC2 instance profile or ECS task role
4. **Environment Variables**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
5. **Default Credential Chain**: Standard AWS credential resolution

## Example Configurations

### Using IAM Role (Recommended)

```yaml
version: "1.0"
providers:
  email-ses:
    name: AWS SES Email
    description: Amazon SES email provider
    provider: email  # can use 'email' with platform: ses, or use 'email.ses' directly
    enabled: true
    config:
      platform: ses
      from: noreply@example.com
      region: us-east-1
```

**Alternative** (using direct provider name, no platform needed):

```yaml
version: "1.0"
providers:
  email-ses:
    name: AWS SES Email
    description: Amazon SES email provider
    provider: email.ses  # directly specify email.ses
    enabled: true
    config:
      from: noreply@example.com
      region: us-east-1
```

### Using AWS Profile

```yaml
version: "1.0"
providers:
  email-ses:
    name: AWS SES Email
    description: Amazon SES email provider with profile
    provider: email
    enabled: true
    config:
      platform: ses
      from: noreply@example.com
      region: us-west-2
      profile: production
```

### Using Static Credentials

```yaml
version: "1.0"
providers:
  email-ses:
    name: AWS SES Email
    description: Amazon SES with static credentials
    provider: email
    enabled: true
    config:
      platform: ses
      from: noreply@example.com
      region: eu-west-1
      access_key_id: <your-access-key-id>
      secret_access_key: <your-secret-access-key>
```

### Using Custom Endpoint (LocalStack)

```yaml
version: "1.0"
providers:
  email-ses-local:
    name: AWS SES Local Testing
    description: LocalStack SES for development
    provider: email
    enabled: true
    config:
      platform: ses
      from: test@example.com
      region: us-east-1
      endpoint: http://localhost:4566
      access_key_id: test
      secret_access_key: test
```

## AWS Setup Requirements

### 1. Verify Email Address or Domain

**Single Email Verification:**
```bash
aws ses verify-email-identity --email-address noreply@example.com --region us-east-1
```

**Domain Verification:**
```bash
aws ses verify-domain-identity --domain example.com --region us-east-1
```

### 2. Request Production Access

SES starts in sandbox mode with limitations:
- Can only send to verified addresses
- Limited to 200 messages per day
- Maximum 1 message per second

Request production access through AWS Console:
1. Navigate to SES Console
2. Select "Account Dashboard"
3. Click "Request production access"
4. Complete the form

### 3. Configure IAM Permissions

Minimum IAM policy for sending emails:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ses:SendEmail",
        "ses:SendRawEmail"
      ],
      "Resource": "*"
    }
  ]
}
```

More restrictive policy with specific sender:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ses:SendEmail",
        "ses:SendRawEmail"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "ses:FromAddress": "noreply@example.com"
        }
      }
    }
  ]
}
```

### 4. Configure Email Authentication

**SPF Record:**
```
"v=spf1 include:amazonses.com ~all"
```

**DKIM Setup:**
```bash
aws ses verify-domain-dkim --domain example.com --region us-east-1
```

Add the returned CNAME records to your DNS.

**DMARC Record:**
```
"v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"
```

## Usage in Workflows

```yaml
workflows:
  - name: access-notification
    steps:
      - task: notify
        type: email
        config:
          provider: email-ses
          subject: "Access Granted"
          to:
            - user@example.com
          body:
            text: "Your access request has been approved."
            html: |
              <html>
                <body>
                  <h1>Access Granted</h1>
                  <p>Your access request has been approved.</p>
                </body>
              </html>
```

## Best Practices

### Security

1. **Use IAM Roles**: Prefer IAM roles over static credentials
2. **Least Privilege**: Grant only necessary SES permissions
3. **Verified Senders**: Only use verified email addresses or domains
4. **Monitor Usage**: Set up CloudWatch alarms for unusual activity
5. **Rotate Credentials**: Regularly rotate access keys if using static credentials

### Deliverability

1. **Warm Up**: Gradually increase sending volume for new accounts
2. **Email Authentication**: Configure SPF, DKIM, and DMARC
3. **List Hygiene**: Remove bounces and complaints promptly
4. **Engagement**: Monitor open and click rates
5. **Reputation**: Maintain low bounce and complaint rates

### Cost Optimization

1. **Regional Pricing**: Choose appropriate AWS region (pricing varies)
2. **Batch Operations**: Group email sends when possible
3. **Monitor Spend**: Use AWS Cost Explorer to track SES costs
4. **Clean Lists**: Don't send to invalid addresses

### Performance

1. **Regional Endpoints**: Use SES in regions closest to your infrastructure
2. **Concurrent Sending**: SES handles high concurrency well
3. **Error Handling**: Implement exponential backoff for throttling
4. **Monitoring**: Track send rates and delivery metrics

## Troubleshooting

### Verification Issues

**Problem**: `MessageRejected: Email address is not verified`

**Solutions**:
- Verify email address or domain in SES console
- Check verification status: `aws ses get-identity-verification-attributes`
- Ensure using correct AWS region
- Wait for domain DNS propagation (up to 72 hours)

### Sandbox Limitations

**Problem**: `MessageRejected: Email address is not verified` (recipients)

**Solutions**:
- Request production access from SES console
- Verify recipient addresses for testing in sandbox
- Check account is out of sandbox: SES Console > Account Dashboard

### Authentication Errors

**Problem**: `UnauthorizedOperation` or credential errors

**Solutions**:
- Verify IAM permissions include `ses:SendEmail`
- Check AWS credentials are correctly configured
- Ensure region matches SES setup region
- Test credentials: `aws ses get-send-quota --region us-east-1`

### Sending Rate Exceeded

**Problem**: `Throttling: Maximum sending rate exceeded`

**Solutions**:
- Check current limits: `aws ses get-send-quota`
- Implement exponential backoff retry logic
- Request sending limit increase
- Distribute sending over time

### Bounce and Complaint Issues

**Problem**: High bounce or complaint rates

**Solutions**:
- Set up bounce and complaint notifications via SNS
- Remove hard bounces immediately
- Implement feedback loop processing
- Maintain clean email lists
- Review content for spam indicators

## Monitoring and Metrics

### CloudWatch Metrics

Monitor these SES metrics:
- `Send`: Number of emails sent
- `Bounce`: Bounce rate
- `Complaint`: Complaint rate
- `Delivery`: Successful deliveries
- `Reject`: Rejected emails

### Example CloudWatch Alarm

```bash
aws cloudwatch put-metric-alarm \
  --alarm-name ses-high-bounce-rate \
  --alarm-description "Alert on high SES bounce rate" \
  --metric-name Reputation.BounceRate \
  --namespace AWS/SES \
  --statistic Average \
  --period 300 \
  --threshold 0.05 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 1
```

### Sending Statistics

Check sending statistics:
```bash
aws ses get-send-statistics --region us-east-1
```

## Cost Estimation

SES pricing (as of 2024, varies by region):
- **First 62,000 emails/month**: Free (if sent from EC2)
- **Additional emails**: $0.10 per 1,000 emails
- **Attachments**: $0.12 per GB
- **Dedicated IPs**: $24.95/month per IP

Example monthly costs:
- 100,000 emails: ~$3.80
- 1,000,000 emails: ~$94.00
- 10,000,000 emails: ~$994.00

## Migration Guide

### From SMTP to SES

1. Verify sender addresses/domains in SES
2. Update configuration to use `platform: ses`
3. Configure IAM permissions
4. Test in sandbox environment
5. Request production access
6. Update DNS records (SPF, DKIM, DMARC)
7. Gradually migrate traffic

### From Other Email Services

Benefits of migrating to SES:
- Lower costs at scale
- Better deliverability with AWS infrastructure
- Native AWS integration
- Detailed metrics and monitoring

## Related Documentation

- [Email Provider Overview](../email/)
- [Email SMTP Provider](../email.smtp/)
- [Azure Communication Services Provider](../email.acs/)
- [AWS SES Documentation](https://docs.aws.amazon.com/ses/)
