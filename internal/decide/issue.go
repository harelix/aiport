package decide

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Issues. An Issue is a problem aiport keeps seeing for one Application,
// keyed by Application, Cause and Subject (for an unknown model: the
// requested model), with its evidence, a diagnosis and two resolutions: the
// aiport fix, offered as a Suggestion in this queue (so the Owner decides
// it like any other and accepting it applies a revertable Change through
// config.Live), and the Application fix, text for the Application's Coding
// agent (ApplicationFix, the one place that writes it).
//
// States: open → resolving (a fix was applied, or the Application said it
// fixed it) → resolved (its calls succeed again, or the error stopped for a
// whole window) or dismissed (the Owner dismissed it; quiet for QuietFor
// unless the evidence doubles). A resolved Issue reopens when the error
// sticks again. Resolution records how it was resolved.
//
// Issues hold Application ids, the cause, model names, Connection ids,
// counts, request ids and fixed advice text; never prompts, responses,
// provider messages, keys or tokens.

// IssueState is where an Issue stands.
type IssueState string

const (
	IssueOpen      IssueState = "open"      // the error sticks; the Owner or the Application can resolve it
	IssueResolving IssueState = "resolving" // a fix was made; aiport waits for calls to succeed
	IssueResolved  IssueState = "resolved"  // calls succeed again (or the error stopped)
	IssueDismissed IssueState = "dismissed" // the Owner dismissed it
)

// Cause is what an Issue is about. Each Cause has a detector that produces
// it; Subject says which instance (for CauseUnknownModel, the model name).
type Cause string

const CauseUnknownModel Cause = "unknown_model"

// How is how an Issue was resolved, or is expected to resolve while it is
// resolving.
type How string

const (
	ByAiportFix      How = "aiport_fix"      // the aiport fix was applied as a Change
	ByApplicationFix How = "application_fix" // the Application changed its calls
	Stopped          How = "stopped"         // the calls succeeded again, or stopped, without a fix
)

// FixState is where an Issue's aiport fix stands.
type FixState string

const (
	FixProposed FixState = "proposed" // a pending Suggestion waits for the Owner
	FixApplied  FixState = "applied"  // accepted: a Change in the configuration
	FixRejected FixState = "rejected" // the Owner rejected it (offered again while the error sticks)
	FixReverted FixState = "reverted" // the Change was reverted
)

// Answers to "which enabled Connections serve this model?" (proxy.Serving).
const (
	ServedByOne     = "one"     // certain: exactly one Connection serves it
	ServedByNone    = "none"    // no Connection lists it
	ServedBySeveral = "several" // ambiguous
	ServedUnknown   = "unknown" // a listing failed: no certain answer
)

// Diagnosis is what aiport found out about an Issue's cause. For an unknown
// model: which enabled Connections of Protocol list the model (Serving),
// which could not be listed (Unknown), and the models Connections of that
// protocol list (Available, bounded), from the free model listings.
type Diagnosis struct {
	Protocol  string   `json:"protocol"`
	Answer    string   `json:"answer"`
	Serving   []string `json:"serving"`
	Unknown   []string `json:"unknown"`
	Available []string `json:"available,omitempty"`
	At        int64    `json:"at_ms"`
}

// ModelFinder answers which enabled Connections of protocol serve model
// (the server adapts proxy.Gateway.Serves). It must not generate anything.
type ModelFinder interface {
	Serves(ctx context.Context, protocol, model string) Diagnosis
}

// Fix is an Issue's aiport fix: the Change it would make (Kind, Scope,
// Proposal; Scope is always application) and the Suggestion in this queue
// that offers it. Once accepted, ChangeID is the Change, By who accepted it
// and At when.
type Fix struct {
	Suggestion string   `json:"suggestion"`
	Kind       Kind     `json:"kind"`
	Scope      Scope    `json:"scope"`
	Proposal   Proposal `json:"proposal"`
	State      FixState `json:"state"`
	ChangeID   string   `json:"change_id,omitempty"`
	By         string   `json:"by,omitempty"`
	At         int64    `json:"at_ms,omitempty"`
}

