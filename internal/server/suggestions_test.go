package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

const suggestionsYAML = "version: 1\nauth: {token_env: SUGG_MGMT_TOKEN}\ntargets:\n  - id: local\n    provider: ollama\n    base_url: http://127.0.0.1:9/v1\ndefaults: {openai_target: local}\norigins:\n  - id: nightly-report\n    name: Nightly Report\n    token_env: SUGG_NIGHTLY_TOKEN\n  - id: other\n    name: Other App\n    token_env: SUGG_OTHER_TOKEN\n"

var suggestionsNow = time.UnixMilli(1_800_000_000_000)

// suggestionsServer starts aiport over a configuration file (extra is appended
// to it) and a database next to it, with request history seeded by seed and
// a clock fixed at *at.
func suggestionsServer(t *testing.T, path, extra string, seed []telemetry.Event, at *time.Time) (*Handler, string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(suggestionsYAML+extra), 0600); err != nil {
			t.Fatal(err)
		}
	}
	management := strings.Repeat("m", 40)
	t.Setenv("SUGG_MGMT_TOKEN", management)
	t.Setenv("SUGG_NIGHTLY_TOKEN", strings.Repeat("n", 40))
	t.Setenv("SUGG_OTHER_TOKEN", strings.Repeat("o", 40))
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	collector := telemetry.New(true, 256)
	for _, e := range seed {
		collector.Offer(e)
	}
	collector.Close(time.Second)
	h := NewConfigured(c, s, collector, path).(*Handler)
	h.now = func() time.Time { return *at }
	t.Cleanup(h.Close)
	return h, management
}

var seq int

// upstream404s returns n requests from app for model that the Connection
// answered with 404, the first one ago before now and one a minute apart.
func upstream404s(app, model string, n int, ago time.Duration) []telemetry.Event {
	var out []telemetry.Event
	for i := 0; i < n; i++ {
		seq++
		m := model
		out = append(out, telemetry.Event{ID: fmt.Sprintf("00000000-0000-4000-8000-%012d", seq), StartedAt: suggestionsNow.Add(-ago + time.Duration(i)*time.Minute).UnixMilli(),
			Origin: app, Target: "local", Route: "/v1/chat/completions", Method: "POST", Status: 404, Outcome: "upstream_error", RequestedModel: &m})
	}
	return out
}

func pendingDecisions(t *testing.T, h *Handler, token string) (bool, []map[string]any) {
	t.Helper()
	w, out := call(h, "GET", "/api/v1/decisions", "", token, false)
	if w.Code != 200 {
		t.Fatalf("list decisions: %d %s", w.Code, w.Body)
	}
	var list []map[string]any
	for _, x := range out["items"].([]any) {
		list = append(list, x.(map[string]any))
	}
	on, _ := out["adapting"].(bool)
	return on, list
}

func seededHistory() []telemetry.Event {
	var seed []telemetry.Event
	seed = append(seed, upstream404s("nightly-report", "fast", 6, 50*time.Minute)...) // the pattern
	seed = append(seed, upstream404s("nightly-report", "slow", 2, 30*time.Minute)...) // too few
	seed = append(seed, upstream404s("other", "fast", 1, 20*time.Minute)...)          // another Application, too few
	seed = append(seed, upstream404s("nightly-report", "old", 9, 5*time.Hour)...)     // outside the window
	ok := "fast"
	seed = append(seed, telemetry.Event{ID: "00000000-0000-4000-9000-000000000001", StartedAt: suggestionsNow.Add(-55 * time.Minute).UnixMilli(), Origin: "nightly-report", Target: "local", Route: "/v1/chat/completions", Status: 200, Outcome: "success", RequestedModel: &ok})
	return seed
}

