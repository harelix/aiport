package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

func suggestConfig(t *testing.T) config.Config {
	t.Helper()
	yaml := "version: 1\nauth: {token_env: SUG_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: hosted, provider: acme, protocol: openai, base_url: 'http://127.0.0.1:1/v1'}\n" +
		"  - {id: cheap, provider: acme3, protocol: openai, base_url: 'http://127.0.0.1:2/v1'}\n" +
		"  - {id: claude, provider: acme2, protocol: anthropic, base_url: 'http://127.0.0.1:3/v1'}\n" +
		"aliases:\n" +
		"  - {target: cheap, alias: fast, model: small-model}\n" +
		"  - {target: hosted, alias: balanced, model: mid-model}\n" +
		"  - {target: hosted, alias: best, model: big-model}\n" +
		"origins:\n  - {id: docs-bot, name: Docs Bot, token_env: SUG_DOCS_TOKEN}\n  - {id: chat-helper, name: Chat Helper, token_env: SUG_CHAT_TOKEN}\n"
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func votes(up, down int) telemetry.Votes {
	v := telemetry.Votes{Calls: up + down, Rated: up + down, Up: up, Down: down}
	if v.Rated > 0 {
		r := float64(up) / float64(v.Rated)
		v.UpRatio = &r
	}
	return v
}

func row(name, connection, model string, v telemetry.Votes) telemetry.FeedbackRow {
	return telemetry.FeedbackRow{Name: name, Connection: connection, Model: model, Effort: "medium", Votes: v}
}

func TestVotesSuggestModuleRoutes(t *testing.T) {
	c := suggestConfig(t)
	apps := []telemetry.FeedbackApp{
		{Application: "docs-bot", Modules: []telemetry.FeedbackModule{
			// insights does badly on mid-model; big-model does clearly better for
			// chat-helper's insights (the same kind of work).
			{Module: "insights", Rows: []telemetry.FeedbackRow{row("balanced", "hosted", "mid-model", votes(4, 8))}},
			// tldr does badly on small-model and no other tldr does better (classify
			// doing well on big-model is other work): one Level up.
			{Module: "tldr", Rows: []telemetry.FeedbackRow{row("fast", "cheap", "small-model", votes(3, 9))}},
			// classify is rated very highly on best: the Level below may do.
			{Module: "classify", Rows: []telemetry.FeedbackRow{row("best", "hosted", "big-model", votes(19, 1))}},
			// too few votes: nothing.
			{Module: "extract", Rows: []telemetry.FeedbackRow{row("balanced", "hosted", "mid-model", votes(1, 5))}},
			// no module: never suggested.
			{Module: "", Rows: []telemetry.FeedbackRow{row("balanced", "hosted", "mid-model", votes(1, 20))}},
		}},
		{Application: "chat-helper", Modules: []telemetry.FeedbackModule{
			{Module: "insights", Rows: []telemetry.FeedbackRow{row("best", "hosted", "big-model", votes(17, 3))}},
			// Better, but on the other protocol: not evidence for an OpenAI module.
			{Module: "answers", Rows: []telemetry.FeedbackRow{row("best", "claude", "claude-large", votes(20, 0))}},
		}},
	}
	got := map[string]RouteSuggestion{}
	for _, s := range suggestRoutes(c, apps) {
		got[s.Application+"/"+s.Module] = s
	}
	if s := got["docs-bot/insights"]; s.Basis != "evidence" || s.To.Model != "big-model" || s.To.Connection != "hosted" || !strings.Contains(s.Why, "big-model gets 85% up over 20 votes in chat-helper / insights") {
		t.Fatalf("insights: %+v", s)
	}
	if s := got["docs-bot/tldr"]; s.Basis != "escalation" || s.To.Model != "mid-model" {
		t.Fatalf("tldr: %+v", s)
	}
	if s := got["docs-bot/classify"]; s.Basis != "cheaper" || s.To.Model != "mid-model" {
		t.Fatalf("classify: %+v", s)
	}
	for _, k := range []string{"docs-bot/extract", "docs-bot/", "chat-helper/answers", "chat-helper/insights"} {
		if s, ok := got[k]; ok {
			t.Errorf("%s: unexpected suggestion %+v", k, s)
		}
	}
}

// Applying a suggestion is a module route: that module's calls for that
// name go to the new model, the Application's other modules do not, and
// the Change can be reverted.
func TestAModuleRouteMovesOnlyThatModule(t *testing.T) {
	h, hosted, cheap := everywhereServer(t)
	put := `{"connection":"cheap","model":"other-model","why":"38% up over 12 votes"}`
	w, out := call(h, "PUT", "/api/v1/module-routes/docs-bot/insights/gpt-raw", put, lvlManagement, true)
	if w.Code != 200 {
		t.Fatalf("module route: %d %v", w.Code, out)
	}
	send := func(module string) {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", strings.NewReader(`{"model":"gpt-raw"}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+lvlDocs)
		r.Header.Set("X-Aiport-Module", module)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", module, w.Code, w.Body)
		}
	}
	send("insights")
	if _, body := cheap.received(); body != `{"model":"other-model"}` {
		t.Fatalf("insights went as %s", body)
	}
	send("tldr")
	if _, body := hosted.received(); body != `{"model":"gpt-raw"}` {
		t.Fatalf("tldr went as %s", body)
	}
	for _, bad := range []string{
		"/api/v1/module-routes/nobody/insights/gpt-raw",
		"/api/v1/module-routes/docs-bot/Not%20A%20Module/gpt-raw",
	} {
		if w, _ := call(h, "PUT", bad, put, lvlManagement, true); w.Code != 400 {
			t.Errorf("%s: %d, want 400", bad, w.Code)
		}
	}
	if w, _ := call(h, "DELETE", "/api/v1/module-routes/docs-bot/insights/gpt-raw", "", lvlManagement, true); w.Code != 200 {
		t.Fatalf("remove: %d", w.Code)
	}
	send("insights")
	if _, body := hosted.received(); body != `{"model":"gpt-raw"}` {
		t.Fatalf("after removal insights went as %s", body)
	}
}
