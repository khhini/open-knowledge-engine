package main

import (
	"log"
	"os"

	"github.com/khhini/open-knowledge-engine.git/pkg/mcp"
	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
	"github.com/khhini/open-knowledge-engine.git/pkg/server"
	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
	"github.com/khhini/open-knowledge-engine.git/pkg/watcher"
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
			log.Printf("Warning: failed to generate index.md: %v", err)
		}
	}

	fw, err := watcher.StartWatcher(knowledgeDir, memStore)
	if err != nil {
		log.Printf("Warning: File watcher failed to initialize: %v", err)
	} else {
		defer fw.Close()
	}

	svr := server.NewServer(memStore, simulator, mcpServer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running at http://localhost:%s\n", port)
	if err := svr.Start(port); err != nil {
		log.Fatal(err.Error())

	}

}
