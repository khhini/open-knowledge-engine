# OKF v0.2 Specification: Model Context Protocol (MCP) Expansion

**Document Version:** 2.0.0 (Updated - Full Toolset & Existing Tool Enhancements)  
**Target Platform:** Open Knowledge Engine (`okf` v0.2)  
**Standard Compliance:** Model Context Protocol (MCP) Specification (2024-11-05)  

---

## 1. Executive Summary & Purpose

To enable autonomous AI agents (such as Claude Code, Cursor, Windsurf, or custom subagents) to perform read, write, audit, link-repair, and N-hop reasoning workflows on the knowledge base, OKF expands its native **Model Context Protocol (MCP)** server.

This specification details:
1. **Enhancements to Existing MCP Tools**: Upgrading `search_knowledge`, `create_concept`, `read_concept`, and `inspect_source_of_truth` with filter support, complete frontmatter fields, and automatic source resolution.
2. **5 New MCP Tools**: `update_concept`, `verify_concept`, `traverse_graph`, `get_backlinks`, `list_broken_links`.
3. **HTTP Server-Sent Events (SSE) Transport**: A remote HTTP SSE transport (`/mcp/sse`) enabling concurrent remote AI agent connections alongside standard `stdio`.

---

## 2. Enhancements to Existing MCP Tools

### 2.1 `search_knowledge` (Enhanced)
Adds optional filtering by **Trust Tier** (`tier`) and **Concept Type** (`concept_type`) leveraging `store.SearchFiltered()`.

#### JSON-RPC Input Schema:
```json
{
  "name": "search_knowledge",
  "description": "Search OKF v0.2 knowledge corpus by keyword query with optional tier and type filters",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string", "description": "Search keyword query" },
      "tier": { "type": "string", "enum": ["all", "human", "machine", "unverified"], "default": "all" },
      "concept_type": { "type": "string", "description": "Filter by concept type e.g. BigQuery Table, Metric, Playbook", "default": "all" }
    },
    "required": ["query"]
  }
}
```

---

### 2.2 `create_concept` (Enhanced)
Allows agents to specify `resource`, `sources`, `tags`, `stale_after`, and `status` directly at creation time, eliminating the need for immediate follow-up updates.

#### JSON-RPC Input Schema:
```json
{
  "name": "create_concept",
  "description": "Create a new OKF v0.2 concept markdown file with complete YAML frontmatter",
  "inputSchema": {
    "type": "object",
    "properties": {
      "concept_id": { "type": "string", "description": "Relative file path without .md, e.g. concepts/customer_orders" },
      "type": { "type": "string", "description": "OKF Concept Type e.g. BigQuery Table, Metric, Playbook" },
      "title": { "type": "string" },
      "description": { "type": "string" },
      "body": { "type": "string", "description": "Markdown body content" },
      "agent_id": { "type": "string", "description": "Agent producer string, e.g. process:claude-3-5-sonnet/v1" },
      "resource": { "type": "string", "description": "Canonical asset URI (bq://, postgresql://, gdrive://, gs://)" },
      "tags": { "type": "array", "items": { "type": "string" } },
      "status": { "type": "string", "enum": ["draft", "stable", "deprecated"], "default": "draft" },
      "stale_after": { "type": "string", "description": "YYYY-MM-DD date string" },
      "sources": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "id": { "type": "string" },
            "resource": { "type": "string" },
            "title": { "type": "string" },
            "author": { "type": "string" }
          },
          "required": ["id", "resource"]
        }
      }
    },
    "required": ["concept_id", "type", "title", "body", "agent_id"]
  }
}
```

---

### 2.3 `read_concept` (Enhanced)
Adds optional `include_source_truth` flag to automatically attach simulated source inspection metadata (`sourcetruth.Inspection`) alongside concept details.

#### JSON-RPC Input Schema:
```json
{
  "name": "read_concept",
  "description": "Read a complete OKF v0.2 concept by Concept ID including metadata, body, links, and optional source of truth inspection",
  "inputSchema": {
    "type": "object",
    "properties": {
      "concept_id": { "type": "string", "description": "e.g. concepts/customer_orders" },
      "include_source_truth": { "type": "boolean", "description": "If true, automatically attaches source inspection data for concept.Frontmatter.Resource", "default": false }
    },
    "required": ["concept_id"]
  }
}
```

---

### 2.4 `inspect_source_of_truth` (Enhanced)
Supports passing `concept_id` directly, automatically looking up the concept's canonical `resource` or primary `source` URI if `uri` is omitted.

#### JSON-RPC Input Schema:
```json
{
  "name": "inspect_source_of_truth",
  "description": "Inspect technical metadata, document text snippets, spreadsheet columns, and local simulation files for an external source URI or Concept ID",
  "inputSchema": {
    "type": "object",
    "properties": {
      "uri": { "type": "string", "description": "Canonical asset URI or source URI (e.g. gdrive://, gs://, https://drive.google.com/file/...)" },
      "concept_id": { "type": "string", "description": "Concept ID to inspect canonical resource from (e.g. concepts/customer_orders)" }
    }
  }
}
```

