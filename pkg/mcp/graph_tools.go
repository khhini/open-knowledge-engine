package mcp

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Backlink struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	TrustTier   string `json:"trust_tier"`
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

type queueItem struct {
	id    string
	depth int
}

type BrokenLink struct {
	SourceConceptID string `json:"source_concept_id"`
	SourceTitle     string `json:"source_title"`
	TargetWikilink  string `json:"target_wikilink"`
}

func (s *MCPServer) handleSearchKnowledge(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args SearchKnowledgeArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	results := s.store.SearchFiltered(args.Query, args.TrustTier, args.ConceptType)
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": marshalJSON(results)},
		},
	}, nil
}

func (s *MCPServer) handleGetBacklinks(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args GetBacklinksArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	conceptView, ok := s.store.GetConceptView(args.ConceptID)
	if !ok {
		return nil, newInvalidParamsError(fmt.Sprintf("Concept '%s' not found", args.ConceptID))
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

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": marshalJSON(map[string]any{
				"concept_id":  args.ConceptID,
				"backlinks":   backlinks,
				"total_count": len(backlinks),
			})},
		},
	}, nil
}

func (s *MCPServer) handleTraverseGraph(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args TraverseGraphArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	if args.MaxDepth > 4 {
		args.MaxDepth = 4
	}

	if args.Direction == "" {
		args.Direction = "both"
	}

	if _, ok := s.store.GetConceptView(args.RootConceptID); !ok {
		return nil, newInvalidParamsError(fmt.Sprintf("Root concept '%s' not found", args.RootConceptID))
	}

	nodesMap := make(map[string]GraphNode)
	edgesMap := make(map[string]GraphEdge)

	queue := []queueItem{{id: args.RootConceptID, depth: 0}}

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

		if curr.depth >= args.MaxDepth {
			continue
		}

		if args.Direction == "both" || args.Direction == "outgoing" {
			for _, out := range view.OutgoingLinks {
				edgeKey := curr.id + "->" + out.ID
				edgesMap[edgeKey] = GraphEdge{Source: curr.id, Target: out.ID}
				if _, visited := nodesMap[out.ID]; !visited {
					queue = append(queue, queueItem{id: out.ID, depth: curr.depth + 1})
				}
			}
		}

		if args.Direction == "both" || args.Direction == "incoming" {
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

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": marshalJSON(map[string]any{
				"root_concept_id": args.RootConceptID,
				"max_depth":       args.MaxDepth,
				"direction":       args.Direction,
				"nodes":           nodeList,
				"edges":           edgeList,
			})},
		},
	}, nil

}

func (s *MCPServer) handleListBrokenLinks() (any, *JSONRPCError) {
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

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": marshalJSON(map[string]any{
				"broken_links": brokenLinks,
				"total_count":  len(brokenLinks),
			})},
		},
	}, nil
}
