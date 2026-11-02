package decide

import (
	"strconv"
	"time"
)

// Branch fixes. Fix it (spec harelix/aiport#21) runs the Application's
// Coding agent in its own git worktree of the Application's repository, on
// a new branch named for the Issue. What it changed is kept with the Issue:
// the branch, the diff summary and, once the Owner keeps it, the commit
// aiport made on that branch. aiport never commits to the Owner's branch,
// merges or pushes.

// BranchState is where a branch fix stands.
type BranchState string

const (
	BranchRunning   BranchState = "running"   // the agent is working in the worktree
	BranchReady     BranchState = "ready"     // the run ended with a diff for the Owner to review
	BranchEmpty     BranchState = "empty"     // the run ended without changing anything; the worktree is gone
	BranchKept      BranchState = "kept"      // the Owner kept it: one commit on the branch, the worktree is gone
	BranchDiscarded BranchState = "discarded" // the Owner discarded it: the worktree and the branch are gone
	BranchStopped   BranchState = "stopped"   // stopped, timed out or failed: the worktree and the branch are gone
)

// MaxBranchFiles bounds the files a branch fix's summary lists.
const MaxBranchFiles = 50

// BranchFix is an Issue's latest Fix it run and what it changed.
type BranchFix struct {
	// Branch is the new branch, named for the Issue (aiport/fix-<id>).
	Branch string `json:"branch"`
	// Base is the commit the branch starts from: the Owner's HEAD when the
	// run started.
	Base string `json:"base"`
	// Worktree is the run's own working directory, while it exists.
	Worktree string `json:"worktree,omitempty"`
	// Repository is the Application's repository the worktree belongs to.
	Repository string      `json:"repository"`
	Agent      string      `json:"agent"`
	Run        string      `json:"run"`
	State      BranchState `json:"state"`
	// Summary is the diff summary ("2 files changed, 3 insertions(+), 1
	// deletion(-)") and Files the paths it changed (at most MaxBranchFiles).
	Summary string   `json:"summary,omitempty"`
	Files   []string `json:"files,omitempty"`
	// Commit is the commit aiport made on Branch when the Owner kept it.
	Commit string `json:"commit,omitempty"`
	// Note says in a sentence why a run left nothing (stopped, timed out,
	// failed).
	Note      string `json:"note,omitempty"`
	StartedAt int64  `json:"started_at_ms"`
	At        int64  `json:"at_ms"`
}

// RecordBranch saves b as Issue id's branch fix, replacing any earlier one.
// Keeping it moves an open Issue to resolving by the Application fix:
// aiport resolves it, as an Application fix, once the Application's calls
// succeed again.
func (q *Queue) RecordBranch(id string, b BranchFix, at time.Time) (Issue, error) {
	if !q.On() {
		return Issue{}, ErrOff
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	i := q.issueIndex(id)
	if i < 0 {
		return Issue{}, ErrNotFound
	}
	is := &q.issues[i]
	now := at.UnixMilli()
	if len(b.Files) > MaxBranchFiles {
		b.Files = b.Files[:MaxBranchFiles]
	}
	b.At = now
	is.BranchFix = &b
	if b.State == BranchKept && is.State == IssueOpen {
		is.State, is.ResolvingBy, is.SinceMS = IssueResolving, ByApplicationFix, now
	}
	is.UpdatedAt = now
	q.saveIssue(is)
	return *is, nil
}

// branchText is the Application fix's paragraph about a kept branch fix.
func branchText(b BranchFix) string {
	s := "A fix is waiting on branch " + strconv.Quote(b.Branch) + " of the Application's repository"
	if b.Commit != "" {
		s += " (commit " + b.Commit[:min(12, len(b.Commit))] + ")"
	}
	s += ", made by the Coding agent " + strconv.Quote(b.Agent) + " and kept by the Owner"
	if b.Summary != "" {
		s += ": " + b.Summary
	}
	return s + ". aiport does not merge it: merge it, deploy, and aiport confirms the fix from the next calls."
}
