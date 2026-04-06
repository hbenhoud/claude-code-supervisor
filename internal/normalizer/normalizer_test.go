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
	if evt.AgentID != "root" {
		t.Errorf("expected agent_id=root for Agent spawn, got %s", evt.AgentID)
	}
}

// TestTemporalWindowSingleAgent verifies that tool calls during a single open agent window
// are attributed to that agent.
func TestTemporalWindowSingleAgent(t *testing.T) {
	n := New()

	// 1. Spawn agent
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_aaa",
		Input:        json.RawMessage(`{"description":"explore"}`),
		Timestamp:    time.Now(),
	})

	// 2. Tool call during agent window → should be attributed to agent
	evt := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_read_001",
		Timestamp:    time.Now(),
	})

	expectedAgentID := "agent-toolu_agent_aaa" // "agent-" + first 16 chars of "toolu_agent_aaa"
	if evt.AgentID != expectedAgentID {
		t.Errorf("expected agent_id=%s, got %s", expectedAgentID, evt.AgentID)
	}
	if evt.ParentAgentID != "root" {
		t.Errorf("expected parent_agent_id=root, got %s", evt.ParentAgentID)
	}

	// 3. Complete agent
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPostToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_aaa",
		Timestamp:    time.Now(),
	})

	// 4. Tool call after agent window → should be root
	evt2 := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_read_002",
		Timestamp:    time.Now(),
	})

	if evt2.AgentID != "root" {
		t.Errorf("expected agent_id=root after window closed, got %s", evt2.AgentID)
	}
}

// TestTemporalWindowParallelAgents verifies that tool calls during multiple open agent windows
// are attributed to root (ambiguous).
func TestTemporalWindowParallelAgents(t *testing.T) {
	n := New()

	// Spawn two agents in parallel
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_aaa",
		Input:        json.RawMessage(`{"description":"explore"}`),
		Timestamp:    time.Now(),
	})
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_bbb",
		Input:        json.RawMessage(`{"description":"plan"}`),
		Timestamp:    time.Now(),
	})

	// Tool call with 2 agents open → should stay root
	evt := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_read_001",
		Timestamp:    time.Now(),
	})

	if evt.AgentID != "root" {
		t.Errorf("expected agent_id=root with parallel agents, got %s", evt.AgentID)
	}

	// Complete one agent → only one left → attributed to remaining
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPostToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_aaa",
		Timestamp:    time.Now(),
	})

	evt2 := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_read_002",
		Timestamp:    time.Now(),
	})

	expectedAgentID := "agent-toolu_agent_bbb" // "agent-" + first 16 chars of "toolu_agent_bbb"
	if evt2.AgentID != expectedAgentID {
		t.Errorf("expected agent_id=%s after one agent closed, got %s", expectedAgentID, evt2.AgentID)
	}
}

// TestTemporalWindowNoAgent verifies that tool calls without any open agent are attributed to root.
func TestTemporalWindowNoAgent(t *testing.T) {
	n := New()

	evt := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Bash",
		ToolUseID:    "toolu_bash_001",
		Timestamp:    time.Now(),
	})

	if evt.AgentID != "root" {
		t.Errorf("expected agent_id=root with no agents, got %s", evt.AgentID)
	}
	if evt.ParentAgentID != "" {
		t.Errorf("expected empty parent_agent_id, got %s", evt.ParentAgentID)
	}
}

// TestTemporalWindowPostToolUseAlsoAttributed verifies PostToolUse events also get the right agent.
func TestTemporalWindowPostToolUseAlsoAttributed(t *testing.T) {
	n := New()

	// Spawn agent
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_aaa",
		Timestamp:    time.Now(),
	})

	// Pre tool call
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_read_001",
		Timestamp:    time.Now(),
	})

	// Post tool call → should also be attributed to agent
	evt := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPostToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_read_001",
		Output:       json.RawMessage(`{"content":"hello"}`),
		Timestamp:    time.Now(),
	})

	if evt.AgentID == "root" {
		t.Error("expected PostToolUse to be attributed to sub-agent, got root")
	}
	if evt.ParentAgentID != "root" {
		t.Errorf("expected parent_agent_id=root, got %s", evt.ParentAgentID)
	}
}

func TestNormalizeSubAgentToolCall(t *testing.T) {
	n := New()

	// When parent_tool_use_id IS set (future-proofing), it should still work via resolve
	// but since parent_tool_use_id isn't sent, this tests the temporal window path
	n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Agent",
		ToolUseID:    "toolu_agent_123",
		Input:        json.RawMessage(`{"description":"explore"}`),
		Timestamp:    time.Now(),
	})

	evt := n.Normalize(hooks.HookEvent{
		Hook:         hooks.HookPreToolUse,
		SessionIDRaw: "sess-1",
		ToolNameRaw:  "Read",
		ToolUseID:    "toolu_sub_read",
		Input:        json.RawMessage(`{"file_path":"/tmp/test.go"}`),
		Timestamp:    time.Now(),
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
