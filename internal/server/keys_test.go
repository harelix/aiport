package server

import (
	"crypto/sha256"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

func TestTargetKeyRoute(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	yaml := "version: 1\nauth: {token_env: LOCAL_TOKEN}\ntargets:\n - {id: meta, provider: meta, protocol: openai, base_url: 'https://api.meta.example/v1', api_key_env: TEST_META_KEY}\n - {id: local, provider: ollama, protocol: openai, base_url: 'http://127.0.0.1:11434/v1'}\ndefaults: {openai_target: meta}\n"
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("m", 43)
	t.Setenv("TEST_META_KEY", "")
	s, err := c.Resolve(func(n string) string { return map[string]string{"LOCAL_TOKEN": token}[n] })
	if err != nil || len(s.MissingKeys) != 1 {
		t.Fatalf("missing key must not block startup: %v %v", err, s.MissingKeys)
	}
	h := NewConfigured(c, s, nil, path)
	call := func(method, p, body string, csrf bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+p, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		if csrf {
			r.Header.Set("X-Aiport-CSRF", "1")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	source := func() map[string]string {
		var out struct {
			Source map[string]string `json:"secret_source"`
		}
		w := call("GET", "/api/v1/config", "", false)
		if strings.Contains(w.Body.String(), "sk-meta-secret") {
			t.Fatal("config response leaked the key")
		}
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return out.Source
	}
	if got := source(); got["meta"] != "missing" || got["local"] != "not_needed" {
		t.Fatalf("initial sources %v", got)
	}
	if w := call("POST", "/v1/chat/completions", `{"model":"m"}`, false); w.Code != 503 || !strings.Contains(w.Body.String(), "upstream_key_missing") {
		t.Fatalf("expected missing-key 503, got %d %s", w.Code, w.Body)
	}
	if w := call("PUT", "/api/v1/targets/meta/api-key", `{"api_key":"sk-meta-secret"}`, false); w.Code != 403 {
		t.Fatalf("CSRF header required, got %d", w.Code)
	}
	if w := call("PUT", "/api/v1/targets/meta/api-key", `{"api_key":"`+token+`"}`, true); w.Code != 400 {
		t.Fatalf("local token accepted as provider key: %d", w.Code)
	}
	if w := call("PUT", "/api/v1/targets/local/api-key", `{"api_key":"sk-x"}`, true); w.Code != 400 {
		t.Fatalf("keyless target accepted a key: %d", w.Code)
	}
	w := call("PUT", "/api/v1/targets/meta/api-key", `{"api_key":"sk-meta-secret"}`, true)
	if w.Code != 200 || strings.Contains(w.Body.String(), "sk-meta-secret") {
		t.Fatalf("save failed or leaked: %d %s", w.Code, w.Body)
	}
	saved, err := config.ReadCredentialFile(config.CredentialFile(path))
	if err != nil || saved["TEST_META_KEY"] != "sk-meta-secret" {
		t.Fatalf("key not persisted: %v", err)
	}
	if st, _ := os.Stat(config.CredentialFile(path)); st.Mode().Perm() != 0600 {
		t.Fatal("credential file not private")
	}
	if got := source(); got["meta"] != "saved" {
		t.Fatalf("after save %v", got)
	}
	// A restart picks the saved key back up.
	h2 := NewConfigured(c, config.Secrets{Credentials: []config.Credential{{Digest: sha256.Sum256([]byte(token)), Management: true}}, Upstream: map[string]string{"meta": "sk-meta-secret"}}, nil, path)
	if _, src := h2.(*Handler).keyStatus(); src["meta"] != "saved" {
		t.Fatalf("restart source %v", src)
	}
	if w := call("DELETE", "/api/v1/targets/meta/api-key", "", true); w.Code != 200 {
		t.Fatalf("delete failed: %d", w.Code)
	}
	if got := source(); got["meta"] != "missing" {
		t.Fatalf("after delete %v", got)
	}
	if saved, _ := config.ReadCredentialFile(config.CredentialFile(path)); saved["TEST_META_KEY"] != "" {
		t.Fatal("key not removed from file")
	}
}