// Resolution is how and when an Issue was resolved; ChangeID is the aiport
// fix's Change when that resolved it.
type Resolution struct {
	How      How    `json:"how"`
	At       int64  `json:"at_ms"`
	ChangeID string `json:"change_id,omitempty"`
}

// Issue is one problem aiport keeps seeing for one Application.
type Issue struct {
	ID          string     `json:"id"`
	Application string     `json:"application"`
	Cause       Cause      `json:"cause"`
	Subject     string     `json:"subject"`
	State       IssueState `json:"state"`
	// Summary says what is wrong and what aiport found, in fixed words.
	Summary   string     `json:"summary"`
	Evidence  Evidence   `json:"evidence"`
	Diagnosis *Diagnosis `json:"diagnosis,omitempty"`
	// Fix is the aiport fix; nil when there is none to offer.
	Fix *Fix `json:"fix,omitempty"`
	// ApplicationFix is the fix in the Application, for its Coding agent
	// (written by ApplicationFix).
	ApplicationFix string `json:"application_fix"`
	// ResolvingBy is the fix aiport waits to see working while resolving.
	ResolvingBy How         `json:"resolving_by,omitempty"`
	Resolution  *Resolution `json:"resolution,omitempty"`
	// Dismissal is the Owner's dismissal (its evidence count decides when
	// the Issue may come back early).
	Dismissal *Verdict `json:"dismissal,omitempty"`
	// Reopened counts how often the Issue came back after being resolved
	// or dismissed.
	Reopened int `json:"reopened,omitempty"`
	// SinceMS is when the current state began.
	SinceMS   int64 `json:"since_ms"`
	CreatedAt int64 `json:"created_at_ms"`
	UpdatedAt int64 `json:"updated_at_ms"`
	// NeedsOwner: the Issue is open with an aiport fix waiting for the
	// Owner's decision (what the Changes badge counts).
	NeedsOwner bool `json:"needs_owner"`
}

// IssueFilter selects Issues. Zero fields don't filter.
type IssueFilter struct {
	Application string
	States      []IssueState
}

// IssueStore keeps Issues across restarts (SQLStore, store migration 7).
// A Store that also implements it keeps Issues; otherwise they are kept for
// this session only.
type IssueStore interface {
	LoadIssues() ([]Issue, error)
	SaveIssue(Issue) error
	DeleteIssues(ids []string) error
}

// ErrNoFix: the Issue has no aiport fix waiting for a decision.
var ErrNoFix = errors.New("this Issue has no aiport fix waiting for a decision")

// Diagnose lets detection ask f which Connections serve a model.
func (q *Queue) Diagnose(f ModelFinder) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.finder = f
}

