package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The rules of an Agent run: what Explain may do, Stop, the time limit, one
// run per Application, and where the run's model calls go.

// Explain is read-only: aiport denies every edit and command the agent asks
// for, and any read outside the run's worktree (also the Owner's own
// copy), without asking the Owner; reading inside the worktree is allowed. aiport only ever answers once,
// so nothing is remembered in the agent's own settings.
func TestExplainDeniesEveryEditAndCommand(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{
		{Ask: &acpAsk{Kind: "read", Title: "Read config.py", Path: "CWD/README.md"}},
		{Ask: &acpAsk{Kind: "read", Title: "Read the Owner's copy", Path: "REPO/README.md"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Edit config.py", Path: "REPO/config.py"}},
		{Ask: &acpAsk{Kind: "execute", Title: "Run pytest"}},
		{Ask: &acpAsk{Kind: "delete", Title: "Delete config.py", Path: "REPO/config.py"}},
		{Ask: &acpAsk{Kind: "read", Title: "Read /etc/hosts", Path: "/etc/hosts"}},
		{Ask: &acpAsk{Kind: "read", Title: "Read outside via ..", Path: "REPO/../x"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Edit with no reject_once", Options: []string{"allow_once", "allow_always", "reject_always"}}},
		{Say: "Done reading."},
	}}, "")
	// REPO in the script's paths is the Owner's copy of the repository.
	writeScript(t, f.agentDir, f.repo)

	code, evs := f.explainEvents(t, f.id)
	if code != 200 || evs[len(evs)-1]["type"] != "done" {
		t.Fatalf("explain: %d %v", code, evs)
	}
	var shown []string
	for _, e := range evs {
		if e["type"] == "step" && m(e["step"])["kind"] == "permission" {
			s := m(e["step"])
			shown = append(shown, s["text"].(string)+"="+s["verdict"].(string))
		}
	}
	want := "Read config.py=allowed|Read the Owner's copy=denied|Edit config.py=denied|Run pytest=denied|Delete config.py=denied|Read /etc/hosts=denied|Read outside via ..=denied|Edit with no reject_once=denied"
	if strings.Join(shown, "|") != want {
		t.Fatalf("permissions shown:\n%s\nwant\n%s", strings.Join(shown, "|"), want)
	}
	var got []string
	for _, raw := range agentRecord(t, f.agentDir)["permission"] {
		var p struct{ Title, Outcome, Option string }
		_ = json.Unmarshal(raw, &p)
		got = append(got, p.Title+"="+p.Outcome+":"+p.Option)
	}
	want = "Read config.py=selected:opt-allow_once|Read the Owner's copy=selected:opt-reject_once|Edit config.py=selected:opt-reject_once|Run pytest=selected:opt-reject_once|Delete config.py=selected:opt-reject_once|Read /etc/hosts=selected:opt-reject_once|Read outside via ..=selected:opt-reject_once|Edit with no reject_once=cancelled:"
	if strings.Join(got, "|") != want {
		t.Fatalf("answers the agent got:\n%s\nwant\n%s", strings.Join(got, "|"), want)
	}
}

// explainStream starts Explain for the fixture's Issue over a real
// connection and returns its events as they arrive; the channel closes when
// the stream ends.
func (f *agentRunFixture) explainStream(t *testing.T) <-chan map[string]any {
	t.Helper()
	srv := httptest.NewServer(f.h)
	t.Cleanup(srv.Close)
	req, _ := http.NewRequest("POST", srv.URL+"/api/v1/issues/"+f.id+"/explain", nil)
	req.Host = "127.0.0.1:8787"
	req.Header.Set("Authorization", "Bearer "+issuesManagement)
	req.Header.Set("X-Aiport-CSRF", "1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("explain: %d", resp.StatusCode)
	}
	out := make(chan map[string]any, 64)
	go func() {
		defer resp.Body.Close()
		defer close(out)
		lines := bufio.NewScanner(resp.Body)
		for lines.Scan() {
			e := map[string]any{}
			_ = json.Unmarshal(lines.Bytes(), &e)
			out <- e
		}
	}()
	return out
}

// until reads events until one of type kind arrives, or the stream ends.
func until(t *testing.T, events <-chan map[string]any, kind string) map[string]any {
	t.Helper()
	timeout := time.After(20 * time.Second)
	for {
		select {
		case e, ok := <-events:
			if !ok {
				t.Fatalf("the stream ended before a %s event", kind)
			}
			if e["type"] == kind {
				return e
			}
		case <-timeout:
			t.Fatalf("no %s event", kind)
		}
	}
}

// Stop ends a run at any time: the agent is asked to cancel its turn,
// nothing is saved, and the Application may run again.
func TestStopEndsAnAgentRun(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "Reading "}, {Cancel: true}}}, "")
	events := f.explainStream(t)
	run := m(until(t, events, "run")["run"])
	if run["agent"] != "claude" || run["mode"] != "explain" || run["state"] != "running" || run["issue"] != f.id {
		t.Fatalf("run: %v", run)
	}
	until(t, events, "text")
	_, out := call(f.h, "GET", "/api/v1/runs", "", issuesManagement, false)
	if runs, _ := out["runs"].([]any); len(runs) != 1 || m(runs[0])["id"] != run["id"] {
		t.Fatalf("runs going: %v", out)
	}
	// Only one run per Application.
	if w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/explain", "", issuesManagement, true); w.Code != 409 {
		t.Fatalf("a second run: %d %v", w.Code, out)
	}
	if w, _ := call(f.h, "POST", "/api/v1/runs/"+run["id"].(string)+"/stop", "", issuesManagement, false); w.Code != 403 {
		t.Fatalf("stop without CSRF: %d", w.Code)
	}
	if w, _ := call(f.h, "POST", "/api/v1/runs/nope/stop", "", issuesManagement, true); w.Code != 404 {
		t.Fatalf("stop an unknown run: %d", w.Code)
	}
	if w, out := call(f.h, "POST", "/api/v1/runs/"+run["id"].(string)+"/stop", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("stop: %d %v", w.Code, out)
	}
	e := until(t, events, "error")
	if m(e["error"])["code"] != "run_stopped" {
		t.Fatalf("stopped: %v", e)
	}
	if is := f.only(t); is["explanation"] != nil {
		t.Fatalf("a stopped run saved %v", is["explanation"])
	}
	if rec := agentRecord(t, f.agentDir); len(rec["cancel"]) != 1 {
		t.Fatalf("the agent was not asked to cancel: %v", rec)
	}
	_, out = call(f.h, "GET", "/api/v1/runs", "", issuesManagement, false)
	if runs, _ := out["runs"].([]any); len(runs) != 0 {
		t.Fatalf("runs going after Stop: %v", out)
	}
	_, out = call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	if m(out["run"])["state"] != "stopped" {
		t.Fatalf("the stopped run: %v", out)
	}
}

