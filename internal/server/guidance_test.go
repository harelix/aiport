package server

import (
	"strings"
	"testing"
	"time"
)

// Guidance versions: aiport numbers the guidance it gives coding agents and
// keeps, per Application, the version its coding agent last confirmed.

func TestGuidanceOfAnApplicationThatNeverConfirmed(t *testing.T) {
	h, _ := requestsServer(t)
	w, out := call(h, "GET", GuidancePath, "", reqHelper, false)
	if w.Code != 200 {
		t.Fatalf("guidance: %d %s", w.Code, w.Body)
	}
	if out["current"] != float64(4) || out["confirmed"] != nil || out["behind"] != true {
		t.Fatalf("an Application integrated before guidance versions must be unknown and behind: %s", w.Body)
	}
	changes, _ := out["changes"].([]any)
	if len(changes) != 4 {
		t.Fatalf("unknown should list every version's changes: %s", w.Body)
	}
	newest := changes[len(changes)-1].(map[string]any)
	if newest["version"] != float64(4) || !strings.Contains(newest["summary"].(string), "X-Aiport-Module") {
		t.Fatalf("version 4 should be the votes per module guidance: %v", newest)
	}
	if line, _ := out["agents_line"].(string); !strings.Contains(line, "aiport") {
		t.Fatalf("agents_line missing: %s", w.Body)
	}
	steps, _ := out["steps"].([]any)
	if len(steps) == 0 || !strings.Contains(steps[len(steps)-1].(string), "aiport_guidance_done") {
		t.Fatalf("the steps must end with confirming: %v", steps)
	}
}

func TestConfirmingGuidanceRecordsTheCurrentVersion(t *testing.T) {
	h, _ := requestsServer(t)
	w, out := call(h, "POST", GuidancePath+"/done", "", reqHelper, false)
	if w.Code != 200 || out["confirmed"] != float64(4) || out["current"] != float64(4) || out["behind"] != false {
		t.Fatalf("done: %d %s", w.Code, w.Body)
	}
	w, out = call(h, "GET", GuidancePath, "", reqHelper, false)
	if out["confirmed"] != float64(4) || out["behind"] != false || len(out["changes"].([]any)) != 0 {
		t.Fatalf("after done nothing should be left to do: %s", w.Body)
	}
	// Own state only: another Application is still unknown.
	_, other := call(h, "GET", GuidancePath, "", reqWriter, false)
	if other["confirmed"] != nil || other["behind"] != true {
		t.Fatalf("one Application's confirmation leaked to another: %v", other)
	}
}

func TestGuidanceConfirmationSurvivesARestart(t *testing.T) {
	h, path := requestsServer(t)
	call(h, "POST", GuidancePath+"/done", "", reqHelper, false)
	h.Close()
	again := reopen(t, path, time.UnixMilli(1_800_000_060_000))
	_, out := call(again, "GET", GuidancePath, "", reqHelper, false)
	if out["confirmed"] != float64(4) {
		t.Fatalf("confirmation lost on restart: %v", out)
	}
}

func TestGuidanceNeedsAnApplicationToken(t *testing.T) {
	h, _ := requestsServer(t)
	for _, c := range []struct{ method, path string }{{"GET", GuidancePath}, {"POST", GuidancePath + "/done"}} {
		w, out := call(h, c.method, c.path, "", reqManagement, false)
		if w.Code != 403 || errorCode(out) != "application_token_required" {
			t.Fatalf("%s %s with the management token: %d %s", c.method, c.path, w.Code, w.Body)
		}
	}
	if w, _ := call(h, "POST", GuidancePath, "", reqHelper, false); w.Code != 405 {
		t.Fatalf("POST guidance: %d", w.Code)
	}
	if w, _ := call(h, "GET", GuidancePath+"/done", "", reqHelper, false); w.Code != 405 {
		t.Fatalf("GET done: %d", w.Code)
	}
}

func TestCapabilitiesShowGuidance(t *testing.T) {
	h, _ := requestsServer(t)
	_, out := call(h, "GET", CapabilitiesPath, "", reqHelper, false)
	g, _ := out["guidance"].(map[string]any)
	if g == nil || g["current"] != float64(4) || g["confirmed"] != nil {
		t.Fatalf("capabilities guidance before confirming: %v", out["guidance"])
	}
	call(h, "POST", GuidancePath+"/done", "", reqHelper, false)
	_, out = call(h, "GET", CapabilitiesPath, "", reqHelper, false)
	if g, _ := out["guidance"].(map[string]any); g["confirmed"] != float64(4) {
		t.Fatalf("capabilities guidance after confirming: %v", out["guidance"])
	}
}

