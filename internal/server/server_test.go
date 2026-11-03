package server

import (
	"crypto/sha256"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

func TestHTTPBoundary(t *testing.T) {
	token := strings.Repeat("a", 43)
	origin := strings.Repeat("b", 43)
	h := New(config.Config{Listen: "127.0.0.1:8787"}, config.Secrets{Credentials: []config.Credential{{Digest: sha256.Sum256([]byte(token)), Management: true}, {Digest: sha256.Sum256([]byte(origin)), OriginID: "app"}}})
	for _, tc := range []struct {
		name, method, path, host, bearer, origin, site string
		status                                         int
	}{
		{name: "health", status: 200, bearer: token},
		{name: "missing auth", status: 401},
		{name: "wrong auth", status: 401, bearer: "secret-canary"},
		{name: "origin cannot manage", status: 401, bearer: origin},
		{name: "dns rebinding", status: 403, bearer: token, host: "attacker.example:8787"},
		{name: "cross origin", status: 403, bearer: token, origin: "https://attacker.example"},
		{name: "null origin", status: 403, bearer: token, origin: "null"},
		{name: "same origin", status: 200, bearer: token, origin: "http://127.0.0.1:8787"},
		{name: "cross site", status: 403, bearer: token, site: "cross-site"},
		{name: "bad method", method: "POST", status: 405, bearer: token},
		{name: "unsupported", path: "/v1/files", status: 404, bearer: token},
		{name: "unconfigured inference", method: "POST", path: "/v1/chat/completions", status: 503, bearer: origin},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.method == "" {
				tc.method = "GET"
			}
			if tc.path == "" {
				tc.path = "/api/v1/health"
			}
			if tc.host == "" {
				tc.host = "127.0.0.1:8787"
			}
			r := httptest.NewRequest(tc.method, "http://"+tc.host+tc.path, strings.NewReader("payload-canary"))
			if tc.bearer != "" {
				r.Header.Set("Authorization", "Bearer "+tc.bearer)
			}
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			if tc.site != "" {
				r.Header.Set("Sec-Fetch-Site", tc.site)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			for _, canary := range []string{token, origin, "secret-canary", "payload-canary"} {
				if strings.Contains(w.Body.String(), canary) {
					t.Fatal("response leaked a canary")
				}
			}
			if len(w.Header().Get("X-Aiport-Request-Id")) != 36 {
				t.Fatal("missing UUID")
			}
			if w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatal("CORS enabled")
			}
			if !json.Valid(w.Body.Bytes()) {
				t.Fatal("not JSON")
			}
		})
	}
}
