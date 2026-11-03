package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/harelix/aiport/internal/agentrun"
	"github.com/harelix/aiport/internal/agents"
	"github.com/harelix/aiport/internal/decide"
	"github.com/harelix/aiport/internal/explain"
	"github.com/harelix/aiport/internal/worktree"
)

// Fix it (harelix/aiport#25): on the Owner's click aiport makes a git
// worktree of the Application's repository on a new branch named for the
// Issue (internal/worktree) and runs the Application's Coding agent there.
// Edits inside the worktree are allowed; a command inside it waits for the
// Owner's answer (commandRule, harelix/aiport#26); anything outside the
// worktree is denied. The run ends with a diff the Owner reviews:
//
//   - Keep: aiport commits the diff on the new branch, as one commit
//     authored by aiport, and removes the worktree; the branch stays, and the
//     Issue becomes resolving by the Application fix (resolved once the
//     Application's calls succeed again). aiport never commits to the
//     Owner's branch, merges or pushes: merging is the Owner's.
//   - Discard: the worktree and the branch are removed; the Issue stays open.
//
// Fix it is offered only for causes that can be in the Application's code
// (FixCauses). Stop by the Owner and a failed run remove the worktree and
// the branch. A run stopped because a command request went unanswered, or
// that reached its time limit (agent_runs.fix_minutes), keeps what it
// changed so far for review like a finished run's diff
// (decide.BranchStoppedWaiting, BranchTimedOut). The branch and the diff summary
// are kept with the Issue (decide.BranchFix). A run still recorded as going
// when aiport starts is marked stopped and its worktree removed
// (stopLeftFixes).

// fixUnavailable says why Fix it is unavailable for Application app, in a
// hint the dashboard shows instead of the button; "" when it is available.
func (h *Handler) fixUnavailable(app string, ca codingAgent) string {
	if h.decisions == nil || !h.decisions.On() {
		return "Fix it is unavailable while adapting is turned off in the configuration."
	}
	switch {
	case ca.ID == "" && ca.Path == "":
		return "Connect a Coding agent and the repository path of " + app + " (Integrate, the Application's row) to use Fix it."
	case ca.ID == "":
		return "Choose the Coding agent of " + app + " (Integrate, the Application's row) to use Fix it."
	case ca.Path == "":
		return "Add the repository path of " + app + " (Integrate, the Application's row) to use Fix it."
	case !ca.OK:
		return "Fix it needs the Coding agent " + ca.ID + " ready (Settings, Coding agents) and " + ca.Path + " to be the top of a git repository."
	}
	return ""
}

// FixCauses are the Issue causes Fix it is offered for: those whose cause
// can be in the Application's code.
var FixCauses = []decide.Cause{decide.CauseUnknownModel}

// FixCauseHint says why Fix it is not offered for other causes.
const FixCauseHint = "Fix it changes the Application's code, so it is offered when the cause can be there: a model name the Application asks for that no Connection serves. " +
	"A failing Connection is not in the Application's code: bind the Application to a working Connection (the aiport fix) instead."

func fixesCause(c decide.Cause) bool { return slices.Contains(FixCauses, c) }

// fixBy says, per Application, whether Fix it is available, with which
// agent, time limit and for which causes, or the hint why not (GET
// /api/v1/issues), from all, every Application's Coding agent.
func (h *Handler) fixBy(all map[string]codingAgent) map[string]any {
	out := map[string]any{}
	for app, ca := range all {
		if hint := h.fixUnavailable(app, ca); hint != "" {
			out[app] = map[string]any{"available": false, "hint": hint}
			continue
		}
		out[app] = map[string]any{"available": true, "agent": ca.Agent.ID, "agent_name": ca.Agent.Name, "model_calls": h.modelCalls(ca.Agent),
			"limit_ms": h.fixTimeLimit().Milliseconds(), "causes": FixCauses, "cause_hint": FixCauseHint}
	}
	return out
}

func (h *Handler) fixTimeLimit() time.Duration {
	if h.fixLimit > 0 {
		return h.fixLimit
	}
	return h.snapshot().AgentRuns.FixLimit()
}

func (h *Handler) worktreeRoot() string {
	if h.worktrees != "" {
		return h.worktrees
	}
	return filepath.Join(os.TempDir(), "aiport-worktrees")
}

