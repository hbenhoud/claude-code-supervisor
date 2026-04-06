package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hbenhoud/claude-code-supervisor/internal/normalizer"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub manages WebSocket clients and broadcasts events per session.
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]map[*Client]bool // sessionID -> set of clients
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	sessionID string
	send      chan []byte
	closeOnce sync.Once
}

type subscribeMsg struct {
	Subscribe     string `json:"subscribe"`
	AfterSequence int    `json:"afterSequence,omitempty"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.sessionID] == nil {
				h.clients[client.sessionID] = make(map[*Client]bool)
			}
			h.clients[client.sessionID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.sessionID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					client.closeOnce.Do(func() { close(client.send) })
					if len(clients) == 0 {
						delete(h.clients, client.sessionID)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Broadcast(sessionID string, evt *normalizer.SupervisorEvent) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[sessionID]; ok {
		for client := range clients {
			select {
			case client.send <- data:
			default:
				// Client too slow, drop message
			}
		}
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Read subscribe message
	var msg subscribeMsg
	if err := conn.ReadJSON(&msg); err != nil {
		conn.Close()
		return
	}

	client := &Client{
		hub:       s.hub,
		conn:      conn,
		sessionID: msg.Subscribe,
		send:      make(chan []byte, 256),
	}

	s.hub.register <- client

	// Historical events are now served via REST (GET /api/sessions/:id/events).
	// The WebSocket only sends live events arriving after afterSequence.
	// If afterSequence > 0, the client already has historical data from REST.

	// Writer goroutine: reads from send channel, writes to WebSocket
	go func() {
		defer conn.Close()
		for data := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()

	// Reader goroutine: detects disconnect, triggers cleanup once
	go func() {
		defer func() {
			s.hub.unregister <- client
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}
