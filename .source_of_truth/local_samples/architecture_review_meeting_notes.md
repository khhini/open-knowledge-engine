# 📝 Architecture Review Meeting Notes: Source of Truth Simulator V2 Upgrades

**Date:** 2026-08-11  
**Location:** Engineering Sync / Architecture Board  
**Attendees:** Architecture Board (`human:arch-board`), Data Operations Lead (`human:data-lead`), Security Ops (`human:sec-ops`)

---

## 🎯 Executive Summary & Architectural Decisions

During today's architecture review, the team evaluated the Source of Truth Local File Storage Simulator and approved three key upgrades to the specification:

### 1. Decision #1: Add Embedded Database Simulation (SQLite & DuckDB)
- **Previous Rule**: Simulator only supported `.pdf.txt`, `.csv`, `.json`, and `.sql` text files under `gdrive/`, `gcs/`, `s3/`, `bigquery/`, and `postgres/`.
- **New Decision**: Extend `.source_of_truth/` directory structure to support embedded local database files under `.source_of_truth/sqlite/` and `.source_of_truth/duckdb/`.
- **Target URIs**: `sqlite://billing.db/transactions` and `duckdb://analytics.duckdb/marts/fct_orders`.

### 2. Decision #2: Enforce 90-Day Freshness TTL for Local Simulator Files
- **Previous Rule**: Local file metadata had no default TTL expiration.
- **New Decision**: All simulated local files with `last_modified` older than 90 days without an explicit `stale_after` date MUST automatically evaluate to **`TierStale`** trust tier.

### 3. Decision #3: Expand Snippet Preview Limit
- **Previous Rule**: Multi-source inspection (`include_source_truth: true`) truncated text snippets at 500 characters.
- **New Decision**: Increase text snippet preview limit to **1,000 characters** to provide richer context for LLM agent reasoning.

---

## 🔗 Impacted Specifications & Concepts
- **Specification Document**: [[docs/SOURCE_OF_TRUTH_LOCAL_SIMULATION_SPEC.md]]
- **Code Handler**: [[pkg/sourcetruth/simulator.go]]
