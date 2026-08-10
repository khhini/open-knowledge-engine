package sourcetruth

import "time"

type AssetType string

const (
	TypeGDrive      AssetType = "Google Drive Document"
	TypeSpreadsheet AssetType = "Google Spreadsheet / CSV"
	TypeGCS         AssetType = "Google Cloud Storage"
	TypeS3          AssetType = "AWS S3 Object"
	TypeBigQuery    AssetType = "BigQuery Table"
	TypePostgres    AssetType = "Postgres Table"
	TypeAPI         AssetType = "API Endpoint"
	TypteDBT        AssetType = "dbt Model"
	TypeLocalFile   AssetType = "Local File"
	TypeUnknown     AssetType = "Unknown Asset"
)

// Column represents a structured field in a spreadsheet or database schema
type Column struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Nullable    string `json:"nullable,omitempty"`
	Description string `json:"description,omitempty"`
}

// Inspection contains technical metadata and previews retrieved from the simulator
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

type SourceStatus struct {
	SourceID     string    `json:"source_id"`
	URI          string    `json:"uri"`
	AssetType    AssetType `json:"asset_type"`
	Exists       bool      `json:"exists"`
	LocalPath    string    `json:"local_path"`
	LastModified string    `json:"last_modified,omitempty"`
}
