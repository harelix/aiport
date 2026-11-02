package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// tryInference sends one inference request with the given bearer token ("" =
// none) at the current fixed time and returns the status.
func tryInference(h *Handler, token string) int {
	w, _ := call(h, "POST", "/v1/chat/completions", `{"model":"fast","messages":[]}`, token, false)
	return w.Code
}

func unknownTokenSuggestions(t *testing.T, h *Handler, management string) []map[string]any {
	t.Helper()
	_, list := pendingDecisions(t, h, management)
	var out []map[string]any
	for _, s := range list {
		if s["kind"] == "pair_application" {
			out = append(out, s)
		}
	}
	return out
}

// Repeated failed authentications become one informational Suggestion that
// points to the Applications panel; the token is stored nowhere.
func TestUnknownTokenSuggestion(t *testing.T) {
	at := suggestionsNow
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management := suggestionsServer(t, path, "", nil, &at)
	unknown := "sk-unpaired-" + strings.Repeat("u", 30)

	at = suggestionsNow.Add(-5 * time.Hour) // outside the window
	for i := 0; i < 20; i++ {
		tryInference(h, unknown)
	}
	for i := 0; i < 11; i++ {
		at = suggestionsNow.Add(-40*time.Minute + time.Duration(i)*time.Minute)
		token := unknown
		if i == 10 {
			token = "" // no token at all is also a failed authentication
		}
		if got := tryInference(h, token); got != 401 {
			t.Fatalf("unknown token: status %d", got)
		}
	}
	// Neither successful requests nor management refusals are counted.
	for i := 0; i < 5; i++ {
		if w, _ := call(h, "GET", "/v1/aiport/capabilities", "", strings.Repeat("n", 40), false); w.Code != 200 {
			t.Fatalf("capabilities: %d", w.Code)
		}
		call(h, "GET", "/api/v1/decisions", "", strings.Repeat("n", 40), false)
	}
	at = suggestionsNow
	h.Detect(at)
	h.Detect(at)
	list := unknownTokenSuggestions(t, h, management)
	if len(list) != 1 {
		t.Fatalf("want one pairing Suggestion, got %v", list)
	}
	s := list[0]
	ev, _ := s["evidence"].(map[string]any)
	if s["source"] != "suggestion" || s["status"] != "pending" || s["opens"] != "applications" || s["informational"] != true {
		t.Fatalf("suggestion: %v", s)
	}
	if ev["count"] != float64(11) || ev["window_minutes"] != float64(60) || ev["from_ms"] != float64(suggestionsNow.Add(-40*time.Minute).UnixMilli()) || ev["to_ms"] != float64(suggestionsNow.Add(-30*time.Minute).UnixMilli()) {
		t.Fatalf("evidence: %v", ev)
	}
	need := s["need"].(string)
	if !strings.Contains(need, "11") || !strings.Contains(need, "pair") {
		t.Fatalf("need does not explain the evidence: %q", need)
	}

	// Informational: nothing to approve, only Dismiss.
	if code, out := decideOn(t, h, management, s["id"].(string), `{"action":"accept"}`); code != 400 {
		t.Fatalf("accepting an informational Suggestion: %d %v", code, out)
	}
	if code, out := decideOn(t, h, management, s["id"].(string), `{"action":"dismiss"}`); code != 200 {
		t.Fatalf("dismiss: %d %v", code, out)
	}
	h.Detect(at)
	if list := unknownTokenSuggestions(t, h, management); len(list) != 0 {
		t.Fatalf("dismissed Suggestion came back: %v", list)
	}

	// The token, and its digest, are kept nowhere aiport writes.
	sum := sha256.Sum256([]byte(unknown))
	h.Close()
	files, _ := filepath.Glob(path + "*")
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, secret := range [][]byte{[]byte(unknown), []byte(hex.EncodeToString(sum[:])), sum[:]} {
			if bytes.Contains(b, secret) {
				t.Fatalf("%s contains the unknown token or its digest", f)
			}
		}
	}
}

// A few failures, or failures only before the window, suggest nothing.
func TestUnknownTokenBelowThreshold(t *testing.T) {
	at := suggestionsNow.Add(-10 * time.Minute)
	h, management := suggestionsServer(t, filepath.Join(t.TempDir(), "aiport.yaml"), "adapting: {unknown_token: {count: 4, window_minutes: 15}}\n", nil, &at)
	for i := 0; i < 3; i++ {
		tryInference(h, "nope")
	}
	at = suggestionsNow.Add(-20 * time.Minute)
	for i := 0; i < 10; i++ {
		tryInference(h, "nope")
	}
	at = suggestionsNow
	h.Detect(at)
	if list := unknownTokenSuggestions(t, h, management); len(list) != 0 {
		t.Fatalf("want no pairing Suggestion, got %v", list)
	}
	at = suggestionsNow.Add(-time.Minute)
	tryInference(h, "nope")
	at = suggestionsNow
	h.Detect(at)
	if list := unknownTokenSuggestions(t, h, management); len(list) != 1 || list[0]["evidence"].(map[string]any)["count"] != float64(4) {
		t.Fatalf("want one pairing Suggestion from 4 failures, got %v", list)
	}
}
