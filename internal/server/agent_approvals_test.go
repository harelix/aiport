package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Approving commands during Fix it (harelix/aiport#26): a command the agent
// wants to run inside its worktree waits on the dashboard for the Owner's
// answer: Allow once, Allow this command for this run, or Deny. Anything
// outside the worktree is denied without asking. A request left unanswered
// stops the run, which keeps what it produced so far.

// pytest is a command request: run "pytest -q" in the worktree; once allowed
// the agent writes its result file (the command's effect).
func pytest(command, result string) acpStep {
	return acpStep{Ask: &acpAsk{Kind: "execute", Title: "Run " + command, Input: map[string]any{"command": command, "cwd": "CWD"}, Path: "CWD/" + result, Write: "passed\n"}}
}

// approvals is the list of commands waiting for the Owner.
func (f *agentRunFixture) approvals(t *testing.T) []map[string]any {
	t.Helper()
	w, out := call(f.h, "GET", "/api/v1/approvals", "", issuesManagement, false)
	if w.Code != 200 {
		t.Fatalf("approvals: %d %s", w.Code, w.Body)
	}
	var list []map[string]any
	for _, a := range out["approvals"].([]any) {
		list = append(list, m(a))
	}
	return list
}

// waitApproval waits until a command waits for the Owner and returns it.
func (f *agentRunFixture) waitApproval(t *testing.T) map[string]any {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		if list := f.approvals(t); len(list) > 0 {
			return list[0]
		}
		if time.Now().After(deadline) {
			_, run := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
			t.Fatalf("no command waits for the Owner; run: %v", run)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// answer answers approval id.
func (f *agentRunFixture) answer(t *testing.T, id, answer string) {
	t.Helper()
	if w, out := call(f.h, "POST", "/api/v1/approvals/"+id, `{"answer":"`+answer+`"}`, issuesManagement, true); w.Code != 200 {
		t.Fatalf("answer %s: %d %v", answer, w.Code, out)
	}
}

// answers are the fake agent's permission answers, title=option.
func answers(t *testing.T, dir string) string {
	var got []string
	for _, raw := range agentRecord(t, dir)["permission"] {
		var p struct{ Title, Option string }
		_ = json.Unmarshal(raw, &p)
		got = append(got, p.Title+"="+strings.TrimPrefix(p.Option, "opt-"))
	}
	return strings.Join(got, "|")
}

func TestTheOwnerApprovesTheCommandsOfAFixItRun(t *testing.T) {
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{
		editConfig.Steps[0],
		pytest("pytest -q", "result-1"),
		pytest("pytest -q", "result-2"),
		pytest("pytest -q tests/<b>models</b>", "result-3"),
		{Ask: &acpAsk{Kind: "execute", Title: "Run outside", Input: map[string]any{"command": "pytest -q", "cwd": "OUTSIDE"}}},
		{Ask: &acpAsk{Kind: "execute", Title: "Run on the Owner's copy", Input: map[string]any{"command": "pytest -q"}, Path: "REPO/config.py"}},
		pytest("make test", "result-4"),
		{Say: "config.py now asks for qwen2.5:7b; the tests pass."},
	}})
	code, out := f.fix(t)
	if code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	run := m(out["run"])["id"]

	// The first command waits for the Owner, with its text, the run and the
	// three answers.
	a := f.waitApproval(t)
	if a["command"] != "pytest -q" || a["application"] != "docs-bot" || a["issue"] != f.id || a["run"] != run || a["expires_at_ms"] == nil {
		t.Fatalf("approval: %v", a)
	}
	if got := strings.Join(strs(a["answers"]), " "); got != "allow_once allow_run deny" {
		t.Fatalf("answers offered: %s", got)
	}
	// The run waits: it says so to the dashboard (the badge).
	_, runs := call(f.h, "GET", "/api/v1/runs", "", issuesManagement, false)
	if r := m(runs["runs"].([]any)[0]); r["waiting"] != float64(1) {
		t.Fatalf("the run does not say it waits: %v", r)
	}
	_, logged := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	if l, _ := logged["approvals"].([]any); len(l) != 1 || m(l[0])["command"] != "pytest -q" {
		t.Fatalf("the Issue's run does not show the waiting command: %v", logged["approvals"])
	}

	// Management only, from the dashboard.
	if w, _ := call(f.h, "GET", "/api/v1/approvals", "", docsBotToken, false); w.Code == 200 {
		t.Fatalf("an Application token listed approvals: %d", w.Code)
	}
	if w, _ := call(f.h, "POST", "/api/v1/approvals/"+a["id"].(string), `{"answer":"allow_once"}`, docsBotToken, true); w.Code == 200 {
		t.Fatalf("an Application token answered: %d", w.Code)
	}
	if w, _ := call(f.h, "POST", "/api/v1/approvals/"+a["id"].(string), `{"answer":"allow_once"}`, issuesManagement, false); w.Code != 403 {
		t.Fatalf("an answer without CSRF: %d", w.Code)
	}
	if w, out := call(f.h, "POST", "/api/v1/approvals/"+a["id"].(string), `{"answer":"always"}`, issuesManagement, true); w.Code != 400 {
		t.Fatalf("an unknown answer: %d %v", w.Code, out)
	}

	// Allow this command for this run: its repeat is allowed without asking;
	// another command asks again.
	f.answer(t, a["id"].(string), "allow_run")
	if w, _ := call(f.h, "POST", "/api/v1/approvals/"+a["id"].(string), `{"answer":"deny"}`, issuesManagement, true); w.Code != 404 {
		t.Fatalf("answering twice: %d", w.Code)
	}
	b := f.waitApproval(t)
	if b["command"] != "pytest -q tests/<b>models</b>" {
		t.Fatalf("second approval: %v", b)
	}
	f.answer(t, b["id"].(string), "deny")
	// Outside the worktree: denied without asking. make test asks; allow once.
	c := f.waitApproval(t)
	if c["command"] != "make test" {
		t.Fatalf("third approval: %v", c)
	}
	f.answer(t, c["id"].(string), "allow_once")

	is := f.fixEnded(t)
	if bf := m(is["branch_fix"]); bf["state"] != "ready" {
		t.Fatalf("branch fix: %v", bf)
	}
	want := "Edit config.py=allow_once|Run pytest -q=allow_once|Run pytest -q=allow_once|Run pytest -q tests/<b>models</b>=reject_once|" +
		"Run outside=reject_once|Run on the Owner's copy=reject_once|Run make test=allow_once"
	if got := answers(t, f.agentDir); got != want {
		t.Fatalf("answers:\n%s\nwant\n%s", got, want)
	}
	if len(f.approvals(t)) != 0 {
		t.Fatal("approvals left after the run")
	}
	// The commands' effects are the agent's, in its worktree: in the diff.
	_, diff := call(f.h, "GET", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, false)
	text := diff["diff"].(string)
	if !strings.Contains(text, "result-1") || !strings.Contains(text, "result-2") || strings.Contains(text, "result-3") || !strings.Contains(text, "result-4") {
		t.Fatalf("diff: %s", text)
	}

	// Allow this command for this run covers that run only: the next run
	// asks again. Stopping a waiting run ends the request.
	if w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/discard", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("discard: %d %v", w.Code, out)
	}
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix again: %d %v", code, out)
	}
	again := f.waitApproval(t)
	if again["command"] != "pytest -q" || again["run"] == run {
		t.Fatalf("the next run: %v", again)
	}
	if w, out := call(f.h, "POST", "/api/v1/runs/"+again["run"].(string)+"/stop", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("stop: %d %v", w.Code, out)
	}
	if bf := m(f.fixEnded(t)["branch_fix"]); bf["state"] != "stopped" || bf["worktree"] != nil {
		t.Fatalf("stopped by the Owner: %v", bf)
	}
	if len(f.approvals(t)) != 0 {
		t.Fatal("approvals left after the run was stopped")
	}
}

