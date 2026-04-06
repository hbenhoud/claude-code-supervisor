package hooks

import (
	"encoding/json"
	"testing"
)

func TestParsePreToolUsePayload(t *testing.T) {
	raw := `{
		"hook": "pre_tool_use",
		"session_id": "session-abc-123",
		"tool_name": "Read",
		"tool_input": {"file_path": "/tmp/test.go"},
		"tool_use_id": "toolu_123",
		"cwd": "/home/user/project"
	}`

	var evt HookEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if evt.Hook != HookPreToolUse {
		t.Errorf("expected hook=%s, got %s", HookPreToolUse, evt.Hook)
	}
	if evt.GetSessionID() != "session-abc-123" {
		t.Errorf("expected session=session-abc-123, got %s", evt.GetSessionID())
	}
	if evt.GetToolName() != "Read" {
		t.Errorf("expected tool=Read, got %s", evt.GetToolName())
	}
	if evt.ToolUseID != "toolu_123" {
		t.Errorf("expected tool_use_id=toolu_123, got %s", evt.ToolUseID)
	}
	if evt.CWD != "/home/user/project" {
		t.Errorf("expected cwd=/home/user/project, got %s", evt.CWD)
	}
}

func TestParsePostToolUsePayload(t *testing.T) {
	raw := `{
		"hook": "post_tool_use",
		"session_id": "session-abc-123",
		"tool_name": "Edit",
		"tool_use_id": "toolu_456",
		"tool_input": {"file_path": "/tmp/test.go"},
		"tool_output": {"type": "patch"}
	}`

	var evt HookEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if evt.Hook != HookPostToolUse {
		t.Errorf("expected hook=%s, got %s", HookPostToolUse, evt.Hook)
	}
	if evt.GetOutput() == nil {
		t.Fatal("expected output to be non-nil")
	}
}

func TestParseNotificationPayload(t *testing.T) {
	raw := `{
		"hook": "notification",
		"session_id": "session-abc-123",
		"title": "Task completed",
		"body": "Successfully fixed the auth bug"
	}`

	var evt HookEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if evt.Hook != HookNotification {
		t.Errorf("expected hook=%s, got %s", HookNotification, evt.Hook)
	}
	if evt.Title != "Task completed" {
		t.Errorf("expected title=Task completed, got %s", evt.Title)
	}
}

func TestParsePayloadWithParentToolUseID(t *testing.T) {
	raw := `{
		"hook": "pre_tool_use",
		"session_id": "session-abc-123",
		"tool_name": "Read",
		"tool_use_id": "toolu_789",
		"parent_tool_use_id": "toolu_parent_123",
		"tool_input": {"file_path": "/tmp/test.go"}
	}`

	var evt HookEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if evt.ParentToolUseID != "toolu_parent_123" {
		t.Errorf("expected parent_tool_use_id=toolu_parent_123, got %s", evt.ParentToolUseID)
	}
}

func TestGetSessionIDFallback(t *testing.T) {
	// Test with "session" field (our jq-added alias)
	evt1 := HookEvent{SessionID: "via-session", SessionIDRaw: "via-session-id"}
	if evt1.GetSessionID() != "via-session" {
		t.Errorf("expected via-session, got %s", evt1.GetSessionID())
	}

	// Test with only "session_id" field (raw Claude Code field)
	evt2 := HookEvent{SessionIDRaw: "via-session-id"}
	if evt2.GetSessionID() != "via-session-id" {
		t.Errorf("expected via-session-id, got %s", evt2.GetSessionID())
	}
}
