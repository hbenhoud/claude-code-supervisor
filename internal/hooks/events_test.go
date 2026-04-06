package hooks

import (
	"encoding/json"
	"testing"
)

func TestParsePreToolUsePayload(t *testing.T) {
	raw := `{
		"hook": "pre_tool_use",
		"session": "session-abc-123",
		"tool": "Read",
		"input": {"file_path": "/tmp/test.go"},
		"cwd": "/home/user/project"
	}`

	var evt HookEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if evt.Hook != HookPreToolUse {
		t.Errorf("expected hook=%s, got %s", HookPreToolUse, evt.Hook)
	}
	if evt.SessionID != "session-abc-123" {
		t.Errorf("expected session=session-abc-123, got %s", evt.SessionID)
	}
	if evt.Tool != "Read" {
		t.Errorf("expected tool=Read, got %s", evt.Tool)
	}
	if evt.CWD != "/home/user/project" {
		t.Errorf("expected cwd=/home/user/project, got %s", evt.CWD)
	}
	if evt.Input == nil {
		t.Fatal("expected input to be non-nil")
	}

	var input map[string]string
	if err := json.Unmarshal(evt.Input, &input); err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}
	if input["file_path"] != "/tmp/test.go" {
		t.Errorf("expected file_path=/tmp/test.go, got %s", input["file_path"])
	}
}

func TestParsePostToolUsePayload(t *testing.T) {
	raw := `{
		"hook": "post_tool_use",
		"session": "session-abc-123",
		"tool": "Edit",
		"input": {"file_path": "/tmp/test.go", "old_string": "foo", "new_string": "bar"},
		"output": {"type": "patch", "file": {"filePath": "/tmp/test.go", "numLines": 42}}
	}`

	var evt HookEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if evt.Hook != HookPostToolUse {
		t.Errorf("expected hook=%s, got %s", HookPostToolUse, evt.Hook)
	}
	if evt.Output == nil {
		t.Fatal("expected output to be non-nil")
	}
}

func TestParseNotificationPayload(t *testing.T) {
	raw := `{
		"hook": "notification",
		"session": "session-abc-123",
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
	if evt.Body != "Successfully fixed the auth bug" {
		t.Errorf("expected body, got %s", evt.Body)
	}
}

func TestParseSubagentSpawnPayload(t *testing.T) {
	raw := `{
		"hook": "subagent_spawn",
		"session": "session-abc-123",
		"agent_id": "agent-xyz",
		"parent_id": "root",
		"description": "Explore codebase for auth patterns"
	}`

	var evt HookEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if evt.Hook != HookSubagentSpawn {
		t.Errorf("expected hook=%s, got %s", HookSubagentSpawn, evt.Hook)
	}
	if evt.AgentID != "agent-xyz" {
		t.Errorf("expected agent_id=agent-xyz, got %s", evt.AgentID)
	}
	if evt.ParentAgentID != "root" {
		t.Errorf("expected parent_id=root, got %s", evt.ParentAgentID)
	}
	if evt.Description != "Explore codebase for auth patterns" {
		t.Errorf("expected description, got %s", evt.Description)
	}
}
