package extension

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
)

// script writes an executable shell script named name into dir and returns
// its absolute path.
func script(t *testing.T, dir, name, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Extensions in these tests are shell scripts")
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func policyInput() Input {
	return Input{Request: &Request{Application: "docs-bot", Kind: "application_alias", Supported: true, Need: "a cheaper model", Scope: "application"}}
}

func TestRun(t *testing.T) {
	long := strings.Repeat("x", MaxReason+20)
	tests := []struct {
		name    string
		body    string
		timeout int
		want    Output
	}{
		{"allow", `echo '{"decision":"allow","reason":"dropped on allow"}'`, 0, Output{Decision: "allow"}},
		{"allow with whitespace around", `printf '\n  {"decision":"allow"}  \n\n'`, 0, Output{Decision: "allow"}},
		{"deny with reason", `echo '{"decision":"deny","reason":"not on weekends"}'`, 0, deny("not on weekends")},
		{"deny without reason", `echo '{"decision":"deny"}'`, 0, deny("no reason given")},
		{"deny with blank reason", `printf '%s' '{"decision":"deny","reason":" \t "}'`, 0, deny("no reason given")},
		{"reason made one line", `printf '{"decision":"deny","reason":"line one\\nline\\ttwo\\u0007  end"}'`, 0, deny("line one line two end")},
		{"long reason cut", `echo '{"decision":"deny","reason":"` + long + `"}'`, 0, deny(strings.Repeat("x", MaxReason) + "…")},
		{"other decision", `echo '{"decision":"maybe"}'`, 0, deny("the Extension's decision must be allow or deny")},
		{"decision case matters", `echo '{"decision":"Allow"}'`, 0, deny("the Extension's decision must be allow or deny")},
		{"not JSON", `echo allow`, 0, deny("the Extension did not answer with a JSON decision")},
		{"two objects", `echo '{"decision":"allow"}{"decision":"allow"}'`, 0, deny("the Extension did not answer with a JSON decision")},
		{"no output", `exit 0`, 0, deny("the Extension did not answer with a JSON decision")},
		{"non-zero exit after allow", `echo '{"decision":"allow"}'; exit 3`, 0, deny("the Extension failed (exit status 3)")},
		{"too much output", `i=0; while [ $i -lt 1100 ]; do printf '%064d' 0; i=$((i+1)); done`, 0, deny("the Extension wrote more than 64 KiB")},
		{"timeout", `sleep 10; echo '{"decision":"allow"}'`, 200, deny("no decision within 200 ms")},
	}
	dir := t.TempDir()
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := script(t, dir, "ext"+string(rune('a'+i)), tt.body)
			started := time.Now()
			got := Run(config.Extension{Name: "check", Hook: config.HookPolicy, Path: p, TimeoutMS: tt.timeout}, policyInput())
			if got != tt.want {
				t.Errorf("Run = %+v, want %+v", got, tt.want)
			}
			if tt.timeout > 0 && time.Since(started) > 5*time.Second {
				t.Errorf("timeout took %v", time.Since(started))
			}
		})
	}
}

func TestRunCannotStart(t *testing.T) {
	dir := t.TempDir()
	notExec := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(notExec, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{filepath.Join(dir, "missing"), notExec, dir} {
		got := Run(config.Extension{Name: "check", Hook: config.HookPolicy, Path: p}, policyInput())
		if got != deny("the Extension could not be started") {
			t.Errorf("%s: Run = %+v", p, got)
		}
	}
}

func TestRunNeverLooksUpPATH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}
	got := Run(config.Extension{Name: "check", Hook: config.HookPolicy, Path: "true"}, policyInput())
	if got.Decision != "deny" || got.Reason != "the Extension could not be started" {
		t.Errorf("Run with a bare name = %+v", got)
	}
}

