package server

import (
	"encoding/json"
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

// The Catalog over HTTP: GET /v1/models answers the calling Application's
// Catalog in the shape its SDK expects, and every answer to an Application
// carries its Catalog version in X-Aiport-Catalog.

var (
	catManagement = strings.Repeat("c", 40)
	catDocs       = strings.Repeat("d", 40)
	catReview     = strings.Repeat("r", 40)
	catKey        = "sk-catalog-upstream-key-1234567890"
)

// catalogUpstream is a provider that lists models and answers model calls
// with the status and body the test sets.
type catalogUpstream struct {
	mu         sync.Mutex
	listStatus int
	list       string
	callStatus int
	callBody   string
	srv        *httptest.Server
}

func newCatalogUpstream(t *testing.T, listStatus int, list string) *catalogUpstream {
	u := &catalogUpstream{listStatus: listStatus, list: list, callStatus: 200, callBody: `{"id":"chatcmpl-1","object":"chat.completion","choices":[]}`}
	u.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.mu.Lock()
		defer u.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/models") {
			w.WriteHeader(u.listStatus)
			_, _ = w.Write([]byte(u.list))
			return
		}
		w.WriteHeader(u.callStatus)
		_, _ = w.Write([]byte(u.callBody))
	}))
	t.Cleanup(u.srv.Close)
	return u
}

func (u *catalogUpstream) answer(status int, body string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.callStatus, u.callBody = status, body
}

type catalogFixture struct {
	h                      *Handler
	hosted, claude, broken *catalogUpstream
}

