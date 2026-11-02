package server

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiport/internal/store"
)

// When kept decisions cannot be read, aiport says so in its log (without
// tokens or request text) and keeps deciding for this session.
func TestUnreadableKeptDecisionsAreLogged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(suggestionsYAML), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(store.Path(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO decisions (id, source, kind, application, subject, status, created_at_ms, updated_at_ms, body) VALUES ('x', 'request', 'k', 'nightly-report', 's', 'pending', 1, 1, 'not json: secret request text')`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	var logged bytes.Buffer
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	at := suggestionsNow
	h, management := suggestionsServer(t, path, "", seededHistory(), &at)

	out := logged.String()
	if !strings.Contains(out, "decisions") || !strings.Contains(out, "this session only") {
		t.Fatalf("the failure was not logged: %q", out)
	}
	if strings.Contains(out, "secret request text") || strings.Contains(out, management) {
		t.Fatalf("the log carries more than the failure: %q", out)
	}
	h.Detect(at)
	if on, list := pendingDecisions(t, h, management); !on || len(list) != 1 {
		t.Fatalf("deciding for this session: adapting=%v items=%v", on, list)
	}
}
