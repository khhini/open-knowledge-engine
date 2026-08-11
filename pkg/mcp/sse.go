package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SSESession struct {
	ID          string
	MessageChan chan string
}

type SSETransport struct {
	MCPServer *MCPServer
	mu        sync.RWMutex
	sessions  map[string]*SSESession
}

func NewSSETransport(mcpServer *MCPServer) *SSETransport {
	return &SSETransport{
		MCPServer: mcpServer,
		sessions:  make(map[string]*SSESession),
	}
}

func (t *SSETransport) HandleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
	}

	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	msgChan := make(chan string, 100)

	sess := &SSESession{
		ID:          sessionID,
		MessageChan: msgChan,
	}

	t.mu.Lock()
	t.sessions[sessionID] = sess
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.sessions, sessionID)
		t.mu.Unlock()
		close(msgChan)
	}()

	messageEndpoint := fmt.Sprintf("/mcp/message?session_id=%s", sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageEndpoint)
	flusher.Flush()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (t *SSETransport) HandleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id parameter", http.StatusBadRequest)
		return
	}

	t.mu.Lock()
	sess, ok := t.sessions[sessionID]
	t.mu.RUnlock()

	if !ok {
		http.Error(w, "Session not found or expired", http.StatusNotFound)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := t.MCPServer.HandleRPC(req)
	if resp != nil {
		respJSON := marshalJSON(resp)
		select {
		case sess.MessageChan <- respJSON:
		default:
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status": "accepted"}`))

}
