package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/telemetry"
)

// An Application that keeps asking for a Level nobody mapped on its
// protocol: the docs-bot fixture maps no Level, so best is refused with
// level_unmapped every time.

// Sticky level_unmapped refusals for one Level open one Issue for that
// Application and Level, with its evidence and an aiport fix (map the
// Level for that Application only) that waits for the Owner, who chooses
// the model: aiport never invents one. Approving maps the Level for
// docs-bot alone as a revertable Change; the next call succeeds and the
// Issue resolves itself as fixed in aiport.
func TestAnUnmappedLevelIssueStartToFinish(t *testing.T) {
	f := newIssuesFixture(t, "")
	if codes := f.chat(t, docsBotToken, "best", 4); fmt.Sprint(codes) != "[400 400 400 400]" {
		t.Fatalf("calls before the fix: %v", codes)
	}
	f.detect()
	if list := f.issues(t, "?state=all"); len(list) != 0 {
		t.Fatalf("an Issue before the refusals stuck: %v", list)
	}
	f.chat(t, docsBotToken, "best", 1)
	f.detect()
	f.detect() // a second pass over the same history adds nothing
	is := f.only(t)
	ev, fix := m(is["evidence"]), m(is["fix"])
	if is["application"] != "docs-bot" || is["cause"] != "level_unmapped" || is["subject"] != "best" || is["protocol"] != "openai" || is["state"] != "open" || is["needs_owner"] != true {
		t.Fatalf("issue: %v", is)
	}
	if ev["count"] != float64(5) || ev["window_minutes"] != float64(60) || ev["status"] != float64(400) || len(ev["examples"].([]any)) == 0 {
		t.Fatalf("evidence: %v", ev)
	}
	p := m(fix["proposal"])
	if fix["kind"] != "application_alias" || fix["scope"] != "application" || fix["state"] != "proposed" || p["alias"] != "best" || p["target"] != "hosted" || p["model"] != nil {
		t.Fatalf("aiport fix (the protocol default proposed, no model invented): %v", fix)
	}
	summary, _ := is["summary"].(string)
	for _, want := range []string{"docs-bot", "best", "5", "60 minutes", "OpenAI"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary %q does not mention %q", summary, want)
		}
	}
	text, _ := is["application_fix"].(string)
	for _, want := range []string{is["id"].(string), "docs-bot", `"best"`, "level_unmapped", "OpenAI", "application_alias"} {
		if !strings.Contains(text, want) {
			t.Fatalf("Application fix %q does not mention %q", text, want)
		}
	}
	_, pending := pendingDecisions(t, f.h, issuesManagement)
	if len(pending) != 1 || pending[0]["issue"] != is["id"] {
		t.Fatalf("the aiport fix is not waiting for the Owner: %v", pending)
	}

	id := is["id"].(string)
	if code, _ := f.decide(t, id, `{"action":"accept"}`); code != 400 {
		t.Fatalf("approve without a model: %d", code)
	}
	code, out := f.decide(t, id, `{"action":"accept","target":"local","model":"llama3:8b"}`)
	if code != 200 || out["change_id"] == "" {
		t.Fatalf("approve: %d %v", code, out)
	}
	c := liveConfig(t, f.h, issuesManagement)
	if mp, ok := c.Level("docs-bot", "openai", "best"); !ok || mp.Shared || mp.Connection != "local" || mp.Model != "llama3:8b" {
		t.Fatalf("best not mapped for docs-bot alone: %+v %v", mp, ok)
	}
	if _, ok := c.Level("other", "openai", "best"); ok {
		t.Fatal("the aiport fix mapped best for another Application")
	}
	is = f.only(t)
	if is["state"] != "resolving" || m(is["fix"])["state"] != "applied" {
		t.Fatalf("after approve: %v", is)
	}

	if codes := f.chat(t, docsBotToken, "best", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("call after the fix: %v", codes)
	}
	f.detect()
	is = f.only(t)
	if res := m(is["resolution"]); is["state"] != "resolved" || res["how"] != "aiport_fix" || res["change_id"] != out["change_id"] {
		t.Fatalf("after a successful call: %v", is)
	}
}

