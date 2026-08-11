---
type: ML Feature Store
title: Customer Churn ML Feature Store
description: Real-time ML feature store holding customer behavioral metrics for churn inference models.
resource: file:///Users/kikihutapea/Documents/khhini-workspaces/open-knowledge-engine/.source_of_truth/local_samples/enterprise_data_platform_spec.txt#section=2
tags: [ml, feature-store, churn, inference]
generated:
  by: process:local-file-ingestor/v1
  at: 2026-08-11T13:56:00Z
sources:
  - id: parent-architecture-spec
    resource: file:///Users/kikihutapea/Documents/khhini-workspaces/open-knowledge-engine/.source_of_truth/local_samples/enterprise_data_platform_spec.txt
    title: Enterprise Data Platform & Streaming Architecture Specification
    author: human:arch-board
status: stable
stale_after: "2027-01-01"
---

# Feature Store Architecture

- **Feature Store Name**: `customer_churn_feature_set`
- **Storage Engine**: Redis / Feast Feature Store
- **Refresh Frequency**: Hourly Streaming Refresh

# Feature Lineage & Derived Inputs

- Active plan pricing joined from [[concepts/subscription_plans]].
- Historical sales transactions joined from [[concepts/customer_orders]].
- Data retention schedules compliance verified against [[concepts/data_retention_policy]].
