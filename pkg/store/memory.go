package store

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
)

type ConceptView struct {
	*okf.Concept
	OutgoingLinks []*okf.Concept `json:"outgoing_links"`
	IncomingLinks []*okf.Concept `json:"incoming_links"`
}

type MemoryStore struct {
	mu       sync.RWMutex
	baseDir  string
	concepts map[string]*okf.Concept // conceptID -> Concept
	outgoing map[string][]string     // conceptID -> list of target conceptIDs
	incoming map[string][]string     // conceptID -> list of referencing conceptIDs
}

func NewMemoryStore(baseDir string) *MemoryStore {
	return &MemoryStore{
		baseDir:  baseDir,
		concepts: make(map[string]*okf.Concept),
		outgoing: make(map[string][]string),
		incoming: make(map[string][]string),
	}
}

func (s *MemoryStore) GetBaseDir() string {
	return s.baseDir
}

func (s *MemoryStore) LoadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.concepts = make(map[string]*okf.Concept)
	s.outgoing = make(map[string][]string)
	s.incoming = make(map[string][]string)

	return filepath.WalkDir(s.baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		if d.Name() == "index.md" || d.Name() == "log.md" {
			return nil // Skip reserved files in root/subdirs if needed
		}

		concept, err := okf.ParseConcept(path)
		if err != nil {
			return nil
		}

		s.concepts[concept.ID] = concept
		s.indexLinks(concept.ID, concept.BodyMarkdown)

		return nil
	})
}

func (s *MemoryStore) Get(id string) (*okf.Concept, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.concepts[id]
	return c, ok
}

func (s *MemoryStore) List() []*okf.Concept {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*okf.Concept, 0, len(s.concepts))
	for _, c := range s.concepts {
		list = append(list, c)
	}
	return list
}

func (s *MemoryStore) Search(query string) []*okf.Concept {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if query == "" {
		return s.List()
	}

	q := strings.ToLower(query)
	var results []*okf.Concept

	for _, c := range s.concepts {
		title := strings.ToLower(c.Frontmatter.Title)
		desc := strings.ToLower(c.Frontmatter.Description)
		body := strings.ToLower(c.BodyMarkdown)

		if strings.Contains(title, q) || strings.Contains(desc, q) || strings.Contains(body, q) {
			results = append(results, c)
		}
	}
	return results
}

func (s *MemoryStore) SearchFiltered(query, tierFilter, typeFilter string) []*okf.Concept {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(query)
	results := make([]*okf.Concept, 0)

	for _, c := range s.concepts {
		if tierFilter != "" && tierFilter != "all" && c.TrustTier != tierFilter {
			continue

		}

		if typeFilter != "" && typeFilter != "all" && !strings.EqualFold(c.Frontmatter.Type, typeFilter) {
			continue
		}

		if q != "" {
			title := strings.ToLower(c.Frontmatter.Title)
			desc := strings.ToLower(c.Frontmatter.Description)
			body := strings.ToLower(c.BodyMarkdown)
			if !strings.Contains(title, q) && !strings.Contains(desc, q) && !strings.Contains(body, q) {
				continue
			}
		}

		results = append(results, c)
	}

	return results
}

func (s *MemoryStore) indexLinks(sourceID, body string) {
	re := regexp.MustCompile(`\[\[(.*?)\]\]`)
	matches := re.FindAllStringSubmatch(body, -1)

	var targets []string
	for _, match := range matches {
		if len(match) > 1 {
			target := match[1]
			targets = append(targets, target)
			s.incoming[target] = append(s.incoming[target], sourceID)
		}
	}

	s.outgoing[sourceID] = targets
}

func (s *MemoryStore) GetConceptView(id string) (*ConceptView, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.concepts[id]
	if !ok {
		return nil, false
	}

	view := &ConceptView{Concept: c,
		OutgoingLinks: make([]*okf.Concept, 0),
		IncomingLinks: make([]*okf.Concept, 0),
	}

	// Resolve Outgoing Links
	if targetIDs, hasOutoging := s.outgoing[id]; hasOutoging {
		// Deduplicate incoming links
		seen := make(map[string]bool)
		for _, targetID := range targetIDs {
			if !seen[targetID] {
				seen[targetID] = true

				if target, exists := s.concepts[targetID]; exists {
					view.OutgoingLinks = append(view.OutgoingLinks, target)
				}
			}
		}
	}

	// Resolve Incoming Links (Backlinks)
	if sourceIDs, hasIncoming := s.incoming[id]; hasIncoming {
		// Deduplicate incoming links
		seen := make(map[string]bool)
		for _, sourceID := range sourceIDs {
			if !seen[sourceID] {
				seen[sourceID] = true

				if source, exists := s.concepts[sourceID]; exists {
					view.IncomingLinks = append(view.IncomingLinks, source)
				}
			}
		}

	}

	return view, true
}
