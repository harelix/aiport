package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aiport/internal/config"
)

func TestRuntimeKeyUpdate(t *testing.T) {
	c, err := config.Parse([]byte(`version: 1
auth: {token_env: LOCAL_TOKEN}
targets:
 - {id: claude, provider: anthropic, protocol: anthropic, base_url: 'https://anthropic.example/v1', api_key_env: ANTHROPIC_API_KEY}
defaults: {anthropic_target: claude}
`))
	if err != nil {
		t.Fatal(err)
	}
	g := New(c, config.Secrets{Upstream: map[string]string{}})
	seen := ""
	g.transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		seen = r.Header.Get("X-Api-Key")
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{}"))}, nil
	})
	send := func() (int, string) {
		code := ""
		w := httptest.NewRecorder()
		g.Forward(w, httptest.NewRequest("POST", "http://local/v1/messages", strings.NewReader("{}")), "unknown", func(status int, c, _ string) { code = c; w.WriteHeader(status) })
		return w.Code, code
	}
	if status, code := send(); status != 503 || code != "upstream_key_missing" {
		t.Fatalf("missing key: %d %s", status, code)
	}
	g.SetKey("claude", "sk-test-value")
	if status, _ := send(); status != 200 || seen != "sk-test-value" {
		t.Fatalf("key not applied: %d %q", status, seen)
	}
	g.SetKey("claude", "")
	if status, _ := send(); status != 503 {
		t.Fatal("removed key still used")
	}
}
