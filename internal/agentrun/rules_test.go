package agentrun

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// worktree is a resolved temporary directory (macOS links /var to
// /private/var) with a src/ folder and a link that points outside it.
func worktree(t *testing.T) (root, outside string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	outside, err = filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	return root, outside
}

func input(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestReadOnly(t *testing.T) {
	root, outside := worktree(t)
	policy := ReadOnly(root)
	cases := []struct {
		name  string
		p     Permission
		allow bool
		why   string
	}{
		{"thinking", Permission{Kind: ToolThink}, true, "thinking"},
		{"read inside", Permission{Kind: ToolRead, Paths: []string{filepath.Join(root, "src/main.go")}}, true, "reading inside"},
		{"relative read", Permission{Kind: ToolRead, Paths: []string{"src/main.go"}}, true, "reading inside"},
		{"search with a relative pattern", Permission{Kind: ToolSearch, Paths: []string{root}, Input: input(t, map[string]any{"pattern": "**/*.go"})}, true, "reading inside"},
		{"read without a path", Permission{Kind: ToolRead}, false, "did not say where"},
		{"read outside", Permission{Kind: ToolRead, Paths: []string{filepath.Join(outside, "x")}}, false, "outside"},
		{"read through a link that leaves", Permission{Kind: ToolRead, Paths: []string{"escape/secret.txt"}}, false, "outside"},
		{"read climbing out", Permission{Kind: ToolRead, Paths: []string{"../x"}}, false, "outside"},
		{"read in the home directory", Permission{Kind: ToolRead, Input: input(t, map[string]any{"file_path": "~/.ssh/id"})}, false, "outside"},
		{"search pattern leaving root", Permission{Kind: ToolSearch, Paths: []string{root}, Input: input(t, map[string]any{"glob": outside + "/**"})}, false, "outside"},
		{"path in nested input", Permission{Kind: ToolRead, Paths: []string{root}, Input: input(t, map[string]any{"args": []any{map[string]any{"path": "/etc/passwd"}}})}, false, "outside"},
		{"unparsable input", Permission{Kind: ToolRead, Paths: []string{root}, Input: json.RawMessage(`{`)}, false, "outside"},
		{"edit", Permission{Kind: ToolEdit, Paths: []string{"src/main.go"}}, false, "read-only"},
		{"command", Permission{Kind: ToolExecute, Input: input(t, "ls")}, false, "read-only"},
		{"fetch", Permission{Kind: ToolFetch}, false, "read-only"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := policy(context.Background(), c.p)
			if v.Allow != c.allow || !strings.Contains(v.Reason, c.why) {
				t.Fatalf("got %+v, want allow=%v with a reason containing %q", v, c.allow, c.why)
			}
		})
	}
}

func TestFixRules(t *testing.T) {
	root, outside := worktree(t)
	var asked []Permission
	command := func(_ context.Context, p Permission) Verdict {
		asked = append(asked, p)
		return Verdict{Allow: true, Reason: "the Owner allowed it"}
	}
	policy := FixRules(root, command)
	cases := []struct {
		name  string
		p     Permission
		allow bool
		why   string
	}{
		{"read inside", Permission{Kind: ToolRead, Paths: []string{"src/main.go"}}, true, "reading inside the run's worktree"},
		{"edit inside", Permission{Kind: ToolEdit, Paths: []string{"src/main.go"}}, true, "editing inside"},
		{"new file in a new folder", Permission{Kind: ToolEdit, Paths: []string{"src/new/deep/file.go"}}, true, "editing inside"},
		{"delete inside", Permission{Kind: ToolDelete, Paths: []string{"src/old.go"}}, true, "editing inside"},
		{"move inside", Permission{Kind: ToolMove, Input: input(t, map[string]any{"from": "src/a.go", "to": "src/b.go"})}, true, "editing inside"},
		{"patch inside", Permission{Kind: ToolEdit, Input: input(t, map[string]any{"changes": map[string]any{"src/a.go": map[string]any{"diff": "+x"}}})}, true, "editing inside"},
		{"edit without a path", Permission{Kind: ToolEdit}, false, "did not say which file"},
		{"edit outside", Permission{Kind: ToolEdit, Paths: []string{filepath.Join(outside, "x.go")}}, false, "outside"},
		{"edit through a link", Permission{Kind: ToolEdit, Paths: []string{"escape/x.go"}}, false, "outside"},
		{"move out of the worktree", Permission{Kind: ToolMove, Input: input(t, map[string]any{"from": "src/a.go", "to": filepath.Join(outside, "a.go")})}, false, "outside"},
		{"patch leaving the worktree", Permission{Kind: ToolEdit, Input: input(t, map[string]any{"changes": map[string]any{"../x.go": map[string]any{}}})}, false, "outside"},
		{"git configuration", Permission{Kind: ToolEdit, Paths: []string{".git/hooks/pre-commit"}}, false, ".git is not"},
		{"agent folder in any case", Permission{Kind: ToolEdit, Paths: []string{"src/.Claude/settings.json"}}, false, ".claude is not"},
		{"agent instructions", Permission{Kind: ToolEdit, Paths: []string{"AGENTS.md"}}, false, "AGENTS.md is not"},
		{"MCP config deep down", Permission{Kind: ToolDelete, Paths: []string{"src/.mcp.json"}}, false, ".mcp.json is not"},
		{"a folder named like a protected file", Permission{Kind: ToolEdit, Paths: []string{"AGENTS.md/notes.txt"}}, true, "editing inside"},
		{"the worktree itself", Permission{Kind: ToolDelete, Paths: []string{root}}, false, "the run's directory itself"},
		{"command inside goes to the Owner", Permission{Kind: ToolExecute, Input: input(t, map[string]any{"command": "go test ./...", "cwd": root})}, true, "the Owner allowed it"},
		{"command outside", Permission{Kind: ToolExecute, Input: input(t, map[string]any{"command": "ls", "cwd": outside})}, false, "a command outside"},
		{"fetch", Permission{Kind: ToolFetch}, false, "only reading and editing"},
		{"switch mode", Permission{Kind: ToolSwitchMode}, false, "only reading and editing"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := policy(context.Background(), c.p)
			if v.Allow != c.allow || !strings.Contains(v.Reason, c.why) {
				t.Fatalf("got %+v, want allow=%v with a reason containing %q", v, c.allow, c.why)
			}
		})
	}
	if len(asked) != 1 {
		t.Fatalf("the Owner was asked %d times, want once (only the command inside the worktree)", len(asked))
	}
}

func TestPatternBase(t *testing.T) {
	for pattern, want := range map[string]string{
		"**/*.go":         "",
		"src/*.go":        "",
		"^func [A-Z]":     "",
		"/etc/*.conf":     "/etc/",
		"~/notes/*":       "~/notes/",
		"../other/**":     "../other/",
		"src/../../x/*.y": "src/../../x/",
		"/*":              "/",
		"*":               "",
	} {
		if got := patternBase(pattern); got != want {
			t.Errorf("patternBase(%q) = %q, want %q", pattern, got, want)
		}
	}
}
