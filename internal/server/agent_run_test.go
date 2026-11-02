package server

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
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
	if real, _ := filepath.EvalSymlinks(start.Cwd); real != f.repo {
		t.Fatalf("the agent ran in %q, not the repository %q", start.Cwd, f.repo)
	}
	var session struct {
		Cwd        string `json:"cwd"`
		MCPServers []any  `json:"mcpServers"`
	}
	_ = json.Unmarshal(rec["session"][0], &session)
	if session.Cwd != f.repo || session.MCPServers == nil || len(session.MCPServers) != 0 {
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
