# OKF v0.2 Specification: Source of Truth & External File Storage Simulator

**Document Version:** 2.0.0 (Revised - File Storage Focus)  
**Status:** Approved Specification  
**Target Platform:** Open Knowledge Engine (`okf` v0.2)  

---

## 1. Executive Summary & Purpose

In **OKF v0.2**, Markdown concepts act as human-readable and machine-actionable semantic wrappers over external data assets and documentation sources.

This specification defines the initial implementation phase of OKF's **Source of Truth & Upstream Provenance Lineage**, prioritizing **External Document & File Storage Simulation** (Google Drive files, Google Cloud Storage objects, AWS S3 buckets, and local file shares containing PDFs, Google Docs, Spreadsheets, CSVs, and technical specs).

Instead of requiring active OAuth tokens or cloud storage API keys during local development and testing, OKF uses a **Local File Store Simulator**. The simulator maps external file URIs (`gdrive://`, `gs://`, `s3://`, `file://`, and Google Drive web links) to local mock files inside `.source_of_truth/`.

This enables:
1. **Document Content Previews**: Displaying document summaries, page counts, text extractions, and owner metadata in the Web UI.
2. **Spreadsheet & CSV Inspection**: Viewing column headers, sample data rows, and record counts.
3. **Upstream Dependency Verification**: Validating that referenced PDFs, spec documents, and data exports actually exist.
4. **Agent Context Expansion via MCP**: Enabling LLM agents to read document text snippets and spreadsheet headers before generating answers or code.

---

## 2. Simulated Storage Directory Layout

The simulator organizes mock external files under `.source_of_truth/` in the repository root:

```
.source_of_truth/
├── gdrive/
│   ├── finance/
│   │   ├── mrr_spec.pdf.meta.json          # Google Drive PDF metadata
│   │   └── mrr_spec.pdf.txt                # Document text extraction preview
│   └── analytics/
│       ├── q3_subscription_billing_report.csv.meta.json  # Google Sheet export meta
│       └── q3_subscription_billing_report.csv            # Spreadsheet sample data
├── gcs/
│   └── data-lake-prod/
│       └── documents/
│           ├── privacy_policy_v2.pdf.meta.json # GCS storage metadata
│           └── privacy_policy_v2.txt           # Extracted policy text
├── s3/
│   └── enterprise-data/
│       └── exports/
│           ├── customer_churn_q2.csv.meta.json
│           └── customer_churn_q2.csv
└── file/
    └── local_specs/
        └── data_retention_rules.md
```

---

## 3. Standardized External File URI Mapping Rules

| Storage Provider | URI Scheme Standard | Local Simulator Path |
| :--- | :--- | :--- |
| **Google Drive Document** | `gdrive://<folder>/<file>`<br>*or* `https://drive.google.com/file/d/<id>/view` | `.source_of_truth/gdrive/<folder>/<file>.txt`<br>`.source_of_truth/gdrive/<folder>/<file>.meta.json` |
| **Google Spreadsheet** | `https://docs.google.com/spreadsheets/d/<id>/edit` | `.source_of_truth/gdrive/<folder>/<filename>.csv`<br>`.source_of_truth/gdrive/<folder>/<filename>.csv.meta.json` |
| **Google Cloud Storage (GCS)** | `gs://<bucket>/<path_to_file>` | `.source_of_truth/gcs/<bucket>/<path_to_file>.txt`<br>`.source_of_truth/gcs/<bucket>/<path_to_file>.meta.json` |
| **AWS S3 Object** | `s3://<bucket>/<path_to_file>` | `.source_of_truth/s3/<bucket>/<path_to_file>.txt`<br>`.source_of_truth/s3/<bucket>/<path_to_file>.meta.json` |
| **Local File / Internal Doc** | `file://<relative_path>` | Local path directly in repository |

---

## 4. Local File Simulator Schemas

Every simulated external file consists of two paired local assets:

### 4.1 Metadata File (`.meta.json`)
```json
{
  "asset_type": "Google Drive Document",
  "file_id": "1A2B3C4D5E6F7G8H9I0J",
  "name": "mrr_spec.pdf",
  "mime_type": "application/pdf",
  "size_bytes": 2458112,
  "drive_path": "Drive/Finance/Accounting Specs/mrr_spec.pdf",
  "web_view_link": "https://drive.google.com/file/d/1A2B3C4D5E6F7G8H9I0J/view",
  "owner": "human:finance-team",
  "last_modified": "2026-06-15T10:00:00Z",
  "document_summary": "Finance Revenue Accounting Principles detailing MRR/ARR definitions, pro-rata calculations for subscription changes, and GAAP compliance guidelines.",
  "page_count": 14
}
```

### 4.2 Content Extraction File (`.txt` or `.csv`)
Provides the actual text or tabular content representation:
```
================================================================================
FINANCE REVENUE ACCOUNTING PRINCIPLES (v2026.1)
Document ID: DOC-FIN-2026-MRR
Owner: Finance & Revenue Operations Team (finance-team@enterprise.com)
Last Revised: 2026-06-15
================================================================================

1. EXECUTIVE SUMMARY & MRR DEFINITION
Monthly Recurring Revenue (MRR) represents the normalized monthly subscription fee
contracted by active customer accounts...
```

---

## 5. Go Backend File Inspector Architecture (`pkg/sourcetruth`)

### 5.1 Data Models
```go
package sourcetruth

import "time"

type AssetType string

const (
	TypeGDrive      AssetType = "Google Drive Document"
	TypeSpreadsheet AssetType = "Google Spreadsheet / CSV"
	TypeGCS         AssetType = "Google Cloud Storage"
	TypeS3          AssetType = "AWS S3 Object"
	TypeLocalFile   AssetType = "Local File"
)

type Column struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type Inspection struct {
	URI          string    `json:"uri"`
	AssetType    AssetType `json:"asset_type"`
	LocalPath    string    `json:"local_path"`
	Exists       bool      `json:"exists"`
	FileName     string    `json:"file_name,omitempty"`
	MimeType     string    `json:"mime_type,omitempty"`
	SizeBytes    int64     `json:"size_bytes,omitempty"`
	Owner        string    `json:"owner,omitempty"`
	LastModified time.Time `json:"last_modified,omitempty"`
	Summary      string    `json:"summary,omitempty"`
	PageCount    int       `json:"page_count,omitempty"`
	RowCount     int64     `json:"row_count,omitempty"`
	Columns      []Column  `json:"columns,omitempty"`
	TextSnippet  string    `json:"text_snippet,omitempty"`
	SampleCSV    string    `json:"sample_csv,omitempty"`
	RawContent   string    `json:"raw_content,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}
```

---

## 6. Interactive Document Inspector UI

### 6.1 Concept Page Banner
When a concept contains a `resource` or `sources` link pointing to an external file (e.g. `https://internal.docs/finance/mrr_spec.pdf` or `gs://data-lake-prod/documents/privacy_policy_v2.pdf`), the page displays an interactive **"📄 Inspect Document"** button.

### 6.2 HTMX Document Inspector Drawer (`templates/fragments/source_inspector.html`)
The drawer slides out to reveal:
- **Document Header**: File name, Storage Provider Badge (Google Drive / GCS / S3), Owner, Size, Page/Row count.
- **Document Summary**: AI-generated or author-provided executive summary.
- **Text / Spreadsheet Preview**: Rendered text snippet or formatted CSV table.
- **Direct Link**: Button to open original external URL.

---

## 7. Model Context Protocol (MCP) Integration

The MCP tool `inspect_source_of_truth` enables AI Agents to read document text snippets, page counts, and spreadsheet headers directly during agent reasoning workflows.
