package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Effort over HTTP: a client says how hard a call is (X-Aiport-Effort, or
// its provider's own effort field) and aiport picks the per-effort model of
// the Level or task name it sent. The body is forwarded unchanged apart from
// the model name, the effort field included.

type effortFixture struct {
	h              *Handler
	hosted, claude *echoUpstream
}

// effortServer: hosted speaks OpenAI, claude speaks Anthropic. The shared
// balanced Level has per-effort models on both protocols; the shared task
// name summarize has only a low model; docs-bot's own task name review has
// only a high model; quick has no per-effort models.
func effortServer(t *testing.T) *effortFixture {
	t.Helper()
	f := &effortFixture{hosted: newEchoUpstream(t), claude: newEchoUpstream(t)}
	yaml := "version: 1\nauth: {token_env: EFF_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + f.hosted.srv.URL + "/v1'}\n" +
		"  - {id: claude, provider: acme2, protocol: anthropic, base_url: '" + f.claude.srv.URL + "/v1'}\n" +
		"defaults: {openai_target: hosted, anthropic_target: claude}\n" +
		"aliases:\n" +
		"  - {target: hosted, alias: balanced, model: gpt-mid, effort: {low: gpt-mini, high: gpt-big}}\n" +
		"  - {target: claude, alias: balanced, model: claude-mid, effort: {low: claude-small, high: claude-large}}\n" +
		"  - {target: hosted, alias: summarize, model: gpt-mid, effort: {low: gpt-mini}}\n" +
		"  - {target: hosted, alias: quick, model: gpt-mini}\n" +
		"origins:\n" +
		"  - id: docs-bot\n    name: Docs Bot\n    token_env: EFF_DOCS_TOKEN\n    aliases:\n      - {alias: review, target: hosted, model: gpt-mid, effort: {high: gpt-big}}\n"
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EFF_MGMT_TOKEN", lvlManagement)
	t.Setenv("EFF_DOCS_TOKEN", lvlDocs)
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	f.h = NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	now := time.UnixMilli(1_800_000_000_000)
	f.h.now = func() time.Time { return now }
	t.Cleanup(f.h.Close)
	return f
}

// effortCall sends body to path as docs-bot, with X-Aiport-Effort when
// header is not "".
func effortCall(h *Handler, path, header, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+lvlDocs)
	if path == "/v1/messages" {
		r.Header.Set("anthropic-version", "2023-06-01")
	}
	if header != "" {
		r.Header.Set("X-Aiport-Effort", header)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// sentAs is body with its model value replaced, as the upstream must see it.
func sentAs(body, from, to string) string {
	return strings.Replace(body, `"model" : "`+from+`"`, `"model" : "`+to+`"`, 1)
}

// Bodies with odd spacing, nested objects and the provider's own effort
// field (and a numeric thinking budget): only the model value may differ.
const (
	chatEffortBody      = "{ \"messages\":[{\"role\":\"user\",\"content\":\"caf\\u00e9\"}],\n\t\"model\" : \"%m\" ,\"reasoning_effort\":%e,\"metadata\":{\"model\":\"best\"}}"
	responsesEffortBody = "{\"input\":\"hi\", \"model\" : \"%m\",\"reasoning\": {\"effort\":%e, \"summary\":\"auto\"}}"
	messagesEffortBody  = "{\"max_tokens\":64, \"model\" : \"%m\", \"thinking\":{\"type\":\"enabled\",\"budget_tokens\":1024},\"output_config\":{\"effort\":%e},\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}"
)

func bodyOf(template, model, effort string) string {
	return strings.Replace(strings.Replace(template, "%m", model, 1), "%e", effort, 1)
}

func expectSent(t *testing.T, w *httptest.ResponseRecorder, up *echoUpstream, want string) {
	t.Helper()
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if _, got := up.received(); got != want {
		t.Fatalf("upstream got\n%s\nwant\n%s", got, want)
	}
}

func TestTheEffortHeaderPicksTheLevelsPerEffortModel(t *testing.T) {
	f := effortServer(t)
	body := bodyOf(chatEffortBody, "balanced", `"medium"`)
	for header, model := range map[string]string{"low": "gpt-mini", "medium": "gpt-mid", "high": "gpt-big"} {
		w := effortCall(f.h, "/v1/chat/completions", header, body)
		expectSent(t, w, f.hosted, sentAs(body, "balanced", model))
	}
}

func TestTheCatalogListsPerEffortModels(t *testing.T) {
	f := effortServer(t)
	b, _ := json.Marshal(catalogOf(t, f.h, lvlDocs))
	got := string(b)
	for _, want := range []string{
		`"effort":{"high":"gpt-big","low":"gpt-mini"}`,          // balanced on OpenAI
		`"effort":{"high":"claude-large","low":"claude-small"}`, // balanced on Anthropic
		`"effort":{"low":"gpt-mini"}`,                           // summarize
		`"effort":{"high":"gpt-big"}`,                           // review
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("the Catalog lacks %s:\n%s", want, got)
		}
	}
	if strings.Count(got, `"effort"`) != 4 {
		t.Fatalf("a name without per-effort models lists them:\n%s", got)
	}
}

func TestEachProvidersOwnEffortFieldPicksThePerEffortModel(t *testing.T) {
	f := effortServer(t)
	cases := []struct {
		path, template string
		up             *echoUpstream
		value, want    string
	}{
		{"/v1/chat/completions", chatEffortBody, f.hosted, `"low"`, "gpt-mini"},
		{"/v1/chat/completions", chatEffortBody, f.hosted, `"high"`, "gpt-big"},
		{"/v1/responses", responsesEffortBody, f.hosted, `"low"`, "gpt-mini"},
		{"/v1/responses", responsesEffortBody, f.hosted, `"high"`, "gpt-big"},
		{"/v1/messages", messagesEffortBody, f.claude, `"low"`, "claude-small"},
		{"/v1/messages", messagesEffortBody, f.claude, `"high"`, "claude-large"},
	}
	for _, c := range cases {
		body := bodyOf(c.template, "balanced", c.value)
		w := effortCall(f.h, c.path, "", body)
		expectSent(t, w, c.up, sentAs(body, "balanced", c.want))
	}
}

func TestProviderEffortValuesMapToLowMediumOrHigh(t *testing.T) {
	f := effortServer(t)
	for value, want := range map[string]string{
		`"none"`: "gpt-mini", `"minimal"`: "gpt-mini", `"low"`: "gpt-mini",
		`"medium"`: "gpt-mid",
		`"high"`:   "gpt-big", `"xhigh"`: "gpt-big", `"max"`: "gpt-big",
	} {
		body := bodyOf(responsesEffortBody, "balanced", value)
		w := effortCall(f.h, "/v1/responses", "", body)
		expectSent(t, w, f.hosted, sentAs(body, "balanced", want))
	}
}

func TestTheEffortHeaderWinsOverTheProviderField(t *testing.T) {
	f := effortServer(t)
	body := bodyOf(messagesEffortBody, "balanced", `"high"`)
	w := effortCall(f.h, "/v1/messages", "low", body)
	expectSent(t, w, f.claude, sentAs(body, "balanced", "claude-small"))
}

func TestNoEffortMeansMedium(t *testing.T) {
	f := effortServer(t)
	body := `{"model" : "balanced","messages":[]}`
	w := effortCall(f.h, "/v1/chat/completions", "", body)
	expectSent(t, w, f.hosted, sentAs(body, "balanced", "gpt-mid"))
	// A header value that is no effort states none: the field, else medium.
	w = effortCall(f.h, "/v1/chat/completions", "extreme", body)
	expectSent(t, w, f.hosted, sentAs(body, "balanced", "gpt-mid"))
	body = bodyOf(chatEffortBody, "balanced", `"low"`)
	w = effortCall(f.h, "/v1/chat/completions", "extreme", body)
	expectSent(t, w, f.hosted, sentAs(body, "balanced", "gpt-mini"))
}

func TestNumericThinkingBudgetsAreNeverTranslated(t *testing.T) {
	f := effortServer(t)
	// Only a budget, no effort: medium. A numeric effort counts as none.
	body := "{\"max_tokens\":64, \"model\" : \"balanced\", \"thinking\":{\"type\":\"enabled\",\"budget_tokens\":32000},\"messages\":[]}"
	w := effortCall(f.h, "/v1/messages", "", body)
	expectSent(t, w, f.claude, sentAs(body, "balanced", "claude-mid"))
	body = bodyOf(messagesEffortBody, "balanced", "3")
	w = effortCall(f.h, "/v1/messages", "", body)
	expectSent(t, w, f.claude, sentAs(body, "balanced", "claude-mid"))
}

func TestAnOddEffortFieldNeverCostsTheModelName(t *testing.T) {
	f := effortServer(t)
	for _, c := range []struct{ path, body string }{
		{"/v1/chat/completions", bodyOf(chatEffortBody, "balanced", `{"x":1}`)},
		{"/v1/chat/completions", bodyOf(chatEffortBody, "balanced", `3`)},
		{"/v1/responses", `{"model" : "balanced","reasoning":"high","input":"hi"}`},
		{"/v1/responses", `{"model" : "balanced","reasoning":null,"input":"hi"}`},
	} {
		w := effortCall(f.h, c.path, "", c.body)
		expectSent(t, w, f.hosted, sentAs(c.body, "balanced", "gpt-mid"))
	}
}

func TestOnlyTheEffortFieldOfTheRequestsOwnAPIIsRead(t *testing.T) {
	f := effortServer(t)
	// reasoning_effort means nothing on Responses, output_config nothing on
	// Chat Completions.
	body := `{"model" : "balanced","reasoning_effort":"low","input":"hi"}`
	w := effortCall(f.h, "/v1/responses", "", body)
	expectSent(t, w, f.hosted, sentAs(body, "balanced", "gpt-mid"))
	body = `{"model" : "balanced","output_config":{"effort":"low"},"messages":[]}`
	w = effortCall(f.h, "/v1/chat/completions", "", body)
	expectSent(t, w, f.hosted, sentAs(body, "balanced", "gpt-mid"))
}

func TestTaskNamesHavePerEffortModelsTheSameWay(t *testing.T) {
	f := effortServer(t)
	cases := []struct{ name, header, want string }{
		{"review", "high", "gpt-big"},    // docs-bot's own task name
		{"review", "low", "gpt-mid"},     // no low model: the Alias's model
		{"summarize", "low", "gpt-mini"}, // a shared task name
		{"summarize", "", "gpt-mid"},
		{"quick", "high", "gpt-mini"}, // no per-effort models at all
	}
	for _, c := range cases {
		body := bodyOf(chatEffortBody, c.name, `"medium"`)
		w := effortCall(f.h, "/v1/chat/completions", c.header, body)
		expectSent(t, w, f.hosted, sentAs(body, c.name, c.want))
	}
	// The provider field picks a task name's model too.
	body := bodyOf(chatEffortBody, "summarize", `"minimal"`)
	w := effortCall(f.h, "/v1/chat/completions", "", body)
	expectSent(t, w, f.hosted, sentAs(body, "summarize", "gpt-mini"))
}
