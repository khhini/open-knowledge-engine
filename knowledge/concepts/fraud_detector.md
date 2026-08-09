---
type: Attested Computation
title: Transaction Fraud Risk Model
description: Sanctioned ML scoring model for detecting high-risk orders.
tags: [fraud, security, ml, attestation]
generated:
  by: process:ci-ml-pipeline
  at: 2026-08-06T14:00:00Z
attestation:
  runtime: wasm-wasmer-v3
  executor: process:fraud-scorer-v2
  computation: sha256-model-eval-v1.4
  parameters:
    threshold_score: 0.85
status: stable
---

# Overview

Evaluates incoming orders in [[concepts/customer_orders]] against user profile signals in [[concepts/customer_profiles]] to compute a real-time risk score.

# Attestation Receipt

Computations run inside a WASM runtime returning cryptographic execution receipts verified deterministically without LLM intervention.

# Related Concepts

- Evaluates order risk for [[concepts/customer_orders]].
- Ingests risk factors from [[concepts/customer_profiles]].
