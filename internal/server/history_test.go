package server

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

// The Activity window is measured from the handler's clock, and to_ms bounds
// the query itself rather than filtering after loading everything.
func TestActivityWindowUsesClock(t *testing.T) {
	c, err := config.Parse([]byte("version: 1\nauth: {token_env: LOCAL_TOKEN, disabled: true}\n"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.UnixMilli(1_800_000_000_000)
	collector := telemetry.New(true, 16)
	for i, age := range []time.Duration{10 * time.Minute, 90 * time.Minute, 30 * time.Hour} {
		collector.Offer(telemetry.Event{ID: fmt.Sprintf("00000000-0000-4000-8000-%012d", i), StartedAt: now.Add(-age).UnixMilli(), Route: "/v1/responses", Outcome: "success"})
	}
	collector.Close(time.Second)
	h := NewWithTelemetry(c, config.Secrets{}, collector).(*Handler)
	h.now = func() time.Time { return now }
	count := func(query string) int {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:8787/api/v1/requests?"+query, nil))
		var out struct{ Requests []telemetry.Event }
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 {
			t.Fatalf("%s: %d %s", query, w.Code, w.Body)
		}
		return len(out.Requests)
	}
	if n := count("hours=1"); n != 1 {
		t.Fatalf("hours=1: %d", n)
	}
	if n := count("hours=24"); n != 2 {
		t.Fatalf("hours=24: %d", n)
	}
	from, to := now.Add(-2*time.Hour).UnixMilli(), now.Add(-time.Hour).UnixMilli()
	if n := count(fmt.Sprintf("from_ms=%d&to_ms=%d", from, to)); n != 1 {
		t.Fatalf("explicit range: %d", n)
	}
}
