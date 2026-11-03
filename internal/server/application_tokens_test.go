package server

import (
	"strings"
	"testing"
)

// Issuing, adding and removing Applications from the dashboard, over HTTP:
// the Owner adds an Application without pairing and gets its token once,
// issues a new token that replaces the old one at once, and removes an
// Application so its token stops working. Reverting the removal brings the
// Application back without a token until the Owner issues one.

func catalogStatus(h *Handler, token string) int {
	w, _ := call(h, "GET", "/v1/models", "", token, false)
	return w.Code
}

func TestAddApplicationWithoutPairing(t *testing.T) {
	h, management, _ := pairingServer(t)
	if w, _ := call(h, "POST", "/api/v1/origins", `{"name":"Docs Bot"}`, management, false); w.Code != 403 {
		t.Fatalf("add without CSRF: %d", w.Code)
	}
	if w, _ := call(h, "POST", "/api/v1/origins", `{"name":""}`, management, true); w.Code != 400 {
		t.Fatalf("add without a name: %d", w.Code)
	}
	w, out := call(h, "POST", "/api/v1/origins", `{"name":"Docs Bot"}`, management, true)
	token, _ := out["token"].(string)
	if w.Code != 201 || out["client_id"] != "docs-bot" || len(token) < 32 {
		t.Fatalf("add: %d %s", w.Code, w.Body)
	}
	if code := catalogStatus(h, token); code != 200 {
		t.Fatalf("new token not live: %d", code)
	}
	if w, _ := call(h, "POST", "/api/v1/origins", `{"name":"Docs Bot","id":"docs-bot"}`, management, true); w.Code != 409 {
		t.Fatalf("duplicate id: %d", w.Code)
	}
	if w, _ := call(h, "POST", "/api/v1/origins", `{"name":"X"}`, token, true); w.Code != 401 {
		t.Fatalf("an Application token added an Application: %d", w.Code)
	}
}

func TestIssueTokenReplacesTheOldOne(t *testing.T) {
	h, management, _ := pairingServer(t)
	_, out := call(h, "POST", "/api/v1/origins", `{"name":"Docs Bot"}`, management, true)
	old := out["token"].(string)
	if w, _ := call(h, "POST", "/api/v1/origins/docs-bot/token", "", management, false); w.Code != 403 {
		t.Fatalf("issue without CSRF: %d", w.Code)
	}
	if w, _ := call(h, "POST", "/api/v1/origins/nobody/token", "", management, true); w.Code != 404 {
		t.Fatalf("issue for an unknown Application: %d", w.Code)
	}
	w, out := call(h, "POST", "/api/v1/origins/docs-bot/token", "", management, true)
	fresh, _ := out["token"].(string)
	if w.Code != 200 || len(fresh) < 32 || fresh == old {
		t.Fatalf("issue: %d %s", w.Code, w.Body)
	}
	if code := catalogStatus(h, old); code != 401 {
		t.Fatalf("old token still works: %d", code)
	}
	if code := catalogStatus(h, fresh); code != 200 {
		t.Fatalf("new token not live: %d", code)
	}
	if n := len(h.live.Changes()); n != 1 {
		t.Fatalf("issuing a token recorded a configuration Change: %d Changes", n)
	}
}

func TestIssueTokenRefusesATokenFromTheEnvironment(t *testing.T) {
	h, management, _ := pairingServer(t)
	call(h, "POST", "/api/v1/origins", `{"name":"Docs Bot"}`, management, true)
	external := strings.Repeat("e", 40)
	t.Setenv("AIPORT_APP_DOCS_BOT_TOKEN", external)
	h.resolveTokens(h.snapshot(), true)
	if w, out := call(h, "POST", "/api/v1/origins/docs-bot/token", "", management, true); w.Code != 409 || !strings.Contains(w.Body.String(), "environment") {
		t.Fatalf("replaced a token set in aiport's environment: %d %v", w.Code, out)
	}
	if w, _ := call(h, "DELETE", "/api/v1/origins/docs-bot", "", management, true); w.Code != 409 {
		t.Fatalf("removed an Application whose token is in aiport's environment: %d", w.Code)
	}
	if code := catalogStatus(h, external); code != 200 {
		t.Fatalf("environment token no longer works: %d", code)
	}
}

