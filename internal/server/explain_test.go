package server

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"aiport/internal/config"
)

// Explain without a coding agent: a Connection and model the Owner chose in
// Settings write a short root cause of an Issue from its evidence alone,
// streamed to the dashboard and saved with the Issue.

const writerKey = "sk-writer-key-of-the-explaining-connection"

// fakeWriter is the Connection that writes explanations. It streams its
// answer in chunks (OpenAI or Anthropic server-sent events, by path); with
// hold set it waits for release between the first chunk and the rest.
// status other than 0 answers that error instead.
type fakeWriter struct {
	mu      sync.Mutex
	chunks  []string
	status  int
	hold    bool
	release chan struct{}
	bodies  []string
	auth    []string
	srv     *httptest.Server
}

func newFakeWriter(t *testing.T, chunks ...string) *fakeWriter {
	f := &fakeWriter{chunks: chunks, release: make(chan struct{})}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" { // model listings during detection
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"writer-model"}]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.bodies = append(f.bodies, string(body))
		f.auth = append(f.auth, r.Header.Get("Authorization")+"|"+r.Header.Get("X-Api-Key"))
		chunks, status, hold := f.chunks, f.status, f.hold
		f.mu.Unlock()
		if status != 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":{"type":"authentication_error","message":"bad key (provider secret detail)"}}`))
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		anthropic := strings.HasSuffix(r.URL.Path, "/messages")
		for i, c := range chunks {
			text, _ := json.Marshal(c)
			if anthropic {
				_, _ = io.WriteString(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":"+string(text)+"}}\n\n")
			} else {
				_, _ = io.WriteString(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":"+string(text)+"}}]}\n\n")
			}
			w.(http.Flusher).Flush()
			if i == 0 && hold {
				<-f.release
			}
		}
		if anthropic {
			_, _ = io.WriteString(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		} else {
			_, _ = io.WriteString(w, "data: [DONE]\n\n")
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeWriter) set(status int, chunks ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.chunks = status, chunks
}

func (f *fakeWriter) requests() ([]string, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.bodies...), append([]string(nil), f.auth...)
}

// newExplainFixture is the docs-bot example with two more Connections that
// can write explanations: writer (OpenAI protocol, with a key) and
// claude-writer (Anthropic protocol).
func newExplainFixture(t *testing.T, writer *fakeWriter, extra string) *issuesFixture {
	t.Helper()
	f := &issuesFixture{path: filepath.Join(t.TempDir(), "aiport.yaml"), at: time.Now(), extra: extra}
	f.local = newFakeConnection(t, `{"object":"list","data":[{"id":"llama3:8b"},{"id":"qwen2.5:7b"}]}`, true)
	f.hosted = newFakeConnection(t, `{"object":"list","data":[{"id":"gpt-4o-mini"}]}`, false)
	yaml := "version: 1\nauth: {token_env: ISSUES_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: local, provider: ollama, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + f.hosted.srv.URL + "/v1', api_key_env: ISSUES_HOSTED_KEY}\n" +
		"  - {id: writer, provider: acme, protocol: openai, base_url: '" + writer.srv.URL + "/v1', api_key_env: ISSUES_WRITER_KEY}\n" +
		"  - {id: claude-writer, provider: acme, protocol: anthropic, base_url: '" + writer.srv.URL + "/v1'}\n" +
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
	t.Setenv("ISSUES_WRITER_KEY", writerKey)
	f.start(t)
	return f
}

// explainEvents posts Explain for Issue id and returns the streamed events,
// read whole.
func (f *issuesFixture) explainEvents(t *testing.T, id string) (int, []map[string]any) {
	t.Helper()
	w, _ := call(f.h, "POST", "/api/v1/issues/"+id+"/explain", "", issuesManagement, true)
	return w.Code, events(t, w.Body.String())
}

func events(t *testing.T, body string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(body), "\n") {
		e := map[string]any{}
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("not one JSON event per line: %q", body)
		}
		out = append(out, e)
	}
	return out
}

func (f *issuesFixture) chooseWriter(t *testing.T, connection, model string) (int, map[string]any) {
	t.Helper()
	w, out := call(f.h, "PUT", "/api/v1/adapting/explainer", `{"connection":"`+connection+`","model":"`+model+`"}`, issuesManagement, true)
	return w.Code, out
}

