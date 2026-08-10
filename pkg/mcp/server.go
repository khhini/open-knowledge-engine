package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	JSONRCP string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func (s *MCPServer) HandleRPC(req JSONRPCRequest) *JSONRPCResponse {
	// Notifications (e.g. notifications/initialized) do not expact a response
	if req.ID == nil {
		return nil
	}

	resp := &JSONRPCResponse{JSONRCP: "2.0", ID: req.ID}

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
					"description": "Search OKF v0.2 knowledge corpus by keyword query",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{
								"type":        "string",
								"description": "Search keyword query",
							},
						},
						"required": []string{"query"},
					},
				},
				{
					"name":        "read_concept",
					"description": "Read a complete OKF v0.2 concept by Concept ID including metadata and body",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"concept_id": map[string]any{
								"type":        "string",
								"description": "e.g. concept/customer_oreders",
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
			var q string
			if val, ok := callParams.Arguments["query"]; ok && val != nil {
				q = fmt.Sprintf("%v", val)
			}
			results := s.store.Search(q)
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
			resp.Result = map[string]any{
				"concept": []map[string]any{
					{"type": "text", "text": marshalJSON(concept)},
				},
			}

		case "create_concept":
			conceptID, _ := callParams.Arguments["concept_id"].(string)
			conceptType, _ := callParams.Arguments["type"].(string)
			title, _ := callParams.Arguments["title"].(string)
			description, _ := callParams.Arguments["description"].(string)
			body, _ := callParams.Arguments["body"].(string)
			agentID, _ := callParams.Arguments["agent_id"].(string)

			fm := okf.Frontmatter{
				Type:        conceptType,
				Title:       title,
				Description: description,
				Status:      "draft",
				Generated: &okf.Generated{
					By: okf.Actor(agentID),
					At: time.Now(),
				},
			}

			yamlBytes, _ := yaml.Marshal(fm)
			fileContent := fmt.Sprintf("---\n%s---\n\n%s", string(yamlBytes), body)

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