// Seeded history plus one detection pass at a fixed time produces exactly one
// Suggestion, for that Application and model, with its counts and window.
func TestUnknownModelSuggestion(t *testing.T) {
	at := suggestionsNow
	h, management := suggestionsServer(t, filepath.Join(t.TempDir(), "aiport.yaml"), "", seededHistory(), &at)
	if on, list := pendingDecisions(t, h, management); !on || len(list) != 0 {
		t.Fatalf("suggestions before any detection pass: %v", list)
	}
	h.Detect(at)
	h.Detect(at) // a second pass over the same history adds nothing
	_, list := pendingDecisions(t, h, management)
	if len(list) != 1 {
		t.Fatalf("want exactly one Suggestion, got %v", list)
	}
	s := list[0]
	evidence, _ := s["evidence"].(map[string]any)
	proposal, _ := s["proposal"].(map[string]any)
	if s["source"] != "suggestion" || s["kind"] != "application_alias" || s["application"] != "nightly-report" || s["status"] != "pending" {
		t.Fatalf("suggestion: %v", s)
	}
	if evidence["count"] != float64(6) || evidence["from_ms"] != float64(suggestionsNow.Add(-50*time.Minute).UnixMilli()) || evidence["to_ms"] != float64(suggestionsNow.Add(-45*time.Minute).UnixMilli()) || evidence["window_minutes"] != float64(60) {
		t.Fatalf("evidence: %v", evidence)
	}
	// local's model list cannot be read, so which Connection serves "fast"
	// is unknown: the Owner chooses.
	if proposal["alias"] != "fast" || proposal["target"] != nil || proposal["model"] != "fast" {
		t.Fatalf("proposal: %v", proposal)
	}
	if !strings.Contains(s["need"].(string), "fast") || !strings.Contains(s["need"].(string), "6") {
		t.Fatalf("need does not explain the evidence: %v", s["need"])
	}
	if w, _ := call(h, "GET", "/api/v1/decisions", "", strings.Repeat("n", 40), false); w.Code != 401 {
		t.Fatal("an Application token reached the decisions list")
	}
}

func decideOn(t *testing.T, h *Handler, management, id, body string) (int, map[string]any) {
	t.Helper()
	w, out := call(h, "POST", "/api/v1/decisions/"+id, body, management, true)
	return w.Code, out
}

