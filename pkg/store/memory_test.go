package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf-store-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	concept1 := `---
type: concept
title: Alpha
description: Starting point
---
This points to [[beta]]`

	concept2 := `---
type: concept
title: Beta
description: Intermediate step
---
This points to [[alpha]] and [[gamma]]`

	concept3 := `---
type: concept
title: Gamma
description: Ending point
---
Leaf node.`

	if err := os.WriteFile(filepath.Join(tmpDir, "alpha.md"), []byte(concept1), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "beta.md"), []byte(concept2), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "gamma.md"), []byte(concept3), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	s := NewMemoryStore(tmpDir)
	if err := s.LoadAll(); err != nil {
		t.Fatalf("failed to load memory store: %v", err)
	}

	// 1. Check list size
	concepts := s.List()
	if len(concepts) != 3 {
		t.Errorf("expected 3 concepts, got %d", len(concepts))
	}

	// 2. Check get and lookup logic
	alphaID := "alpha"
	betaID := "beta"
	gammaID := "gamma"

	c, ok := s.Get(alphaID)
	if !ok {
		t.Fatalf("expected to find concept alpha")
	}
	if c.Frontmatter.Title != "Alpha" {
		t.Errorf("expected title 'Alpha', got '%s'", c.Frontmatter.Title)
	}

	_, ok = s.Get(gammaID)
	if !ok {
		t.Fatalf("expected to find concept gamma")
	}

	// 3. Verify Links (Outgoing / Incoming / Backlinks)
	view, ok := s.GetConceptView(betaID)
	if !ok {
		t.Fatalf("expected to get concept view for beta")
	}

	if len(view.OutgoingLinks) != 2 {
		t.Errorf("expected 2 outgoing links from beta, got %d", len(view.OutgoingLinks))
	}

	if len(view.IncomingLinks) != 1 || view.IncomingLinks[0].ID != alphaID {
		t.Errorf("expected 1 incoming link (backlink) from alpha to beta")
	}

	// 4. Search logic
	results := s.Search("Intermediate")
	if len(results) != 1 || results[0].Frontmatter.Title != "Beta" {
		t.Errorf("expected search for 'Intermediate' to return Beta")
	}
}

func TestMemoryStoreConcurrency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf-store-concurrency-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a single concept file that we'll reload repeatedly
	cContent := `---
type: concept
title: Concurrent Node
---
Body text pointing to [[another-node]]`

	if err := os.WriteFile(filepath.Join(tmpDir, "node.md"), []byte(cContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	s := NewMemoryStore(tmpDir)
	if err := s.LoadAll(); err != nil {
		t.Fatalf("failed to load store initially: %v", err)
	}

	var wg sync.WaitGroup
	workers := 10
	iterations := 100

	// Concurrent Readers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			nodeID := filepath.Join(tmpDir, "node")
			for j := 0; j < iterations; j++ {
				_, _ = s.Get(nodeID)
				_ = s.List()
				_ = s.Search("Concurrent")
				_, _ = s.GetConceptView(nodeID)
			}
		}(i)
	}

	// Concurrent Writers/Reloaders
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Modify files to trigger parser/walk
				newContent := fmt.Sprintf(`---
type: concept
title: Concurrent Node %d-%d
---
Ref [[another-node]]`, workerID, j)
				_ = os.WriteFile(filepath.Join(tmpDir, "node.md"), []byte(newContent), 0644)
				_ = s.LoadAll()
			}
		}(i)
	}

	wg.Wait()
}