func TestExplainWaitsForTheOwnerToChooseAConnection(t *testing.T) {
	writer := newFakeWriter(t, "unused")
	f := newExplainFixture(t, writer, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)

	_, out := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if ex := m(out["explain"]); ex["available"] != false || !strings.Contains(ex["hint"].(string), "Settings") {
		t.Fatalf("Explain before a Connection is chosen: %v", out["explain"])
	}
	w, body := call(f.h, "POST", "/api/v1/issues/"+id+"/explain", "", issuesManagement, true)
	if w.Code != 409 || m(body["error"])["code"] != "explainer_not_chosen" {
		t.Fatalf("Explain without a chosen Connection: %d %s", w.Code, w.Body)
	}
	if w, _ := call(f.h, "POST", "/api/v1/issues/"+id+"/explain", "", issuesManagement, false); w.Code != 403 {
		t.Fatalf("Explain without the dashboard's CSRF header: %d", w.Code)
	}
	if code, out := f.chooseWriter(t, "nowhere", "m"); code != 400 {
		t.Fatalf("an unknown Connection was accepted: %d %v", code, out)
	}
	code, out := f.chooseWriter(t, "writer", "writer-model")
	if code != 200 || out["change_id"] == "" || m(out["explainer"])["connection"] != "writer" {
		t.Fatalf("choose the writer: %d %v", code, out)
	}
	_, out = call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if ex := m(out["explain"]); ex["available"] != true || ex["connection"] != "writer" || ex["model"] != "writer-model" {
		t.Fatalf("Explain once chosen: %v", out["explain"])
	}
	if c := liveConfig(t, f.h, issuesManagement); c.Adapting.Explain.Connection != "writer" || c.Adapting.Explain.Model != "writer-model" {
		t.Fatalf("not in the configuration: %+v", c.Adapting.Explain)
	}
	// Clearing it turns Explain off again.
	if code, _ := f.chooseWriter(t, "", ""); code != 200 {
		t.Fatalf("clear: %d", code)
	}
	if w, _ := call(f.h, "POST", "/api/v1/issues/"+id+"/explain", "", issuesManagement, true); w.Code != 409 {
		t.Fatalf("Explain after clearing: %d", w.Code)
	}
	if bodies, _ := writer.requests(); len(bodies) != 0 {
		t.Fatal("the writer was asked before it was chosen")
	}
}

