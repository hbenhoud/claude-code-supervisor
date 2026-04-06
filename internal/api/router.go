package api

import (
	"net/http"

	"github.com/hbenhoud/claude-code-supervisor/internal/normalizer"
	"github.com/hbenhoud/claude-code-supervisor/internal/store"
)

type Server struct {
	db         *store.DB
	normalizer *normalizer.Normalizer
	hub        *Hub
}

func NewRouter(db *store.DB) http.Handler {
	s := &Server{
		db:         db,
		normalizer: normalizer.New(),
		hub:        NewHub(),
	}

	go s.hub.Run()

	mux := http.NewServeMux()

	// Event ingestion from hooks
	mux.HandleFunc("POST /api/events", s.handleIngest)

	// REST API
	mux.HandleFunc("GET /api/sessions", s.handleListSessions)
	mux.HandleFunc("GET /api/sessions/{id}", s.handleGetSession)
	mux.HandleFunc("GET /api/sessions/{id}/events", s.handleGetSessionEvents)

	// WebSocket
	mux.HandleFunc("GET /ws", s.handleWebSocket)

	// CORS middleware for dev
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
