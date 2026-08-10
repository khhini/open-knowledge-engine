# OKF v0.2 Specification: Source of Truth, Resource URIs & Provenance Lineage

This specification defines the standards, data models, URI schemes, and rendering requirements for implementing **Source of Truth & Upstream Provenance Lineage** in the Open Knowledge Engine.

---

## 1. Executive Summary

In OKF v0.2, Markdown concepts serve as semantic metadata wrappers over underlying technical assets (BigQuery tables, PostgreSQL databases, cloud storage buckets, microservices, and dbt models). To bridge human readability with machine rigor, concepts must explicitly declare:
1. **Canonical Resource (`resource`)**: The exact URI / location of the target physical or logical asset.
2. **Provenance Sources (`sources`)**: An array of upstream data dependencies, schemas, and origins.
3. **Attestation (`attestation`)**: Execution metadata proving deterministic computation or dynamic verification.

---

## 2. Frontmatter Metadata Standards

### 2.1 Frontmatter Schema Definition

```yaml
---
# 1. Identity & Canonical Target
type: BigQuery Table                          # REQUIRED: Concept category
title: Customer Orders Table                   # Optional: Human readable title
resource: https://console.cloud.google.com/... # Canonical asset URI
tags: [sales, revenue, analytics]

# 2. Provenance Sources (Upstream Data Lineage)
sources:
  - id: ga4-export                            # Unique source identifier
    resource: https://developers.google.com/...# Upstream source URI
    title: GA4 BigQuery Export Schema         # Human title
    author: team:ga4-docs                      # Actor format: human:*, process:*, team:*
    usage_count: 5200                          # Operational usage metric
    last_modified: 2026-06-15                  # ISO-8601 date string (YYYY-MM-DD)

# 3. Computational Attestation
attestation:
  runtime: dbt-core/v1.7                      # Execution engine & version
  executor: process:airflow-dag-daily-orders  # Actor executing computation
  computation: sql/marts/fct_orders.sql      # Script or query identifier
  parameters:                                 # Key-value runtime arguments
    cluster: prod-us-central1
    partition_date: "2026-08-01"

status: stable
---
```

---

## 3. Standardized Resource URI Schemes

To ensure machine parsability and visual categorization in the UI, canonical `resource` and provenance `sources[].resource` strings must conform to the following URI standards:

| Enterprise Asset Type | URI Scheme Standard | Example | UI Icon / Badge |
| :--- | :--- | :--- | :--- |
| **BigQuery Table** | `https://console.cloud.google.com/bigquery?...` or `bq://<project>.<dataset>.<table>` | `bq://enterprise-prod.sales.customer_orders` | 🟢 BigQuery (Cyan) |
| **PostgreSQL Table** | `postgresql://<host>:<port>/<db>/<schema>/<table>` | `postgresql://db.prod:5432/analytics/public/orders` | 🔵 Postgres (Blue) |
| **Cloud Storage (GCS/S3)**| `gs://<bucket>/<path>` or `s3://<bucket>/<path>` | `gs://data-lake-prod/raw/orders/v1/` | 🟡 Cloud Storage (Amber) |
| **dbt Model** | `dbt://<repository>/<path_to_model>.sql` | `dbt://analytics-repo/models/marts/fct_orders.sql` | 🟠 dbt Model (Orange) |
| **REST / OpenAPI** | `https://<domain>/<api_path>` | `https://api.enterprise.com/v1/orders` | 🟣 API Endpoint (Purple) |
| **Airflow / Dagster** | `airflow://<dag_id>/<task_id>` | `airflow://daily_etl_pipeline/extract_orders` | 🟢 Workflow DAG (Emerald) |

---

## 4. Go Data Model Implementation

