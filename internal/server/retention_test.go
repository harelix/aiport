package server

import (
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

// Decided Suggestions and Change requests are kept as long as request
// history (telemetry.retention_days, 7 by default); items still waiting and
// the Changes themselves are kept.

// reopen starts aiport again over the configuration file at path and its
// database, as after a restart, with the clock at at.
func reopen(t *testing.T, path string, at time.Time) *Handler {
	t.Helper()
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	h.now = func() time.Time { return at }
	t.Cleanup(h.Close)
	return h
}

func myRequests(t *testing.T, h *Handler, token string) []any {
	t.Helper()
	w, out := call(h, "GET", RequestsPath, "", token, false)
	if w.Code != 200 {
		t.Fatalf("list requests: %d %s", w.Code, w.Body)
	}
	list, _ := out["requests"].([]any)
	return list
}

func TestDecidedItemsArePrunedWithRequestHistory(t *testing.T) {
	h, path := requestsServer(t)
	start := time.UnixMilli(1_800_000_000_000)
	decided := file(t, h, reqHelper, nil)
	if code, out := decideOn(t, h, reqManagement, decided, acceptFast); code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	waiting := file(t, h, reqHelper, map[string]any{"need": "Another alias"})

	sixDays := start.Add(6 * 24 * time.Hour)
	h.now = func() time.Time { return sixDays }
	h.Detect(sixDays)
	if list := myRequests(t, h, reqHelper); len(list) != 2 {
		t.Fatalf("within retention both requests are kept: %v", list)
	}

	eightDays := start.Add(8 * 24 * time.Hour)
	h.now = func() time.Time { return eightDays }
	h.Detect(eightDays)
	list := myRequests(t, h, reqHelper)
	if len(list) != 1 || list[0].(map[string]any)["id"] != waiting {
		t.Fatalf("after retention only the waiting request is kept: %v", list)
	}
	_, changes := call(h, "GET", "/api/v1/changes", "", reqManagement, false)
	if kept, _ := changes["changes"].([]any); len(kept) != 1 {
		t.Fatalf("the Change itself is kept: %v", changes)
	}

	again := reopen(t, path, eightDays)
	if list := myRequests(t, again, reqHelper); len(list) != 1 || list[0].(map[string]any)["id"] != waiting {
		t.Fatalf("a pruned request came back after a restart: %v", list)
	}
}
