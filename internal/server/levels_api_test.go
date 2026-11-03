package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAnApplicationsOwnMappingIsACatalogChangeForThatApplicationOnly(t *testing.T) {
	f := levelsServer(t)
	code, out := levelEdit(f, "PUT", "/api/v1/levels/anthropic/best", `{"application":"docs-bot","connection":"claude","model":"claude-opus"}`)
	if code != 200 {
		t.Fatalf("%d %v", code, out)
	}
	changes := catalogChanges(t, catalogOf(t, f.h, lvlDocs))
	if len(changes) != 1 || changes[0]["change_id"] != out["change_id"] {
		t.Fatalf("docs-bot catalog changes %v", changes)
	}
	if other := catalogChanges(t, catalogOf(t, f.h, lvlReview)); len(other) != 0 {
		t.Fatalf("review-code was told of docs-bot's mapping: %v", other)
	}
}

func TestLevelEditsAreRefusedWithoutTheDashboardOrForAMappingThatCannotServe(t *testing.T) {
	f := levelsServer(t)
	before, _ := json.Marshal(f.h.snapshot().Aliases)
	if w, _ := call(f.h, "PUT", "/api/v1/levels/openai/best", `{"connection":"local","model":"x"}`, lvlManagement, false); w.Code != 403 {
		t.Fatalf("without CSRF: %d", w.Code)
	}
	if w, _ := call(f.h, "PUT", "/api/v1/levels/openai/best", `{"connection":"local","model":"x"}`, lvlDocs, true); w.Code != 401 && w.Code != 403 {
		t.Fatalf("with an Application token: %d", w.Code)
	}
	cases := []struct {
		method, path, body string
		code               int
	}{
		{"PUT", "/api/v1/levels/anthropic/best", `{"connection":"hosted","model":"gpt-4o"}`, 400}, // an OpenAI Connection
		{"PUT", "/api/v1/levels/openai/strongest", `{"connection":"hosted","model":"gpt-4o"}`, 400},
		{"PUT", "/api/v1/levels/openai/best", `{"connection":"hosted"}`, 400},
		{"PUT", "/api/v1/levels/openai/best", `{"application":"nobody","connection":"hosted","model":"gpt-4o"}`, 404},
		{"PUT", "/api/v1/levels/openai/best", `{"connection":"hosted","model":"gpt-4o","effort":{"low":"bad\u0007"}}`, 400},
		{"PUT", "/api/v1/aliases/effort", `{"alias":"summarize","connection":"hosted","effort":{"low":"x"}}`, 400}, // no such Alias
		{"PUT", "/api/v1/aliases/effort", `{"alias":"best","effort":{"low":"x"}}`, 400},                            // which Connection?
		{"POST", "/api/v1/levels/openai/best", `{}`, 405},
	}
	for _, c := range cases {
		if code, out := levelEdit(f, c.method, c.path, c.body); code != c.code {
			t.Fatalf("%s %s %s: %d %v, want %d", c.method, c.path, c.body, code, out, c.code)
		}
	}
	if after, _ := json.Marshal(f.h.snapshot().Aliases); string(after) != string(before) {
		t.Fatalf("a refused edit changed the mappings:\n%s\n%s", before, after)
	}
}

// The Owner edits Levels and per-effort models over the management API:
// each edit is one revertable Change.

// levelEdit sends a management request with the dashboard's CSRF header.
func levelEdit(f *levelsFixture, method, path, body string) (int, map[string]any) {
	w, out := call(f.h, method, path, body, lvlManagement, true)
	return w.Code, out
}

// Each edit's Change says what it did, in words, for whom.
func TestLevelAndEffortChangesDescribeTheEdit(t *testing.T) {
	f := levelsServer(t)
	edits := []struct{ method, path, body, want string }{
		{"PUT", "/api/v1/levels/openai/best", `{"connection":"local","model":"llama3:70b"}`, "map Level best on OpenAI to llama3:70b on local for every Application"},
		{"PUT", "/api/v1/levels/anthropic/fast", `{"application":"docs-bot","connection":"claude","model":"claude-haiku","effort":{"high":"claude-sonnet"}}`, "map Level fast on Anthropic to claude-haiku on claude for docs-bot (high claude-sonnet)"},
		{"PUT", "/api/v1/aliases/effort", `{"alias":"best","connection":"local","effort":{"low":"llama3:8b"}}`, "per-effort models of best on local for every Application: low llama3:8b"},
		{"PUT", "/api/v1/aliases/effort", `{"application":"docs-bot","alias":"fast","connection":"claude","effort":{}}`, "per-effort models of fast on claude for docs-bot: none"},
		{"DELETE", "/api/v1/levels/anthropic/fast", `{"application":"docs-bot"}`, "clear Level fast on Anthropic for docs-bot"},
		{"DELETE", "/api/v1/levels/openai/best", ``, "clear Level best on OpenAI for every Application"},
	}
	for _, e := range edits {
		code, out := levelEdit(f, e.method, e.path, e.body)
		if code != 200 {
			t.Fatalf("%s %s %s: %d %v", e.method, e.path, e.body, code, out)
		}
		var what string
		for _, c := range f.h.live.Changes() {
			if c.ID == out["change_id"] {
				what = c.What
			}
		}
		if what != e.want {
			t.Fatalf("%s %s: Change %q, want %q", e.method, e.path, what, e.want)
		}
	}
	if code, _ := levelEdit(f, "DELETE", "/api/v1/aliases/effort", ``); code != 405 {
		t.Fatalf("DELETE on aliases/effort: %d", code)
	}
}

