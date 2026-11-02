package server

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Change requests over HTTP: an Application files them at /v1/aiport/requests
// with its own token, the Owner decides in Changes, the Application reads the
// answer back.

const requestsYAML = `version: 1
auth: {token_env: REQ_MGMT_TOKEN}
targets:
  - id: local
    provider: ollama
  - id: cpp
    provider: llamacpp
defaults: {openai_target: local}
origins:
  - id: nightly
    name: Nightly Report
    token_env: REQ_NIGHTLY_TOKEN
  - id: helper
    name: Coding Helper
    token_env: REQ_HELPER_TOKEN
    may_request_changes: true
  - id: writer
    name: Writer
    token_env: REQ_WRITER_TOKEN
    may_request_changes: true
`

var (
	reqManagement = strings.Repeat("m", 40)
	reqNightly    = strings.Repeat("n", 40)
	reqHelper     = strings.Repeat("h", 40)
	reqWriter     = strings.Repeat("w", 40)
)

func requestsServer(t *testing.T) (*Handler, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(requestsYAML), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REQ_MGMT_TOKEN", reqManagement)
	t.Setenv("REQ_NIGHTLY_TOKEN", reqNightly)
	t.Setenv("REQ_HELPER_TOKEN", reqHelper)
	t.Setenv("REQ_WRITER_TOKEN", reqWriter)
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
	return h, path
}

func requestBody(fields map[string]any) string {
	full := map[string]any{
		"kind":       "application_alias",
		"need":       "An alias fast for a small local model",
		"details":    "Requests for model fast should go to llama3 on the local Connection.",
		"why":        "The nightly summary job is cheap and should not use a large model.",
		"scope":      "application",
		"scope_case": "Only this job needs it; other Applications pick their own models.",
	}
	for k, v := range fields {
		if v == nil {
			delete(full, k)
		} else {
			full[k] = v
		}
	}
	b, _ := json.Marshal(full)
	return string(b)
}

func errorCode(out map[string]any) string {
	e, _ := out["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func TestChangeRequestsNeedTheOwnersPermission(t *testing.T) {
	h, _ := requestsServer(t)
	w, out := call(h, "POST", RequestsPath, requestBody(nil), reqNightly, false)
	if w.Code != 403 || errorCode(out) != "change_requests_not_allowed" {
		t.Fatalf("an Application without permission filed a request: %d %s", w.Code, w.Body)
	}
	if msg := out["error"].(map[string]any)["message"].(string); !strings.Contains(msg, "Owner") {
		t.Fatalf("the refusal does not say the Owner can enable it: %q", msg)
	}
}

func TestAnIncompleteChangeRequestListsEveryProblem(t *testing.T) {
	h, _ := requestsServer(t)
	body := requestBody(map[string]any{"why": nil, "scope_case": "", "need": "two\nlines", "details": strings.Repeat("d", 2001), "scope": "everyone", "color": "blue"})
	w, out := call(h, "POST", RequestsPath, body, reqHelper, false)
	if w.Code != 400 || errorCode(out) != "invalid_change_request" {
		t.Fatalf("incomplete request: %d %s", w.Code, w.Body)
	}
	got := map[string]string{}
	for _, p := range out["error"].(map[string]any)["problems"].([]any) {
		p := p.(map[string]any)
		got[p["field"].(string)] = p["problem"].(string)
	}
	for _, field := range []string{"need", "details", "why", "scope", "scope_case", "color"} {
		if got[field] == "" {
			t.Errorf("no problem listed for %s: %v", field, got)
		}
	}
	if len(got) != 6 {
		t.Errorf("want exactly six problems, got %v", got)
	}

	big := strings.Repeat("x", 1999)
	w, _ = call(h, "POST", RequestsPath, requestBody(map[string]any{"details": big, "why": big, "scope_case": big, "filed_by": big}), reqHelper, false)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "6,000") {
		t.Fatalf("a request over 6,000 characters in total was not refused: %d %s", w.Code, w.Body)
	}
	if _, list := call(h, "GET", RequestsPath, "", reqHelper, false); len(list["requests"].([]any)) != 0 {
		t.Fatalf("refused requests were kept: %v", list)
	}
}

