package mcp

func (s *MCPServer) listTools() []map[string]any {
	return []map[string]any{
		{
			"name":        "search_knowledge",
			"description": "Search OKF v0.2 knowledge corpus by keyword query with optional tier and type filters",
			"inputSchema": argsSchema[SearchKnowledgeArgs](),
		},
		{
			"name":        "read_concept",
			"description": "Read a complete OKF v0.2 concept by Concept ID including metadata, body, links, and optional source of truth inspection",
			"inputSchema": argsSchema[ReadConceptArgs](),
		},

		{
			"name":        "create_concept",
			"description": "Create a new OKF v0.2 concept markdown file with YAML frontmatter",
			"inputSchema": argsSchema[CreateConceptArgs](),
		},
		{
			"name":        "inspect_source_of_truth",
			"description": "Inspect technical meatadata, document text snippets, spreadsheet columns, and local simulation files for an external source URI",
			"inputSchema": argsSchema[InspectSourceTruthArgs](),
		},
		{
			"name":        "update_concept",
			"description": "Modify an existing OKF concept's frontmatter fields, replace body, or  append body content",
			"inputSchema": argsSchema[UpdateConceptArgs](),
		},
		{
			"name":        "verify_concept",
			"description": "Submit a machine or human attestation for a concept, prompting its Trust Tier",
			"inputSchema": argsSchema[VerifyConceptArgs](),
		},
		{
			"name":        "get_backlinks",
			"description": "Retrieve all incoming backlinks pointing to a concept ID",
			"inputSchema": argsSchema[GetBacklinksArgs](),
		},
		{
			"name":        "traverse_graph",
			"description": "Traverse the knowledge graph N-hops starting from a root concept ID",
			"inputSchema": argsSchema[TraverseGraphArgs](),
		},
		{
			"name":        "list_broken_links",
			"description": "Scan the corpus for broken/dangling [[wikilinks]] pointing to  missing concepts",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}
