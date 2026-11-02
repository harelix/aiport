package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// fakeConnection is a provider the test controls: the models it lists and
// whether it answers chat calls (with a short completion) or 404s them as
// an unknown model.
type fakeConnection struct {
	mu      sync.Mutex
	models  string // the body of GET /v1/models; "" = the listing fails
	serves  bool
	prompts []string
	srv     *httptest.Server
}

func newFakeConnection(t *testing.T, models string, serves bool) *fakeConnection {
	f := &fakeConnection{models: models, serves: serves}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/models") {
			if f.models == "" {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(`{"error":{"message":"listing broke: provider secret detail"}}`))
				return
			}
			_, _ = w.Write([]byte(f.models))
			return
		}
		var body struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.prompts = append(f.prompts, body.Model)
		if !f.serves {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error":{"type":"invalid_request_error","code":"model_not_found","message":"The model ` + body.Model + ` does not exist (provider secret detail)"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"c1","object":"chat.completion","model":"` + body.Model + `","choices":[{"index":0,"message":{"role":"assistant","content":"hello"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1,"total_tokens":4}}`))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeConnection) set(models string, serves bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.models, f.serves = models, serves
}

const (
	issuesManagement = "issues-management-token-0123456789abcdef"
	docsBotToken     = "docs-bot-application-token-0123456789abc"
	otherAppToken    = "other-app-application-token-0123456789ab"
	hostedKey        = "sk-hosted-key-never-in-an-issue-0123456"
)

// issuesFixture is the docs-bot example: docs-bot is bound to nothing, the
// protocol default is a hosted Connection that does not serve llama3:8b,
// and a local Connection does.
type issuesFixture struct {
	h             *Handler
	path          string
	at            time.Time
	local, hosted *fakeConnection
	collector     *telemetry.Collector
	extra         string
}

