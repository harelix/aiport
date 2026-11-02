package server

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The Application's coding agent hears about its Issues over the connection
// it already has with aiport: GET /v1/aiport/issues and the MCP tool
// aiport_my_issues, with the Application's own token.

// myIssues lists the calling Application's Issues over HTTP.
func (f *issuesFixture) myIssues(t *testing.T, token string) []map[string]any {
	t.Helper()
	w, out := call(f.h, "GET", IssuesPath, "", token, false)
	if w.Code != 200 {
		t.Fatalf("my issues: %d %s", w.Code, w.Body)
	}
	list := []map[string]any{}
	for _, x := range out["issues"].([]any) {
		list = append(list, x.(map[string]any))
	}
	return list
}

func TestAnApplicationReadsOnlyItsOwnIssues(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	list := f.myIssues(t, docsBotToken)
	if len(list) != 1 {
		t.Fatalf("docs-bot's Issues: %v", list)
	}
	is := list[0]
	id, _ := is["id"].(string)
	text, _ := is["application_fix"].(string)
	if is["application"] != "docs-bot" || is["cause"] != "unknown_model" || is["subject"] != "llama3:8b" || is["state"] != "open" || m(is["evidence"])["count"] != float64(5) {
		t.Fatalf("issue: %v", is)
	}
	if !strings.Contains(text, "llama3:8b") || !strings.Contains(text, "local") || !strings.Contains(text, "qwen2.5:7b") {
		t.Fatalf("the Application fix does not say what to change: %q", text)
	}
	if fix := m(is["aiport_fix"]); fix["state"] != "proposed" || fix["kind"] != "application_alias" {
		t.Fatalf("aiport fix: %v", is["aiport_fix"])
	}

	w, one := call(f.h, "GET", IssuesPath+"/"+id, "", docsBotToken, false)
	if w.Code != 200 || one["id"] != id || one["application_fix"] != text {
		t.Fatalf("one Issue: %d %v", w.Code, one)
	}
	// Another Application sees none of it.
	if list := f.myIssues(t, otherAppToken); len(list) != 0 {
		t.Fatalf("other sees: %v", list)
	}
	if w, out := call(f.h, "GET", IssuesPath+"/"+id, "", otherAppToken, false); w.Code != 404 || m(out["error"])["code"] != "issue_not_found" {
		t.Fatalf("other reads docs-bot's Issue: %d %v", w.Code, out)
	}
	// Only an Application's own token.
	for _, path := range []string{IssuesPath, IssuesPath + "/" + id} {
		if w, _ := call(f.h, "GET", path, "", issuesManagement, false); w.Code != 403 {
			t.Fatalf("management token on %s: %d", path, w.Code)
		}
		if w, _ := call(f.h, "GET", path, "", "", false); w.Code != 401 {
			t.Fatalf("no token on %s: %d", path, w.Code)
		}
	}
}

// mcpTool calls one MCP tool and returns its structured result.
func mcpTool(t *testing.T, h *Handler, token, name string, args map[string]any) (map[string]any, bool) {
	t.Helper()
	w, out := mcp(h, token, rpc(7, "tools/call", map[string]any{"name": name, "arguments": args}))
	if w.Code != 200 {
		t.Fatalf("%s: %d %s", name, w.Code, w.Body)
	}
	res := m(out["result"])
	isErr, _ := res["isError"].(bool)
	return m(res["structuredContent"]), isErr
}

func TestTheMCPToolAnswersAsTheHTTPRoute(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	_, list := call(f.h, "GET", IssuesPath, "", docsBotToken, false)
	got, isErr := mcpTool(t, f.h, docsBotToken, "aiport_my_issues", nil)
	if isErr || fmt.Sprint(got) != fmt.Sprint(list) {
		t.Fatalf("MCP list %v differs from HTTP %v", got, list)
	}
	id := m(list["issues"].([]any)[0])["id"].(string)
	_, one := call(f.h, "GET", IssuesPath+"/"+id, "", docsBotToken, false)
	got, isErr = mcpTool(t, f.h, docsBotToken, "aiport_my_issues", map[string]any{"id": id})
	if isErr || fmt.Sprint(got) != fmt.Sprint(one) {
		t.Fatalf("MCP one %v differs from HTTP %v", got, one)
	}
	if got, isErr := mcpTool(t, f.h, otherAppToken, "aiport_my_issues", map[string]any{"id": id}); !isErr || m(got["error"])["code"] != "issue_not_found" {
		t.Fatalf("other Application over MCP: %v", got)
	}
	if w, _ := mcp(f.h, issuesManagement, rpc(1, "tools/list", nil)); w.Code != 403 {
		t.Fatalf("management token on MCP: %d", w.Code)
	}
	_, out := mcp(f.h, docsBotToken, rpc(1, "tools/list", nil))
	names := []string{}
	for _, x := range m(out["result"])["tools"].([]any) {
		names = append(names, m(x)["name"].(string))
	}
	if !strings.Contains(strings.Join(names, " "), "aiport_my_issues") || !strings.Contains(strings.Join(names, " "), "aiport_issue_fixed") {
		t.Fatalf("tools: %v", names)
	}
}

