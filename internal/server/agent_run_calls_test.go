package server

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Where an Agent run's model calls go. Claude Code is pointed at aiport by
// environment variables, with a credential of its own that authenticates
// only model calls, as the built-in Application aiport, to the Connection
// the Owner chose, and only while the run lasts. Codex and Gemini CLI use
// their own login, and the run says so. aiport never edits an agent's
// configuration files: it only sets the run's environment.

// agentEnv is the fake agent's environment, by name.
func agentEnv(t *testing.T, dir string) map[string]string {
	t.Helper()
	var start struct {
		Env []string `json:"env"`
	}
	rec := agentRecord(t, dir)
	if len(rec["start"]) == 0 {
		t.Fatal("the agent did not start")
	}
	_ = json.Unmarshal(rec["start"][0], &start)
	env := map[string]string{}
	for _, kv := range start.Env {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
	}
	return env
}

func TestClaudeCodeRunsCallModelsThroughAiport(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "Looking. "}, {Wait: "go"}, {Say: "Found it."}}},
		"agent_runs: {connection: claude-writer, model: claude-model}\n")
	events := f.explainStream(t)
	run := m(until(t, events, "run")["run"])
	calls := m(run["model_calls"])
	if calls["through_aiport"] != true || calls["connection"] != "claude-writer" || calls["model"] != "claude-model" {
		t.Fatalf("model calls: %v", run)
	}
	until(t, events, "text")
	env := agentEnv(t, f.agentDir)
	token := env["ANTHROPIC_AUTH_TOKEN"]
	if env["ANTHROPIC_BASE_URL"] != "http://127.0.0.1:8787" || len(token) < 32 || env["ANTHROPIC_MODEL"] != "claude-model" {
		t.Fatalf("the agent's environment: %v", env)
	}

	// The run's credential makes model calls as the built-in Application.
	w, _ := call(f.h, "POST", "/v1/messages", `{"model":"claude-model","max_tokens":5,"messages":[{"role":"user","content":"hi"}]}`, token, false)
	if w.Code != 200 {
		t.Fatalf("the run's model call: %d %s", w.Code, w.Body)
	}
	// ... and nothing else.
	for _, path := range []string{"/api/v1/issues", "/mcp", "/v1/aiport/issues", "/v1/aiport/capabilities"} {
		if w, _ := call(f.h, "GET", path, "", token, false); w.Code != 403 && w.Code != 401 {
			t.Fatalf("the run's credential reached %s: %d", path, w.Code)
		}
	}
	from, to := time.Now().Add(-time.Hour).UnixMilli(), time.Now().Add(time.Hour).UnixMilli()
	waitFor(t, func() bool {
		_, act := call(f.h, "GET", "/api/v1/requests?origin=aiport&from_ms="+strconv.FormatInt(from, 10)+"&to_ms="+strconv.FormatInt(to, 10), "", issuesManagement, false)
		rows, _ := act["requests"].([]any)
		return len(rows) == 1 && m(rows[0])["target_id"] == "claude-writer" && m(rows[0])["route"] == "/v1/messages"
	})

	release(t, f.agentDir, "go")
	if done := until(t, events, "done"); m(m(done["issue"])["explanation"])["text"] != "Looking. Found it." {
		t.Fatalf("done: %v", done)
	}
	// Revoked once the run ended.
	if w, _ := call(f.h, "POST", "/v1/messages", `{"model":"claude-model","max_tokens":5,"messages":[]}`, token, false); w.Code != 401 {
		t.Fatalf("the run's credential after the run: %d", w.Code)
	}
	_, out := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	if raw, _ := json.Marshal(out); strings.Contains(string(raw), token) {
		t.Fatal("the run's credential is in its log")
	}
}

