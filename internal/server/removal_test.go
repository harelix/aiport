package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Removing a Connection with a plan, over HTTP: the Owner reads what uses the
// Connection with a suggested replacement for each item and the Applications
// affected, then removes it with a plan as one revertable Change.

var (
	remManagement = strings.Repeat("o", 40)
	remDocs       = strings.Repeat("p", 40)
	remReview     = strings.Repeat("q", 40)
	remNightly    = strings.Repeat("s", 40)
)

// removalServer: hosted (OpenAI protocol, gpt-4o and gpt-4o-mini) is the
// OpenAI default, carries the shared Alias "quick" (gpt-4o-mini), docs-bot's
// own Alias "summarize" (gpt-4o), review-code's binding and the evaluator.
// spare also serves gpt-4o; local serves only llama3:8b and is nightly's
// binding; claude is the only Anthropic-protocol Connection.
func removalServer(t *testing.T) (*Handler, string) {
	t.Helper()
	f := removalFixture(t, "")
	return f.h, f.path
}

// removal is removalServer with its fake upstreams, so a test can make a
// Connection's calls fail.
type removal struct {
	h                            *Handler
	path                         string
	hosted, spare, local, claude *catalogUpstream
}

// removalFixture builds removalServer's gateway with extra appended to its
// configuration.
func removalFixture(t *testing.T, extra string) removal {
	t.Helper()
	hosted := newCatalogUpstream(t, 200, `{"object":"list","data":[{"id":"gpt-4o"},{"id":"gpt-4o-mini"}]}`)
	spare := newCatalogUpstream(t, 200, `{"object":"list","data":[{"id":"gpt-4o"}]}`)
	local := newCatalogUpstream(t, 200, `{"object":"list","data":[{"id":"llama3:8b"}]}`)
	claude := newCatalogUpstream(t, 200, `{"data":[{"id":"claude-haiku"}]}`)
	yaml := "version: 1\nauth: {token_env: REM_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + hosted.srv.URL + "/v1'}\n" +
		"  - {id: spare, provider: acme2, protocol: openai, base_url: '" + spare.srv.URL + "/v1'}\n" +
		"  - {id: local, provider: acme3, protocol: openai, base_url: '" + local.srv.URL + "/v1'}\n" +
		"  - {id: claude, provider: acme4, protocol: anthropic, base_url: '" + claude.srv.URL + "/v1'}\n" +
		"defaults: {openai_target: hosted, anthropic_target: claude}\n" +
		"aliases:\n  - {target: hosted, alias: quick, model: gpt-4o-mini}\n" +
		"evaluation: {connection: hosted, model: gpt-4o-mini}\n" +
		"origins:\n" +
		"  - id: docs-bot\n    name: Docs Bot\n    token_env: REM_DOCS_TOKEN\n    aliases:\n      - {alias: summarize, target: hosted, model: gpt-4o}\n" +
		"  - id: review-code\n    name: Review Code\n    token_env: REM_REVIEW_TOKEN\n    target: hosted\n" +
		"  - id: nightly\n    name: Nightly\n    token_env: REM_NIGHTLY_TOKEN\n    target: local\n" + extra
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REM_MGMT_TOKEN", remManagement)
	t.Setenv("REM_DOCS_TOKEN", remDocs)
	t.Setenv("REM_REVIEW_TOKEN", remReview)
	t.Setenv("REM_NIGHTLY_TOKEN", remNightly)
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	now := time.UnixMilli(1_800_000_000_000)
	h.now = func() time.Time { return now }
	t.Cleanup(h.Close)
	return removal{h: h, path: path, hosted: hosted, spare: spare, local: local, claude: claude}
}

// suggestions maps each listed dependent's id to its suggested replacement.
func suggestions(t *testing.T, out map[string]any) map[string]string {
	t.Helper()
	list, ok := out["dependents"].([]any)
	if !ok {
		t.Fatalf("no dependents in %v", out)
	}
	got := map[string]string{}
	for _, d := range list {
		m := d.(map[string]any)
		s, _ := m["suggestion"].(string)
		got[m["id"].(string)] = s
	}
	return got
}

