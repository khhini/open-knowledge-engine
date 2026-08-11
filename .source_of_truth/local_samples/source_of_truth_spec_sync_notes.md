# 📝 Meeting Notes: Source of Truth Local Simulation Spec V2 Sync

**Date:** 2026-08-11  
**Attendees:** Architecture Board (`human:arch-board`), Data Ops (`human:data-lead`), Systems Eng (`human:khhini`)

---

## 🎯 Architectural Decisions Updating `concepts/source_of_truth_local_simulation_spec`

The team reviewed `concepts/source_of_truth_local_simulation_spec` and approved three new technical updates:

### 1. Decision #1: Local S3 & GCS Mock Bucket Sync
- Extend `.source_of_truth/s3/` and `.source_of_truth/gcs/` mock metadata to include `bucket_name`, `region`, and `kms_encryption_key` fields.

### 2. Decision #2: Auto-Detect Tabular CSV/TSV Delimiters
- Update `pkg/sourcetruth/simulator.go` to auto-detect `,`, `;`, and `\t` delimiters when parsing local tabular files.

### 3. Decision #3: Strict JSON Schema Validation
- All `.meta.json` sidecar files in `.source_of_truth/` MUST validate against the Go struct models in `pkg/sourcetruth/models.go`.

---

## 🔗 Target Concept Updated
- Primary Target: [[concepts/source_of_truth_local_simulation_spec]]
