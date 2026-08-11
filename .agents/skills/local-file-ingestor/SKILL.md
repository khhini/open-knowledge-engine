---
name: local-file-ingestor
description: Reads local documents, data files, and code schemas (.pdf, .csv, .md, .sql, .yaml), semantically identifies domain concepts & types from content, decomposes multi-concept files, and ingests them into OKF via MCP tools.
---

# Local File Ingestion Skill

This skill guides AI agents in semantically analyzing and ingesting local files into the **OKF Knowledge Engine**.

## Ingestion Principles & Rules

1. **Content-Based Semantic Identification**:
   - NEVER create a concept ID based solely on a raw filename (e.g. do NOT turn `architecture_v2_final.pdf` into `concepts/architecture_v2_final`).
   - READ the actual content of the file and identify real domain entities (e.g. `concepts/customer_orders`, `concepts/data_retention_policy`, `concepts/churn_feature_store`).

2. **Multi-Concept Atomic Decomposition**:
   - If a file contains multiple domain entities (e.g. a PDF describing 3 database tables, a policy, and a data pipeline), decompose the document into separate, atomic concept files in `knowledge/concepts/`.
   - Each extracted concept MUST point back to the parent local file URI in its `resource` or `sources` list:
     ```yaml
     resource: file:///absolute/path/to/parent_doc.pdf#section=3
     sources:
       - id: parent-file
         resource: file:///absolute/path/to/parent_doc.pdf
         title: Parent Document Title
         author: human:document-owner
     ```

3. **Dynamic Concept Type Discovery**:
   - Use standard concept types (`BigQuery Table`, `PostgreSQL Table`, `Metric`, `Policy`, `API Endpoint`, `Playbook`, `dbt Model`, `Data Pipeline`, `Attested Computation`).
   - If the file describes a novel domain entity (e.g. `Kafka Topic`, `ML Feature Store`, `Data Contract`), define a new concise Title Case concept type.

4. **Bidirectional Wikilink Weaving**:
   - Cross-reference sibling concepts extracted from the same file and pre-existing concepts in the OKF engine using `[[wikilinks]]` (e.g. `[[concepts/customer_orders]]`).

## Execution Protocol (Step-by-Step)

### Step 1: Local File Inspection
Call `inspect_source_of_truth` with `uri: "file://<absolute-path>"`.
Inspect the text snippet, page count, CSV headers, or schema definition.

### Step 2: Content Analysis & Decomposition
Analyze the text content to identify:
- How many atomic domain concepts are in the document?
- What are their semantic titles, descriptions, and concept types?
- What cross-referencing `[[wikilinks]]` connect them?

### Step 3: Knowledge Search & Deduplication
For each identified concept:
- Call `search_knowledge` with the entity title.
- If a concept already exists: Call `update_concept` (using `append_body` or `body`).
- If it is new: Call `create_concept`.

### Step 4: Frontmatter & Content Formatting
Construct valid OKF v0.2 frontmatter:
```yaml
type: <ConceptType>
title: <Semantic Entity Title>
description: <Concise 1-sentence summary>
resource: file://<absolute-path>
tags: [<relevant-tags>]
generated:
  by: process:local-file-ingestor/v1
  at: <ISO-8601-timestamp>
sources:
  - id: parent-local-file
    resource: file://<absolute-path>
    title: <Original File Title>
    author: human:<owner-or-team>
status: draft
stale_after: "YYYY-MM-DD"
```

### Step 5: Ingestion, Audit & Attestation
- Execute `create_concept` or `update_concept`.
- Call `list_broken_links` to audit wikilink integrity and fix any dangling links.
- Call `verify_concept`:
  ```json
  {
    "concept_id": "concepts/<target_id>",
    "actor": "process:local-file-ingestor/v1",
    "notes": "Semantically ingested and verified from local file file://<absolute-path>"
  }
  ```
