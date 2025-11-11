---
layout: default
title: Email ACS
description: Azure Communication Services email provider configuration
parent: Providers
grand_parent: Configuration
---

# Email ACS Provider

Azure Communication Services (ACS) Email integration for enterprise-grade email delivery.

## Overview

The Azure Communication Services email provider (`email.acs`) enables email notifications through Microsoft's Azure Communication Services platform. ACS provides reliable, scalable email delivery with enterprise features and tight integration with Azure services.

## Capabilities

- **Enterprise Grade**: Built for business-critical communications
- **Azure Integration**: Native integration with Azure ecosystem
- **Scalability**: Handle high-volume email delivery
- **Compliance**: Meet regulatory and compliance requirements
- **HTML & Plain Text**: Support for both email formats
- **Multiple Recipients**: Send to multiple addresses simultaneously

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `platform` | string | Yes | Set to `acs` |
| `endpoint` | string | Yes | Azure Communication Services endpoint URL |
| `from` | string | Yes | Default sender email address (must be configured in ACS) |
| `client_id` | string | No | Azure service principal client ID |
| `client_secret` | string | No | Azure service principal client secret |
| `tenant_id` | string | No | Azure tenant ID |

## Authentication Methods

The ACS provider supports multiple Azure authentication methods:

1. **Service Principal**: `client_id`, `client_secret`, and `tenant_id`
2. **Managed Identity**: Available when running on Azure resources (VM, Container, Function)
3. **Default Credential Chain**: Standard Azure credential resolution

## Example Configurations

### Using Managed Identity (Recommended for Azure)

```yaml
version: "1.0"
providers:
  email-acs:
    name: Azure Communication Services Email
    description: Azure ACS email provider
    provider: email  # can use 'email' with platform: acs, or use 'email.acs' directly
    enabled: true
    config:
      platform: acs
      endpoint: https://my-acs-resource.communication.azure.com
      from: DoNotReply@acs.example.com
```

**Alternative** (using direct provider name, no platform needed):

```yaml
version: "1.0"
providers:
  email-acs:
    name: Azure Communication Services Email
    description: Azure ACS email provider
    provider: email.acs  # directly specify email.acs
    enabled: true
    config:
      endpoint: https://my-acs-resource.communication.azure.com
      from: DoNotReply@acs.example.com
```

### Using Service Principal

```yaml
version: "1.0"
providers:
  email-acs:
    name: Azure Communication Services Email
    description: Azure ACS with service principal
    provider: email
    enabled: true
    config:
      platform: acs
      endpoint: https://my-acs-resource.communication.azure.com
      from: DoNotReply@acs.example.com
      client_id: <your-client-id>
      client_secret: <your-client-secret>
      tenant_id: <your-tenant-id>
```

### Using Environment Variables

```yaml
version: "1.0"
providers:
  email-acs:
    name: Azure Communication Services Email
    description: Azure ACS with environment variables
    provider: email
    enabled: true
    config:
      platform: acs
      endpoint: ${ACS_ENDPOINT}
      from: DoNotReply@acs.example.com
      client_id: ${AZURE_CLIENT_ID}
      client_secret: ${AZURE_CLIENT_SECRET}
      tenant_id: ${AZURE_TENANT_ID}
```

## Azure Setup Requirements

### 1. Create Communication Services Resource

**Using Azure Portal:**
1. Navigate to Azure Portal
2. Create a new "Communication Services" resource
3. Note the endpoint URL from the resource overview

**Using Azure CLI:**
```bash
az communication create \
  --name my-acs-resource \
  --resource-group my-resource-group \
  --location global \
  --data-location UnitedStates
```

### 2. Configure Email Domain

**Add Email Domain:**
1. In your Communication Services resource, select "Domains"
2. Click "Add domain"
3. Choose between:
   - **Azure Managed Domain**: Quick setup, limited customization
   - **Custom Domain**: Your own domain, requires DNS configuration

**For Custom Domain:**
```bash
az communication email domain create \
  --domain-name example.com \
  --resource-group my-resource-group \
  --email-service-name my-email-service
```