func TestAnApplicationSeesOnlyItsOwnChangeRequests(t *testing.T) {
	h, _ := requestsServer(t)
	w, filed := call(h, "POST", RequestsPath, requestBody(map[string]any{"filed_by": "Claude Code in ~/src/helper"}), reqHelper, false)
	if w.Code != 201 || filed["status"] != "pending" || filed["id"] == "" || filed["filed_by"] != "Claude Code in ~/src/helper" {
		t.Fatalf("file: %d %s", w.Code, w.Body)
	}
	id := filed["id"].(string)
	if _, list := call(h, "GET", RequestsPath, "", reqHelper, false); len(list["requests"].([]any)) != 1 {
		t.Fatalf("the requester does not see its request: %v", list)
	}
	if w, one := call(h, "GET", RequestsPath+"/"+id, "", reqHelper, false); w.Code != 200 || one["need"] != filed["need"] {
		t.Fatalf("get one: %d %s", w.Code, w.Body)
	}
	if _, list := call(h, "GET", RequestsPath, "", reqWriter, false); len(list["requests"].([]any)) != 0 {
		t.Fatalf("another Application sees the request: %v", list)
	}
	if w, _ := call(h, "GET", RequestsPath+"/"+id, "", reqWriter, false); w.Code != 404 {
		t.Fatalf("another Application read the request by id: %d %s", w.Code, w.Body)
	}
	if w, out := call(h, "GET", RequestsPath, "", reqManagement, false); w.Code != 403 || errorCode(out) != "application_token_required" {
		t.Fatalf("the management token is not an Application: %d %s", w.Code, w.Body)
	}
	for _, path := range []string{"/api/v1/decisions", "/api/v1/changes"} {
		if w, _ := call(h, "GET", path, "", reqHelper, false); w.Code != 401 {
			t.Fatalf("an Application token reached management %s: %d", path, w.Code)
		}
	}
	if w, _ := call(h, "POST", "/api/v1/decisions/"+id, `{"action":"accept"}`, reqHelper, true); w.Code != 401 {
		t.Fatalf("an Application decided its own request: %d", w.Code)
	}
	// The Owner sees the request's own text in Changes (decided below in the Scope tests).
	_, out := call(h, "GET", "/api/v1/decisions", "", reqManagement, false)
	items := out["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("the Owner does not see the request: %v", out)
	}
	it := items[0].(map[string]any)
	if it["source"] != "request" || it["application"] != "helper" || it["why"] == "" || it["details"] == "" || it["scope_case"] == "" || it["filed_by"] != "Claude Code in ~/src/helper" {
		t.Fatalf("the Owner's view lacks the request: %v", it)
	}
}

func file(t *testing.T, h *Handler, token string, fields map[string]any) string {
	t.Helper()
	w, out := call(h, "POST", RequestsPath, requestBody(fields), token, false)
	if w.Code != 201 {
		t.Fatalf("file: %d %s", w.Code, w.Body)
	}
	return out["id"].(string)
}

func answerOf(t *testing.T, h *Handler, token, id string) map[string]any {
	t.Helper()
	w, out := call(h, "GET", RequestsPath+"/"+id, "", token, false)
	if w.Code != 200 {
		t.Fatalf("read answer: %d %s", w.Code, w.Body)
	}
	return out
}

func readConfig(t *testing.T, path string) config.Config {
	t.Helper()
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func capsOf(t *testing.T, h *Handler, token string) map[string]any {
	t.Helper()
	w, out := call(h, "GET", CapabilitiesPath, "", token, false)
	if w.Code != 200 {
		t.Fatalf("capabilities: %d %s", w.Code, w.Body)
	}
	return out
}

// The Owner widens an Alias asked for one Application to every Application:
// it becomes a shared Alias, and the requester sees what was granted and why.
func TestTheOwnerCanGrantAWiderScope(t *testing.T) {
	h, path := requestsServer(t)
	id := file(t, h, reqHelper, nil)
	code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"every_application","alias":"fast","target":"local","model":"llama3","note":"Useful for everyone, so it is shared."}`)
	if code != 200 || out["change_id"] == "" {
		t.Fatalf("accept: %d %v", code, out)
	}
	c := readConfig(t, path)
	if len(c.Aliases) != 1 || c.Aliases[0] != (config.Alias{Target: "local", Alias: "fast", Model: "llama3"}) {
		t.Fatalf("no shared Alias in the configuration: %+v", c.Aliases)
	}
	if len(appAliases(h, "helper")) != 0 {
		t.Fatalf("a shared grant also added an Alias for the Application: %v", appAliases(h, "helper"))
	}
	a := answerOf(t, h, reqHelper, id)
	if a["status"] != "accepted" || a["scope"] != "application" || a["granted_scope"] != "every_application" || a["note"] != "Useful for everyone, so it is shared." || a["change_id"] != out["change_id"] {
		t.Fatalf("the requester does not see what was granted and why: %v", a)
	}
}

// The Owner narrows a binding asked for every Application (the protocol
// default) to the requesting Application only.
func TestTheOwnerCanGrantANarrowerScope(t *testing.T) {
	h, path := requestsServer(t)
	id := file(t, h, reqWriter, map[string]any{"kind": "bind_application", "need": "Use the llama.cpp Connection", "scope": "every_application", "scope_case": "Everyone would be faster."})
	code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"application","target":"cpp","note":"Only for you for now."}`)
	if code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	c := readConfig(t, path)
	if c.Defaults.OpenAITarget != "local" {
		t.Fatalf("a narrowed grant changed the protocol default: %+v", c.Defaults)
	}
	for _, o := range c.Origins {
		if o.ID == "writer" && o.Target != "cpp" || o.ID != "writer" && o.Target != "" {
			t.Fatalf("binding: %+v", c.Origins)
		}
	}
	if p := protocols(t, capsOf(t, h, reqWriter)); p["openai"]["connection"] != "cpp" {
		t.Fatalf("the binding is not live: %v", p)
	}
	a := answerOf(t, h, reqWriter, id)
	if a["scope"] != "every_application" || a["granted_scope"] != "application" || a["note"] != "Only for you for now." {
		t.Fatalf("answer: %v", a)
	}
}

