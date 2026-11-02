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

// rejectServer is a pairing server over the configuration at path, so a test
// can start it again over the same file and database.
func rejectServer(t *testing.T, path string) (*Handler, string, *time.Time) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		if err := os.WriteFile(path, []byte("version: 1\nauth: {token_env: PAIR_MGMT_TOKEN}\ntargets:\n  - id: local\n    provider: ollama\ndefaults: {openai_target: local}\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	management := strings.Repeat("m", 40)
	t.Setenv("PAIR_MGMT_TOKEN", management)
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	t.Cleanup(h.Close)
	now := time.UnixMilli(1_800_000_000_000)
	h.now = func() time.Time { return now }
	return h, management, &now
}

func rejectedList(t *testing.T, h *Handler, management string) []map[string]any {
	t.Helper()
	w, out := call(h, "GET", "/api/v1/pairings?status=rejected", "", management, false)
	if w.Code != 200 {
		t.Fatalf("rejected list: %d %s", w.Code, w.Body)
	}
	raw, _ := out["pairings"].([]any)
	list := []map[string]any{}
	for _, x := range raw {
		list = append(list, x.(map[string]any))
	}
	return list
}

func reject(t *testing.T, h *Handler, management, name string) map[string]any {
	t.Helper()
	w, a := call(h, "POST", PairPath, `{"name":"`+name+`","description":"about `+name+`"}`, "", false)
	if w.Code != 202 {
		t.Fatalf("pair %s: %d %s", name, w.Code, w.Body)
	}
	if w, _ := call(h, "POST", "/api/v1/pairings/"+a["code"].(string)+"/reject", "", management, true); w.Code != 200 {
		t.Fatalf("reject %s: %d %s", name, w.Code, w.Body)
	}
	return a
}

// A rejected Application is remembered: after it collected its answer,
// after the pending window and after a restart, the Owner still sees it.
func TestARejectedApplicationIsRemembered(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management, now := rejectServer(t, path)
	a := reject(t, h, management, "Nightly Report Job")
	code := a["code"].(string)
	call(h, "GET", a["poll"].(string), "", "", false) // the Application collects its answer
	*now = now.Add(time.Hour)

	list := rejectedList(t, h, management)
	if len(list) != 1 || list[0]["name"] != "Nightly Report Job" || list[0]["description"] != "about Nightly Report Job" ||
		list[0]["code"] != code || list[0]["status"] != "rejected" || list[0]["rejected_at_ms"] != float64(1_800_000_000_000) {
		t.Fatalf("rejected list: %v", list)
	}
	if _, pending := call(h, "GET", "/api/v1/pairings", "", management, false); len(pending["pairings"].([]any)) != 0 {
		t.Fatalf("a rejected application is listed as pending: %v", pending)
	}

	h.Close()
	again, _, _ := rejectServer(t, path)
	if list := rejectedList(t, again, management); len(list) != 1 || list[0]["code"] != code {
		t.Fatalf("forgotten after a restart: %v", list)
	}
}

// The list of rejected Applications is management only and carries no
// secrets; at most the last 100 are kept, newest first.
func TestRejectedApplicationsAreManagementOnlyAndBounded(t *testing.T) {
	h, management, now := rejectServer(t, filepath.Join(t.TempDir(), "aiport.yaml"))
	var secrets []string
	for i := 0; i < 103; i++ {
		*now = now.Add(5 * time.Second) // stays under the pairing rate limit
		a := reject(t, h, management, "App "+itoa(i))
		secrets = append(secrets, strings.TrimPrefix(a["poll"].(string), PairPath+"/"))
	}
	list := rejectedList(t, h, management)
	if len(list) != 100 || list[0]["name"] != "App 102" || list[99]["name"] != "App 3" {
		t.Fatalf("kept %d; first %v last %v", len(list), list[0], list[len(list)-1])
	}
	w, _ := call(h, "GET", "/api/v1/pairings?status=rejected", "", management, false)
	for _, s := range secrets {
		if strings.Contains(w.Body.String(), s) {
			t.Fatal("the rejected list carries a poll secret")
		}
	}
	if w, _ := call(h, "GET", "/api/v1/pairings?status=rejected", "", "", false); w.Code != 401 {
		t.Fatal("rejected list visible without management access")
	}
	if w, _ := call(h, "GET", "/api/v1/pairings?status=other", "", management, false); w.Code != 400 {
		t.Fatalf("unknown status accepted: %d", w.Code)
	}
}

// The Owner may forget a rejected Application (same-origin dashboard only).
func TestTheOwnerCanForgetARejectedApplication(t *testing.T) {
	h, management, _ := rejectServer(t, filepath.Join(t.TempDir(), "aiport.yaml"))
	reject(t, h, management, "Keep")
	reject(t, h, management, "Forget")
	id := ""
	for _, x := range rejectedList(t, h, management) {
		if x["name"] == "Forget" {
			id, _ = x["id"].(string)
		}
	}
	if id == "" {
		t.Fatal("no id to forget a rejected application by")
	}
	if w, _ := call(h, "DELETE", "/api/v1/pairings/rejected/"+id, "", management, false); w.Code != 403 {
		t.Fatal("forgot without CSRF header")
	}
	if w, _ := call(h, "DELETE", "/api/v1/pairings/rejected/"+id, "", management, true); w.Code != 200 {
		t.Fatalf("forget: %d %s", w.Code, w.Body)
	}
	if list := rejectedList(t, h, management); len(list) != 1 || list[0]["name"] != "Keep" {
		t.Fatalf("after forget: %v", list)
	}
	if w, _ := call(h, "DELETE", "/api/v1/pairings/rejected/"+id, "", management, true); w.Code != 404 {
		t.Fatalf("forgot twice: %d", w.Code)
	}
}

// Without a database (a configuration that is not a file), rejections are
// remembered for this session.
func TestRejectedApplicationsWithoutADatabaseLastTheSession(t *testing.T) {
	h, management, now := pairingServer(t)
	h.Close() // no database from here on
	a := reject(t, h, management, "Session Only")
	call(h, "GET", a["poll"].(string), "", "", false)
	*now = now.Add(time.Hour)
	if list := rejectedList(t, h, management); len(list) != 1 || list[0]["name"] != "Session Only" {
		t.Fatalf("not remembered for the session: %v", list)
	}
}
