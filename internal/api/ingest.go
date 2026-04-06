package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/hbenhoud/claude-code-supervisor/internal/hooks"
)

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	var raw hooks.HookEvent
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if raw.SessionID == "" {
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

	// Broadcast to WebSocket clients
	s.hub.Broadcast(evt.SessionID, evt)

	w.WriteHeader(http.StatusOK)
}
