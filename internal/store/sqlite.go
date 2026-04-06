package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func Open(dbPath string) (*DB, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot find home directory: %w", err)
		}
		dir := filepath.Join(home, ".claude-supervisor")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("cannot create data directory: %w", err)
		}
		dbPath = filepath.Join(dir, "data.db")
	}

	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}

func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			cwd TEXT,
			status TEXT NOT NULL DEFAULT 'running',
			started_at INTEGER,
			finished_at INTEGER,
			tool_count INTEGER DEFAULT 0,
			agent_count INTEGER DEFAULT 0,
			prompt TEXT,
			model TEXT,
			total_cost_usd REAL,
			total_tokens INTEGER,
			num_turns INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			sequence INTEGER NOT NULL,
			timestamp TEXT NOT NULL,
			event_type TEXT NOT NULL,
			event_subtype TEXT,
			agent_id TEXT,
			parent_agent_id TEXT,
			tool_name TEXT,
			tool_use_id TEXT,
			data JSON NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id, sequence)`,
		`CREATE TABLE IF NOT EXISTS session_scores (
			session_id TEXT PRIMARY KEY,
			score INTEGER,
			grade TEXT,
			error_rate REAL,
			retry_density REAL,
			cost_per_file REAL,
			duration_percentile REAL,
			tool_count INTEGER,
			files_touched INTEGER,
			agent_count INTEGER,
			exploration_ratio REAL,
			duration_ms INTEGER,
			enriched INTEGER DEFAULT 0,
			computed_at TEXT NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS anti_patterns (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			pattern_type TEXT NOT NULL,
			severity TEXT NOT NULL,
			description TEXT,
			recommendation TEXT,
			event_ids JSON,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_anti_patterns_session ON anti_patterns(session_id)`,
		`CREATE TABLE IF NOT EXISTS skill_stats (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			skill_name TEXT NOT NULL,
			error_count INTEGER DEFAULT 0,
			duration_ms INTEGER,
			tool_calls INTEGER,
			tool_breakdown JSON,
			estimated_cost_usd REAL,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_skill_stats_name ON skill_stats(skill_name)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}
	return nil
}
