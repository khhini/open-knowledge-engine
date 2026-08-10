package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/khhini/open-knowledge-engine.git/pkg/mcp"
	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
	"github.com/khhini/open-knowledge-engine.git/pkg/watcher"
	"github.com/khhini/open-knowledge-engine.git/templates"
)

func main() {
	knowledgeDir := "./knowledge"
	memStore := store.NewMemoryStore(knowledgeDir)
	simulator := sourcetruth.NewSimulator(".source_of_truth")
	mcpServer := mcp.NewMCPServer(memStore, knowledgeDir, simulator)

	if err := memStore.LoadAll(); err != nil {
		log.Fatalf("Failed to load knowledge bundle: %v", err)
	}

	if concepts := memStore.List(); len(concepts) > 0 {
		if err := okf.GenerateBundleIndex(concepts, knowledgeDir); err != nil {
			log.Printf("Warnig: failed to generate index.md: %v", err)
		}
	}

	fw, err := watcher.StartWatcher(knowledgeDir, memStore)
	if err != nil {
		log.Printf("Waring: File watcher failed to initialize: %v", err)
	} else {
		defer fw.Close()
	}

	tmpl := template.Must(template.ParseFS(templates.Files, "index.html", "fragments/*.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		tier := r.URL.Query().Get("tier")
		if tier == "" {
			tier = "all"
		}

		conceptType := r.URL.Query().Get("type")
		if conceptType == "" {
			conceptType = "all"
		}

		allConcepts := memStore.List()
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

		if err = tmpl.ExecuteTemplate(w, "index.html", map[string]any{
			"Concepts":       allConcepts,
			"SelectedTier":   tier,
			"SelectedType":   conceptType,
			"Query":          q,
			"AvailableTypes": availableTypes,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		tier := r.URL.Query().Get("tier")
		if tier == "" {
			tier = "all"
		}

		conceptType := r.URL.Query().Get("type")
		if conceptType == "" {
			conceptType = "all"
		}

		results := memStore.SearchFiltered(q, tier, conceptType)

		allConcepts := memStore.List()
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

		if err = tmpl.ExecuteTemplate(w, "list.html", map[string]any{
			"Concepts":       results,
			"SelectedTier":   tier,
			"SelectedType":   conceptType,
			"Query":          q,
			"AvailableTypes": availableTypes,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})

	http.HandleFunc("/concept", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		conceptView, ok := memStore.GetConceptView(id)
		if !ok {
			http.Error(w, "Concept Not Found", http.StatusNotFound)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "concept.html", conceptView); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/concept/verify", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		concept, ok := memStore.Get(id)
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
		_ = okf.AppendLogEntry(knowledgeDir, actor, "VERIFIED", concept.ID, "Human verification approved via web UI")

		_ = okf.GenerateBundleIndex(memStore.List(), knowledgeDir)

		if err = tmpl.ExecuteTemplate(w, "trust_badge.html", concept); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

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

	http.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		concepts := memStore.List()

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
			view, ok := memStore.GetConceptView(c.ID)
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
	})

	http.HandleFunc("/api/source-truth/inspect", func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("uri")
		inspection, err := simulator.Inspect(uri)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "source_inspector.html", inspection); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.Handle("/mcp", mcpServer)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
