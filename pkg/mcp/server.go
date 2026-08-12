package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
	"gopkg.in/yaml.v3"
)

type MCPServer struct {
	store     *store.MemoryStore
	baseDir   string
	simulator *sourcetruth.Simulator
}

func NewMCPServer(store *store.MemoryStore, baseDir string, simulator *sourcetruth.Simulator) *MCPServer {
	return &MCPServer{
		store:     store,
		baseDir:   baseDir,
		simulator: simulator,
	}
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func (s *MCPServer) HandleRPC(req JSONRPCRequest) *JSONRPCResponse {
	// Notifications (e.g. notifications/initialized) do not expact a response
	if req.ID == nil {
		return nil
	}

	resp := &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "okf-knowledge-engine",
				"version": "0.2.0",
			},
		}

	case "tools/list":
		resp.Result = map[string]any{
			"tools": []map[string]any{
				{
					"name":        "search_knowledge",
					"description": "Search OKF v0.2 knowledge corpus by keyword query with optional tier and type filters",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{
								"type":        "string",
								"description": "Search keyword query",
							},
							"tier": map[string]any{
								"type":        "string",
								"description": "Filter by trust tier: all, human, machine, unverified",
								"enum":        []string{"all", "human", "machine", "unverified"},
								"default":     "all",
							},
							"concept_type": map[string]any{
								"type":        "string",
								"description": "Filter by concept type e.g. BigQuery Table, Matric, Playbook",
								"default":     "all",
							},
						},
						"required": []string{"query"},
					},
				},
				{
					"name":        "read_concept",
					"description": "Read a complete OKF v0.2 concept by Concept ID including metadata, body, links, and optional source of truth inspection",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"concept_id": map[string]any{
								"type":        "string",
								"description": "e.g. concept/customer_oreders",
							},
							"include_source_truth": map[string]any{
								"type":        "boolean",
								"description": "If true, automatically attaches source inspection data for concept.Frontmatter.Resource",
								"default":     false,
							},
						},
						"required": []string{"concept_id"},
					},
				},

				{
					"name":        "create_concept",
					"description": "Create a new OKF v0.2 concept markdown file with YAML frontmatter",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"concept_id": map[string]any{
								"type":        "string",
								"description": "Relative file path without .md, e.g. concept/customer_oreders",
							},
							"type": map[string]any{
								"type":        "string",
								"description": "OKF Concept Type, e.g. Playbook",
							},
							"title": map[string]any{
								"type": "string",
							},
							"description": map[string]any{
								"type": "string",
							},
							"body": map[string]any{
								"type":        "string",
								"description": "Markdown body content",
							},
							"agent_id": map[string]any{
								"type":        "string",
								"description": "Agent producer string, e.g. claude-3-5-sonnet/v1",
							},
							"resource": map[string]any{
								"type":        "string",
								"description": "Caconical asset URI (bq://, postgresql://, gdrive://, gs://)",
							},
							"tags": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
							"status": map[string]any{
								"type":    "string",
								"enum":    []string{"draft", "stable", "deprecated"},
								"default": "draft",
							},
							"stale_after": map[string]any{
								"type":        "string",
								"description": "YYYY-MM-DD date string",
							},
							"sources": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"id":       map[string]any{"type": "string"},
										"resource": map[string]any{"type": "string"},
										"title":    map[string]any{"type": "string"},
										"author":   map[string]any{"type": "string"},
									},
								},
								"required": []string{"id", "resource"},
							},
						},
						"required": []string{"concept_id", "type", "title", "body", "agent_id"},
					},
				},
				{
					"name":        "inspect_source_of_truth",
					"description": "Inspect technical meatadata, document text snippets, spreadsheet columns, and local simulation files for an external source URI",
					"inputSchema": map[string]any{
						"uri": map[string]any{
							"type":        "string",
							"description": "Canonical asset URI or source URI (e.g. gdrive://, gs://, https://drive.google.com/file/...)",
						},
					},
					"required": []string{"uri"},
				},
				{
					"name":        "update_concept",
					"description": "Modify an existing OKF concept's frontmatter fields, replace body, or  append body content",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"concept_id": map[string]any{
								"type":        "string",
								"description": "Concept ID e.g. concept/customer_oreders",
							},
							"agent_id": map[string]any{
								"type":        "string",
								"description": "Actor string identifying the modifying agent  e.g. agent:claude-3-5-sonnet/v1",
							},
							"body": map[string]any{
								"type":        "string",
								"description": "Full replacement markdown body content (replaces existing body)",
							},
							"append_body": map[string]any{
								"type":        "string",
								"description": "Markdown text to append to the existing body (ignored if 'body' is supplied)",
							},
							"frontmatter_updates": map[string]any{
								"type":        "object",
								"description": "Key-value pairs to update in YAML frontmatter",
								"properties": map[string]any{

									"type": map[string]any{
										"type":        "string",
										"description": "OKF Concept Type, e.g. Playbook",
									},
									"title": map[string]any{
										"type": "string",
									},
									"description": map[string]any{
										"type": "string",
									},
									"resource": map[string]any{
										"type":        "string",
										"description": "Caconical asset URI (bq://, postgresql://, gdrive://, gs://)",
									},
									"tags": map[string]any{
										"type":  "array",
										"items": map[string]any{"type": "string"},
									},
									"status": map[string]any{
										"type":    "string",
										"enum":    []string{"draft", "stable", "deprecated"},
										"default": "draft",
									},
									"stale_after": map[string]any{
										"type":        "string",
										"description": "YYYY-MM-DD date string",
									},
									"sources": map[string]any{
										"type": "array",
										"items": map[string]any{
											"type": "object",
											"properties": map[string]any{
												"id":       map[string]any{"type": "string"},
												"resource": map[string]any{"type": "string"},
												"title":    map[string]any{"type": "string"},
												"author":   map[string]any{"type": "string"},
											},
										},
										"required": []string{"id", "resource"},
									},
								},
							},
						},
						"required": []string{"concept_id", "agent_id"},
					},
				},
				{
					"name":        "verify_concept",
					"description": "Submit a machine or human attestation for a concept, prompting its Trust Tier",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"concept_id": map[string]any{
								"type":        "string",
								"description": "Concept ID e.g. concept/revenue_metric",
							},
							"actor": map[string]any{
								"type":        "string",
								"description": "Actor identifier e.g. process:data-quality-bot or human:username",
							},
							"notes": map[string]any{
								"type":        "string",
								"description": "Verification notes or audit log entry text",
							},
						},
						"required": []string{"concept_id", "actor"},
					},
				},
				{
					"name":        "get_backlinks",
					"description": "Retrieve all incoming backlinks pointing to a concept ID",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"concept_id": map[string]any{
								"type":        "string",
								"description": "Concept ID e.g. concept/revenue_metric",
							},
						},
						"required": []string{"concept_id"},
					},
				},
				{
					"name":        "traverse_graph",
					"description": "Traverse the knowledge graph N-hops starting from a root concept ID",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"root_concept_id": map[string]any{
								"type":        "string",
								"description": "Strarting concept Concept ID e.g. concept/revenue_metric",
							},
							"max_depth": map[string]any{
								"type":        "integer",
								"description": "Maxium grap traversal depth (default: 2, max: 4)",
								"default":     2,
							},
							"direction": map[string]any{
								"type":        "string",
								"enum":        []string{"both", "outgoing", "incoming"},
								"default":     "both",
								"description": "Traversal direction: both, outgoing, or incoming",
							},
						},
						"required": []string{"root_concept_id"},
					},
				},

				{
					"name":        "list_broken_links",
					"description": "Scan the corpus for broken/dangling [[wikilinks]] pointing to  missing concepts",
					"inputSchema": map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
			},
		}

	case "tools/call":
		var callParams struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}

		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			resp.Error = map[string]any{
				"code": -32602, "message": "Invalid params",
			}
			break
		}

		switch callParams.Name {
		case "search_knowledge":
			var q, tier, conceptType string
			if val, ok := callParams.Arguments["query"]; ok && val != nil {
				q = fmt.Sprintf("%v", val)
			}

			if val, ok := callParams.Arguments["tier"]; ok && val != nil {
				tier = fmt.Sprintf("%v", val)
			}

			if val, ok := callParams.Arguments["concept_type"]; ok && val != nil {
				conceptType = fmt.Sprintf("%v", val)
			}

			results := s.store.SearchFiltered(q, tier, conceptType)
			resp.Result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": marshalJSON(results)},
				},
			}

		case "read_concept":
			var conceptID string
			if val, ok := callParams.Arguments["concept_id"]; ok && val != nil {
				conceptID = fmt.Sprintf("%v", val)
			}
			concept, ok := s.store.GetConceptView(conceptID)
			if !ok {
				resp.Error = map[string]any{
					"code": -32602, "message": "Concept not found",
				}
				break
			}

			respMap := map[string]any{
				"concept": concept,
			}

			includeSourceTruth := false
			if val, ok := callParams.Arguments["include_source_truth"]; ok && val != nil {
				if b, ok := val.(bool); ok {
					includeSourceTruth = b
				}
			}

			if includeSourceTruth && s.simulator != nil {
				var inspections []*sourcetruth.Inspection
				visitedURIs := make(map[string]bool)

				addInspection := func(uri string) {
					if uri == "" || visitedURIs[uri] {
						return
					}

					visitedURIs[uri] = true
					if insp, err := s.simulator.Inspect(uri); err == nil {
						if len(insp.TextSnippet) > 500 {
							insp.TextSnippet = insp.TextSnippet[:500] + "\n... [Truncated for multi-source]"
						}
						inspections = append(inspections, insp)
					}
				}

				// 1. Inspect Caconical Resource if present
				addInspection(concept.Frontmatter.Resource)

				// 2. Inspect All Upstream Provence Sources
				for _, src := range concept.Frontmatter.Sources {
					addInspection(src.Resource)
				}

				respMap["source_truth_inspections"] = inspections
			}

			resp.Result = map[string]any{
				"concept": []map[string]any{
					{"type": "text", "text": marshalJSON(respMap)},
				},
			}

		case "create_concept":
			conceptID, _ := callParams.Arguments["concept_id"].(string)
			conceptType, _ := callParams.Arguments["type"].(string)
			title, _ := callParams.Arguments["title"].(string)
			description, _ := callParams.Arguments["description"].(string)
			body, _ := callParams.Arguments["body"].(string)
			agentID, _ := callParams.Arguments["agent_id"].(string)

			resource, _ := callParams.Arguments["resource"].(string)
			status, _ := callParams.Arguments["status"].(string)
			if status == "" {
				status = "draft"
			}
			staleAfter, _ := callParams.Arguments["stale_after"].(string)

			var tags []string
			if tagsRow, ok := callParams.Arguments["tags"].([]any); ok {
				for _, t := range tagsRow {
					tags = append(tags, fmt.Sprintf("%v", t))
				}
			}

			var sources []okf.Source
			if sourcesRaw, ok := callParams.Arguments["sources"].([]any); ok {
				for _, sRaw := range sourcesRaw {
					if sMap, ok := sRaw.(map[string]any); ok {
						sources = append(sources, okf.Source{
							ID:       fmt.Sprintf("%v", sMap["id"]),
							Resource: fmt.Sprintf("%v", sMap["resource"]),
							Title:    fmt.Sprintf("%v", sMap["title"]),
							Author:   okf.Actor(fmt.Sprintf("%v", sMap["author"])),
						})
					}
				}
			}

			fm := &okf.Frontmatter{
				Type:        conceptType,
				Title:       title,
				Description: description,
				Resource:    resource,
				Status:      "draft",
				Tags:        tags,
				StaleAfter:  staleAfter,
				Sources:     []okf.Source{},
				Generated: &okf.Generated{
					By: okf.Actor(agentID),
					At: time.Now(),
				},
			}

			fileContent, _ := okf.FormatConcept(fm, body)

			fullPath := filepath.Join(s.baseDir, conceptID+".md")
			_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
			if err := os.WriteFile(fullPath, []byte(fileContent), 0644); err != nil {
				resp.Error = map[string]any{"code": -32603, "message": err.Error()}
				break
			}

			// Reload store
			_ = s.store.LoadAll()

			resp.Result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": fmt.Sprintf("Created concept '%s' successfully.", conceptID)},
				},
			}

		case "inspect_source_of_truth":
			var uri string
			if val, ok := callParams.Arguments["uri"]; ok && val != nil {
				uri = fmt.Sprintf("%v", val)
			}

			inspection, err := s.simulator.Inspect(uri)
			if err != nil {
				resp.Error = map[string]any{"code": -32603, "message": err.Error()}
				break
			}

			resp.Result = map[string]any{
				"inspection": []map[string]any{
					{"type": "text", "text": marshalJSON(inspection)},
				},
			}

		case "update_concept":
			conceptID, _ := callParams.Arguments["concept_id"].(string)
			agentID, _ := callParams.Arguments["agent_id"].(string)

			fullPath := filepath.Join(s.baseDir, conceptID+".md")
			contentBytes, err := os.ReadFile(fullPath)
			if err != nil {
				resp.Error = map[string]any{
					"code":    -32602,
					"message": fmt.Sprintf("Concept file '%s' not found: %v", conceptID, err),
				}
				break
			}

			// Split Frontmatter and Body
			rawStr := string(contentBytes)
			parts := strings.SplitN(rawStr, "---", 3)
			var existingFM okf.Frontmatter
			var existingBody string

			if len(parts) >= 3 {
				_ = yaml.Unmarshal([]byte(parts[1]), &existingFM)
				existingBody = strings.TrimSpace(parts[2])
			} else {
				existingBody = strings.TrimSpace(rawStr)
			}

			if fmUpdates, ok := callParams.Arguments["frontmatter_updates"].(map[string]any); ok {
				if val, ok := fmUpdates["type"].(string); ok && val != "" {
					existingFM.Type = val
				}

				if val, ok := fmUpdates["title"].(string); ok && val != "" {
					existingFM.Title = val
				}

				if val, ok := fmUpdates["description"].(string); ok && val != "" {
					existingFM.Description = val
				}

				if val, ok := fmUpdates["status"].(string); ok && val != "" {
					existingFM.Status = val
				}

				if val, ok := fmUpdates["resource"].(string); ok && val != "" {
					existingFM.Resource = val
				}

				if val, ok := fmUpdates["stale_after"].(string); ok && val != "" {
					existingFM.StaleAfter = val
				}
				if tagsRaw, ok := fmUpdates["tags"].([]any); ok {
					var newTags []string
					for _, t := range tagsRaw {
						newTags = append(newTags, fmt.Sprintf("%v", t))
					}
					existingFM.Tags = newTags
				}
				if sourcesRaw, ok := fmUpdates["sources"].([]any); ok {
					var newSources []okf.Source
					for _, sRaw := range sourcesRaw {
						if sMap, ok := sRaw.(map[string]any); ok {
							src := okf.Source{
								ID:       fmt.Sprintf("%v", sMap["id"]),
								Resource: fmt.Sprintf("%v", sMap["resource"]),
							}
							if t, ok := sMap["title"].(string); ok {
								src.Title = t
							}
							if a, ok := sMap["author"].(string); ok {
								src.Author = okf.Actor(a)
							}
							newSources = append(newSources, src)
						}
					}
					existingFM.Sources = newSources
				}
			}

			// Handle Body: 'body' (replace) va 'append_body' (append)
			if replaceBody, ok := callParams.Arguments["body"].(string); ok && strings.TrimSpace(replaceBody) != "" {
				existingBody = strings.TrimSpace(replaceBody)
			} else if appendBody, ok := callParams.Arguments["append_body"].(string); ok && strings.TrimSpace(appendBody) != "" {
				existingBody = existingBody + "\n\n" + strings.TrimSpace(appendBody)
			}

			updatedContent, _ := okf.FormatConcept(&existingFM, existingBody)

			if err := os.WriteFile(fullPath, []byte(updatedContent), 0644); err != nil {
				resp.Error = map[string]any{"code": -32603, "message": err.Error()}
			}

			_ = okf.AppendLogEntry(s.baseDir, okf.Actor(agentID), "UPDATED", conceptID, "Concept updated via MCP agent tool")

			_ = s.store.LoadAll()

			resp.Result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": fmt.Sprintf("Concept '%s' updated successfully.", conceptID)},
				},
			}

		case "verify_concept":
			conceptID, _ := callParams.Arguments["concept_id"].(string)
			actorStr, _ := callParams.Arguments["actor"].(string)
			notes, _ := callParams.Arguments["notes"].(string)

			if notes == "" {
				notes = "Attestation submitted via MCP agent tool"
			}

			fullPath := filepath.Join(s.baseDir, conceptID+".md")
			contentBytes, err := os.ReadFile(fullPath)
			if err != nil {
				resp.Error = map[string]any{
					"code":    -32602,
					"message": fmt.Sprintf("Concept file '%s' not found: %v", conceptID, err),
				}
				break
			}

			rawStr := string(contentBytes)
			parts := strings.SplitN(rawStr, "---", 3)
			var existingFM okf.Frontmatter
			var existingBody string

			if len(parts) >= 3 {
				_ = yaml.Unmarshal([]byte(parts[1]), &existingFM)
				existingBody = strings.TrimSpace(parts[2])
			} else {
				existingBody = strings.TrimSpace(rawStr)
			}

			actor := okf.Actor(actorStr)
			existingFM.Verified = append(existingFM.Verified, okf.Verification{
				Actor: actor,
				At:    time.Now(),
				Notes: notes,
			})

			updatedContent, _ := okf.FormatConcept(&existingFM, existingBody)

			if err := os.WriteFile(fullPath, []byte(updatedContent), 0644); err != nil {
				resp.Error = map[string]any{"code": -32603, "message": err.Error()}
				break
			}

			_ = okf.AppendLogEntry(s.baseDir, actor, "VERIFIED", conceptID, notes)
			_ = s.store.LoadAll()

			resp.Result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": fmt.Sprintf("Concept '%s' verified by '%s' successfully.", conceptID, actorStr)},
				},
			}

		case "get_backlinks":
			conceptID, _ := callParams.Arguments["concept_id"].(string)
			conceptView, ok := s.store.GetConceptView(conceptID)

			if !ok {
				resp.Error = map[string]any{"code": -32602, "message": fmt.Sprintf("Concept '%s' not found", conceptID)}
				break
			}

			type Backlink struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Type        string `json:"type"`
				Description string `json:"description"`
				TrustTier   string `json:"trust_tier"`
			}

			var backlinks []Backlink
			for _, inc := range conceptView.IncomingLinks {
				title := inc.Frontmatter.Title
				if title == "" {
					title = inc.ID
				}

				backlinks = append(backlinks, Backlink{
					ID:          inc.ID,
					Title:       title,
					Type:        inc.Frontmatter.Type,
					Description: inc.Frontmatter.Description,
					TrustTier:   inc.TrustTier,
				})
			}

			resp.Result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": marshalJSON(map[string]any{
						"concept_id":  conceptID,
						"backlinks":   backlinks,
						"total_count": len(backlinks),
					})},
				},
			}

		case "traverse_graph":
			rootID, _ := callParams.Arguments["root_concept_id"].(string)
			maxDepth := 2
			if d, ok := callParams.Arguments["max_depth"].(float64); ok && d > 0 {
				maxDepth = int(d)
				if maxDepth > 4 {
					maxDepth = 4
				}
			}

			direction := "both"
			if dir, ok := callParams.Arguments["direction"].(string); ok && dir != "" {
				direction = dir
			}

			if _, ok := s.store.GetConceptView(rootID); !ok {
				resp.Error = map[string]any{"code": -32602, "message": fmt.Sprintf("Root concept '%s' not found", rootID)}
				break
			}

			type GraphNode struct {
				ID        string `json:"id"`
				Title     string `json:"title"`
				Type      string `json:"type"`
				TrustTier string `json:"trust_tier"`
				Depth     int    `josn:"depth"`
			}

			type GraphEdge struct {
				Source string `json:"source"`
				Target string `json:"target"`
			}

			nodesMap := make(map[string]GraphNode)
			edgesMap := make(map[string]GraphEdge)

			type queueItem struct {
				id    string
				depth int
			}

			queue := []queueItem{{id: rootID, depth: 0}}

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]

				view, ok := s.store.GetConceptView(curr.id)
				if !ok {
					continue
				}

				title := view.Concept.Frontmatter.Title
				if title == "" {
					title = view.Concept.ID
				}

				if _, exists := nodesMap[curr.id]; !exists {
					nodesMap[curr.id] = GraphNode{
						ID:        view.Concept.ID,
						Title:     title,
						Type:      view.Concept.Frontmatter.Type,
						TrustTier: view.Concept.TrustTier,
						Depth:     curr.depth,
					}
				}

				if curr.depth >= maxDepth {
					continue
				}

				if direction == "both" || direction == "outgoing" {
					for _, out := range view.OutgoingLinks {
						edgeKey := curr.id + "->" + out.ID
						edgesMap[edgeKey] = GraphEdge{Source: curr.id, Target: out.ID}
						if _, visited := nodesMap[out.ID]; !visited {
							queue = append(queue, queueItem{id: out.ID, depth: curr.depth + 1})
						}
					}
				}

				if direction == "both" || direction == "incoming" {
					for _, inc := range view.IncomingLinks {
						edgeKey := inc.ID + "->" + curr.id
						edgesMap[edgeKey] = GraphEdge{Source: inc.ID, Target: curr.id}
						if _, visited := nodesMap[inc.ID]; !visited {
							queue = append(queue, queueItem{id: inc.ID, depth: curr.depth + 1})
						}
					}
				}
			}

			var nodeList []GraphNode
			for _, n := range nodesMap {
				nodeList = append(nodeList, n)
			}

			var edgeList []GraphEdge
			for _, e := range edgesMap {
				edgeList = append(edgeList, e)
			}

			resp.Result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": marshalJSON(map[string]any{
						"root_concept_id": rootID,
						"max_depth":       maxDepth,
						"direction":       direction,
						"nodes":           nodeList,
						"edges":           edgeList,
					})},
				},
			}

		case "list_broken_links":
			type BrokenLink struct {
				SourceConceptID string `json:"source_concept_id"`
				SourceTitle     string `json:"source_title"`
				TargetWikilink  string `json:"target_wikilink"`
			}

			var brokenLinks []BrokenLink
			allConcepts := s.store.List()
			wikilinkRegex := regexp.MustCompile(`\[\[(.*?)\]\]`)

			for _, concept := range allConcepts {
				matches := wikilinkRegex.FindAllStringSubmatch(concept.BodyMarkdown, -1)
				visitedTargets := make(map[string]bool)

				for _, match := range matches {
					if len(match) > 1 {
						targetID := match[1]
						if visitedTargets[targetID] {
							continue
						}
						visitedTargets[targetID] = true

						if _, exists := s.store.Get(targetID); !exists {
							title := concept.Frontmatter.Title
							if title == "" {
								title = concept.ID
							}
							brokenLinks = append(brokenLinks, BrokenLink{
								SourceConceptID: concept.ID,
								SourceTitle:     title,
								TargetWikilink:  targetID,
							})
						}
					}
				}
			}

			resp.Result = map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": marshalJSON(map[string]any{
						"broken_links": brokenLinks,
						"total_count":  len(brokenLinks),
					})},
				},
			}
		default:
			resp.Error = map[string]any{"code": -32601, "message": fmt.Sprintf("Tool '%s' not found", callParams.Name)}
		}
	default:
		resp.Error = map[string]any{"code": -32601, "message": fmt.Sprintf("Method '%s' not supported", req.Method)}
	}
	return resp
}

func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := s.HandleRPC(req)
	if resp != nil {
		json.NewEncoder(w).Encode(resp)
	}
}

func marshalJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}
