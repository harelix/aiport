package decide

import (
	"errors"
	"time"

	"aiport/internal/telemetry"
)

// Issues as the Application's Coding agent sees them (GET /v1/aiport/issues
// and the MCP tool aiport_my_issues): only that Application's own Issues,
// with what is wrong, the evidence, the Application fix and where the
// aiport fix stands, and never the Owner's notes or other Applications'
// data. The agent can say it fixed an Issue on its side; aiport then waits
// for real calls to confirm it.

// ErrNotOpen: the Issue is not open, so it cannot be marked fixed.
var ErrNotOpen = errors.New("this Issue is not open: it is already being resolved, resolved or dismissed")

// AgentFix is where an Issue's aiport fix stands, for the Application.
type AgentFix struct {
	Kind     Kind     `json:"kind"`
	Proposal Proposal `json:"proposal"`
	State    FixState `json:"state"`
	ChangeID string   `json:"change_id,omitempty"`
	By       string   `json:"by,omitempty"`
	At       int64    `json:"at_ms,omitempty"`
}

// AgentIssue is an Issue as its Application's Coding agent reads it.
type AgentIssue struct {
	ID          string     `json:"id"`
	Application string     `json:"application"`
	Cause       Cause      `json:"cause"`
	Subject     string     `json:"subject"`
	State       IssueState `json:"state"`
	Summary     string     `json:"summary"`
	Evidence    Evidence   `json:"evidence"`
	Diagnosis   *Diagnosis `json:"diagnosis,omitempty"`
	// ApplicationFix is what to change in the Application (ApplicationFix).
	ApplicationFix string `json:"application_fix"`
	// AiportFix is the fix in aiport's configuration, if one is offered or
	// was made: proposed (waiting for the Owner), applied, rejected or
	// reverted.
	AiportFix   *AgentFix   `json:"aiport_fix,omitempty"`
	ResolvingBy How         `json:"resolving_by,omitempty"`
	Resolution  *Resolution `json:"resolution,omitempty"`
	SinceMS     int64       `json:"since_ms"`
	CreatedAt   int64       `json:"created_at_ms"`
	UpdatedAt   int64       `json:"updated_at_ms"`
}

// ForAgent returns the Issue as its Application's Coding agent reads it.
func (is Issue) ForAgent() AgentIssue {
	a := AgentIssue{
		ID: is.ID, Application: is.Application, Cause: is.Cause, Subject: is.Subject, State: is.State,
		Summary: is.Summary, Evidence: is.Evidence, Diagnosis: is.Diagnosis, ApplicationFix: is.ApplicationFix,
		ResolvingBy: is.ResolvingBy, Resolution: is.Resolution, SinceMS: is.SinceMS, CreatedAt: is.CreatedAt, UpdatedAt: is.UpdatedAt,
	}
	if f := is.Fix; f != nil {
		a.AiportFix = &AgentFix{Kind: f.Kind, Proposal: f.Proposal, State: f.State, ChangeID: f.ChangeID, By: f.By, At: f.At}
	}
	return a
}

// IssuesOf returns Application app's Issues (open, being resolved, and
// recently resolved or dismissed), newest first, as its agent reads them.
func (q *Queue) IssuesOf(app string) []AgentIssue {
	out := []AgentIssue{}
	if app == "" {
		return out
	}
	for _, is := range q.Issues(IssueFilter{Application: app}) {
		out = append(out, is.ForAgent())
	}
	return out
}

// IssueOf returns Issue id if it belongs to Application app; another
// Application's Issue is ErrNotFound.
func (q *Queue) IssueOf(app, id string) (AgentIssue, error) {
	is, ok := q.Issue(id)
	if !ok || app == "" || is.Application != app {
		return AgentIssue{}, ErrNotFound
	}
	return is.ForAgent(), nil
}

// OpenIssues counts Application app's Issues that are open or being
// resolved.
func (q *Queue) OpenIssues(app string) int {
	return len(q.Issues(IssueFilter{Application: app, States: []IssueState{IssueOpen, IssueResolving}}))
}

// HintIssue is the open (or resolving) Issue a failing call from app names
// in its X-Aiport-Hint: the one whose failures the call is one of (e as
// request history will record it: for an unknown model, a provider 404 for
// the Issue's model; for a failing Connection, a call to that Connection
// that failed against it). Any other failing call names none. It reads the
// snapshot published after each change to the Issues and never takes the
// queue's lock, so it is safe on the forwarding path.
func (q *Queue) HintIssue(app string, e telemetry.Event) (string, bool) {
	snap := q.hints.Load()
	if snap == nil || app == "" {
		return "", false
	}
	for _, is := range (*snap)[app] {
		if r, ok := causes[is.cause]; ok && r.counts(e, is.subject) {
			return is.id, true
		}
	}
	return "", false
}

// hintIssue is an open or resolving Issue as HintIssue reads it.
type hintIssue struct {
	id, subject string
	cause       Cause
}

// publishHints swaps in a fresh snapshot of the open and resolving Issues
// per Application for HintIssue. Called with the lock held after each change
// to the Issues.
func (q *Queue) publishHints() {
	snap := map[string][]hintIssue{}
	for i := len(q.issues) - 1; i >= 0; i-- {
		is := q.issues[i]
		if is.State == IssueOpen || is.State == IssueResolving {
			snap[is.Application] = append(snap[is.Application], hintIssue{id: is.ID, subject: is.Subject, cause: is.Cause})
		}
	}
	q.hints.Store(&snap)
}

// FixedByApplication records that Application app fixed its open Issue id
// on its side: the Issue becomes resolving by the Application fix, so the
// Owner sees it is being handled. An aiport fix waiting for the Owner keeps
// waiting: the Owner can still approve it. aiport resolves the Issue only
// when the Application's calls succeed again.
func (q *Queue) FixedByApplication(app, id string, at time.Time) (AgentIssue, error) {
	if !q.On() {
		return AgentIssue{}, ErrOff
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	i := q.issueIndex(id)
	if i < 0 || app == "" || q.issues[i].Application != app {
		return AgentIssue{}, ErrNotFound
	}
	is := &q.issues[i]
	if is.State == IssueResolving && is.ResolvingBy == ByApplicationFix {
		return is.ForAgent(), nil // already said
	}
	if is.State != IssueOpen {
		return AgentIssue{}, ErrNotOpen
	}
	now := at.UnixMilli()
	is.State, is.ResolvingBy, is.SinceMS, is.UpdatedAt = IssueResolving, ByApplicationFix, now, now
	q.saveIssue(is)
	return is.ForAgent(), nil
}
