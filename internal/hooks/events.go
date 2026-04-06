package hooks

import (
	"encoding/json"
	"time"
)

// HookType identifies which Claude Code hook fired.
type HookType string

const (
	HookPreToolUse   HookType = "pre_tool_use"
	HookPostToolUse  HookType = "post_tool_use"
	HookNotification HookType = "notification"
)

// HookEvent is the raw payload received from a Claude Code hook via HTTP POST.
// Fields are populated from the jq-transformed stdin JSON.
// We also keep the full raw JSON for fields we don't explicitly parse.
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

	// Raw fields forwarded from Claude Code's stdin JSON
	SessionIDRaw    string          `json:"session_id,omitempty"`
	ToolNameRaw     string          `json:"tool_name,omitempty"`
	ToolInputRaw    json.RawMessage `json:"tool_input,omitempty"`
	ToolOutputRaw   json.RawMessage `json:"tool_output,omitempty"`
	ToolUseID       string          `json:"tool_use_id,omitempty"`
	ParentToolUseID string          `json:"parent_tool_use_id,omitempty"`
}

// GetSessionID returns the session ID from whichever field has data.
func (e *HookEvent) GetSessionID() string {
	if e.SessionID != "" {
		return e.SessionID
	}
	return e.SessionIDRaw
}

// GetToolName returns the tool name from whichever field has data.
func (e *HookEvent) GetToolName() string {
	if e.Tool != "" {
		return e.Tool
	}
	return e.ToolNameRaw
}

// GetInput returns tool input from whichever field has data.
func (e *HookEvent) GetInput() json.RawMessage {
	if e.Input != nil {
		return e.Input
	}
	return e.ToolInputRaw
}

// GetOutput returns tool output from whichever field has data.
func (e *HookEvent) GetOutput() json.RawMessage {
	if e.Output != nil {
		return e.Output
	}
	return e.ToolOutputRaw
}
