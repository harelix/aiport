package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/telemetry"
)

// Three Connections; nightly-report is bound to local, which keeps failing.
const failingYAML = "version: 1\nauth: {token_env: SUGG_MGMT_TOKEN}\ntargets:\n  - id: local\n    provider: ollama\n  - id: backup\n    provider: ollama\n  - id: spare\n    provider: ollama\ndefaults: {openai_target: local}\norigins:\n  - id: nightly-report\n    name: Nightly Report\n    token_env: SUGG_NIGHTLY_TOKEN\n    target: local\n  - id: other\n    name: Other App\n    token_env: SUGG_OTHER_TOKEN\n"

// calls returns n requests from app to Connection target with outcome and
// status, the first one ago before now and one a minute apart.
func calls(app, target, outcome string, status, n int, ago time.Duration) []telemetry.Event {
	var out []telemetry.Event
	for i := 0; i < n; i++ {
		seq++
		out = append(out, telemetry.Event{ID: fmt.Sprintf("00000000-0000-4000-a000-%012d", seq), StartedAt: suggestionsNow.Add(-ago + time.Duration(i)*time.Minute).UnixMilli(),
			Origin: app, Target: target, Route: "/v1/chat/completions", Method: "POST", Status: status, Outcome: outcome})
	}
	return out
}

func failingServer(t *testing.T, seed []telemetry.Event) (*Handler, string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(failingYAML), 0600); err != nil {
		t.Fatal(err)
	}
	at := suggestionsNow
	h, management := suggestionsServer(t, path, "", seed, &at)
	return h, management, path
}

func failingHistory() []telemetry.Event {
	var seed []telemetry.Event
	seed = append(seed, calls("nightly-report", "local", "gateway_error", 502, 4, 50*time.Minute)...) // the pattern: 4 gateway errors
	seed = append(seed, calls("nightly-report", "local", "timeout", 504, 3, 40*time.Minute)...)       // and 3 timeouts
	seed = append(seed, calls("nightly-report", "local", "cancelled", 0, 9, 30*time.Minute)...)       // not failures
	seed = append(seed, calls("nightly-report", "local", "rejected", 429, 9, 30*time.Minute)...)      // not failures
	seed = append(seed, calls("nightly-report", "local", "gateway_error", 502, 9, 5*time.Hour)...)    // outside the window
	seed = append(seed, calls("other", "local", "timeout", 504, 2, 20*time.Minute)...)                // another Application, too few
	seed = append(seed, calls("other", "backup", "success", 200, 3, 20*time.Minute)...)               // backup works
	return seed
}

// Repeated failures for one Application and Connection produce one
// Suggestion naming both, with the counts and the window.
func TestFailingConnectionSuggestion(t *testing.T) {
	h, management, _ := failingServer(t, failingHistory())
	h.Detect(suggestionsNow)
	h.Detect(suggestionsNow)
	_, list := pendingDecisions(t, h, management)
	if len(list) != 1 {
		t.Fatalf("want exactly one Suggestion, got %v", list)
	}
	s := list[0]
	evidence, _ := s["evidence"].(map[string]any)
	proposal, _ := s["proposal"].(map[string]any)
	if s["source"] != "suggestion" || s["kind"] != "bind_application" || s["application"] != "nightly-report" || s["subject"] != "local" || s["scope"] != "application" {
		t.Fatalf("suggestion: %v", s)
	}
	if evidence["count"] != float64(7) || evidence["from_ms"] != float64(suggestionsNow.Add(-50*time.Minute).UnixMilli()) || evidence["to_ms"] != float64(suggestionsNow.Add(-38*time.Minute).UnixMilli()) || evidence["window_minutes"] != float64(60) {
		t.Fatalf("evidence: %v", evidence)
	}
	if proposal["target"] != "backup" {
		t.Fatalf("proposal should offer the Connection that works: %v", proposal)
	}
	need := s["need"].(string)
	for _, want := range []string{"nightly-report", "local", "7", "60 minutes"} {
		if !strings.Contains(need, want) {
			t.Fatalf("need %q does not mention %q", need, want)
		}
	}
}

// Cancellations and refused requests are not failures, however many there are.
func TestCancelledAndRefusedAreNotFailures(t *testing.T) {
	var seed []telemetry.Event
	seed = append(seed, calls("nightly-report", "local", "cancelled", 0, 20, 50*time.Minute)...)
	seed = append(seed, calls("nightly-report", "local", "rejected", 401, 20, 50*time.Minute)...)
	seed = append(seed, calls("nightly-report", "local", "gateway_error", 502, 4, 10*time.Minute)...) // one short of the threshold
	h, management, _ := failingServer(t, seed)
	h.Detect(suggestionsNow)
	if _, list := pendingDecisions(t, h, management); len(list) != 0 {
		t.Fatalf("cancelled or refused requests counted as failures: %v", list)
	}
}

// The Owner chooses the Connection when approving; the binding is live at once
// and the Change can be reverted.
func TestApproveBindingIsLiveAndRevertable(t *testing.T) {
	h, management, path := failingServer(t, failingHistory())
	h.Detect(suggestionsNow)
	_, list := pendingDecisions(t, h, management)
	id := list[0]["id"].(string)
	before, _ := os.ReadFile(path)
	app := strings.Repeat("n", 40)

	if code, out := decideOn(t, h, management, id, `{"action":"accept","target":"local"}`); code != 400 || !strings.Contains(fmt.Sprint(out["error"]), "another Connection") {
		t.Fatalf("binding to the failing Connection itself: %d %v", code, out)
	}
	if code, out := decideOn(t, h, management, id, `{"action":"accept","target":"nowhere"}`); code != 400 {
		t.Fatalf("binding to a missing Connection: %d %v", code, out)
	}
	code, out := decideOn(t, h, management, id, `{"action":"accept","target":"spare","note":"backup is for other"}`)
	if code != 200 || out["change_id"] == "" {
		t.Fatalf("accept: %d %v", code, out)
	}
	if _, caps := call(h, "GET", CapabilitiesPath, "", app, false); caps["binding"] != "spare" {
		t.Fatalf("binding not live: %v", caps["binding"])
	}
	if _, caps := call(h, "GET", CapabilitiesPath, "", strings.Repeat("o", 40), false); caps["binding"] != nil {
		t.Fatalf("another Application was bound: %v", caps["binding"])
	}
	changes := listChanges(t, h, management)
	if len(changes) != 1 || changes[0]["id"] != out["change_id"] || !strings.Contains(changes[0]["what"].(string), "spare") {
		t.Fatalf("Change not recorded: %v", changes)
	}
	// The old failures to local no longer matter: nightly-report is bound elsewhere.
	h.Detect(suggestionsNow)
	if _, list := pendingDecisions(t, h, management); len(list) != 0 {
		t.Fatalf("suggested again after approval: %v", list)
	}
	if w, _ := call(h, "POST", "/api/v1/changes/"+changes[0]["id"].(string)+"/revert", "", management, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	if got, _ := os.ReadFile(path); string(got) != string(before) {
		t.Fatalf("revert did not restore the configuration:\n%s", got)
	}
	if _, caps := call(h, "GET", CapabilitiesPath, "", app, false); caps["binding"] != "local" {
		t.Fatalf("revert not live: %v", caps["binding"])
	}
}
