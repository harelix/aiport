package server

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Fix it (harelix/aiport#25): aiport makes a git worktree of the
// Application's repository on a new branch named for the Issue and runs the
// Coding agent there. Edits inside the worktree are allowed; every command
// and anything outside the worktree is denied. The run ends with a diff:
// Keep commits it on the new branch (as aiport) and leaves the branch;
// Discard removes the worktree and the branch. The Owner's working copy and
// branch are never touched; nothing is merged or pushed.

// gitIn runs git in dir for a test and returns its trimmed output.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=Owner", "-c", "user.email=owner@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newFixFixture is the agent-run fixture whose repository has one commit on
// the Owner's branch "main": config.py asking for llama3:8b, and a symbolic
// link "shared" to a directory outside the repository.
func newFixFixture(t *testing.T, script acpScript) (*agentRunFixture, string) {
	t.Helper()
	f := newAgentRunFixture(t, "claude", script, "")
	outside := t.TempDir()
	if real, err := filepath.EvalSymlinks(outside); err == nil {
		outside = real
	}
	gitIn(t, f.repo, "checkout", "-q", "-B", "main")
	if err := os.WriteFile(filepath.Join(f.repo, "config.py"), []byte("MODEL = 'llama3:8b'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(f.repo, "shared")); err != nil {
		t.Fatal(err)
	}
	gitIn(t, f.repo, "add", "-A")
	gitIn(t, f.repo, "commit", "-q", "-m", "first")
	writeScript(t, f.agentDir, f.repo)
	path := filepath.Join(f.agentDir, "script.json")
	data, _ := os.ReadFile(path)
	_ = os.WriteFile(path, []byte(strings.ReplaceAll(string(data), "OUTSIDE", outside)), 0o600)
	return f, outside
}

// fix starts Fix it for the fixture's Issue.
func (f *agentRunFixture) fix(t *testing.T) (int, map[string]any) {
	t.Helper()
	w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, true)
	return w.Code, out
}

// fixEnded waits until the Issue's Fix it run has ended and returns the
// Issue.
func (f *agentRunFixture) fixEnded(t *testing.T) map[string]any {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		is := f.only(t)
		if b := m(is["branch_fix"]); b != nil && b["state"] != "running" {
			return is
		}
		if time.Now().After(deadline) {
			_, run := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
			t.Fatalf("the Fix it run did not end: %v\nrun: %v", is["branch_fix"], run)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// owner is the state of the Owner's working copy: branch, commit, status
// and config.py.
func owner(t *testing.T, repo string) string {
	t.Helper()
	data, _ := os.ReadFile(filepath.Join(repo, "config.py"))
	return gitIn(t, repo, "symbolic-ref", "HEAD") + " " + gitIn(t, repo, "rev-parse", "HEAD") + " [" + gitIn(t, repo, "status", "--porcelain") + "] " + string(data)
}

func TestFixItEditsOnlyInsideItsOwnWorktree(t *testing.T) {
	f, outside := newFixFixture(t, acpScript{Steps: []acpStep{
		{Ask: &acpAsk{Kind: "read", Title: "Read config.py", Path: "CWD/config.py"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Edit config.py", Path: "CWD/config.py", Write: "MODEL = 'qwen2.5:7b'\n"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Add models/pick.py", Path: "CWD/models/pick.py", Write: "def pick():\n    return 'qwen2.5:7b'\n"}},
		{Ask: &acpAsk{Kind: "execute", Title: "Run pytest"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Edit the Owner's copy", Path: "REPO/config.py", Write: "MODEL = 'owner copy changed'\n"}},
		{Ask: &acpAsk{Kind: "read", Title: "Read the Owner's copy", Path: "REPO/config.py"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Escape with ..", Path: "CWD/../escape.py", Write: "x"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Write outside", Path: "OUTSIDE/abs.py", Write: "x"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Write through a link", Path: "CWD/shared/linked.py", Write: "x"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Edit .git", Path: "CWD/.git", Write: "gitdir: /nowhere\n"}},
		{Ask: &acpAsk{Kind: "edit", Title: "Edit somewhere unsaid"}},
		{Ask: &acpAsk{Kind: "fetch", Title: "Fetch a page"}},
		{Say: "config.py now asks for qwen2.5:7b, which the local Connection serves."},
	}})
	before := owner(t, f.repo)

	code, out := f.fix(t)
	run := m(out["run"])
	if code != 202 || run["mode"] != "fix" || run["agent"] != "claude" || run["issue"] != f.id {
		t.Fatalf("fix: %d %v", code, out)
	}
	is := f.fixEnded(t)
	b := m(is["branch_fix"])
	branch := "aiport/fix-" + f.id
	if b["state"] != "ready" || b["branch"] != branch || b["agent"] != "claude" || b["run"] != run["id"] {
		t.Fatalf("branch fix: %v", b)
	}
	if b["summary"] != "2 files changed, 3 insertions(+), 1 deletion(-)" || strings.Join(strs(b["files"]), " ") != "config.py models/pick.py" {
		t.Fatalf("diff summary: %v", b)
	}

	// The agent worked in a worktree on the new branch, outside the repository.
	var start struct{ Cwd string }
	_ = json.Unmarshal(agentRecord(t, f.agentDir)["start"][0], &start)
	tree, _ := filepath.EvalSymlinks(start.Cwd)
	if tree == "" || tree == f.repo || strings.HasPrefix(tree, f.repo+string(filepath.Separator)) {
		t.Fatalf("the agent worked in %q (repository %q)", start.Cwd, f.repo)
	}
	if got := gitIn(t, tree, "symbolic-ref", "--short", "HEAD"); got != branch {
		t.Fatalf("the worktree is on %q", got)
	}

	// What aiport allowed and denied.
	var got []string
	for _, raw := range agentRecord(t, f.agentDir)["permission"] {
		var p struct{ Title, Option string }
		_ = json.Unmarshal(raw, &p)
		got = append(got, p.Title+"="+strings.TrimPrefix(p.Option, "opt-"))
	}
	want := "Read config.py=allow_once|Edit config.py=allow_once|Add models/pick.py=allow_once|Run pytest=reject_once|Edit the Owner's copy=reject_once|Read the Owner's copy=reject_once|" +
		"Escape with ..=reject_once|Write outside=reject_once|Write through a link=reject_once|Edit .git=reject_once|Edit somewhere unsaid=reject_once|Fetch a page=reject_once"
	if strings.Join(got, "|") != want {
		t.Fatalf("answers:\n%s\nwant\n%s", strings.Join(got, "|"), want)
	}
	_, logged := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
	reasons := ""
	for _, u := range logged["log"].([]any) {
		if m(u)["kind"] == "permission" {
			reasons += m(u)["text"].(string) + ": " + m(u)["reason"].(string) + "\n"
		}
	}
	if !strings.Contains(reasons, "Run pytest: ") || !strings.Contains(reasons, "command") {
		t.Fatalf("the run log does not say why the command was denied:\n%s", reasons)
	}

	// The diff the Owner reviews.
	w, diff := call(f.h, "GET", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, false)
	text, _ := diff["diff"].(string)
	if w.Code != 200 || !strings.Contains(text, "+MODEL = 'qwen2.5:7b'") || !strings.Contains(text, "-MODEL = 'llama3:8b'") || !strings.Contains(text, "models/pick.py") || m(diff["branch_fix"])["branch"] != branch {
		t.Fatalf("diff: %d %v", w.Code, diff)
	}

	// Nothing outside the worktree changed; nothing was committed yet.
	if after := owner(t, f.repo); after != before {
		t.Fatalf("the Owner's working copy changed:\n%s\n%s", before, after)
	}
	for _, p := range []string{filepath.Join(outside, "abs.py"), filepath.Join(outside, "linked.py"), filepath.Join(filepath.Dir(tree), "escape.py")} {
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("%s was written", p)
		}
	}
	if gitIn(t, f.repo, "rev-parse", branch) != gitIn(t, f.repo, "rev-parse", "main") {
		t.Fatal("aiport committed before the Owner reviewed the diff")
	}
}

func strs(v any) []string {
	var out []string
	l, _ := v.([]any)
	for _, x := range l {
		out = append(out, x.(string))
	}
	return out
}

// editConfig is a Fix it script that changes config.py to ask for
// qwen2.5:7b.
var editConfig = acpScript{Steps: []acpStep{
	{Ask: &acpAsk{Kind: "edit", Title: "Edit config.py", Path: "CWD/config.py", Write: "MODEL = 'qwen2.5:7b'\n"}},
	{Say: "config.py now asks for qwen2.5:7b."},
}}

// worktrees lists the repository's worktrees (the Owner's first).
func worktrees(t *testing.T, repo string) []string {
	var out []string
	for _, line := range strings.Split(gitIn(t, repo, "worktree", "list", "--porcelain"), "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok {
			out = append(out, p)
		}
	}
	return out
}

// hasBranch reports whether the repository has branch.
func hasBranch(t *testing.T, repo, branch string) bool {
	return gitIn(t, repo, "branch", "--list", branch) != ""
}

// Keep commits the reviewed diff on the new branch, as aiport, removes the
// worktree and leaves the branch; the Issue becomes resolving and resolves
// as an Application fix once the Application's calls succeed again.
func TestKeepLeavesTheBranchAndTheIssueResolving(t *testing.T) {
	f, _ := newFixFixture(t, editConfig)
	before := owner(t, f.repo)
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	b := m(f.fixEnded(t)["branch_fix"])
	branch, base, tree := b["branch"].(string), b["base"].(string), b["worktree"].(string)
	if b["state"] != "ready" || len(worktrees(t, f.repo)) != 2 {
		t.Fatalf("ready: %v %v", b, worktrees(t, f.repo))
	}

	if w, _ := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/keep", "", issuesManagement, false); w.Code != 403 {
		t.Fatalf("keep without CSRF: %d", w.Code)
	}
	w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/keep", "", issuesManagement, true)
	is := m(out["issue"])
	kept := m(is["branch_fix"])
	if w.Code != 200 || is["state"] != "resolving" || is["resolving_by"] != "application_fix" || kept["state"] != "kept" || kept["branch"] != branch {
		t.Fatalf("keep: %d %v", w.Code, out)
	}
	if !strings.Contains(is["application_fix"].(string), branch) {
		t.Fatalf("the Application fix does not name the branch: %s", is["application_fix"])
	}

	// One commit on the new branch, by aiport, on top of the base; the
	// worktree is gone, the branch stays; the Owner's copy is untouched.
	commit := kept["commit"].(string)
	if got := gitIn(t, f.repo, "log", "-1", "--format=%H|%an|%ae|%cn|%P", branch); got != commit+"|aiport|aiport@localhost|aiport|"+base {
		t.Fatalf("the branch's commit: %s", got)
	}
	if got := gitIn(t, f.repo, "show", branch+":config.py"); got != "MODEL = 'qwen2.5:7b'" {
		t.Fatalf("the kept config.py: %q", got)
	}
	if _, err := os.Stat(tree); err == nil || len(worktrees(t, f.repo)) != 1 {
		t.Fatalf("the worktree is still there: %v", worktrees(t, f.repo))
	}
	if after := owner(t, f.repo); after != before {
		t.Fatalf("the Owner's working copy changed:\n%s\n%s", before, after)
	}
	if gitIn(t, f.repo, "remote") != "" {
		t.Fatal("aiport added a remote")
	}
	_, diff := call(f.h, "GET", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, false)
	if !strings.Contains(diff["diff"].(string), "+MODEL = 'qwen2.5:7b'") {
		t.Fatalf("the kept diff: %v", diff)
	}
	if w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/keep", "", issuesManagement, true); w.Code != 409 || m(out["error"])["code"] != "fix_not_waiting" {
		t.Fatalf("keep twice: %d %v", w.Code, out)
	}
	if w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/discard", "", issuesManagement, true); w.Code != 409 || !hasBranch(t, f.repo, branch) {
		t.Fatalf("discard a kept fix: %d %v", w.Code, out)
	}

	// Kept with the Issue across a restart.
	f.start(t)
	f.useAgentDirs()
	if b := m(f.only(t)["branch_fix"]); b["branch"] != branch || b["commit"] != commit || b["summary"] != "1 file changed, 1 insertion(+), 1 deletion(-)" {
		t.Fatalf("after a restart: %v", b)
	}

	// The Owner merged and deployed: docs-bot's calls succeed again.
	f.hosted.set(`{"object":"list","data":[{"id":"llama3:8b"}]}`, true)
	if codes := f.chat(t, docsBotToken, "llama3:8b", 1); codes[0] != 200 {
		t.Fatalf("call after the fix: %v", codes)
	}
	f.detect()
	if is := f.only(t); is["state"] != "resolved" || m(is["resolution"])["how"] != "application_fix" {
		t.Fatalf("after a successful call: %v", is)
	}
}