// Issues returns matching Issues, newest first. With the loop off it
// returns nothing.
func (q *Queue) Issues(f IssueFilter) []Issue {
	if !q.On() {
		return []Issue{}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []Issue{}
	for i := len(q.issues) - 1; i >= 0; i-- {
		is := q.issues[i]
		if f.Application != "" && is.Application != f.Application {
			continue
		}
		if len(f.States) > 0 && !containsState(f.States, is.State) {
			continue
		}
		out = append(out, is)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out
}

// Issue returns Issue id.
func (q *Queue) Issue(id string) (Issue, bool) {
	if !q.On() {
		return Issue{}, false
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if i := q.issueIndex(id); i >= 0 {
		return q.issues[i], true
	}
	return Issue{}, false
}

// DecideIssue decides an open Issue's aiport fix: Accept applies it (as
// Decide does, with d.Proposal filling in what the fix leaves open) and the
// Issue becomes resolving; Dismiss dismisses the Issue (quiet period);
// Reject rejects the fix and leaves the Issue open.
func (q *Queue) DecideIssue(id string, d Decision, at time.Time) (Issue, config.Change, error) {
	if !q.On() {
		return Issue{}, config.Change{}, ErrOff
	}
	q.mu.Lock()
	i := q.issueIndex(id)
	if i < 0 {
		q.mu.Unlock()
		return Issue{}, config.Change{}, ErrNotFound
	}
	is := q.issues[i]
	q.mu.Unlock()
	if is.State != IssueOpen {
		return Issue{}, config.Change{}, ErrDecided
	}
	if is.Fix == nil || is.Fix.State != FixProposed {
		return Issue{}, config.Change{}, ErrNoFix
	}
	_, ch, err := q.Decide(is.Fix.Suggestion, d, at)
	if err != nil {
		return Issue{}, config.Change{}, err
	}
	is, _ = q.Issue(id)
	return is, ch, nil
}

// ApplicationFix is the fix in the Application for Issue is, written for
// its Coding agent: what is wrong, why, and what to change in the
// Application. It is the one place that writes it (the Issues routes, MCP
// and "Copy fix for the coding agent" all show this text).
func ApplicationFix(is Issue) string {
	switch is.Cause {
	case CauseUnknownModel:
		return unknownModelFix(is)
	}
	return ""
}

func unknownModelFix(is Issue) string {
	var b strings.Builder
	model := strconv.Quote(is.Subject)
	ev := is.Evidence
	b.WriteString("aiport Issue " + is.ID + ": calls from " + is.Application + " for model " + model + " fail with " + statusText(ev))
	b.WriteString(" (" + strconv.Itoa(ev.Count) + " times in the last " + strconv.Itoa(ev.WindowMinutes) + " minutes")
	if len(ev.Connections) > 0 {
		b.WriteString(", sent to Connection " + strings.Join(ev.Connections, ", "))
	}
	b.WriteString(").\n")
	b.WriteString("Cause: the Connection these calls go to does not serve " + model + ", and no Alias routes it for " + is.Application + ".\n")
	d := is.Diagnosis
	if d != nil {
		switch d.Answer {
		case ServedByOne:
			b.WriteString("Connection " + strconv.Quote(d.Serving[0]) + " serves " + model + ".\n")
		case ServedBySeveral:
			b.WriteString("Connections " + strings.Join(d.Serving, ", ") + " serve " + model + ".\n")
		case ServedByNone:
			b.WriteString("No enabled Connection of this gateway serves " + model + ".\n")
		default:
			b.WriteString("aiport could not list the models of every Connection, so it is not certain which serves " + model + ".\n")
		}
	}
	b.WriteString("Fix in the Application: send a model this gateway serves")
	if d != nil && len(d.Available) > 0 {
		b.WriteString(" for the " + d.Protocol + " protocol (" + strings.Join(d.Available, ", ") + ")")
	}
	b.WriteString(", or ask aiport to route " + model + " to a Connection that serves it with a Change request of kind " + string(KindApplicationAlias) + ". Do not retry the same model unchanged.")
	return b.String()
}

func statusText(ev Evidence) string {
	s := strconv.Itoa(ev.Status)
	if ev.Error != "" {
		s += " " + ev.Error
	}
	return s
}

func containsState(list []IssueState, s IssueState) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func (q *Queue) issueIndex(id string) int {
	for i := range q.issues {
		if q.issues[i].ID == id {
			return i
		}
	}
	return -1
}

func (q *Queue) issueFor(app string, cause Cause, subject string) int {
	for i := len(q.issues) - 1; i >= 0; i-- {
		is := q.issues[i]
		if is.Application == app && is.Cause == cause && is.Subject == subject {
			return i
		}
	}
	return -1
}

func (q *Queue) saveIssue(is *Issue) {
	is.NeedsOwner = is.State == IssueOpen && is.Fix != nil && is.Fix.State == FixProposed
	is.ApplicationFix = ApplicationFix(*is)
	if s, ok := q.store.(IssueStore); ok && q.store != nil {
		_ = s.SaveIssue(*is) // kept for this session even if the database fails
	}
}

// symptom is what one detection pass saw for one Application and model in
// the window: the errors that count, and the successful calls.
type symptom struct {
	app, model  string
	protocol    string
	errors      []telemetry.Event // oldest first
	lastSuccess int64
	diagnosis   *Diagnosis
}

func (s *symptom) since(ms int64) int {
	n := 0
	for _, e := range s.errors {
		if e.StartedAt > ms {
			n++
		}
	}
	return n
}

func (s *symptom) evidence(window int) Evidence {
	ev := Evidence{Count: len(s.errors), WindowMinutes: window, Status: 404}
	if len(s.errors) == 0 {
		return ev
	}
	ev.FromMS, ev.ToMS = s.errors[0].StartedAt, s.errors[len(s.errors)-1].StartedAt
	seen := map[string]bool{}
	for i := len(s.errors) - 1; i >= 0; i-- {
		e := s.errors[i]
		if len(ev.Examples) < 5 {
			ev.Examples = append(ev.Examples, e.ID)
		}
		if e.Target != "" && !seen[e.Target] {
			seen[e.Target] = true
			ev.Connections = append(ev.Connections, e.Target)
		}
		if ev.Error == "" {
			ev.Error = e.UpstreamErrorType
		}
	}
	sort.Strings(ev.Connections)
	return ev
}

// unknownModelSymptoms groups the window's inference calls by Application
// and requested model: upstream 404s for a model the Application has no
// Alias for, and successful calls.
func unknownModelSymptoms(c config.Config, history History, at time.Time) []*symptom {
	th := c.Adapting.UnknownModel
	from := at.Add(-time.Duration(th.WindowMinutes) * time.Minute)
	apps := map[string]config.Origin{}
	for _, o := range c.Origins {
		apps[o.ID] = o
	}
	protocols := map[string]string{}
	for _, t := range c.Targets {
		protocols[t.ID] = t.Protocol
	}
	groups := map[[2]string]*symptom{}
	events := history.Find(telemetry.Query{From: from.UnixMilli(), To: at.UnixMilli(), InferenceOnly: true})
	sort.SliceStable(events, func(i, j int) bool { return events[i].StartedAt < events[j].StartedAt })
	for _, e := range events {
		o, known := apps[e.Origin]
		if !known || e.RequestedModel == nil || *e.RequestedModel == "" {
			continue
		}
		failed := e.Outcome == "upstream_error" && e.Status == 404 && !hasAlias(o, *e.RequestedModel)
		if !failed && e.Outcome != "success" {
			continue
		}
		key := [2]string{e.Origin, *e.RequestedModel}
		g := groups[key]
		if g == nil {
			g = &symptom{app: e.Origin, model: *e.RequestedModel}
			groups[key] = g
		}
		if !failed {
			g.lastSuccess = max(g.lastSuccess, e.StartedAt)
			continue
		}
		g.errors = append(g.errors, e)
		if g.protocol = e.Protocol; g.protocol == "" {
			g.protocol = protocols[e.Target]
		}
	}
	out := make([]*symptom, 0, len(groups))
	for _, g := range groups {
		if g.protocol == "" {
			g.protocol = "openai"
		}
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].app+"\x00"+out[i].model < out[j].app+"\x00"+out[j].model })
	return out
}

// diagnose asks finder which Connections serve each model whose error
// sticks. It runs outside the queue's lock: listings may take a moment.
func diagnose(finder ModelFinder, symptoms []*symptom, count int, at time.Time) {
	if finder == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, s := range symptoms {
		if len(s.errors) < count {
			continue
		}
		d := finder.Serves(ctx, s.protocol, s.model)
		d.At = at.UnixMilli()
		if d.Serving == nil {
			d.Serving = []string{}
		}
		if d.Unknown == nil {
			d.Unknown = []string{}
		}
		if len(d.Available) > 20 {
			d.Available = d.Available[:20]
		}
		s.diagnosis = &d
	}
}

// trackUnknownModels opens, updates, reopens and resolves unknown-model
// Issues from one pass's symptoms, offering each open Issue's aiport fix as
// a Suggestion. It returns the Suggestions offered. Called with the lock
// held.
func (q *Queue) trackUnknownModels(c config.Config, symptoms []*symptom, at time.Time) []Item {
	th := c.Adapting.UnknownModel
	window := time.Duration(th.WindowMinutes) * time.Minute
	now := at.UnixMilli()
	q.revertedFixes(now)
	var offered []Item
	seen := map[int]*symptom{}
	for _, s := range symptoms {
		i := q.issueFor(s.app, CauseUnknownModel, s.model)
		if i >= 0 {
			seen[i] = s
		}
		if len(s.errors) < th.Count {
			continue
		}
		if i < 0 {
			q.issues = append(q.issues, Issue{ID: newID(), Application: s.app, Cause: CauseUnknownModel, Subject: s.model, State: IssueOpen, SinceMS: now, CreatedAt: now})
			i = len(q.issues) - 1
			seen[i] = s
		}
		is := &q.issues[i]
		switch is.State {
		case IssueResolved:
			if s.since(is.SinceMS) < th.Count {
				continue
			}
			q.reopen(is, now)
		case IssueDismissed:
			if quietUntil(is.Dismissal, len(s.errors), at) {
				continue
			}
			q.reopen(is, now)
		case IssueResolving:
			if s.since(is.SinceMS) < th.Count {
				continue // the fix may still be working; errors before it don't count
			}
			is.State, is.ResolvingBy, is.SinceMS = IssueOpen, "", now // the fix did not work
		}
		is.Evidence = s.evidence(th.WindowMinutes)
		if s.diagnosis != nil {
			is.Diagnosis = s.diagnosis
		}
		is.Summary = unknownModelSummary(*is)
		is.UpdatedAt = now
		if it, ok := q.offerFix(is, now); ok {
			offered = append(offered, it)
		}
		q.saveIssue(is)
	}
	// Self-resolution: calls succeed again after the last error (and after
	// the fix), or a whole window passed without the error.
	for i := range q.issues {
		is := &q.issues[i]
		if is.Cause != CauseUnknownModel || (is.State != IssueOpen && is.State != IssueResolving) {
			continue
		}
		after := is.Evidence.ToMS
		if is.State == IssueResolving {
			after = max(after, is.SinceMS)
		}
		s := seen[i]
		if s != nil && len(s.errors) > 0 {
			after = max(after, s.errors[len(s.errors)-1].StartedAt) // errors below the threshold still mean it goes on
		}
		if (s == nil || s.lastSuccess <= after) && now-after < window.Milliseconds() {
			continue
		}
		q.resolve(is, now)
	}
	return offered
}

// revertedFixes moves resolving Issues whose aiport fix was reverted back to
// open, offering the fix again.
func (q *Queue) revertedFixes(now int64) {
	reverted := map[string]bool{}
	for _, ch := range q.live.Changes() {
		if ch.RevertedAt != 0 {
			reverted[ch.ID] = true
		}
	}
	for i := range q.issues {
		is := &q.issues[i]
		if is.State != IssueResolving || is.Fix == nil || is.Fix.State != FixApplied || !reverted[is.Fix.ChangeID] {
			continue
		}
		is.Fix.State, is.State, is.ResolvingBy, is.SinceMS, is.UpdatedAt = FixReverted, IssueOpen, "", now, now
		q.saveIssue(is)
	}
}

func (q *Queue) reopen(is *Issue, now int64) {
	is.State, is.SinceMS, is.Resolution, is.ResolvingBy, is.Fix = IssueOpen, now, nil, "", nil
	is.Reopened++
}

// resolve marks is resolved, recording how, and withdraws a fix still
// waiting for the Owner.
func (q *Queue) resolve(is *Issue, now int64) {
	r := &Resolution{How: Stopped, At: now}
	if is.State == IssueResolving && is.ResolvingBy != "" {
		r.How = is.ResolvingBy
	}
	if r.How == ByAiportFix && is.Fix != nil {
		r.ChangeID = is.Fix.ChangeID
	}
	if is.Fix != nil && is.Fix.State == FixProposed {
		if j := q.index(is.Fix.Suggestion); j >= 0 && q.items[j].Status == Pending {
			q.items[j].Status, q.items[j].UpdatedAt = Withdrawn, now
			q.save(q.items[j])
		}
		is.Fix = nil
	}
	is.State, is.Resolution, is.ResolvingBy, is.SinceMS, is.UpdatedAt = IssueResolved, r, "", now, now
	q.saveIssue(is)
}

// offerFix offers an open Issue's aiport fix as a pending Suggestion (or
// updates the one waiting). With exactly one Connection serving the model
// the fix is complete; otherwise the Owner chooses the Connection.
func (q *Queue) offerFix(is *Issue, now int64) (Item, bool) {
	if is.State != IssueOpen {
		return Item{}, false
	}
	p := Proposal{Alias: is.Subject, Model: is.Subject}
	if d := is.Diagnosis; d != nil && d.Answer == ServedByOne && len(d.Serving) == 1 {
		p.Target = d.Serving[0]
	}
	ev := is.Evidence
	it := q.submit(Item{
		Source: FromSuggestion, Kind: KindApplicationAlias, Application: is.Application, Subject: is.Subject, Scope: ScopeApplication,
		Need: is.Summary, Evidence: &ev, Proposal: p, Issue: is.ID,
	}, now)
	is.Fix = &Fix{Suggestion: it.ID, Kind: KindApplicationAlias, Scope: ScopeApplication, Proposal: p, State: FixProposed}
	return it, true
}

// settledFix keeps an Issue in step with the Owner's decision on the
// Suggestion that offers its aiport fix. Called with the lock held.
func (q *Queue) settledFix(it Item) {
	if it.Issue == "" || it.Decision == nil {
		return
	}
	i := q.issueIndex(it.Issue)
	if i < 0 {
		return
	}
	is := &q.issues[i]
	if is.Fix == nil || is.Fix.Suggestion != it.ID || is.State != IssueOpen {
		return
	}
	v := it.Decision
	switch it.Status {
	case Accepted:
		is.Fix.State, is.Fix.Proposal, is.Fix.ChangeID, is.Fix.By, is.Fix.At = FixApplied, it.Proposal, v.ChangeID, v.By, v.At
		is.State, is.ResolvingBy, is.SinceMS = IssueResolving, ByAiportFix, v.At
	case Rejected:
		is.Fix.State = FixRejected
	case Dismissed:
		is.State, is.Dismissal, is.SinceMS = IssueDismissed, v, v.At
	}
	is.UpdatedAt = v.At
	q.saveIssue(is)
}

// quietUntil reports whether a dismissal d still keeps an Issue quiet at
// time at with count errors: within QuietFor, unless the evidence doubled.
func quietUntil(d *Verdict, count int, at time.Time) bool {
	if d == nil {
		return false
	}
	return at.UnixMilli()-d.At < QuietFor.Milliseconds() && count < 2*d.EvidenceCount
}

func unknownModelSummary(is Issue) string {
	ev := is.Evidence
	s := is.Application + " asked for model \"" + is.Subject + "\" " + strconv.Itoa(ev.Count) + " times in the last " + strconv.Itoa(ev.WindowMinutes) + " minutes and got " + statusText(ev) + " each time."
	d := is.Diagnosis
	if d == nil {
		return s + " Choose the Connection that serves it: add an Alias."
	}
	switch d.Answer {
	case ServedByOne:
		return s + " Only Connection \"" + d.Serving[0] + "\" serves it: add an Alias for " + is.Application + " to it."
	case ServedBySeveral:
		return s + " Connections " + strings.Join(d.Serving, ", ") + " serve it: choose one for an Alias."
	case ServedByNone:
		return s + " No enabled Connection serves it: choose a Connection and model for an Alias, or fix the Application."
	}
	return s + " aiport could not list the models of " + strings.Join(d.Unknown, ", ") + ": choose the Connection that serves it."
}

// pruneIssues forgets resolved and dismissed Issues older than retention
// (dismissed ones are kept at least QuietFor). Called with the lock held.
func (q *Queue) pruneIssues(at time.Time, retention time.Duration) {
	kept, gone := q.issues[:0], []string{}
	for _, is := range q.issues {
		keep := retention
		if is.State == IssueDismissed {
			keep = max(retention, QuietFor)
		}
		if (is.State == IssueResolved || is.State == IssueDismissed) && at.UnixMilli()-is.UpdatedAt > keep.Milliseconds() {
			gone = append(gone, is.ID)
			continue
		}
		kept = append(kept, is)
	}
	q.issues = kept
	if s, ok := q.store.(IssueStore); ok && len(gone) > 0 {
		_ = s.DeleteIssues(gone)
	}
}
