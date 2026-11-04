package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

// everywhereServer: docs-bot's calls go to hosted (the OpenAI default). The
// shared task name classify lives on cheap and is set to everywhere, with
// reasoning always high; tidy lives on cheap without everywhere.
func everywhereServer(t *testing.T) (*Handler, *echoUpstream, *echoUpstream) {
	t.Helper()
	hosted, cheap := newEchoUpstream(t), newEchoUpstream(t)
	yaml := "version: 1\nauth: {token_env: EVW_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + hosted.srv.URL + "/v1'}\n" +
		"  - {id: cheap, provider: acme3, protocol: openai, base_url: '" + cheap.srv.URL + "/v1'}\n" +
		"defaults: {openai_target: hosted}\n" +
		"aliases:\n" +
		"  - {target: cheap, alias: classify, model: small-model, everywhere: true, reasoning: high}\n" +
		"  - {target: cheap, alias: tidy, model: small-model}\n" +
		"origins:\n  - id: docs-bot\n    name: Docs Bot\n    token_env: EVW_DOCS_TOKEN\n"
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EVW_MGMT_TOKEN", lvlManagement)
	t.Setenv("EVW_DOCS_TOKEN", lvlDocs)
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	t.Cleanup(h.Close)
	return h, hosted, cheap
}

func TestATaskNameSetToEverywhereRoutesEveryApplicationToItsConnection(t *testing.T) {
	h, hosted, cheap := everywhereServer(t)
	// classify goes to cheap although docs-bot's calls go to hosted, with
	// reasoning high whatever effort the header states.
	for _, header := range []string{"", "low", "high"} {
		w := effortCall(h, "/v1/chat/completions", header, `{"model":"classify","messages":[]}`)
		expectSent(t, w, cheap, `{"model":"small-model","messages":[],"reasoning_effort":"high"}`)
	}
	// The body's own effort is never changed.
	w := effortCall(h, "/v1/chat/completions", "", `{"model":"classify","reasoning_effort":"low"}`)
	expectSent(t, w, cheap, `{"model":"small-model","reasoning_effort":"low"}`)
	// tidy is not set to everywhere: on hosted it is just a model id.
	w = effortCall(h, "/v1/chat/completions", "", `{"model":"tidy"}`)
	expectSent(t, w, hosted, `{"model":"tidy"}`)

	b, _ := json.Marshal(catalogOf(t, h, lvlDocs))
	got := string(b)
	if !strings.Contains(got, `{"connection":"cheap","everywhere":true,"model":"small-model","name":"classify","reasoning":"high","shared":true}`) {
		t.Fatalf("the Catalog lacks classify with its reasoning:\n%s", got)
	}
	if strings.Contains(got, `"name":"tidy"`) {
		t.Fatalf("the Catalog lists tidy, which docs-bot cannot reach:\n%s", got)
	}
}

func TestEverywhereAndReasoningAreValidated(t *testing.T) {
	base := "version: 1\ntargets:\n  - {id: a, provider: acme, protocol: openai, base_url: 'http://127.0.0.1:1/v1'}\n  - {id: b, provider: acme, protocol: openai, base_url: 'http://127.0.0.1:2/v1'}\naliases:\n"
	for name, aliases := range map[string]string{
		"reasoning that is no effort": "  - {target: a, alias: classify, model: m, reasoning: extreme}\n",
		"reasoning in capitals":       "  - {target: a, alias: classify, model: m, reasoning: High}\n",
		"a Level set to everywhere":   "  - {target: a, alias: fast, model: m, everywhere: true}\n",
		"everywhere twice":            "  - {target: a, alias: classify, model: m, everywhere: true}\n  - {target: b, alias: classify, model: m, everywhere: true}\n",
	} {
		path := filepath.Join(t.TempDir(), "aiport.yaml")
		if err := os.WriteFile(path, []byte(base+aliases), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := config.Load(path); err == nil {
			t.Errorf("%s: loaded", name)
		}
	}
}

func TestTheOwnerAddsChangesAndRemovesATaskName(t *testing.T) {
	h, hosted, cheap := everywhereServer(t)
	put := func(name, body string) (int, map[string]any) {
		w, out := call(h, "PUT", "/api/v1/task-names/"+name, body, lvlManagement, true)
		return w.Code, out
	}
	if code, out := put("tldr", `{"connection":"cheap","model":"small-model","everywhere":true,"reasoning":"high"}`); code != 200 || out["change_id"] == "" {
		t.Fatalf("add tldr: %d %v", code, out)
	}
	w := effortCall(h, "/v1/chat/completions", "", `{"model":"tldr"}`)
	expectSent(t, w, cheap, `{"model":"small-model","reasoning_effort":"high"}`)
	// Changing it without everywhere keeps it on cheap only.
	if code, out := put("tldr", `{"connection":"cheap","model":"small-model-2"}`); code != 200 {
		t.Fatalf("change tldr: %d %v", code, out)
	}
	w = effortCall(h, "/v1/chat/completions", "", `{"model":"tldr"}`)
	expectSent(t, w, hosted, `{"model":"tldr"}`)
	for name, body := range map[string]string{
		"fast":    `{"connection":"cheap","model":"m"}`,
		"tldr":    `{"connection":"nowhere","model":"m"}`,
		"summary": `{"connection":"cheap","model":"m","reasoning":"extreme"}`,
	} {
		if code, _ := put(name, body); code != 400 {
			t.Errorf("%s %s: %d, want 400", name, body, code)
		}
	}
	if code, _ := put("classify", `{"connection":"hosted","model":"m","everywhere":true}`); code != 400 {
		t.Errorf("a second everywhere classify: %d, want 400", code)
	}
	if w, out := call(h, "DELETE", "/api/v1/task-names/tldr", `{"connection":"cheap"}`, lvlManagement, true); w.Code != 200 {
		t.Fatalf("remove tldr: %d %v", w.Code, out)
	}
	for _, a := range h.snapshot().Aliases {
		if a.Alias == "tldr" {
			t.Fatalf("tldr is still there: %+v", a)
		}
	}
}

// Every forwarded call says which model answered and why, so the application
// can keep it with the answer and its vote.
func TestAForwardedCallNamesTheModelThatAnswered(t *testing.T) {
	h, _, _ := everywhereServer(t)
	for _, c := range []struct{ body, module, model, route string }{
		{`{"model":"classify"}`, "", "small-model", "connection=cheap; name=classify; effort=medium; via=task_name"},
		{`{"model":"gpt-raw"}`, "", "gpt-raw", "connection=hosted; name=gpt-raw; effort=medium; via=model"},
	} {
		w := effortCall(h, "/v1/chat/completions", "", c.body)
		if w.Code != 200 || w.Header().Get("X-Aiport-Model") != c.model || w.Header().Get("X-Aiport-Route") != c.route {
			t.Fatalf("%s: %d model %q route %q", c.body, w.Code, w.Header().Get("X-Aiport-Model"), w.Header().Get("X-Aiport-Route"))
		}
	}
	if w, out := call(h, "PUT", "/api/v1/module-routes/docs-bot/insights/classify", `{"connection":"hosted","model":"big-model"}`, lvlManagement, true); w.Code != 200 {
		t.Fatalf("module route: %d %v", w.Code, out)
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", strings.NewReader(`{"model":"classify"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+lvlDocs)
	r.Header.Set("X-Aiport-Module", "insights")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Header().Get("X-Aiport-Model") != "big-model" || !strings.HasSuffix(w.Header().Get("X-Aiport-Route"), "via=module_route") {
		t.Fatalf("module route headers: %v", w.Header())
	}
}
