package okf

import (
	"html/template"
	"time"
)

type Actor string

type Source struct {
	ID           string `yaml:"id,omitempty" json:"id,omitempty"`
	Resource     string `yaml:"resource" json:"resource"` // Required
	Title        string `yaml:"title,omitempty" json:"title,omitempty"`
	Author       Actor  `yaml:"author,omitempty" json:"author,omitempty"`
	UsageCount   int    `yaml:"usage_count,omitempty" json:"usage_count,omitempty"`
	LastModified string `yaml:"last_modified,omitempty" json:"last_modified,omitempty"` // YYYY-MM-DD
}

type Verification struct {
	Actor Actor     `yaml:"actor" json:"actor"`
	At    time.Time `yaml:"at" json:"at"`
	Notes string    `yaml:"notes,omitempty" json:"notes,omitempty"`
}

type Attestation struct {
	Runtime     string         `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Executor    Actor          `yaml:"executor,omitempty" json:"executor,omitempty"`
	Computation string         `yaml:"computation,omitempty" json:"computation,omitempty"`
	Parameters  map[string]any `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

type UsageWindow struct {
	From string `yaml:"from,omitempty" json:"from,omitempty"`
	To   string `yaml:"to,omitempty" json:"to,omitempty"`
}

type Generated struct {
	By Actor     `yaml:"by" json:"by"`
	At time.Time `yaml:"at" json:"at"`
}

// Frontmatter represents OKF v0.2 specification metadata
type Frontmatter struct {
	// Required
	Type string `yaml:"type" json:"type"` // REQUIRED

	// Identity
	Title       string   `yaml:"title,omitempty" json:"title,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Resource    string   `yaml:"resource,omitempty" json:"resource,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Province
	Sources     []Source     `yaml:"sources,omitempty" json:"sources,omitempty"`
	UsageWindow *UsageWindow `yaml:"usage_window,omitempty" json:"usage_window,omitempty"`

	// Trust
	Generated *Generated     `yaml:"generated,omitempty" json:"generated,omitempty"`
	Verified  []Verification `yaml:"verified,omitempty" json:"verified,omitempty"`

	// Lifecycle
	Status     string `yaml:"status,omitempty" json:"status,omitempty"` // draft | stable | deprecated
	StaleAfter string `yaml:"stale_after,omitempty" json:"stale_after,omitempty"`

	// Attestation
	Attestation *Attestation `yaml:"attestation,omitempty" json:"attestation,omitempty"`

	// Preserve arbitary unknown YAML extension fileds
	Extra map[string]any `yaml:",inline" json:"extra,omitempty"`
}

type Concept struct {
	ID           string        `json:"id"`        // Relative path without .md e.g. "concepts/orders"
	FilePath     string        `json:"file_path"` // Absolute or bundle-relative
	Frontmatter  Frontmatter   `json:"frontmatter"`
	BodyMarkdown string        `json:"body_markdown"`
	BodyHTML     template.HTML `json:"body_html"`
	TrustTier    string        `json:"trust_tier"`
	IsStale      bool          `json:"is_stale"`
}
