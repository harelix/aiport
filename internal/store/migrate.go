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

	// 5: rejected pairing requests, so the Owner still sees an Application
	// they rejected after it collected its answer or aiport restarted: what
	// the Application said about itself and when, never its poll secret or
	// any token. The writer keeps only the newest 100.
	`CREATE TABLE rejected_pairings (
		id             TEXT PRIMARY KEY,
		code           TEXT NOT NULL,
		name           TEXT NOT NULL,
		description    TEXT NOT NULL DEFAULT '',
		rejected_at_ms INTEGER NOT NULL
	);
	CREATE INDEX rejected_pairings_time ON rejected_pairings(rejected_at_ms DESC);`,

	// 6: why a request failed, as fixed identifiers only: aiport's own
	// refusal code, the provider's short error type, and a connection error
	// class. Never message text, prompts, headers or keys. Additive: older
	// rows read as empty.
	`ALTER TABLE requests ADD COLUMN refusal_code TEXT NOT NULL DEFAULT '';
	ALTER TABLE requests ADD COLUMN upstream_error_type TEXT NOT NULL DEFAULT '';
	ALTER TABLE requests ADD COLUMN connection_error TEXT NOT NULL DEFAULT '';`,

	// 7: Issues, problems aiport keeps seeing for one Application, with
	// their evidence, diagnosis and resolutions. The Issue is JSON in body;
	// the columns are what it is looked up by. Application ids, causes,
	// model names, Connection ids, counts, request ids and fixed advice
	// text only: never prompts, responses, provider messages, keys or
	// tokens. Additive.
	`CREATE TABLE issues (
		id            TEXT PRIMARY KEY,
		application   TEXT NOT NULL,
		cause         TEXT NOT NULL, -- unknown_model, failing_connection
		subject       TEXT NOT NULL, -- for unknown_model the requested model, for failing_connection the Connection id
		state         TEXT NOT NULL, -- open, resolving, resolved, dismissed
		created_at_ms INTEGER NOT NULL,
		updated_at_ms INTEGER NOT NULL,
		body          TEXT NOT NULL
	);
	CREATE INDEX issues_key ON issues(application, cause, subject);`,

	// 8: Catalog changes, what one configuration Change (or revert) changed
	// in one Application's Catalog. The change is JSON in body; the columns
	// are what it is looked up and pruned by (kept as long as request
	// history). Application ids, task names, model names and Connection ids
	// only: never keys, URLs, prompts, responses or tokens. Additive.
	`CREATE TABLE catalog_changes (
		change_id   TEXT NOT NULL,
		application TEXT NOT NULL,
		at_ms       INTEGER NOT NULL,
		body        TEXT NOT NULL,
		PRIMARY KEY (change_id, application)
	);
	CREATE INDEX catalog_changes_application ON catalog_changes(application, at_ms DESC);
	CREATE INDEX catalog_changes_time ON catalog_changes(at_ms);`,
	// 9: the guidance version each Application's coding agent last
	// confirmed (aiport_guidance_done) and when. No row = never confirmed.
	// Application ids, version numbers and times only. Additive.
	`CREATE TABLE guidance (
		application     TEXT PRIMARY KEY,
		version         INTEGER NOT NULL,
		confirmed_at_ms INTEGER NOT NULL
	);`,
	// 10: what a model call asked for and what served it: its effort (low,
	// medium, high), where the effort came from (header, field, default)
	// and the model sent upstream. Names only, never body content. The
	// index serves Activity's filter by the name sent. Additive: older rows
	// read as empty.
	`ALTER TABLE requests ADD COLUMN effort TEXT NOT NULL DEFAULT '';
	ALTER TABLE requests ADD COLUMN effort_source TEXT NOT NULL DEFAULT '';
	ALTER TABLE requests ADD COLUMN resolved_model TEXT NOT NULL DEFAULT '';
	CREATE INDEX requests_model ON requests(requested_model, started_at_ms);`,
	// 11: pairing requests, so a restart drops neither a waiting request nor
	// an approved one whose token was not collected yet. The SHA-256 hashes
	// of the poll secrets only (space-separated), never a secret or a token:
	// an approved request's token is read from the credential file when it
	// is collected. Additive.
	`CREATE TABLE pairings (
		code            TEXT PRIMARY KEY,
		name            TEXT NOT NULL,
		description     TEXT NOT NULL DEFAULT '',
		repository      TEXT NOT NULL DEFAULT '',
		requested_at_ms INTEGER NOT NULL,
		expires_at_ms   INTEGER NOT NULL,
		status          TEXT NOT NULL,
		origin_id       TEXT NOT NULL DEFAULT '',
		secret_hashes   TEXT NOT NULL,
		done_at_ms      INTEGER NOT NULL DEFAULT 0
	);`,
	// 12: the section of the calling application a model call was made for
	// (X-Aiport-Task, or the task sent with its feedback), so votes can be
	// read per section. A short identifier only. Additive: older rows read
	// as empty.
	`ALTER TABLE requests ADD COLUMN task TEXT NOT NULL DEFAULT '';
	CREATE INDEX requests_task ON requests(origin_id, task, started_at_ms);`,
}

// preRelease maps the schema version of a database written by a pre-release
// build to the released one. Those builds added the changes table as
// migration 4 and the decisions table as migration 5; both are migration 4
// now. Released databases are told apart by their tables, not the number:
//
//   - version 5 with rejected_pairings is a released version 5: as it is.
//   - version 5 without it is pre-release (changes and decisions): at 4.
//   - version 4 without decisions is pre-release (changes only): at 3, so
//     migration 4 (idempotent) runs again.
//
// A database an earlier build already moved to 4 is a released version-4
// database and simply gains migration 5. Later migrations need no change
// here: only versions 4 and 5 are ever looked at, and a database at 6 or
// later is released (no pre-release build reached 6). A pre-release version
// 5 moved back to 4 then gains migrations 5 and 6 in order.
func preRelease(db *sql.DB, current int) (int, error) {
	if current != 4 && current != 5 {
		return current, nil
	}
	has := func(table string) (bool, error) {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&n); err != nil {
			return false, errors.New("cannot read database tables")
		}
		return n > 0, nil
	}
	to := current
	if current == 5 {
		released, err := has("rejected_pairings")
		if err != nil {
			return current, err
		}
		if !released {
			to = 4
		}
	}
	if to == 4 {
		decisions, err := has("decisions")
		if err != nil {
			return current, err
		}
		if !decisions {
			to = 3
		}
	}
	if to == current {
		return current, nil
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
