package telemetry

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/store"
)

func openDB(t *testing.T, path string) Store {
	t.Helper()
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return Store{DB: db, RetentionDays: 7, BatchSize: 10, FlushInterval: 10 * time.Millisecond}
}

func event(i int, at int64) Event {
	model := "gpt-x"
	ttfb := int64(5)
	return Event{ID: fmt.Sprintf("%036d", i), StartedAt: at, Origin: "codex", Target: "main", Provider: "openai", Protocol: "openai", Route: "/v1/chat/completions", Method: "POST", Status: 200, Outcome: "success", LatencyMS: 12, TTFBMS: &ttfb, RequestedModel: &model, UsageStatus: "unobserved", CostStatus: "unknown", Compression: "json", CompressionStatus: "applied", BytesBefore: 100, BytesAfter: 80, Effort: "high", EffortSource: "field", ResolvedModel: "gpt-x-large"}
}

func TestHistorySurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml.db")
	now := time.Now().UnixMilli()
	c := NewPersistent(true, 1024, openDB(t, path))
	for i := 0; i < 600; i++ {
		c.Offer(event(i, now-int64(600-i)))
	}
	c.Close(2 * time.Second)
	if h := c.Health(); h["state"] != "persisting" || h["failed_writes"].(uint64) != 0 || h["evicted_events"].(uint64) != 0 {
		t.Fatalf("health: %v", h)
	}

	again := NewPersistent(true, 16, openDB(t, path))
	defer again.Close(time.Second)
	rows := again.Find(Query{Origin: "codex"})
	if len(rows) != 600 || rows[0].ID != fmt.Sprintf("%036d", 599) {
		t.Fatalf("restart lost history: %d rows", len(rows))
	}
	if rows[0].TTFBMS == nil || *rows[0].TTFBMS != 5 || rows[0].RequestedModel == nil || rows[0].Compression != "json" || rows[0].BytesAfter != 80 || rows[0].Effort != "high" || rows[0].EffortSource != "field" || rows[0].ResolvedModel != "gpt-x-large" {
		t.Fatalf("columns not round-tripped: %+v", rows[0])
	}
	if len(again.Find(Query{Model: "gpt-x"})) != 600 || len(again.Find(Query{Model: "gpt-x-large"})) != 0 {
		t.Fatal("the filter by the name sent is wrong in SQLite")
	}
	if len(again.Find(Query{Origin: "other"})) != 0 {
		t.Fatal("origin filter ignored in SQLite")
	}
	if e, ok := again.Get(fmt.Sprintf("%036d", 3)); !ok || e.StartedAt != now-597 {
		t.Fatal("Get missed a persisted row")
	}
}

func TestFeedbackBeforeAndAfterFlush(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml.db")
	s := openDB(t, path)
	s.FlushInterval = time.Hour // nothing flushes until Close
	s.BatchSize = 1000
	c := NewPersistent(true, 16, s)
	c.Offer(event(1, time.Now().UnixMilli()))
	deadline := time.Now().Add(time.Second)
	for len(c.Find(Query{Origin: ""})) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !c.SetFeedback(event(1, 0).ID, "codex", 0.75, "") || c.SetFeedback(event(1, 0).ID, "someone-else", 0.1, "") {
		t.Fatal("feedback ownership on pending row")
	}
	c.Close(time.Second)

	again := NewPersistent(true, 16, openDB(t, path))
	defer again.Close(time.Second)
	e, ok := again.Get(event(1, 0).ID)
	if !ok || e.Feedback == nil || *e.Feedback != 0.75 {
		t.Fatal("feedback on an unflushed row was lost")
	}
	if !again.SetFeedback(e.ID, "codex", 0.25, "tldr") {
		t.Fatal("feedback on a persisted-only row not found")
	}
	if e, _ = again.Get(e.ID); *e.Feedback != 0.25 {
		t.Fatal("feedback update not visible")
	}
}

func TestRetentionPrunesOldRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml.db")
	old := time.Now().Add(-8 * 24 * time.Hour).UnixMilli()
	c := NewPersistent(true, 16, openDB(t, path))
	c.Offer(event(1, old))
	c.Offer(event(2, time.Now().UnixMilli()))
	c.Close(time.Second)

	again := NewPersistent(true, 16, openDB(t, path)) // prunes on start
	again.Close(time.Second)
	saved, err := again.query(Query{})
	if err != nil || len(saved) != 1 || saved[0].ID != event(2, 0).ID {
		t.Fatalf("retention: %v %d", err, len(saved))
	}
}
