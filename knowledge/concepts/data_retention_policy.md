---
type: Policy
title: GDPR Customer Data Retention Policy
description: Compliance policy governing data retention windows and deletion schedules.
tags: [compliance, gdpr, governance]
generated:
  by: human:legal-team
  at: 2026-08-07T09:00:00Z
verified:
  - actor: "human:khhini"
    at: "2026-08-07T10:00:00Z"
    notes: "Reviewed and approved for production data pipelines."
status: stable
stale_after: "2027-01-01"
---

# Policy Scope

Defines retention schedules and erasure profiles for personal identifiable information (PII) stored in [[concepts/customer_profiles]] and transactional history in [[concepts/customer_orders]].

# Retention Windows

1. **Active Profiles**: Retained while subscription in [[concepts/subscription_plans]] remains active.
2. **Inactive Accounts**: Purged 30 days after deletion request.
