package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hbenhoud/claude-code-supervisor/internal/normalizer"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenCreatesDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected DB file to be created")
	}
}

func TestWALMode(t *testing.T) {
	db := testDB(t)

	var mode string
	err := db.Conn().QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("expected journal_mode=wal, got %s", mode)
	}
}

func TestEnsureSessionAndList(t *testing.T) {
	db := testDB(t)

	db.EnsureSession("sess-1", "/tmp/project")
	db.EnsureSession("sess-1", "/tmp/project") // idempotent

	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "sess-1" {
		t.Errorf("expected id=sess-1, got %s", sessions[0].ID)
	}
	if sessions[0].Status != "running" {
		t.Errorf("expected status=running, got %s", sessions[0].Status)
	}
}

func TestGetSession(t *testing.T) {
	db := testDB(t)

	db.EnsureSession("sess-1", "/tmp/project")

	s, err := db.GetSession("sess-1")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if s.ID != "sess-1" {
		t.Errorf("expected id=sess-1, got %s", s.ID)
	}

	_, err = db.GetSession("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestInsertAndGetEvents(t *testing.T) {
	db := testDB(t)

	db.EnsureSession("sess-1", "/tmp")

	evt := &normalizer.SupervisorEvent{
		ID:           "evt-1",
		SessionID:    "sess-1",
		Sequence:     1,
		EventType:    "tool_call",
		EventSubtype: "start",
		AgentID:      "root",
		ToolName:     "Read",
		ToolInput:    json.RawMessage(`{"file_path":"/tmp/test.go"}`),
	}
	if err := db.InsertEvent(evt); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	events, err := db.GetEventsBySession("sess-1", 0)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ToolName != "Read" {
		t.Errorf("expected tool_name=Read, got %s", events[0].ToolName)
	}
}

func TestGetEventsAfterSequence(t *testing.T) {
	db := testDB(t)
	db.EnsureSession("sess-1", "/tmp")

	for i := 1; i <= 5; i++ {
		db.InsertEvent(&normalizer.SupervisorEvent{
			ID:        fmt.Sprintf("evt-%d", i),
			SessionID: "sess-1",
			Sequence:  i,
			EventType: "tool_call",
			AgentID:   "root",
		})
	}

	events, err := db.GetEventsBySession("sess-1", 3)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events (seq 4,5), got %d", len(events))
	}
}

func TestConcurrentWrites(t *testing.T) {
	db := testDB(t)
	db.EnsureSession("sess-1", "/tmp")

	var wg sync.WaitGroup
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db.InsertEvent(&normalizer.SupervisorEvent{
				ID:        fmt.Sprintf("evt-%d", i),
				SessionID: "sess-1",
				Sequence:  i,
				EventType: "tool_call",
				AgentID:   "root",
			})
		}(i)
	}
	wg.Wait()

	events, err := db.GetEventsBySession("sess-1", 0)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if len(events) != 100 {
		t.Errorf("expected 100 events, got %d", len(events))
	}
}
