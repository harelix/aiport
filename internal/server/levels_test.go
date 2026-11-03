package server

import (
	"encoding/json"
	"io"
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

// Levels over HTTP: a client sends fast, balanced or best as its model and
// reaches the Connection and model the Owner mapped for its protocol, with
// the body forwarded unchanged apart from the model name.

var (
	lvlManagement = strings.Repeat("m", 40)
	lvlDocs       = strings.Repeat("o", 40)
	lvlReview     = strings.Repeat("v", 40)
)

// echoUpstream is a provider that keeps the last request body it received
// and answers 200.
type echoUpstream struct {
	mu    sync.Mutex
	calls int
	body  string
	path  string
	srv   *httptest.Server
}

func newEchoUpstream(t *testing.T) *echoUpstream {
	u := &echoUpstream{}
	u.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.mu.Lock()
		defer u.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/models") {
			_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
			return
		}
		b, _ := io.ReadAll(r.Body)
		u.calls++
		u.body, u.path = string(b), r.URL.Path
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(u.srv.Close)
	return u
}

func (u *echoUpstream) received() (int, string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.calls, u.body
}

type levelsFixture struct {
	h                     *Handler
	hosted, local, claude *echoUpstream
}

// levelsServer: hosted and local speak OpenAI, claude speaks Anthropic.
// Shared Level mappings: best → hosted gpt-4o, fast → local llama3:8b,
// balanced (Anthropic) → claude claude-sonnet. docs-bot follows the
// defaults; review-code has its own best on local (OpenAI only).
func levelsServer(t *testing.T) *levelsFixture {
	t.Helper()
	f := &levelsFixture{hosted: newEchoUpstream(t), local: newEchoUpstream(t), claude: newEchoUpstream(t)}
	yaml := "version: 1\nauth: {token_env: LVL_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + f.hosted.srv.URL + "/v1'}\n" +
		"  - {id: local, provider: ollama-ish, protocol: openai, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: claude, provider: acme2, protocol: anthropic, base_url: '" + f.claude.srv.URL + "/v1'}\n" +
		"defaults: {openai_target: hosted, anthropic_target: claude}\n" +
		"aliases:\n" +
		"  - {target: hosted, alias: best, model: gpt-4o}\n" +
		"  - {target: local, alias: fast, model: 'llama3:8b'}\n" +
		"  - {target: claude, alias: balanced, model: claude-sonnet}\n" +
		"origins:\n" +
		"  - id: docs-bot\n    name: Docs Bot\n    token_env: LVL_DOCS_TOKEN\n" +
		"  - id: review-code\n    name: Review Code\n    token_env: LVL_REVIEW_TOKEN\n    aliases:\n      - {alias: best, target: local, model: 'llama3:70b'}\n"
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LVL_MGMT_TOKEN", lvlManagement)
	t.Setenv("LVL_DOCS_TOKEN", lvlDocs)
	t.Setenv("LVL_REVIEW_TOKEN", lvlReview)
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

// levelCall sends body to an inference path with an Application token.
func levelCall(h *Handler, path, token, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+token)
	if path == "/v1/messages" {
		r.Header.Set("anthropic-version", "2023-06-01")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// A body with odd spacing, key order, nested objects and escapes: only the
// model value may differ upstream.
const levelChatBody = "{ \"messages\":[{\"role\":\"user\",\"content\":\"caf\\u00e9 \\\"best\\\"\"}],\n\t\"model\" :  \"%s\" ,\"reasoning_effort\":\"high\",\"metadata\":{\"model\":\"best\"}}"

func chatBody(model string) string { return strings.Replace(levelChatBody, "%s", model, 1) }

func TestEachLevelResolvesPerProtocolWithTheBodyUnchangedApartFromTheModel(t *testing.T) {
	f := levelsServer(t)
	cases := []struct {
		path, level string
		up          *echoUpstream
		want        string
	}{
		{"/v1/chat/completions", "best", f.hosted, "{ \"messages\":[{\"role\":\"user\",\"content\":\"caf\\u00e9 \\\"best\\\"\"}],\n\t\"model\" :  \"gpt-4o\" ,\"reasoning_effort\":\"high\",\"metadata\":{\"model\":\"best\"}}"},
		{"/v1/chat/completions", "fast", f.local, "{ \"messages\":[{\"role\":\"user\",\"content\":\"caf\\u00e9 \\\"best\\\"\"}],\n\t\"model\" :  \"llama3:8b\" ,\"reasoning_effort\":\"high\",\"metadata\":{\"model\":\"best\"}}"},
		{"/v1/messages", "balanced", f.claude, "{ \"messages\":[{\"role\":\"user\",\"content\":\"caf\\u00e9 \\\"best\\\"\"}],\n\t\"model\" :  \"claude-sonnet\" ,\"reasoning_effort\":\"high\",\"metadata\":{\"model\":\"best\"}}"},
	}
	for _, c := range cases {
		w := levelCall(f.h, c.path, lvlDocs, chatBody(c.level))
		if w.Code != 200 {
			t.Fatalf("%s on %s: %d %s", c.level, c.path, w.Code, w.Body)
		}
		if _, got := c.up.received(); got != c.want {
			t.Fatalf("%s on %s reached upstream as\n%s\nwant\n%s", c.level, c.path, got, c.want)
		}
	}
}

func TestAnApplicationsOwnLevelComesBeforeTheSharedOneOnItsProtocolOnly(t *testing.T) {
	f := levelsServer(t)
	if w := levelCall(f.h, "/v1/chat/completions", lvlReview, `{"model":"best"}`); w.Code != 200 {
		t.Fatalf("review-code best: %d %s", w.Code, w.Body)
	}
	if n, got := f.local.received(); n != 1 || got != `{"model":"llama3:70b"}` {
		t.Fatalf("review-code's own best reached local as %q (%d calls)", got, n)
	}
	if n, _ := f.hosted.received(); n != 0 {
		t.Fatal("review-code's best went to the shared mapping")
	}
	// Its own best is on OpenAI only: on Anthropic, best is not mapped.
	w := levelCall(f.h, "/v1/messages", lvlReview, `{"model":"best"}`)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "level_unmapped") {
		t.Fatalf("review-code best on Anthropic: %d %s", w.Code, w.Body)
	}
	// docs-bot has no own mapping: the shared one serves it.
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"best"}`); w.Code != 200 {
		t.Fatalf("docs-bot best: %d %s", w.Code, w.Body)
	}
	if _, got := f.hosted.received(); got != `{"model":"gpt-4o"}` {
		t.Fatalf("docs-bot best reached hosted as %q", got)
	}
}

func TestALevelNotMappedOnTheProtocolIsRefusedNamingLevelAndProtocol(t *testing.T) {
	f := levelsServer(t)
	w := levelCall(f.h, "/v1/messages", lvlDocs, `{"model":"best","max_tokens":5}`)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if w.Code != 400 || errorCode(out) != "level_unmapped" {
		t.Fatalf("unmapped Level: %d %s", w.Code, w.Body)
	}
	if m := errorMessage(out); !strings.Contains(m, `"best"`) || !strings.Contains(m, "Anthropic") {
		t.Fatalf("the refusal does not name the Level and protocol: %q", m)
	}
	if n, _ := f.claude.received(); n != 0 {
		t.Fatal("an unmapped Level was forwarded to the protocol default")
	}
	// A name that is not a Level still goes to the protocol default unchanged.
	if w := levelCall(f.h, "/v1/messages", lvlDocs, `{"model":"claude-haiku"}`); w.Code != 200 {
		t.Fatalf("plain model: %d %s", w.Code, w.Body)
	}
	if _, got := f.claude.received(); got != `{"model":"claude-haiku"}` {
		t.Fatalf("plain model reached claude as %q", got)
	}
}

func TestALevelIsServedWithoutAProtocolDefault(t *testing.T) {
	f := levelsServer(t)
	if _, _, err := f.h.live.Change("no Anthropic default", config.SetDefaultTarget("anthropic", "")); err != nil {
		t.Fatal(err)
	}
	if w := levelCall(f.h, "/v1/messages", lvlDocs, `{"model":"balanced"}`); w.Code != 200 {
		t.Fatalf("balanced without a default: %d %s", w.Code, w.Body)
	}
	if _, got := f.claude.received(); got != `{"model":"claude-sonnet"}` {
		t.Fatalf("balanced reached claude as %q", got)
	}
	// Any other name still finds no Connection.
	if w := levelCall(f.h, "/v1/messages", lvlDocs, `{"model":"claude-haiku"}`); w.Code != 503 || !strings.Contains(w.Body.String(), "target_unconfigured") {
		t.Fatalf("plain model without a default: %d %s", w.Code, w.Body)
	}
}

// On a protocol with no default and no Level mapped, a Level is still
// refused as an unmapped Level, not as a missing Connection.
func TestALevelOnAProtocolWithNothingIsUnmappedNotUnconfigured(t *testing.T) {
	f := levelsServer(t)
	if _, _, err := f.h.live.Change("no OpenAI default", config.SetDefaultTarget("openai", "")); err != nil {
		t.Fatal(err)
	}
	for _, l := range []string{"best", "fast"} {
		e, err := config.SetLevel(f.h.live.Config(), config.LevelEdit{Protocol: "openai", Level: l})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := f.h.live.Change("unmap "+l, e); err != nil {
			t.Fatal(err)
		}
	}
	w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"best"}`)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "level_unmapped") {
		t.Fatalf("best with nothing on OpenAI: %d %s", w.Code, w.Body)
	}
	// Any other name still finds no Connection.
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"gpt-4o"}`); w.Code != 503 || !strings.Contains(w.Body.String(), "target_unconfigured") {
		t.Fatalf("plain model without a default: %d %s", w.Code, w.Body)
	}
}

func TestMappedLevelsAreListedInModelsOnTheirProtocol(t *testing.T) {
	f := levelsServer(t)
	_, out := listModels(f.h, lvlDocs)
	got := entries(t, out, "owned_by")
	want := [][2]string{{"fast", "aiport"}, {"best", "aiport"}} // in Level order
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("OpenAI list %v, want %v", got, want)
	}
	_, out = listModels(f.h, lvlDocs, "anthropic-version", "2023-06-01")
	got = entries(t, out, "display_name")
	if len(got) != 1 || got[0][0] != "balanced" || !strings.Contains(got[0][1], "Level") {
		t.Fatalf("Anthropic list %v, want only balanced marked as a Level", got)
	}
}

func TestMappingALevelIsARevertableChangeAndACatalogChange(t *testing.T) {
	f := levelsServer(t)
	edit, err := config.SetLevel(f.h.snapshot(), config.LevelEdit{Protocol: "anthropic", Level: "best", Connection: "claude", Model: "claude-opus"})
	if err != nil {
		t.Fatal(err)
	}
	_, ch, err := f.h.live.Change("map best on Anthropic", edit)
	if err != nil {
		t.Fatal(err)
	}
	if w := levelCall(f.h, "/v1/messages", lvlDocs, `{"model":"best"}`); w.Code != 200 {
		t.Fatalf("best after mapping: %d %s", w.Code, w.Body)
	}
	if _, got := f.claude.received(); got != `{"model":"claude-opus"}` {
		t.Fatalf("best reached claude as %q", got)
	}
	changes := catalogChanges(t, catalogOf(t, f.h, lvlDocs))
	if len(changes) != 1 || changes[0]["change_id"] != ch.ID {
		t.Fatalf("catalog changes %v", changes)
	}
	added, _ := changes[0]["added"].([]any)
	if len(added) != 1 || added[0].(map[string]any)["name"] != "best" || added[0].(map[string]any)["protocol"] != "anthropic" || added[0].(map[string]any)["level"] != true {
		t.Fatalf("the Catalog change does not add the Anthropic best: %v", changes[0])
	}
	// The OpenAI best is untouched.
	if w := levelCall(f.h, "/v1/chat/completions", lvlDocs, `{"model":"best"}`); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if _, got := f.hosted.received(); got != `{"model":"gpt-4o"}` {
		t.Fatalf("OpenAI best reached hosted as %q", got)
	}
	if _, _, err := f.h.live.Revert(ch.ID); err != nil {
		t.Fatal(err)
	}
	if w := levelCall(f.h, "/v1/messages", lvlDocs, `{"model":"best"}`); w.Code != 400 {
		t.Fatalf("best after the revert: %d %s", w.Code, w.Body)
	}
}

func TestALevelMappingMustSpeakItsProtocol(t *testing.T) {
	f := levelsServer(t)
	if _, err := config.SetLevel(f.h.snapshot(), config.LevelEdit{Protocol: "anthropic", Level: "best", Connection: "hosted", Model: "gpt-4o"}); err == nil {
		t.Fatal("mapped the Anthropic best to an OpenAI Connection")
	}
	if _, err := config.SetLevel(f.h.snapshot(), config.LevelEdit{Protocol: "openai", Level: "best", Connection: "nowhere", Model: "gpt-4o"}); err == nil {
		t.Fatal("mapped best to a missing Connection")
	}
	if _, err := config.SetLevel(f.h.snapshot(), config.LevelEdit{Protocol: "openai", Level: "strongest", Connection: "hosted", Model: "gpt-4o"}); err == nil {
		t.Fatal("mapped a name that is not a Level")
	}
}