func TestRemoveApplicationAndRevert(t *testing.T) {
	h, management, _ := pairingServer(t)
	_, out := call(h, "POST", "/api/v1/origins", `{"name":"Docs Bot","target":"local"}`, management, true)
	token := out["token"].(string)
	if w, _ := call(h, "DELETE", "/api/v1/origins/docs-bot", "", management, false); w.Code != 403 {
		t.Fatalf("remove without CSRF: %d", w.Code)
	}
	if w, _ := call(h, "DELETE", "/api/v1/origins/nobody", "", management, true); w.Code != 404 {
		t.Fatalf("remove an unknown Application: %d", w.Code)
	}
	w, out := call(h, "DELETE", "/api/v1/origins/docs-bot", "", management, true)
	if w.Code != 200 || out["removed"] != "docs-bot" {
		t.Fatalf("remove: %d %s", w.Code, w.Body)
	}
	if _, ok := h.snapshot().Application("docs-bot"); ok {
		t.Fatal("Application still configured")
	}
	if code := catalogStatus(h, token); code != 401 {
		t.Fatalf("removed Application's token still works: %d", code)
	}
	// Revert: the Application is back, without a token, and listed as needing one.
	if w, _ := call(h, "POST", "/api/v1/changes/"+out["change_id"].(string)+"/revert", "", management, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	if _, ok := h.snapshot().Application("docs-bot"); !ok {
		t.Fatal("revert did not bring the Application back")
	}
	if code := catalogStatus(h, token); code != 401 {
		t.Fatalf("deleted token works again after the revert: %d", code)
	}
	_, cfg := call(h, "GET", "/api/v1/config", "", management, false)
	if missing, _ := cfg["missing_tokens"].([]any); len(missing) != 1 || missing[0] != "docs-bot" {
		t.Fatalf("missing_tokens: %v", cfg["missing_tokens"])
	}
	// Other Applications keep working meanwhile, and a new one can be added.
	_, other := call(h, "POST", "/api/v1/origins", `{"name":"Nightly"}`, management, true)
	if code := catalogStatus(h, other["token"].(string)); code != 200 {
		t.Fatalf("a new Application does not work while another has no token: %d", code)
	}
	_, out = call(h, "POST", "/api/v1/origins/docs-bot/token", "", management, true)
	if code := catalogStatus(h, out["token"].(string)); code != 200 {
		t.Fatalf("issued token after the revert does not work: %d", code)
	}
	_, cfg = call(h, "GET", "/api/v1/config", "", management, false)
	if missing, _ := cfg["missing_tokens"].([]any); len(missing) != 0 {
		t.Fatalf("still missing after issuing: %v", missing)
	}
}

// Pairing dedup: a second request from the same application replaces the
// waiting one, and a request from an Application that is already configured
// is approved as that Application (a new token) instead of a duplicate.
func TestPairingDedup(t *testing.T) {
	h, management, _ := pairingServer(t)
	_, first := call(h, "POST", PairPath, `{"name":"Docs Bot"}`, "", false)
	_, second := call(h, "POST", PairPath, `{"name":"docs bot!"}`, "", false)
	if _, st := call(h, "GET", first["poll"].(string), "", "", false); st["status"] != "superseded" || st["token"] != nil {
		t.Fatalf("first request not superseded: %v", st)
	}
	_, list := call(h, "GET", "/api/v1/pairings", "", management, false)
	if waiting := list["pairings"].([]any); len(waiting) != 1 || waiting[0].(map[string]any)["code"] != second["code"] {
		t.Fatalf("waiting after a repeat request: %v", list)
	}
	_, other := call(h, "POST", PairPath, `{"name":"Nightly"}`, "", false)
	if _, list = call(h, "GET", "/api/v1/pairings", "", management, false); len(list["pairings"].([]any)) != 2 {
		t.Fatalf("a different application was deduplicated: %v", list)
	}
	call(h, "POST", "/api/v1/pairings/"+other["code"].(string)+"/reject", "", management, true)

	// Approve the first; the same application asks again later.
	call(h, "POST", "/api/v1/pairings/"+second["code"].(string)+"/approve", `{}`, management, true)
	_, got := call(h, "GET", second["poll"].(string), "", "", false)
	old := got["token"].(string)
	_, again := call(h, "POST", PairPath, `{"name":"Docs Bot"}`, "", false)
	_, list = call(h, "GET", "/api/v1/pairings", "", management, false)
	existing, _ := list["pairings"].([]any)[0].(map[string]any)["existing"].(map[string]any)
	if existing["id"] != "docs-bot" || existing["by"] != "name" {
		t.Fatalf("existing Application not named: %v", list)
	}
	w, out := call(h, "POST", "/api/v1/pairings/"+again["code"].(string)+"/approve", `{"replace":"docs-bot"}`, management, true)
	if w.Code != 200 || out["client_id"] != "docs-bot" || out["replaced"] != true {
		t.Fatalf("approve as existing: %d %s", w.Code, w.Body)
	}
	_, got = call(h, "GET", again["poll"].(string), "", "", false)
	fresh, _ := got["token"].(string)
	if got["client_id"] != "docs-bot" || fresh == "" || fresh == old {
		t.Fatalf("poll after replace: %v", got)
	}
	if n := len(h.snapshot().Origins); n != 1 {
		t.Fatalf("a duplicate Application was added: %d Applications", n)
	}
	if catalogStatus(h, old) != 401 || catalogStatus(h, fresh) != 200 {
		t.Fatal("replace did not swap the token")
	}
}
