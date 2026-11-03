package decide

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/harelix/aiport/internal/catalog"
	"github.com/harelix/aiport/internal/refusal"
	"github.com/harelix/aiport/internal/telemetry"
)

// Catalog-change Issues. A configuration Change that takes something away
// from an Application's Catalog (a task name it had, or the binding or a
// protocol default it follows) opens an Issue of CauseCatalogChanged for
// that Application, one per Change, Subject the Change id, when the
// Application used what went away within request history (it sent the
// removed task name, or a model name that followed the moved binding or
// default). It needs no
// Owner decision (there is no aiport fix) and is delivered like every
// Issue. It resolves itself when the Application's calls succeed after the
// change (and after any failing call that sent a removed name), when the
// Application says it fixed it, or when the Change is reverted.

// CauseCatalogChanged: a Change took away something the Application's
// Catalog had. Subject is the Change id.
const CauseCatalogChanged Cause = "catalog_changed"

// CatalogNotice is one Application's Catalog change as a Catalog-change
// Issue keeps it: the change itself and what the Application can still use
// after it (its task names and the Connections its model names reach).
// Application ids, task names, model names, Connection ids and Change ids
// only.
type CatalogNotice struct {
	catalog.Record
	Available   []catalog.TaskName `json:"available"`
	Connections []string           `json:"connections"`
}

// takesAway reports whether the change took something away: a task name,
// or the binding or a protocol default the Application followed moved.
func (n CatalogNotice) takesAway() bool {
	if len(n.Removed) > 0 {
		return true
	}
	return slices.ContainsFunc(n.Routing, func(s catalog.Switch) bool { return s.Before != "" })
}

// removedNames are the task names the change removed.
func (n CatalogNotice) removedNames() []string {
	out := make([]string, 0, len(n.Removed))
	for _, t := range n.Removed {
		out = append(out, t.Name)
	}
	return out
}

// CatalogChanged takes the Catalog changes of one configuration Change.
// It runs while that Change is applied (under the configuration's lock), so
// it only queues them: they become Issues before the next read of the
// Issues, or right after in the background. The caller leaves them out
// while aiport does not adapt.
func (q *Queue) CatalogChanged(notices []CatalogNotice) {
	if len(notices) == 0 {
		return
	}
	q.inboxMu.Lock()
	q.inbox = append(q.inbox, notices...)
	q.inboxMu.Unlock()
	go q.drainCatalog()
}

// drainCatalog turns the queued Catalog changes into Issues, in the order
// the Changes were made: the inbox is taken while the lock is held, so a
// drain that took later notices (a revert) never handles them before an
// earlier drain has handled the ones before (the Change it reverts). Never
// called with the lock held.
func (q *Queue) drainCatalog() {
	q.inboxMu.Lock()
	waiting := len(q.inbox) > 0
	q.inboxMu.Unlock()
	if !waiting {
		return // a notice queued after this has its own drain
	}
	retention := time.Duration(q.live.Config().Telemetry.RetentionDays) * 24 * time.Hour
	q.mu.Lock()
	defer q.mu.Unlock()
	q.inboxMu.Lock()
	notices := q.inbox
	q.inbox = nil
	q.inboxMu.Unlock()
	for _, n := range notices {
		q.catalogNotice(n, retention)
	}
}

// usedTakenAway reports whether Application n.Application used, within
// request history before the change (retention long), what the change took
// away: it sent a removed task name, or sent a model name that is not a task
// name and went to the Connection its binding or a protocol default it
// follows pointed at before. Called with the lock held.
func (q *Queue) usedTakenAway(n CatalogNotice, retention time.Duration) bool {
	if q.history == nil {
		return false
	}
	removed := n.removedNames()
	taskName := func(name string) bool {
		return slices.ContainsFunc(n.Available, func(t catalog.TaskName) bool { return t.Name == name })
	}
	from := time.UnixMilli(n.At).Add(-retention).UnixMilli()
	for _, e := range q.history.Find(telemetry.Query{Origin: n.Application, From: from, To: n.At, InferenceOnly: true}) {
		if e.RequestedModel == nil {
			continue
		}
		if slices.Contains(removed, *e.RequestedModel) {
			return true
		}
		if taskName(*e.RequestedModel) {
			continue
		}
		for _, s := range n.Routing {
			if s.Before != "" && e.Target == s.Before && (s.Of == catalog.RoutedBinding || e.Protocol == string(s.Of)) {
				return true
			}
		}
	}
	return false
}

