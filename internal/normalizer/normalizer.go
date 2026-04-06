package normalizer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hbenhoud/claude-code-supervisor/internal/hooks"
)

// SupervisorEvent is the canonical event format stored and broadcast.
type SupervisorEvent struct {
	ID            string          `json:"id"`
	SessionID     string          `json:"session_id"`
	Timestamp     time.Time       `json:"timestamp"`
	Sequence      int             `json:"sequence"`
	EventType     string          `json:"event_type"`
	EventSubtype  string          `json:"event_subtype"`
	AgentID       string          `json:"agent_id"`
	ParentAgentID string          `json:"parent_agent_id,omitempty"`
	ToolName      string          `json:"tool_name,omitempty"`
	ToolUseID     string          `json:"tool_use_id,omitempty"`
	ToolInput     json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput    json.RawMessage `json:"tool_output,omitempty"`
	DurationMs    *int64          `json:"duration_ms,omitempty"`
	Error         string          `json:"error,omitempty"`
}

// Normalizer transforms raw hook events into canonical SupervisorEvents.
type Normalizer struct {
	mu       sync.Mutex
	seqBySession map[string]int
	pendingPre   map[string]*pendingToolUse // keyed by session_id + tool_name + sequence
}

type pendingToolUse struct {
	Event     *SupervisorEvent
	Timestamp time.Time
}

func New() *Normalizer {
	return &Normalizer{
		seqBySession: make(map[string]int),
		pendingPre:   make(map[string]*pendingToolUse),
	}
}

func (n *Normalizer) nextSequence(sessionID string) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.seqBySession[sessionID]++
	return n.seqBySession[sessionID]
}

// Normalize transforms a raw HookEvent into a SupervisorEvent.
func (n *Normalizer) Normalize(raw hooks.HookEvent) *SupervisorEvent {
	now := time.Now()
	if !raw.Timestamp.IsZero() {
		now = raw.Timestamp
	}

	seq := n.nextSequence(raw.SessionID)

	switch raw.Hook {
	case hooks.HookPreToolUse:
		evt := &SupervisorEvent{
			ID:           generateID(),
			SessionID:    raw.SessionID,
			Timestamp:    now,
			Sequence:     seq,
			EventType:    "tool_call",
			EventSubtype: "start",
			AgentID:      "root",
			ToolName:     raw.Tool,
			ToolInput:    raw.Input,
		}
		// Store as pending for pairing with PostToolUse
		key := raw.SessionID + ":" + raw.Tool
		n.mu.Lock()
		n.pendingPre[key] = &pendingToolUse{Event: evt, Timestamp: now}
		n.mu.Unlock()
		return evt

	case hooks.HookPostToolUse:
		evt := &SupervisorEvent{
			ID:           generateID(),
			SessionID:    raw.SessionID,
			Timestamp:    now,
			Sequence:     seq,
			EventType:    "tool_call",
			EventSubtype: "complete",
			AgentID:      "root",
			ToolName:     raw.Tool,
			ToolInput:    raw.Input,
			ToolOutput:   raw.Output,
		}
		// Try to pair with pending PreToolUse
		key := raw.SessionID + ":" + raw.Tool
		n.mu.Lock()
		if pending, ok := n.pendingPre[key]; ok {
			duration := now.Sub(pending.Timestamp).Milliseconds()
			evt.DurationMs = &duration
			delete(n.pendingPre, key)
		}
		n.mu.Unlock()
		return evt

	case hooks.HookNotification:
		return &SupervisorEvent{
			ID:           generateID(),
			SessionID:    raw.SessionID,
			Timestamp:    now,
			Sequence:     seq,
			EventType:    "notification",
			EventSubtype: raw.Title,
			AgentID:      "root",
		}

	case hooks.HookSubagentSpawn:
		return &SupervisorEvent{
			ID:            generateID(),
			SessionID:     raw.SessionID,
			Timestamp:     now,
			Sequence:      seq,
			EventType:     "agent_spawn",
			AgentID:       raw.AgentID,
			ParentAgentID: raw.ParentAgentID,
		}

	default:
		return &SupervisorEvent{
			ID:        generateID(),
			SessionID: raw.SessionID,
			Timestamp: now,
			Sequence:  seq,
			EventType: string(raw.Hook),
			AgentID:   "root",
		}
	}
}

func generateID() string {
	// Simple unique ID generator using timestamp + random
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%10000)
}