// fixPrompt is what an agent fixing is is told: aiport's evidence
// (explain.Prompt: never tokens or keys; of what an Application wrote,
// only a model name that looks like a model id),
// the saved explanation if any, and the task.
func fixPrompt(is decide.Issue, branch string) string {
	s := "aiport, the local AI gateway this repository's Application calls, keeps seeing the problem below with the Application's model calls.\n\n" + explain.Prompt(is)
	if e := is.Explanation; e != nil && e.Text != "" {
		s += "\nRoot cause found earlier:\n" + e.Text + "\n"
	}
	return s + "\nFix it in this repository's code. You are in a git worktree on the new branch " + branch + ", made for this fix: " +
		"edit files only inside this directory (aiport allows that). Before each command you run here (tests, builds), aiport asks the Owner, who may deny it; " +
		"aiport denies anything outside this directory. " +
		"Do not commit, merge or push: aiport shows your diff to the Owner, who keeps or discards it. " +
		"End with a short summary of what you changed and why, in plain text."
}

// discardTree removes worktree t and its branch. A worktree recorded
// outside aiport's worktrees directory is left alone and logged.
func (h *Handler) discardTree(t worktree.Tree) error {
	err := t.Discard()
	if errors.Is(err, worktree.ErrNotOwned) {
		log.Printf("aiport: left %s alone: it is not in aiport's worktrees directory %s", t.Dir, t.Root)
	}
	return err
}

// treeOf is the worktree of branch fix b.
func (h *Handler) treeOf(b decide.BranchFix) worktree.Tree {
	return worktree.Tree{Repo: b.Repository, Dir: b.Worktree, Branch: b.Branch, Base: b.Base, Root: h.worktreeRoot()}
}

// refused is why a request was refused: a 409 with code and message.
type refused struct{ code, message string }

// startFix starts Fix it for Issue is: the worktree, the branch fix
// recorded as running, and the run. It returns the run, or why not.
func (h *Handler) startFix(is decide.Issue) (*agentrun.Run, *refused) {
	h.fixMu.Lock()
	defer h.fixMu.Unlock()
	if is.State != decide.IssueOpen {
		return nil, &refused{"issue_not_open", "Fix it works on an open Issue; this one is " + string(is.State)}
	}
	if !fixesCause(is.Cause) {
		return nil, &refused{"fix_not_for_cause", FixCauseHint}
	}
	if b := is.BranchFix; b != nil && (b.State == decide.BranchRunning || b.Reviewable()) {
		return nil, &refused{"fix_waiting", "This Issue's last fix is still going or waiting for the Owner: keep or discard it first"}
	}
	ca := h.codingAgentOf(is.Application)
	if hint := h.fixUnavailable(is.Application, ca); hint != "" {
		return nil, &refused{"fix_unavailable", hint}
	}
	a := ca.Agent
	if _, going := h.runner().Active(is.Application); going {
		return nil, &refused{"agent_run_running", agentrun.ErrBusy.Error()}
	}
	t, err := worktree.Create(ca.Repo, h.worktreeRoot(), is.ID)
	if err != nil {
		return nil, &refused{"worktree_failed", err.Error()}
	}
	calls := h.modelCalls(a)
	env, token, revoke := h.runEnv(a, calls)
	b := decide.BranchFix{Branch: t.Branch, Base: t.Base, Worktree: t.Dir, Repository: t.Repo, Agent: a.ID, State: decide.BranchRunning, StartedAt: h.clock().UnixMilli()}
	allowed := &allowances{}
	parent := h.fixContext
	if parent == nil {
		parent = context.Background()
	}
	run, err := h.runner().Start(parent, agentrun.Spec{
		Application: is.Application, Issue: is.ID, Mode: agentrun.Fix, Agent: a.ID, AgentName: a.Name,
		Program: agentrun.Program{Path: a.ACP.Path, Args: a.ACP.Args}, Dir: t.Dir,
		Prompt: fixPrompt(is, t.Branch), Env: env, Secret: token, ModelCalls: calls, Limit: h.fixTimeLimit(), Warnings: agents.SettingsWarnings(t.Dir),
		Policy: fixRules(t.Dir, h.commandRule(is.Application, is.ID, allowed)), OnEdit: allowed.clear,
		OnEnd: func(run *agentrun.Run) {
			revoke()
			h.fixEnded(is.ID, b, run.Info())
		},
	})
	if err != nil {
		revoke()
		_ = h.discardTree(t)
		if errors.Is(err, agentrun.ErrBusy) {
			return nil, &refused{"agent_run_running", err.Error()}
		}
		return nil, &refused{"fix_failed", "aiport could not start the agent"}
	}
	b.Run = run.Info().ID
	if _, err := h.decisions.RecordBranch(is.ID, b, h.clock()); err != nil {
		run.Stop()
		return nil, &refused{"fix_failed", "aiport could not record the fix on the Issue"}
	}
	return run, nil
}

