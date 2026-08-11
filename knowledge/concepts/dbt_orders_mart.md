---
type: dbt Model
title: Customer Orders Data Mart Model
description: Sanitized dbt analytics model joining raw orders with active subscription tiers and fraud risk scores.
resource: dbt://analytics-repo/models/marts/fct_orders.sql
tags: [dbt, analytics, data-mart, attestation]
generated:
  by: process:dbt-cloud-runner
  at: 2026-08-08T06:00:00Z
attestation:
  runtime: dbt-core/v1.7
  executor: process:airflow-dag-daily-orders
  computation: models/marts/fct_orders.sql
  parameters:
    target_schema: analytics_marts
    partition_date: "2026-08-08"
sources:
  - id: bq-orders-table
    resource: bq://enterprise-prod.sales.customer_orders
    title: Customer Orders BigQuery Table
    author: process:ga4-exporter
    usage_count: 14200
  - id: pg-plans-table
    resource: postgresql://db.internal:5432/billing/public/subscription_plans
    title: Subscription Plans PostgreSQL Table
    author: process:schema-migrator
    usage_count: 18500
status: stable
---

# Overview

The `fct_orders` dbt data mart materializes consolidated transactional records for downstream reporting, BI dashboards, and financial audit pipelines.

# Transformation Logic

This model joins transactional records from [[concepts/customer_orders]] with active subscription plan tiers in [[concepts/subscription_plans]] and calculates Net USD totals after applying risk filters from [[concepts/fraud_detector]].

```sql
SELECT
  o.order_id,
  o.customer_id,
  p.plan_tier,
  o.total_usd,
  o.placed_at
FROM {{ ref('stg_orders') }} o
LEFT JOIN {{ ref('stg_subscription_plans') }} p ON o.customer_id = p.customer_id;
```

# Related Concepts

- Source orders table: [[concepts/customer_orders]].
- Source subscription pricing: [[concepts/subscription_plans]].
- Downstream MRR calculations: [[concepts/revenue_metric]].
