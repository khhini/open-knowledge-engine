package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
)

type Node struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	TrustTier string `json:"trustTier"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Edge `json:"links"`
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/search", s.handleSearch)
	s.mux.HandleFunc("/concept", s.handleGetConcept)
	s.mux.HandleFunc("/concept/verify", s.handlerVerifyConcept)

	s.mux.HandleFunc("/api/graph", s.handleGraph)
	s.mux.HandleFunc("/api/source-truth/inspect", s.handleSourceTruthInspect)

	s.mux.Handle("/mcp", s.mcpServer)
	s.mux.HandleFunc("/mcp/sse", s.sseTransport.HandleSSE)
	s.mux.HandleFunc("/mcp/message", s.sseTransport.HandleMessage)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	tier := r.URL.Query().Get("tier")
	if tier == "" {
		tier = "all"
	}

	conceptType := r.URL.Query().Get("type")
	if conceptType == "" {
		conceptType = "all"
	}

	allConcepts := s.store.List()
	typeSet := make(map[string]bool)
	for _, c := range allConcepts {
		if c.Frontmatter.Type != "" {
			typeSet[c.Frontmatter.Type] = true
		}
	}

	availableTypes := make([]string, 0, len(typeSet))
	for t := range typeSet {
		availableTypes = append(availableTypes, t)
	}

	if err := s.tmpl.ExecuteTemplate(w, "index.html", map[string]any{
		"Concepts":       allConcepts,
		"SelectedTier":   tier,
		"SelectedType":   conceptType,
		"Query":          q,
		"AvailableTypes": availableTypes,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	tier := r.URL.Query().Get("tier")
	if tier == "" {
		tier = "all"
	}

	conceptType := r.URL.Query().Get("type")
	if conceptType == "" {
		conceptType = "all"
	}

	results := s.store.SearchFiltered(q, tier, conceptType)

	allConcepts := s.store.List()
	typeSet := make(map[string]bool)
	for _, c := range allConcepts {
		if c.Frontmatter.Type != "" {
			typeSet[c.Frontmatter.Type] = true
		}
	}

	availableTypes := make([]string, 0, len(typeSet))
	for t := range typeSet {
		availableTypes = append(availableTypes, t)
	}

	if err := s.tmpl.ExecuteTemplate(w, "list.html", map[string]any{
		"Concepts":       results,
		"SelectedTier":   tier,
		"SelectedType":   conceptType,
		"Query":          q,
		"AvailableTypes": availableTypes,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s *Server) handleGetConcept(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	conceptView, ok := s.store.GetConceptView(id)
	if !ok {
		http.Error(w, "Concept Not Found", http.StatusNotFound)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "concept.html", conceptView); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handlerVerifyConcept(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	concept, ok := s.store.Get(id)
	if !ok {
		http.Error(w, "Concept Not Found", http.StatusNotFound)
		return
	}

	actor := okf.Actor("human:khhini")

	concept.Frontmatter.Verified = append(concept.Frontmatter.Verified, okf.Verification{
		Actor: actor,
		At:    time.Now(),
		Notes: "Verified vai Web UI",
	})
	concept.TrustTier = string(okf.EvaluateTrustTier(&concept.Frontmatter, time.Now()))

	// Record to log.md & refresh index.md
	_ = okf.AppendLogEntry(s.store.GetBaseDir(), actor, "VERIFIED", concept.ID, "Human verification approved via web UI")

	_ = okf.GenerateBundleIndex(s.store.List(), s.store.GetBaseDir())

	if err := s.tmpl.ExecuteTemplate(w, "trust_badge.html", concept); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleSourceTruthInspect(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	inspection, err := s.simulator.Inspect(uri)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "source_inspector.html", inspection); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	concepts := s.store.List()

	nodes := make([]Node, 0, len(concepts))
	links := make([]Edge, 0)
	nodeMap := make(map[string]bool)

	for _, c := range concepts {
		nodeMap[c.ID] = true
		title := c.Frontmatter.Title
		if title == "" {
			title = c.ID
		}
		nodes = append(nodes, Node{
			ID:        c.ID,
			Title:     title,
			Type:      c.Frontmatter.Type,
			TrustTier: c.TrustTier,
		})

		// Add outgoing edges
		view, ok := s.store.GetConceptView(c.ID)
		if ok {
			for _, target := range view.OutgoingLinks {
				links = append(links, Edge{
					Source: c.ID,
					Target: target.ID,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(GraphData{Nodes: nodes, Links: links}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
