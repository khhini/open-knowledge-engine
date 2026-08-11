---
type: Kafka Topic
title: Real-Time Checkout Events Stream
description: High-throughput Kafka stream topic broadcasting real-time customer checkout attempts and payloads.
resource: file:///Users/kikihutapea/Documents/khhini-workspaces/open-knowledge-engine/.source_of_truth/local_samples/enterprise_data_platform_spec.txt#section=1
tags: [kafka, streaming, checkout, events]
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

# Topic Configuration

- **Topic Name**: `real_time_checkout_events`
- **Partition Count**: 16
- **Retention Window**: 7 Days (168 Hours)
- **Schema Format**: Apache Avro v1.4

# Consumer Lineage & Downstream Systems

1. Ingested by pipeline in [[concepts/order_fulfillment_pipeline]].
2. Evaluated for risk by [[concepts/fraud_detector]].
3. Materialized into BigQuery table [[concepts/customer_orders]].
