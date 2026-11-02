package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Explain by a Coding agent: an Application with a connected Coding agent
// and repository gets an Agent run. aiport starts the installed agent's ACP
// entry point by absolute path, with a minimal environment, in the
// repository, sends the Issue's evidence as the prompt, streams the agent's
// words and saves them with the Issue.

// agentRunFixture is the docs-bot example with an Issue, a temporary git
// repository and Coding agent agent connected to docs-bot, the fake ACP
// agent following script.
type agentRunFixture struct {
	*issuesFixture
	repo, agentDir, bin string
	id                  string // the Issue
}

func newAgentRunFixture(t *testing.T, agent string, script acpScript, extra string) *agentRunFixture {
	t.Helper()
	writer := newFakeWriter(t, "unused")
	repo := gitRepository(t)
	// One commit: an Agent run works in a worktree of the Owner's HEAD.
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("docs-bot\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-q", "-m", "start")
	bin := t.TempDir()
	dir := installFakeAgent(t, bin, agent, script)
	f := &agentRunFixture{repo: repo, agentDir: dir, bin: bin}
	f.issuesFixture = newExplainFixture(t, writer, extra)
	f.useAgentDirs()
	if w, _ := call(f.h, "PATCH", "/api/v1/origins/docs-bot", `{"agent":"`+agent+`","repository":"`+repo+`"}`, issuesManagement, true); w.Code != 200 {
		t.Fatalf("connect %s: %d %s", agent, w.Code, w.Body)
	}
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	f.id = f.only(t)["id"].(string)
	return f
}

// useAgentDirs makes the handler (also after a restart) find agents in bin
// only.
func (f *agentRunFixture) useAgentDirs() { f.h.agentDirs = []string{f.bin} }

func TestExplainByACodingAgentStreamsAndSavesTheRootCause(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{
		{Think: "Looking for where the model is set."},
		{Tool: "Read config.py"},
		{Say: "config.py asks for llama3:8b, "},
		{Say: "which the hosted Connection does not serve."},
	}}, "")

	_, out := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	by := m(m(out["explain_by"])["docs-bot"])
	if by["available"] != true || by["source"] != "coding_agent" || by["agent"] != "claude" {
		t.Fatalf("Explain for docs-bot: %v", out["explain_by"])
	}

	code, evs := f.explainEvents(t, f.id)
	if code != 200 || len(evs) < 3 {
		t.Fatalf("explain: %d %v", code, evs)
	}
	if s := evs[0]; s["type"] != "start" || m(s["by"])["source"] != "coding_agent" || m(s["by"])["agent"] != "claude" {
		t.Fatalf("start: %v", s)
	}
	text, steps := "", []string{}
	for _, e := range evs {
		switch e["type"] {
		case "text":
			text += e["text"].(string)
		case "step":
			steps = append(steps, m(e["step"])["kind"].(string)+": "+m(e["step"])["text"].(string))
		}
	}
	const want = "config.py asks for llama3:8b, which the hosted Connection does not serve."
	last := evs[len(evs)-1]
	if text != want || last["type"] != "done" {
		t.Fatalf("streamed %q, last %v", text, last)
	}
	if strings.Join(steps, "|") != "thought: Looking for where the model is set.|tool: Read config.py" {
		t.Fatalf("steps: %v", steps)
	}
	ex := m(m(last["issue"])["explanation"])
	if ex["text"] != want || ex["source"] != "coding_agent" || ex["agent"] != "claude" {
		t.Fatalf("saved: %v", ex)
	}
	if is := f.only(t); !strings.Contains(is["application_fix"].(string), want) {
		t.Fatalf("the Application fix does not carry it: %v", is["application_fix"])
	}

	// What the agent was given.
	rec := agentRecord(t, f.agentDir)
	var start struct {
		Args []string `json:"args"`
		Env  []string `json:"env"`
		Cwd  string   `json:"cwd"`
	}
	_ = json.Unmarshal(rec["start"][0], &start)
	// It read a throwaway worktree of the Owner's HEAD, never the Owner's
	// working copy, removed once the run ended; no branch was made.
	root, _ := filepath.EvalSymlinks(filepath.Dir(f.h.worktreeRoot()))
	root = filepath.Join(root, filepath.Base(f.h.worktreeRoot()))
	if filepath.Dir(start.Cwd) != root {
		t.Fatalf("the agent ran in %q, not a worktree under %q (repository %q)", start.Cwd, root, f.repo)
	}
	if _, err := os.Stat(start.Cwd); err == nil {
		t.Fatalf("the Explain worktree %s is still there", start.Cwd)
	}
	if len(worktrees(t, f.repo)) != 1 || strings.Contains(gitIn(t, f.repo, "branch", "--list"), "aiport/") {
		t.Fatalf("left behind: %v %s", worktrees(t, f.repo), gitIn(t, f.repo, "branch", "--list"))
	}
	var session struct {
		Cwd        string `json:"cwd"`
		MCPServers []any  `json:"mcpServers"`
	}
	_ = json.Unmarshal(rec["session"][0], &session)
	if session.Cwd != start.Cwd || session.MCPServers == nil || len(session.MCPServers) != 0 {
		t.Fatalf("session/new: %s", rec["session"][0])
	}
	var initialize struct {
		ProtocolVersion int `json:"protocolVersion"`
	}
	_ = json.Unmarshal(rec["initialize"][0], &initialize)
	if initialize.ProtocolVersion != 1 {
		t.Fatalf("initialize: %s", rec["initialize"][0])
	}
	allowed := map[string]bool{"PATH": true, "HOME": true, "TMPDIR": true, "LANG": true, "LC_ALL": true, "USER": true, "LOGNAME": true, "ANTHROPIC_BASE_URL": true, "ANTHROPIC_AUTH_TOKEN": true, "ANTHROPIC_MODEL": true,
		"PWD": true, "SHLVL": true, "_": true} // the last three the fake agent's /bin/sh wrapper sets itself
	all := strings.Join(start.Env, "\n")
	for _, kv := range start.Env {
		if name, _, _ := strings.Cut(kv, "="); !allowed[name] {
			t.Fatalf("the agent's environment holds %s", name)
		}
	}
	for _, secret := range []string{issuesManagement, docsBotToken, otherAppToken, hostedKey, writerKey} {
		if strings.Contains(all, secret) {
			t.Fatalf("the agent's environment holds a token or key: %v", start.Env)
		}
	}
	var prompt string
	_ = json.Unmarshal(rec["prompt"][0], &prompt)
	for _, s := range []string{"llama3:8b", "404", "hosted", "read-only"} {
		if !strings.Contains(prompt, s) {
			t.Fatalf("the prompt lacks %q: %s", s, prompt)
		}
	}
	for _, s := range []string{"PRIVATE PROMPT TEXT", issuesManagement, docsBotToken, hostedKey} {
		if strings.Contains(prompt, s) {
			t.Fatalf("the prompt holds %q", s)
		}
	}
}