func newIssuesFixture(t *testing.T, extra string) *issuesFixture {
	t.Helper()
	f := &issuesFixture{path: filepath.Join(t.TempDir(), "aiport.yaml"), at: time.Now(), extra: extra}
	f.local = newFakeConnection(t, `{"object":"list","data":[{"id":"llama3:8b"},{"id":"qwen2.5:7b"}]}`, true)
	f.hosted = newFakeConnection(t, `{"object":"list","data":[{"id":"gpt-4o-mini"}]}`, false)
	yaml := "version: 1\nauth: {token_env: ISSUES_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: local, provider: ollama, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + f.hosted.srv.URL + "/v1', api_key_env: ISSUES_HOSTED_KEY}\n" +
		"defaults: {openai_target: hosted}\norigins:\n" +
		"  - {id: docs-bot, name: Docs Bot, token_env: ISSUES_DOCS_TOKEN}\n" +
		"  - {id: other, name: Other App, token_env: ISSUES_OTHER_TOKEN}\n" + extra
	if err := os.WriteFile(f.path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ISSUES_MGMT_TOKEN", issuesManagement)
	t.Setenv("ISSUES_DOCS_TOKEN", docsBotToken)
	t.Setenv("ISSUES_OTHER_TOKEN", otherAppToken)
	t.Setenv("ISSUES_HOSTED_KEY", hostedKey)
	f.start(t)
	return f
}

// start runs aiport over the fixture's configuration and database, as after
// a restart.
func (f *issuesFixture) start(t *testing.T) {
	t.Helper()
	c, err := config.Load(f.path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(f.path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	f.collector = telemetry.New(true, 256)
	f.h = NewConfigured(c, s, f.collector, f.path).(*Handler)
	f.h.now = func() time.Time { return f.at }
	h := f.h
	t.Cleanup(h.Close)
}

// chat sends n chat calls for model with token and waits until request
// history has them. Each call happens a second after the last.
func (f *issuesFixture) chat(t *testing.T, token, model string, n int) []int {
	t.Helper()
	before := len(f.collector.Find(telemetry.Query{}))
	var codes []int
	for i := 0; i < n; i++ {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", strings.NewReader(`{"model":"`+model+`","messages":[{"role":"user","content":"PRIVATE PROMPT TEXT"}]}`))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		f.h.ServeHTTP(w, r)
		codes = append(codes, w.Code)
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(f.collector.Find(telemetry.Query{})) < before+n {
		if time.Now().After(deadline) {
			t.Fatal("request history did not record the calls")
		}
		time.Sleep(2 * time.Millisecond)
	}
	return codes
}

// detect runs one background detection pass right after the last call; the
// clock stays there until the next pass.
func (f *issuesFixture) detect() {
	time.Sleep(2 * time.Millisecond)
	f.at = time.Now()
	f.h.Detect(f.at)
	time.Sleep(2 * time.Millisecond) // later calls happen strictly after it
}

func (f *issuesFixture) issues(t *testing.T, query string) []map[string]any {
	t.Helper()
	w, out := call(f.h, "GET", "/api/v1/issues"+query, "", issuesManagement, false)
	if w.Code != 200 {
		t.Fatalf("list issues: %d %s", w.Code, w.Body)
	}
	list := []map[string]any{}
	for _, x := range out["issues"].([]any) {
		list = append(list, x.(map[string]any))
	}
	return list
}

func (f *issuesFixture) only(t *testing.T) map[string]any {
	t.Helper()
	list := f.issues(t, "?state=all")
	if len(list) != 1 {
		t.Fatalf("want exactly one Issue, got %d: %v", len(list), list)
	}
	return list[0]
}

func (f *issuesFixture) decide(t *testing.T, id, body string) (int, map[string]any) {
	t.Helper()
	w, out := call(f.h, "POST", "/api/v1/issues/"+id, body, issuesManagement, true)
	return w.Code, out
}

// aliases are Application app's own Aliases in the running configuration.
func (f *issuesFixture) aliases(t *testing.T, app string) []config.OriginAlias {
	t.Helper()
	o, _ := liveConfig(t, f.h, issuesManagement).Application(app)
	return o.Aliases
}

func m(v any) map[string]any { x, _ := v.(map[string]any); return x }

// The docs-bot example, start to finish: sticky 404s for llama3:8b open one
// Issue with its evidence and the diagnosis that only the local Connection
// serves the model; Approve adds an Alias for docs-bot only as a revertable
// Change; the next calls succeed and the Issue resolves itself as fixed in
// aiport.
func TestUnknownModelIssueStartToFinish(t *testing.T) {
	f := newIssuesFixture(t, "")
	if codes := f.chat(t, docsBotToken, "llama3:8b", 4); fmt.Sprint(codes) != "[404 404 404 404]" {
		t.Fatalf("calls before the fix: %v", codes)
	}
	f.detect()
	if list := f.issues(t, "?state=all"); len(list) != 0 {
		t.Fatalf("an Issue before the error stuck: %v", list)
	}
	f.chat(t, docsBotToken, "llama3:8b", 2)
	f.detect()
	f.detect() // a second pass over the same history adds nothing
	is := f.only(t)
	ev, diag, fix := m(is["evidence"]), m(is["diagnosis"]), m(is["fix"])
	if is["application"] != "docs-bot" || is["cause"] != "unknown_model" || is["subject"] != "llama3:8b" || is["state"] != "open" {
		t.Fatalf("issue: %v", is)
	}
	if ev["count"] != float64(6) || ev["window_minutes"] != float64(60) || ev["status"] != float64(404) || ev["error"] != "model_not_found" || len(ev["examples"].([]any)) == 0 || fmt.Sprint(ev["connections"]) != "[hosted]" {
		t.Fatalf("evidence: %v", ev)
	}
	if diag["answer"] != "one" || fmt.Sprint(diag["serving"]) != "[local]" || diag["protocol"] != "openai" || diag["summary"] != `Connection "local" serves "llama3:8b" (from the Connections' free model lists).` {
		t.Fatalf("diagnosis: %v", diag)
	}
	p := m(fix["proposal"])
	if fix["kind"] != "application_alias" || fix["scope"] != "application" || fix["state"] != "proposed" || p["alias"] != "llama3:8b" || p["target"] != "local" || p["model"] != "llama3:8b" {
		t.Fatalf("aiport fix: %v", fix)
	}
	if text, _ := is["application_fix"].(string); !strings.Contains(text, "llama3:8b") || !strings.Contains(text, "local") {
		t.Fatalf("Application fix: %q", text)
	}
	// The Changes page: the Issue needs the Owner, so it counts for the badge.
	_, pending := pendingDecisions(t, f.h, issuesManagement)
	if len(pending) != 1 || pending[0]["issue"] != is["id"] {
		t.Fatalf("the aiport fix is not waiting for the Owner: %v", pending)
	}

	code, out := f.decide(t, is["id"].(string), `{"action":"accept"}`)
	if code != 200 || out["change_id"] == "" {
		t.Fatalf("approve: %d %v", code, out)
	}
	if got := f.aliases(t, "docs-bot"); len(got) != 1 || got[0] != (config.OriginAlias{Alias: "llama3:8b", Target: "local", Model: "llama3:8b"}) {
		t.Fatalf("alias not live for docs-bot: %v", got)
	}
	if len(f.aliases(t, "other")) != 0 || len(liveConfig(t, f.h, issuesManagement).Aliases) != 0 {
		t.Fatal("the aiport fix reached another Application")
	}
	is = f.only(t)
	if is["state"] != "resolving" || m(is["fix"])["state"] != "applied" || m(is["fix"])["change_id"] != out["change_id"] {
		t.Fatalf("after approve: %v", is)
	}
	if _, pending := pendingDecisions(t, f.h, issuesManagement); len(pending) != 0 {
		t.Fatalf("still waiting for the Owner after approve: %v", pending)
	}

	if codes := f.chat(t, docsBotToken, "llama3:8b", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("call after the fix: %v", codes)
	}
	f.detect()
	is = f.only(t)
	res := m(is["resolution"])
	if is["state"] != "resolved" || res["how"] != "aiport_fix" || res["change_id"] != out["change_id"] {
		t.Fatalf("after a successful call: %v", is)
	}
	if list := f.issues(t, ""); len(list) != 0 {
		t.Fatalf("a resolved Issue is listed as open: %v", list)
	}
}

// An approved aiport fix is credited only when a call shows it works: if
// docs-bot makes no calls for a whole window after the fix, the Issue
// resolves as stopped, not as fixed by aiport.
func TestAnApprovedFixWithNoCallsAfterItResolvesAsStopped(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	if code, out := f.decide(t, f.only(t)["id"].(string), `{"action":"accept"}`); code != 200 {
		t.Fatalf("approve: %d %v", code, out)
	}
	f.h.Detect(f.at.Add(30 * time.Minute))
	if is := f.only(t); is["state"] != "resolving" {
		t.Fatalf("resolved before a whole quiet window: %v", is)
	}
	f.h.Detect(f.at.Add(61 * time.Minute))
	is := f.only(t)
	if res := m(is["resolution"]); is["state"] != "resolved" || res["how"] != "stopped" || res["change_id"] != nil {
		t.Fatalf("after a quiet window: %v", is)
	}
}

// The Owner can dismiss any Issue that is not resolved or dismissed yet,
// whatever its aiport fix stands at.
func TestAnyUndecidedIssueCanBeDismissed(t *testing.T) {
	for _, first := range []string{"accept", "reject"} {
		t.Run(first, func(t *testing.T) {
			f := newIssuesFixture(t, "")
			f.chat(t, docsBotToken, "llama3:8b", 5)
			f.detect()
			id := f.only(t)["id"].(string)
			if code, out := f.decide(t, id, `{"action":"`+first+`"}`); code != 200 {
				t.Fatalf("%s: %d %v", first, code, out)
			}
			code, out := f.decide(t, id, `{"action":"dismiss","note":"not now"}`)
			if code != 200 || m(out["issue"])["state"] != "dismissed" {
				t.Fatalf("dismiss after %s: %d %v", first, code, out)
			}
			if is := f.only(t); is["state"] != "dismissed" || is["needs_owner"] != false {
				t.Fatalf("after dismiss: %v", is)
			}
			if code, out := f.decide(t, id, `{"action":"dismiss"}`); code != 409 {
				t.Fatalf("dismissing twice: %d %v", code, out)
			}
		})
	}
}

// With no Connection, several or an unlisted one serving the model, the
// Issue stays open with no proposed Connection; the Owner chooses one.
func TestUncertainDiagnosesLeaveTheConnectionToTheOwner(t *testing.T) {
	cases := map[string]struct {
		local, hosted string
		answer        string
		serving       string
	}{
		"none":    {`{"data":[{"id":"qwen2.5:7b"}]}`, `{"data":[{"id":"gpt-4o-mini"}]}`, "none", "[]"},
		"several": {`{"data":[{"id":"llama3:8b"}]}`, `{"data":[{"id":"llama3:8b"}]}`, "several", "[hosted local]"},
		"unknown": {"", `{"data":[{"id":"gpt-4o-mini"}]}`, "unknown", "[]"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newIssuesFixture(t, "")
			f.local.set(tc.local, true)
			f.hosted.set(tc.hosted, false)
			f.chat(t, docsBotToken, "llama3:8b", 5)
			f.detect()
			is := f.only(t)
			diag, fix := m(is["diagnosis"]), m(is["fix"])
			if is["state"] != "open" || diag["answer"] != tc.answer || fmt.Sprint(diag["serving"]) != tc.serving {
				t.Fatalf("issue: %v", is)
			}
			if p := m(fix["proposal"]); p["target"] != nil || fix["state"] != "proposed" || is["needs_owner"] != true {
				t.Fatalf("a Connection was proposed without certainty: %v", fix)
			}
			if code, out := f.decide(t, is["id"].(string), `{"action":"accept"}`); code != 400 {
				t.Fatalf("accepted without a Connection: %d %v", code, out)
			}
			code, out := f.decide(t, is["id"].(string), `{"action":"accept","target":"local"}`)
			if code != 200 || m(out["issue"])["state"] != "resolving" {
				t.Fatalf("accept with the Owner's Connection: %d %v", code, out)
			}
			if got := f.aliases(t, "docs-bot"); len(got) != 1 || got[0] != (config.OriginAlias{Alias: "llama3:8b", Target: "local", Model: "llama3:8b"}) {
				t.Fatalf("alias: %v", got)
			}
		})
	}
}

// Reverting the aiport fix reopens the Issue and offers the fix again; an
// Issue whose calls succeed again without a fix resolves as stopped; a
// resolved Issue reopens when the error sticks again, and only new errors
// count.
func TestAnIssueReopensWhenTheErrorReturns(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)
	_, out := f.decide(t, id, `{"action":"accept"}`)
	change := out["change_id"].(string)
	if w, _ := call(f.h, "POST", "/api/v1/changes/"+change+"/revert", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	f.detect()
	is := f.only(t)
	if is["state"] != "open" || m(is["fix"])["state"] != "proposed" || is["needs_owner"] != true {
		t.Fatalf("after reverting the fix: %v", is)
	}

	// The hosted Connection starts serving the model: the calls succeed
	// again without an aiport fix.
	f.hosted.set(`{"data":[{"id":"gpt-4o-mini"},{"id":"llama3:8b"}]}`, true)
	f.chat(t, docsBotToken, "llama3:8b", 1)
	f.detect()
	is = f.only(t)
	if is["state"] != "resolved" || m(is["resolution"])["how"] != "stopped" {
		t.Fatalf("after a successful call without a fix: %v", is)
	}
	if _, pending := pendingDecisions(t, f.h, issuesManagement); len(pending) != 0 {
		t.Fatalf("the fix still waits for the Owner after the Issue resolved: %v", pending)
	}
	f.detect()
	if is := f.only(t); is["state"] != "resolved" {
		t.Fatalf("the errors from before the resolution reopened it: %v", is)
	}

	f.hosted.set(`{"data":[{"id":"gpt-4o-mini"}]}`, false)
	f.chat(t, docsBotToken, "llama3:8b", 4)
	f.detect()
	if is := f.only(t); is["state"] != "resolved" {
		t.Fatalf("reopened before the error stuck: %v", is)
	}
	f.chat(t, docsBotToken, "llama3:8b", 1)
	f.detect()
	is = f.only(t)
	if is["id"] != id || is["state"] != "open" || is["reopened"] != float64(1) || is["resolution"] != nil || m(is["fix"])["state"] != "proposed" {
		t.Fatalf("the error came back: %v", is)
	}
}

// Dismissing an Issue keeps it quiet, also across a restart, until its
// evidence doubles; Issues persist across restarts.
func TestADismissedIssueStaysQuiet(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)
	if code, out := f.decide(t, id, `{"action":"dismiss","note":"not now"}`); code != 200 || m(out["issue"])["state"] != "dismissed" {
		t.Fatalf("dismiss: %d %v", code, out)
	}
	if code, _ := f.decide(t, id, `{"action":"accept"}`); code != 409 {
		t.Fatalf("decided a dismissed Issue: %d", code)
	}
	if list := f.issues(t, ""); len(list) != 0 {
		t.Fatalf("a dismissed Issue is listed as open: %v", list)
	}
	if _, pending := pendingDecisions(t, f.h, issuesManagement); len(pending) != 0 {
		t.Fatalf("a dismissed Issue still waits for the Owner: %v", pending)
	}

	f.h.Close()
	f.start(t) // a restart: request history is gone, the Issue is kept
	if is := f.only(t); is["id"] != id || is["state"] != "dismissed" || m(is["evidence"])["count"] != float64(5) {
		t.Fatalf("after a restart: %v", is)
	}
	f.chat(t, docsBotToken, "llama3:8b", 9)
	f.detect()
	if is := f.only(t); is["state"] != "dismissed" {
		t.Fatalf("came back before the evidence doubled: %v", is)
	}
	f.chat(t, docsBotToken, "llama3:8b", 1)
	f.detect()
	if is := f.only(t); is["id"] != id || is["state"] != "open" || m(is["evidence"])["count"] != float64(10) {
		t.Fatalf("evidence doubled: %v", is)
	}
}

// Issues hold no prompts, responses, provider messages, keys or tokens, and
// only the Owner's dashboard reads them.
func TestIssuesKeepNothingPrivate(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	w, _ := call(f.h, "GET", "/api/v1/issues?state=all", "", issuesManagement, false)
	_, pending := call(f.h, "GET", "/api/v1/decisions", "", issuesManagement, false)
	db, err := os.ReadFile(f.path + ".db")
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{w.Body.String(), fmt.Sprint(pending), string(db)} {
		for _, secret := range []string{"PRIVATE PROMPT TEXT", "provider secret detail", "does not exist", hostedKey, docsBotToken, otherAppToken, issuesManagement} {
			if strings.Contains(body, secret) {
				t.Fatalf("an Issue keeps %q", secret)
			}
		}
	}
	for _, token := range []string{docsBotToken, ""} {
		if w, _ := call(f.h, "GET", "/api/v1/issues", "", token, false); w.Code != 401 {
			t.Fatalf("issues read without the dashboard: %d", w.Code)
		}
	}
	if w, _ := call(f.h, "POST", "/api/v1/issues/"+f.only(t)["id"].(string), `{"action":"dismiss"}`, issuesManagement, false); w.Code != 403 {
		t.Fatalf("decided without the CSRF header: %d", w.Code)
	}
}

// With the adapting loop off no Issue is created or listed, and forwarding
// answers as before.
func TestNoIssuesWithAdaptingOff(t *testing.T) {
	f := newIssuesFixture(t, "adapting: {disabled: true}\n")
	if codes := f.chat(t, docsBotToken, "llama3:8b", 6); fmt.Sprint(codes) != "[404 404 404 404 404 404]" {
		t.Fatalf("forwarding: %v", codes)
	}
	f.detect()
	w, out := call(f.h, "GET", "/api/v1/issues?state=all", "", issuesManagement, false)
	if w.Code != 200 || out["adapting"] != false || len(out["issues"].([]any)) != 0 {
		t.Fatalf("issues with the loop off: %d %v", w.Code, out)
	}
	db, _ := os.ReadFile(f.path + ".db")
	if strings.Contains(string(db), "unknown_model") {
		t.Fatal("an Issue was kept with the loop off")
	}
}

// Resolved Issues are forgotten with request history; open ones are kept.
func TestResolvedIssuesFollowRequestHistoryRetention(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.chat(t, otherAppToken, "llama3:8b", 5)
	f.detect()
	f.hosted.set(`{"data":[{"id":"gpt-4o-mini"},{"id":"llama3:8b"}]}`, true)
	f.chat(t, docsBotToken, "llama3:8b", 1)
	f.detect()
	if list := f.issues(t, "?state=resolved"); len(list) != 1 || list[0]["application"] != "docs-bot" {
		t.Fatalf("resolved: %v", list)
	}
	if list := f.issues(t, "?application=other"); len(list) != 1 || list[0]["state"] != "open" {
		t.Fatalf("open for other: %v", list)
	}
	f.at = f.at.Add(8 * 24 * time.Hour) // retention_days defaults to 7
	f.h.Detect(f.at)
	if list := f.issues(t, "?state=all"); len(list) != 1 || list[0]["application"] != "other" {
		t.Fatalf("after the retention period: %v", list)
	}
}
