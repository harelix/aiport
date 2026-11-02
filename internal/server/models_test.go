package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"aiport/internal/config"
)

// fakeProvider is an upstream whose model list the test controls.
type fakeProvider struct {
	mu     sync.Mutex
	status int
	body   string
	hits   int
	auth   string
	srv    *httptest.Server
}

func newFakeProvider(t *testing.T, status int, body string) *fakeProvider {
	p := &fakeProvider{status: status, body: body}
	p.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.hits++
		p.auth = r.Header.Get("Authorization") + r.Header.Get("X-Api-Key")
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/models") {
			t.Errorf("listing must be a non-generating GET /models, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(p.status)
		_, _ = w.Write([]byte(p.body))
	}))
	t.Cleanup(p.srv.Close)
	return p
}

func (p *fakeProvider) set(status int, body string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status, p.body = status, body
}

func (p *fakeProvider) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hits
}

type modelsFixture struct {
	h                                 *Handler
	now                               time.Time
	local, hosted, spare, claude, off *fakeProvider
}

// newModelsFixture is a gateway with three enabled OpenAI-protocol
// Connections (a local server, a hosted one with a key answering in the
// Ollama shape, and a spare one), one Anthropic-protocol Connection and one
// disabled Connection, each with its own fake provider.
func newModelsFixture(t *testing.T) *modelsFixture {
	f := &modelsFixture{now: time.UnixMilli(1_800_000_000_000)}
	f.local = newFakeProvider(t, 200, `{"object":"list","data":[{"id":"llama3:8b"},{"id":"qwen2.5:7b"}]}`)
	f.hosted = newFakeProvider(t, 200, `{"models":[{"name":"gpt-4o-mini","model":"gpt-4o-mini"}]}`)
	f.spare = newFakeProvider(t, 200, `{"data":[{"id":"mistral:7b"}]}`)
	f.claude = newFakeProvider(t, 200, `{"data":[{"id":"llama3:8b"}]}`)
	f.off = newFakeProvider(t, 200, `{"data":[{"id":"llama3:8b"}]}`)
	yaml := "version: 1\nauth: {token_env: T, disabled: true}\ntargets:\n" +
		"  - {id: local, provider: ollama, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: '" + f.hosted.srv.URL + "/v1', api_key_env: HOSTED_KEY}\n" +
		"  - {id: spare, provider: vllm, protocol: openai, base_url: '" + f.spare.srv.URL + "/v1'}\n" +
		"  - {id: claude, provider: acme2, protocol: anthropic, base_url: '" + f.claude.srv.URL + "/v1'}\n" +
		"  - {id: off, provider: vllm2, protocol: openai, base_url: '" + f.off.srv.URL + "/v1', disabled: true}\n" +
		"defaults: {openai_target: hosted, anthropic_target: claude}\n"
	c, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	f.h = New(c, config.Secrets{Upstream: map[string]string{"hosted": "sk-hosted-secret"}}).(*Handler)
	f.h.now = func() time.Time { return f.now }
	return f
}

// get sends a management GET from the dashboard and decodes the answer.
func (f *modelsFixture) get(t *testing.T, path string, out any) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("GET", "http://127.0.0.1:8787"+path, nil)
	r.Header.Set("X-Aiport-CSRF", "1")
	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)
	if out != nil {
		if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
	}
	return w
}

type modelList struct {
	Connection string   `json:"connection"`
	Listed     bool     `json:"listed"`
	State      string   `json:"state"`
	Models     []string `json:"models"`
	ListedAt   int64    `json:"listed_at_ms"`
}

func TestConnectionModelsAreListed(t *testing.T) {
	f := newModelsFixture(t)
	var got modelList
	if w := f.get(t, "/api/v1/targets/local/models", &got); w.Code != 200 {
		t.Fatalf("models of local: %d %s", w.Code, w.Body)
	}
	if !got.Listed || got.Connection != "local" || strings.Join(got.Models, ",") != "llama3:8b,qwen2.5:7b" {
		t.Fatalf("OpenAI-shaped listing: %+v", got)
	}
	if got.ListedAt != f.now.UnixMilli() {
		t.Fatalf("listed_at_ms = %d, want the gateway clock", got.ListedAt)
	}
	if w := f.get(t, "/api/v1/targets/hosted/models", &got); w.Code != 200 || !got.Listed || strings.Join(got.Models, ",") != "gpt-4o-mini" {
		t.Fatalf("Ollama-shaped listing: %d %+v", w.Code, got)
	}
	if f.hosted.auth != "Bearer sk-hosted-secret" {
		t.Fatalf("listing must use the Connection's key, sent %q", f.hosted.auth)
	}
}

