package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"aiport/internal/config"
	"aiport/internal/connect"
)

func TestConnectDocument(t *testing.T) {
	c, err := config.Parse([]byte("version: 1\nauth: {token_env: LOCAL_TOKEN}\ntargets:\n - {id: anthropic, provider: anthropic, protocol: anthropic, base_url: 'https://api.anthropic.example/v1', api_key_env: SECRET_KEY}\ndefaults: {anthropic_target: anthropic}\n"))
	if err != nil {
		t.Fatal(err)
	}
	h := New(c, config.Secrets{Upstream: map[string]string{"anthropic": "sk-very-secret"}})
	get := func(path, accept string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "http://127.0.0.1:8787"+path, nil)
		if accept != "" {
			r.Header.Set("Accept", accept)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	w := get("/_connect", "")
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("json: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	var doc connect.Document
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil || doc.Protocol != "aiport-connect" || doc.BaseURL != "http://127.0.0.1:8787/v1" || !doc.Auth.Required {
		t.Fatalf("bad document: %v %+v", err, doc)
	}
	if doc.Protocols[0].Available || !doc.Protocols[1].Available {
		t.Fatal("protocol availability must follow configured defaults")
	}
	for _, leak := range []string{"sk-very-secret", "SECRET_KEY", "api.anthropic.example", "LOCAL_TOKEN"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Fatalf("document leaks %q", leak)
		}
	}
	md := get("/_connect", "text/markdown")
	if md.Code != 200 || !strings.HasPrefix(md.Header().Get("Content-Type"), "text/markdown") || !strings.Contains(md.Body.String(), "# aiport Connect v1") {
		t.Fatalf("markdown: %d", md.Code)
	}
	if get("/_connect?format=markdown", "").Code != 200 || get("/_connect?format=xml", "").Code != 400 {
		t.Fatal("format parameter")
	}
	r := httptest.NewRequest("GET", "http://evil.example/_connect", nil)
	foreign := httptest.NewRecorder()
	h.ServeHTTP(foreign, r)
	if foreign.Code != 403 {
		t.Fatal("foreign Host must still be refused")
	}
}