// A command left unanswered stops the run; what it produced so far stays in
// its worktree for the Owner to review, keep or discard.
func TestAnUnansweredCommandStopsTheRunAndKeepsItsWork(t *testing.T) {
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{
		editConfig.Steps[0],
		pytest("pytest -q", "result-1"),
		{Say: "Done."},
	}})
	f.h.approvalWait = 300 * time.Millisecond
	before := owner(t, f.repo)
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	if a := f.waitApproval(t); a["command"] != "pytest -q" {
		t.Fatalf("approval: %v", a)
	}
	is := f.fixEnded(t)
	b := m(is["branch_fix"])
	if b["state"] != "stopped_waiting" || b["summary"] != "1 file changed, 1 insertion(+), 1 deletion(-)" || !strings.Contains(b["note"].(string), "unanswered") {
		t.Fatalf("branch fix: %v", b)
	}
	tree := b["worktree"].(string)
	if _, err := os.Stat(filepath.Join(tree, "config.py")); err != nil {
		t.Fatalf("the worktree is gone: %v", err)
	}
	if len(f.approvals(t)) != 0 {
		t.Fatal("the unanswered request is still listed")
	}
	if got := answers(t, f.agentDir); !strings.HasPrefix(got, "Edit config.py=allow_once") || strings.Contains(got, "Run pytest -q=allow_once") {
		t.Fatalf("answers: %s", got)
	}
	_, diff := call(f.h, "GET", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, false)
	if !strings.Contains(diff["diff"].(string), "+MODEL = 'qwen2.5:7b'") {
		t.Fatalf("diff: %v", diff)
	}
	if w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, true); w.Code != 409 || m(out["error"])["code"] != "fix_waiting" {
		t.Fatalf("Fix it again before deciding: %d %v", w.Code, out)
	}
	w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/keep", "", issuesManagement, true)
	if w.Code != 200 || m(m(out["issue"])["branch_fix"])["state"] != "kept" {
		t.Fatalf("keep: %d %v", w.Code, out)
	}
	if after := owner(t, f.repo); after != before {
		t.Fatalf("the Owner's working copy changed:\n%s\n%s", before, after)
	}
}