---

## 3. New MCP Toolset Specifications

### 3.1 `update_concept`
Allows an AI agent to modify existing concept markdown sections, update YAML frontmatter fields across all 5 OKF v0.2 metadata families (`type`, `title`, `description`, `status`, `tags`, `stale_after`, `resource`, `sources`, `usage_window`, `attestation`), or append body text.

#### JSON-RPC Input Schema:
```json
{
  "name": "update_concept",
  "description": "Modify an existing OKF concept's frontmatter fields or append body content",
  "inputSchema": {
    "type": "object",
    "properties": {
      "concept_id": { "type": "string", "description": "Concept ID (e.g. concepts/customer_orders)" },
      "agent_id": { "type": "string", "description": "Actor string identifying the modifying agent (e.g. process:doc-updater/v1)" },
      "append_body": { "type": "string", "description": "Markdown text to append to the concept body" },
      "frontmatter_updates": {
        "type": "object",
        "description": "Key-value pairs to update in YAML frontmatter",
        "properties": {
          "type": { "type": "string" },
          "title": { "type": "string" },
          "description": { "type": "string" },
          "status": { "type": "string", "enum": ["draft", "stable", "deprecated"] },
          "tags": { "type": "array", "items": { "type": "string" } },
          "stale_after": { "type": "string", "description": "YYYY-MM-DD date string" },
          "resource": { "type": "string", "description": "Canonical asset URI (bq://, postgresql://, gdrive://, gs://)" },
          "sources": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "id": { "type": "string" },
                "resource": { "type": "string" },
                "title": { "type": "string" },
                "author": { "type": "string" },
                "usage_count": { "type": "integer" },
                "last_modified": { "type": "string" }
              },
              "required": ["id", "resource"]
            }
          },
          "usage_window": {
            "type": "object",
            "properties": { "from": { "type": "string" }, "to": { "type": "string" } }
          },
          "attestation": {
            "type": "object",
            "properties": {
              "runtime": { "type": "string" },
              "executor": { "type": "string" },
              "computation": { "type": "string" },
              "parameters": { "type": "object" }
            }
          }
        }
      }
    },
    "required": ["concept_id", "agent_id"]
  }
}
```

---

### 3.2 `verify_concept`
Enables automated verification agents (CI/CD bots, data quality verifiers) to append machine attestations to a concept, promoting its Trust Tier.

#### JSON-RPC Input Schema:
```json
{
  "name": "verify_concept",
  "description": "Submit a machine attestation for a concept, promoting its Trust Tier",
  "inputSchema": {
    "type": "object",
    "properties": {
      "concept_id": { "type": "string", "description": "Concept ID (e.g. concepts/revenue_metric)" },
      "actor": { "type": "string", "description": "Actor identifier (e.g. process:data-quality-bot or human:username)" },
      "notes": { "type": "string", "description": "Verification notes or audit log entry text" }
    },
    "required": ["concept_id", "actor"]
  }
}
```

---

### 3.3 `traverse_graph`
Exposes N-hop graph traversal so agents can explore incoming and outgoing concept dependencies up to `max_depth`.

#### JSON-RPC Input Schema:
```json
{
  "name": "traverse_graph",
  "description": "Traverse the knowledge graph N-hops starting from a root concept ID",
  "inputSchema": {
    "type": "object",
    "properties": {
      "root_concept_id": { "type": "string", "description": "Starting concept ID (e.g. concepts/revenue_metric)" },
      "max_depth": { "type": "integer", "description": "Maximum graph traversal depth (default: 2, max: 4)", "default": 2 },
      "direction": { "type": "string", "enum": ["both", "outgoing", "incoming"], "default": "both" }
    },
    "required": ["root_concept_id"]
  }
}
```

---

### 3.4 `get_backlinks`
Returns all concepts referencing the target concept ID via `[[wikilinks]]`.

#### JSON-RPC Input Schema:
```json
{
  "name": "get_backlinks",
  "description": "Retrieve all incoming backlinks pointing to a concept ID",
  "inputSchema": {
    "type": "object",
    "properties": {
      "concept_id": { "type": "string", "description": "Target concept ID (e.g. concepts/customer_profiles)" }
    },
    "required": ["concept_id"]
  }
}
```

---

### 3.5 `list_broken_links`
Scans the entire knowledge corpus for dangling `[[wikilinks]]` that point to non-existent concepts.

#### JSON-RPC Input Schema:
```json
{
  "name": "list_broken_links",
  "description": "Scans corpus for broken/dangling [[wikilinks]] pointing to missing concepts",
  "inputSchema": { "type": "object", "properties": {} }
}
```

---

## 4. HTTP Server-Sent Events (SSE) Transport Specification

Adding `/mcp/sse` and `/mcp/message` endpoints to support remote HTTP streaming connections for remote AI agents.
