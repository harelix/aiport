package server

import (
	"crypto/sha256"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

func TestDashboardAndAuthenticatedMetadata(t *testing.T) {
	c, err := config.Parse([]byte("version: 1\nauth: {token_env: LOCAL_TOKEN}\n"))
	if err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("m", 43)
	collector := telemetry.New(true, 16)
	collector.Offer(telemetry.Event{ID: "00000000-0000-4000-8000-000000000002", StartedAt: time.Now().UnixMilli(), Origin: "codex", Outcome: "success"})
	collector.Offer(telemetry.Event{ID: "00000000-0000-4000-8000-000000000001", StartedAt: time.Now().UnixMilli() - 1, Origin: "unknown", Outcome: "success"})
	collector.Close(time.Second)
	h := NewWithTelemetry(c, config.Secrets{Credentials: []config.Credential{{Digest: sha256.Sum256([]byte(token)), Management: true}}}, collector)
	request := func(path string, auth bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "http://127.0.0.1:8787"+path, nil)
		if auth {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	page := request("/", false)
	if page.Code != 200 || !strings.Contains(page.Body.String(), "aiport") {
		t.Fatal("embedded page unavailable")
	}
	if !strings.Contains(page.Header().Get("Content-Security-Policy"), "script-src 'self'") {
		t.Fatal("asset CSP missing")
	}
	if request("/api/v1/config", false).Code != 401 {
		t.Fatal("public configuration exposed")
	}
	list := request("/api/v1/requests?limit=1", true)
	var result struct {
		Requests []telemetry.Event `json:"requests"`
		Cursor   string            `json:"next_cursor"`
	}
	if json.Unmarshal(list.Body.Bytes(), &result) != nil || len(result.Requests) != 1 || result.Cursor == "" {
		t.Fatal("pagination failed", list.Body.String())
	}
	older := request("/api/v1/requests?limit=1&cursor="+result.Cursor, true)
	if older.Code != 200 || strings.Contains(older.Body.String(), result.Requests[0].ID) {
		t.Fatal("cursor duplicates")
	}
	for _, path := range []string{"/api/v1/stats?hours=0", "/api/v1/requests?limit=101", "/api/v1/requests?cursor=bad", "/api/v1/stats?unknown=x"} {
		if request(path, true).Code != 400 {
			t.Fatal("invalid filter accepted", path)
		}
	}
	settings := request("/api/v1/config", true)
	if !strings.Contains(settings.Body.String(), `"targets":[]`) || !strings.Contains(settings.Body.String(), `"origins":[]`) {
		t.Fatal("empty settings collections must remain iterable JSON arrays")
	}
	if strings.Contains(settings.Body.String(), token) {
		t.Fatal("config leaked credential")
	}
}
