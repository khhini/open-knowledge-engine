package mcp

import "encoding/json"

func (s *MCPServer) handleInspectSourceTruth(rawArgs json.RawMessage) (any, *JSONRPCError) {
	var args InspectSourceTruthArgs
	if err := DecodeArgsTo(rawArgs, &args); err != nil {
		return nil, err
	}

	inspection, err := s.simulator.Inspect(args.URI)
	if err != nil {
		return nil, newInternalError(err.Error())
	}

	return map[string]any{
		"inspection": []map[string]any{
			{"type": "text", "text": marshalJSON(inspection)},
		},
	}, nil
}