// The docs-bot example: Explain streams the writer's words to the
// dashboard as they come, saves them with the Issue, and both the Owner's
// copy of the Application fix and the Application's own agent get them.
// The writer gets only the Issue's evidence.
func TestExplainStreamsAndSavesAnExplanationFromTheEvidence(t *testing.T) {
	writer := newFakeWriter(t, "The hosted Connection ", "does not serve llama3:8b; local does.")
	writer.mu.Lock()
	writer.hold = true
	writer.mu.Unlock()
	f := newExplainFixture(t, writer, "adapting: {explain: {connection: writer, model: writer-model}}\n")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	is := f.only(t)
	id := is["id"].(string)

	srv := httptest.NewServer(f.h)
	defer srv.Close()
	req, _ := http.NewRequest("POST", srv.URL+"/api/v1/issues/"+id+"/explain", nil)
	req.Host = "127.0.0.1:8787"
	req.Header.Set("Authorization", "Bearer "+issuesManagement)
	req.Header.Set("X-Aiport-CSRF", "1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/x-ndjson") {
		t.Fatalf("explain: %d %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	lines := bufio.NewScanner(resp.Body)
	next := func() map[string]any {
		if !lines.Scan() {
			t.Fatalf("the stream ended early: %v", lines.Err())
		}
		e := map[string]any{}
		if err := json.Unmarshal(lines.Bytes(), &e); err != nil {
			t.Fatalf("not a JSON event: %q", lines.Text())
		}
		return e
	}
	start := next()
	if start["type"] != "start" || m(start["by"])["source"] != "connection" || m(start["by"])["connection"] != "writer" || m(start["by"])["model"] != "writer-model" {
		t.Fatalf("first event: %v", start)
	}
	// The first words arrive while the writer is still writing.
	if e := next(); e["type"] != "text" || e["text"] != "The hosted Connection " {
		t.Fatalf("first words: %v", e)
	}
	close(writer.release)
	text := "The hosted Connection "
	var done map[string]any
	for done == nil {
		switch e := next(); e["type"] {
		case "text":
			text += e["text"].(string)
		case "done":
			done = e
		default:
			t.Fatalf("unexpected event: %v", e)
		}
	}
	const want = "The hosted Connection does not serve llama3:8b; local does."
	ex := m(m(done["issue"])["explanation"])
	if text != want || ex["text"] != want || ex["source"] != "connection" || ex["connection"] != "writer" || ex["model"] != "writer-model" {
		t.Fatalf("explanation: streamed %q, saved %v", text, ex)
	}

	// What the writer was sent: the evidence, with the writer's own key only.
	bodies, auth := writer.requests()
	if len(bodies) != 1 || auth[0] != "Bearer "+writerKey+"|" {
		t.Fatalf("writer requests: %d, auth %v", len(bodies), auth)
	}
	var sent struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	_ = json.Unmarshal([]byte(bodies[0]), &sent)
	ev := m(is["evidence"])
	example := ev["examples"].([]any)[0].(string)
	if sent.Model != "writer-model" || !sent.Stream {
		t.Fatalf("writer request: %s", bodies[0])
	}
	for _, s := range []string{"llama3:8b", "404", "model_not_found", example, "hosted", "local", "qwen2.5:7b", strconv.Itoa(5)} {
		if !strings.Contains(bodies[0], s) {
			t.Fatalf("the writer was not given %q: %s", s, bodies[0])
		}
	}
	for _, s := range []string{"PRIVATE PROMPT TEXT", "provider secret detail", docsBotToken, otherAppToken, hostedKey, writerKey, issuesManagement, "docs-bot", "Docs Bot"} {
		if strings.Contains(bodies[0], s) {
			t.Fatalf("the writer was sent %q: %s", s, bodies[0])
		}
	}

	// Saved with the Issue, and carried by the Application fix.
	is = f.only(t)
	fix, _ := is["application_fix"].(string)
	if m(is["explanation"])["text"] != want || !strings.Contains(fix, want) {
		t.Fatalf("not saved with the Issue: %v / %q", is["explanation"], fix)
	}
	mine := f.myIssues(t, docsBotToken)
	if len(mine) != 1 || m(mine[0]["explanation"])["text"] != want || mine[0]["application_fix"] != fix {
		t.Fatalf("the Application's agent does not get it: %v", mine)
	}
	got, isErr := mcpTool(t, f.h, docsBotToken, "aiport_my_issues", map[string]any{"id": id})
	if isErr || m(got["explanation"])["text"] != want || got["application_fix"] != fix {
		t.Fatalf("aiport_my_issues: %v", got)
	}

	// The call shows in Activity under the built-in aiport Application.
	from, to := time.Now().Add(-time.Hour).UnixMilli(), time.Now().Add(time.Hour).UnixMilli()
	_, act := call(f.h, "GET", "/api/v1/requests?origin=aiport&from_ms="+strconv.FormatInt(from, 10)+"&to_ms="+strconv.FormatInt(to, 10), "", issuesManagement, false)
	rows, _ := act["requests"].([]any)
	if len(rows) != 1 {
		t.Fatalf("Activity under aiport: %v", act)
	}
	if r := m(rows[0]); r["target_id"] != "writer" || r["requested_model"] != "writer-model" || r["http_status"] != float64(200) || r["outcome"] != "success" || r["route"] != "/v1/chat/completions" {
		t.Fatalf("Activity row: %v", r)
	}

	// Explain again replaces it; it survives a restart.
	writer.set(0, "Second look: ", "send a model hosted serves.")
	code, evs := f.explainEvents(t, id)
	if code != 200 || evs[len(evs)-1]["type"] != "done" {
		t.Fatalf("explain again: %d %v", code, evs)
	}
	f.h.Close()
	f.start(t)
	is = f.only(t)
	if m(is["explanation"])["text"] != "Second look: send a model hosted serves." || strings.Contains(is["application_fix"].(string), want) {
		t.Fatalf("after Explain again and a restart: %v", is["explanation"])
	}
}

// An Anthropic-protocol Connection writes explanations too.
func TestExplainWithAnAnthropicConnection(t *testing.T) {
	writer := newFakeWriter(t, "Root cause: ", "no Alias for llama3:8b.")
	f := newExplainFixture(t, writer, "adapting: {explain: {connection: claude-writer, model: claude-model}}\n")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	code, evs := f.explainEvents(t, f.only(t)["id"].(string))
	if code != 200 || evs[len(evs)-1]["type"] != "done" || m(m(evs[len(evs)-1]["issue"])["explanation"])["text"] != "Root cause: no Alias for llama3:8b." {
		t.Fatalf("anthropic explain: %d %v", code, evs)
	}
	bodies, _ := writer.requests()
	var sent struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		Stream    bool   `json:"stream"`
	}
	_ = json.Unmarshal([]byte(bodies[0]), &sent)
	if sent.Model != "claude-model" || sent.MaxTokens < 1 || !sent.Stream {
		t.Fatalf("anthropic request: %s", bodies[0])
	}
}

// A writer that fails ends the stream with aiport's own words (never the
// provider's message), keeps the saved explanation, and shows in Activity.
func TestAFailingWriterKeepsTheSavedExplanation(t *testing.T) {
	writer := newFakeWriter(t, "First explanation.")
	f := newExplainFixture(t, writer, "adapting: {explain: {connection: writer, model: writer-model}}\n")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)
	if code, _ := f.explainEvents(t, id); code != 200 {
		t.Fatal("first explain failed")
	}
	writer.set(401)
	code, evs := f.explainEvents(t, id)
	last := evs[len(evs)-1]
	msg, _ := m(last["error"])["message"].(string)
	if code != 200 || last["type"] != "error" || m(last["error"])["code"] != "explanation_failed" || !strings.Contains(msg, "401") || strings.Contains(msg, "secret") {
		t.Fatalf("failing writer: %d %v", code, evs)
	}
	if is := f.only(t); m(is["explanation"])["text"] != "First explanation." {
		t.Fatalf("the saved explanation changed: %v", is["explanation"])
	}
	from, to := time.Now().Add(-time.Hour).UnixMilli(), time.Now().Add(time.Hour).UnixMilli()
	_, act := call(f.h, "GET", "/api/v1/requests?origin=aiport&outcome=upstream_error&from_ms="+strconv.FormatInt(from, 10)+"&to_ms="+strconv.FormatInt(to, 10), "", issuesManagement, false)
	if rows, _ := act["requests"].([]any); len(rows) != 1 || m(rows[0])["http_status"] != float64(401) {
		t.Fatalf("Activity: %v", act)
	}
	if w, out := call(f.h, "POST", "/api/v1/issues/nope/explain", "", issuesManagement, true); w.Code != 404 || m(out["error"])["code"] != "issue_not_found" {
		t.Fatalf("unknown Issue: %d %v", w.Code, out)
	}
}

