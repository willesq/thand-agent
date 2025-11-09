# Okta Provider

The Okta provider enables integration with Okta for identity and access management (IAM) operations. This provider supports role-based access control (RBAC), user management, and group operations.

## Features

- **RBAC Support**: Assign and revoke Okta administrator roles to users
- **Permissions Management**: Access to Okta's OAuth 2.0 scopes and admin permissions
- **Role Management**: Support for predefined Okta administrator roles
- **User Management**: List, get, and manage Okta users
- **Group Management**: List, get, and manage Okta groups
- **Search Functionality**: Fast search across permissions and roles using Bleve indexing

## Configuration

### Required Configuration

The Okta provider requires the following configuration parameters:

```yaml
provider: okta
config:
  org_url: https://your-org.okta.com    # Your Okta organization URL
  api_token: your_api_token_here        # Okta API token with appropriate permissions
```

### Example Configuration

```yaml
version: "1.0"
providers:
  okta-prod:
    name: Okta Production
    description: Production Okta environment
    provider: okta
    config:
      org_url: https://your-org.okta.com
      api_token: your_api_token_here
    enabled: true
```

### Getting an API Token

1. Log in to your Okta organization as an administrator
2. Navigate to **Security** → **API** → **Tokens**
3. Click **Create Token**
4. Provide a descriptive name for the token
5. Copy the token value (it will only be shown once)

**Note**: The API token should have sufficient permissions to manage users, groups, and roles in your Okta organization.

## Supported Administrator Roles

The Okta provider supports the following predefined administrator roles:

- **SUPER_ADMIN**: Full administrative access to the Okta organization
- **ORG_ADMIN**: Full administrative access except for managing super administrators
- **APP_ADMIN**: Can create and manage applications and their assignments
- **USER_ADMIN**: Can create and manage users and groups
- **GROUP_MEMBERSHIP_ADMIN**: Can manage group membership
- **GROUP_ADMIN**: Can create, manage, and delete groups
- **HELP_DESK_ADMIN**: Can reset passwords and MFA factors for users
- **READ_ONLY_ADMIN**: Can view all aspects of the Okta organization
- **MOBILE_ADMIN**: Can manage mobile device management settings
- **API_ACCESS_MANAGEMENT_ADMIN**: Can manage authorization servers and scopes
- **REPORT_ADMIN**: Can create and view reports

## Supported Permissions (OAuth 2.0 Scopes)

The provider includes support for Okta's OAuth 2.0 scopes and admin permissions:

### User Permissions
- `okta.users.read` - Read information about users
- `okta.users.manage` - Create, update, and delete users
- `okta.users.credentials.manage` - Manage user credentials
- And more...

### Group Permissions
- `okta.groups.read` - Read information about groups
- `okta.groups.manage` - Create, update, and delete groups

### Application Permissions
- `okta.apps.read` - Read information about applications
- `okta.apps.manage` - Create, update, and delete applications

### Additional Permissions
- Authorization Server permissions
- Event Hook permissions
- Identity Provider permissions
- Policy permissions
- Session permissions
- And many more...

See the full list in `permissions.go`.

## Usage Examples

### Assigning a Role to a User

```go
ctx := context.Background()

req := &models.AuthorizeRoleRequest{
    RoleRequest: &models.RoleRequest{
        User: &models.User{
            Email: "user@example.com",
        },
        Role: &models.Role{
            Name: "USER_ADMIN",
        },
    },
}

resp, err := provider.AuthorizeRole(ctx, req)
if err != nil {
    log.Fatal(err)
}
```

### Revoking a Role from a User

```go
ctx := context.Background()

req := &models.RevokeRoleRequest{
    RoleRequest: &models.RoleRequest{
        User: &models.User{
            Email: "user@example.com",
        },
        Role: &models.Role{
            Name: "USER_ADMIN",
        },
    },
    AuthorizeRoleResponse: resp, // Response from AuthorizeRole
}

_, err := provider.RevokeRole(ctx, req)
if err != nil {
    log.Fatal(err)
}
```

### Listing Users

```go
ctx := context.Background()

users, err := provider.ListIdentities(ctx, "john") // Optional filter
if err != nil {
    log.Fatal(err)
}

for _, user := range users {
    fmt.Printf("User: %s (%s)\n", user.Label, user.ID)
}
```

### Searching Permissions

```go
ctx := context.Background()

permissions, err := provider.ListPermissions(ctx, "users")
if err != nil {
    log.Fatal(err)
}

for _, perm := range permissions {
    fmt.Printf("Permission: %s - %s\n", perm.Name, perm.Description)
}
```

## Implementation Details

### Files

- **main.go**: Provider initialization and configuration
- **permissions.go**: Permission definitions and management
- **roles.go**: Role definitions and management
- **rbac.go**: Role-based access control operations
- **search.go**: Bleve search index for fast lookups

### Dependencies

The provider uses the official Okta Golang SDK v2:
```
github.com/okta/okta-sdk-golang/v2
```

### Caching

The Okta SDK includes built-in caching for API responses, which is enabled by default when creating the client.

## Security Considerations

1. **API Token Security**: Store API tokens securely using environment variables or secret management systems
2. **Least Privilege**: Use API tokens with the minimum required permissions
3. **Token Rotation**: Regularly rotate API tokens
4. **Audit Logging**: Monitor role assignments and changes through Okta's system logs
5. **Network Security**: Ensure communication with Okta is over HTTPS

## Limitations

1. Custom roles are not dynamically fetched (only predefined roles are supported)
2. Resource management is basic and can be extended for specific use cases
3. Some advanced Okta features may require additional implementation

## Future Enhancements

- Support for custom role definitions
- Advanced resource management (apps, policies, etc.)
- Support for Okta Workflows integration
- Enhanced group-based access control
- Support for Okta's advanced security features

## References

- [Okta Developer Documentation](https://developer.okta.com/docs/)
- [Okta API Reference](https://developer.okta.com/docs/reference/)
- [Okta Administrator Roles](https://help.okta.com/en-us/content/topics/security/administrators-admin-comparison.htm)
- [Okta OAuth 2.0 Scopes](https://developer.okta.com/docs/reference/api/oidc/#scopes)
- [Okta Golang SDK](https://github.com/okta/okta-sdk-golang)