The Go data types in [`pkg/okf/spec.go`](file:///Users/kikihutapea/Documents/khhini-workspaces/open-knowledge-engine/pkg/okf/spec.go) map directly to these structures:

```go
type Source struct {
	ID           string `yaml:"id,omitempty" json:"id,omitempty"`
	Resource     string `yaml:"resource" json:"resource"` // Required field
	Title        string `yaml:"title,omitempty" json:"title,omitempty"`
	Author       Actor  `yaml:"author,omitempty" json:"author,omitempty"`
	UsageCount   int    `yaml:"usage_count,omitempty" json:"usage_count,omitempty"`
	LastModified string `yaml:"last_modified,omitempty" json:"last_modified,omitempty"`
}

type Attestation struct {
	Runtime     string         `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Executor    Actor          `yaml:"executor,omitempty" json:"executor,omitempty"`
	Computation string         `yaml:"computation,omitempty" json:"computation,omitempty"`
	Parameters  map[string]any `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

type Frontmatter struct {
	Type        string       `yaml:"type" json:"type"`
	Title       string       `yaml:"title,omitempty" json:"title,omitempty"`
	Description string       `yaml:"description,omitempty" json:"description,omitempty"`
	Resource    string       `yaml:"resource,omitempty" json:"resource,omitempty"`
	Tags        []string     `yaml:"tags,omitempty" json:"tags,omitempty"`
	Sources     []Source     `yaml:"sources,omitempty" json:"sources,omitempty"`
	Attestation *Attestation `yaml:"attestation,omitempty" json:"attestation,omitempty"`
	// ... lifecycle & verification fields
}
```

---

## 5. UI Rendering Specification (`templates/fragments/concept.html`)

The concept viewer workspace must visually highlight **Source of Truth** and **Provenance Lineage** through distinct UI components:

### 5.1 Canonical Resource Banner
Positioned directly under the concept header:
- Display a clickable badge leading to `.Frontmatter.Resource`.
- Automatically format URI badges based on scheme (e.g. `BigQuery`, `PostgreSQL`, `Cloud Storage`, `Web/API`).

```html
{{ if .Frontmatter.Resource }}
<div class="bg-slate-900/80 border border-slate-800 rounded-lg p-3 flex items-center justify-between my-4">
  <div class="flex items-center gap-2">
    <span class="text-xs px-2 py-0.5 rounded bg-cyan-950 text-cyan-400 border border-cyan-800 font-mono">
      Source of Truth
    </span>
    <span class="text-xs text-slate-300 font-mono truncate max-w-md">{{ .Frontmatter.Resource }}</span>
  </div>
  <a href="{{ .Frontmatter.Resource }}" target="_blank" rel="noopener noreferrer"
    class="text-xs px-3 py-1 rounded bg-cyan-600 hover:bg-cyan-500 text-white font-medium transition-colors flex items-center gap-1">
    Open Asset ↗
  </a>
</div>
{{ end }}
```

### 5.2 Upstream Lineage Cards Section
Positioned in a dedicated section alongside Backlinks & Links:
- Render each item in `.Frontmatter.Sources`.
- Display source title, URI link, author badge, last modified date, and usage count.

```html
{{ if .Frontmatter.Sources }}
<div class="mt-6 bg-slate-900/40 border border-slate-800 rounded-xl p-4">
  <h3 class="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">Upstream Lineage Sources</h3>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
    {{ range .Frontmatter.Sources }}
    <div class="p-3 rounded-lg bg-slate-950 border border-slate-800 text-xs">
      <div class="flex items-center justify-between mb-1">
        <span class="font-bold text-slate-200">{{ if .Title }}{{ .Title }}{{ else }}{{ .ID }}{{ end }}</span>
        <span class="text-slate-500 font-mono">{{ .Author }}</span>
      </div>
      <a href="{{ .Resource }}" target="_blank" class="text-cyan-400 hover:underline block font-mono truncate mb-2">
        {{ .Resource }}
      </a>
      <div class="flex items-center gap-3 text-slate-400 text-[11px]">
        {{ if .LastModified }}<span>Updated: {{ .LastModified }}</span>{{ end }}
        {{ if .UsageCount }}<span>Usage: {{ .UsageCount }} refs</span>{{ end }}
      </div>
    </div>
    {{ end }}
  </div>
</div>
{{ end }}
```

### 5.3 Computation Attestation Panel
When `.Frontmatter.Attestation` is present:
- Display executor actor, runtime version, and parameters in a code block component.

---

## 6. MCP Agent Integration

When AI agents query concepts via Model Context Protocol (`read_concept` or `search_knowledge` in [`pkg/mcp/server.go`](file:///Users/kikihutapea/Documents/khhini-workspaces/open-knowledge-engine/pkg/mcp/server.go)):
- The JSON response automatically includes `.frontmatter.resource`, `.frontmatter.sources`, and `.frontmatter.attestation`.
- AI agents use these fields to verify table schemas, trace data upstream origins, and generate accurate SQL queries referencing exact database connections.

---

## 7. Implementation Checklist

- [ ] Define frontmatter data structures in [`pkg/okf/spec.go`](file:///Users/kikihutapea/Documents/khhini-workspaces/open-knowledge-engine/pkg/okf/spec.go).
- [ ] Create example Markdown concepts with `resource` & `sources` in `knowledge/concepts/`.
- [ ] Update [`templates/fragments/concept.html`](file:///Users/kikihutapea/Documents/khhini-workspaces/open-knowledge-engine/templates/fragments/concept.html) to render canonical resource banner and provenance lineage cards.
- [ ] Add URI scheme parsing helper to display customized icons (BigQuery, PostgreSQL, GCS, API).