// fixEnded records how a Fix it run ended: a diff for the Owner, or nothing
// left behind.
func (h *Handler) fixEnded(issue string, b decide.BranchFix, info agentrun.Info) {
	h.fixMu.Lock() // after startFix has recorded the run
	defer h.fixMu.Unlock()
	b.Run = info.ID
	t := h.treeOf(b)
	discard := func(state decide.BranchState, note string) {
		if err := h.discardTree(t); err != nil {
			note += " " + err.Error()
		}
		b.State, b.Note, b.Worktree = state, strings.TrimSpace(note), ""
	}
	switch info.State {
	case agentrun.Done:
		s, err := t.Summary()
		switch {
		case err != nil:
			discard(decide.BranchStopped, err.Error()+"; the worktree was removed.")
		case s.Empty():
			discard(decide.BranchEmpty, "The agent ended its turn without changing anything.")
		default:
			b.State, b.Summary, b.Files = decide.BranchReady, s.Text, s.Files
		}
	case agentrun.Stopped:
		if !h.stoppedUnanswered(info.ID) {
			discard(decide.BranchStopped, "Stopped; the worktree and the branch were removed.")
			break
		}
		why := "A command request was unanswered for " + h.approvalTimeout().String() + ", so aiport stopped the run"
		s, err := t.Summary()
		switch {
		case err != nil:
			discard(decide.BranchStopped, why+"; "+err.Error()+"; the worktree was removed.")
		case s.Empty():
			discard(decide.BranchStopped, why+" before it changed anything; the worktree and the branch were removed.")
		default:
			b.State, b.Summary, b.Files = decide.BranchStoppedWaiting, s.Text, s.Files
			b.Note = why + "; what it changed so far waits for your review."
		}
	case agentrun.TimedOut:
		why := "Reached its time limit of " + time.Duration(info.LimitMS*int64(time.Millisecond)).String()
		s, err := t.Summary()
		switch {
		case err != nil:
			discard(decide.BranchStopped, why+"; "+err.Error()+"; the worktree was removed.")
		case s.Empty():
			discard(decide.BranchStopped, why+" before it changed anything; the worktree and the branch were removed.")
		default:
			b.State, b.Summary, b.Files = decide.BranchTimedOut, s.Text, s.Files
			b.Note = why + "; what it changed so far waits for your review."
		}
	default:
		discard(decide.BranchStopped, "The run failed: "+info.Error+"; the worktree and the branch were removed.")
	}
	if h.decisions != nil {
		_, _ = h.decisions.RecordBranch(issue, b, h.clock())
	}
}

// stopLeftFixes marks stopped every branch fix still recorded as running,
// with no run going: aiport stopped (or crashed) while it ran. Its worktree
// and branch are removed, as for a stopped run.
func (h *Handler) stopLeftFixes() {
	if h.worktrees != "" && len(h.runner().Running()) == 0 {
		// Explain's throwaway worktrees of runs aiport did not finish.
		var repos []string
		for _, o := range h.snapshot().Origins {
			repos = append(repos, o.Repository)
		}
		if n := worktree.RemoveDetached(h.worktreeRoot(), repos); n > 0 {
			log.Printf("aiport: removed %d Explain worktrees left from before the restart", n)
		}
	}
	if h.decisions == nil || !h.decisions.On() {
		return
	}
	h.fixMu.Lock()
	defer h.fixMu.Unlock()
	for _, is := range h.decisions.Issues(decide.IssueFilter{}) {
		b := is.BranchFix
		if b == nil || b.State != decide.BranchRunning {
			continue
		}
		if run, going := h.runner().Active(is.Application); going && run.Info().ID == b.Run {
			continue
		}
		note := "aiport stopped while the run was going; the worktree and the branch were removed."
		if err := h.discardTree(h.treeOf(*b)); err != nil {
			note += " " + err.Error()
		}
		fixed := *b
		fixed.State, fixed.Worktree, fixed.Note = decide.BranchStopped, "", note
		_, _ = h.decisions.RecordBranch(is.ID, fixed, h.clock())
	}
}

