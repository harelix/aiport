package server

import (
	"strconv"
	"testing"
	"time"
)

// Activity records what was asked and what served it: the name sent, the
// effort and where it came from, the Connection and the model sent upstream.

// activityOf is docs-bot's Activity around now, with extra filters (such as
// "&model=balanced"), once want rows are there.
func activityOf(t *testing.T, h *Handler, extra string, want int) []map[string]any {
	t.Helper()
	from, to := time.Now().Add(-time.Hour).UnixMilli(), time.Now().Add(time.Hour).UnixMilli()
	path := "/api/v1/requests?origin=docs-bot&from_ms=" + strconv.FormatInt(from, 10) + "&to_ms=" + strconv.FormatInt(to, 10) + extra
	var rows []map[string]any
	waitFor(t, func() bool {
		w, act := call(h, "GET", path, "", lvlManagement, false)
		if w.Code != 200 {
			t.Fatalf("activity: %d %s", w.Code, w.Body)
		}
		list, _ := act["requests"].([]any)
		rows = nil
		for _, r := range list {
			rows = append(rows, m(r))
		}
		return len(rows) >= want
	})
	if len(rows) != want {
		t.Fatalf("activity%s: %d rows, want %d: %v", extra, len(rows), want, rows)
	}
	return rows
}

func TestActivityRecordsTheNameEffortAndModelOfEachRequest(t *testing.T) {
	f := effortServer(t)
	cases := []struct {
		path, header, body                     string
		name, effort, source, target, resolved string
	}{
		{"/v1/chat/completions", "low", bodyOf(chatEffortBody, "balanced", `"high"`), "balanced", "low", "header", "hosted", "gpt-mini"},
		{"/v1/messages", "", bodyOf(messagesEffortBody, "balanced", `"high"`), "balanced", "high", "field", "claude", "claude-large"},
		{"/v1/responses", "", `{"model":"balanced","input":"hi"}`, "balanced", "medium", "default", "hosted", "gpt-mid"},
		{"/v1/chat/completions", "", `{"model":"gpt-other","messages":[]}`, "gpt-other", "medium", "default", "hosted", "gpt-other"},
	}
	for i, c := range cases {
		w := effortCall(f.h, c.path, c.header, c.body)
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", c.path, w.Code, w.Body)
		}
		id := w.Header().Get("X-Aiport-Request-Id")
		var row map[string]any
		for _, r := range activityOf(t, f.h, "", i+1) {
			if r["id"] == id {
				row = r
			}
		}
		if row == nil || row["requested_model"] != c.name || row["effort"] != c.effort || row["effort_source"] != c.source || row["target_id"] != c.target || row["resolved_model"] != c.resolved {
			t.Fatalf("case %d: Activity row %v", i, row)
		}
		// The request's details carry the same.
		_, one := call(f.h, "GET", "/api/v1/requests/"+id, "", lvlManagement, false)
		if one["requested_model"] != c.name || one["effort"] != c.effort || one["effort_source"] != c.source || one["resolved_model"] != c.resolved {
			t.Fatalf("case %d: details %v", i, one)
		}
	}
}

func TestActivityFiltersByTheNameSent(t *testing.T) {
	f := effortServer(t)
	for _, model := range []string{"balanced", "quick", "balanced"} {
		if w := effortCall(f.h, "/v1/chat/completions", "", `{"model":"`+model+`","messages":[]}`); w.Code != 200 {
			t.Fatalf("%s: %d", model, w.Code)
		}
	}
	activityOf(t, f.h, "", 3)
	for _, r := range activityOf(t, f.h, "&model=balanced", 2) {
		if r["requested_model"] != "balanced" {
			t.Fatalf("the filter let through %v", r)
		}
	}
	if rows := activityOf(t, f.h, "&model=quick", 1); rows[0]["resolved_model"] != "gpt-mini" {
		t.Fatalf("quick: %v", rows[0])
	}
}

func TestARequestNotForwardedRecordsNoServingModel(t *testing.T) {
	f := effortServer(t)
	// Two model members: aiport cannot tell which to rewrite and refuses.
	w := effortCall(f.h, "/v1/chat/completions", "high", `{"model":"balanced","model":"quick"}`)
	if w.Code == 200 {
		t.Fatalf("an ambiguous model was forwarded: %s", w.Body)
	}
	row := activityOf(t, f.h, "", 1)[0]
	if _, ok := row["resolved_model"]; ok || row["effort"] != "high" || row["effort_source"] != "header" {
		t.Fatalf("refused row: %v", row)
	}
}
