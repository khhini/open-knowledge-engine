package mcp

import "encoding/json"

func marshalJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}
