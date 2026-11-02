package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/store"
	"aiport/internal/telemetry"
)

// activityJSON is one request as /api/v1/requests returns it, with its reason.
type activityJSON struct {
	Route             string `json:"route"`
	Status            int    `json:"http_status"`
	Outcome           string `json:"outcome"`
	RefusalCode       string `json:"refusal_code"`
	UpstreamErrorType string `json:"upstream_error_type"`
	ConnectionError   string `json:"connection_error"`
	Reason            *struct {
		Code        string `json:"code"`
		Summary     string `json:"summary"`
		Explanation string `json:"explanation"`
		Fix         string `json:"fix"`
		Page        string `json:"page"`
		Section     string `json:"section"`
		Config      bool   `json:"config"`
	} `json:"reason"`
}

// reasonServer is a handler over yaml with an open collector; send makes an
// inference call, activity closes the collector and lists Activity.
type reasonServer struct {
	t         *testing.T
	h         *Handler
	collector *telemetry.Collector
}

func newReasonServer(t *testing.T, yaml string, collector *telemetry.Collector) *reasonServer {
	t.Helper()
	c, err := config.Parse([]byte("version: 1\nauth: {token_env: LOCAL_TOKEN, disabled: true}\nproxy: {max_concurrent: 4, max_request_bytes: 1048576, connect_timeout_ms: 1000, response_header_timeout_ms: 1000, stream_idle_timeout_ms: 1000, request_read_timeout_ms: 1000}\n" + yaml))
	if err != nil {
		t.Fatal(err)
	}
	if collector == nil {
		collector = telemetry.New(true, 16)
	}
	return &reasonServer{t: t, h: NewWithTelemetry(c, config.Secrets{}, collector).(*Handler), collector: collector}
}

func (s *reasonServer) send(path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, httptest.NewRequest("POST", "http://127.0.0.1:8787"+path, strings.NewReader(body)))
	return w
}

func (s *reasonServer) activity() []activityJSON {
	s.t.Helper()
	s.collector.Close(time.Second)
	return listActivity(s.t, s.h)
}

func listActivity(t *testing.T, h *Handler) []activityJSON {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/v1/requests", nil))
	var out struct{ Requests []activityJSON }
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 {
		t.Fatalf("requests: %d %s", w.Code, w.Body)
	}
	return out.Requests
}

// The Owner's report: a /v1/messages call refused because no Connection is
// the Anthropic protocol default shows why, and where to fix it.
func TestActivityShowsWhyARequestWasRefused(t *testing.T) {
	s := newReasonServer(t, "", nil)
	if w := s.send("/v1/messages", `{"model":"claude-x","messages":[{"role":"user","content":"secret prompt"}]}`); w.Code != 503 {
		t.Fatalf("refusal: %d %s", w.Code, w.Body)
	}
	rows := s.activity()
	if len(rows) != 1 {
		t.Fatalf("rows: %+v", rows)
	}
	r := rows[0]
	if r.RefusalCode != "target_unconfigured" || r.Reason == nil {
		t.Fatalf("row: %+v", r)
	}
	if r.Reason.Code != "target_unconfigured" || r.Reason.Summary != "Refused: no Connection for the Anthropic protocol" {
		t.Fatalf("summary: %+v", r.Reason)
	}
	if r.Reason.Page != "Routing" || r.Reason.Section != "Protocol defaults" || !r.Reason.Config || r.Reason.Fix == "" || r.Reason.Explanation == "" {
		t.Fatalf("fix: %+v", r.Reason)
	}
}

// A provider error is classified by the provider's own short error type or
// code, never by its message, and its body still reaches the Application
// unchanged.
func TestProviderErrorsKeepOnlyTheirErrorType(t *testing.T) {
	bodies := map[string]string{
		"/anthropic/v1/messages":        `{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key, you said: secret prompt"}}`,
		"/openai/v1/chat/completions":   `{"error":{"message":"The model gpt-x-secret does not exist","type":"invalid_request_error","param":null,"code":"model_not_found"}}`,
		"/ollama/v1/chat/completions":   `{"error":"model 'secret prompt' not found"}`,
		"/notjson/v1/chat/completions":  `upstream exploded: secret prompt`,
		"/longtype/v1/chat/completions": `{"error":{"type":"` + strings.Repeat("a", 65) + `"}}`,
	}
	status := map[string]int{"/anthropic/v1/messages": 401, "/openai/v1/chat/completions": 404, "/ollama/v1/chat/completions": 404, "/notjson/v1/chat/completions": 500, "/longtype/v1/chat/completions": 400}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status[r.URL.Path])
		_, _ = w.Write([]byte(bodies[r.URL.Path]))
	}))
	defer provider.Close()
	want := map[string]struct{ kind, page, summary string }{
		"anthropic": {"authentication_error", "Providers", "Provider error: the key was rejected (authentication_error, HTTP 401)"},
		"openai":    {"model_not_found", "", "Provider error: unknown model (model_not_found, HTTP 404)"},
		"ollama":    {"", "", "Provider error: unknown model (HTTP 404)"},
		"notjson":   {"", "", "Provider error: the provider failed (HTTP 500)"},
		"longtype":  {"", "", "Provider error: the request was not accepted (HTTP 400)"},
	}
	for name, expect := range want {
		protocol, path := "openai", "/v1/chat/completions"
		if name == "anthropic" {
			protocol, path = "anthropic", "/v1/messages"
		}
		s := newReasonServer(t, "targets:\n - {id: p, provider: ollama, protocol: "+protocol+", base_url: '"+provider.URL+"/"+name+"/v1'}\ndefaults: {"+protocol+"_target: p}\n", nil)
		w := s.send(path, `{"model":"m"}`)
		if w.Body.String() != bodies["/"+name+path] {
			t.Fatalf("%s: body changed: %s", name, w.Body)
		}
		rows := s.activity()
		if len(rows) != 1 || rows[0].Reason == nil {
			t.Fatalf("%s: %+v", name, rows)
		}
		r := rows[0]
		if r.UpstreamErrorType != expect.kind || r.RefusalCode != "" || r.Reason.Summary != expect.summary || r.Reason.Page != expect.page {
			t.Fatalf("%s: %+v %+v", name, r, r.Reason)
		}
		raw, _ := json.Marshal(r)
		if strings.Contains(string(raw), "secret prompt") || strings.Contains(string(raw), "gpt-x-secret") {
			t.Fatalf("%s: message text kept: %s", name, raw)
		}
	}
}

