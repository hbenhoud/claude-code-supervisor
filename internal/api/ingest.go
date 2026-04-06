package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hbenhoud/claude-code-supervisor/internal/hooks"
)

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	// Read body and log for debugging
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	log.Printf("[ingest] %s", string(body))

	var raw hooks.HookEvent
	if err := json.Unmarshal(body, &raw); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if raw.GetSessionID() == "" {
		http.Error(w, "missing session ID", http.StatusBadRequest)
		return
	}

	if raw.Timestamp.IsZero() {
		raw.Timestamp = time.Now()
	}

	// Normalize the raw hook event
	evt := s.normalizer.Normalize(raw)

	// Ensure session exists
	s.db.EnsureSession(evt.SessionID, raw.CWD)

	// Persist event
	if err := s.db.InsertEvent(evt); err != nil {
		log.Printf("Failed to persist event: %v", err)
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	// Update session counters
	if evt.EventType == "tool_call" && evt.EventSubtype == "complete" {
		s.db.IncrementToolCount(evt.SessionID)
	}
	if evt.EventType == "agent_spawn" && evt.EventSubtype == "start" {
		s.db.IncrementAgentCount(evt.SessionID)
	}

	// Broadcast to WebSocket clients
	s.hub.Broadcast(evt.SessionID, evt)

	w.WriteHeader(http.StatusOK)
}
