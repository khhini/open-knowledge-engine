# Enterprise Open Knowledge Engine (OKF v0.2)

An open-spec, high-performance knowledge management and human-AI collaboration platform implementing **Open Knowledge Format (OKF) v0.2**.

Built with a text-native, zero-database architecture in **Go**, **htmx**, **UnoCSS**, and **Goldmark**, this platform bridges human readability (Markdown + YAML frontmatter) with machine rigor (provenance, attestation, and dynamic trust signals).

---

## Key Capabilities

### 1. OKF v0.2 Specification Compliance
- **Text-Native Storage**: Conforms to the official [Google Cloud Platform OKF v0.2 Specification](https://github.com/GoogleCloudPlatform/knowledge-catalog).
- **5 Metadata Families**:
  1. **Identity**: `type` (required), `title`, `description`, `resource`, `tags`.
  2. **Provenance**: `sources` list (`id`, `resource`, `title`, `author`, `usage_count`, `last_modified`) and `usage_window`.
  3. **Trust & Verification**: `generated` (`by`, `at`) and `verified` (`[{ actor, at, notes }]`).
  4. **Lifecycle**: `status` (`draft`, `stable`, `deprecated`) and `stale_after`.
  5. **Attestation**: `attestation` block (`runtime`, `executor`, `computation`, `parameters`) for deterministic computations.
- **Reserved Documents (§3.1)**: Automated generation of `index.md` (directory table for agent discovery) and `log.md` (chronological audit log).

### 2. Dynamic Trust Tier Engine
Computes credibility at read-time instead of storing static scores:
- 🟢 **Human-Reviewed**: Verified by a human actor (`human:username`).
- 🔵 **Machine-Confirmed**: Verified by automated agents/pipelines (`process:id` or `producer/version`).
- 🟡 **Stale**: Current timestamp exceeds `stale_after` expiration threshold.
- ⚪ **Unverified**: Default state for unreviewed concepts.

### 3. Bidirectional Knowledge Graph & Backlinks
- Scans `[[wikilinks]]` in concept bodies to build bidirectional graph edges in memory.
- Displays **Referenced By (Backlinks)** and **Links To** cards for instant graph traversal.

### 4. Model Context Protocol (MCP) Server for LLM Agents
- Native JSON-RPC 2.0 MCP handler in Go (`pkg/mcp/server.go`) supporting `initialize`, `notifications/initialized`, `tools/list`, and `tools/call`.
- Exposes tools (`search_knowledge`, `read_concept`, `create_concept`) to LLM agents (Claude Code, Cursor, Windsurf, or custom subagents).
- Includes panic recovery and type-safe argument extraction.

### 5. Interactive 2D Knowledge Graph Visualizer
- Force-directed graph canvas modal (`force-graph` + UnoCSS).
- Color-coded nodes by **Trust Tier**.
- Smoothly centers, zooms, and highlights the currently active concept with a glowing cyan ring.

### 6. Real-Time Hot-Reload File Watcher (`fsnotify`)
- Recursively monitors the `knowledge/` directory for `.md` file edits from Git (`git pull`), Obsidian, text editors, or AI agents.
- Automatically re-indexes the memory store and updates `index.md` in real-time without server restarts.

### 7. Dynamic Sidebar Filters
- Real-time filtering by keyword query, **Trust Tier** (Human, Machine, Unverified), and **Concept Type** (`BigQuery Table`, `PostgreSQL Table`, `Metric`, `Policy`, `API Endpoint`, `Playbook`).

---

## Directory Architecture

```text
.
├── cmd/
│   └── server/
│       └── main.go              # Server routing, flag parsing, and HTTP handlers
├── pkg/
│   ├── okf/
│   │   ├── spec.go              # OKF v0.2 structs & actor conventions
│   │   ├── trust.go             # Dynamic trust tier calculation algorithm
│   │   └── generator.go         # Reserved index.md and log.md auto-generators
│   ├── store/
│   │   └── memory.go            # Thread-safe in-memory graph store & search
│   ├── mcp/
│   │   └── server.go            # Panic-safe MCP JSON-RPC 2.0 server
│   └── watcher/
│       └── watcher.go           # Real-time fsnotify file watcher with debouncing
├── templates/
│   ├── index.html               # Main htmx layout with UnoCSS & Force Graph modal
│   └── fragments/
│       ├── list.html            # Sidebar concept list with Trust Tier & Type filters
│       ├── concept.html         # Concept workspace with backlinks & UnoCSS typography
│       └── trust_badge.html     # Interactive human verification fragment
└── knowledge/                   # OKF v0.2 Markdown Corpus
    ├── index.md                 # Auto-generated bundle index
    ├── log.md                   # Auto-generated audit log
    └── concepts/
        ├── customer_orders.md
        ├── customer_profiles.md
        ├── freshness_playbook.md
        ├── revenue_metric.md
        ├── subscription_plans.md
        ├── fraud_detector.md
        └── data_retention_policy.md
```

---

## Getting Started

### Prerequisites
- **Go**: 1.22+ installed.

### Installation

```bash
# Clone the repository
git clone https://github.com/khhini/enterprise-open-knowledge-graph.git
cd enterprise-open-knowledge-graph

# Download Go dependencies
go get github.com/yuin/goldmark
go get gopkg.in/yaml.v3
go get github.com/fsnotify/fsnotify
```

### Running the Web Server

```bash
go run cmd/server/main.go
```

Open your browser at `http://localhost:8080` to access the interactive platform.

### Connecting to LLM AI Agents (MCP Mode)

To run the MCP server over standard input/output (`stdio`) for CLI agents like `agy` or Claude Desktop:

```bash
go run cmd/server/main.go -mcp-stdio
```

#### MCP Plugin Config (`mcp_config.json`)

```json
{
  "mcpServers": {
    "okf-knowledge-engine": {
      "command": "go",
      "args": ["run", "cmd/server/main.go", "-mcp-stdio"],
      "cwd": "/path/to/enterprise-open-knowledge-graph"
    }
  }
}
```

---

## License

Open source under the [Apache 2.0 License](LICENSE).
