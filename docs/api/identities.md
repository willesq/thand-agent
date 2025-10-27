---
layout: default
title: Identity Management
parent: API Reference
nav_order: 9
---

# Identity Management

Manage user and group identities across providers.

## List Identities

Get available identities (users/groups) from identity providers.

**GET** `/identities`

### Query Parameters

- `q` - Search filter for identity name or email

### Response

```json
{
  "identities": [
    {
      "id": "alice@example.com",
      "label": "Alice Smith",
      "user": {
        "email": "alice@example.com",
        "name": "Alice Smith",
        "id": "user_123"
      }
    },
    {
      "id": "developers",
      "label": "Development Team",
      "type": "group"
    }
  ],
  "providers": 2
}
```

### Notes

- Only available in server mode
- Requires authentication
- Aggregates identities from all configured identity providers
- Removes duplicates across providers
- Respects provider permissions for the authenticated user