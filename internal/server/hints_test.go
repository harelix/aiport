package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

// Hints in refusals: when an Application allowed to request changes hits a
// wall a Change could remove, aiport tells it how to ask. Status codes and
// error codes stay exactly as they are, and Applications without the
// permission see no hint.

const providerNotFound = `{"error":{"message":"The model 'nope' does not exist","type":"invalid_request_error"}}`

var (
	hintAllowed = strings.Repeat("a", 40)
	hintDenied  = strings.Repeat("d", 40)
)

func hintsServer(t *testing.T) *Handler { return hintsServerWith(t, "") }

// hintsServerWith is hintsServer with extra top-level YAML appended.
func hintsServerWith(t *testing.T, extra string) *Handler {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Model string }
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		if body.Model == "nope" {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(providerNotFound))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(upstream.Close)
	yaml := `version: 1
auth: {token_env: HINT_MGMT_TOKEN}
targets:
  - {id: up, provider: ollama, protocol: openai, base_url: '` + upstream.URL + `/v1'}
  - {id: off, provider: ollama, protocol: openai, disabled: true}
defaults: {openai_target: up}
origins:
  - id: helper
    name: Coding Helper
    token_env: HINT_ALLOWED_TOKEN
    may_request_changes: true
    aliases:
      - {alias: old, target: off, model: x}
  - id: nightly
    name: Nightly Report
    token_env: HINT_DENIED_TOKEN
    aliases:
      - {alias: old, target: off, model: x}
` + extra
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HINT_MGMT_TOKEN", strings.Repeat("m", 40))
	t.Setenv("HINT_ALLOWED_TOKEN", hintAllowed)
	t.Setenv("HINT_DENIED_TOKEN", hintDenied)
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
	return h
}

func infer(h *Handler, token, path, model string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+path, strings.NewReader(`{"model":"`+model+`","messages":[]}`))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func gatewayError(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out struct {
		Error map[string]any `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || out.Error == nil {
		t.Fatalf("not a gateway error: %d %s", w.Code, w.Body)
	}
	return out.Error
}

func TestRefusalsAChangeCouldFixCarryAHintOnlyForPermittedApplications(t *testing.T) {
	h := hintsServer(t)
	cases := []struct {
		name, path, model string
		status            int
		code, kind        string
	}{
		{"no Connection for this protocol", "/v1/messages", "claude-x", 503, "target_unconfigured", "bind_application"},
		{"a disabled Connection", "/v1/chat/completions", "old", 503, "target_disabled", "application_alias"},
	}
	for _, c := range cases {
		allowed := infer(h, hintAllowed, c.path, c.model)
		denied := infer(h, hintDenied, c.path, c.model)
		for _, w := range []*httptest.ResponseRecorder{allowed, denied} {
			if w.Code != c.status || gatewayError(t, w)["code"] != c.code {
				t.Fatalf("%s: status or code changed: %d %s", c.name, w.Code, w.Body)
			}
		}
		hint, _ := gatewayError(t, allowed)["hint"].(string)
		for _, want := range []string{"/v1/aiport/requests", c.kind, "aiport_request_change"} {
			if !strings.Contains(hint, want) {
				t.Fatalf("%s: hint %q does not mention %q", c.name, hint, want)
			}
		}
		if _, ok := gatewayError(t, denied)["hint"]; ok {
			t.Fatalf("%s: an Application without the permission got a hint: %s", c.name, denied.Body)
		}
	}
}

func TestAnUpstreamUnknownModelGetsAHintHeaderAndAnUnchangedBody(t *testing.T) {
	h := hintsServer(t)
	allowed := infer(h, hintAllowed, "/v1/chat/completions", "nope")
	denied := infer(h, hintDenied, "/v1/chat/completions", "nope")
	for _, w := range []*httptest.ResponseRecorder{allowed, denied} {
		if w.Code != 404 || w.Body.String() != providerNotFound {
			t.Fatalf("the provider's 404 must pass through unchanged: %d %s", w.Code, w.Body)
		}
	}
	hint := allowed.Header().Get("X-Aiport-Hint")
	for _, want := range []string{"/v1/aiport/requests", "application_alias", "aiport_request_change"} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint header %q does not mention %q", hint, want)
		}
	}
	if got := denied.Header().Get("X-Aiport-Hint"); got != "" {
		t.Fatalf("an Application without the permission got a hint: %q", got)
	}
	if ok := infer(h, hintAllowed, "/v1/chat/completions", "fine"); ok.Code != 200 || ok.Header().Get("X-Aiport-Hint") != "" {
		t.Fatalf("a successful call carries no hint: %d %q", ok.Code, ok.Header().Get("X-Aiport-Hint"))
	}
}

func TestWithAdaptingOffForwardingIsAsBeforeForEveryApplication(t *testing.T) {
	h := hintsServerWith(t, "adapting: {disabled: true}\n")
	for _, token := range []string{hintAllowed, hintDenied} {
		w := infer(h, token, "/v1/chat/completions", "nope")
		if w.Code != 404 || w.Body.String() != providerNotFound || w.Header().Get("X-Aiport-Hint") != "" {
			t.Fatalf("with adapting off the provider's 404 must pass through with no hint: %d %q %s", w.Code, w.Header().Get("X-Aiport-Hint"), w.Body)
		}
		refused := infer(h, token, "/v1/messages", "claude-x")
		if refused.Code != 503 || gatewayError(t, refused)["code"] != "target_unconfigured" {
			t.Fatalf("status or code changed: %d %s", refused.Code, refused.Body)
		}
		if _, ok := gatewayError(t, refused)["hint"]; ok {
			t.Fatalf("with adapting off a refusal must not hint at change requests: %s", refused.Body)
		}
	}
	w, caps := call(h, "GET", "/v1/aiport/capabilities", "", hintAllowed, false)
	if w.Code != 200 {
		t.Fatalf("capabilities: %d %s", w.Code, w.Body)
	}
	if kinds, _ := caps["change_kinds"].([]any); len(kinds) != 0 {
		t.Fatalf("with adapting off capabilities must list no change kinds: %s", w.Body)
	}
}

func TestForbiddingAnApplicationToRequestChangesStopsItsHintsAtOnce(t *testing.T) {
	h := hintsServer(t)
	if w := infer(h, hintAllowed, "/v1/chat/completions", "nope"); w.Header().Get("X-Aiport-Hint") == "" {
		t.Fatalf("an allowed Application gets a hint: %d %s", w.Code, w.Body)
	}
	if w, _ := call(h, "PATCH", "/api/v1/origins/helper", `{"may_request_changes":false}`, strings.Repeat("m", 40), true); w.Code != 200 {
		t.Fatalf("forbid: %d %s", w.Code, w.Body)
	}
	if w := infer(h, hintAllowed, "/v1/chat/completions", "nope"); w.Header().Get("X-Aiport-Hint") != "" {
		t.Fatalf("a forbidden Application still gets a hint: %q", w.Header().Get("X-Aiport-Hint"))
	}
	if w := infer(h, hintAllowed, "/v1/messages", "claude-x"); strings.Contains(w.Body.String(), "hint") {
		t.Fatalf("a forbidden Application still gets a refusal hint: %s", w.Body)
	}
}