// unreachable returns base URLs where connecting fails in known ways: a port
// nobody listens on, and a TLS server whose certificate is not trusted.
func unreachable(t *testing.T) map[string]string {
	closed := httptest.NewServer(http.NotFoundHandler())
	refused := closed.URL + "/v1"
	closed.Close()
	untrusted := httptest.NewTLSServer(http.NotFoundHandler())
	t.Cleanup(untrusted.Close)
	return map[string]string{"connection_refused": refused, "tls_error": untrusted.URL + "/v1"}
}

// A connection that cannot be made is recorded as a fixed class, so Activity
// says what went wrong instead of only "gateway error".
func TestConnectionFailuresSayWhy(t *testing.T) {
	summaries := map[string]string{
		"connection_refused": "Cannot reach the provider: connection refused",
		"tls_error":          "Cannot reach the provider: the secure connection (TLS) failed",
	}
	for class, base := range unreachable(t) {
		s := newReasonServer(t, "targets:\n - {id: p, provider: ollama, protocol: openai, base_url: '"+base+"'}\ndefaults: {openai_target: p}\n", nil)
		if w := s.send("/v1/chat/completions", `{"model":"m"}`); w.Code != 502 || !strings.Contains(w.Body.String(), `"upstream_unavailable"`) {
			t.Fatalf("%s: %d %s", class, w.Code, w.Body)
		}
		rows := s.activity()
		if len(rows) != 1 || rows[0].Reason == nil {
			t.Fatalf("%s: %+v", class, rows)
		}
		r := rows[0]
		if r.RefusalCode != "upstream_unavailable" || r.ConnectionError != class || r.Reason.Code != class || r.Reason.Summary != summaries[class] || r.Reason.Page != "Providers" {
			t.Fatalf("%s: %+v %+v", class, r, r.Reason)
		}
	}
}

// Check connection says why it failed: the connection error class, the HTTP
// status and the provider's error type, with the same advice as Activity.
func TestCheckConnectionSaysWhyItFailed(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`))
	}))
	defer provider.Close()
	bases := unreachable(t)
	bases["key"] = provider.URL + "/v1"
	type probe struct {
		State             string `json:"state"`
		Status            int    `json:"http_status"`
		Message           string `json:"message"`
		ConnectionError   string `json:"connection_error"`
		UpstreamErrorType string `json:"upstream_error_type"`
		Reason            *struct {
			Code, Fix, Page string
		} `json:"reason"`
	}
	for name, base := range bases {
		s := newReasonServer(t, "targets:\n - {id: p, provider: ollama, protocol: anthropic, base_url: '"+base+"'}\n", nil)
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/v1/targets/p/check", nil)
		r.Header.Set("X-Aiport-CSRF", "1")
		w := httptest.NewRecorder()
		s.h.ServeHTTP(w, r)
		var out probe
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 || out.Reason == nil || out.Reason.Fix == "" || out.Reason.Page != "Providers" {
			t.Fatalf("%s: %d %s", name, w.Code, w.Body)
		}
		switch name {
		case "key":
			if out.State != "key_rejected" || out.Status != 401 || out.UpstreamErrorType != "authentication_error" || !strings.Contains(out.Message, "authentication_error") || strings.Contains(out.Message, "invalid x-api-key") {
				t.Fatalf("key: %+v", out)
			}
		case "connection_refused":
			if out.State != "unreachable" || out.ConnectionError != name || !strings.Contains(out.Message, "connection refused") {
				t.Fatalf("refused: %+v", out)
			}
		case "tls_error":
			if out.State != "unreachable" || out.ConnectionError != name || !strings.Contains(out.Message, "TLS") {
				t.Fatalf("tls: %+v", out)
			}
		}
	}
}

// The reason survives a restart: it is kept in SQLite with the request.
func TestRefusalReasonIsKeptAcrossRestarts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml.db")
	open := func() *telemetry.Collector {
		db, err := store.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { db.Close() })
		return telemetry.NewPersistent(true, 16, telemetry.Store{DB: db, RetentionDays: 30})
	}
	s := newReasonServer(t, "", open())
	s.send("/v1/messages", `{"model":"claude-x"}`)
	s.collector.Close(time.Second)

	rows := listActivity(t, newReasonServer(t, "", open()).h)
	if len(rows) != 1 || rows[0].RefusalCode != "target_unconfigured" || rows[0].Reason == nil || rows[0].Reason.Page != "Routing" {
		t.Fatalf("after restart: %+v", rows)
	}
}
