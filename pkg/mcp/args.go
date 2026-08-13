package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
)

type SearchKnowledgeArgs struct {
	// Search keyword query
	Query string `json:"query"`
	// Filter by trust tier
	// enum: all, human, machine, unverified
	// default: all
	TrustTier string `json:"trust_tier,omitempty"`

	// Filter by concept type e.g. BigQuery Table / Metric / Playbook
	// default: all
	ConceptType string `json:"concept_type"`
}

type CreateConceptArgs struct {
	// Relative file path without .md e.g. concept/customer_orders
	ConceptID string `json:"concept_id"`

	// Agent produser string e.g. cloude-3-5-sonet/v1
	AgentID string `json:"agent_id"`

	// Morkdown body content
	Body string `json:"body"`
	FrontmatterArgs
}

type ReadConceptArgs struct {
	// Relative file path without .md e.g. concept/customer_orders
	ConceptID string `json:"concept_id"`

	// If true, automatically attaches source inspection data for concept.Frontmatter.Resource
	// default: false
	IncludeSourceTruth bool `json:"include_source_truth,omitempty"`
}

type InspectSourceTruthArgs struct {
	// Canonical asset URI or source URI (e.g. gdrive://, gs://, https://drive.google.com/file/...)
	URI string `json:"uri"`
}

type FrontmatterArgs struct {
	// OKF Concept Type, e.g. Playbook
	Type string `json:"type"`

	Title       string `json:"title"`
	Description string `json:"description,omitempty"`

	// Canonical asset URI (bq://, postgresql://, gdrive://, gs://)
	Resource string `json:"resource,omitempty"`

	// enum: draft, stable, deprecated
	// default: draft
	Status string `json:"status,omitempty"`

	// YYYY-MM-DD date string
	StaleAfter string `json:"stale_after,omitempty"`

	Tags    []string     `json:"tags,omitempty"`
	Sources []okf.Source `json:"sources,omitempty"`
}

type UpdateConceptArgs struct {
	// Relative file path without .md e.g. concept/customer_orders
	ConceptID string `json:"concept_id"`

	// Agent producer string e.g. claude-3-5-sonnet/v1
	AgentID string `json:"agent_id"`

	// Full replacement markdown bdoy content (replaces existing body)
	Body string `json:"body,omitempty"`

	// Markdown text to append to the existing body (ignored if 'body' is supplied)
	AppendBody string `json:"append_body,omitempty"`

	FrontmatterUpdate *FrontmatterArgs `json:"frontmatter_updates,omitempty"`
}

type VerifyConceptArgs struct {
	// Relative file path without .md e.g. concept/customer_orders
	ConceptID string `json:"concept_id"`

	// Actor identifier e.g. process:data-quality-bot or human:username
	Actor string `json:"actor"`

	// Verification notes or audit log entry text
	Notes string `json:"notes"`
}

type GetBacklinksArgs struct {
	// Relative file path without .md e.g. concept/customer_orders
	ConceptID string `json:"concept_id"`
}

type TraverseGraphArgs struct {
	// Starting concept ID e.g. concept/customer_orders
	RootConceptID string `json:"root_concept_id"`

	// Maximum graph traversal depth
	// default: 2
	// max: 4
	MaxDepth int `json:"max_depth"`

	// Traversal direction
	// enum: both, outgoing, incoming
	// default: both
	Direction string `json:"direction"`
}

func DecodeArgsTo[T any](rawBytes json.RawMessage, dest *T) *JSONRPCError {
	if len(rawBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(rawBytes, dest); err != nil {
		return newInvalidParamsError(fmt.Sprintf("invalid arguments structure: %v", err))
	}

	return nil
}

func argsSchema[T any]() *jsonschema.Schema {
	s, err := jsonschema.For[T](&jsonschema.ForOptions{})
	if err != nil {
		panic(fmt.Sprintf("invalid schema definition: %v", err))
	}
	return s
}