// The mapping is an ordinary Change: reverting it unmaps the Level again
// and reopens the Issue with its fix offered again.
func TestRevertingTheLevelMappingReopensTheIssue(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "best", 5)
	f.detect()
	id := f.only(t)["id"].(string)
	_, out := f.decide(t, id, `{"action":"accept","target":"local","model":"llama3:8b"}`)
	if w, _ := call(f.h, "POST", "/api/v1/changes/"+out["change_id"].(string)+"/revert", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	if _, ok := liveConfig(t, f.h, issuesManagement).Level("docs-bot", "openai", "best"); ok {
		t.Fatal("best still mapped after the revert")
	}
	f.detect()
	is := f.only(t)
	if is["id"] != id || is["state"] != "open" || m(is["fix"])["state"] != "proposed" || is["needs_owner"] != true {
		t.Fatalf("after reverting the mapping: %v", is)
	}
}

// One Issue per Application and Level: another Level, and another
// Application, each open their own.
func TestEachApplicationAndLevelHasItsOwnIssue(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "best", 5)
	f.chat(t, docsBotToken, "fast", 5)
	f.chat(t, otherAppToken, "best", 5)
	f.detect()
	got := map[string]bool{}
	for _, is := range f.issues(t, "?state=all") {
		if is["cause"] != "level_unmapped" {
			t.Fatalf("unexpected Issue: %v", is)
		}
		got[fmt.Sprint(is["application"], "/", is["subject"])] = true
	}
	if len(got) != 3 || !got["docs-bot/best"] || !got["docs-bot/fast"] || !got["other/best"] {
		t.Fatalf("Issues: %v", got)
	}
}

// With adapting off, unmapped Levels are still refused but no Issue opens.
func TestNoUnmappedLevelIssuesWithAdaptingOff(t *testing.T) {
	f := newIssuesFixture(t, "adapting: {disabled: true}\n")
	if codes := f.chat(t, docsBotToken, "best", 6); fmt.Sprint(codes) != "[400 400 400 400 400 400]" {
		t.Fatalf("forwarding: %v", codes)
	}
	f.detect()
	w, out := call(f.h, "GET", "/api/v1/issues?state=all", "", issuesManagement, false)
	if w.Code != 200 || out["adapting"] != false || len(out["issues"].([]any)) != 0 {
		t.Fatalf("issues with the loop off: %d %v", w.Code, out)
	}
	db, _ := os.ReadFile(f.path + ".db")
	if strings.Contains(string(db), `"cause":"level_unmapped"`) {
		t.Fatal("an Issue was kept with the loop off")
	}
	if c := liveConfig(t, f.h, issuesManagement); len(c.Aliases) != 0 {
		t.Fatalf("something was mapped: %v", c.Aliases)
	}
}