// Without a Connection chosen for Agent runs (agent_runs.connection),
// Claude Code uses its own login and the run says so, even when there is an
// Anthropic protocol default: aiport never falls back to it silently.
func TestClaudeCodeWithoutAConnectionUsesItsOwnLogin(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "ok"}}}, "")
	if w, _ := call(f.h, "PUT", "/api/v1/defaults", `{"openai_target":"hosted","anthropic_target":"claude-writer"}`, issuesManagement, true); w.Code != 200 {
		t.Fatalf("defaults: %d %s", w.Code, w.Body)
	}
	code, evs := f.explainEvents(t, f.id)
	if code != 200 || evs[len(evs)-1]["type"] != "done" {
		t.Fatalf("explain: %d %v", code, evs)
	}
	calls := m(m(evs[1]["run"])["model_calls"])
	if evs[1]["type"] != "run" || calls["through_aiport"] != false || !strings.Contains(calls["note"].(string), "own login") {
		t.Fatalf("run: %v", evs[1])
	}
	if env := agentEnv(t, f.agentDir); env["ANTHROPIC_BASE_URL"] != "" || env["ANTHROPIC_AUTH_TOKEN"] != "" {
		t.Fatalf("pointed at aiport without a Connection: %v", env)
	}
}

func TestCodexAndGeminiRunsUseTheirOwnLogin(t *testing.T) {
	for _, agent := range []string{"codex", "gemini"} {
		t.Run(agent, func(t *testing.T) {
			f := newAgentRunFixture(t, agent, acpScript{Steps: []acpStep{{Say: "The model name is wrong."}}},
				"agent_runs: {connection: claude-writer}\n")
			_, out := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
			by := m(m(out["explain_by"])["docs-bot"])
			if by["agent"] != agent || m(by["model_calls"])["through_aiport"] != false || !strings.Contains(m(by["model_calls"])["note"].(string), "own login") {
				t.Fatalf("explain_by: %v", by)
			}
			code, evs := f.explainEvents(t, f.id)
			last := evs[len(evs)-1]
			if code != 200 || last["type"] != "done" || m(m(last["issue"])["explanation"])["agent"] != agent {
				t.Fatalf("explain: %d %v", code, evs)
			}
			env := agentEnv(t, f.agentDir)
			for k := range env {
				if strings.HasPrefix(k, "ANTHROPIC_") || strings.HasPrefix(k, "OPENAI_") || strings.HasPrefix(k, "GOOGLE_") || strings.HasPrefix(k, "GEMINI_") {
					t.Fatalf("%s run got %s", agent, k)
				}
			}
			var start struct {
				Args []string `json:"args"`
			}
			_ = json.Unmarshal(agentRecord(t, f.agentDir)["start"][0], &start)
			if want := map[string]string{"codex": "", "gemini": "--acp"}[agent]; strings.Join(start.Args, " ") != want {
				t.Fatalf("%s started with %v", agent, start.Args)
			}
		})
	}
}

// The run's log is kept while the Issue is open, and dropped when it
// closes; the explanation stays with the Issue.
func TestTheRunLogIsDroppedWhenTheIssueCloses(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Tool: "Read app.py"}, {Say: "app.py asks for llama3:8b."}}}, "")
	if w, _ := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false); w.Code != 404 {
		t.Fatalf("a run log before any run: %d", w.Code)
	}
	if code, evs := f.explainEvents(t, f.id); code != 200 || evs[len(evs)-1]["type"] != "done" {
		t.Fatalf("explain: %d %v", code, evs)
	}
	w, out := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	run, log := m(out["run"]), out["log"].([]any)
	if w.Code != 200 || run["state"] != "done" || run["agent_name"] != "Claude Code" || len(log) < 3 || m(log[0])["text"] != "Read app.py" {
		t.Fatalf("the run log: %d %v", w.Code, out)
	}
	if code, out := f.decide(t, f.id, `{"action":"dismiss"}`); code != 200 {
		t.Fatalf("dismiss: %d %v", code, out)
	}
	if w, _ := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false); w.Code != 404 {
		t.Fatalf("the run log after the Issue closed: %d", w.Code)
	}
	if is := f.only(t); m(is["explanation"])["text"] != "app.py asks for llama3:8b." {
		t.Fatalf("the explanation: %v", is["explanation"])
	}
}

