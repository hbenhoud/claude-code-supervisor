package hooks

import (
	"encoding/json"
	"time"
)

// HookType identifies which Claude Code hook fired.
type HookType string

const (
	HookPreToolUse    HookType = "pre_tool_use"
	HookPostToolUse   HookType = "post_tool_use"
	HookNotification  HookType = "notification"
	HookSubagentSpawn HookType = "subagent_spawn"
)

// HookEvent is the raw payload received from a Claude Code hook via HTTP POST.
type HookEvent struct {
	Hook      HookType        `json:"hook"`
	SessionID string          `json:"session"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
	Tool      string          `json:"tool,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Output    json.RawMessage `json:"output,omitempty"`
	CWD       string          `json:"cwd,omitempty"`

	// Notification fields
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`

	// SubagentSpawn fields
	AgentID       string `json:"agent_id,omitempty"`
	ParentAgentID string `json:"parent_id,omitempty"`
	Description   string `json:"description,omitempty"`
}
