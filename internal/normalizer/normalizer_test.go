package normalizer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hbenhoud/claude-code-supervisor/internal/hooks"
)

func TestNormalizePreToolUse(t *testing.T) {
	n := New()

	raw := hooks.HookEvent{
		Hook:      hooks.HookPreToolUse,
		SessionID: "sess-1",
		Tool:      "Read",
		Input:     json.RawMessage(`{"file_path":"/tmp/test.go"}`),
		CWD:       "/home/user",
		Timestamp: time.Now(),
	}

	evt := n.Normalize(raw)

	if evt.EventType != "tool_call" {
		t.Errorf("expected event_type=tool_call, got %s", evt.EventType)
	}
	if evt.EventSubtype != "start" {
		t.Errorf("expected event_subtype=start, got %s", evt.EventSubtype)
	}
	if evt.SessionID != "sess-1" {
		t.Errorf("expected session_id=sess-1, got %s", evt.SessionID)
	}
	if evt.ToolName != "Read" {
		t.Errorf("expected tool_name=Read, got %s", evt.ToolName)
	}
	if evt.Sequence != 1 {
		t.Errorf("expected sequence=1, got %d", evt.Sequence)
	}
	if evt.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestNormalizePostToolUsePairsWithPre(t *testing.T) {
	n := New()

	preTime := time.Now()
	postTime := preTime.Add(150 * time.Millisecond)

	pre := hooks.HookEvent{
		Hook:      hooks.HookPreToolUse,
		SessionID: "sess-1",
		Tool:      "Edit",
		Input:     json.RawMessage(`{"file_path":"/tmp/test.go"}`),
		Timestamp: preTime,
	}
	n.Normalize(pre)

	post := hooks.HookEvent{
		Hook:      hooks.HookPostToolUse,
		SessionID: "sess-1",
		Tool:      "Edit",
		Output:    json.RawMessage(`{"type":"patch"}`),
		Timestamp: postTime,
	}
	evt := n.Normalize(post)

	if evt.EventSubtype != "complete" {
		t.Errorf("expected event_subtype=complete, got %s", evt.EventSubtype)
	}
	if evt.DurationMs == nil {
		t.Fatal("expected duration to be computed")
	}
	if *evt.DurationMs < 100 || *evt.DurationMs > 250 {
		t.Errorf("expected duration ~150ms, got %dms", *evt.DurationMs)
	}
}

func TestNormalizeOrphanPostToolUse(t *testing.T) {
	n := New()

	post := hooks.HookEvent{
		Hook:      hooks.HookPostToolUse,
		SessionID: "sess-1",
		Tool:      "Bash",
		Output:    json.RawMessage(`{"stdout":"hello"}`),
		Timestamp: time.Now(),
	}
	evt := n.Normalize(post)

	if evt.DurationMs != nil {
		t.Error("expected nil duration for orphan PostToolUse")
	}
	if evt.EventSubtype != "complete" {
		t.Errorf("expected event_subtype=complete, got %s", evt.EventSubtype)
	}
}

func TestNormalizeSequenceMonotonic(t *testing.T) {
	n := New()

	for i := 1; i <= 5; i++ {
		evt := n.Normalize(hooks.HookEvent{
			Hook:      hooks.HookPreToolUse,
			SessionID: "sess-1",
			Tool:      "Read",
			Timestamp: time.Now(),
		})
		if evt.Sequence != i {
			t.Errorf("expected sequence=%d, got %d", i, evt.Sequence)
		}
	}
}

func TestNormalizeSequencePerSession(t *testing.T) {
	n := New()

	evt1 := n.Normalize(hooks.HookEvent{
		Hook: hooks.HookPreToolUse, SessionID: "sess-A", Tool: "Read", Timestamp: time.Now(),
	})
	evt2 := n.Normalize(hooks.HookEvent{
		Hook: hooks.HookPreToolUse, SessionID: "sess-B", Tool: "Read", Timestamp: time.Now(),
	})
	evt3 := n.Normalize(hooks.HookEvent{
		Hook: hooks.HookPreToolUse, SessionID: "sess-A", Tool: "Edit", Timestamp: time.Now(),
	})

	if evt1.Sequence != 1 {
		t.Errorf("sess-A first event: expected 1, got %d", evt1.Sequence)
	}
	if evt2.Sequence != 1 {
		t.Errorf("sess-B first event: expected 1, got %d", evt2.Sequence)
	}
	if evt3.Sequence != 2 {
		t.Errorf("sess-A second event: expected 2, got %d", evt3.Sequence)
	}
}

func TestNormalizeSubagentSpawn(t *testing.T) {
	n := New()

	evt := n.Normalize(hooks.HookEvent{
		Hook:          hooks.HookSubagentSpawn,
		SessionID:     "sess-1",
		AgentID:       "agent-xyz",
		ParentAgentID: "root",
		Description:   "explore code",
		Timestamp:     time.Now(),
	})

	if evt.EventType != "agent_spawn" {
		t.Errorf("expected event_type=agent_spawn, got %s", evt.EventType)
	}
	if evt.AgentID != "agent-xyz" {
		t.Errorf("expected agent_id=agent-xyz, got %s", evt.AgentID)
	}
	if evt.ParentAgentID != "root" {
		t.Errorf("expected parent_agent_id=root, got %s", evt.ParentAgentID)
	}
}

func TestNormalizeNotification(t *testing.T) {
	n := New()

	evt := n.Normalize(hooks.HookEvent{
		Hook:      hooks.HookNotification,
		SessionID: "sess-1",
		Title:     "Task completed",
		Body:      "Done",
		Timestamp: time.Now(),
	})

	if evt.EventType != "notification" {
		t.Errorf("expected event_type=notification, got %s", evt.EventType)
	}
	if evt.EventSubtype != "Task completed" {
		t.Errorf("expected event_subtype=Task completed, got %s", evt.EventSubtype)
	}
}
