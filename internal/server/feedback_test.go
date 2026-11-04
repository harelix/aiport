package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/telemetry"
)

// An Application names the section each call is for (X-Aiport-Task) and
// votes on answers; the Owner reads the votes per Application, section,
// name, model and effort.
func TestVotesAreReportedPerApplicationAndSection(t *testing.T) {
	h, _, _ := everywhereServer(t)
	send := func(task, model string) string {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", strings.NewReader(`{"model":"`+model+`"}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+lvlDocs)
		if task != "" {
			r.Header.Set("X-Aiport-Task", task)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("call: %d %s", w.Code, w.Body)
		}
		return w.Header().Get("X-Aiport-Request-Id")
	}
	vote := func(body string) int {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/aiport/feedback", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+lvlDocs)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	a, b := send("TLDR", "classify"), send("tldr", "classify")
	c := send("", "gpt-raw")
	send("Not a section!", "gpt-raw") // an invalid section is recorded as none
	deadline := time.Now().Add(2 * time.Second)
	for len(h.collector.Find(telemetry.Query{})) < 4 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	for body, want := range map[string]int{
		`{"request_id":"` + a + `","vote":"up"}`:                    200,
		`{"request_id":"` + b + `","vote":"down"}`:                  200,
		`{"request_id":"` + c + `","vote":"up","task":"insights"}`:  200, // the section, given with the vote
		`{"request_id":"` + a + `","vote":"sideways"}`:              400,
		`{"request_id":"` + a + `","vote":"up","task":"no spaces"}`: 400,
		`{"request_id":"unknown","vote":"up"}`:                      404,
	} {
		if got := vote(body); got != want {
			t.Errorf("%s: %d, want %d", body, got, want)
		}
	}
	w, _ := call(h, "GET", "/api/v1/feedback?hours=24", "", lvlManagement, false)
	var out struct {
		Applications []telemetry.FeedbackApp `json:"applications"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || len(out.Applications) != 1 {
		t.Fatalf("report: %d %s", w.Code, w.Body)
	}
	app := out.Applications[0]
	if app.Application != "docs-bot" || app.Calls != 4 || app.Rated != 3 || app.Up != 2 || app.Down != 1 {
		t.Fatalf("docs-bot: %+v", app.Votes)
	}
	got := map[string]telemetry.FeedbackSection{}
	for _, s := range app.Sections {
		got[s.Task] = s
	}
	tldr := got["tldr"]
	if tldr.Rated != 2 || tldr.Up != 1 || tldr.UpRatio == nil || *tldr.UpRatio != 0.5 || len(tldr.Rows) != 1 || tldr.Rows[0].Name != "classify" || tldr.Rows[0].Model != "small-model" || tldr.Rows[0].Effort != "medium" {
		t.Fatalf("tldr: %+v", tldr)
	}
	if s := got["insights"]; s.Rated != 1 || s.Up != 1 || s.Rows[0].Model != "gpt-raw" {
		t.Fatalf("insights: %+v", s)
	}
	if s := got[""]; s.Calls != 1 || s.Rated != 0 || s.UpRatio != nil {
		t.Fatalf("no section: %+v", s)
	}
	if last := app.Sections[len(app.Sections)-1]; last.Task != "" {
		t.Fatalf("calls without a section are not last: %+v", app.Sections)
	}
}