func TestRunPassesInputArgsDirAndMinimalEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIPORT_TEST_PROVIDER_KEY", "not-a-real-key")
	t.Setenv("HOME", "/home/docs-owner")
	p := script(t, dir, "record", `cat > input.json
printf '%s\n' "$@" > args.txt
env > env.txt
pwd -P > pwd.txt
echo '{"decision":"allow"}'`)
	in := policyInput()
	in.Hook, in.Extension = "ignored", "ignored"
	if got := Run(config.Extension{Name: "recorder", Hook: config.HookPolicy, Path: p, Args: []string{"--mode", "strict mode", "$HOME"}}, in); got.Decision != "allow" {
		t.Fatalf("Run = %+v", got)
	}
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	var gotIn Input
	if err := json.Unmarshal([]byte(read("input.json")), &gotIn); err != nil {
		t.Fatal(err)
	}
	if gotIn.Request == nil || *gotIn.Request != *in.Request || gotIn.Change != nil {
		t.Errorf("input = %s", read("input.json"))
	}
	if got := read("args.txt"); got != "--mode\nstrict mode\n$HOME\n" {
		t.Errorf("args = %q", got)
	}
	realDir, _ := filepath.EvalSymlinks(dir)
	if got := strings.TrimSpace(read("pwd.txt")); got != realDir {
		t.Errorf("dir = %q, want %q", got, realDir)
	}
	env := read("env.txt")
	if strings.Contains(env, "AIPORT_TEST_PROVIDER_KEY") || strings.Contains(env, "not-a-real-key") {
		t.Errorf("a secret reached the Extension:\n%s", env)
	}
	if !strings.Contains(env, "HOME=/home/docs-owner\n") {
		t.Errorf("HOME not passed:\n%s", env)
	}
}

func TestCheck(t *testing.T) {
	dir := t.TempDir()
	allow := script(t, dir, "allow", `echo '{"decision":"allow"}'`)
	denyWeekend := script(t, dir, "deny", `echo '{"decision":"deny","reason":"not on weekends"}'`)
	// mark records each run and the hook and name it was given.
	mark := script(t, dir, "mark", `input=$(cat); echo "$input" >> marks.txt; echo '{"decision":"allow"}'`)
	tests := []struct {
		name   string
		list   []config.Extension
		hook   string
		err    error
		marked int
	}{
		{"none listed", nil, config.HookPolicy, nil, 0},
		{"all allow", []config.Extension{{Name: "first", Hook: config.HookPolicy, Path: mark}, {Name: "second", Hook: config.HookPolicy, Path: allow}}, config.HookPolicy, nil, 1},
		{"first deny wins, later ones do not run", []config.Extension{{Name: "weekday", Hook: config.HookPolicy, Path: denyWeekend}, {Name: "later", Hook: config.HookPolicy, Path: mark}}, config.HookPolicy, Denied{Extension: "weekday", Hook: config.HookPolicy, Reason: "not on weekends"}, 0},
		{"other hook is not run", []config.Extension{{Name: "weekday", Hook: config.HookGuardrail, Path: denyWeekend}, {Name: "m", Hook: config.HookPolicy, Path: mark}}, config.HookPolicy, nil, 1},
		{"guardrail deny", []config.Extension{{Name: "m", Hook: config.HookGuardrail, Path: mark}, {Name: "weekday", Hook: config.HookGuardrail, Path: denyWeekend}}, config.HookGuardrail, Denied{Extension: "weekday", Hook: config.HookGuardrail, Reason: "not on weekends"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(filepath.Join(dir, "marks.txt"))
			err := Check(config.Config{Extensions: tt.list}, tt.hook, Input{Change: &Change{Item: "sug-1", Source: "suggestion", Application: "docs-bot", Kind: "bind_application", Scope: "application", By: "owner"}})
			if err != tt.err {
				t.Errorf("Check = %v, want %v", err, tt.err)
			}
			if tt.err != nil {
				var d Denied
				if !errors.As(err, &d) {
					t.Errorf("Check returned %T", err)
				}
			}
			b, _ := os.ReadFile(filepath.Join(dir, "marks.txt"))
			lines := strings.Split(strings.TrimSpace(string(b)), "\n")
			if len(b) == 0 {
				lines = nil
			}
			if len(lines) != tt.marked {
				t.Fatalf("mark ran %d times, want %d", len(lines), tt.marked)
			}
			for _, l := range lines {
				var in Input
				if err := json.Unmarshal([]byte(l), &in); err != nil {
					t.Fatal(err)
				}
				if in.Protocol != Protocol || in.Hook != tt.hook || in.Extension == "" || in.Change == nil || in.Change.Item != "sug-1" {
					t.Errorf("mark got %s", l)
				}
			}
		})
	}
}