// Explain reads a throwaway worktree: even an edit the agent's own settings
// allow without asking (aiport never sees it) cannot touch the Owner's
// working copy.
func TestAnEditExplainNeverSeesStaysOutOfTheOwnersCopy(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{
		{Edit: &acpEdit{Path: "CWD/README.md", Text: "changed by the agent\n"}},
		{Say: "The model name is wrong."},
	}}, "")
	before, _ := os.ReadFile(filepath.Join(f.repo, "README.md"))
	code, evs := f.explainEvents(t, f.id)
	if code != 200 || evs[len(evs)-1]["type"] != "done" {
		t.Fatalf("explain: %d %v", code, evs)
	}
	after, _ := os.ReadFile(filepath.Join(f.repo, "README.md"))
	if string(after) != string(before) || gitIn(t, f.repo, "status", "--porcelain") != "" {
		t.Fatalf("the Owner's working copy changed: %q", after)
	}
	if len(worktrees(t, f.repo)) != 1 {
		t.Fatalf("the Explain worktree is left: %v", worktrees(t, f.repo))
	}
}

// An Explain run does not depend on the dashboard's request: closing the
// tab never stops it; it ends on its own, saves its explanation, and its
// log stays readable at GET /api/v1/issues/{id}/run.
func TestClosingTheTabDoesNotStopExplain(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "Looking. "}, {Wait: "go"}, {Say: "Found it."}}}, "")
	srv := httptest.NewServer(f.h)
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "POST", srv.URL+"/api/v1/issues/"+f.id+"/explain", nil)
	req.Host = "127.0.0.1:8787"
	req.Header.Set("Authorization", "Bearer "+issuesManagement)
	req.Header.Set("X-Aiport-CSRF", "1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("explain: %v %v", err, resp)
	}
	lines := bufio.NewScanner(resp.Body)
	for lines.Scan() && !strings.Contains(lines.Text(), `"type":"text"`) {
	}
	cancel() // the Owner closes the tab
	resp.Body.Close()
	time.Sleep(100 * time.Millisecond)
	_, out := call(f.h, "GET", "/api/v1/runs", "", issuesManagement, false)
	if runs, _ := out["runs"].([]any); len(runs) != 1 {
		t.Fatalf("the run ended with the request: %v", out)
	}
	release(t, f.agentDir, "go")
	waitFor(t, func() bool {
		ex := m(f.only(t)["explanation"])
		return ex != nil && ex["text"] == "Looking. Found it."
	})
	_, logged := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	if m(logged["run"])["state"] != "done" || len(logged["log"].([]any)) < 2 {
		t.Fatalf("the run log: %v", logged)
	}
}

// aiport only sees the permission requests an agent sends. Agent settings
// in the repository that allow tools without asking are named in the run
// panel before the run starts.
func TestARunWarnsOfAgentSettingsThatAllowWithoutAsking(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "ok"}}}, "")
	for path, text := range map[string]string{
		".claude/settings.json": `{"permissions": {"allow": ["Bash(npm test:*)"], "defaultMode": "bypassPermissions"}}`,
		".gemini/settings.json": `{"autoAccept": true}`,
		".codex/config.toml":    "approval_policy = \"never\"\nsandbox_mode = \"danger-full-access\"\n",
	} {
		if err := os.MkdirAll(filepath.Join(f.repo, filepath.Dir(path)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(f.repo, path), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, f.repo, "add", "-A")
	gitIn(t, f.repo, "commit", "-q", "-m", "agent settings")
	_, evs := f.explainEvents(t, f.id)
	var warnings []string
	for _, e := range evs {
		if e["type"] == "run" {
			warnings = strs(m(e["run"])["warnings"])
		}
	}
	all := strings.Join(warnings, "\n")
	for _, s := range []string{".claude/settings.json", "bypassPermissions", "Bash(npm test:*)", ".gemini/settings.json", ".codex/config.toml", "never"} {
		if !strings.Contains(all, s) {
			t.Fatalf("the warnings do not name %q:\n%s", s, all)
		}
	}

	quiet := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "ok"}}}, "")
	_, evs = quiet.explainEvents(t, quiet.id)
	for _, e := range evs {
		if e["type"] == "run" && m(e["run"])["warnings"] != nil {
			t.Fatalf("warnings without settings: %v", e["run"])
		}
	}
}