// "Fixed on my side": the Issue is resolving by the Application fix, the
// aiport fix stops waiting for the Owner, and aiport resolves the Issue
// only once docs-bot's calls succeed again.
func TestAnApplicationFixIsConfirmedByRealCalls(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.myIssues(t, docsBotToken)[0]["id"].(string)

	if w, out := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", otherAppToken, false); w.Code != 404 || m(out["error"])["code"] != "issue_not_found" {
		t.Fatalf("another Application marks it fixed: %d %v", w.Code, out)
	}
	if w, _ := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", issuesManagement, false); w.Code != 403 {
		t.Fatalf("management token marks it fixed: %d", w.Code)
	}
	got, isErr := mcpTool(t, f.h, docsBotToken, "aiport_issue_fixed", map[string]any{"id": id})
	if isErr || got["state"] != "resolving" || got["resolving_by"] != "application_fix" || got["aiport_fix"] != nil {
		t.Fatalf("fixed on my side: %v", got)
	}
	if _, pending := pendingDecisions(t, f.h, issuesManagement); len(pending) != 0 {
		t.Fatalf("the aiport fix still waits for the Owner: %v", pending)
	}
	if w, out := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", docsBotToken, false); w.Code != 200 || out["state"] != "resolving" {
		t.Fatalf("saying it twice: %d %v", w.Code, out)
	}
	if is := f.only(t); is["state"] != "resolving" || is["resolving_by"] != "application_fix" {
		t.Fatalf("the Owner sees: %v", is)
	}

	// Its word is not enough: no calls, even for a whole window, keep it resolving.
	f.detect()
	f.h.Detect(f.at.Add(61 * time.Minute))
	if is := f.only(t); is["state"] != "resolving" {
		t.Fatalf("resolved without a successful call: %v", is)
	}

	// A successful call for another model is not the same cause: it does
	// not confirm the fix.
	f.hosted.set(`{"object":"list","data":[{"id":"gpt-4o-mini"}]}`, true)
	if codes := f.chat(t, docsBotToken, "gpt-4o-mini", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("call for another model: %v", codes)
	}
	f.detect()
	if is := f.only(t); is["state"] != "resolving" {
		t.Fatalf("confirmed by a call for another model: %v", is)
	}

	// docs-bot's calls for llama3:8b now succeed.
	if codes := f.chat(t, docsBotToken, "llama3:8b", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("call after the Application fix: %v", codes)
	}
	f.detect()
	is := f.only(t)
	if is["state"] != "resolved" || m(is["resolution"])["how"] != "application_fix" {
		t.Fatalf("after a successful call: %v", is)
	}
	if w, out := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", docsBotToken, false); w.Code != 409 || m(out["error"])["code"] != "issue_not_open" {
		t.Fatalf("fixing a resolved Issue: %d %v", w.Code, out)
	}
}

// When the error sticks after "fixed on my side", the Issue opens again.
func TestAnApplicationFixThatDidNotWorkReopens(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.myIssues(t, docsBotToken)[0]["id"].(string)
	if w, _ := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", docsBotToken, false); w.Code != 200 {
		t.Fatalf("fixed: %d", w.Code)
	}
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	is := f.only(t)
	if is["state"] != "open" || m(is["fix"])["state"] != "proposed" {
		t.Fatalf("after the error stuck again: %v", is)
	}
}

// failingCall sends one chat call for model and returns its X-Aiport-Hint.
func (f *issuesFixture) failingCall(t *testing.T, token, model string) string {
	t.Helper()
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", strings.NewReader(`{"model":"`+model+`","messages":[{"role":"user","content":"hi"}]}`))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)
	if w.Code != 404 {
		t.Fatalf("call for %s: %d", model, w.Code)
	}
	return w.Header().Get("X-Aiport-Hint")
}