// A Fix it run going when aiport stopped is marked stopped at the next
// start, and its worktree and branch are removed.
func TestARunLeftGoingByARestartIsMarkedStopped(t *testing.T) {
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{editConfig.Steps[0], {Wait: "never"}}})
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	b := m(f.only(t)["branch_fix"])
	if b["state"] != "running" {
		t.Fatalf("not running: %v", b)
	}
	old := f.h
	f.start(t) // the same configuration and database: as after a crash
	f.useAgentDirs()
	after := m(f.only(t)["branch_fix"])
	if after["state"] != "stopped" || after["worktree"] != nil || !strings.Contains(after["note"].(string), "aiport") {
		t.Fatalf("after a restart: %v", after)
	}
	if hasBranch(t, f.repo, b["branch"].(string)) || len(worktrees(t, f.repo)) != 1 {
		t.Fatalf("left behind: %v", worktrees(t, f.repo))
	}
	old.Close()
}

// An Explain run's throwaway worktree left by a crash is removed the next
// time aiport starts.
func TestAnExplainWorktreeLeftByARestartIsRemoved(t *testing.T) {
	f := newAgentRunFixture(t, "claude", acpScript{Steps: []acpStep{{Say: "Reading"}, {Wait: "never"}}}, "")
	events := f.explainStream(t)
	until(t, events, "text")
	var start struct{ Cwd string }
	_ = json.Unmarshal(agentRecord(t, f.agentDir)["start"][0], &start)
	if _, err := os.Stat(start.Cwd); err != nil {
		t.Fatalf("no Explain worktree: %v", err)
	}
	t.Cleanup(f.h.Close) // before the stream's server closes
	f.start(t)           // as after a crash
	if _, err := os.Stat(start.Cwd); err == nil {
		t.Fatalf("the Explain worktree %s is left", start.Cwd)
	}
}

// run is a command request for argv (a list of arguments) in the worktree.
func run(title string, argv ...any) acpStep {
	return acpStep{Ask: &acpAsk{Kind: "execute", Title: title, Input: map[string]any{"command": argv, "cwd": "CWD"}}}
}