// fixRoute serves Fix it to the Owner:
//
//	POST /api/v1/issues/{id}/fix           202 {run, issue}: starts Fix it
//	GET  /api/v1/issues/{id}/fix           {branch_fix, diff, truncated}: the diff to review (or the kept commit's)
//	POST /api/v1/issues/{id}/fix/keep      {issue}: commits the diff on the branch, removes the worktree; the Issue becomes resolving
//	POST /api/v1/issues/{id}/fix/discard   {issue}: removes the worktree and the branch; the Issue stays open
//
// The run's log is GET /api/v1/issues/{id}/run, Stop is POST
// /api/v1/runs/{id}/stop. The caller has already verified management
// access.
func (h *Handler) fixRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/issues/")
	if !ok {
		return false
	}
	id, action, _ := strings.Cut(rest, "/")
	switch action {
	case "fix", "fix/keep", "fix/discard":
	default:
		return false
	}
	if id == "" {
		return false
	}
	if action == "fix" && r.Method == http.MethodGet {
		h.showFix(w, id, fail)
		return true
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		if action == "fix" {
			w.Header().Set("Allow", "GET, POST")
		}
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard for Fix it")
		return true
	}
	if h.decisions == nil || !h.decisions.On() {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return true
	}
	is, found := h.decisions.Issue(id)
	if !found {
		fail(404, "issue_not_found", "No Issue with this id")
		return true
	}
	switch action {
	case "fix":
		run, no := h.startFix(is)
		if no != nil {
			fail(409, no.code, no.message)
			return true
		}
		is, _ = h.decisions.Issue(id)
		writeJSON(w, 202, map[string]any{"run": run.Info(), "issue": is})
	case "fix/keep":
		h.keepFix(w, is, fail)
	case "fix/discard":
		h.discardFix(w, is, fail)
	}
	return true
}

// readyFix is Issue is's branch fix waiting for the Owner, or false after
// failing the request.
func readyFix(is decide.Issue, fail func(int, string, string)) (decide.BranchFix, bool) {
	b := is.BranchFix
	switch {
	case b == nil:
		fail(409, "no_fix", "This Issue has no fix from Fix it")
	case b.State == decide.BranchRunning:
		fail(409, "fix_running", "The Fix it run is still going: wait for its diff or stop it")
	case !b.Reviewable():
		fail(409, "fix_not_waiting", "This Issue's fix is "+string(b.State)+": nothing waits to be kept or discarded")
	default:
		return *b, true
	}
	return decide.BranchFix{}, false
}

func (h *Handler) keepFix(w http.ResponseWriter, is decide.Issue, fail func(int, string, string)) {
	h.fixMu.Lock()
	defer h.fixMu.Unlock()
	is, _ = h.decisions.Issue(is.ID)
	b, ok := readyFix(is, fail)
	if !ok {
		return
	}
	message := "Fix aiport Issue " + is.ID + "\n\n" + strings.TrimSpace(is.Summary) + "\n\nChanged by the Coding agent " + b.Agent + " in an aiport Fix it run (" + b.Run + "), kept by the Owner after reviewing the diff."
	commit, err := h.treeOf(b).Keep(message)
	if commit == "" {
		fail(409, "keep_failed", err.Error())
		return
	}
	b.State, b.Commit, b.Worktree, b.Note = decide.BranchKept, commit, "", ""
	if err != nil {
		b.Note = err.Error()
	}
	fresh, err := h.decisions.RecordBranch(is.ID, b, h.clock())
	if err != nil {
		fail(500, "issue_update_failed", "The Issue could not be updated")
		return
	}
	writeJSON(w, 200, map[string]any{"issue": fresh})
}

func (h *Handler) discardFix(w http.ResponseWriter, is decide.Issue, fail func(int, string, string)) {
	h.fixMu.Lock()
	defer h.fixMu.Unlock()
	is, _ = h.decisions.Issue(is.ID)
	b, ok := readyFix(is, fail)
	if !ok {
		return
	}
	if err := h.discardTree(h.treeOf(b)); err != nil {
		fail(409, "discard_failed", err.Error())
		return
	}
	b.State, b.Worktree, b.Note = decide.BranchDiscarded, "", ""
	fresh, err := h.decisions.RecordBranch(is.ID, b, h.clock())
	if err != nil {
		fail(500, "issue_update_failed", "The Issue could not be updated")
		return
	}
	writeJSON(w, 200, map[string]any{"issue": fresh})
}

// showFix answers GET /api/v1/issues/{id}/fix: the branch fix and its diff
// (the worktree's while it waits for the Owner, the kept commit's after).
func (h *Handler) showFix(w http.ResponseWriter, id string, fail func(int, string, string)) {
	if h.decisions == nil {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return
	}
	is, found := h.decisions.Issue(id)
	if !found {
		fail(404, "issue_not_found", "No Issue with this id")
		return
	}
	b := is.BranchFix
	if b == nil {
		fail(404, "no_fix", "This Issue has no fix from Fix it")
		return
	}
	diff, truncated := "", false
	var err error
	switch {
	case b.Reviewable():
		h.fixMu.Lock()
		diff, truncated, err = h.treeOf(*b).Diff()
		h.fixMu.Unlock()
	case b.State == decide.BranchKept && b.Commit != "":
		diff, truncated, err = worktree.KeptDiff(b.Repository, b.Base, b.Commit)
	}
	out := map[string]any{"branch_fix": b, "diff": diff, "truncated": truncated}
	if err != nil {
		out["diff_error"] = err.Error()
	}
	writeJSON(w, 200, out)
}
