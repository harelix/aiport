package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// A failing call's X-Aiport-Hint names an Issue only when the call is one
// of that Issue's failures: same Application and, for an unknown model, the
// same requested model on a provider 404; for a failing Connection, a call
// to that Connection that failed with a 5xx, a timeout or a connection
// error. No other Issue is named, and aiport's own refusals name none
// unless the refusal is the Issue's cause.

// callHint sends one call for model on path and returns its status and
// X-Aiport-Hint.
func (f *issuesFixture) callHint(t *testing.T, token, path, model string) (int, string) {
	t.Helper()
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+path, strings.NewReader(`{"model":"`+model+`","messages":[{"role":"user","content":"hi"}]}`))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)
	return w.Code, w.Header().Get("X-Aiport-Hint")
}

func TestAnUnknownModelHintNamesTheIssueOnlyForThatModel(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.myIssues(t, docsBotToken)[0]["id"].(string)

	if code, hint := f.callHint(t, docsBotToken, "/v1/chat/completions", "llama3:8b"); code != 404 || !strings.Contains(hint, id) {
		t.Fatalf("a failing call for the Issue's model: %d %q", code, hint)
	}
	if code, hint := f.callHint(t, docsBotToken, "/v1/chat/completions", "mistral:7b"); code != 404 || strings.Contains(hint, "Issue") {
		t.Fatalf("a failing call for another model names an Issue: %d %q", code, hint)
	}
	// aiport's own refusal (no Connection for this protocol) is not the
	// Issue's cause.
	if code, hint := f.callHint(t, docsBotToken, "/v1/messages", "llama3:8b"); code != 503 || strings.Contains(hint, "Issue") {
		t.Fatalf("aiport's own refusal names an Issue: %d %q", code, hint)
	}
	if _, hint := f.callHint(t, otherAppToken, "/v1/chat/completions", "llama3:8b"); strings.Contains(hint, id) {
		t.Fatalf("another Application's call names docs-bot's Issue: %q", hint)
	}
}

func TestAFailingConnectionHintNamesTheIssueForCallsToThatConnection(t *testing.T) {
	f, flaky := failingFixture(t)
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)

	if code, hint := f.callHint(t, docsBotToken, "/v1/chat/completions", "llama3:8b"); code != 502 || !strings.Contains(hint, id) {
		t.Fatalf("a 502 from the failing Connection: %d %q", code, hint)
	}
	// The Connection cannot be reached at all: aiport's refusal is the cause.
	flaky.srv.Close()
	if code, hint := f.callHint(t, docsBotToken, "/v1/chat/completions", "llama3:8b"); code != 502 || !strings.Contains(hint, id) {
		t.Fatalf("an unreachable failing Connection: %d %q", code, hint)
	}
	// Another protocol, refused by aiport itself: not this Issue.
	if code, hint := f.callHint(t, docsBotToken, "/v1/messages", "claude-x"); code < 400 || strings.Contains(hint, "Issue") {
		t.Fatalf("an unrelated refusal names the Issue: %d %q", code, hint)
	}
	if code, hint := f.callHint(t, otherAppToken, "/v1/chat/completions", "llama3:8b"); code != 200 || hint != "" {
		t.Fatalf("another Application: %d %q", code, hint)
	}
}