func TestModelListsAreCachedBrieflyAndRefreshedOnDemand(t *testing.T) {
	f := newModelsFixture(t)
	var got modelList
	f.get(t, "/api/v1/targets/local/models", &got)
	f.local.set(200, `{"data":[{"id":"llama3:8b"},{"id":"phi3:mini"}]}`)
	f.get(t, "/api/v1/targets/local/models", &got)
	if f.local.count() != 1 || strings.Join(got.Models, ",") != "llama3:8b,qwen2.5:7b" {
		t.Fatalf("a recent list must be reused: %d listings, %v", f.local.count(), got.Models)
	}
	f.get(t, "/api/v1/targets/local/models?refresh=1", &got)
	if f.local.count() != 2 || strings.Join(got.Models, ",") != "llama3:8b,phi3:mini" {
		t.Fatalf("refresh must list again: %d listings, %v", f.local.count(), got.Models)
	}
	f.now = f.now.Add(10 * time.Minute)
	f.get(t, "/api/v1/targets/local/models", &got)
	if f.local.count() != 3 || got.ListedAt != f.now.UnixMilli() {
		t.Fatalf("an old list must be listed again: %d listings, %+v", f.local.count(), got)
	}
	// The connection check is the same listing and refreshes the list.
	f.local.set(200, `{"data":[{"id":"gemma2:9b"}]}`)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/v1/targets/local/check", nil)
	r.Header.Set("X-Aiport-CSRF", "1")
	f.h.ServeHTTP(httptest.NewRecorder(), r)
	f.get(t, "/api/v1/targets/local/models", &got)
	if f.local.count() != 4 || strings.Join(got.Models, ",") != "gemma2:9b" {
		t.Fatalf("Check connection must refresh the list: %d listings, %v", f.local.count(), got.Models)
	}
}

type serving struct {
	Model    string   `json:"model"`
	Protocol string   `json:"protocol"`
	Answer   string   `json:"answer"`
	Serving  []string `json:"serving"`
	Unknown  []string `json:"unknown"`
}

func (f *modelsFixture) serves(t *testing.T, protocol, model string) serving {
	t.Helper()
	var s serving
	if w := f.get(t, "/api/v1/serving?protocol="+protocol+"&model="+model+"&refresh=1", &s); w.Code != 200 {
		t.Fatalf("serving %s %s: %d %s", protocol, model, w.Code, w.Body)
	}
	return s
}

func TestWhoServesAModel(t *testing.T) {
	f := newModelsFixture(t)
	s := f.serves(t, "openai", "llama3:8b")
	if s.Answer != "one" || strings.Join(s.Serving, ",") != "local" || len(s.Unknown) != 0 || s.Model != "llama3:8b" || s.Protocol != "openai" {
		t.Fatalf("one Connection serves llama3:8b: %+v", s)
	}
	if s = f.serves(t, "openai", "gpt-4o-mini"); s.Answer != "one" || strings.Join(s.Serving, ",") != "hosted" {
		t.Fatalf("the hosted Connection serves gpt-4o-mini: %+v", s)
	}
	if s = f.serves(t, "openai", "no-such-model"); s.Answer != "none" || len(s.Serving) != 0 || len(s.Unknown) != 0 {
		t.Fatalf("nobody serves no-such-model: %+v", s)
	}
	f.spare.set(200, `{"data":[{"id":"mistral:7b"},{"id":"llama3:8b"}]}`)
	if s = f.serves(t, "openai", "llama3:8b"); s.Answer != "several" || strings.Join(s.Serving, ",") != "local,spare" {
		t.Fatalf("two Connections serve llama3:8b: %+v", s)
	}
}

