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
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_abc",
		Input:        json.RawMessage(`{"file_path":"/tmp/test.go"}`),
		Timestamp:    time.Now(),
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
	if evt.AgentID != "root" {
		t.Errorf("expected agent_id=root, got %s", evt.AgentID)
	}
	if evt.Sequence != 1 {
		t.Errorf("expected sequence=1, got %d", evt.Sequence)
	}
}

func TestNormalizePostToolUsePairsWithPre(t *testing.T) {
	n := New()

	preTime := time.Now()
	postTime := preTime.Add(150 * time.Millisecond)

	pre := hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Edit",
		ToolUseID:    "toolu_xyz",
		Input:        json.RawMessage(`{"file_path":"/tmp/test.go"}`),
		Timestamp:    preTime,
	}
	n.Normalize(pre)

	post := hooks.HookEvent{
		Hook:         hooks.HookPostToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Edit",
		ToolUseID:    "toolu_xyz",
		Output:       json.RawMessage(`{"type":"patch"}`),
		Timestamp:    postTime,
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
		Hook:         hooks.HookPostToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Bash",
		ToolUseID:    "toolu_orphan",
		Output:       json.RawMessage(`{"stdout":"hello"}`),
		Timestamp:    time.Now(),
	}
	evt := n.Normalize(post)

	if evt.DurationMs != nil {
		t.Error("expected nil duration for orphan PostToolUse")
	}
}

func TestNormalizeSequenceMonotonic(t *testing.T) {
	n := New()

	for i := 1; i <= 5; i++ {
		evt := n.Normalize(hooks.HookEvent{
			Hook:         hooks.HookPreToolUse,
			SessionIDRaw: "sess-1",
			ToolNameRaw:  "Read",
			Timestamp:    time.Now(),
		})
		if evt.Sequence != i {
			t.Errorf("expected sequence=%d, got %d", i, evt.Sequence)
		}
	}
}

func TestNormalizeSequencePerSession(t *testing.T) {
	n := New()

	evt1 := n.Normalize(hooks.HookEvent{
		Hook: hooks.HookPreToolUse, SessionIDRaw: "sess-A", ToolNameRaw: "Read", Timestamp: time.Now(),
	})
	evt2 := n.Normalize(hooks.HookEvent{
		Hook: hooks.HookPreToolUse, SessionIDRaw: "sess-B", ToolNameRaw: "Read", Timestamp: time.Now(),
	})
	evt3 := n.Normalize(hooks.HookEvent{
		Hook: hooks.HookPreToolUse, SessionIDRaw: "sess-A", ToolNameRaw: "Edit", Timestamp: time.Now(),
	})

	if evt1.Sequence != 1 {
		t.Errorf("sess-A first: expected 1, got %d", evt1.Sequence)
	}
	if evt2.Sequence != 1 {
		t.Errorf("sess-B first: expected 1, got %d", evt2.Sequence)
	}
	if evt3.Sequence != 2 {
		t.Errorf("sess-A second: expected 2, got %d", evt3.Sequence)
	}
}

func TestNormalizeAgentToolCallCreatesAgentSpawn(t *testing.T) {
	n := New()

	evt := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_123",
		Input:        json.RawMessage(`{"description":"explore code","subagent_type":"Explore"}`),
		Timestamp:    time.Now(),
	})

	if evt.EventType != "agent_spawn" {
		t.Errorf("expected event_type=agent_spawn, got %s", evt.EventType)
	}
	if evt.ToolName != "Agent" {
		t.Errorf("expected tool_name=Agent, got %s", evt.ToolName)
	}
}

func TestNormalizeSubAgentToolCall(t *testing.T) {
	n := New()

	evt := n.Normalize(hooks.HookEvent{
		Hook:            hooks.HookPreToolUse,
		SessionIDRaw:    "sess-1",
		ToolNameRaw:     "Read",
		ToolUseID:       "toolu_sub_read",
		ParentToolUseID: "toolu_agent_123",
		Input:           json.RawMessage(`{"file_path":"/tmp/test.go"}`),
		Timestamp:       time.Now(),
	})

	if evt.AgentID == "root" {
		t.Error("expected agent_id to NOT be root for sub-agent tool call")
	}
	if evt.ParentAgentID == "" {
		t.Error("expected parent_agent_id to be set")
	}
}

func TestNormalizeNotification(t *testing.T) {
	n := New()

	evt := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookNotification,
		SessionIDRaw: "sess-1",
		Title:        "Task completed",
		Timestamp:    time.Now(),
	})

	if evt.EventType != "notification" {
		t.Errorf("expected event_type=notification, got %s", evt.EventType)
	}
	if evt.EventSubtype != "Task completed" {
		t.Errorf("expected event_subtype=Task completed, got %s", evt.EventSubtype)
	}
}