func TestDeniedError(t *testing.T) {
	tests := []struct {
		d    Denied
		want string
	}{
		{Denied{Extension: "weekday", Hook: config.HookPolicy, Reason: "not on weekends"}, "refused by policy weekday: not on weekends"},
		{Denied{Extension: "budget", Hook: config.HookGuardrail, Reason: "over budget"}, "blocked by guardrail budget: over budget"},
		{Denied{Extension: "x", Reason: "r"}, "refused by policy x: r"},
	}
	for _, tt := range tests {
		if got := tt.d.Error(); got != tt.want {
			t.Errorf("Error() = %q, want %q", got, tt.want)
		}
	}
}

func TestEnviron(t *testing.T) {
	t.Setenv("PATH", "/usr/bin:/bin")
	t.Setenv("AIPORT_TEST_EXTRA", "extra-value")
	t.Setenv("AIPORT_TEST_SECRET", "not-a-real-secret")
	env := Environ("AIPORT_TEST_EXTRA", "AIPORT_TEST_UNSET_VARIABLE")
	has := map[string]bool{}
	for _, kv := range env {
		has[kv] = true
		k, _, _ := strings.Cut(kv, "=")
		allowed := k == "AIPORT_TEST_EXTRA"
		for _, p := range passedEnv {
			allowed = allowed || k == p
		}
		if !allowed {
			t.Errorf("Environ passed %q", kv)
		}
	}
	if !has["PATH=/usr/bin:/bin"] || !has["AIPORT_TEST_EXTRA=extra-value"] {
		t.Errorf("Environ = %q", env)
	}
	for _, kv := range env {
		if strings.HasPrefix(kv, "AIPORT_TEST_UNSET_VARIABLE") {
			t.Error("an unset variable was passed")
		}
	}
	before := append([]string{}, passedEnv...)
	Environ("A", "B", "C")
	if strings.Join(before, ",") != strings.Join(passedEnv, ",") {
		t.Error("Environ changed passedEnv")
	}
}

func TestCapped(t *testing.T) {
	tests := []struct {
		name   string
		max    int
		writes []string
		kept   string
		over   bool
	}{
		{"under", 10, []string{"abc", "def"}, "abcdef", false},
		{"exactly", 6, []string{"abc", "def"}, "abcdef", false},
		{"over in one write", 4, []string{"abcdef"}, "abcd", true},
		{"over across writes", 5, []string{"abc", "def", "ghi"}, "abcde", true},
		{"zero", 0, []string{"a"}, "", true},
		{"nothing written", 3, nil, "", false},
		{"empty write when full", 3, []string{"abc", ""}, "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Capped{Max: tt.max}
			for _, w := range tt.writes {
				n, err := c.Write([]byte(w))
				if n != len(w) || err != nil {
					t.Fatalf("Write(%q) = %d, %v", w, n, err)
				}
			}
			if c.String() != tt.kept || string(c.Bytes()) != tt.kept || c.Over() != tt.over {
				t.Errorf("kept %q over %v, want %q %v", c.String(), c.Over(), tt.kept, tt.over)
			}
		})
	}
}

func TestIsolateKillsTheProcessGroup(t *testing.T) {
	dir := t.TempDir()
	// The child keeps stdout open; without killing the group the run would
	// wait for it.
	p := script(t, dir, "spawn", `sleep 30 &
sleep 30`)
	started := time.Now()
	got := Run(config.Extension{Name: "spawner", Hook: config.HookPolicy, Path: p, TimeoutMS: 200}, policyInput())
	if got != deny("no decision within 200 ms") {
		t.Errorf("Run = %+v", got)
	}
	if d := time.Since(started); d > 3*time.Second {
		t.Errorf("run took %v: the background child was not stopped", d)
	}
}
