---
name: local-file-ingestor
description: Ingests local files (.pdf, .csv, .md, .sql, .yaml) into OKF concepts, decomposing multi-concept files, discovering new types, and generating bidirectional wikilinks.
---

# Local File Ingestion Skill

This skill guides AI agents in ingesting local documents, data files, and code schemas into the **OKF Knowledge Engine**.

## Step-by-Step Execution Workflow

### 1. Inspect Local Source
When given a local file path (e.g., `/Users/username/docs/spec.pdf`):
- Construct a `file://` URI: `file:///Users/username/docs/spec.pdf`
- Invoke the MCP tool `inspect_source_of_truth` with `uri: "file:///Users/username/docs/spec.pdf"`.
- Review the returned text snippets, page counts, CSV columns, or schema metadata.

### 2. Decompose Content & Identify Concepts
- Determine if the file contains one or multiple domain entities.
- For each distinct entity:
  - Assign a standard concept type (`BigQuery Table`, `PostgreSQL Table`, `Policy`, `Playbook`, `dbt Model`, `Data Pipeline`) or define a concise PascalCase novel type (e.g. `ML Feature Store`, `Kafka Topic`, `Data Contract`).
  - Formulate a clear `concept_id` (e.g. `concepts/customer_churn_feature_set`).

### 3. Deduplicate Existing Knowledge
- For each concept:
  - Invoke `search_knowledge` with the entity title.
  - If a matching concept exists: Use `update_concept` with `append_body` or `body` replacement.
  - If no matching concept exists: Prepare for `create_concept`.

### 4. Format OKF v0.2 Frontmatter & Body
- Ensure YAML frontmatter includes:
  ```yaml
  type: <ConceptType>
  title: <Concept Title>
  description: <Summary>
  resource: file://<absolute-file-path>#<section-anchor>
  tags: [<relevant-tags>]
  generated:
    by: process:local-file-ingestor/v1
    at: <ISO-timestamp>
  sources:
    - id: parent-local-file
      resource: file://<absolute-file-path>
      title: <Original Document Title>
      author: human:<owner>
  status: draft
  stale_after: "YYYY-MM-DD"
  ```
- Insert `[[wikilinks]]` referencing related existing or sibling concepts.

### 5. Programmatic Ingestion via MCP
- Invoke `create_concept` or `update_concept`.
- Invoke `list_broken_links` to audit wikilink integrity.
- Submit a machine attestation receipt via `verify_concept`:
  ```json
  {
    "concept_id": "concepts/<target-id>",
    "actor": "process:local-file-ingestor/v1",
    "notes": "Ingested from local file file://<absolute-file-path>"
  }
  ```
