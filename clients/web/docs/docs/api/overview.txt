# API Overview

The rtylr API provides programmatic access to your organization's data and operations. Use it to build custom integrations, automate workflows, or extend rtylr's functionality.

## Base URL

```
https://api.rtylr.com/v1
```

## Authentication

All API requests require a Bearer token in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

Generate API keys in **Hub > Settings > API Keys**. Each key is scoped to specific apps and permissions.

## Rate Limits

API requests are rate-limited per organization:

- **Standard plan:** 100 requests per minute
- **Pro plan:** 500 requests per minute
- **Enterprise:** Custom limits

Rate limit headers are included in every response:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1700000000
```

## Response Format

All responses are JSON. Successful responses include a `data` field:

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 150
  }
}
```

## Errors

Error responses include a `code` and `message`:

```json
{
  "error": {
    "code": "not_found",
    "message": "The requested resource was not found."
  }
}
```

## Pagination

List endpoints support cursor-based pagination using `page` and `per_page` query parameters. The default page size is 25, with a maximum of 100.