### 3. Verify Domain Ownership

Add required DNS records:

**For Azure Domain:**
- Use provided Azure subdomain (e.g., `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.azurecomm.net`)
- No DNS configuration required

**For Custom Domain:**
- Add TXT record for domain verification
- Add SPF, DKIM, and DMARC records
- Verification can take up to 48 hours

Example DNS records:
```
# Verification
TXT @ azure-domain-verification=xxxxxxxxxxxxx

# SPF
TXT @ "v=spf1 include:spf.protection.outlook.com -all"

# DKIM (add records provided by Azure)
CNAME selector1._domainkey selector1-{domain}._domainkey.{region}.azurecomm.net
CNAME selector2._domainkey selector2-{domain}._domainkey.{region}.azurecomm.net

# DMARC
TXT _dmarc "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"
```

### 4. Configure Sender Addresses

**Add sender address:**
```bash
az communication email domain sender-username create \
  --domain-name example.com \
  --resource-group my-resource-group \
  --email-service-name my-email-service \
  --sender-username DoNotReply \
  --display-name "My Application"
```

### 5. Set Up RBAC Permissions

**Using Service Principal:**

Create service principal:
```bash
az ad sp create-for-rbac --name acs-email-sender
```

Assign Contributor role:
```bash
az role assignment create \
  --assignee <service-principal-id> \
  --role "Contributor" \
  --scope /subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Communication/CommunicationServices/<acs-resource-name>
```

**Using Managed Identity:**

Enable managed identity on your Azure resource (VM, App Service, etc.):
```bash
az vm identity assign --name my-vm --resource-group my-resource-group
```

Assign permissions:
```bash
az role assignment create \
  --assignee <managed-identity-principal-id> \
  --role "Contributor" \
  --scope /subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Communication/CommunicationServices/<acs-resource-name>
```

## Usage in Workflows

```yaml
workflows:
  - name: user-notification
    steps:
      - task: notify
        type: email
        config:
          provider: email-acs
          subject: "Welcome to Our Platform"
          to:
            - user@example.com
          body:
            text: "Thank you for joining our platform."
            html: |
              <html>
                <body>
                  <h1>Welcome!</h1>
                  <p>Thank you for joining our platform.</p>
                </body>
              </html>
```

## Best Practices

### Security

1. **Use Managed Identity**: Prefer managed identity over service principals when running on Azure
2. **Least Privilege**: Grant only necessary permissions (use custom roles if needed)
3. **Secure Endpoints**: Restrict access to ACS resources using network rules
4. **Rotate Credentials**: Regularly rotate service principal secrets
5. **Monitor Access**: Enable diagnostic logging for security monitoring

### Deliverability

1. **Domain Verification**: Always verify custom domains
2. **Email Authentication**: Configure SPF, DKIM, and DMARC
3. **Sender Reputation**: Monitor bounce and complaint rates
4. **Content Quality**: Avoid spam triggers in email content
5. **Engagement**: Track and improve email engagement metrics

### Compliance

1. **Data Residency**: Choose appropriate data location during resource creation
2. **Audit Logging**: Enable diagnostic settings for compliance
3. **Privacy**: Implement unsubscribe mechanisms
4. **GDPR**: Follow GDPR requirements for EU recipients
5. **Retention**: Configure appropriate data retention policies

### Cost Optimization

1. **Resource Planning**: Choose appropriate service tier
2. **Monitor Usage**: Track email volumes and costs
3. **Batch Operations**: Group email sends when possible
4. **Clean Lists**: Remove invalid addresses to avoid waste
5. **Regional Resources**: Use resources in cost-effective regions

## Troubleshooting

### Endpoint Configuration Issues

**Problem**: `failed to send email via Azure Communication Services: 401`