// A run that reaches its time limit is stopped, even when the agent ignores
// the request to cancel: aiport kills its process group.
func TestAnAgentRunStopsAtItsTimeLimit(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "Thinking hard"}, {Ignore: true}}}, "")
	f.h.explainLimit = 300 * time.Millisecond
	started := time.Now()
	code, evs := f.explainEvents(t, f.id)
	last := evs[len(evs)-1]
	if code != 200 || last["type"] != "error" || m(last["error"])["code"] != "explanation_timed_out" || !strings.Contains(m(last["error"])["message"].(string), "300ms") {
		t.Fatalf("timed out: %d %v", code, evs)
	}
	if took := time.Since(started); took > 10*time.Second {
		t.Fatalf("the run took %v to stop", took)
	}
	if is := f.only(t); is["explanation"] != nil {
		t.Fatal("a timed-out run saved an explanation")
	}
	_, out := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	if m(out["run"])["state"] != "timed_out" {
		t.Fatalf("the run: %v", out)
	}
}

// An agent that fails is reported in aiport's words; nothing is saved.
func TestAnAgentThatFailsSavesNothing(t *testing.T) {
	one := 1
	for name, script := range map[string]acpScript{
		"exits":            {Steps: []acpStep{{Say: "partial"}, {Exit: &one}}},
		"refuses":          {Steps: []acpStep{{Say: "no"}}, Stop: "refusal"},
		"speaks version 2": {Protocol: 2},
		"says nothing":     {},
	} {
		t.Run(name, func(t *testing.T) {
			f := newAgentRunFixture(t, "claude", script, "")
			code, evs := f.explainEvents(t, f.id)
			last := evs[len(evs)-1]
			if code != 200 || last["type"] != "error" || m(last["error"])["code"] != "explanation_failed" {
				t.Fatalf("%d %v", code, evs)
			}
			if is := f.only(t); is["explanation"] != nil {
				t.Fatal("saved")
			}
		})
	}
}