// The fix maps the Level on the protocol the calls are refused on: a
// Connection of another protocol is refused and nothing is mapped.
func TestTheLevelFixStaysOnTheProtocolOfTheCalls(t *testing.T) {
	f := &issuesFixture{path: filepath.Join(t.TempDir(), "aiport.yaml"), at: time.Now()}
	f.local = newFakeConnection(t, `{"object":"list","data":[{"id":"llama3:8b"}]}`, true)
	yaml := "version: 1\nauth: {token_env: ISSUES_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: local, provider: ollama, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: claude, provider: acme2, protocol: anthropic, base_url: 'http://127.0.0.1:9/v1'}\n" +
		"defaults: {openai_target: local}\norigins:\n" +
		"  - {id: docs-bot, name: Docs Bot, token_env: ISSUES_DOCS_TOKEN}\n"
	if err := os.WriteFile(f.path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ISSUES_MGMT_TOKEN", issuesManagement)
	t.Setenv("ISSUES_DOCS_TOKEN", docsBotToken)
	f.start(t)
	f.chat(t, docsBotToken, "balanced", 5)
	f.detect()
	is := f.only(t)
	if is["protocol"] != "openai" || m(m(is["fix"])["proposal"])["target"] != "local" {
		t.Fatalf("issue: %v", is)
	}
	if code, out := f.decide(t, is["id"].(string), `{"action":"accept","target":"claude","model":"claude-sonnet"}`); code != 400 || !strings.Contains(fmt.Sprint(out["error"]), "OpenAI") {
		t.Fatalf("a fix on the Anthropic protocol: %d %v", code, out)
	}
	c := liveConfig(t, f.h, issuesManagement)
	for _, p := range []string{"openai", "anthropic"} {
		if _, ok := c.Level("docs-bot", p, "balanced"); ok {
			t.Fatalf("balanced was mapped on %s", p)
		}
	}
}

// A Level unmapped on both protocols opens one Issue per protocol, each
// with its fix on its own protocol, and calls on one protocol never move
// the other's Issue.
func TestALevelUnmappedOnBothProtocolsOpensAnIssueForEach(t *testing.T) {
	f := &issuesFixture{path: filepath.Join(t.TempDir(), "aiport.yaml"), at: time.Now()}
	f.local = newFakeConnection(t, `{"object":"list","data":[{"id":"llama3:8b"}]}`, true)
	yaml := "version: 1\nauth: {token_env: ISSUES_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: local, provider: ollama, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: claude, provider: acme2, protocol: anthropic, base_url: 'http://127.0.0.1:9/v1'}\n" +
		"defaults: {openai_target: local, anthropic_target: claude}\norigins:\n" +
		"  - {id: docs-bot, name: Docs Bot, token_env: ISSUES_DOCS_TOKEN}\n"
	if err := os.WriteFile(f.path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ISSUES_MGMT_TOKEN", issuesManagement)
	t.Setenv("ISSUES_DOCS_TOKEN", docsBotToken)
	f.start(t)
	f.chat(t, docsBotToken, "best", 5)
	before := len(f.collector.Find(telemetry.Query{}))
	for i := 0; i < 5; i++ {
		w := levelCall(f.h, "/v1/messages", docsBotToken, `{"model":"best","max_tokens":5}`)
		if w.Code != 400 {
			t.Fatalf("best on Anthropic: %d %s", w.Code, w.Body)
		}
	}
	for deadline := time.Now().Add(2 * time.Second); len(f.collector.Find(telemetry.Query{})) < before+5; time.Sleep(2 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("request history did not record the Anthropic calls")
		}
	}
	f.detect()
	f.detect()
	byProtocol := map[string]map[string]any{}
	for _, is := range f.issues(t, "?state=all") {
		if is["cause"] != "level_unmapped" || is["subject"] != "best" {
			t.Fatalf("unexpected Issue: %v", is)
		}
		byProtocol[fmt.Sprint(is["protocol"])] = is
	}
	if len(byProtocol) != 2 || byProtocol["openai"] == nil || byProtocol["anthropic"] == nil {
		t.Fatalf("Issues by protocol: %v", byProtocol)
	}
	for p, target := range map[string]string{"openai": "local", "anthropic": "claude"} {
		is := byProtocol[p]
		if got := m(m(is["fix"])["proposal"])["target"]; got != target {
			t.Fatalf("%s Issue proposes %v, want %s", p, got, target)
		}
		if m(is["evidence"])["count"] != float64(5) {
			t.Fatalf("%s Issue counts %v", p, m(is["evidence"])["count"])
		}
	}
	// Each fix maps the Level on its own protocol only.
	if code, out := f.decide(t, byProtocol["anthropic"]["id"].(string), `{"action":"accept","target":"claude","model":"claude-sonnet"}`); code != 200 {
		t.Fatalf("approve the Anthropic fix: %d %v", code, out)
	}
	c := liveConfig(t, f.h, issuesManagement)
	if _, ok := c.Level("docs-bot", "anthropic", "best"); !ok {
		t.Fatal("best not mapped on Anthropic")
	}
	if _, ok := c.Level("docs-bot", "openai", "best"); ok {
		t.Fatal("the Anthropic fix mapped best on OpenAI")
	}
	if is := f.issues(t, "?state=all"); len(is) != 2 {
		t.Fatalf("Issues after the fix: %v", is)
	}
}
