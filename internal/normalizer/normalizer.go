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

// openAgent tracks a currently running sub-agent for temporal window attribution.
type openAgent struct {
	toolUseID string
	agentID   string // "agent-{toolUseID[:8]}"
}

// Normalizer transforms raw hook events into canonical SupervisorEvents.
type Normalizer struct {
	mu           sync.Mutex
	seqBySession map[string]int
	// pendingPre tracks PreToolUse events by (session_id, tool_use_id) for duration pairing
	pendingPre map[string]*pendingToolUse
	// openAgents tracks currently running sub-agents per session for temporal window attribution.
	// Key: sessionID, Value: map of toolUseID -> openAgent
	openAgents map[string]map[string]*openAgent
}

type pendingToolUse struct {
	Event     *SupervisorEvent
	Timestamp time.Time
}

func New() *Normalizer {
	return &Normalizer{
		seqBySession: make(map[string]int),
		pendingPre:   make(map[string]*pendingToolUse),
		openAgents:   make(map[string]map[string]*openAgent),
	}
}

func (n *Normalizer) nextSequence(sessionID string) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.seqBySession[sessionID]++
	return n.seqBySession[sessionID]
}

// resolveAgent determines the agent_id for a tool call using temporal windows.
// If exactly one sub-agent is open for this session, attribute the tool call to it.
// If zero or multiple are open, attribute to root.
func (n *Normalizer) resolveAgent(sessionID string) (agentID, parentAgentID string) {
	agents := n.openAgents[sessionID]
	if len(agents) == 1 {
		for _, a := range agents {
			return a.agentID, "root"
		}
	}
	return "root", ""
}

// Normalize transforms a raw HookEvent into a SupervisorEvent.
func (n *Normalizer) Normalize(raw hooks.HookEvent) *SupervisorEvent {
	now := time.Now()
	if !raw.Timestamp.IsZero() {
		now = raw.Timestamp
	}

	sessionID := raw.GetSessionID()
	toolName := raw.GetToolName()
	seq := n.nextSequence(sessionID)

	switch raw.Hook {
	case hooks.HookPreToolUse:
		// If this is an Agent tool call, register it as an open agent
		if toolName == "Agent" && raw.ToolUseID != "" {
			aid := "agent-" + raw.ToolUseID[:min(16, len(raw.ToolUseID))]
			n.mu.Lock()
			if n.openAgents[sessionID] == nil {
				n.openAgents[sessionID] = make(map[string]*openAgent)
			}
			n.openAgents[sessionID][raw.ToolUseID] = &openAgent{
				toolUseID: raw.ToolUseID,
				agentID:   aid,
			}
			n.mu.Unlock()

			evt := &SupervisorEvent{
				ID:           generateID(),
				SessionID:    sessionID,
				Timestamp:    now,
				Sequence:     seq,
				EventType:    "agent_spawn",
				EventSubtype: "start",
				AgentID:      "root",
				ToolName:     toolName,
				ToolUseID:    raw.ToolUseID,
				ToolInput:    raw.GetInput(),
			}
			// Store as pending for duration pairing
			key := sessionID + ":" + raw.ToolUseID
			n.mu.Lock()
			n.pendingPre[key] = &pendingToolUse{Event: evt, Timestamp: now}
			n.mu.Unlock()
			return evt
		}

		// Regular tool call — resolve agent via temporal window
		n.mu.Lock()
		agentID, parentAgentID := n.resolveAgent(sessionID)
		n.mu.Unlock()

		evt := &SupervisorEvent{
			ID:            generateID(),
			SessionID:     sessionID,
			Timestamp:     now,
			Sequence:      seq,
			EventType:     "tool_call",
			EventSubtype:  "start",
			AgentID:       agentID,
			ParentAgentID: parentAgentID,
			ToolName:      toolName,
			ToolUseID:     raw.ToolUseID,
			ToolInput:     raw.GetInput(),
		}

		// Store as pending for duration pairing
		if raw.ToolUseID != "" {
			key := sessionID + ":" + raw.ToolUseID
			n.mu.Lock()
			n.pendingPre[key] = &pendingToolUse{Event: evt, Timestamp: now}
			n.mu.Unlock()
		}

		return evt

	case hooks.HookPostToolUse:
		// If this is an Agent tool completing, close the window
		if toolName == "Agent" && raw.ToolUseID != "" {
			n.mu.Lock()
			if agents, ok := n.openAgents[sessionID]; ok {
				delete(agents, raw.ToolUseID)
				if len(agents) == 0 {
					delete(n.openAgents, sessionID)
				}
			}
			n.mu.Unlock()

			evt := &SupervisorEvent{
				ID:           generateID(),
				SessionID:    sessionID,
				Timestamp:    now,
				Sequence:     seq,
				EventType:    "agent_spawn",
				EventSubtype: "complete",
				AgentID:      "root",
				ToolName:     toolName,
				ToolUseID:    raw.ToolUseID,
				ToolInput:    raw.GetInput(),
				ToolOutput:   raw.GetOutput(),
			}

			// Pair with pending PreToolUse for duration
			key := sessionID + ":" + raw.ToolUseID
			n.mu.Lock()
			if pending, ok := n.pendingPre[key]; ok {
				duration := now.Sub(pending.Timestamp).Milliseconds()
				evt.DurationMs = &duration
				delete(n.pendingPre, key)
			}
			n.mu.Unlock()

			return evt
		}

		// Regular tool call completing — resolve agent via temporal window
		n.mu.Lock()
		agentID, parentAgentID := n.resolveAgent(sessionID)
		n.mu.Unlock()

		evt := &SupervisorEvent{
			ID:            generateID(),
			SessionID:     sessionID,
			Timestamp:     now,
			Sequence:      seq,
			EventType:     "tool_call",
			EventSubtype:  "complete",
			AgentID:       agentID,
			ParentAgentID: parentAgentID,
			ToolName:      toolName,
			ToolUseID:     raw.ToolUseID,
			ToolInput:     raw.GetInput(),
			ToolOutput:    raw.GetOutput(),
		}

		// Pair with pending PreToolUse for duration
		if raw.ToolUseID != "" {
			key := sessionID + ":" + raw.ToolUseID
			n.mu.Lock()
			if pending, ok := n.pendingPre[key]; ok {
				duration := now.Sub(pending.Timestamp).Milliseconds()
				evt.DurationMs = &duration
				delete(n.pendingPre, key)
			}
			n.mu.Unlock()
		}

		return evt

	case hooks.HookNotification:
		return &SupervisorEvent{
			ID:           generateID(),
			SessionID:    sessionID,
			Timestamp:    now,
			Sequence:     seq,
			EventType:    "notification",
			EventSubtype: raw.Title,
			AgentID:      "root",
		}

	default:
		return &SupervisorEvent{
			ID:        generateID(),
			SessionID: sessionID,
			Timestamp: now,
			Sequence:  seq,
			EventType: string(raw.Hook),
			AgentID:   "root",
		}
	}
}

func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%10000)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
