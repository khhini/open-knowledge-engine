package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
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

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newInvalidParamsError(msg string) *JSONRPCError {
	return &JSONRPCError{Code: -32602, Message: msg}
}

func newInternalError(msg string) *JSONRPCError {
	return &JSONRPCError{Code: -32603, Message: msg}
}

func (s *MCPServer) initializeResult() map[string]any {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "okf-knowledge-engine",
			"version": "0.2.0",
		},
	}
}

func (s *MCPServer) HandleRPC(req JSONRPCRequest) *JSONRPCResponse {
	// Notifications (e.g. notifications/initialized) do not expact a response
	if req.ID == nil {
		return nil
	}

	resp := &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = s.initializeResult()
	case "tools/list":
		resp.Result = map[string]any{"tools": s.listTools()}
	case "tools/call":
		var callParams struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}

		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			resp.Error = newInvalidParamsError(err.Error())
			break
		}

		var result any
		var err *JSONRPCError

		switch callParams.Name {
		case "search_knowledge":
			result, err = s.handleSearchKnowledge(callParams.Arguments)
		case "read_concept":
			result, err = s.handleReadConcept(callParams.Arguments)
		case "create_concept":
			result, err = s.handleCreateConcept(callParams.Arguments)
		case "inspect_source_of_truth":
			result, err = s.handleInspectSourceTruth(callParams.Arguments)
		case "update_concept":
			result, err = s.handleUpdateConcept(callParams.Arguments)
		case "verify_concept":
			result, err = s.handleVerifyConcept(callParams.Arguments)
		case "get_backlinks":
			result, err = s.handleGetBacklinks(callParams.Arguments)
		case "traverse_graph":
			result, err = s.handleTraverseGraph(callParams.Arguments)
		case "list_broken_links":
			result, err = s.handleListBrokenLinks()
		default:
			resp.Error = map[string]any{"code": -32601, "message": fmt.Sprintf("Tool '%s' not found", callParams.Name)}
		}

		if err != nil {
			resp.Error = err
		} else if result != nil {
			resp.Result = result
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
