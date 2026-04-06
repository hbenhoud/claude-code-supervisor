package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Session struct {
	ID           string  `json:"id"`
	CWD          string  `json:"cwd,omitempty"`
	Status       string  `json:"status"`
	StartedAt    int64   `json:"started_at"`
	FinishedAt   *int64  `json:"finished_at,omitempty"`
	ToolCount    int     `json:"tool_count"`
	AgentCount   int     `json:"agent_count"`
	Prompt       *string `json:"prompt,omitempty"`
	Model        *string `json:"model,omitempty"`
	TotalCostUSD *float64 `json:"total_cost_usd,omitempty"`
	TotalTokens  *int    `json:"total_tokens,omitempty"`
	NumTurns     *int    `json:"num_turns,omitempty"`
}

func (db *DB) EnsureSession(sessionID, cwd string) {
	_, err := db.conn.Exec(
		`INSERT OR IGNORE INTO sessions (id, cwd, status, started_at) VALUES (?, ?, 'running', ?)`,
		sessionID, cwd, time.Now().Unix(),
	)
	if err != nil {
		// Log but don't fail — session may already exist
		return
	}
}

func (db *DB) ListSessions() ([]Session, error) {
	rows, err := db.conn.Query(
		`SELECT id, cwd, status, started_at, finished_at, tool_count, agent_count FROM sessions ORDER BY started_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.CWD, &s.Status, &s.StartedAt, &s.FinishedAt, &s.ToolCount, &s.AgentCount); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []Session{}
	}
	return sessions, nil
}

func (db *DB) GetSession(id string) (*Session, error) {
	var s Session
	err := db.conn.QueryRow(
		`SELECT id, cwd, status, started_at, finished_at, tool_count, agent_count, prompt, model, total_cost_usd, total_tokens, num_turns FROM sessions WHERE id = ?`,
		id,
	).Scan(&s.ID, &s.CWD, &s.Status, &s.StartedAt, &s.FinishedAt, &s.ToolCount, &s.AgentCount, &s.Prompt, &s.Model, &s.TotalCostUSD, &s.TotalTokens, &s.NumTurns)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) IncrementToolCount(sessionID string) {
	db.conn.Exec(`UPDATE sessions SET tool_count = tool_count + 1 WHERE id = ?`, sessionID)
}

func (db *DB) IncrementAgentCount(sessionID string) {
	db.conn.Exec(`UPDATE sessions SET agent_count = agent_count + 1 WHERE id = ?`, sessionID)
}