// Discard removes the worktree and the branch and leaves the Issue open, so
// the Owner can try again.
func TestDiscardRemovesTheWorktreeAndTheBranch(t *testing.T) {
	f, _ := newFixFixture(t, editConfig)
	before := owner(t, f.repo)
	f.fix(t)
	b := m(f.fixEnded(t)["branch_fix"])
	branch := b["branch"].(string)
	if w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, true); w.Code != 409 || m(out["error"])["code"] != "fix_waiting" {
		t.Fatalf("Fix it again before deciding: %d %v", w.Code, out)
	}
	w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/discard", "", issuesManagement, true)
	is := m(out["issue"])
	if w.Code != 200 || is["state"] != "open" || m(is["branch_fix"])["state"] != "discarded" {
		t.Fatalf("discard: %d %v", w.Code, out)
	}
	if hasBranch(t, f.repo, branch) || len(worktrees(t, f.repo)) != 1 {
		t.Fatalf("left behind: branch %v, worktrees %v", hasBranch(t, f.repo, branch), worktrees(t, f.repo))
	}
	if _, err := os.Stat(b["worktree"].(string)); err == nil {
		t.Fatal("the worktree directory is still there")
	}
	if after := owner(t, f.repo); after != before {
		t.Fatalf("the Owner's working copy changed:\n%s\n%s", before, after)
	}
	// Try again: a fresh worktree on the same branch name.
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("Fix it again: %d %v", code, out)
	}
	if b := m(f.fixEnded(t)["branch_fix"]); b["state"] != "ready" || b["branch"] != branch {
		t.Fatalf("again: %v", b)
	}
}

