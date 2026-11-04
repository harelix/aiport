package server

import (
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/telemetry"
)

// The Owner reports a failed request from Activity to its Application's
// coding agent: it becomes an Issue of cause reported_request the agent
// reads with its own token, with the request's metadata and the Owner's
// note; another failure of the same kind adds to it, and the agent's next
// failing call of that kind names it in X-Aiport-Hint. The agent says it
// fixed it; the next successful call of that name resolves it.

func (f *issuesFixture) lastRequest(t *testing.T) telemetry.Event {
	t.Helper()
	rows := f.collector.Find(telemetry.Query{})
	if len(rows) == 0 {
		t.Fatal("no request in history")
	}
	return rows[0]
}

func (f *issuesFixture) report(t *testing.T, body string) (int, map[string]any) {
	t.Helper()
	w, out := call(f.h, "POST", "/api/v1/issues", body, issuesManagement, true)
	return w.Code, out
}

func TestTheOwnerReportsAFailedRequestToTheApplicationsCodingAgent(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "mistral:7b", 1) // the hosted default does not serve it: 404
	first := f.lastRequest(t)
	if first.Status != 404 {
		t.Fatalf("setup: want a 404, got %d", first.Status)
	}
	code, out := f.report(t, `{"request_id":"`+first.ID+`","note":"The app showed: model not found"}`)
	if code != 200 {
		t.Fatalf("report: %d %v", code, out)
	}
	is := m(out["issue"])
	id, _ := is["id"].(string)
	if is["cause"] != "reported_request" || is["application"] != "docs-bot" || is["state"] != "open" {
		t.Fatalf("issue: %v", is)
	}
	fix, _ := is["application_fix"].(string)
	for _, want := range []string{id, first.ID, "The app showed: model not found", "HTTP 404", `"mistral:7b"`, "aiport_issue_fixed"} {
		if !strings.Contains(fix, want) {
			t.Fatalf("application fix lacks %q:\n%s", want, fix)
		}
	}
	if strings.Contains(fix, "PRIVATE PROMPT TEXT") || strings.Contains(fix, hostedKey) {
		t.Fatalf("application fix leaks the prompt or a key:\n%s", fix)
	}

	mine := f.myIssues(t, docsBotToken)
	if len(mine) != 1 || mine[0]["id"] != id || m(mine[0]["report"])["note"] != "The app showed: model not found" {
		t.Fatalf("the agent's Issues: %v", mine)
	}
	if other := f.myIssues(t, otherAppToken); len(other) != 0 {
		t.Fatalf("another Application sees the report: %v", other)
	}

	// The next failure of that kind names the Issue and, reported, joins it.
	if code, hint := f.callHint(t, docsBotToken, "/v1/chat/completions", "mistral:7b"); code != 404 || !strings.Contains(hint, id) {
		t.Fatalf("hint on the same failure: %d %q", code, hint)
	}
	waitFor(t, func() bool { return f.lastRequest(t).ID != first.ID })
	second := f.lastRequest(t)
	if code, out := f.report(t, `{"request_id":"`+second.ID+`"}`); code != 200 || m(out["issue"])["id"] != id || m(m(out["issue"])["evidence"])["count"] != float64(2) || m(m(out["issue"])["report"])["note"] != "The app showed: model not found" {
		t.Fatalf("second report: %d %v", code, out)
	}
	if n := len(f.issues(t, "?state=all")); n != 1 {
		t.Fatalf("want one Issue for one kind of failure, got %d", n)
	}

	// The agent fixes it; the next successful call of that name resolves it.
	if w, out := call(f.h, "POST", IssuesPath+"/"+id+"/fixed", "", docsBotToken, false); w.Code != 200 || out["state"] != "resolving" {
		t.Fatalf("fixed: %d %v", w.Code, out)
	}
	f.hosted.set(`{"object":"list","data":[{"id":"mistral:7b"}]}`, true)
	if codes := f.chat(t, docsBotToken, "mistral:7b", 1); codes[0] != 200 {
		t.Fatalf("the fixed call: %v", codes)
	}
	f.detect()
	if st := f.only(t)["state"]; st != "resolved" {
		t.Fatalf("after a successful call the Issue is %v, want resolved", st)
	}
}

func TestOnlyAFailedRequestOfAnApplicationCanBeReported(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.hosted.set(`{"object":"list","data":[{"id":"gpt-4o-mini"}]}`, true)
	f.chat(t, docsBotToken, "gpt-4o-mini", 1)
	ok := f.lastRequest(t)
	if code, out := f.report(t, `{"request_id":"`+ok.ID+`"}`); code != 400 || !strings.Contains(m(out["error"])["message"].(string), "only a failed request") {
		t.Fatalf("reporting a success: %d %v", code, out)
	}
	if code, _ := f.report(t, `{"request_id":"no-such-request"}`); code != 404 {
		t.Fatalf("an unknown request: %d", code)
	}
	if code, _ := f.report(t, `{"request_id":"`+ok.ID+`","note":"`+strings.Repeat("x", 2001)+`"}`); code != 400 {
		t.Fatalf("a long note: %d", code)
	}
	if w, _ := call(f.h, "POST", "/api/v1/issues", `{"request_id":"`+ok.ID+`"}`, issuesManagement, false); w.Code != 403 {
		t.Fatalf("without the dashboard's CSRF header: %d", w.Code)
	}
}
