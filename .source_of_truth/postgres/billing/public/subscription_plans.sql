-- Source: postgresql://db.internal:5432/billing/public/subscription_plans
-- Migration File: 004_plans.sql
-- Last Modified: 2026-08-05
-- Owner: team:billing-devs

CREATE TABLE public.subscription_plans (
    subscription_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id VARCHAR(64) NOT NULL,
    plan_tier VARCHAR(32) NOT NULL CHECK (plan_tier IN ('BASIC', 'PRO', 'ENTERPRISE')),
    monthly_rate_usd NUMERIC(10, 2) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'PAUSED', 'CANCELLED')),
    billing_cycle VARCHAR(16) NOT NULL DEFAULT 'MONTHLY',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sub_plans_customer ON public.subscription_plans(customer_id);
CREATE INDEX idx_sub_plans_status ON public.subscription_plans(status);