**Solutions**:
- Verify endpoint URL is correct (check Azure portal)
- Ensure endpoint includes full URL with protocol (https://)
- Check service principal or managed identity has permissions
- Verify tenant ID is correct

### Authentication Failures

**Problem**: `failed to get access token`

**Solutions**:
- Verify client ID, client secret, and tenant ID are correct
- Check service principal is not expired
- Ensure managed identity is enabled and assigned
- Verify RBAC permissions are correctly set
- Test authentication: `az login --service-principal`

### Domain Not Verified

**Problem**: `Sender address not verified` or domain validation errors

**Solutions**:
- Check domain verification status in Azure portal
- Verify DNS records are correctly configured
- Wait for DNS propagation (up to 48 hours)
- Ensure sender address matches configured domain
- Check domain is not expired or suspended

### API Rate Limiting

**Problem**: `TooManyRequests` or `429` status codes

**Solutions**:
- Implement exponential backoff retry logic
- Reduce sending rate
- Contact Azure support for limit increases
- Consider upgrading service tier

### Email Delivery Failures

**Problem**: Emails not being delivered

**Solutions**:
- Check sender address is properly configured
- Verify recipient addresses are valid
- Review Azure Communication Services logs
- Check DNS records (SPF, DKIM, DMARC)
- Monitor for bounces and complaints
- Ensure domain reputation is good

## Monitoring and Diagnostics

### Enable Diagnostic Settings

```bash
az monitor diagnostic-settings create \
  --name acs-diagnostics \
  --resource /subscriptions/<sub-id>/resourceGroups/<rg>/providers/Microsoft.Communication/CommunicationServices/<acs-name> \
  --logs '[{"category": "EmailSendMailOperational", "enabled": true}]' \
  --workspace /subscriptions/<sub-id>/resourceGroups/<rg>/providers/Microsoft.OperationalInsights/workspaces/<workspace-name>
```

### Key Metrics to Monitor

- **EmailDeliveryAttempts**: Total email send attempts
- **EmailDeliverySuccess**: Successfully delivered emails
- **EmailDeliveryFailure**: Failed email deliveries
- **EmailBounce**: Bounced emails
- **EmailComplaint**: Complaint notifications

### Example Log Analytics Query

```kusto
ACSEmailSendMailOperational
| where TimeGenerated > ago(24h)
| summarize 
    Attempts = count(),
    Success = countif(ResultType == "Success"),
    Failures = countif(ResultType == "Failure")
    by bin(TimeGenerated, 1h)
| project TimeGenerated, Attempts, Success, Failures, SuccessRate = (Success * 100.0) / Attempts
```

## Cost Estimation

Azure Communication Services Email pricing (as of 2024):

- **Email Messages**: $0.25 per 1,000 messages
- **No minimum**: Pay only for what you use
- **Free Tier**: First 500 emails/month free

Example monthly costs:
- 10,000 emails: ~$2.50
- 100,000 emails: ~$25.00
- 1,000,000 emails: ~$250.00

**Note**: Pricing may vary by region and is subject to change.

## Migration Guide

### From SMTP to ACS

1. Create Azure Communication Services resource
2. Configure email domain and sender addresses
3. Set up authentication (managed identity or service principal)
4. Update configuration to use `platform: acs`
5. Test email delivery
6. Gradually migrate traffic
7. Monitor delivery metrics

### From Other Cloud Providers

Benefits of migrating to ACS:
- Tight Azure integration
- Enterprise compliance features
- Simplified authentication with managed identity
- Comprehensive monitoring and diagnostics
- Predictable pricing

## Security Considerations

### Network Security

- Use Azure Private Link for private connectivity
- Configure firewall rules to restrict access
- Enable network service tags in NSGs
- Use VNet service endpoints where applicable

### Identity and Access

- Prefer managed identity over service principals
- Use Azure AD Conditional Access policies
- Implement just-in-time access
- Regular access reviews

### Data Protection

- Enable encryption at rest and in transit
- Configure appropriate data retention
- Implement backup and disaster recovery
- Follow data residency requirements

## Related Documentation

- [Email Provider Overview](../email/)
- [Email SMTP Provider](../email.smtp/)
- [AWS SES Provider](../email.ses/)
- [Azure Communication Services Documentation](https://docs.microsoft.com/azure/communication-services/)
- [Azure Email Communication Services](https://docs.microsoft.com/azure/communication-services/concepts/email/email-overview)
