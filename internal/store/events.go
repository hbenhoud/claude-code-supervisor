package store

import (
	"encoding/json"

	"github.com/hbenhoud/claude-code-supervisor/internal/normalizer"
)

func (db *DB) InsertEvent(evt *normalizer.SupervisorEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, err = db.conn.Exec(
		`INSERT INTO events (id, session_id, sequence, timestamp, event_type, event_subtype, agent_id, parent_agent_id, tool_name, tool_use_id, data)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		evt.ID, evt.SessionID, evt.Sequence, evt.Timestamp.Format("2006-01-02T15:04:05.000Z"),
		evt.EventType, evt.EventSubtype, evt.AgentID, evt.ParentAgentID,
		evt.ToolName, evt.ToolUseID, string(data),
	)
	return err
}

func (db *DB) GetEventsBySession(sessionID string, afterSequence int) ([]*normalizer.SupervisorEvent, error) {
	rows, err := db.conn.Query(
		`SELECT data FROM events WHERE session_id = ? AND sequence > ? ORDER BY sequence ASC`,
		sessionID, afterSequence,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*normalizer.SupervisorEvent
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var evt normalizer.SupervisorEvent
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}
		events = append(events, &evt)
	}
	if events == nil {
		events = []*normalizer.SupervisorEvent{}
	}
	return events, nil
}
