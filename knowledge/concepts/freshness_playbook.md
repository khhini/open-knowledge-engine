---
type: Playbook
title: Data Freshness Incident Response
description: Triage playbook when data pipelines fall behind SLA thresholds.
tags: [oncall, incident, playbook]
generated:
  by: human:khhini
  at: 2026-08-03T09:00:00Z
verified:
  - actor: "human:khhini"
    at: "2026-08-03T09:00:00Z"
    notes: "Initial playbook authoring."
status: stable
stale_after: "2026-11-01"
---

# Incident Response Steps

When an SLA alert fires for [[concepts/customer_orders]]:

1. **Check Ingestion Worker**: Navigate to the Airflow / Temporal worker dashboard.
2. **Inspect Backpressure**: Verify if upstream GA4 stream is throttling requests.
3. **Trigger Backfill**: If lag > 60 minutes, execute backfill job `job_order_backfill_v2`.