func initialize(t *testing.T, h *Handler, token string) string {
	t.Helper()
	_, out := mcp(h, token, rpc(1, "initialize", map[string]any{"protocolVersion": "2025-06-18"}))
	res, _ := out["result"].(map[string]any)
	s, _ := res["instructions"].(string)
	return s
}

// texts returns the text blocks of a tool result.
func texts(res map[string]any) []string {
	out := []string{}
	for _, b := range res["content"].([]any) {
		out = append(out, b.(map[string]any)["text"].(string))
	}
	return out
}

func TestMCPRemindsOnlyWhileBehind(t *testing.T) {
	h, _ := requestsServer(t)
	first := strings.SplitN(initialize(t, h, reqHelper), "\n", 2)[0]
	if !strings.Contains(first, "aiport_guidance") || !strings.Contains(first, "version 4") {
		t.Fatalf("initialize instructions should start with the reminder: %q", first)
	}
	for _, name := range []string{"aiport_capabilities", "aiport_catalog", "aiport_my_requests"} {
		res := toolCall(t, h, reqHelper, name, nil)
		got := texts(res)
		if len(got) != 2 || got[1] != first {
			t.Fatalf("%s should carry the reminder as its last block: %v", name, got)
		}
	}
	// An error answer carries it too.
	if got := texts(toolCall(t, h, reqHelper, "aiport_my_requests", map[string]any{"id": "nope"})); len(got) != 2 {
		t.Fatalf("an error answer lacks the reminder: %v", got)
	}

	done := toolCall(t, h, reqHelper, "aiport_guidance_done", nil)
	w, _ := call(h, "GET", GuidancePath, "", reqHelper, false)
	sameJSON(t, toolCall(t, h, reqHelper, "aiport_guidance", nil), w.Body.Bytes())
	if done["isError"] == true || len(texts(done)) != 1 {
		t.Fatalf("guidance_done: %v", done)
	}
	if s := initialize(t, h, reqHelper); strings.Contains(s, first) || strings.Contains(s, "\n") {
		t.Fatalf("reminder still there after confirming: %q", s)
	}
	if got := texts(toolCall(t, h, reqHelper, "aiport_capabilities", nil)); len(got) != 1 {
		t.Fatalf("reminder still there after confirming: %v", got)
	}
	// The other Application is still reminded.
	if s := initialize(t, h, reqWriter); !strings.HasPrefix(s, first) {
		t.Fatalf("another Application lost its reminder: %q", s)
	}
}

func TestMCPGuidanceMatchesHTTP(t *testing.T) {
	h, _ := requestsServer(t)
	res := toolCall(t, h, reqWriter, "aiport_guidance", nil)
	w, _ := call(h, "GET", GuidancePath, "", reqWriter, false)
	sameJSON(t, res, w.Body.Bytes())
	if len(texts(res)) != 1 {
		t.Fatalf("aiport_guidance needs no reminder of itself: %v", texts(res))
	}
}

func TestOwnerSeesWhoIsBehind(t *testing.T) {
	h, _ := requestsServer(t)
	call(h, "POST", GuidancePath+"/done", "", reqHelper, false)
	w, out := call(h, "GET", "/api/v1/guidance", "", reqManagement, false)
	if w.Code != 200 || out["current"] != float64(4) {
		t.Fatalf("management guidance: %d %s", w.Code, w.Body)
	}
	if s, _ := out["instruction"].(string); !strings.Contains(s, "aiport_guidance") || !strings.Contains(s, "aiport_guidance_done") || !strings.Contains(s, GuidancePath) {
		t.Fatalf("refresh instruction: %q", s)
	}
	apps, _ := out["applications"].(map[string]any)
	helper, _ := apps["helper"].(map[string]any)
	writer, _ := apps["writer"].(map[string]any)
	if helper["confirmed"] != float64(4) || helper["confirmed_at_ms"] != float64(1_800_000_000_000) || writer == nil || writer["confirmed"] != nil {
		t.Fatalf("applications: %v", apps)
	}
	if out["behind"] != float64(2) { // nightly and writer have not confirmed
		t.Fatalf("behind count: %v", out["behind"])
	}
	if out["never_confirmed"] != float64(2) { // neither ever confirmed any version
		t.Fatalf("never confirmed count: %v", out["never_confirmed"])
	}
	// An Application token never reaches it.
	if w, _ := call(h, "GET", "/api/v1/guidance", "", reqHelper, false); w.Code != 401 && w.Code != 403 {
		t.Fatalf("Application token read the Owner's view: %d", w.Code)
	}
}