// Stop, the time limit, a failed run and a run that changes nothing leave
// nothing behind: no worktree, no branch; the Issue stays open.
func TestAFixItRunThatEndsWithoutADiffLeavesNothing(t *testing.T) {
	for name, tc := range map[string]struct {
		script acpScript
		limit  time.Duration
		stop   bool
		state  string
		note   string
	}{
		"stopped":   {script: acpScript{Steps: []acpStep{editConfig.Steps[0], {Say: "Working"}, {Cancel: true}}}, stop: true, state: "stopped", note: "Stopped"},
		"timed out": {script: acpScript{Steps: []acpStep{{Say: "Thinking"}, {Ignore: true}}}, limit: 300 * time.Millisecond, state: "stopped", note: "time limit of 300ms"},
		"no change": {script: acpScript{Steps: []acpStep{{Say: "Nothing to change here."}}}, state: "empty", note: "without changing anything"},
		"failed":    {script: acpScript{Steps: []acpStep{editConfig.Steps[0]}, Stop: "refusal"}, state: "stopped", note: "refused"},
	} {
		t.Run(name, func(t *testing.T) {
			f, _ := newFixFixture(t, tc.script)
			f.h.fixLimit = tc.limit
			before := owner(t, f.repo)
			code, out := f.fix(t)
			if code != 202 {
				t.Fatalf("fix: %d %v", code, out)
			}
			if tc.stop {
				deadline := time.Now().Add(10 * time.Second)
				for {
					_, logged := call(f.h, "GET", "/api/v1/issues/"+f.id+"/run", "", issuesManagement, false)
					if l, _ := logged["log"].([]any); len(l) >= 2 {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("the agent did not start working")
					}
					time.Sleep(5 * time.Millisecond)
				}
				if w, out := call(f.h, "POST", "/api/v1/runs/"+m(out["run"])["id"].(string)+"/stop", "", issuesManagement, true); w.Code != 200 {
					t.Fatalf("stop: %d %v", w.Code, out)
				}
			}
			is := f.fixEnded(t)
			b := m(is["branch_fix"])
			if is["state"] != "open" || b["state"] != tc.state || !strings.Contains(b["note"].(string), tc.note) || b["worktree"] != nil {
				t.Fatalf("ended: %v", is)
			}
			if hasBranch(t, f.repo, b["branch"].(string)) || len(worktrees(t, f.repo)) != 1 {
				t.Fatalf("left behind: %v", worktrees(t, f.repo))
			}
			if entries, _ := os.ReadDir(f.h.worktreeRoot()); len(entries) != 0 {
				t.Fatalf("worktree directories left: %v", entries)
			}
			if after := owner(t, f.repo); after != before {
				t.Fatalf("the Owner's working copy changed:\n%s\n%s", before, after)
			}
		})
	}
}

