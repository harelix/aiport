package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

func manageHandler(t *testing.T) (*Handler, string, func(method, path, body string) *httptest.ResponseRecorder) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	yaml := "version: 1\nauth: {token_env: AIPORT_T, disabled: true}\n# my connections\ntargets:\n  - id: bedrock\n    provider: bedrock\n    region: us-east-1\n"
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, config.Secrets{}, nil, path).(*Handler)
	send := func(method, p, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+p, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Aiport-CSRF", "1")
		w := httptest.NewRecorder()
		if !h.targetRoute(w, r, func(code int, c, m string) { w.WriteHeader(code); _, _ = w.Write([]byte(c + ": " + m)) }) {
			t.Fatal("route not matched", p)
		}
		return w
	}
	return h, path, send
}

func TestPatchTargetAppliesLiveAndPersists(t *testing.T) {
	h, path, send := manageHandler(t)
	if w := send("PATCH", "/api/v1/targets/bedrock", `{"region":"eu-west-1","endpoint":"mantle"}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if got := h.snapshot().Targets[0].URL(); got != "https://bedrock-mantle.eu-west-1.api.aws/v1" {
		t.Fatal("not applied live:", got)
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "endpoint: mantle") || !strings.Contains(string(data), "# my connections") {
		t.Fatal("not persisted in place")
	}
	if w := send("PATCH", "/api/v1/targets/bedrock", `{"region":"Not A Region"}`); w.Code != 400 {
		t.Fatal("bad region accepted")
	}
	if h.snapshot().Targets[0].Region != "eu-west-1" {
		t.Fatal("rejected update changed live config")
	}
}

func TestAddDefaultRemoveConnection(t *testing.T) {
	h, path, send := manageHandler(t)
	if w := send("POST", "/api/v1/targets", `{"id":"claude-eu","provider":"bedrock","region":"eu-central-1","protocol":"anthropic"}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := send("POST", "/api/v1/targets", `{"id":"claude-eu","provider":"anthropic"}`); w.Code != 409 {
		t.Fatal("duplicate id accepted", w.Code)
	}
	if w := send("POST", "/api/v1/targets", `{"id":"box","provider":"acme","protocol":"openai","base_url":"http://127.0.0.1:9000/v1"}`); w.Code != 200 {
		t.Fatal("custom local provider rejected", w.Body.String())
	}
	live := h.snapshot()
	if len(live.Targets) != 3 || live.Targets[1].URL() != "https://bedrock-runtime.eu-central-1.amazonaws.com/anthropic/v1" || live.Targets[2].Location != config.Local {
		t.Fatalf("not applied: %+v", live.Targets)
	}
	if _, src := h.keyStatus(); src["claude-eu"] != keyMissing || src["box"] != keyNotNeeded {
		t.Fatal("key sources not refreshed", src)
	}
	if w := send("PUT", "/api/v1/defaults", `{"anthropic_target":"bedrock"}`); w.Code != 400 {
		t.Fatal("protocol-incompatible default accepted")
	}
	if w := send("PUT", "/api/v1/defaults", `{"anthropic_target":"claude-eu"}`); w.Code != 200 || h.snapshot().Defaults.AnthropicTarget != "claude-eu" {
		t.Fatal("default not set", w.Body.String())
	}
	if w := send("DELETE", "/api/v1/targets/claude-eu", ""); w.Code != 400 || !strings.Contains(w.Body.String(), "default") {
		t.Fatal("removed a default connection", w.Body.String())
	}
	send("PUT", "/api/v1/defaults", `{}`)
	if w := send("DELETE", "/api/v1/targets/claude-eu", ""); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "claude-eu") || !strings.Contains(string(data), "# my connections") {
		t.Fatalf("file:\n%s", data)
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0600 {
		t.Fatal("permissions changed")
	}
	r := httptest.NewRequest("DELETE", "http://127.0.0.1:8787/api/v1/targets/box", nil)
	w := httptest.NewRecorder()
	h.targetRoute(w, r, func(code int, _, _ string) { w.WriteHeader(code) })
	if w.Code != 403 {
		t.Fatal("missing CSRF accepted")
	}
}
