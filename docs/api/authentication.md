---
layout: default
title: Authentication
parent: API Reference
nav_order: 2
---

# Authentication Endpoints

OAuth2 authentication flows and session management.

## Auth Request

Initiate OAuth2 authentication flow for a provider.

**GET** `/auth/request/{provider}`

### Query Parameters

- `callback` - Callback URL after authentication

### Response

Redirects to provider's OAuth2 authorization URL.

## Auth Callback

Handle OAuth2 callback from provider.

**GET** `/auth/callback/{provider}`

### Query Parameters

- `code` - Authorization code from provider
- `state` - State parameter for CSRF protection

### Response

Redirects to callback URL or shows success page.

## Logout

Clear authentication session.

**GET** `/auth/logout/{provider}` or `/auth/logout`

Clears authentication session for specific provider or all providers.