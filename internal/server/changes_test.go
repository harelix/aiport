package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

const changesYAML = "version: 1\nauth: {token_env: CHANGES_MGMT_TOKEN}\n# my connections\ntargets:\n  - id: local\n    provider: ollama\ndefaults: {openai_target: local}\n"

// changesServer starts aiport over the configuration file at path (and its
// database next to it), as a fresh process would, with a fixed clock.
func changesServer(t *testing.T, path string, at time.Time) (*Handler, string) {
	t.Helper()
	management := strings.Repeat("m", 40)
	t.Setenv("CHANGES_MGMT_TOKEN", management)
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	h.now = func() time.Time { return at }
	t.Cleanup(h.Close)
	return h, management
}

func changesFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(changesYAML), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func listChanges(t *testing.T, h *Handler, token string) []map[string]any {
	t.Helper()
	w, out := call(h, "GET", "/api/v1/changes", "", token, false)
	if w.Code != 200 {
		t.Fatalf("list changes: %d %s", w.Code, w.Body)
	}
	var list []map[string]any
	for _, c := range out["changes"].([]any) {
		list = append(list, c.(map[string]any))
	}
	return list
}

// A Change applied before a restart is listed and revertable after it, and
// the revert restores the configuration file byte for byte.
func TestChangeRevertAfterRestart(t *testing.T) {
	path := changesFile(t)
	first, management := changesServer(t, path, time.UnixMilli(1_800_000_000_000))
	if w, _ := call(first, "POST", "/api/v1/targets", `{"id":"box","provider":"acme","protocol":"openai","base_url":"http://127.0.0.1:9000/v1"}`, management, true); w.Code != 200 {
		t.Fatalf("add connection: %d %s", w.Code, w.Body)
	}
	first.Close()

	second, management := changesServer(t, path, time.UnixMilli(1_800_000_060_000))
	list := listChanges(t, second, management)
	if len(list) != 1 || list[0]["what"] != "add connection box" || list[0]["by"] != "Owner" || list[0]["reverted"] != false {
		t.Fatalf("change not kept across restart: %v", list)
	}
	if w, _ := call(second, "GET", "/api/v1/changes", "", "", false); w.Code != 401 {
		t.Fatal("changes visible without management access")
	}
	id := list[0]["id"].(string)
	if w, _ := call(second, "POST", "/api/v1/changes/"+id+"/revert", "", management, false); w.Code != 403 {
		t.Fatal("revert without CSRF header")
	}
	if w, _ := call(second, "POST", "/api/v1/changes/"+id+"/revert", "", management, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	got, _ := os.ReadFile(path)
	if string(got) != changesYAML {
		t.Fatalf("revert did not restore the file byte for byte:\n%s", got)
	}
	// Live at once: the reverted connection is gone from the running config.
	for _, target := range second.snapshot().Targets {
		if target.ID == "box" {
			t.Fatal("reverted connection still live")
		}
	}
}

// Revert is refused with an explanation when a newer edit touched the
// configuration, and the file is left as the newer edit wrote it.
func TestRevertRefusedAfterNewerEdit(t *testing.T) {
	path := changesFile(t)
	h, management := changesServer(t, path, time.UnixMilli(1_800_000_000_000))
	if w, _ := call(h, "POST", "/api/v1/targets", `{"id":"box","provider":"acme","protocol":"openai","base_url":"http://127.0.0.1:9000/v1"}`, management, true); w.Code != 200 {
		t.Fatalf("add connection: %d %s", w.Code, w.Body)
	}
	if w, _ := call(h, "PUT", "/api/v1/defaults", `{"openai_target":"box"}`, management, true); w.Code != 200 {
		t.Fatalf("set defaults: %d %s", w.Code, w.Body)
	}
	newer, _ := os.ReadFile(path)
	list := listChanges(t, h, management)
	if len(list) != 2 || list[0]["what"] != "protocol defaults" {
		t.Fatalf("changes not newest first: %v", list)
	}
	w, out := call(h, "POST", "/api/v1/changes/"+list[1]["id"].(string)+"/revert", "", management, true)
	failure, _ := out["error"].(map[string]any)
	if w.Code != 409 || failure["code"] != "revert_conflict" || !strings.Contains(failure["message"].(string), "edited after this change") {
		t.Fatalf("conflicting revert not refused clearly: %d %s", w.Code, w.Body)
	}
	if got, _ := os.ReadFile(path); string(got) != string(newer) {
		t.Fatal("refused revert changed the file")
	}
}

// Reverted Changes stay listed and marked as reverted after a restart, next to
// the revert itself, and cannot be reverted twice.
func TestRevertedChangesMarkedAcrossRestart(t *testing.T) {
	path := changesFile(t)
	first, management := changesServer(t, path, time.UnixMilli(1_800_000_000_000))
	if w, _ := call(first, "POST", "/api/v1/targets", `{"id":"box","provider":"acme","protocol":"openai","base_url":"http://127.0.0.1:9000/v1"}`, management, true); w.Code != 200 {
		t.Fatalf("add connection: %d %s", w.Code, w.Body)
	}
	id := listChanges(t, first, management)[0]["id"].(string)
	if w, _ := call(first, "POST", "/api/v1/changes/"+id+"/revert", "", management, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	first.Close()

	second, management := changesServer(t, path, time.UnixMilli(1_800_000_060_000))
	list := listChanges(t, second, management)
	if len(list) != 2 {
		t.Fatalf("want the change and its revert, got %v", list)
	}
	revert, original := list[0], list[1]
	if revert["what"] != "revert: add connection box" || revert["revert_of"] != id || revert["by"] != "Owner" {
		t.Fatalf("revert not listed: %v", revert)
	}
	if original["id"] != id || original["reverted"] != true || original["reverted_by"] != revert["id"] || original["reverted_at_ms"] != float64(1_800_000_000_000) {
		t.Fatalf("reverted change not marked: %v", original)
	}
	w, out := call(second, "POST", "/api/v1/changes/"+id+"/revert", "", management, true)
	if failure, _ := out["error"].(map[string]any); w.Code != 409 || failure["code"] != "already_reverted" {
		t.Fatalf("second revert not refused: %d %s", w.Code, w.Body)
	}
	if w, _ := call(second, "POST", "/api/v1/changes/nope/revert", "", management, true); w.Code != 404 {
		t.Fatalf("unknown change: %d", w.Code)
	}
}
