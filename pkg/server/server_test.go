package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/khhini/open-knowledge-engine.git/pkg/mcp"
	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
)

func TestServerRoutes(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "okf-server-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	s := store.NewMemoryStore(tmpDir)
	sim := sourcetruth.NewSimulator(tmpDir)
	mcpServer := mcp.NewMCPServer(s, tmpDir, sim)

	srv := NewServer(s, sim, mcpServer)
	if srv == nil {
		t.Fatal("expected server to not be nil")
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"Index Page", "/", http.StatusOK},
		{"Search Page", "/search", http.StatusOK},
		{"Concept Page - Not Found", "/concept?id=nonexistent", http.StatusNotFound},
		{"API Graph", "/api/graph", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			srv.mux.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}
