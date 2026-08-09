---
type: PostgreSQL Table
title: Subscription Plans Table
description: Primary relational store for customer subscription tiers and billing schedules.
resource: postgresql://db.internal:5432/billing.subscription_plans
tags: [billing, subscriptions, postgres]
generated:
  by: process:schema-migrator
  at: 2026-08-05T11:00:00Z
verified:
  - actor: "human:khhini"
    at: "2026-08-06T15:00:00Z"
    notes: "Verified table constraints and foreign key indexes."
sources:
  - id: billing-schema
    resource: https://git.internal/billing/migrations/004_plans.sql
    author: human:billing-devs
    usage_count: 18500
status: stable
---

# Overview

Tracks recurring billing schedules, price points, and active lifecycle states for customer accounts.

# Schema

| Column Name | Type | Description |
| :--- | :--- | :--- |
| `subscription_id` | UUID | Primary Key |
| `customer_id` | STRING | Foreign key linking to [[concepts/customer_profiles]] |
| `plan_tier` | STRING | Tier level (BASIC, PRO, ENTERPRISE) |
| `monthly_rate_usd` | NUMERIC | Monthly rate billed to customer |

# Related Concepts

- Revenue metrics computed from this table are defined in [[concepts/revenue_metric]].
- Account owner details are queried from [[concepts/customer_profiles]].