// affected lists the ids of the affected Applications in an answer.
func affected(t *testing.T, out map[string]any) []string {
	t.Helper()
	list, ok := out["applications"].([]any)
	if !ok {
		t.Fatalf("no applications in %v", out)
	}
	got := []string{}
	for _, a := range list {
		got = append(got, a.(map[string]any)["application"].(string))
	}
	sort.Strings(got)
	return got
}

func planBody(plan map[string]string) string {
	b, _ := json.Marshal(map[string]any{"plan": plan})
	return string(b)
}

func TestDependentsListEveryUseWithASuggestionAndTheAffectedApplications(t *testing.T) {
	h, _ := removalServer(t)
	w, out := call(h, "GET", "/api/v1/targets/hosted/dependents", "", remManagement, false)
	if w.Code != 200 {
		t.Fatalf("dependents: %d %s", w.Code, w.Body)
	}
	want := map[string]string{
		"default:openai":                       "spare", // serves most of hosted's models
		"binding:review-code":                  "spare",
		"alias:quick":                          "spare", // nobody else serves gpt-4o-mini: the (new) protocol default
		"application_alias:docs-bot:summarize": "spare", // serves gpt-4o
		"evaluation":                           "spare",
	}
	got := suggestions(t, out)
	if len(got) != len(want) {
		t.Fatalf("dependents %v, want %v", got, want)
	}
	for id, s := range want {
		if got[id] != s {
			t.Fatalf("%s suggested %q, want %q (all: %v)", id, got[id], s, got)
		}
	}
	if a := affected(t, out); strings.Join(a, ",") != "aiport,docs-bot,review-code" {
		t.Fatalf("affected Applications %v", a)
	}
	if strings.Contains(w.Body.String(), "127.0.0.1") {
		t.Fatalf("the answer discloses a Connection URL: %s", w.Body)
	}

	// The only Connection of its protocol has no replacement to suggest.
	_, out = call(h, "GET", "/api/v1/targets/claude/dependents", "", remManagement, false)
	if got := suggestions(t, out); len(got) != 1 || got["default:anthropic"] != "" {
		t.Fatalf("claude's dependents: %v", got)
	}
	// Another plan, another set of affected Applications.
	w, out = call(h, "POST", "/api/v1/targets/local/dependents", planBody(map[string]string{"binding:nightly": ""}), remManagement, true)
	if w.Code != 200 || strings.Join(affected(t, out), ",") != "nightly" || suggestions(t, out)["binding:nightly"] != "hosted" {
		t.Fatalf("preview of local's removal: %d %s", w.Code, w.Body)
	}
	if w, _ := call(h, "GET", "/api/v1/targets/nowhere/dependents", "", remManagement, false); w.Code != 404 {
		t.Fatalf("dependents of a missing Connection: %d", w.Code)
	}
	if w, _ := call(h, "GET", "/api/v1/targets/hosted/dependents", "", remDocs, false); w.Code == 200 {
		t.Fatal("an Application token read management data")
	}
}

