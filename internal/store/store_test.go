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

func TestANewDatabaseHasChangesAndDecisionsFromOneMigration(t *testing.T) {
	if SchemaVersion() != 4 {
		t.Fatalf("want 4 migrations, have %d", SchemaVersion())
	}
	db, err := Open(filepath.Join(t.TempDir(), "aiport.yaml.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	count(t, db, "changes")
	count(t, db, "decisions")
}