// writeScript rewrites the fake agent's script with REPO replaced by repo.
func writeScript(t *testing.T, dir, repo string) {
	t.Helper()
	path := filepath.Join(dir, "script.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(data), "REPO", repo)), 0o600); err != nil {
		t.Fatal(err)
	}
}

// setMode is the mode the fake agent was asked to switch to, "" when none.
func setMode(t *testing.T, dir string) string {
	t.Helper()
	rec := agentRecord(t, dir)["set_mode"]
	if len(rec) == 0 {
		return ""
	}
	var p struct {
		SessionID string `json:"sessionId"`
		ModeID    string `json:"modeId"`
	}
	_ = json.Unmarshal(rec[len(rec)-1], &p)
	return p.SessionID + ":" + p.ModeID
}

// The session mode an agent starts in comes from its own settings. When it
// offers modes, aiport picks the most restrictive suitable one: plan or
// read-only for Explain, one that asks before acting for Fix it; never
// bypass or auto-accept.
func TestARunPicksTheMostRestrictiveSessionMode(t *testing.T) {
	modes := &acpModes{Current: "bypassPermissions", Available: []string{"default", "acceptEdits", "plan", "bypassPermissions"}}
	f := newAgentRunFixture(t, "claude", acpScript{Modes: modes, Steps: []acpStep{{Say: "Read it."}}}, "")
	code, evs := f.explainEvents(t, f.id)
	if code != 200 || evs[len(evs)-1]["type"] != "done" {
		t.Fatalf("explain: %d %v", code, evs)
	}
	if got := setMode(t, f.agentDir); got != "sess-1:plan" {
		t.Fatalf("Explain switched to %q", got)
	}
	_, out := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	if m(out["run"])["session_mode"] != "plan" {
		t.Fatalf("the run: %v", out["run"])
	}

	fix, _ := newFixFixture(t, acpScript{Modes: modes, Steps: editConfig.Steps})
	if code, out := fix.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	fix.fixEnded(t)
	if got := setMode(t, fix.agentDir); got != "sess-1:default" {
		t.Fatalf("Fix it switched to %q", got)
	}

	// Already in the right mode: nothing to switch.
	plan := newAgentRunFixture(t, "gemini", acpScript{Modes: &acpModes{Current: "plan", Available: []string{"default", "autoEdit", "yolo", "plan"}}, Steps: []acpStep{{Say: "ok"}}}, "")
	plan.explainEvents(t, plan.id)
	if got := setMode(t, plan.agentDir); got != "" {
		t.Fatalf("switched from plan to %q", got)
	}
}

// A session that starts in a mode acting without asking, with no other mode
// offered, never gets the prompt.
func TestARunRefusesASessionThatNeverAsks(t *testing.T) {
	f := newAgentRunFixture(t, "codex", acpScript{Modes: &acpModes{Current: "full-access", Available: []string{"full-access", "auto"}}, Steps: []acpStep{{Say: "should not run"}}}, "")
	code, evs := f.explainEvents(t, f.id)
	last := evs[len(evs)-1]
	if code != 200 || last["type"] != "error" || !strings.Contains(m(last["error"])["message"].(string), "without asking") {
		t.Fatalf("explain: %d %v", code, evs)
	}
	if rec := agentRecord(t, f.agentDir); len(rec["prompt"]) != 0 {
		t.Fatal("the agent was prompted")
	}
}

// A permission request naming a session aiport did not open is denied.
func TestAPermissionForAnotherSessionIsDenied(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{
		{Ask: &acpAsk{Kind: "read", Title: "Read README.md", Path: "CWD/README.md", Session: "sess-other"}},
		{Say: "ok"},
	}}, "")
	f.explainEvents(t, f.id)
	if got := answers(t, f.agentDir); got != "Read README.md=reject_once" {
		t.Fatalf("answers: %s", got)
	}
}