// catalogNotice opens the Issue for notice n when the Application used what
// it took away within request history (retention long), or, for a revert,
// resolves the Issue the reverted Change opened. Called with the lock held.
func (q *Queue) catalogNotice(n CatalogNotice, retention time.Duration) {
	if n.RevertOf != "" {
		if i := q.issueFor(n.Application, CauseCatalogChanged, n.RevertOf, ""); i >= 0 && undecided(q.issues[i].State) {
			q.resolve(&q.issues[i], Stopped, n.At)
			return // back as it was: nothing new to tell
		}
	}
	if !n.takesAway() || q.issueFor(n.Application, CauseCatalogChanged, n.ChangeID, "") >= 0 {
		return
	}
	// An Issue's aiport fix for this Application, or its revert, is told by
	// that Issue itself (it resolves, or reopens): no second Issue.
	if q.fixChange(n.Application, n.ChangeID) || q.fixChange(n.Application, n.RevertOf) {
		return
	}
	// An Application that never sent what went away has nothing to fix: the
	// Catalog change is recorded and readable, but no Issue opens.
	if !q.usedTakenAway(n, retention) {
		return
	}
	notice := n
	q.issues = append(q.issues, Issue{
		ID: newID(), Application: n.Application, Cause: CauseCatalogChanged, Subject: n.ChangeID, State: IssueOpen,
		Evidence: Evidence{}, Catalog: &notice, SinceMS: n.At, CreatedAt: n.At, UpdatedAt: n.At,
	})
	is := &q.issues[len(q.issues)-1]
	is.Summary = catalogSummary(*is)
	q.saveIssue(is)
}

// fixChange reports whether Change id applied the aiport fix of one of
// Application app's Issues. Called with the lock held.
func (q *Queue) fixChange(app, id string) bool {
	if id == "" {
		return false
	}
	return slices.ContainsFunc(q.issues, func(is Issue) bool {
		return is.Application == app && is.Fix != nil && is.Fix.ChangeID == id
	})
}

// sentRemoved reports whether call e failed and sent one of names.
func sentRemoved(e telemetry.Event, names []string) bool {
	failed := telemetry.IsError(e.Outcome) || e.Status >= 400
	return failed && e.RequestedModel != nil && slices.Contains(names, *e.RequestedModel)
}

// catalogResolutions resolves open Catalog-change Issues whose
// Application's calls succeeded after the change and after its last failing
// call that sent a removed name. Called with the lock held.
func (q *Queue) catalogResolutions(at time.Time) {
	now := at.UnixMilli()
	for i := range q.issues {
		is := &q.issues[i]
		if is.Cause != CauseCatalogChanged || !undecided(is.State) || is.Catalog == nil {
			continue
		}
		names := is.Catalog.removedNames()
		after, success := is.CreatedAt, int64(0)
		for _, e := range q.history.Find(telemetry.Query{Origin: is.Application, From: is.CreatedAt, To: now, InferenceOnly: true}) {
			switch {
			case sentRemoved(e, names):
				after = max(after, e.StartedAt)
			case telemetry.Succeeded(e.Outcome):
				success = max(success, e.StartedAt)
			}
		}
		if success > after {
			q.resolve(is, Stopped, now)
		}
	}
}