func TestAFailedListingMakesTheAnswerUnknown(t *testing.T) {
	f := newModelsFixture(t)
	f.spare.set(500, `{"error":{"type":"server_error","message":"internal detail from the provider"}}`)
	s := f.serves(t, "openai", "llama3:8b")
	if s.Answer != "unknown" || strings.Join(s.Serving, ",") != "local" || strings.Join(s.Unknown, ",") != "spare" {
		t.Fatalf("a failed listing is unknown, not 'does not serve': %+v", s)
	}
	if s = f.serves(t, "openai", "no-such-model"); s.Answer != "unknown" || strings.Join(s.Unknown, ",") != "spare" {
		t.Fatalf("with a failed listing, 'none' is not certain: %+v", s)
	}
	// A listing that worked before and fails now is unknown too: the old
	// list is not trusted after a failure.
	f.spare.set(200, `{"data":[{"id":"mistral:7b"}]}`)
	f.serves(t, "openai", "mistral:7b")
	f.spare.set(404, `not found`)
	if s = f.serves(t, "openai", "mistral:7b"); s.Answer != "unknown" || len(s.Serving) != 0 {
		t.Fatalf("a Connection with no model list is unknown: %+v", s)
	}
	var l modelList
	f.get(t, "/api/v1/targets/spare/models", &l)
	if l.Listed || l.State != "no_model_list" || len(l.Models) != 0 {
		t.Fatalf("failed listing: %+v", l)
	}
	f.spare.srv.Close()
	if s = f.serves(t, "openai", "mistral:7b"); s.Answer != "unknown" || strings.Join(s.Unknown, ",") != "spare" {
		t.Fatalf("an unreachable Connection is unknown: %+v", s)
	}
}

func TestOnlyEnabledConnectionsOfTheProtocolAreConsidered(t *testing.T) {
	f := newModelsFixture(t)
	s := f.serves(t, "openai", "llama3:8b")
	for _, id := range append(s.Serving, s.Unknown...) {
		if id == "off" || id == "claude" {
			t.Fatalf("%s must not be considered: %+v", id, s)
		}
	}
	if f.off.count() != 0 || f.claude.count() != 0 {
		t.Fatalf("disabled and other-protocol Connections must not be listed: off %d, claude %d", f.off.count(), f.claude.count())
	}
	if s = f.serves(t, "anthropic", "llama3:8b"); s.Answer != "one" || strings.Join(s.Serving, ",") != "claude" {
		t.Fatalf("on the Anthropic protocol only claude counts: %+v", s)
	}
	var l modelList
	f.get(t, "/api/v1/targets/off/models", &l)
	if l.Listed || l.State != "disabled" || f.off.count() != 0 {
		t.Fatalf("a disabled Connection is not listed: %+v", l)
	}
}

func TestModelListingsKeepNoSecretsOrProviderMessages(t *testing.T) {
	f := newModelsFixture(t)
	f.hosted.set(401, `{"error":{"type":"authentication_error","message":"key sk-hosted-secret is wrong"}}`)
	f.spare.set(200, `{"data":[{"id":"mistral:7b","owned_by":"provider note"}]}`)
	for _, p := range []string{"/api/v1/targets/hosted/models", "/api/v1/targets/spare/models", "/api/v1/serving?protocol=openai&model=llama3:8b"} {
		b := f.get(t, p, nil).Body.String()
		for _, secret := range []string{"sk-hosted-secret", "is wrong", "provider note", "authentication_error"} {
			if strings.Contains(b, secret) {
				t.Fatalf("%s leaks %q: %s", p, secret, b)
			}
		}
	}
}

func TestModelRoutesNeedTheDashboard(t *testing.T) {
	f := newModelsFixture(t)
	for _, c := range []struct {
		method, path string
		csrf         bool
		want         int
	}{
		{"GET", "/api/v1/targets/local/models", false, 403},
		{"GET", "/api/v1/serving?protocol=openai&model=x", false, 403},
		{"POST", "/api/v1/targets/local/models", true, 405},
		{"GET", "/api/v1/targets/missing/models", true, 404},
		{"GET", "/api/v1/serving?protocol=gemini&model=x", true, 400},
		{"GET", "/api/v1/serving?protocol=openai", true, 400},
	} {
		r := httptest.NewRequest(c.method, "http://127.0.0.1:8787"+c.path, nil)
		if c.csrf {
			r.Header.Set("X-Aiport-CSRF", "1")
		}
		w := httptest.NewRecorder()
		f.h.ServeHTTP(w, r)
		if w.Code != c.want {
			t.Errorf("%s %s: %d, want %d", c.method, c.path, w.Code, c.want)
		}
	}
	if f.local.count() != 0 {
		t.Fatal("refused requests must not list models")
	}
}
