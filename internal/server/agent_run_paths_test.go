package server

import (
	"strings"
	"testing"
)

// The paths a permission request names are its locations and every path in
// its raw input (path, file_path, cwd, a move's destination, a patch's
// changed files, the base of a search pattern…): all must be inside the
// run's directory, and a read or search that names none is denied. In both
// modes the agent's configuration in the run's directory (and git's) is
// never the agent's to change.

func TestFixItNeverChangesAgentOrGitConfiguration(t *testing.T) {
	edit := func(title, path string) acpStep {
		return acpStep{Ask: &acpAsk{Kind: "edit", Title: title, Path: "CWD/" + path, Write: "x\n"}}
	}
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{
		edit("Claude settings", ".claude/settings.json"),
		edit("Claude memory", "CLAUDE.md"),
		edit("Claude memory in any case", "Claude.md"),
		edit("Nested AGENTS.md", "docs/AGENTS.md"),
		edit("Gemini memory", "GEMINI.md"),
		edit("MCP servers", ".mcp.json"),
		edit("Codex config", ".codex/config.toml"),
		edit("Gemini settings", ".gemini/settings.json"),
		edit("git attributes", ".gitattributes"),
		edit("nested git attributes", "src/.gitattributes"),
		edit("a git hook", ".git/hooks/pre-commit"),
		edit("husky hook", ".husky/pre-commit"),
		{Ask: &acpAsk{Kind: "move", Title: "Move into .claude", Path: "CWD/config.py", Input: map[string]any{"source": "CWD/config.py", "destination": "CWD/.claude/commands/run.md"}}},
		{Ask: &acpAsk{Kind: "edit", Title: "Patch outside", Path: "CWD/config.py", Input: map[string]any{"changes": map[string]any{"OUTSIDE/x.py": map[string]any{"type": "add"}}}}},
		edit("Notes about Claude", "docs/claude-notes.md"),
		{Say: "Done."},
	}})
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	f.fixEnded(t)
	want := "Claude settings=reject_once|Claude memory=reject_once|Claude memory in any case=reject_once|Nested AGENTS.md=reject_once|Gemini memory=reject_once|" +
		"MCP servers=reject_once|Codex config=reject_once|Gemini settings=reject_once|git attributes=reject_once|nested git attributes=reject_once|" +
		"a git hook=reject_once|husky hook=reject_once|Move into .claude=reject_once|Patch outside=reject_once|Notes about Claude=allow_once"
	if got := answers(t, f.agentDir); got != want {
		t.Fatalf("answers:\n%s\nwant\n%s", got, want)
	}
}

func TestReadsMustSayWhereAndStayInside(t *testing.T) {
	steps := []acpStep{
		{Ask: &acpAsk{Kind: "read", Title: "Read by input", Input: map[string]any{"file_path": "CWD/README.md"}}},
		{Ask: &acpAsk{Kind: "search", Title: "Search relative", Input: map[string]any{"path": "src", "pattern": "MODEL"}}},
		{Ask: &acpAsk{Kind: "read", Title: "Read unsaid"}},
		{Ask: &acpAsk{Kind: "search", Title: "Search unsaid", Input: map[string]any{"pattern": "MODEL"}}},
		{Ask: &acpAsk{Kind: "search", Title: "Search /etc", Input: map[string]any{"path": "/etc", "pattern": "x"}}},
		{Ask: &acpAsk{Kind: "read", Title: "Location inside, input outside", Path: "CWD/README.md", Input: map[string]any{"file_path": "/etc/hosts"}}},
		{Ask: &acpAsk{Kind: "search", Title: "Glob in home", Input: map[string]any{"pattern": "~/.ssh/*"}}},
		{Ask: &acpAsk{Kind: "search", Title: "Glob outside", Input: map[string]any{"pattern": "/etc/**/*.conf"}}},
		{Ask: &acpAsk{Kind: "search", Title: "Glob up", Input: map[string]any{"pattern": "../**/*.py"}}},
		{Ask: &acpAsk{Kind: "read", Title: "Nested cwd outside", Input: map[string]any{"options": map[string]any{"cwd": "/"}, "file_path": "README.md"}}},
		{Say: "Read."},
	}
	want := "Read by input=allow_once|Search relative=allow_once|Read unsaid=reject_once|Search unsaid=reject_once|Search /etc=reject_once|" +
		"Location inside, input outside=reject_once|Glob in home=reject_once|Glob outside=reject_once|Glob up=reject_once|Nested cwd outside=reject_once"

	explain := newAgentRunFixture(t, "claude", acpScript{Steps: steps}, "")
	explain.explainEvents(t, explain.id)
	if got := answers(t, explain.agentDir); got != want {
		t.Fatalf("Explain answers:\n%s\nwant\n%s", got, want)
	}

	fix, _ := newFixFixture(t, acpScript{Steps: steps})
	fix.fix(t)
	fix.fixEnded(t)
	if got := answers(t, fix.agentDir); got != want {
		t.Fatalf("Fix it answers:\n%s\nwant\n%s", got, want)
	}
	_, logged := call(fix.h, "GET", "/api/v1/issues/"+fix.id+"/run", "", issuesManagement, false)
	reasons := ""
	for _, u := range logged["log"].([]any) {
		if m(u)["kind"] == "permission" {
			reasons += m(u)["reason"].(string) + "\n"
		}
	}
	if !strings.Contains(reasons, "did not say") || !strings.Contains(reasons, "outside") {
		t.Fatalf("reasons:\n%s", reasons)
	}
}
