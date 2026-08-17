package mcp

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newInvalidParamsError(msg string) *JSONRPCError {
	return &JSONRPCError{Code: -32602, Message: msg}
}

func newInternalError(msg string) *JSONRPCError {
	return &JSONRPCError{Code: -32603, Message: msg}
}
