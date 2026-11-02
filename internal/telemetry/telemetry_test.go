package telemetry

import (
	"fmt"
	"testing"
	"time"
)

func TestQueueDropIsImmediate(t *testing.T) {
	c := &Collector{enabled: true, queue: make(chan Event, 1), stop: make(chan struct{}), done: make(chan struct{})}
	c.Offer(Event{ID: "one"})
	start := time.Now()
	c.Offer(Event{ID: "two"})
	if time.Since(start) > 100*time.Millisecond || c.dropped.Load() != 1 {
		t.Fatal("full telemetry queue blocked or failed to report loss")
	}
}
func TestBoundedSessionAndFilters(t *testing.T) {
	c := New(true, 1024)
	for i := 0; i < 600; i++ {
		c.Offer(Event{ID: fmt.Sprintf("%036d", i), StartedAt: int64(i + 1), Origin: "codex", Target: "main", Provider: "openai", Outcome: "success", LatencyMS: int64(i)})
	}
	c.Close(time.Second)
	rows := c.Find(Query{Origin: "codex", Target: "main", Provider: "openai", Outcome: "success"})
	if len(rows) != 500 || c.evicted.Load() != 100 || rows[0].StartedAt != 600 {
		t.Fatalf("unbounded or unsorted history: %d", len(rows))
	}
	if len(c.Find(Query{Origin: "other"})) != 0 {
		t.Fatal("origin filter ignored")
	}
	summary := Summary(rows)
	if summary["estimated_cost_usd"] != nil || summary["input_tokens"] != nil {
		t.Fatal("unknown metrics fabricated")
	}
	c.Offer(Event{ID: "after-close"})
	if c.dropped.Load() != 1 {
		t.Fatal("offer after shutdown not counted")
	}
}
func TestDisabledAndErrorRate(t *testing.T) {
	c := New(false, 1)
	c.Offer(Event{ID: "not-recorded"})
	c.Close(time.Second)
	if len(c.Find(Query{Origin: ""})) != 0 {
		t.Fatal("disabled telemetry recorded a call")
	}
	s := Summary([]Event{{Outcome: "success", LatencyMS: 10}, {Outcome: "upstream_error", LatencyMS: 20}, {Outcome: "cancelled", LatencyMS: 30}, {Outcome: "rejected", LatencyMS: 40}})
	if s["error_rate"].(float64) != 1.0/3 || s["p95_latency_ms"] != int64(40) {
		t.Fatal("aggregate semantics")
	}
}