// Fix it is hidden, with a hint, for an Application without a Coding agent
// or repository path; with both, the Issues list says which agent fixes.
func TestFixItNeedsACodingAgentAndARepository(t *testing.T) {
	f := newExplainFixture(t, newFakeWriter(t, "unused"), "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	id := f.only(t)["id"].(string)
	_, out := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if by := m(m(out["fix_by"])["docs-bot"]); by["available"] != false || !strings.Contains(by["hint"].(string), "Connect a Coding agent and the repository path of docs-bot") {
		t.Fatalf("fix_by without an agent: %v", out["fix_by"])
	}
	if w, out := call(f.h, "POST", "/api/v1/issues/"+id+"/fix", "", issuesManagement, true); w.Code != 409 || m(out["error"])["code"] != "fix_unavailable" {
		t.Fatalf("Fix it without an agent: %d %v", w.Code, out)
	}
	if w, _ := call(f.h, "PATCH", "/api/v1/origins/docs-bot", `{"agent":"claude"}`, issuesManagement, true); w.Code != 200 {
		t.Fatalf("connect: %d %s", w.Code, w.Body)
	}
	_, out = call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if by := m(m(out["fix_by"])["docs-bot"]); by["available"] != false || !strings.Contains(by["hint"].(string), "Add the repository path of docs-bot") {
		t.Fatalf("fix_by without a repository: %v", out["fix_by"])
	}

	a, _ := newFixFixture(t, editConfig)
	_, out = call(a.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if by := m(m(out["fix_by"])["docs-bot"]); by["available"] != true || by["agent"] != "claude" || by["limit_ms"] != float64(30*60*1000) {
		t.Fatalf("fix_by with an agent: %v", out["fix_by"])
	}
	five := newAgentRunFixture(t, "claude", editConfig, "agent_runs: {fix_minutes: 5}\n")
	_, out = call(five.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	if by := m(m(out["fix_by"])["docs-bot"]); by["limit_ms"] != float64(5*60*1000) {
		t.Fatalf("agent_runs.fix_minutes: %v", out["fix_by"])
	}
}

// Fix it changes the Application's code, so it is offered only for causes
// that can be there (an unknown model name), not for a failing Connection.
func TestFixItIsOnlyForCausesInTheApplicationsCode(t *testing.T) {
	f, _ := failingFixture(t)
	repo := gitRepository(t)
	gitIn(t, repo, "commit", "-q", "--allow-empty", "-m", "start")
	bin := t.TempDir()
	installFakeAgent(t, bin, "claude", editConfig)
	f.h.agentDirs = []string{bin}
	if w, _ := call(f.h, "PATCH", "/api/v1/origins/docs-bot", `{"agent":"claude","repository":"`+repo+`"}`, issuesManagement, true); w.Code != 200 {
		t.Fatalf("connect: %d %s", w.Code, w.Body)
	}
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	is := f.only(t)
	if is["cause"] != "failing_connection" {
		t.Fatalf("issue: %v", is)
	}
	_, out := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false)
	by := m(m(out["fix_by"])["docs-bot"])
	if by["available"] != true || strings.Join(strs(by["causes"]), " ") != "unknown_model" || !strings.Contains(by["cause_hint"].(string), "Connection") {
		t.Fatalf("fix_by: %v", by)
	}
	w, res := call(f.h, "POST", "/api/v1/issues/"+is["id"].(string)+"/fix", "", issuesManagement, true)
	if w.Code != 409 || m(res["error"])["code"] != "fix_not_for_cause" {
		t.Fatalf("Fix it for a failing Connection: %d %v", w.Code, res)
	}
	if len(worktrees(t, repo)) != 1 {
		t.Fatal("a worktree was made")
	}
}

// A Fix it run that reaches its time limit keeps its worktree and diff for
// review, like one whose command went unanswered.
func TestATimedOutFixKeepsItsDiffForReview(t *testing.T) {
	f, _ := newFixFixture(t, acpScript{Steps: []acpStep{editConfig.Steps[0], {Ignore: true}}})
	f.h.fixLimit = 500 * time.Millisecond
	if code, out := f.fix(t); code != 202 {
		t.Fatalf("fix: %d %v", code, out)
	}
	b := m(f.fixEnded(t)["branch_fix"])
	if b["state"] != "timed_out" || b["worktree"] == nil || !strings.Contains(b["note"].(string), "time limit") {
		t.Fatalf("branch fix: %v", b)
	}
	_, diff := call(f.h, "GET", "/api/v1/issues/"+f.id+"/fix", "", issuesManagement, false)
	if !strings.Contains(diff["diff"].(string), "+MODEL = 'qwen2.5:7b'") {
		t.Fatalf("diff: %v", diff)
	}
	if w, out := call(f.h, "POST", "/api/v1/issues/"+f.id+"/fix/keep", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("keep: %d %v", w.Code, out)
	}
}
