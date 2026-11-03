package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

func TestCompressionRecordedAndReported(t *testing.T) {
	var got string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		if r.Header.Get("X-Aiport-Compress") != "" {
			t.Error("compression header leaked upstream")
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()
	c, err := config.Parse([]byte("version: 1\nauth: {token_env: T, disabled: true}\ntargets:\n  - {id: up, provider: acme, protocol: openai, base_url: '" + upstream.URL + "/v1'}\ndefaults: {openai_target: up}\n"))
	if err != nil {
		t.Fatal(err)
	}
	collector := telemetry.New(true, 64)
	h := NewWithTelemetry(c, config.Secrets{}, collector)
	send := func(method, path, body string, header map[string]string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Aiport-CSRF", "1")
		for k, v := range header {
			r.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	body := `{"model":"m","messages":[{"role":"user","content":"hi   \n\n\n\n\nthere"}]}`
	if w := send("POST", "/v1/chat/completions", body, map[string]string{"X-Aiport-Compress": "nope"}); w.Code != 400 {
		t.Fatal("unknown method accepted", w.Code)
	}
	w := send("POST", "/v1/chat/completions", body, map[string]string{"X-Aiport-Compress": "whitespace"})
	if w.Code != 200 || !strings.Contains(got, `hi\n\nthere`) || !strings.Contains(w.Header().Get("X-Aiport-Compression"), "status=applied") {
		t.Fatal(w.Code, got, w.Header())
	}
	id := w.Header().Get("X-Aiport-Request-Id")
	send("POST", "/v1/chat/completions", body, nil) // uncompressed baseline
	deadline := time.Now().Add(2 * time.Second)
	for len(collector.Find(telemetry.Query{})) < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if f := send("POST", "/v1/aiport/feedback", `{"request_id":"`+id+`","score":0.9}`, nil); f.Code != 200 {
		t.Fatal("feedback", f.Code, f.Body.String())
	}
	if f := send("POST", "/v1/aiport/feedback", `{"request_id":"`+id+`","score":2}`, nil); f.Code != 400 {
		t.Fatal("out-of-range score accepted")
	}
	rep := send("GET", "/api/v1/compression?hours=24", "", nil)
	var report struct {
		Groups []struct {
			Key           string
			Requests      int
			FeedbackCount int `json:"feedback_count"`
			Verdict       string
			SavedRatio    float64 `json:"saved_ratio"`
		}
		Baseline struct{ Requests int }
	}
	if rep.Code != 200 || json.Unmarshal(rep.Body.Bytes(), &report) != nil {
		t.Fatal(rep.Code, rep.Body.String())
	}
	if len(report.Groups) != 1 || report.Groups[0].Key != "whitespace" || report.Groups[0].FeedbackCount != 1 || report.Groups[0].SavedRatio <= 0 || report.Groups[0].Verdict != "insufficient_data" || report.Baseline.Requests != 1 {
		t.Fatalf("report: %s", rep.Body.String())
	}
}
