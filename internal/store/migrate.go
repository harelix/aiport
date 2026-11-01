package store

import (
	"database/sql"
	"errors"
)

// migrations are applied in order, each in one transaction. Never edit a
// released migration; append a new one.
var migrations = []string{
	// 1: request metadata (no prompts or responses, ever).
	`CREATE TABLE requests (
		id              TEXT PRIMARY KEY,
		started_at_ms   INTEGER NOT NULL,
		origin_id       TEXT NOT NULL,
		target_id       TEXT NOT NULL DEFAULT '',
		provider        TEXT NOT NULL DEFAULT '',
		protocol        TEXT NOT NULL DEFAULT '',
		route           TEXT NOT NULL,
		method          TEXT NOT NULL,
		http_status     INTEGER NOT NULL,
		outcome         TEXT NOT NULL,
		latency_ms      INTEGER NOT NULL,
		ttfb_ms         INTEGER,
		requested_model TEXT,
		usage_status    TEXT NOT NULL,
		cost_status     TEXT NOT NULL,
		compression        TEXT NOT NULL DEFAULT '',
		compression_status TEXT NOT NULL DEFAULT '',
		bytes_before       INTEGER NOT NULL DEFAULT 0,
		bytes_after        INTEGER NOT NULL DEFAULT 0,
		feedback           REAL
	);
	CREATE INDEX requests_time ON requests(started_at_ms DESC, id DESC);
	CREATE INDEX requests_origin ON requests(origin_id, started_at_ms);
	CREATE INDEX requests_target ON requests(target_id, started_at_ms);
	CREATE INDEX requests_outcome ON requests(outcome, started_at_ms);
	CREATE INDEX requests_compression ON requests(compression, started_at_ms);`,

	// 2: task queue.
	`CREATE TABLE tasks (
		uuid          TEXT PRIMARY KEY,
		name          TEXT NOT NULL UNIQUE,
		title         TEXT NOT NULL,
		instructions  TEXT NOT NULL DEFAULT '',
		acceptance    TEXT NOT NULL DEFAULT '',
		status        TEXT NOT NULL,
		priority      INTEGER NOT NULL DEFAULT 0,
		created_at_ms INTEGER NOT NULL,
		updated_at_ms INTEGER NOT NULL
	);
	CREATE TABLE task_history (
		task_uuid TEXT NOT NULL REFERENCES tasks(uuid),
		seq       INTEGER NOT NULL,
		at_ms     INTEGER NOT NULL,
		actor     TEXT NOT NULL,
		action    TEXT NOT NULL,
		from_val  TEXT NOT NULL DEFAULT '',
		to_val    TEXT NOT NULL DEFAULT '',
		note      TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (task_uuid, seq)
	);
	CREATE INDEX tasks_created ON tasks(created_at_ms DESC);`,
}

// SchemaVersion is the newest schema this build understands.
func SchemaVersion() int { return len(migrations) }

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`); err != nil {
		return errors.New("cannot initialise database: " + err.Error())
	}
	var current int
	err := db.QueryRow(`SELECT version FROM schema_version`).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err = db.Exec(`INSERT INTO schema_version (version) VALUES (0)`); err != nil {
			return errors.New("cannot initialise database version")
		}
	} else if err != nil {
		return errors.New("cannot read database version")
	}
	if current > len(migrations) {
		return ErrNewerSchema
	}
	for v := current; v < len(migrations); v++ {
		tx, err := db.Begin()
		if err != nil {
			return errors.New("cannot start database migration")
		}
		if _, err = tx.Exec(migrations[v]); err == nil {
			_, err = tx.Exec(`UPDATE schema_version SET version = ?`, v+1)
		}
		if err != nil {
			_ = tx.Rollback()
			return errors.New("database migration failed: " + err.Error())
		}
		if err = tx.Commit(); err != nil {
			return errors.New("cannot commit database migration")
		}
	}
	return nil
}
