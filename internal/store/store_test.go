package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// Before release, the changes and decisions tables were added as two
// migrations, 4 and 5. They are one migration 4 now, and databases written
// by those builds still open with everything they kept.

const preReleaseChanges = `CREATE TABLE changes (
	id TEXT PRIMARY KEY, at_ms INTEGER NOT NULL, what TEXT NOT NULL, made_by TEXT NOT NULL,
	revert_of TEXT NOT NULL DEFAULT '', reverted_by TEXT NOT NULL DEFAULT '', reverted_at_ms INTEGER NOT NULL DEFAULT 0,
	before_config BLOB NOT NULL, after_config BLOB NOT NULL);`

const preReleaseDecisions = `CREATE TABLE decisions (
	id TEXT PRIMARY KEY, source TEXT NOT NULL, kind TEXT NOT NULL, application TEXT NOT NULL, subject TEXT NOT NULL,
	status TEXT NOT NULL, created_at_ms INTEGER NOT NULL, updated_at_ms INTEGER NOT NULL, body TEXT NOT NULL);
	CREATE INDEX decisions_subject ON decisions(source, kind, application, subject);`

// preReleaseDB writes a database as a pre-release build left it: migrations 1–3,
// then the given statements, at schema version version.
func preReleaseDB(t *testing.T, version int, statements ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml.db")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	all := append(append([]string{`CREATE TABLE schema_version (version INTEGER NOT NULL)`}, migrations[:3]...), statements...)
	for _, s := range all {
		if _, err := db.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (?)`, version); err != nil {
		t.Fatal(err)
	}
	return path
}

func count(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&n); err != nil {
		t.Fatalf("%s: %v", table, err)
	}
	return n
}

func TestADatabaseFromBeforeTheMigrationsWereFoldedStillOpens(t *testing.T) {
	cases := map[string]struct {
		version int
		tables  []string
		kept    map[string]int
	}{
		"changes and decisions as migrations 4 and 5": {5, []string{preReleaseChanges, preReleaseDecisions,
			`INSERT INTO changes VALUES ('c', 1, 'w', 'Owner', '', '', 0, x'00', x'00')`,
			`INSERT INTO decisions VALUES ('d', 'request', 'k', 'a', 's', 'pending', 1, 1, '{}')`}, map[string]int{"changes": 1, "decisions": 1}},
		"only changes, as migration 4": {4, []string{preReleaseChanges,
			`INSERT INTO changes VALUES ('c', 1, 'w', 'Owner', '', '', 0, x'00', x'00')`}, map[string]int{"changes": 1, "decisions": 0}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			path := preReleaseDB(t, tc.version, tc.tables...)
			db, err := Open(path)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer db.Close()
			for table, n := range tc.kept {
				if got := count(t, db, table); got != n {
					t.Fatalf("%s: %d rows, want %d", table, got, n)
				}
			}
			var v int
			if err := db.QueryRow(`SELECT version FROM schema_version`).Scan(&v); err != nil || v != SchemaVersion() {
				t.Fatalf("schema version %d, want %d (%v)", v, SchemaVersion(), err)
			}
		})
	}
}

func TestANewDatabaseHasEveryTable(t *testing.T) {
	if SchemaVersion() != 10 {
		t.Fatalf("want 10 migrations, have %d", SchemaVersion())
	}
	db, err := Open(filepath.Join(t.TempDir(), "aiport.yaml.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	count(t, db, "changes")
	count(t, db, "decisions")
	count(t, db, "rejected_pairings")
	count(t, db, "issues")
	count(t, db, "catalog_changes")
	count(t, db, "guidance")
}

// A database a released build left at version 6 keeps its rows and gains
// migration 7, the Issues table.
func TestAVersion6DatabaseGainsIssues(t *testing.T) {
	path := atVersion(t, 6,
		`INSERT INTO decisions VALUES ('d', 'suggestion', 'application_alias', 'docs-bot', 'llama3:8b', 'pending', 1, 1, '{}')`,
		`INSERT INTO rejected_pairings (id, code, name, description, rejected_at_ms) VALUES ('r', 'ABC-DEF', 'n', '', 1)`)
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if count(t, db, "decisions") != 1 || count(t, db, "rejected_pairings") != 1 || count(t, db, "issues") != 0 {
		t.Fatal("rows lost or the issues table missing")
	}
	var v int
	if err := db.QueryRow(`SELECT version FROM schema_version`).Scan(&v); err != nil || v != SchemaVersion() {
		t.Fatalf("schema version %d (%v)", v, err)
	}
}

// atVersion writes a database the way this or an earlier release left it:
// the first n migrations applied, at schema version n.
func atVersion(t *testing.T, n int, statements ...string) string {
	t.Helper()
	path := preReleaseDB(t, n, append(append([]string{}, migrations[3:n]...), statements...)...)
	return path
}

// A database a released build left at version 4 (also one the pre-release
// shim moved there) gains the rejected pairings table and keeps its rows.
func TestAVersion4DatabaseUpgrades(t *testing.T) {
	path := atVersion(t, 4,
		`INSERT INTO changes VALUES ('c', 1, 'w', 'Owner', '', '', 0, x'00', x'00')`,
		`INSERT INTO decisions VALUES ('d', 'request', 'k', 'a', 's', 'pending', 1, 1, '{}')`)
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if count(t, db, "changes") != 1 || count(t, db, "decisions") != 1 || count(t, db, "rejected_pairings") != 0 {
		t.Fatal("rows lost or table missing")
	}
}

// A database already at version 5 is not mistaken for a pre-release one: it
// keeps its rows and only gains migration 6, the failure reason columns.
func TestAVersion5DatabaseOpensAsItIs(t *testing.T) {
	path := atVersion(t, 5,
		`INSERT INTO changes VALUES ('c', 1, 'w', 'Owner', '', '', 0, x'00', x'00')`,
		`INSERT INTO rejected_pairings (id, code, name, description, rejected_at_ms) VALUES ('r', 'ABC-DEF', 'n', '', 1)`,
		`INSERT INTO requests (id, started_at_ms, origin_id, route, method, http_status, outcome, latency_ms, usage_status, cost_status) VALUES ('q', 1, 'o', '/v1/messages', 'POST', 503, 'rejected', 1, 'unknown', 'unknown')`)
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if count(t, db, "changes") != 1 || count(t, db, "rejected_pairings") != 1 || count(t, db, "requests") != 1 {
		t.Fatal("rows lost")
	}
	var v int
	if err := db.QueryRow(`SELECT version FROM schema_version`).Scan(&v); err != nil || v != SchemaVersion() {
		t.Fatalf("schema version %d (%v)", v, err)
	}
	var code, kind, conn string
	if err := db.QueryRow(`SELECT refusal_code, upstream_error_type, connection_error FROM requests WHERE id = 'q'`).Scan(&code, &kind, &conn); err != nil || code+kind+conn != "" {
		t.Fatalf("reason columns: %q %q %q (%v)", code, kind, conn, err)
	}
}

func TestRejectionsKeepTheNewestAndCanBeForgotten(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "aiport.yaml.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := Rejections{DB: db}
	for i := 0; i < 5; i++ {
		id := string(rune('a' + i))
		if err := r.Add(Rejection{ID: id, Code: "ABC-DE" + id, Name: "App " + id, Description: "d", RejectedAt: int64(i)}, 3); err != nil {
			t.Fatal(err)
		}
	}
	list, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[0].ID != "e" || list[2].ID != "c" || list[0].Name != "App e" || list[0].Code != "ABC-DEe" || list[0].RejectedAt != 4 {
		t.Fatalf("kept %+v", list)
	}
	if ok, err := r.Forget("d"); !ok || err != nil {
		t.Fatalf("forget: %v %v", ok, err)
	}
	if ok, _ := r.Forget("d"); ok {
		t.Fatal("forgot twice")
	}
	if list, _ := r.List(); len(list) != 2 {
		t.Fatalf("after forget %+v", list)
	}
}
