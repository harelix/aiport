package decide

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
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
// it; Subject says which instance (for CauseUnknownModel, the model name;
// for CauseFailingConnection, the Connection id).
type Cause string

const (
	// CauseUnknownModel: upstream 404s for a model the Application has no
	// Alias for. The aiport fix is an Alias for that Application.
	CauseUnknownModel Cause = "unknown_model"
	// CauseFailingConnection: the Application's calls to one Connection keep
	// failing (gateway errors and timeouts). The aiport fix binds the
	// Application to a working Connection; it is never applied automatically.
	CauseFailingConnection Cause = "failing_connection"
)

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

// Served answers "which enabled Connections serve this model?", with the
// wire values of proxy.Answer. (Answer in this package is a Change
// request's answer.)
type Served string

const (
	ServedByOne     Served = "one"     // certain: exactly one Connection serves it
	ServedByNone    Served = "none"    // no Connection lists it
	ServedBySeveral Served = "several" // ambiguous
	ServedUnknown   Served = "unknown" // a listing failed: no certain answer
)

// Diagnosis is what aiport found out about an Issue's cause. For an unknown
// model: which enabled Connections of Protocol list the model (Serving),
// which could not be listed (Unknown), and the models Connections of that
// protocol list (Available, bounded), from the free model listings.
type Diagnosis struct {
	Protocol  string   `json:"protocol"`
	Answer    Served   `json:"answer"`
	Serving   []string `json:"serving"`
	Unknown   []string `json:"unknown"`
	Available []string `json:"available,omitempty"`
	// Summary says what aiport found, in fixed words (diagnosisSummary).
	Summary string `json:"summary"`
	At      int64  `json:"at_ms"`
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
	ID          string `json:"id"`
	Application string `json:"application"`
	Cause       Cause  `json:"cause"`
	Subject     string `json:"subject"`
	// Protocol is the protocol the Issue is on (CauseLevelUnmapped only).
	Protocol string     `json:"protocol,omitempty"`
	State    IssueState `json:"state"`
	// Summary says what is wrong and what aiport found, in fixed words.
	Summary   string     `json:"summary"`
	Evidence  Evidence   `json:"evidence"`
	Diagnosis *Diagnosis `json:"diagnosis,omitempty"`
	// Fix is the aiport fix; nil when there is none to offer.
	Fix *Fix `json:"fix,omitempty"`
	// ApplicationFix is the fix in the Application, for its Coding agent
	// (written by ApplicationFix).
	ApplicationFix string `json:"application_fix"`
	// Explanation is the root cause in words, written on the Owner's click
	// (Queue.Explain); nil until then.
	Explanation *Explanation `json:"explanation,omitempty"`
	// Catalog is a Catalog-change Issue's change and what the Application
	// can still use (CauseCatalogChanged only).
	Catalog *CatalogNotice `json:"catalog_change,omitempty"`
	// Report is a reported request's failure and the Owner's note
	// (CauseReportedRequest only).
	Report *Reported `json:"report,omitempty"`
	// BranchFix is the latest Fix it run's branch and diff summary
	// (Queue.RecordBranch); nil until Fix it ran.
	BranchFix *BranchFix `json:"branch_fix,omitempty"`
	// ResolvingBy is the fix aiport waits to see working while resolving.
	ResolvingBy How         `json:"resolving_by,omitempty"`
	Resolution  *Resolution `json:"resolution,omitempty"`
	// Dismissal is the Owner's dismissal (its evidence count decides when
	// the Issue may come back early).
	Dismissal *Verdict `json:"dismissal,omitempty"`
	// Reopened counts how often the Issue came back after being resolved
	// or dismissed.
	Reopened int `json:"reopened,omitempty"`
	// NeverAuto lists the aiport fixes aiport must never apply automatically
	// again for this Application and cause: the ones the Owner reverted after
	// aiport applied them automatically, and the ones the Owner rejected (a
	// rejected fix may still be approved by hand, never automatically). An
	// Issue holding any is kept beyond request-history retention, so the
	// Owner's decision survives (pruneIssues). The JSON key stays
	// refused_auto_fixes for the stored Issues.
	NeverAuto []Proposal `json:"refused_auto_fixes,omitempty"`
	// SinceMS is when the current state began.
	SinceMS   int64 `json:"since_ms"`
	CreatedAt int64 `json:"created_at_ms"`
	UpdatedAt int64 `json:"updated_at_ms"`
	// NeedsOwner: the Issue is open with an aiport fix waiting for the
	// Owner's decision (what the Changes badge counts).
	NeedsOwner bool `json:"needs_owner"`
}