func TestRemovalWithAPlanIsOneChangeAndOneRevertRestoresEverything(t *testing.T) {
	h, path := removalServer(t)
	before, _ := os.ReadFile(path)
	_, out := call(h, "GET", "/api/v1/targets/hosted/dependents", "", remManagement, false)
	plan := suggestions(t, out)
	plan["alias:quick"] = "" // retire this task name on purpose
	w, out := call(h, "DELETE", "/api/v1/targets/hosted", planBody(plan), remManagement, true)
	if w.Code != 200 {
		t.Fatalf("removal with a plan: %d %s", w.Code, w.Body)
	}
	id, _ := out["change_id"].(string)
	if id == "" || strings.Join(affected(t, out), ",") != "aiport,docs-bot,review-code" {
		t.Fatalf("removal answer: %s", w.Body)
	}
	c := readConfig(t, path)
	for _, tg := range c.Targets {
		if tg.ID == "hosted" {
			t.Fatal("hosted is still configured")
		}
	}
	docs, _ := c.Application("docs-bot")
	review, _ := c.Application("review-code")
	if c.Defaults.OpenAITarget != "spare" || review.Target != "spare" || docs.Aliases[0].Target != "spare" || len(c.Aliases) != 0 || c.Evaluation.Connection != "spare" {
		t.Fatalf("not applied: %+v %+v %+v %+v %+v", c.Defaults, review, docs, c.Aliases, c.Evaluation)
	}
	if live := h.snapshot(); live.Defaults.OpenAITarget != "spare" {
		t.Fatal("not live")
	}
	changes := listChanges(t, h, remManagement)
	if len(changes) != 1 || changes[0]["id"] != id {
		t.Fatalf("removal was not one Change: %v", changes)
	}
	if w, _ := call(h, "POST", "/api/v1/changes/"+id+"/revert", "", remManagement, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatalf("one revert did not restore everything:\n%s", after)
	}
}

// Per-effort models name models of their Connection: an Alias or Level
// mapping re-pointed to another Connection loses them, as when the Level
// is mapped to another Connection.
func TestRePointingDropsPerEffortModels(t *testing.T) {
	h, path := removalServer(t)
	for _, e := range [][2]string{
		{"/api/v1/levels/openai/best", `{"connection":"hosted","model":"gpt-4o","effort":{"high":"gpt-4o-2"}}`},
		{"/api/v1/aliases/effort", `{"application":"docs-bot","alias":"summarize","effort":{"low":"gpt-4o-mini"}}`},
	} {
		if w, _ := call(h, "PUT", e[0], e[1], remManagement, true); w.Code != 200 {
			t.Fatalf("%s: %d %s", e[0], w.Code, w.Body)
		}
	}
	_, out := call(h, "GET", "/api/v1/targets/hosted/dependents", "", remManagement, false)
	plan := suggestions(t, out)
	if plan["level:best:openai"] != "spare" || plan["application_alias:docs-bot:summarize"] != "spare" {
		t.Fatalf("suggested plan: %v", plan)
	}
	if w, _ := call(h, "DELETE", "/api/v1/targets/hosted", planBody(plan), remManagement, true); w.Code != 200 {
		t.Fatalf("removal: %d %s", w.Code, w.Body)
	}
	c := readConfig(t, path)
	best, ok := c.Level("docs-bot", "openai", "best")
	if !ok || best.Connection != "spare" || best.Model != "gpt-4o" || !best.Effort.IsZero() {
		t.Fatalf("best after re-pointing: %+v", best)
	}
	docs, _ := c.Application("docs-bot")
	if a := docs.Aliases[0]; a.Target != "spare" || a.Model != "gpt-4o" || !a.Effort.IsZero() {
		t.Fatalf("summarize after re-pointing: %+v", a)
	}
}

func TestRemovalWithoutAPlanIsRefusedAsBefore(t *testing.T) {
	h, path := removalServer(t)
	before, _ := os.ReadFile(path)
	w, _ := call(h, "DELETE", "/api/v1/targets/hosted", "", remManagement, true)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "OpenAI-protocol default") {
		t.Fatalf("removal without a plan: %d %s", w.Code, w.Body)
	}
	// A plan that leaves an item unanswered is refused too, naming it.
	w, _ = call(h, "DELETE", "/api/v1/targets/hosted", planBody(map[string]string{"default:openai": "spare"}), remManagement, true)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "evaluation") {
		t.Fatalf("incomplete plan: %d %s", w.Code, w.Body)
	}
	// A replacement the result cannot accept (an OpenAI default on an
	// Anthropic Connection) is refused by validation.
	_, out := call(h, "GET", "/api/v1/targets/hosted/dependents", "", remManagement, false)
	plan := suggestions(t, out)
	plan["default:openai"] = "claude"
	if w, _ = call(h, "DELETE", "/api/v1/targets/hosted", planBody(plan), remManagement, true); w.Code != 400 {
		t.Fatalf("invalid plan: %d %s", w.Code, w.Body)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("a refused removal changed the file")
	}
	if n := len(listChanges(t, h, remManagement)); n != 0 {
		t.Fatalf("a refused removal recorded %d Changes", n)
	}
}

