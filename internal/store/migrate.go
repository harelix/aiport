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

	// 3: Intent Queue tables. The feature lives on the intent-queue branch;
	// the schema stays here because databases already carry version 3.
	// instructions holds the editable description; request is the original
	// wording and is never updated.
	`ALTER TABLE tasks ADD COLUMN workspace TEXT NOT NULL DEFAULT '';
	ALTER TABLE tasks ADD COLUMN request TEXT NOT NULL DEFAULT '';
	ALTER TABLE tasks ADD COLUMN source TEXT NOT NULL DEFAULT 'owner';
	ALTER TABLE tasks ADD COLUMN location TEXT NOT NULL DEFAULT '';
	ALTER TABLE tasks ADD COLUMN position REAL NOT NULL DEFAULT 0;
	ALTER TABLE tasks ADD COLUMN superseded_by TEXT NOT NULL DEFAULT '';
	ALTER TABLE tasks ADD COLUMN execution_ref TEXT NOT NULL DEFAULT '';
	ALTER TABLE tasks ADD COLUMN needs_approval INTEGER NOT NULL DEFAULT 0;
	ALTER TABLE tasks ADD COLUMN validation TEXT NOT NULL DEFAULT '';
	ALTER TABLE tasks ADD COLUMN revision INTEGER NOT NULL DEFAULT 1;
	CREATE INDEX tasks_workspace ON tasks(workspace, position);
	CREATE TABLE task_revisions (
		task_uuid   TEXT NOT NULL REFERENCES tasks(uuid) ON DELETE CASCADE,
		revision    INTEGER NOT NULL,
		at_ms       INTEGER NOT NULL,
		actor       TEXT NOT NULL,
		title       TEXT NOT NULL,
		description TEXT NOT NULL,
		acceptance  TEXT NOT NULL,
		PRIMARY KEY (task_uuid, revision)
	);
	CREATE TABLE task_links (
		task_uuid  TEXT NOT NULL REFERENCES tasks(uuid) ON DELETE CASCADE,
		other_uuid TEXT NOT NULL REFERENCES tasks(uuid) ON DELETE CASCADE,
		kind       TEXT NOT NULL, -- depends_on, related
		PRIMARY KEY (task_uuid, other_uuid, kind)
	);
	CREATE TABLE batches (
		id               TEXT PRIMARY KEY,
		name             TEXT NOT NULL UNIQUE,
		workspace        TEXT NOT NULL,
		title            TEXT NOT NULL DEFAULT '',
		status           TEXT NOT NULL, -- approved, active, planned, completed, released
		created_at_ms    INTEGER NOT NULL,
		updated_at_ms    INTEGER NOT NULL,
		lease_holder     TEXT NOT NULL DEFAULT '',
		lease_expires_ms INTEGER NOT NULL DEFAULT 0,
		summary          TEXT NOT NULL DEFAULT '',
		checks           TEXT NOT NULL DEFAULT ''
	);
	CREATE INDEX batches_workspace ON batches(workspace, status);
	CREATE TABLE batch_tasks (
		batch_id  TEXT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
		task_uuid TEXT NOT NULL REFERENCES tasks(uuid) ON DELETE CASCADE,
		position  INTEGER NOT NULL,
		interrupt INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (batch_id, task_uuid)
	);
	CREATE TABLE batch_history (
		batch_id TEXT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
		seq      INTEGER NOT NULL,
		at_ms    INTEGER NOT NULL,
		actor    TEXT NOT NULL,
		action   TEXT NOT NULL,
		note     TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (batch_id, seq)
	);
	CREATE TABLE proposals (
		id            TEXT PRIMARY KEY,
		name          TEXT NOT NULL UNIQUE,
		workspace     TEXT NOT NULL,
		actor         TEXT NOT NULL,
		status        TEXT NOT NULL, -- open, done, dismissed
		created_at_ms INTEGER NOT NULL,
		updated_at_ms INTEGER NOT NULL,
		fingerprint   TEXT NOT NULL,
		items         TEXT NOT NULL
	);
	CREATE INDEX proposals_workspace ON proposals(workspace, status);`,

	// 4: adapting aiport. Changes: every configuration edit made through
	// aiport, with the exact file before and after so it can be reverted
	// after a restart (kept, no retention). Decisions: Suggestions and
	// Change requests waiting for or given the Owner's decision; the item
	// itself is JSON in body, the columns are what the queue looks items up
	// by. Metadata only, never prompts or responses. IF NOT EXISTS because
	// pre-release databases may already have one of the tables (see
	// preRelease).
	`CREATE TABLE IF NOT EXISTS changes (
		id             TEXT PRIMARY KEY,
		at_ms          INTEGER NOT NULL,
		what           TEXT NOT NULL,
		made_by        TEXT NOT NULL,
		revert_of      TEXT NOT NULL DEFAULT '',
		reverted_by    TEXT NOT NULL DEFAULT '',
		reverted_at_ms INTEGER NOT NULL DEFAULT 0,
		before_config  BLOB NOT NULL,
		after_config   BLOB NOT NULL
	);
	CREATE TABLE IF NOT EXISTS decisions (
		id             TEXT PRIMARY KEY,
		source         TEXT NOT NULL, -- suggestion, request
		kind           TEXT NOT NULL,
		application    TEXT NOT NULL,
		subject        TEXT NOT NULL, -- what repeats of the same item share
		status         TEXT NOT NULL, -- pending, not_supported, accepted, rejected, dismissed
		created_at_ms  INTEGER NOT NULL,
		updated_at_ms  INTEGER NOT NULL,
		body           TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS decisions_subject ON decisions(source, kind, application, subject);`,
}

// preRelease maps the schema version of a database written by a pre-release
// build to this one. Those builds added the changes table as migration 4 and
// the decisions table as migration 5; both are migration 4 now. A version-5
// database has both tables, so it is at 4; a version-4 one may lack
// decisions, so migration 4 (idempotent) runs again. Remove this, and think
// again, before adding a migration 5.
func preRelease(db *sql.DB, current int) (int, error) {
	if len(migrations) != 4 || current < 4 || current > 5 {
		return current, nil
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'decisions'`).Scan(&n); err != nil {
		return current, errors.New("cannot read database tables")
	}
	to := 4
	if n == 0 {
		to = 3
	}
	if _, err := db.Exec(`UPDATE schema_version SET version = ?`, to); err != nil {
		return current, errors.New("cannot update database version")
	}
	return to, nil
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
	if current, err = preRelease(db, current); err != nil {
		return err
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