// undecided reports whether an Issue in state s is still for someone to
// act on: open, or resolving while aiport waits to see a fix work.
func undecided(s IssueState) bool { return s == IssueOpen || s == IssueResolving }

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

// Automatic is who makes a Change aiport applies itself (adapting.auto_fix).
const Automatic = "aiport (automatic)"

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
	q.drainCatalog()
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []Issue{}
	for i := len(q.issues) - 1; i >= 0; i-- {
		is := q.issues[i]
		if f.Application != "" && is.Application != f.Application {
			continue
		}
		if len(f.States) > 0 && !slices.Contains(f.States, is.State) {
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
	q.drainCatalog()
	q.mu.Lock()
	defer q.mu.Unlock()
	if i := q.issueIndex(id); i >= 0 {
		return q.issues[i], true
	}
	return Issue{}, false
}

// DecideIssue decides an open or resolving Issue's aiport fix: Accept
// applies it (as Decide does, with d.Proposal filling in what the fix
// leaves open) and the Issue becomes resolving; Reject rejects the fix and
// leaves the Issue as it is. Dismiss dismisses any Issue that is not
// resolved or dismissed yet (quiet period), whatever its fix stands at.
func (q *Queue) DecideIssue(id string, d Decision, at time.Time) (Issue, config.Change, error) {
	if !q.On() {
		return Issue{}, config.Change{}, ErrOff
	}
	q.drainCatalog()
	q.mu.Lock()
	i := q.issueIndex(id)
	if i < 0 {
		q.mu.Unlock()
		return Issue{}, config.Change{}, ErrNotFound
	}
	is := q.issues[i]
	if !undecided(is.State) {
		q.mu.Unlock()
		return Issue{}, config.Change{}, ErrDecided
	}
	if is.Fix == nil || is.Fix.State != FixProposed {
		defer q.mu.Unlock()
		if d.Action != Dismiss {
			return Issue{}, config.Change{}, ErrNoFix
		}
		// Nothing waits for the Owner: dismiss the Issue itself.
		if utf8.RuneCountInString(d.Note) > 2000 {
			return Issue{}, config.Change{}, Invalid{"the note must be at most 2,000 characters"}
		}
		if d.By == "" {
			d.By = config.Owner
		}
		p := &q.issues[i]
		now := at.UnixMilli()
		p.Dismissal = &Verdict{Action: Dismiss, Note: d.Note, By: d.By, At: now, EvidenceCount: p.Evidence.Count}
		p.State, p.ResolvingBy, p.SinceMS, p.UpdatedAt = IssueDismissed, "", now, now
		q.saveIssue(p)
		return *p, config.Change{}, nil
	}
	suggestion := is.Fix.Suggestion
	q.mu.Unlock()
	// Decide checks again under the lock that the Suggestion still waits.
	_, ch, err := q.Decide(suggestion, d, at)
	if err != nil {
		return Issue{}, config.Change{}, err
	}
	is, _ = q.Issue(id)
	return is, ch, nil
}

// ApplicationFix is the fix in the Application for Issue is, written for
// its Coding agent so it can be pasted as-is: what is wrong, the evidence,
// what aiport found, what to change in the Application and how to confirm
// the fix. It is the one place that writes it (the Issues routes, MCP and
// "Copy fix for the Coding agent" all show this text). It holds only ids,
// model names, counts, times and request ids: never prompts, provider
// messages, keys or tokens.
//
// A saved explanation comes right after the first paragraph, saying who
// wrote it; a branch fix the Owner kept comes before it.
func ApplicationFix(is Issue) string {
	r, ok := causes[is.Cause]
	if !ok {
		return ""
	}
	s := r.applicationFix(is)
	if e := is.Explanation; e != nil && e.Text != "" {
		head, rest, _ := strings.Cut(s, "\n\n")
		s = head + "\n\n" + explanationText(*e) + "\n\n" + rest
	}
	if b := is.BranchFix; b != nil && b.State == BranchKept {
		head, rest, _ := strings.Cut(s, "\n\n")
		s = head + "\n\n" + branchText(*b) + "\n\n" + rest
	}
	return s
}

func unknownModelFix(is Issue) string {
	var b strings.Builder
	app, model := is.Application, strconv.Quote(is.Subject)
	ev, d := is.Evidence, is.Diagnosis
	b.WriteString("aiport Issue " + is.ID + ": calls from the Application " + app + " for model " + model + " fail with " + statusText(ev) + ".\n\n")

	b.WriteString("What is wrong: the Connection these calls go to")
	if len(ev.Connections) > 0 {
		b.WriteString(" (" + strings.Join(ev.Connections, ", ") + ")")
	}
	b.WriteString(" does not serve " + model + ", and no Alias routes " + model + " for " + app + ".\n\n")

	b.WriteString(evidenceText(ev, true) + "\n")
	if d != nil {
		b.WriteString("What aiport found: " + diagnosisSummary(*d, is.Subject) + "\n")
	}
	b.WriteString("\n")

	if f := is.Fix; f != nil && f.State == FixApplied && f.Proposal.Target != "" {
		b.WriteString("Already fixed in aiport: calls from " + app + " for " + model + " now go to Connection " + strconv.Quote(f.Proposal.Target) + " (model " + strconv.Quote(f.Proposal.Model) + "). No code change is needed; aiport confirms the fix from the next calls.\n\n")
	} else {
		b.WriteString("What to change in " + app + ", one of:\n")
		target := ""
		if d != nil && d.Answer == ServedByOne {
			target = d.Serving[0]
		}
		if target != "" {
			b.WriteString("1. Ask aiport to route " + model + " for " + app + ": file a Change request of kind " + string(KindApplicationAlias) + " (alias " + model + ", prefer_connection " + strconv.Quote(target) + ", prefer_model " + model + "). The Owner decides which Connection and model serve it.")
		} else {
			b.WriteString("1. Ask aiport to route " + model + " for " + app + ": file a Change request of kind " + string(KindApplicationAlias) + " (alias " + model + "; prefer_connection and prefer_model only if you know what should serve it). The Owner decides which Connection and model serve it.")
		}
		b.WriteString(" If the Owner approves aiport's own fix for this Issue instead, no code change is needed.\n")
		b.WriteString("2. Send a model this gateway serves.")
		if d != nil && len(d.Available) > 0 {
			b.WriteString(" Models the " + d.Protocol + " Connections of this gateway list: " + strings.Join(d.Available, ", ") + ".")
		}
		b.WriteString("\nDo not retry " + model + " unchanged.\n\n")
	}

	b.WriteString(confirmText(app, ev.Status, "after the change, ", " for "+model))
	return b.String()
}

// evidenceText is the "Evidence:" line of an Application fix: how many calls
// failed, when, the window, and (with sentTo) the Connections they went to,
// with example request ids.
func evidenceText(ev Evidence, sentTo bool) string {
	s := "Evidence: " + strconv.Itoa(ev.Count) + " failed calls"
	if ev.FromMS > 0 && ev.ToMS > 0 {
		s += " between " + utc(ev.FromMS) + " and " + utc(ev.ToMS)
	}
	s += " (looking back " + strconv.Itoa(ev.WindowMinutes) + " minutes)"
	if sentTo && len(ev.Connections) > 0 {
		s += ", sent to Connection " + strings.Join(ev.Connections, ", ")
	}
	s += "."
	if len(ev.Examples) > 0 {
		s += " Example request ids: " + strings.Join(ev.Examples, ", ") + "."
	}
	return s
}

// confirmText is the "How to confirm" paragraph of an Application fix:
// when (such as "after the change, ") one call with app's token answers 200
// instead of status, and aiport resolves the Issue once app's calls (which:
// such as " for \"llama3:8b\"") succeed again.
func confirmText(app string, status int, when, which string) string {
	return "How to confirm: " + when + "make one call through aiport with " + app + "'s own token and check it answers 200 instead of " + strconv.Itoa(status) + ". aiport marks this Issue resolved by itself when calls from " + app + which + " succeed again; it checks request history, so nothing else needs reporting."
}

// diagnosisSummary says what aiport found out about which Connections
// serve model, in fixed words (Diagnosis.Summary, also in the Application
// fix).
func diagnosisSummary(d Diagnosis, model string) string {
	m := strconv.Quote(model)
	var s string
	switch d.Answer {
	case ServedByOne:
		s = "Connection " + strconv.Quote(d.Serving[0]) + " serves " + m
	case ServedBySeveral:
		s = "Connections " + strings.Join(d.Serving, ", ") + " serve " + m
	case ServedByNone:
		s = "no enabled Connection of this gateway serves " + m
	default:
		s = "not every Connection could be listed, so it is not certain which serves " + m
	}
	return s + " (from the Connections' free model lists)."
}

func utc(ms int64) string { return time.UnixMilli(ms).UTC().Format(time.RFC3339) }

func statusText(ev Evidence) string {
	s := strconv.Itoa(ev.Status)
	if ev.Error != "" {
		s += " " + ev.Error
	}
	return s
}

func (q *Queue) issueIndex(id string) int {
	for i := range q.issues {
		if q.issues[i].ID == id {
			return i
		}
	}
	return -1
}

// issueFor is the latest Issue of app for cause and subject, and for a
// Cause kept per protocol (causeRules.onProtocol) on protocol; -1 if none.
func (q *Queue) issueFor(app string, cause Cause, subject, protocol string) int {
	perProtocol := causes[cause].onProtocol
	for i := len(q.issues) - 1; i >= 0; i-- {
		is := q.issues[i]
		if is.Application == app && is.Cause == cause && is.Subject == subject && (!perProtocol || is.Protocol == protocol) {
			return i
		}
	}
	return -1
}

func (q *Queue) saveIssue(is *Issue) {
	is.NeedsOwner = undecided(is.State) && is.Fix != nil && is.Fix.State == FixProposed
	is.ApplicationFix = ApplicationFix(*is)
	if s, ok := q.store.(IssueStore); ok {
		_ = s.SaveIssue(*is) // kept for this session even if the database fails
	}
	q.publishHints()
}

// symptom is what one detection pass saw for one Application and Subject
// of a Cause in the window: the errors that count, and the successful
// calls. For an unknown model, protocol and diagnosis; for a failing
// Connection, proposedConnection, the working Connection to propose.
type symptom struct {
	cause              Cause
	app, subject       string
	protocol           string
	errors             []telemetry.Event // oldest first
	lastSuccess        int64
	diagnosis          *Diagnosis
	proposedConnection string
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
	ev := Evidence{Count: len(s.errors), WindowMinutes: window}
	if len(s.errors) == 0 {
		return ev
	}
	ev.Status = s.errors[len(s.errors)-1].Status
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
		failed := notFound(e) && !hasAlias(o, *e.RequestedModel)
		if !failed && !telemetry.Succeeded(e.Outcome) {
			continue
		}
		key := [2]string{e.Origin, *e.RequestedModel}
		g := groups[key]
		if g == nil {
			g = &symptom{cause: CauseUnknownModel, app: e.Origin, subject: *e.RequestedModel}
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
	sort.Slice(out, func(i, j int) bool { return out[i].app+"\x00"+out[i].subject < out[j].app+"\x00"+out[j].subject })
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
		d := finder.Serves(ctx, s.protocol, s.subject)
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
		d.Summary = diagnosisSummary(d, s.subject)
		s.diagnosis = &d
	}
}

// causeRules is everything aiport knows about one Cause, so adding a Cause
// touches only its entry in causes (and its detector): its threshold, its
// aiport fix, its summary, its Application fix and which failing calls
// count as its errors.
type causeRules struct {
	threshold func(config.Config) config.Threshold
	// fix is the aiport fix for is from this pass's symptom s.
	fix     func(is Issue, s *symptom) (Kind, Proposal)
	summary func(is Issue, p Proposal) string
	// applicationFix writes the Application fix (see ApplicationFix).
	applicationFix func(is Issue) string
	// counts reports whether call e is one of the errors of Issue is (what
	// X-Aiport-Hint uses to name the Issue).
	counts func(e telemetry.Event, is hintIssue) bool
	// onProtocol: the Issue keeps the protocol of its symptoms.
	onProtocol bool
}

var causes = map[Cause]causeRules{
	CauseUnknownModel: {
		threshold: func(c config.Config) config.Threshold { return c.Adapting.UnknownModel },
		fix: func(is Issue, _ *symptom) (Kind, Proposal) {
			p := Proposal{Alias: is.Subject, Model: is.Subject}
			if d := is.Diagnosis; d != nil && d.Answer == ServedByOne && len(d.Serving) == 1 {
				p.Target = d.Serving[0] // certain; otherwise the Owner chooses
			}
			return KindApplicationAlias, p
		},
		summary:        func(is Issue, _ Proposal) string { return unknownModelSummary(is) },
		applicationFix: unknownModelFix,
		counts: func(e telemetry.Event, is hintIssue) bool {
			return notFound(e) && e.RequestedModel != nil && *e.RequestedModel == is.subject
		},
	},
	CauseFailingConnection: {
		threshold: func(c config.Config) config.Threshold { return c.Adapting.FailingConnection },
		fix: func(_ Issue, s *symptom) (Kind, Proposal) {
			return KindBindApplication, Proposal{Target: s.proposedConnection} // empty: the Owner chooses
		},
		summary:        func(is Issue, p Proposal) string { return failingSummary(is, p.Target) },
		applicationFix: failingFix,
		counts:         func(e telemetry.Event, is hintIssue) bool { return failed(e) && e.Target == is.subject },
	},
	CauseLevelUnmapped: {
		threshold: func(c config.Config) config.Threshold { return c.Adapting.LevelUnmapped },
		fix: func(is Issue, s *symptom) (Kind, Proposal) {
			// The model is the Owner's to choose: aiport never invents one.
			return KindApplicationAlias, Proposal{Alias: is.Subject, Target: s.proposedConnection}
		},
		summary:        levelUnmappedSummary,
		applicationFix: levelUnmappedFix,
		counts: func(e telemetry.Event, is hintIssue) bool {
			return levelUnmapped(e) && *e.RequestedModel == is.subject && config.ProtocolOf(e.Route) == is.protocol
		},
		onProtocol: true,
	},
	// A Catalog change is not detected from errors (CatalogChanged opens
	// it), so it has no threshold and no aiport fix.
	CauseCatalogChanged: {
		applicationFix: catalogFix,
		counts:         func(e telemetry.Event, is hintIssue) bool { return sentRemoved(e, is.names) },
	},
	// A reported request is opened by the Owner (Report), so it has no
	// threshold and no aiport fix.
	CauseReportedRequest: {
		applicationFix: reportedFix,
		counts:         func(e telemetry.Event, is hintIssue) bool { return is.report != nil && is.report.mayBeKind(e) },
	},
}

// notFound reports whether the provider answered e with a 404: the error
// of an unknown model.
func notFound(e telemetry.Event) bool { return e.Outcome == "upstream_error" && e.Status == 404 }

// track opens, updates, reopens and resolves the Issues of one Cause from
// one pass's symptoms, offering each open Issue's aiport fix as a
// Suggestion. It returns the Suggestions offered. Called with the lock held.
//
// succeeded, when set, says when an Application's last successful call for
// an Issue's Subject was (for Causes whose successes are not grouped with
// the errors).
func (q *Queue) track(c config.Config, cause Cause, symptoms []*symptom, succeeded func(app, subject string) int64, at time.Time) []Item {
	r := causes[cause]
	th := r.threshold(c)
	window := time.Duration(th.WindowMinutes) * time.Minute
	now := at.UnixMilli()
	var offered []Item
	seen := map[int]*symptom{}
	for _, s := range symptoms {
		i := q.issueFor(s.app, cause, s.subject, s.protocol)
		if i >= 0 {
			seen[i] = s
		}
		if len(s.errors) < th.Count {
			continue
		}
		if i < 0 {
			q.issues = append(q.issues, Issue{ID: newID(), Application: s.app, Cause: cause, Subject: s.subject, State: IssueOpen, SinceMS: now, CreatedAt: now})
			if r.onProtocol {
				q.issues[len(q.issues)-1].Protocol = s.protocol
			}
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
			if stillQuiet(is.Dismissal, len(s.errors), at) {
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
		if r.onProtocol {
			is.Protocol = s.protocol
		}
		if s.diagnosis != nil {
			is.Diagnosis = s.diagnosis
		}
		kind, p := r.fix(*is, s)
		is.Summary = r.summary(*is, p)
		is.UpdatedAt = now
		if it, ok := q.offerFix(is, kind, p, now); ok {
			offered = append(offered, it)
		}
		q.saveIssue(is)
	}
	// Self-resolution: calls succeed again after the last error (and after
	// the fix), or a whole window passed without the error.
	for i := range q.issues {
		is := &q.issues[i]
		if is.Cause != cause || (is.State != IssueOpen && is.State != IssueResolving) {
			continue
		}
		after := is.Evidence.ToMS
		if is.State == IssueResolving {
			after = max(after, is.SinceMS)
		}
		s := seen[i]
		var success int64
		if s != nil {
			success = s.lastSuccess
			if len(s.errors) > 0 {
				after = max(after, s.errors[len(s.errors)-1].StartedAt) // errors below the threshold still mean it goes on
			}
		}
		if succeeded != nil {
			success = max(success, succeeded(is.Application, is.Subject))
		}
		confirmed, quiet := success > after, now-after >= window.Milliseconds()
		switch {
		case is.State == IssueResolving && confirmed:
			q.resolve(is, is.ResolvingBy, now) // the fix works: a call for the same cause succeeded
		case is.State == IssueResolving && is.ResolvingBy == ByApplicationFix:
			// The Application's word is not enough: its calls for the
			// same cause must succeed (silence never confirms it).
		case confirmed || quiet:
			// No fix, or an aiport fix no call has shown working yet.
			q.resolve(is, Stopped, now)
		}
	}
	return offered
}

// revertedFixes moves resolving Issues whose aiport fix was reverted back to
// open, offering the fix again. A reverted automatic fix (resolving or
// already resolved) is recorded in NeverAuto, so it is never applied
// automatically again.
func (q *Queue) revertedFixes(now int64) {
	reverted := map[string]bool{}
	for _, ch := range q.live.Changes() {
		if ch.RevertedAt != 0 {
			reverted[ch.ID] = true
		}
	}
	for i := range q.issues {
		is := &q.issues[i]
		if is.Fix == nil || is.Fix.State != FixApplied || !reverted[is.Fix.ChangeID] {
			continue
		}
		if is.Fix.By == Automatic && !neverAuto(*is, is.Fix.Proposal) {
			is.NeverAuto = append(is.NeverAuto, is.Fix.Proposal)
		}
		is.Fix.State, is.UpdatedAt = FixReverted, now
		if is.State == IssueResolving {
			is.State, is.ResolvingBy, is.SinceMS = IssueOpen, "", now
		}
		q.saveIssue(is)
	}
}

func (q *Queue) reopen(is *Issue, now int64) {
	is.State, is.SinceMS, is.Resolution, is.ResolvingBy, is.Fix = IssueOpen, now, nil, "", nil
	is.Reopened++
}

// withdrawFix withdraws is's aiport fix while it waits for the Owner (the
// Suggestion becomes Withdrawn). Only an Issue that resolves itself does
// this. Called with the lock held.
func (q *Queue) withdrawFix(is *Issue, now int64) {
	if is.Fix == nil || is.Fix.State != FixProposed {
		return
	}
	if j := q.index(is.Fix.Suggestion); j >= 0 && q.items[j].Status == Pending {
		q.items[j].Status, q.items[j].UpdatedAt = Withdrawn, now
		q.save(q.items[j])
	}
	is.Fix = nil
}

// resolve marks is resolved as how, and withdraws a fix still waiting for
// the Owner: the Issue resolved itself.
func (q *Queue) resolve(is *Issue, how How, now int64) {
	r := &Resolution{How: how, At: now}
	if how == ByAiportFix && is.Fix != nil {
		r.ChangeID = is.Fix.ChangeID
	}
	q.withdrawFix(is, now)
	is.State, is.Resolution, is.ResolvingBy, is.SinceMS, is.UpdatedAt = IssueResolved, r, "", now, now
	q.saveIssue(is)
}

// offerFix offers an open Issue's aiport fix, a Change of kind for
// Proposal p scoped to its Application, as a pending Suggestion (or updates
// the one waiting). What p leaves empty the Owner chooses.
func (q *Queue) offerFix(is *Issue, kind Kind, p Proposal, now int64) (Item, bool) {
	if is.State != IssueOpen {
		return Item{}, false
	}
	ev := is.Evidence
	it := q.submit(Item{
		Source: FromSuggestion, Kind: kind, Application: is.Application, Subject: is.Subject, Scope: ScopeApplication,
		Need: is.Summary, Evidence: &ev, Proposal: p, Issue: is.ID,
	}, now)
	is.Fix = &Fix{Suggestion: it.ID, Kind: kind, Scope: ScopeApplication, Proposal: p, State: FixProposed}
	return it, true
}

// automaticFixes picks, from the Suggestions offered in this pass, the
// aiport fixes aiport applies itself (their positions in offered): only with
// adapting.auto_fix on, for an Issue whose error sticks (it was offered in
// this pass), diagnosed in this pass as certain (exactly one Connection
// serves the model), with a fix that is an Alias for that one Application
// and that the Owner has not refused. Called with the lock held.
func (q *Queue) automaticFixes(c config.Config, offered []Item, now int64) []int {
	if !c.Adapting.AutoFix {
		return nil
	}
	var out []int
	for j, it := range offered {
		i := q.issueIndex(it.Issue)
		if it.Issue == "" || i < 0 || it.Status != Pending {
			continue
		}
		is := q.issues[i]
		d, fix := is.Diagnosis, is.Fix
		certain := d != nil && d.At == now && d.Answer == ServedByOne && len(d.Serving) == 1
		if !certain || is.State != IssueOpen || fix == nil || fix.Suggestion != it.ID || fix.State != FixProposed ||
			fix.Kind != KindApplicationAlias || fix.Scope != ScopeApplication || fix.Proposal.Target != d.Serving[0] ||
			it.Scope != ScopeApplication || len(it.AskedBy) > 1 || neverAuto(is, fix.Proposal) {
			continue
		}
		out = append(out, j)
	}
	return out
}

// neverAuto reports whether aiport must never apply fix p for Issue is
// automatically (the Owner reverted or rejected it).
func neverAuto(is Issue, p Proposal) bool { return slices.Contains(is.NeverAuto, p) }

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
	if is.Fix == nil || is.Fix.Suggestion != it.ID || !undecided(is.State) {
		return
	}
	v := it.Decision
	switch it.Status {
	case Accepted:
		is.Fix.State, is.Fix.Proposal, is.Fix.ChangeID, is.Fix.By, is.Fix.At = FixApplied, it.Proposal, v.ChangeID, v.By, v.At
		is.State, is.ResolvingBy, is.SinceMS = IssueResolving, ByAiportFix, v.At
	case Rejected:
		is.Fix.State = FixRejected
		if !neverAuto(*is, it.Proposal) {
			is.NeverAuto = append(is.NeverAuto, it.Proposal)
		}
	case Dismissed:
		is.State, is.Dismissal, is.SinceMS = IssueDismissed, v, v.At
	}
	is.UpdatedAt = v.At
	q.saveIssue(is)
}

// stillQuiet reports whether a dismissal d still keeps an Issue quiet at
// time at with count errors: within QuietFor, unless the evidence doubled.
func stillQuiet(d *Verdict, count int, at time.Time) bool {
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
		// An Issue holding fixes never to apply automatically (NeverAuto)
		// is kept: the Owner's revert or reject is final.
		if (is.State == IssueResolved || is.State == IssueDismissed) && len(is.NeverAuto) == 0 && at.UnixMilli()-is.UpdatedAt > keep.Milliseconds() {
			gone = append(gone, is.ID)
			continue
		}
		kept = append(kept, is)
	}
	q.issues = kept
	q.publishHints()
	if s, ok := q.store.(IssueStore); ok && len(gone) > 0 {
		_ = s.DeleteIssues(gone)
	}
}
