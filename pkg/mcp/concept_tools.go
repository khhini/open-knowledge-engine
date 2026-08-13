package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
	"github.com/khhini/open-knowledge-engine.git/pkg/sourcetruth"
)

func (s *MCPServer) handleReadConcept(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args ReadConceptArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	concept, ok := s.store.GetConceptView(args.ConceptID)
	if !ok {
		return nil, newInvalidParamsError("Concept not found")
	}

	respMap := map[string]any{
		"concept": concept,
	}

	if args.IncludeSourceTruth && s.simulator != nil {
		var inspections []*sourcetruth.Inspection
		visitedURIs := make(map[string]bool)

		addInspection := func(uri string) {
			if uri == "" || visitedURIs[uri] {
				return
			}

			visitedURIs[uri] = true
			if insp, err := s.simulator.Inspect(uri); err == nil {
				if len(insp.TextSnippet) > 500 {
					insp.TextSnippet = insp.TextSnippet[:500] + "\n... [Truncated for multi-source]"
				}
				inspections = append(inspections, insp)
			}
		}

		// 1. Inspect Caconical Resource if present
		addInspection(concept.Frontmatter.Resource)

		// 2. Inspect All Upstream Provence Sources
		for _, src := range concept.Frontmatter.Sources {
			addInspection(src.Resource)
		}

		respMap["source_truth_inspections"] = inspections
	}

	return map[string]any{
		"concept": []map[string]any{
			{"type": "text", "text": marshalJSON(respMap)},
		},
	}, nil
}

func (s *MCPServer) handleCreateConcept(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args CreateConceptArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	fm := &okf.Frontmatter{
		Type:        args.Type,
		Title:       args.Title,
		Description: args.Description,
		Resource:    args.Resource,
		Status:      "draft",
		Tags:        args.Tags,
		StaleAfter:  args.StaleAfter,
		Sources:     args.Sources,
		Generated: &okf.Generated{
			By: okf.Actor(args.AgentID),
			At: time.Now(),
		},
	}

	fileContent, err := okf.FormatConcept(fm, args.Body)
	if err != nil {
		return nil, newInternalError(err.Error())
	}

	fullPath := filepath.Join(s.baseDir, args.ConceptID+".md")
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err := os.WriteFile(fullPath, []byte(fileContent), 0644); err != nil {
		return nil, newInternalError(err.Error())
	}

	_ = s.store.LoadAll()

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": fmt.Sprintf("Created concept '%s' successfully.", args.ConceptID)},
		},
	}, nil

}

func (s *MCPServer) handleUpdateConcept(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args UpdateConceptArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	fullPath := filepath.Join(s.baseDir, args.ConceptID+".md")
	existingConcept, err := okf.ParseConcept(fullPath)
	if err != nil {
		return nil, newInvalidParamsError(fmt.Sprintf("Failed to parse concept file '%s': %v", fullPath, err.Error()))
	}

	var existingFM okf.Frontmatter = existingConcept.Frontmatter
	var existingBody string = existingConcept.BodyMarkdown

	// Handle frontmatter update
	if fm := args.FrontmatterUpdate; fm != nil {
		if fm.Type != "" {
			existingFM.Type = fm.Type
		}
		if fm.Title != "" {
			existingFM.Title = fm.Title
		}
		if fm.Description != "" {
			existingFM.Description = fm.Description
		}
		if fm.Status != "" {
			existingFM.Status = fm.Status
		}
		if fm.Resource != "" {
			existingFM.Resource = fm.Resource
		}
		if fm.StaleAfter != "" {
			existingFM.StaleAfter = fm.StaleAfter
		}
		if len(fm.Tags) > 0 {
			existingFM.Tags = fm.Tags
		}
		if len(fm.Sources) > 0 {
			existingFM.Sources = fm.Sources
		}
	}

	// Handle Body: 'body' (replace) va 'append_body' (append)
	if body := strings.TrimSpace(args.Body); body != "" {
		existingBody = body
	} else if appendBody := strings.TrimSpace(args.AppendBody); appendBody != "" {
		existingBody = existingBody + "\n\n" + appendBody
	}

	updatedContent, _ := okf.FormatConcept(&existingFM, existingBody)

	if err := os.WriteFile(fullPath, []byte(updatedContent), 0644); err != nil {
		return nil, newInternalError(err.Error())
	}

	_ = okf.AppendLogEntry(s.baseDir, okf.Actor(args.AgentID), "UPDATED", args.ConceptID, "Concept updated via MCP agent tool")

	_ = s.store.LoadAll()

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": fmt.Sprintf("Concept '%s' updated successfully.", args.ConceptID)},
		},
	}, nil
}

func (s *MCPServer) handleVerifyConcept(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args VerifyConceptArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	if args.Notes == "" {
		args.Notes = "Attestation submitted via MCP agent tool"
	}

	fullPath := filepath.Join(s.baseDir, args.ConceptID+".md")
	existingConcept, err := okf.ParseConcept(fullPath)
	if err != nil {
		return nil, newInvalidParamsError(fmt.Sprintf("Failed to parse concept file '%s': %v", fullPath, err.Error()))
	}

	var existingFM okf.Frontmatter = existingConcept.Frontmatter
	var existingBody string = existingConcept.BodyMarkdown

	actor := okf.Actor(args.Actor)
	existingFM.Verified = append(existingFM.Verified, okf.Verification{
		Actor: actor,
		At:    time.Now(),
		Notes: args.Notes,
	})

	updatedContent, _ := okf.FormatConcept(&existingFM, existingBody)

	if err := os.WriteFile(fullPath, []byte(updatedContent), 0644); err != nil {
		return nil, newInternalError(err.Error())
	}

	_ = okf.AppendLogEntry(s.baseDir, actor, "VERIFIED", args.ConceptID, args.Notes)
	_ = s.store.LoadAll()

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": fmt.Sprintf("Concept '%s' verified by '%s' successfully.", args.ConceptID, args.Actor)},
		},
	}, nil
}