func TestTheOwnerMapsASharedLevelAndRevertRestoresThePreviousMapping(t *testing.T) {
	f := levelsServer(t)
	code, out := levelEdit(f, "PUT", "/api/v1/levels/openai/best", `{"connection":"local","model":"llama3:70b"}`)
	if code != 200 || out["change_id"] == nil {
		t.Fatalf("map best: %d %v", code, out)
	}
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"best"}`); w.Code != 200 {
		t.Fatalf("best after mapping: %d %s", w.Code, w.Body)
	}
	if _, got := f.local.received(); got != `{"model":"llama3:70b"}` {
		t.Fatalf("best reached local as %q", got)
	}
	if code, out := levelEdit(f, "POST", "/api/v1/changes/"+out["change_id"].(string)+"/revert", ""); code != 200 {
		t.Fatalf("revert: %d %v", code, out)
	}
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"best"}`); w.Code != 200 {
		t.Fatalf("best after revert: %d %s", w.Code, w.Body)
	}
	if _, got := f.hosted.received(); got != `{"model":"gpt-4o"}` {
		t.Fatalf("after the revert best reached hosted as %q", got)
	}
}

func TestAnApplicationsOwnMappingSetOverTheAPIComesBeforeTheSharedOneUntilCleared(t *testing.T) {
	f := levelsServer(t)
	if code, out := levelEdit(f, "PUT", "/api/v1/levels/openai/fast", `{"application":"docs-bot","connection":"hosted","model":"gpt-4o-mini"}`); code != 200 {
		t.Fatalf("docs-bot fast: %d %v", code, out)
	}
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"fast"}`); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if _, got := f.hosted.received(); got != `{"model":"gpt-4o-mini"}` {
		t.Fatalf("docs-bot's own fast reached hosted as %q", got)
	}
	// review-code still gets the shared fast.
	if w := levelCall(f.h, "/v1/chat/completions", lvlReview, `{"model":"fast"}`); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if _, got := f.local.received(); got != `{"model":"llama3:8b"}` {
		t.Fatalf("review-code fast reached local as %q", got)
	}
	if code, out := levelEdit(f, "DELETE", "/api/v1/levels/openai/fast", `{"application":"docs-bot"}`); code != 200 {
		t.Fatalf("clear docs-bot fast: %d %v", code, out)
	}
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"fast"}`); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if n, got := f.local.received(); n != 2 || got != `{"model":"llama3:8b"}` {
		t.Fatalf("after clearing, docs-bot fast reached local as %q (%d calls)", got, n)
	}
	// Clearing the shared fast leaves fast unmapped on OpenAI.
	if code, out := levelEdit(f, "DELETE", "/api/v1/levels/openai/fast", ""); code != 200 {
		t.Fatalf("clear shared fast: %d %v", code, out)
	}
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"fast"}`); w.Code != 400 {
		t.Fatalf("fast after clearing both: %d %s", w.Code, w.Body)
	}
}

// sentModel sends model with X-Aiport-Effort effort as docs-bot on OpenAI and
// returns the model hosted received.
func sentModel(t *testing.T, f *effortFixture, model, effort string) string {
	t.Helper()
	if w := effortCall(f.h, "/v1/chat/completions", effort, `{"model":"`+model+`"}`); w.Code != 200 {
		t.Fatalf("%s at %s: %d %s", model, effort, w.Code, w.Body)
	}
	_, got := f.hosted.received()
	return strings.TrimSuffix(strings.TrimPrefix(got, `{"model":"`), `"}`)
}

func effortEdit(f *effortFixture, method, path, body string) (int, map[string]any) {
	w, out := call(f.h, method, path, body, lvlManagement, true)
	return w.Code, out
}

func TestTheOwnerSetsPerEffortModelsOnALevelInTheSameChangeAsItsMapping(t *testing.T) {
	f := effortServer(t)
	code, out := effortEdit(f, "PUT", "/api/v1/levels/openai/best", `{"connection":"hosted","model":"gpt-big","effort":{"low":"gpt-mid"}}`)
	if code != 200 {
		t.Fatalf("map best: %d %v", code, out)
	}
	if got := sentModel(t, f, "best", "low"); got != "gpt-mid" {
		t.Fatalf("best at low effort went as %q", got)
	}
	if got := sentModel(t, f, "best", "high"); got != "gpt-big" {
		t.Fatalf("best at high effort went as %q", got)
	}
	// Changing the model with effort left out keeps the per-effort models.
	if code, out := effortEdit(f, "PUT", "/api/v1/levels/openai/best", `{"connection":"hosted","model":"gpt-max"}`); code != 200 {
		t.Fatalf("change best: %d %v", code, out)
	}
	if got := sentModel(t, f, "best", "low"); got != "gpt-mid" {
		t.Fatalf("best lost its low model: %q", got)
	}
	// An empty effort clears them.
	if code, _ := effortEdit(f, "PUT", "/api/v1/levels/openai/best", `{"connection":"hosted","model":"gpt-max","effort":{}}`); code != 200 {
		t.Fatal("clear per-effort models")
	}
	if got := sentModel(t, f, "best", "low"); got != "gpt-max" {
		t.Fatalf("best at low effort after clearing went as %q", got)
	}
}

func TestTheOwnerSetsPerEffortModelsOnTaskNamesAndRevertsThem(t *testing.T) {
	f := effortServer(t)
	// docs-bot's own task name, found by its name alone.
	code, out := effortEdit(f, "PUT", "/api/v1/aliases/effort", `{"application":"docs-bot","alias":"review","effort":{"low":"gpt-mini"}}`)
	if code != 200 {
		t.Fatalf("review: %d %v", code, out)
	}
	if got := sentModel(t, f, "review", "low"); got != "gpt-mini" {
		t.Fatalf("review at low went as %q", got)
	}
	if got := sentModel(t, f, "review", "high"); got != "gpt-mid" {
		t.Fatalf("review at high kept its old high model: %q", got)
	}
	if code, out := effortEdit(f, "POST", "/api/v1/changes/"+out["change_id"].(string)+"/revert", ""); code != 200 {
		t.Fatalf("revert: %d %v", code, out)
	}
	if got := sentModel(t, f, "review", "high"); got != "gpt-big" {
		t.Fatalf("after the revert review at high went as %q", got)
	}
	// A shared task name, found by its name and Connection.
	if code, out := effortEdit(f, "PUT", "/api/v1/aliases/effort", `{"alias":"quick","connection":"hosted","effort":{"high":"gpt-big"}}`); code != 200 {
		t.Fatalf("quick: %d %v", code, out)
	}
	if got := sentModel(t, f, "quick", "high"); got != "gpt-big" {
		t.Fatalf("quick at high went as %q", got)
	}
	// A shared Level, the same way.
	if code, out := effortEdit(f, "PUT", "/api/v1/aliases/effort", `{"alias":"balanced","connection":"hosted","effort":{}}`); code != 200 {
		t.Fatalf("balanced: %d %v", code, out)
	}
	if got := sentModel(t, f, "balanced", "low"); got != "gpt-mid" {
		t.Fatalf("balanced kept its low model: %q", got)
	}
}

func TestLevelsListsTheSharedAndEachApplicationsOwnMappings(t *testing.T) {
	f := levelsServer(t)
	code, out := levelEdit(f, "GET", "/api/v1/levels", "")
	if code != 200 {
		t.Fatalf("%d %v", code, out)
	}
	shared, _ := out["shared"].([]any)
	if len(shared) != 3 {
		t.Fatalf("shared %v", shared)
	}
	if m := shared[2].(map[string]any); m["level"] != "balanced" || m["protocol"] != "anthropic" || m["connection"] != "claude" || m["model"] != "claude-sonnet" {
		t.Fatalf("shared balanced %v", m)
	}
	apps, _ := out["applications"].(map[string]any)
	own, _ := apps["review-code"].([]any)
	if len(own) != 1 || own[0].(map[string]any)["model"] != "llama3:70b" || own[0].(map[string]any)["protocol"] != "openai" {
		t.Fatalf("review-code %v", apps)
	}
	if docs, _ := apps["docs-bot"].([]any); docs == nil || len(docs) != 0 {
		t.Fatalf("docs-bot %v", apps["docs-bot"])
	}
}