func TestNoneLeavesEveryItemWithoutAConnection(t *testing.T) {
	h, path := removalServer(t)
	plan := map[string]string{"default:openai": "", "binding:review-code": "", "alias:quick": "", "application_alias:docs-bot:summarize": "", "evaluation": ""}
	if w, _ := call(h, "DELETE", "/api/v1/targets/hosted", planBody(plan), remManagement, true); w.Code != 200 {
		t.Fatalf("removal with none everywhere: %d %s", w.Code, w.Body)
	}
	c := readConfig(t, path)
	docs, _ := c.Application("docs-bot")
	review, _ := c.Application("review-code")
	if c.Defaults.OpenAITarget != "" || review.Target != "" || len(docs.Aliases) != 0 || len(c.Aliases) != 0 || c.Evaluation != (config.Evaluation{}) {
		t.Fatalf("something still points somewhere: %+v %+v %+v %+v %+v", c.Defaults, review, docs, c.Aliases, c.Evaluation)
	}
}

// The built-in aiport Application (Explanations and Agent runs' model calls)
// follows the protocol defaults unless agent_runs.connection binds it, so a
// removal that moves a default changes its Catalog: it is listed among the
// affected Applications, by name, and its Catalog change is kept.
func TestTheBuiltInApplicationIsAmongTheAffectedApplications(t *testing.T) {
	h, _ := removalServer(t)
	_, out := call(h, "GET", "/api/v1/targets/hosted/dependents", "", remManagement, false)
	var builtin map[string]any
	for _, a := range out["applications"].([]any) {
		if m := a.(map[string]any); m["application"] == config.BuiltinApplication {
			builtin = m
		}
	}
	if builtin == nil || builtin["name"] != "aiport" {
		t.Fatalf("the built-in Application is not listed: %v", out["applications"])
	}
	if r, _ := builtin["routing"].([]any); len(r) != 1 || r[0].(map[string]any)["of"] != "openai" {
		t.Fatalf("the built-in Application's change: %v", builtin)
	}
	// Removing local (nightly's binding) changes nothing for it.
	_, out = call(h, "POST", "/api/v1/targets/local/dependents", planBody(map[string]string{"binding:nightly": ""}), remManagement, true)
	if a := affected(t, out); strings.Join(a, ",") != "nightly" {
		t.Fatalf("affected by removing local: %v", a)
	}

	// Removing hosted keeps the built-in Application's Catalog change, and
	// opens no Issue for it (it has no coding agent to tell).
	_, out = call(h, "GET", "/api/v1/targets/hosted/dependents", "", remManagement, false)
	if w, _ := call(h, "DELETE", "/api/v1/targets/hosted", planBody(suggestions(t, out)), remManagement, true); w.Code != 200 {
		t.Fatalf("removal: %d %s", w.Code, w.Body)
	}
	if got := h.recentCatalogChanges(config.BuiltinApplication, 10); len(got) != 1 {
		t.Fatalf("the built-in Application's Catalog changes: %v", got)
	}
	for _, is := range ownerIssues(t, h, "?state=all") {
		if is["application"] == config.BuiltinApplication {
			t.Fatalf("an Issue opened for the built-in Application: %v", is)
		}
	}
}

// A removal body larger than the API accepts is refused like elsewhere.
func TestAnOversizedRemovalBodyIsRefused(t *testing.T) {
	h, path := removalServer(t)
	before, _ := os.ReadFile(path)
	body := `{"plan":{"x":"` + strings.Repeat("a", 70<<10) + `"}}`
	w, out := call(h, "DELETE", "/api/v1/targets/hosted", body, remManagement, true)
	if w.Code != 413 || errorCode(out) != "request_too_large" {
		t.Fatalf("an oversized removal body: %d %s", w.Code, w.Body)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("a refused removal changed the file")
	}
}
