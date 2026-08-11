---
type: Data Pipeline
title: Daily Order Fulfillment & Billing Sync Pipeline
description: Orchestrated Airflow workflow extracting web transactions, verifying customer identities, and updating billing ledgers.
resource: airflow://daily_etl_pipeline/extract_orders
tags: [airflow, pipeline, etl, orchestration]
generated:
  by: human:data-eng-team
  at: 2026-08-09T08:00:00Z
verified:
  - actor: "human:khhini"
    at: "2026-08-09T09:30:00Z"
    notes: "Verified DAG execution graph and alert callbacks."
sources:
  - id: customer-api-endpoint
    resource: https://api.internal.enterprise/v1/customers
    title: Customer Profiles Microservice API
    author: human:crm-team
    usage_count: 14200
  - id: quarterly-billing-spreadsheet
    resource: https://docs.google.com/spreadsheets/d/9Z8Y7X6W5V4U3T2S1R0Q/edit
    title: Q3 Subscription Billing Report
    author: human:analytics-team
    usage_count: 5200
status: stable
stale_after: "2027-01-01"
---

# Pipeline Architecture

The order fulfillment pipeline runs daily at 02:00 UTC to ingest completed sales transactions into [[concepts/customer_orders]] and synchronize account tiers in [[concepts/customer_profiles]].

```mermaid
graph TD
    A["Web/Mobile Checkout API"] --> B["Airflow DAG: daily_etl_pipeline"]
    B --> C["Verify Identity via [[concepts/customer_profiles]]"]
    B --> D["Risk Scoring via [[concepts/fraud_detector]]"]
    C & D --> E["Load into [[concepts/customer_orders]]"]
```

# Failure & SLA Procedures

In the event of DAG failure or pipeline lag exceeding 60 minutes, immediately initiate the incident response steps in [[concepts/freshness_playbook]].
