---
type: BigQuery Table
title: Customer Orders Table
description: Consolidated customer orders across all web and mobile sales channels.
resource: https://console.cloud.google.com/bigquery?p=enterprise&d=sales&t=orders
tags: [sales, revenue, analytics]
generated:
  by: reference_agent/gemini-2.5-pro
  at: 2026-08-01T08:00:00Z
sources:
  - id: ga4-export
    resource: https://developers.google.com/analytics/bigquery/export-schema
    title: GA4 BigQuery Export Schema
    author: team:ga4-docs
    usage_count: 5200
    last_modified: 2026-06-15
status: stable
---

# Overview

The `customer_orders` table contains real-time transactional records for completed and pending customer orders.

# Schema

| Column Name | Data Type | Nullable | Description |
| :--- | :--- | :--- | :--- |
| `order_id` | STRING | NO | Globally unique order identifier |
| `customer_id` | STRING | NO | Foreign key reference to [[concepts/customer_profiles]] |
| `total_usd` | NUMERIC | NO | Total transaction amount in USD |
| `placed_at` | TIMESTAMP | NO | Order placement timestamp in UTC |

# Related Concepts

- Customer identity details are maintained in [[concepts/customer_profiles]].
- In case of pipeline lag, follow [[concepts/freshness_playbook]].
