package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/khhini/open-knowledge-engine.git/pkg/mcp"
	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
	"github.com/khhini/open-knowledge-engine.git/templates"
)

type Server struct {
	store        *store.MemoryStore
	simulator    *sourcetruth.Simulator
	mcpServer    *mcp.MCPServer
	sseTransport *mcp.SSETransport
	tmpl         *template.Template
	mux          *http.ServeMux
}

func NewServer(s *store.MemoryStore, sim *sourcetruth.Simulator, mcpServer *mcp.MCPServer) *Server {
	srv := &Server{
		store:        s,
		simulator:    sim,
		mcpServer:    mcpServer,
		sseTransport: mcp.NewSSETransport(mcpServer),
		tmpl:         template.Must(template.ParseFS(templates.Files, "index.html", "fragments/*.html")),
		mux:          &http.ServeMux{},
	}
	srv.registerRoutes()
	return srv
}

func (s *Server) Start(port string) error {
	return http.ListenAndServe(fmt.Sprintf(":%s", port), s.mux)
}
