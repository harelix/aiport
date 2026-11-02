package telemetry

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

// Overview and Compression must agree on what an error is.
func TestOneErrorRule(t *testing.T) {
	route := "/v1/chat/completions"
	rows := []Event{
		{Route: route, Outcome: "success", StartedAt: 1},
		{Route: route, Outcome: "upstream_error", StartedAt: 1},
		{Route: route, Outcome: "cancelled", StartedAt: 1},
		{Route: route, Outcome: "rejected", StartedAt: 1},
	}
	summary := Summary(rows)
	report := CompressionReport(rows, 0, time.Hour)
	base := report["baseline"].(*Group)
	if summary["error_rate"].(float64) != 1.0/3 || base.ErrorRate != 1.0/3 || base.Requests != 3 {
		t.Fatalf("summary %.3f vs compression %.3f over %d", summary["error_rate"], base.ErrorRate, base.Requests)
	}
	if IsError("cancelled") || IsError("rejected") || !IsError("timeout") || Admitted("rejected") {
		t.Fatal("outcome rules")
	}
}

func TestQueryBoundsInMemoryAndSQLite(t *testing.T) {
	for _, persisted := range []bool{false, true} {
		t.Run(fmt.Sprint("persisted=", persisted), func(t *testing.T) {
			var c *Collector
			if persisted {
				c = NewPersistent(true, 64, openDB(t, filepath.Join(t.TempDir(), "h.db")))
			} else {
				c = New(true, 64)
			}
			for i, route := range []string{"/v1/chat/completions", "/v1/messages", "/api/v1/stats", "/v1/responses"} {
				c.Offer(Event{ID: fmt.Sprintf("%036d", i), StartedAt: int64(100 * (i + 1)), Route: route, Outcome: "success", Origin: "codex"})
			}
			c.Close(2 * time.Second)
			if n := len(c.Find(Query{From: 150, To: 350})); n != 2 {
				t.Fatalf("From/To: %d", n)
			}
			if n := len(c.Find(Query{InferenceOnly: true})); n != 3 {
				t.Fatalf("InferenceOnly: %d", n)
			}
			if got := c.Find(Query{To: 250}); len(got) != 2 || got[0].StartedAt != 200 {
				t.Fatalf("newest first within bound: %+v", got)
			}
		})
	}
}