// With adapting off, Explain is unavailable.
func TestExplainIsUnavailableWithAdaptingOff(t *testing.T) {
	writer := newFakeWriter(t, "unused")
	f := newExplainFixture(t, writer, "adapting: {disabled: true, explain: {connection: writer, model: writer-model}}\n")
	_, out := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if ex := m(out["explain"]); ex["available"] != false || !strings.Contains(ex["hint"].(string), "adapting") {
		t.Fatalf("explain with adapting off: %v", out["explain"])
	}
	if w, out := call(f.h, "POST", "/api/v1/issues/any/explain", "", issuesManagement, true); w.Code != 409 || m(out["error"])["code"] != "adapting_off" {
		t.Fatalf("explain with adapting off: %d %v", w.Code, out)
	}
	if bodies, _ := writer.requests(); len(bodies) != 0 {
		t.Fatal("the writer was asked with adapting off")
	}
}

// aiport is the built-in Application of aiport's own calls: no configured
// Application may take its id.
func TestNoApplicationMayBeCalledAiport(t *testing.T) {
	writer := newFakeWriter(t, "unused")
	f := newExplainFixture(t, writer, "")
	data, _ := os.ReadFile(f.path)
	bad := strings.Replace(string(data), "id: other,", "id: aiport,", 1)
	if err := os.WriteFile(f.path, []byte(bad), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(f.path); err == nil || !strings.Contains(err.Error(), "aiport") {
		t.Fatalf("an Application called aiport was accepted: %v", err)
	}
}