// Only an installed agent with its ACP entry point and a repository that
// passes its check gets an Agent run; otherwise the chosen Connection
// explains. With adapting off, no run starts.
func TestNoAgentRunWithoutAReadyAgentOrWithAdaptingOff(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "unused"}}}, "adapting: {explain: {connection: writer, model: writer-model}}\n")
	f.h.agentDirs = []string{t.TempDir()} // the adapter is gone
	_, out := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if len(m(out["explain_by"])) != 0 || m(out["explain"])["source"] != "connection" {
		t.Fatalf("without the adapter: %v %v", out["explain_by"], out["explain"])
	}
	if code, evs := f.explainEvents(t, f.id); code != 200 || m(evs[0]["by"])["source"] != "connection" {
		t.Fatalf("explain without the adapter: %d %v", code, evs)
	}
	if rec := agentRecord(t, f.agentDir); len(rec["start"]) != 0 {
		t.Fatal("an agent started without its adapter found")
	}

	g := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "unused"}}}, "")
	data, _ := os.ReadFile(g.path)
	if err := os.WriteFile(g.path, append(data, []byte("adapting: {disabled: true}\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	g.h.Close()
	g.start(t)
	g.useAgentDirs()
	_, out = call(g.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if len(m(out["explain_by"])) != 0 {
		t.Fatalf("explain_by with adapting off: %v", out["explain_by"])
	}
	if w, out := call(g.h, "POST", "/api/v1/issues/"+g.id+"/explain", "", issuesManagement, true); w.Code != 409 || m(out["error"])["code"] != "adapting_off" {
		t.Fatalf("explain with adapting off: %d %v", w.Code, out)
	}
	if rec := agentRecord(t, g.agentDir); len(rec["start"]) != 0 {
		t.Fatal("an agent started with adapting off")
	}
}

// The Owner chooses the Connection (and, optionally, the model) of the
// runs' model calls in Settings; it must speak the Anthropic protocol.
func TestTheOwnerChoosesTheConnectionOfAgentRuns(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "unused"}}}, "")
	if w, _ := call(f.h, "PUT", "/api/v1/agents/calls", `{"connection":"claude-writer"}`, issuesManagement, false); w.Code != 403 {
		t.Fatalf("without CSRF: %d", w.Code)
	}
	if w, out := call(f.h, "PUT", "/api/v1/agents/calls", `{"connection":"writer"}`, issuesManagement, true); w.Code != 400 || !strings.Contains(m(out["error"])["message"].(string), "Anthropic") {
		t.Fatalf("an OpenAI-protocol Connection: %d %v", w.Code, out)
	}
	w, out := call(f.h, "PUT", "/api/v1/agents/calls", `{"connection":"claude-writer","model":"claude-model"}`, issuesManagement, true)
	if w.Code != 200 || out["change_id"] == nil || m(out["agent_runs"])["connection"] != "claude-writer" {
		t.Fatalf("choose: %d %v", w.Code, out)
	}
	_, out = call(f.h, "GET", "/api/v1/agents", "", issuesManagement, false)
	if m(out["agent_runs"])["connection"] != "claude-writer" || m(out["agent_runs"])["model"] != "claude-model" || m(out["agent_runs"])["explain_minutes"] != float64(10) {
		t.Fatalf("agents: %v", out)
	}
	_, out = call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if calls := m(m(m(out["explain_by"])["docs-bot"])["model_calls"]); calls["through_aiport"] != true || calls["connection"] != "claude-writer" {
		t.Fatalf("explain_by: %v", out["explain_by"])
	}
	if w, _ := call(f.h, "PUT", "/api/v1/agents/calls", `{"connection":""}`, issuesManagement, true); w.Code != 200 {
		t.Fatalf("clear: %d", w.Code)
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