func catalogSummary(is Issue) string {
	n := is.Catalog
	parts := []string{}
	if names := n.removedNames(); len(names) > 0 {
		parts = append(parts, "removed task name "+quoteAll(names))
	}
	for _, s := range n.Routing {
		if s.Before != "" {
			parts = append(parts, routingOf(s)+" moved from Connection "+strconv.Quote(s.Before)+" to "+connectionOrNone(s.After))
		}
	}
	return "A Change (" + n.ChangeID + ") changed what " + is.Application + " can use: " + strings.Join(parts, "; ") + ". The Application was told; nothing waits for the Owner."
}

// catalogFix is the Application fix of a Catalog-change Issue: what
// changed, what replaced it, what the Application can still use, and how
// to confirm.
func catalogFix(is Issue) string {
	n, app := is.Catalog, is.Application
	if n == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("aiport Issue " + is.ID + ": the Owner changed what the Application " + app + " can use through aiport (Change " + n.ChangeID + " at " + utc(n.At) + ").\n\n")
	b.WriteString("What changed:\n")
	for _, t := range n.Removed {
		b.WriteString("- Task name " + strconv.Quote(t.Name) + " was removed (it went to model " + strconv.Quote(t.Model) + " on Connection " + strconv.Quote(t.Connection) + "); nothing replaces it.\n")
	}
	for _, m := range n.Moved {
		b.WriteString("- Task name " + strconv.Quote(m.Name) + " now goes to model " + strconv.Quote(m.To.Model) + " on Connection " + strconv.Quote(m.To.Connection) + " (was " + strconv.Quote(m.From.Model) + " on " + strconv.Quote(m.From.Connection) + ").\n")
	}
	for _, s := range n.Routing {
		b.WriteString("- " + capitalize(routingOf(s)) + " now goes to " + connectionOrNone(s.After))
		if s.Before != "" {
			b.WriteString(" (was Connection " + strconv.Quote(s.Before) + ")")
		}
		b.WriteString(".\n")
	}
	b.WriteString("\nWhat " + app + " can still use: ")
	if len(n.Available) > 0 {
		names := make([]string, 0, len(n.Available))
		for _, t := range n.Available {
			names = append(names, t.Name)
		}
		b.WriteString("task names " + quoteAll(names))
	} else {
		b.WriteString("no task names")
	}
	if len(n.Connections) > 0 {
		word := "Connection "
		if len(n.Connections) > 1 {
			word = "Connections "
		}
		b.WriteString(", and the models " + word + strings.Join(n.Connections, ", ") + " list")
	}
	b.WriteString(". Read the full list with GET /v1/models (" + app + "'s own token), and the Catalog with its recent changes at " + refusal.CatalogPath + " or the MCP tool aiport_catalog.\n\n")

	b.WriteString("What to change in " + app + ": stop sending ")
	if names := n.removedNames(); len(names) > 0 {
		b.WriteString(quoteAll(names) + "; ")
	} else {
		b.WriteString("model names the new routing does not serve; ")
	}
	b.WriteString("send one of the names above instead. If " + app + " needs a removed task name back, file a Change request of kind " + string(KindApplicationAlias) + "; the Owner decides.\n\n")
	b.WriteString("How to confirm: make one call through aiport with " + app + "'s own token and check it answers 200. aiport marks this Issue resolved by itself when calls from " + app + " succeed again; or report the fix with the MCP tool aiport_issue_fixed (POST " + refusal.IssuesPath + "/" + is.ID + "/fixed).")
	return b.String()
}

func routingOf(s catalog.Switch) string {
	if s.Of == catalog.RoutedBinding {
		return "the binding (every model name without an Alias)"
	}
	return "the " + string(s.Of) + " protocol default (model names without an Alias)"
}

func connectionOrNone(id string) string {
	if id == "" {
		return "no Connection"
	}
	return "Connection " + strconv.Quote(id)
}

func quoteAll(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = strconv.Quote(n)
	}
	return strings.Join(q, ", ")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