// "Allow until the next edit" lets the same command run again without
// asking only until the agent changes a file, asked or not: an edit may
// change what the command does.
func TestAllowUntilTheNextEdit(t *testing.T) {
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{
		pytest("pytest -q", "r1"),
		pytest("pytest -q", "r2"),
		{Edit: &acpEdit{Path: "CWD/tests/conftest.py", Text: "import os\n"}},
		pytest("pytest -q", "r3"),
		pytest("pytest -q", "r4"),
		editConfig.Steps[0],
		pytest("pytest -q", "r5"),
		{Say: "Done."},
	}})
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	f.answer(t, f.waitApproval(t)["id"].(string), "allow_run") // r1
	f.answer(t, f.waitApproval(t)["id"].(string), "allow_run") // r3: after the unasked edit
	f.answer(t, f.waitApproval(t)["id"].(string), "deny")      // r5: after the allowed edit
	f.fixEnded(t)
	want := "Run pytest -q=allow_once|Run pytest -q=allow_once|Run pytest -q=allow_once|Run pytest -q=allow_once|Edit config.py=allow_once|Run pytest -q=reject_once"
	if got := answers(t, f.agentDir); got != want {
		t.Fatalf("answers:\n%s\nwant\n%s", got, want)
	}
}

// An allowance matches the exact command the agent sent: the same list of
// arguments (or the same string), never one that only looks the same when
// joined with spaces.
func TestAnAllowanceMatchesTheExactArguments(t *testing.T) {
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{
		run("one argument", "echo", "a b"),
		run("two arguments", "echo", "a", "b"),
		{Ask: &acpAsk{Kind: "execute", Title: "a string", Input: map[string]any{"command": "echo a b", "cwd": "CWD"}}},
		run("one argument again", "echo", "a b"),
		{Say: "Done."},
	}})
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	a := f.waitApproval(t)
	if a["command"] != "echo 'a b'" {
		t.Fatalf("shown: %q", a["command"])
	}
	f.answer(t, a["id"].(string), "allow_run")
	b := f.waitApproval(t)
	if b["command"] != "echo a b" || b["title"] != "two arguments" {
		t.Fatalf("second: %v", b)
	}
	f.answer(t, b["id"].(string), "deny")
	c := f.waitApproval(t)
	if c["command"] != "echo a b" || c["title"] != "a string" {
		t.Fatalf("third: %v", c)
	}
	f.answer(t, c["id"].(string), "deny")
	f.fixEnded(t)
	want := "one argument=allow_once|two arguments=reject_once|a string=reject_once|one argument again=allow_once"
	if got := answers(t, f.agentDir); got != want {
		t.Fatalf("answers:\n%s\nwant\n%s", got, want)
	}
}

// The Owner approves exactly what runs: the whole command, however long,
// with invisible and direction-changing characters shown as escapes. A
// command too long to show whole is denied without asking.
func TestTheOwnerSeesTheWholeCommand(t *testing.T) {
	long := "pytest " + strings.Repeat("tests/test_models.py ", 300)
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{
		{Ask: &acpAsk{Kind: "execute", Title: "Run tricky", Input: map[string]any{"command": "echo safe\u202e;rm -rf .\u200b\x1b[2K", "cwd": "CWD"}}},
		{Ask: &acpAsk{Kind: "execute", Title: "Run long", Input: map[string]any{"command": long, "cwd": "CWD"}}},
		{Ask: &acpAsk{Kind: "execute", Title: "Run huge", Input: map[string]any{"command": strings.Repeat("x", 64<<10), "cwd": "CWD"}}},
		{Say: "Done."},
	}})
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	a := f.waitApproval(t)
	if want := "echo safe\\u202e;rm -rf .\\u200b\\u001b[2K"; a["command"] != want {
		t.Fatalf("shown: %q want %q", a["command"], want)
	}
	f.answer(t, a["id"].(string), "deny")
	b := f.waitApproval(t)
	if b["command"] != long {
		t.Fatalf("the long command was cut: %d bytes of %d", len(b["command"].(string)), len(long))
	}
	f.answer(t, b["id"].(string), "deny")
	f.fixEnded(t)
	if got := answers(t, f.agentDir); got != "Run tricky=reject_once|Run long=reject_once|Run huge=reject_once" {
		t.Fatalf("answers: %s", got)
	}
}
