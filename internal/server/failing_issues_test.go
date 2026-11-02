package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// flakyConnection is a provider the test controls: while broken it answers
// every chat call with a 502, otherwise with a short completion.
type flakyConnection struct {
	mu     sync.Mutex
	broken bool
	srv    *httptest.Server
}

func newFlakyConnection(t *testing.T) *flakyConnection {
	f := &flakyConnection{broken: true}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if f.broken {
			w.WriteHeader(502)
			_, _ = w.Write([]byte(`{"error":{"message":"bad gateway (provider secret detail)"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"c1","object":"chat.completion","model":"llama3:8b","choices":[{"index":0,"message":{"role":"assistant","content":"hello"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1,"total_tokens":4}}`))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *flakyConnection) set(broken bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.broken = broken
}

// failingFixture: docs-bot is bound to the Connection "flaky", which keeps
// failing; the Connection "local" works and is the protocol default the
// other Application uses.
func failingFixture(t *testing.T) (*issuesFixture, *flakyConnection) {
	t.Helper()
	f := &issuesFixture{path: filepath.Join(t.TempDir(), "aiport.yaml"), at: time.Now()}
	f.local = newFakeConnection(t, `{"object":"list","data":[{"id":"llama3:8b"}]}`, true)
	flaky := newFlakyConnection(t)
	yaml := "version: 1\nauth: {token_env: ISSUES_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: local, provider: ollama, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: flaky, provider: ollama, base_url: '" + flaky.srv.URL + "/v1'}\n" +
		"defaults: {openai_target: local}\norigins:\n" +
		"  - {id: docs-bot, name: Docs Bot, token_env: ISSUES_DOCS_TOKEN, target: flaky}\n" +
		"  - {id: other, name: Other App, token_env: ISSUES_OTHER_TOKEN}\n"
	if err := os.WriteFile(f.path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ISSUES_MGMT_TOKEN", issuesManagement)
	t.Setenv("ISSUES_DOCS_TOKEN", docsBotToken)
	t.Setenv("ISSUES_OTHER_TOKEN", otherAppToken)
	f.start(t)
	return f, flaky
}

func (f *issuesFixture) binding(t *testing.T, app string) string {
	t.Helper()
	o, _ := liveConfig(t, f.h, issuesManagement).Application(app)
	return o.Target
}

// Sticky failures on one Connection open one Issue for the Application and
// that Connection, with evidence, an aiport fix (bind to a working
// Connection) that waits for the Owner, and the Application fix. Approving
// binds only that Application as a revertable Change; the next calls succeed
// and the Issue resolves itself as fixed in aiport.
func TestFailingConnectionIssueStartToFinish(t *testing.T) {
	f, _ := failingFixture(t)
	if codes := f.chat(t, otherAppToken, "llama3:8b", 2); fmt.Sprint(codes) != "[200 200]" {
		t.Fatalf("the working Connection: %v", codes)
	}
	if codes := f.chat(t, docsBotToken, "llama3:8b", 4); fmt.Sprint(codes) != "[502 502 502 502]" {
		t.Fatalf("calls to the failing Connection: %v", codes)
	}
	f.detect()
	if list := f.issues(t, "?state=all"); len(list) != 0 {
		t.Fatalf("an Issue before the failures stuck: %v", list)
	}
	f.chat(t, docsBotToken, "llama3:8b", 1)
	f.detect()
	f.detect() // a second pass over the same history adds nothing
	is := f.only(t)
	ev, fix := m(is["evidence"]), m(is["fix"])
	if is["application"] != "docs-bot" || is["cause"] != "failing_connection" || is["subject"] != "flaky" || is["state"] != "open" || is["needs_owner"] != true {
		t.Fatalf("issue: %v", is)
	}
	if ev["count"] != float64(5) || ev["window_minutes"] != float64(60) || ev["status"] != float64(502) || len(ev["examples"].([]any)) == 0 || fmt.Sprint(ev["connections"]) != "[flaky]" {
		t.Fatalf("evidence: %v", ev)
	}
	if fix["kind"] != "bind_application" || fix["scope"] != "application" || fix["state"] != "proposed" || m(fix["proposal"])["target"] != "local" {
		t.Fatalf("aiport fix: %v", fix)
	}
	summary, _ := is["summary"].(string)
	for _, want := range []string{"docs-bot", "flaky", "5", "60 minutes", "local"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary %q does not mention %q", summary, want)
		}
	}
	text, _ := is["application_fix"].(string)
	for _, want := range []string{is["id"].(string), "docs-bot", "\"flaky\"", "bind_application", "\"local\""} {
		if !strings.Contains(text, want) {
			t.Fatalf("Application fix %q does not mention %q", text, want)
		}
	}
	_, pending := pendingDecisions(t, f.h, issuesManagement)
	if len(pending) != 1 || pending[0]["issue"] != is["id"] || pending[0]["kind"] != "bind_application" {
		t.Fatalf("the aiport fix is not waiting for the Owner: %v", pending)
	}
	// Never applied automatically: docs-bot is still bound to flaky.
	if got := f.binding(t, "docs-bot"); got != "flaky" {
		t.Fatalf("the binding changed without the Owner: %q", got)
	}

	id := is["id"].(string)
	if code, out := f.decide(t, id, `{"action":"accept","target":"flaky"}`); code != 400 || !strings.Contains(fmt.Sprint(out["error"]), "another Connection") {
		t.Fatalf("binding to the failing Connection itself: %d %v", code, out)
	}
	code, out := f.decide(t, id, `{"action":"accept","target":"local"}`)
	if code != 200 || out["change_id"] == "" {
		t.Fatalf("approve: %d %v", code, out)
	}
	if got := f.binding(t, "docs-bot"); got != "local" {
		t.Fatalf("binding not live: %q", got)
	}
	if got := f.binding(t, "other"); got != "" {
		t.Fatalf("another Application was bound: %q", got)
	}
	is = f.only(t)
	if is["state"] != "resolving" || m(is["fix"])["state"] != "applied" || m(is["fix"])["change_id"] != out["change_id"] {
		t.Fatalf("after approve: %v", is)
	}
	if text, _ := is["application_fix"].(string); !strings.Contains(text, "Already fixed in aiport: docs-bot is now bound to Connection \"local\". No code change is needed") {
		t.Fatalf("Application fix after the aiport fix:\n%s", text)
	}

	if codes := f.chat(t, docsBotToken, "llama3:8b", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("call after the fix: %v", codes)
	}
	f.detect()
	is = f.only(t)
	if res := m(is["resolution"]); is["state"] != "resolved" || res["how"] != "aiport_fix" || res["change_id"] != out["change_id"] {
		t.Fatalf("after a successful call: %v", is)
	}
}

// Reverting the binding reopens the Issue with its fix offered again.
func TestRevertingTheBindingReopensTheIssue(t *testing.T) {
	f, _ := failingFixture(t)
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)
	_, out := f.decide(t, id, `{"action":"accept","target":"local"}`)
	if w, _ := call(f.h, "POST", "/api/v1/changes/"+out["change_id"].(string)+"/revert", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	f.detect()
	is := f.only(t)
	if is["id"] != id || is["state"] != "open" || m(is["fix"])["state"] != "proposed" || is["needs_owner"] != true || f.binding(t, "docs-bot") != "flaky" {
		t.Fatalf("after reverting the binding: %v", is)
	}
}

// When the Connection works again the Issue resolves itself as stopped and
// its fix stops waiting; it reopens when the failures stick again, and only
// new failures count.
func TestFailingConnectionIssueResolvesAndReopens(t *testing.T) {
	f, flaky := failingFixture(t)
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)

	flaky.set(false)
	if codes := f.chat(t, docsBotToken, "llama3:8b", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("the Connection works again: %v", codes)
	}
	f.detect()
	is := f.only(t)
	if is["state"] != "resolved" || m(is["resolution"])["how"] != "stopped" {
		t.Fatalf("after a successful call: %v", is)
	}
	if _, pending := pendingDecisions(t, f.h, issuesManagement); len(pending) != 0 {
		t.Fatalf("the fix still waits for the Owner: %v", pending)
	}
	f.detect()
	if is := f.only(t); is["state"] != "resolved" {
		t.Fatalf("the failures from before the resolution reopened it: %v", is)
	}

	flaky.set(true)
	f.chat(t, docsBotToken, "llama3:8b", 4)
	f.detect()
	if is := f.only(t); is["state"] != "resolved" {
		t.Fatalf("reopened before the failures stuck: %v", is)
	}
	f.chat(t, docsBotToken, "llama3:8b", 1)
	f.detect()
	is = f.only(t)
	if is["id"] != id || is["state"] != "open" || is["reopened"] != float64(1) || m(is["fix"])["state"] != "proposed" {
		t.Fatalf("the failures came back: %v", is)
	}
	// The calls of another Application succeeding elsewhere do not resolve it.
	f.chat(t, otherAppToken, "llama3:8b", 1)
	f.detect()
	if is := f.only(t); is["state"] != "open" {
		t.Fatalf("another Application's success resolved it: %v", is)
	}
}

// "Fixed on my side" for a failing Connection is confirmed only by a
// successful call routed for that Application after it said so.
func TestAFailingConnectionApplicationFixIsConfirmedByItsOwnCalls(t *testing.T) {
	f, flaky := failingFixture(t)
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)
	if w, _ := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", docsBotToken, false); w.Code != 200 {
		t.Fatalf("fixed: %d", w.Code)
	}
	f.chat(t, otherAppToken, "llama3:8b", 1)
	f.detect()
	if is := f.only(t); is["state"] != "resolving" {
		t.Fatalf("another Application's success confirmed it: %v", is)
	}
	flaky.set(false)
	if codes := f.chat(t, docsBotToken, "llama3:8b", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("docs-bot's call: %v", codes)
	}
	f.detect()
	if is := f.only(t); is["state"] != "resolved" || m(is["resolution"])["how"] != "application_fix" {
		t.Fatalf("after docs-bot's successful call: %v", is)
	}
}
