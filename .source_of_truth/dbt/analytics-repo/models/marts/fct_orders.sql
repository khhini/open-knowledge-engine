-- Model: dbt://analytics-repo/models/marts/fct_orders.sql
-- Materialization: Increated / Table
-- Owner: team:analytics-eng

{{ config(
    materialized='table',
    partition_by={
      "field": "placed_at",
      "data_type": "timestamp",
      "granularity": "day"
    },
    cluster_by=["customer_id"]
) }}

WITH raw_orders AS (
    SELECT * FROM {{ source('sales', 'customer_orders') }}
),

active_plans AS (
    SELECT * FROM {{ source('billing', 'subscription_plans') }}
    WHERE status = 'ACTIVE'
)

SELECT
    o.order_id,
    o.customer_id,
    p.plan_tier,
    o.total_usd,
    o.status AS order_status,
    o.placed_at
FROM raw_orders o
LEFT JOIN active_plans p ON o.customer_id = p.customer_id;
