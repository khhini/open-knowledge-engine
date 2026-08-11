package mcp

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
)

func setupTestServer(t *testing.T) (*MCPServer, string) {
	tempDir, err := os.MkdirTemp("", "mcp_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create mock concept file with wikilinks and external resources
	conceptContent := `---
type: BigQuery Table
title: Customer Orders Table
description: Consolidated customer orders
resource: https://internal.docs/finance/mrr_spec.pdf
sources:
  - id: ga4-export
    resource: https://docs.google.com/spreadsheets/d/9Z8Y7X6W5V4U3T2S1R0Q/edit
    title: GA4 Export
    author: team:ga4
---

# Overview
Orders table content with [[concepts/customer_profiles]] and [[concepts/missing_stub]].
`
	conceptsDir := filepath.Join(tempDir, "concepts")
	_ = os.MkdirAll(conceptsDir, 0755)
	_ = os.WriteFile(filepath.Join(conceptsDir, "customer_orders.md"), []byte(conceptContent), 0644)

	memStore := store.NewMemoryStore(tempDir)
	_ = memStore.LoadAll()

	simulator := sourcetruth.NewSimulator(".source_of_truth")
	if _, err := os.Stat("../../.source_of_truth"); err == nil {
		simulator = sourcetruth.NewSimulator("../../.source_of_truth")
	}

	mcpServer := NewMCPServer(memStore, tempDir, simulator)
	return mcpServer, tempDir
}

func TestMCP_ToolsList(t *testing.T) {
	server, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1,
	}

	resp := server.HandleRPC(req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("Expected valid tools/list response, got error: %v", resp)
	}

	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Invalid result map")
	}

	tools, ok := resultMap["tools"].([]map[string]any)
	if !ok || len(tools) == 0 {
		t.Fatalf("Expected tools array, got empty")
	}

	foundSearch := false
	foundUpdate := false
	foundBroken := false
	for _, tool := range tools {
		if tool["name"] == "search_knowledge" {
			foundSearch = true
		}
		if tool["name"] == "update_concept" {
			foundUpdate = true
		}
		if tool["name"] == "list_broken_links" {
			foundBroken = true
		}
	}

	if !foundSearch || !foundUpdate || !foundBroken {
		t.Errorf("Missing expected tools in tools/list response")
	}
}

func TestMCP_UpdateConcept_DualParameter(t *testing.T) {
	server, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	// Test append_body & frontmatter updates
	args := map[string]any{
		"concept_id":  "concepts/customer_orders",
		"agent_id":    "process:test-agent",
		"append_body": "## New Appendix Section\nAdditional notes.",
		"frontmatter_updates": map[string]any{
			"status": "stable",
			"tags":   []any{"sales", "test"},
		},
	}

	argsBytes, _ := json.Marshal(map[string]any{
		"name":      "update_concept",
		"arguments": args,
	})

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      2,
	}

	resp := server.HandleRPC(req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("update_concept failed: %v", resp)
	}

	// Verify file was updated on disk
	updatedContent, _ := os.ReadFile(filepath.Join(tempDir, "concepts/customer_orders.md"))
	if !strings.Contains(string(updatedContent), "status: stable") {
		t.Errorf("Expected status: stable in updated frontmatter")
	}
	if !strings.Contains(string(updatedContent), "## New Appendix Section") {
		t.Errorf("Expected appended body content in updated file")
	}
}

func TestMCP_ListBrokenLinks(t *testing.T) {
	server, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	argsBytes, _ := json.Marshal(map[string]any{
		"name":      "list_broken_links",
		"arguments": map[string]any{},
	})

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      3,
	}

	resp := server.HandleRPC(req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("list_broken_links failed: %v", resp)
	}

	resStr := marshalJSON(resp.Result)
	if !strings.Contains(resStr, "concepts/missing_stub") {
		t.Errorf("Expected missing_stub in broken links output")
	}
}

func TestMCP_SSETransport(t *testing.T) {
	server, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	sseTransport := NewSSETransport(server)

	// Test GET /mcp/sse session creation
	reqSSE := httptest.NewRequest("GET", "/mcp/sse", nil)
	wSSE := httptest.NewRecorder()

	go sseTransport.HandleSSE(wSSE, reqSSE)
	time.Sleep(100 * time.Millisecond)

	bodyStr := wSSE.Body.String()
	if !strings.Contains(bodyStr, "event: endpoint") {
		t.Errorf("Expected endpoint event in SSE stream, got: %s", bodyStr)
	}
}