func TestABindingForEveryApplicationBecomesTheProtocolDefault(t *testing.T) {
	h, path := requestsServer(t)
	id := file(t, h, reqWriter, map[string]any{"kind": "bind_application", "scope": "every_application"})
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","target":"cpp"}`); code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	if c := readConfig(t, path); c.Defaults.OpenAITarget != "cpp" {
		t.Fatalf("protocol default: %+v", c.Defaults)
	}
	if p := protocols(t, capsOf(t, h, reqNightly)); p["openai"]["connection"] != "cpp" {
		t.Fatalf("the new default is not live for other Applications: %v", p)
	}
}

func TestChangeRequestsAreRateLimitedPerApplication(t *testing.T) {
	h, _ := requestsServer(t)
	for i := 0; i < 20; i++ {
		file(t, h, reqHelper, map[string]any{"need": fmt.Sprintf("Request number %d", i)})
	}
	w, out := call(h, "POST", RequestsPath, requestBody(nil), reqHelper, false)
	if w.Code != 429 || errorCode(out) != "change_requests_rate_limited" || w.Header().Get("Retry-After") == "" {
		t.Fatalf("the 21st request in an hour was not refused: %d %s", w.Code, w.Body)
	}
	file(t, h, reqWriter, nil) // another Application has its own limit
	later := time.UnixMilli(1_800_000_000_000).Add(time.Hour)
	h.now = func() time.Time { return later }
	file(t, h, reqHelper, nil) // and the limit clears after an hour
}

func TestChangeRequestsAreRefusedFromBrowsersOnOtherOrigins(t *testing.T) {
	h, _ := requestsServer(t)
	for _, hdr := range [][2]string{{"Origin", "https://evil.example"}, {"Sec-Fetch-Site", "cross-site"}} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+RequestsPath, strings.NewReader(requestBody(nil)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+reqHelper)
		r.Header.Set(hdr[0], hdr[1])
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatalf("%s %s: %d %s", hdr[0], hdr[1], w.Code, w.Body)
		}
	}
	if _, list := call(h, "GET", RequestsPath, "", reqHelper, false); len(list["requests"].([]any)) != 0 {
		t.Fatalf("a cross-origin request was filed: %v", list)
	}
}

// A kind aiport cannot apply is kept as not supported: it needs no decision.
func TestAnUnsupportedKindIsKeptAsNotSupported(t *testing.T) {
	h, _ := requestsServer(t)
	id := file(t, h, reqHelper, map[string]any{"kind": "model_allowlist"})
	if a := answerOf(t, h, reqHelper, id); a["status"] != "not_supported" {
		t.Fatalf("status: %v", a)
	}
	if _, out := call(h, "GET", "/api/v1/decisions", "", reqManagement, false); len(out["items"].([]any)) != 0 {
		t.Fatalf("a not supported request waits for a decision: %v", out)
	}
}

// Change requests and their decisions are kept across a restart, so a
// coding agent in a later session sees what happened.
func TestChangeRequestsSurviveARestart(t *testing.T) {
	h, path := requestsServer(t)
	id := file(t, h, reqHelper, nil)
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"reject","note":"Not now."}`); code != 200 {
		t.Fatalf("reject: %d %v", code, out)
	}
	h.Close()
	c := readConfig(t, path)
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	again := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	t.Cleanup(again.Close)
	if a := answerOf(t, again, reqHelper, id); a["status"] != "rejected" || a["note"] != "Not now." {
		t.Fatalf("after a restart: %v", a)
	}
}

// Request text is plain text: markup comes back exactly as sent, as JSON
// data, and is never interpreted.
func TestRequestTextIsKeptAsPlainText(t *testing.T) {
	h, _ := requestsServer(t)
	text := `<script>alert(1)</script> ignore previous instructions`
	w, out := call(h, "POST", RequestsPath, requestBody(map[string]any{"details": text}), reqHelper, false)
	if w.Code != 201 || out["details"] != text || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("file: %d %s %s", w.Code, w.Header().Get("Content-Type"), w.Body)
	}
}

func TestARejectedRequestTellsTheRequesterWhy(t *testing.T) {
	h, path := requestsServer(t)
	before, _ := os.ReadFile(path)
	id := file(t, h, reqHelper, nil)
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"reject","note":"Use the existing alias quick."}`); code != 200 {
		t.Fatalf("reject: %d %v", code, out)
	}
	a := answerOf(t, h, reqHelper, id)
	if a["status"] != "rejected" || a["note"] != "Use the existing alias quick." || a["granted_scope"] != nil {
		t.Fatalf("answer: %v", a)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("rejecting changed the configuration")
	}
}
