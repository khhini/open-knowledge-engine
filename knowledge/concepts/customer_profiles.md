---
type: API Endpoint
title: Customer Profiles API
description: Microservice endpoint for looking up customer identity and tier status.
resource: https://api.internal.enterprise/v1/customers
tags: [identity, crm, api]
generated:
  by: process:swagger-exporter
  at: 2026-08-02T12:00:00Z
verified:
  - actor: "human:khhini"
    at: "2026-08-05T14:30:00Z"
    notes: "Verified endpoint contract against production OpenAPI spec."
sources:
  - id: openapi-spec
    resource: https://git.internal/crm/spec.yaml
    author: human:dev-team
    usage_count: 14200
status: stable
---

# Overview

Provides RESTful access to customer demographics, membership tiers, and fraud risk scores.

# API Definition

`GET /v1/customers/{customer_id}`

### Response Payload

```json
{
  "customer_id": "cust_99812",
  "email": "user@example.com",
  "tier": "GOLD",
  "created_at": "2025-01-10T00:00:00Z"
}
```

# Cross References

- Used for joining transactions in [[concepts/customer_orders]].