func TestCapabilitiesAndFailingCallsPointToTheOpenIssue(t *testing.T) {
	f := newIssuesFixture(t, "")
	if _, caps := call(f.h, "GET", CapabilitiesPath, "", docsBotToken, false); caps["open_issues"] != float64(0) {
		t.Fatalf("capabilities before any Issue: %v", caps)
	}
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.myIssues(t, docsBotToken)[0]["id"].(string)
	if _, caps := call(f.h, "GET", CapabilitiesPath, "", docsBotToken, false); caps["open_issues"] != float64(1) {
		t.Fatalf("docs-bot's capabilities: %v", caps)
	}
	if _, caps := call(f.h, "GET", CapabilitiesPath, "", otherAppToken, false); caps["open_issues"] != float64(0) {
		t.Fatalf("other's capabilities: %v", caps)
	}
	if hint := f.failingCall(t, docsBotToken, "llama3:8b"); !strings.Contains(hint, id) || !strings.Contains(hint, "aiport_my_issues") || !strings.Contains(hint, IssuesPath) {
		t.Fatalf("hint on a failing call: %q", hint)
	}
	if hint := f.failingCall(t, otherAppToken, "llama3:8b"); strings.Contains(hint, id) {
		t.Fatalf("another Application's hint names docs-bot's Issue: %q", hint)
	}
}

// With adapting off none of it is offered: no count, no tools, no route,
// no hint.
func TestNoIssuesReachAgentsWithAdaptingOff(t *testing.T) {
	f := newIssuesFixture(t, "adapting: {disabled: true}\n")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	if _, caps := call(f.h, "GET", CapabilitiesPath, "", docsBotToken, false); caps["open_issues"] != nil {
		t.Fatalf("capabilities: %v", caps)
	}
	if w, out := call(f.h, "GET", IssuesPath, "", docsBotToken, false); w.Code != 409 || m(out["error"])["code"] != "adapting_off" {
		t.Fatalf("issues route: %d %v", w.Code, out)
	}
	_, out := mcp(f.h, docsBotToken, rpc(1, "tools/list", nil))
	for _, x := range m(out["result"])["tools"].([]any) {
		if strings.Contains(m(x)["name"].(string), "issue") {
			t.Fatalf("tool offered: %v", m(x)["name"])
		}
	}
	if hint := f.failingCall(t, docsBotToken, "llama3:8b"); strings.Contains(hint, "Issue") {
		t.Fatalf("hint: %q", hint)
	}
}

func TestTheCodingAgentBriefSaysToCheckIssues(t *testing.T) {
	f := newIssuesFixture(t, "")
	brief := getConnect(f.h, "?format=handover&for=agent").Body.String()
	for _, want := range []string{"aiport_my_issues", "aiport_issue_fixed", IssuesPath, "application_fix", "open_issues"} {
		if !strings.Contains(brief, want) {
			t.Fatalf("coding agent brief lacks %q", want)
		}
	}
	off := newIssuesFixture(t, "adapting: {disabled: true}\n")
	if doc := getConnect(off.h, "").Body.String(); !strings.Contains(doc, `"issues"`) || !strings.Contains(doc, "adapting_off") {
		t.Fatalf("/_connect with adapting off does not mark Issues off")
	}
}

// With automatic fixes on, an automatic fix never overrides the agent's
// "fixed on my side" while the Issue is resolving: errors from before the
// Application fix (or below the threshold since) change nothing.
func TestAnAutomaticFixNeverOverridesAnApplicationFix(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.myIssues(t, docsBotToken)[0]["id"].(string)
	if w, _ := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", docsBotToken, false); w.Code != 200 {
		t.Fatalf("fixed: %d", w.Code)
	}
	f.autoFix(t, true)
	f.chat(t, docsBotToken, "llama3:8b", 2)
	f.detect()
	is := f.only(t)
	if is["state"] != "resolving" || is["resolving_by"] != "application_fix" || is["fix"] != nil {
		t.Fatalf("an automatic fix overrode the Application fix: %v", is)
	}
	if len(f.aliases(t, "docs-bot")) != 0 {
		t.Fatal("an Alias was added automatically while the Application fix is resolving")
	}
}