// catalogServer: docs-bot follows the protocol defaults (hosted for OpenAI,
// claude for Anthropic) and has its own Alias "summarize"; review-code is
// bound to a Connection whose model list cannot be read and has its own
// Alias "review". The shared Alias "fast" is on hosted.
func catalogServer(t *testing.T) *catalogFixture {
	t.Helper()
	f := &catalogFixture{
		hosted: newCatalogUpstream(t, 200, `{"object":"list","data":[{"id":"gpt-4o"},{"id":"gpt-4o-mini"}]}`),
		claude: newCatalogUpstream(t, 200, `{"data":[{"id":"claude-haiku"},{"id":"claude-sonnet"}]}`),
		broken: newCatalogUpstream(t, 500, `{"error":"down"}`),
	}
	yaml := "version: 1\nauth: {token_env: CAT_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + f.hosted.srv.URL + "/v1', api_key_env: CAT_HOSTED_KEY}\n" +
		"  - {id: claude, provider: acme2, protocol: anthropic, base_url: '" + f.claude.srv.URL + "/v1'}\n" +
		"  - {id: broken, provider: acme3, protocol: openai, base_url: '" + f.broken.srv.URL + "/v1'}\n" +
		"defaults: {openai_target: hosted, anthropic_target: claude}\n" +
		"aliases:\n  - {target: hosted, alias: fast, model: gpt-4o-mini}\n" +
		"origins:\n" +
		"  - id: docs-bot\n    name: Docs Bot\n    token_env: CAT_DOCS_TOKEN\n    aliases:\n      - {alias: summarize, target: hosted, model: gpt-4o}\n" +
		"  - id: review-code\n    name: Review Code\n    token_env: CAT_REVIEW_TOKEN\n    target: broken\n    aliases:\n      - {alias: review, target: hosted, model: gpt-4o}\n"
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CAT_MGMT_TOKEN", catManagement)
	t.Setenv("CAT_DOCS_TOKEN", catDocs)
	t.Setenv("CAT_REVIEW_TOKEN", catReview)
	t.Setenv("CAT_HOSTED_KEY", catKey)
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

// listModels calls GET /v1/models; header pairs are added to the request.
func listModels(h *Handler, token string, header ...string) (*httptest.ResponseRecorder, map[string]any) {
	r := httptest.NewRequest("GET", "http://127.0.0.1:8787/v1/models", nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	for i := 0; i+1 < len(header); i += 2 {
		r.Header.Set(header[i], header[i+1])
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}

// entries returns id and the given field of each listed model.
func entries(t *testing.T, out map[string]any, field string) [][2]string {
	t.Helper()
	data, ok := out["data"].([]any)
	if !ok {
		t.Fatalf("no data list in %v", out)
	}
	got := [][2]string{}
	for _, d := range data {
		m := d.(map[string]any)
		v, _ := m[field].(string)
		got = append(got, [2]string{m["id"].(string), v})
	}
	return got
}

// A Catalog belongs to one Application: the management token has none.
func TestModelsNeedsAnApplicationToken(t *testing.T) {
	f := catalogServer(t)
	w, out := listModels(f.h, catManagement)
	if w.Code != 403 || errorCode(out) != "application_token_required" {
		t.Fatalf("GET /v1/models with the management token: %d %s", w.Code, w.Body)
	}
	if w, _ := listModels(f.h, ""); w.Code != 401 {
		t.Fatalf("GET /v1/models without a token: %d", w.Code)
	}
}

func TestModelsListsTheApplicationsTaskNamesFirstInTheOpenAIShape(t *testing.T) {
	f := catalogServer(t)
	w, out := listModels(f.h, catDocs)
	if w.Code != 200 || out["object"] != "list" {
		t.Fatalf("GET /v1/models: %d %s", w.Code, w.Body)
	}
	want := [][2]string{{"summarize", "aiport"}, {"fast", "aiport"}, {"gpt-4o", "hosted"}, {"gpt-4o-mini", "hosted"}}
	got := entries(t, out, "owned_by")
	if len(got) != len(want) {
		t.Fatalf("listed %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("listed %v, want %v", got, want)
		}
	}
	for _, d := range out["data"].([]any) {
		if d.(map[string]any)["object"] != "model" {
			t.Fatalf("entry is not an OpenAI model object: %v", d)
		}
	}
	for _, secret := range []string{f.hosted.srv.URL, f.claude.srv.URL, "127.0.0.1", catKey, "review"} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatalf("the Catalog discloses %q: %s", secret, w.Body)
		}
	}
}

func TestModelsAnswersTheAnthropicShapeToAnAnthropicSDK(t *testing.T) {
	f := catalogServer(t)
	// Anthropic SDKs authenticate with X-Api-Key.
	w, out := listModels(f.h, "", "X-Api-Key", catDocs, "anthropic-version", "2023-06-01")
	if w.Code != 200 || out["has_more"] != false {
		t.Fatalf("GET /v1/models (Anthropic): %d %s", w.Code, w.Body)
	}
	got := entries(t, out, "display_name")
	want := [][2]string{{"claude-haiku", "claude-haiku"}, {"claude-sonnet", "claude-sonnet"}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("listed %v, want %v", got, want)
	}
	if out["first_id"] != "claude-haiku" || out["last_id"] != "claude-sonnet" {
		t.Fatalf("first/last id: %v %v", out["first_id"], out["last_id"])
	}
	for _, d := range out["data"].([]any) {
		if d.(map[string]any)["type"] != "model" {
			t.Fatalf("entry is not an Anthropic model object: %v", d)
		}
	}
}

func TestATaskNameIsMarkedInTheAnthropicShape(t *testing.T) {
	f := catalogServer(t)
	if _, _, err := f.h.live.Change("add an Alias on claude", config.AddAlias(config.Alias{Target: "claude", Alias: "careful", Model: "claude-sonnet"})); err != nil {
		t.Fatal(err)
	}
	_, out := listModels(f.h, catDocs, "anthropic-version", "2023-06-01")
	got := entries(t, out, "display_name")
	if len(got) == 0 || got[0][0] != "careful" || !strings.Contains(got[0][1], "task name") {
		t.Fatalf("the task name is not first and marked: %v", got)
	}
}

func TestAnUnreadableModelListDoesNotFailTheAnswer(t *testing.T) {
	f := catalogServer(t)
	w, out := listModels(f.h, catReview)
	if w.Code != 200 {
		t.Fatalf("GET /v1/models: %d %s", w.Code, w.Body)
	}
	// review-code is bound to broken: its own Alias only; docs-bot's
	// "summarize" and the shared "fast" (on hosted) are not its.
	got := entries(t, out, "owned_by")
	if len(got) != 1 || got[0] != [2]string{"review", "aiport"} {
		t.Fatalf("listed %v, want only review", got)
	}
}

func catalogCall(h *Handler, method, path, body, token string) *httptest.ResponseRecorder {
	w, _ := call(h, method, path, body, token, false)
	return w
}

func TestEveryAnswerToAnApplicationCarriesItsCatalogVersion(t *testing.T) {
	f := catalogServer(t)
	_, _ = listModels(f.h, catDocs)
	w, _ := listModels(f.h, catDocs)
	docs := w.Header().Get(CatalogHeader)
	if docs == "" {
		t.Fatal("GET /v1/models has no Catalog version")
	}
	w, _ = listModels(f.h, catReview)
	if v := w.Header().Get(CatalogHeader); v == "" || v == docs {
		t.Fatalf("review-code's Catalog version %q (docs-bot's %q)", v, docs)
	}
	chat := `{"model":"summarize","messages":[{"role":"user","content":"hi"}]}`

	// A forwarded call that succeeds: the provider's body is unchanged.
	w = catalogCall(f.h, "POST", "/v1/chat/completions", chat, catDocs)
	if w.Code != 200 || w.Header().Get(CatalogHeader) != docs || w.Body.String() != `{"id":"chatcmpl-1","object":"chat.completion","choices":[]}` {
		t.Fatalf("forwarded success: %d %q %s", w.Code, w.Header().Get(CatalogHeader), w.Body)
	}
	// A provider error passes through unchanged, with the header.
	f.hosted.answer(404, `{"error":{"type":"model_not_found"}}`)
	w = catalogCall(f.h, "POST", "/v1/chat/completions", chat, catDocs)
	if w.Code != 404 || w.Header().Get(CatalogHeader) != docs || w.Body.String() != `{"error":{"type":"model_not_found"}}` {
		t.Fatalf("forwarded provider error: %d %q %s", w.Code, w.Header().Get(CatalogHeader), w.Body)
	}
	// aiport's own errors to an Application.
	for _, c := range []struct{ method, path, body string }{
		{"GET", "/v1/chat/completions", ""},
		{"GET", "/v1/nowhere", ""},
		{"POST", "/v1/messages", `{"model":"claude-haiku","max_tokens":1,"messages":[]}`}, // review-code's binding is OpenAI
		{"POST", "/v1/models", ""},
	} {
		w = catalogCall(f.h, c.method, c.path, c.body, catReview)
		if w.Code < 400 || w.Header().Get(CatalogHeader) == "" {
			t.Fatalf("%s %s: %d without a Catalog version", c.method, c.path, w.Code)
		}
	}
	// Not for the Owner, nor for a caller without a token.
	if w = catalogCall(f.h, "GET", "/api/v1/health", "", catManagement); w.Header().Get(CatalogHeader) != "" {
		t.Fatal("the management token got a Catalog version")
	}
	if w = catalogCall(f.h, "GET", "/v1/models", "", ""); w.Code != 401 || w.Header().Get(CatalogHeader) != "" {
		t.Fatalf("an unauthenticated call: %d %q", w.Code, w.Header().Get(CatalogHeader))
	}
}

func TestTheCatalogVersionFollowsTheProtocolDefaultsOnlyWhenTheApplicationDoes(t *testing.T) {
	f := catalogServer(t)
	version := func(token string) string {
		w, _ := listModels(f.h, token)
		return w.Header().Get(CatalogHeader)
	}
	docs, review := version(catDocs), version(catReview)
	w, _ := call(f.h, "PUT", "/api/v1/defaults", `{"openai_target":"broken","anthropic_target":"claude"}`, catManagement, true)
	if w.Code >= 300 {
		t.Fatalf("PUT /api/v1/defaults: %d %s", w.Code, w.Body)
	}
	if version(catDocs) == docs {
		t.Fatal("docs-bot follows the OpenAI default, but its Catalog version did not change")
	}
	if version(catReview) != review {
		t.Fatal("review-code is bound, but its Catalog version changed with the defaults")
	}
	// Reverting the Change brings the version back.
	changes, out := call(f.h, "GET", "/api/v1/changes", "", catManagement, false)
	list, _ := out["changes"].([]any)
	if changes.Code != 200 || len(list) == 0 {
		t.Fatalf("GET /api/v1/changes: %d %s", changes.Code, changes.Body)
	}
	id := list[0].(map[string]any)["id"].(string)
	if w, _ := call(f.h, "POST", "/api/v1/changes/"+id+"/revert", "", catManagement, true); w.Code >= 300 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	if version(catDocs) != docs {
		t.Fatal("the revert did not bring docs-bot's Catalog version back")
	}
}
