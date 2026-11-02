package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Refused inference requests are recorded with the outcome their refusal
// carries, not one guessed from the error code text.
func TestRefusalOutcomesReachHistory(t *testing.T) {
	hang := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(300 * time.Millisecond) }))
	defer hang.Close()
	c, err := config.Parse([]byte("version: 1\nauth: {token_env: LOCAL_TOKEN, disabled: true}\nproxy: {max_concurrent: 4, max_request_bytes: 1048576, connect_timeout_ms: 1000, response_header_timeout_ms: 50, stream_idle_timeout_ms: 1000, request_read_timeout_ms: 1000}\ntargets:\n - {id: slow, provider: ollama, protocol: openai, base_url: '" + hang.URL + "/v1'}\n"))
	if err != nil {
		t.Fatal(err)
	}
	collector := telemetry.New(true, 16)
	h := NewWithTelemetry(c, config.Secrets{}, collector)
	send := func(path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+path, strings.NewReader(`{"model":"m"}`))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := send("/v1/messages"); w.Code != 503 || !strings.Contains(w.Body.String(), `"code":"target_unconfigured"`) {
		t.Fatalf("unconfigured: %d %s", w.Code, w.Body)
	}
	c.Defaults.OpenAITarget = "slow"
	h.(*Handler).gateway.Apply(c)
	if w := send("/v1/chat/completions"); w.Code != 504 {
		t.Fatalf("timeout: %d %s", w.Code, w.Body)
	}
	collector.Close(time.Second)
	outcomes := map[string]string{}
	for _, e := range collector.Find(telemetry.Query{}) {
		outcomes[e.Route] = e.Outcome
	}
	if outcomes["/v1/messages"] != "rejected" || outcomes["/v1/chat/completions"] != "timeout" {
		t.Fatalf("outcomes: %v", outcomes)
	}
}