// liveConfig is the running configuration as the dashboard reads it.
func liveConfig(t *testing.T, h *Handler, management string) config.Config {
	t.Helper()
	w, _ := call(h, "GET", "/api/v1/config", "", management, false)
	var out struct {
		Config config.Config `json:"config"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &out) != nil {
		t.Fatalf("read the running configuration: %d %s", w.Code, w.Body)
	}
	return out.Config
}

// appAliases are Application app's own Aliases in the running configuration.
// Every test server here uses the same management token.
func appAliases(t *testing.T, h *Handler, app string) []config.OriginAlias {
	t.Helper()
	o, _ := liveConfig(t, h, strings.Repeat("m", 40)).Application(app)
	return o.Aliases
}

// Approving a Suggestion creates the Alias for that Application only, live at
// once, as a Change the Owner can revert.
func TestApproveSuggestionCreatesRevertableAlias(t *testing.T) {
	at := suggestionsNow
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management := suggestionsServer(t, path, "", seededHistory(), &at)
	h.Detect(at)
	_, list := pendingDecisions(t, h, management)
	id := list[0]["id"].(string)
	before, _ := os.ReadFile(path)

	if w, _ := call(h, "POST", "/api/v1/decisions/"+id, `{"action":"accept","model":"llama3"}`, management, false); w.Code != 403 {
		t.Fatal("decision without CSRF header")
	}
	if code, out := decideOn(t, h, management, id, `{"action":"accept"}`); code != 400 || !strings.Contains(fmt.Sprint(out["error"]), "Connection") {
		t.Fatalf("accept without a Connection: %d %v", code, out)
	}
	code, out := decideOn(t, h, management, id, `{"action":"accept","target":"local","model":"llama3","note":"fast is llama3 here"}`)
	if code != 200 || out["change_id"] == "" {
		t.Fatalf("accept: %d %v", code, out)
	}
	if got := appAliases(t, h, "nightly-report"); len(got) != 1 || got[0] != (config.OriginAlias{Alias: "fast", Target: "local", Model: "llama3"}) {
		t.Fatalf("alias not live for nightly-report: %v", got)
	}
	if got := appAliases(t, h, "other"); len(got) != 0 || len(liveConfig(t, h, management).Aliases) != 0 {
		t.Fatal("the alias reached another Application or every Application")
	}
	if _, list := pendingDecisions(t, h, management); len(list) != 0 {
		t.Fatalf("accepted Suggestion still pending: %v", list)
	}
	if code, out := decideOn(t, h, management, id, `{"action":"dismiss"}`); code != 409 {
		t.Fatalf("decided twice: %d %v", code, out)
	}
	changes := listChanges(t, h, management)
	if len(changes) != 1 || changes[0]["id"] != out["change_id"] || !strings.Contains(changes[0]["what"].(string), "fast") {
		t.Fatalf("Change not recorded: %v", changes)
	}
	// Old 404s in the window don't bring the Suggestion back once the alias exists.
	h.Detect(at)
	if _, list := pendingDecisions(t, h, management); len(list) != 0 {
		t.Fatalf("suggested again after approval: %v", list)
	}
	if w, _ := call(h, "POST", "/api/v1/changes/"+changes[0]["id"].(string)+"/revert", "", management, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	if got, _ := os.ReadFile(path); string(got) != string(before) || len(appAliases(t, h, "nightly-report")) != 0 {
		t.Fatalf("revert did not remove the alias:\n%s", got)
	}
}

// Only dismissing starts the quiet period: a rejected Suggestion is
// suggested again on the next detection pass while the pattern lasts.
func TestARejectedSuggestionIsNotQuiet(t *testing.T) {
	at := suggestionsNow
	h, management := suggestionsServer(t, filepath.Join(t.TempDir(), "aiport.yaml"), "", seededHistory(), &at)
	h.Detect(at)
	_, list := pendingDecisions(t, h, management)
	if len(list) != 1 {
		t.Fatalf("want one Suggestion: %v", list)
	}
	if code, out := decideOn(t, h, management, list[0]["id"].(string), `{"action":"reject","note":"wrong idea"}`); code != 200 {
		t.Fatalf("reject: %d %v", code, out)
	}
	h.Detect(at)
	if _, again := pendingDecisions(t, h, management); len(again) != 1 || again[0]["id"] == list[0]["id"] {
		t.Fatalf("a rejected Suggestion stayed quiet: %v", again)
	}
}

// A dismissed Suggestion stays hidden for 7 days, also across a restart,
// unless its evidence doubles.
func TestDismissedSuggestionStaysQuiet(t *testing.T) {
	day := 24 * time.Hour
	var seed []telemetry.Event
	seed = append(seed, upstream404s("nightly-report", "fast", 6, 50*time.Minute)...)         // day 0: 6
	seed = append(seed, upstream404s("nightly-report", "fast", 11, -day+30*time.Minute)...)   // day 1: 11, not double
	seed = append(seed, upstream404s("nightly-report", "fast", 12, -2*day+30*time.Minute)...) // day 2: 12, double
	seed = append(seed, upstream404s("nightly-report", "fast", 6, -10*day+30*time.Minute)...) // day 10: quiet period over
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	at := suggestionsNow
	h, management := suggestionsServer(t, path, "", seed, &at)
	pending := func(h *Handler, when time.Time) []map[string]any {
		at = when
		h.Detect(when)
		_, list := pendingDecisions(t, h, management)
		return list
	}
	list := pending(h, suggestionsNow)
	if len(list) != 1 {
		t.Fatalf("day 0: %v", list)
	}
	if code, out := decideOn(t, h, management, list[0]["id"].(string), `{"action":"dismiss","note":"not now"}`); code != 200 {
		t.Fatalf("dismiss: %d %v", code, out)
	}
	if list := pending(h, suggestionsNow); len(list) != 0 {
		t.Fatalf("dismissed Suggestion came straight back: %v", list)
	}
	h.Close()

	h, management = suggestionsServer(t, path, "", seed, &at)
	if list := pending(h, suggestionsNow.Add(day)); len(list) != 0 {
		t.Fatalf("day 1, after a restart, evidence not doubled: %v", list)
	}
	list = pending(h, suggestionsNow.Add(2*day))
	if len(list) != 1 || list[0]["evidence"].(map[string]any)["count"] != float64(12) {
		t.Fatalf("day 2, evidence doubled: %v", list)
	}
	if code, out := decideOn(t, h, management, list[0]["id"].(string), `{"action":"dismiss"}`); code != 200 {
		t.Fatalf("dismiss: %d %v", code, out)
	}
	if list := pending(h, suggestionsNow.Add(10*day)); len(list) != 1 || list[0]["evidence"].(map[string]any)["count"] != float64(6) {
		t.Fatalf("day 10, 8 days after dismissal: %v", list)
	}
}

// With the adapting loop off in configuration, detection creates nothing and
// the dashboard is told there is nothing to show.
func TestAdaptingOff(t *testing.T) {
	at := suggestionsNow
	h, management := suggestionsServer(t, filepath.Join(t.TempDir(), "aiport.yaml"), "adapting: {disabled: true}\n", seededHistory(), &at)
	h.Detect(at)
	stop := make(chan struct{})
	h.StartAdapting(stop)
	close(stop)
	on, list := pendingDecisions(t, h, management)
	if on || len(list) != 0 {
		t.Fatalf("loop off, yet adapting=%v items=%v", on, list)
	}
	if code, out := decideOn(t, h, management, "anything", `{"action":"dismiss"}`); code != 409 || fmt.Sprint(out["error"]) == "" {
		t.Fatalf("decision with the loop off: %d %v", code, out)
	}
}

// Turning the adapting loop on in configuration starts detection without a
// restart: the background ticker runs from the start and checks each time.
func TestTurningAdaptingOnStartsDetectionWithoutARestart(t *testing.T) {
	at := suggestionsNow
	h, management := suggestionsServer(t, filepath.Join(t.TempDir(), "aiport.yaml"), "adapting: {disabled: true}\n", seededHistory(), &at)
	h.adaptEvery = 5 * time.Millisecond
	stop := make(chan struct{})
	defer close(stop)
	h.StartAdapting(stop)
	time.Sleep(30 * time.Millisecond)
	if on, list := pendingDecisions(t, h, management); on || len(list) != 0 {
		t.Fatalf("loop off, yet adapting=%v items=%v", on, list)
	}
	// No dashboard endpoint turns the loop on; the Owner edits aiport.yaml,
	// which reaches aiport as a Change.
	if _, _, err := h.live.Change("turn adapting on", replaceConfig(suggestionsYAML)); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		on, list := pendingDecisions(t, h, management)
		if on && len(list) == 1 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("no detection after turning the loop on: adapting=%v items=%v", on, list)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// Requests forwarded with the loop on and off get identical bytes back, and
// detection, run separately afterwards, finds the pattern in those requests.
func TestForwardingUnchangedByAdapting(t *testing.T) {
	var received []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received = append(received, r.URL.String()+" "+string(b))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error":{"message":"model \"fast\" not found"}}`))
	}))
	defer upstream.Close()
	forward := func(extra string) (*Handler, string, []string) {
		path := filepath.Join(t.TempDir(), "aiport.yaml")
		yaml := strings.Replace(suggestionsYAML+extra, "base_url: http://127.0.0.1:9/v1\n", "protocol: openai\n    base_url: "+upstream.URL+"/v1\n", 1)
		if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
			t.Fatal(err)
		}
		at := suggestionsNow
		h, management := suggestionsServer(t, path, "", nil, &at)
		h.collector = telemetry.New(true, 64)
		h.openDecisions()
		stop := make(chan struct{})
		defer close(stop)
		h.StartAdapting(stop)
		var got []string
		for i := 0; i < 5; i++ {
			r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", strings.NewReader(`{"model":"fast","messages":[]}`))
			r.Header.Set("Authorization", "Bearer "+strings.Repeat("n", 40))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			w.Header().Del("X-Aiport-Request-Id")
			w.Header().Del("Date") // the provider's clock, not aiport's doing
			got = append(got, fmt.Sprint(w.Code, w.Header(), w.Body.String()))
		}
		h.collector.Close(time.Second)
		return h, management, got
	}
	on, management, gotOn := forward("")
	_, _, gotOff := forward("adapting: {disabled: true}\n")
	if fmt.Sprint(gotOn) != fmt.Sprint(gotOff) || len(received) != 10 || fmt.Sprint(received[:5]) != fmt.Sprint(received[5:]) {
		t.Fatalf("forwarding differs with the loop on and off:\non:  %v\noff: %v\nupstream: %v", gotOn, gotOff, received)
	}
	if _, list := pendingDecisions(t, on, management); len(list) != 0 {
		t.Fatal("a Suggestion appeared without a detection pass")
	}
	on.Detect(time.Now())
	if _, list := pendingDecisions(t, on, management); len(list) != 1 || list[0]["application"] != "nightly-report" {
		t.Fatalf("detection over forwarded requests: %v", list)
	}
}
