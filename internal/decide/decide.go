// Package decide is aiport's decision queue: it holds Suggestions and Change
// requests until the Owner decides, applies accepted ones through
// config.Live so they become revertable Changes, and runs the detectors that
// turn patterns in request history into Suggestions.
//
// Its interface is small on purpose; the dashboard API, the detectors and the
// Change request endpoints all use the same few calls: Submit, List, Decide,
// Detect (one pass at a given time), Prune (forget decided items older than
// request history) and Run (Prune and Detect on a background ticker); File,
// Requests and RequestOf serve Change requests. Nothing here runs on the
// forwarding path.
package decide

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aiport/internal/capability"
	"aiport/internal/config"
	"aiport/internal/extension"
	"aiport/internal/telemetry"

	"go.yaml.in/yaml/v3"
)

// Source is where an item came from.
type Source string

const (
	FromSuggestion Source = "suggestion" // aiport noticed a pattern in request history
	FromRequest    Source = "request"    // an Application asked (Change request)
)

// Status is where an item stands. Only Pending items need a decision.
type Status string

const (
	Pending   Status = "pending"
	Accepted  Status = "accepted"
	Rejected  Status = "rejected"
	Dismissed Status = "dismissed"
	// Withdrawn: aiport took a Suggestion back before the Owner decided,
	// because the Issue it was the fix for resolved itself.
	Withdrawn Status = "withdrawn"
)

// Action is what the Owner does with a pending item.
type Action string

const (
	Accept  Action = "accept"
	Reject  Action = "reject"
	Dismiss Action = "dismiss"
)

// Scope is who a Change affects.
type Scope string

const (
	ScopeApplication Scope = "application"
	ScopeEvery       Scope = "every_application"
)

// Kind is the kind of Change an item is about. The built-in kinds are the
// ones capability.Kinds lists; applyKind says how each is applied.
type Kind string

const (
	KindApplicationAlias Kind = capability.ApplicationAlias // an Alias only one Application sees; for every Application a shared Alias
	KindBindApplication  Kind = capability.BindApplication  // all of one Application's calls to one Connection; for every Application the protocol default
)

// KindPairApplication tells the Owner that something keeps trying to connect
// without a valid token and should be paired. It is informational: there is
// nothing to apply, the Owner dismisses it.
const KindPairApplication Kind = "pair_application"

// OpensApplications is where the dashboard sends the Owner for a
// KindPairApplication Suggestion: the Applications panel, where Pairing is.
const OpensApplications = "applications"

// QuietFor is how long a dismissed Suggestion stays hidden unless its
// evidence doubles.
const QuietFor = 7 * 24 * time.Hour

// Evidence is what a Suggestion is based on: Count matching requests, the
// first at FromMS and the last at ToMS, found looking WindowMinutes back.
type Evidence struct {
	Count         int   `json:"count"`
	FromMS        int64 `json:"from_ms"`
	ToMS          int64 `json:"to_ms"`
	WindowMinutes int   `json:"window_minutes"`
	// For an Issue: the HTTP status and the provider's short error type
	// (a fixed identifier, never its message), up to five example request
	// ids (newest first) and the Connections the calls went to.
	Status      int      `json:"status,omitempty"`
	Error       string   `json:"error,omitempty"`
	Examples    []string `json:"examples,omitempty"`
	Connections []string `json:"connections,omitempty"`
}

// Proposal is the Change an item would make. For an Alias: requests for
// model Alias go to Model on Connection Target. Empty fields are for the
// Owner to fill in when accepting.
type Proposal struct {
	Alias  string `json:"alias,omitempty"`
	Target string `json:"target,omitempty"`
	Model  string `json:"model,omitempty"`
}

// Verdict is the Owner's decision on an item.
type Verdict struct {
	Action   Action `json:"action"`
	Note     string `json:"note,omitempty"`
	Scope    Scope  `json:"scope,omitempty"`
	By       string `json:"by"`
	At       int64  `json:"at_ms"`
	ChangeID string `json:"change_id,omitempty"`
	// EvidenceCount is the evidence when decided (a dismissed Suggestion
	// comes back early once its evidence doubles).
	EvidenceCount int `json:"evidence_count,omitempty"`
}

// Item is one Suggestion or Change request in the queue.
type Item struct {
	ID          string `json:"id"`
	Source      Source `json:"source"`
	Kind        Kind   `json:"kind"`
	Application string `json:"application"`
	// Subject is what repeats of the same item share (for an unknown model,
	// the model name).
	Subject string `json:"subject"`
	Status  Status `json:"status"`
	Need    string `json:"need"`
	Scope   Scope  `json:"scope"`
	// A Change request's own text (plain text, never sent to a model): the
	// details, why, the argument for its Scope and who filed it (free text,
	// never used for identity).
	Details   string    `json:"details,omitempty"`
	Why       string    `json:"why,omitempty"`
	ScopeCase string    `json:"scope_case,omitempty"`
	FiledBy   string    `json:"filed_by,omitempty"`
	Evidence  *Evidence `json:"evidence,omitempty"`
	Proposal  Proposal  `json:"proposal"`
	// Repeats counts how often the same pending item was submitted again.
	Repeats int `json:"repeats,omitempty"`
	// AskedBy lists each Application that asked for this Change for every
	// Application, with its own argument (merged requests).
	AskedBy   []Asker  `json:"asked_by,omitempty"`
	CreatedAt int64    `json:"created_at_ms"`
	UpdatedAt int64    `json:"updated_at_ms"`
	Decision  *Verdict `json:"decision,omitempty"`
	// Informational items have nothing to apply: they can only be dismissed.
	Informational bool `json:"informational,omitempty"`
	// Opens names the dashboard panel the Owner acts in (OpensApplications).
	Opens string `json:"opens,omitempty"`
	// Issue is the Issue this Suggestion is the aiport fix for.
	Issue string `json:"issue,omitempty"`
}

// Decision is what the Owner sends: an action, a note for the requester, the
// Scope granted, and the details of the Change where the proposal leaves
// them open (non-empty fields override the proposal).
type Decision struct {
	Action   Action
	Note     string
	Scope    Scope
	Proposal Proposal
	By       string // empty = config.Owner
}

// Filter selects items for List. Zero fields don't filter. Application
// matches the Application that filed an item or any that joined it.
type Filter struct {
	Application string
	Status      Status
	Source      Source
}

// Store keeps items across restarts (SQLStore in the database).
type Store interface {
	Load() ([]Item, error)
	Save(Item) error
	Delete(ids []string) error
}

// History is where detectors read request metadata (telemetry.Collector).
type History interface {
	Find(telemetry.Query) []telemetry.Event
}

// FailedAuth is where the unknown-token detector reads how many
// authentications failed in a time range (telemetry.FailedAuth): counts only,
// never tokens.
type FailedAuth interface {
	Count(from, to time.Time) (n int, first, last int64)
}

var (
	ErrNotFound = errors.New("no such item")
	ErrDecided  = errors.New("this item was already decided")
	ErrOff      = errors.New("the adapting loop is turned off in configuration")
	// ErrChanged: the configuration changed while a guardrail checked the
	// Change, so it would no longer be the Change the guardrail judged.
	ErrChanged = errors.New("the configuration changed while a guardrail checked this Change; nothing was applied, decide again")
)

// Invalid is a decision that cannot be applied as given; the message says why.
type Invalid struct{ Reason string }

func (e Invalid) Error() string { return e.Reason }

// Queue is the decision queue. It keeps its items in memory and, when it has
// a Store (SQLStore in the database next to aiport.yaml), there too. It is
// safe for concurrent use.
type Queue struct {
	mu      sync.Mutex
	live    *config.Live
	history History
	store   Store
	items   []Item
	filed   map[string][]int64 // Application → when it filed Change requests, for the rate limit
	failed  FailedAuth
	issues  []Issue
	finder  ModelFinder
	// hints is the open and resolving Issues per Application, read without
	// the lock on the forwarding path (HintIssue); publishHints swaps it.
	hints atomic.Pointer[map[string][]hintIssue]
	// inbox is Catalog changes waiting to become Issues (CatalogChanged).
	inboxMu sync.Mutex
	inbox   []CatalogNotice
}

// WatchFailedAuth lets detection read failed-authentication counts from f, so
// repeated failures become a Suggestion to pair an Application.
func (q *Queue) WatchFailedAuth(f FailedAuth) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.failed = f
}

// New returns a queue that applies Changes through live, reads request
// history from history (nil = no detection) and keeps items in store
// (nil = this session only), loading the items kept there.
func New(live *config.Live, history History, store Store) (*Queue, error) {
	q := &Queue{live: live, history: history, store: store}
	if store != nil {
		kept, err := store.Load()
		if err != nil {
			return nil, err
		}
		q.items = kept
		if s, ok := store.(IssueStore); ok {
			issues, err := s.LoadIssues()
			if err != nil {
				return nil, err
			}
			q.issues = issues
		}
	}
	q.publishHints()
	return q, nil
}

// On reports whether the adapting loop is on in the current configuration.
func (q *Queue) On() bool { return !q.live.Config().Adapting.Disabled }

// Submit adds a pending item at time at. If an identical pending item exists
// (same source, kind, Application and subject), that one is updated instead:
// its evidence and proposal are replaced and Repeats goes up.
func (q *Queue) Submit(it Item, at time.Time) (Item, error) {
	if !q.On() {
		return Item{}, ErrOff
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.submit(it, at.UnixMilli()), nil
}

func (q *Queue) submit(it Item, at int64) Item {
	status := Pending
	if it.Status == NotSupported {
		status = NotSupported // kept and counted as demand, never decided
	}
	if i := q.latest(it, status); i >= 0 {
		cur := &q.items[i]
		cur.Evidence, cur.Proposal, cur.Need, cur.UpdatedAt, cur.Issue = it.Evidence, it.Proposal, it.Need, at, it.Issue
		cur.Repeats++
		q.save(*cur)
		return *cur
	}
	it.ID, it.Status, it.CreatedAt, it.UpdatedAt, it.Decision, it.Repeats = newID(), status, at, at, nil, 0
	if it.Scope == "" {
		it.Scope = ScopeApplication
	}
	q.items = append(q.items, it)
	q.save(it)
	return it
}

// latest returns the index of the newest item like it (same source, kind,
// Application and subject, and the same Issue when both fix one: one
// Level unmapped on two protocols has two Issues, each with its own fix)
// with one of statuses (any status when none), or -1.
func (q *Queue) latest(it Item, statuses ...Status) int {
	for i := len(q.items) - 1; i >= 0; i-- {
		c := q.items[i]
		if c.Source != it.Source || c.Kind != it.Kind || c.Application != it.Application || c.Subject != it.Subject || c.Issue != "" && it.Issue != "" && c.Issue != it.Issue {
			continue
		}
		if len(statuses) == 0 {
			return i
		}
		for _, s := range statuses {
			if c.Status == s {
				return i
			}
		}
	}
	return -1
}

// List returns matching items, newest first. With the loop off it returns
// nothing.
func (q *Queue) List(f Filter) []Item {
	if !q.On() {
		return []Item{}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []Item{}
	for i := len(q.items) - 1; i >= 0; i-- {
		it := q.items[i]
		if (f.Application == "" || askedBy(it, f.Application)) && (f.Status == "" || it.Status == f.Status) && (f.Source == "" || it.Source == f.Source) {
			out = append(out, it)
		}
	}
	return out
}

// Decide records the Owner's decision on pending item id at time at. Accept
// applies the Change through config.Live (so it can be reverted) and returns
// it, unless a guardrail Extension blocks it (extension.Denied, with its
// reason; the item stays pending); reject and dismiss change nothing in the
// configuration and run no Extension.
//
// Guardrails run outside the lock, so the configuration may change while
// they check. The Change is then built again from the configuration as it
// is, and applied only if it is the Change the guardrails judged; otherwise
// nothing is applied, the item stays pending and Decide returns ErrChanged.
func (q *Queue) Decide(id string, d Decision, at time.Time) (Item, config.Change, error) {
	if !q.On() {
		return Item{}, config.Change{}, ErrOff
	}
	if len(d.Note) > 2000 {
		return Item{}, config.Change{}, Invalid{"the note must be at most 2,000 characters"}
	}
	if d.By == "" {
		d.By = config.Owner
	}
	var judged *guarded
	for {
		it, ch, check, err := q.settle(id, d, at, judged)
		if check == nil {
			return it, ch, err
		}
		if judged != nil {
			return Item{}, config.Change{}, ErrChanged
		}
		if err := extension.Check(check.config, config.HookGuardrail, extension.Input{Change: &check.change}); err != nil {
			return Item{}, config.Change{}, err
		}
		judged = check
	}
}

// guarded is an accepted Change as the guardrails see it, with the
// guardrails that judge it.
type guarded struct {
	config     config.Config
	change     extension.Change
	guardrails string
}

func (g *guarded) same(o *guarded) bool {
	return g != nil && o != nil && g.change == o.change && g.guardrails == o.guardrails
}

// settle decides pending item id under the lock. When the Change needs a
// guardrail check that judged does not cover, nothing is decided and the
// Change to check is returned instead.
func (q *Queue) settle(id string, d Decision, at time.Time, judged *guarded) (Item, config.Change, *guarded, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	i := q.index(id)
	if i < 0 {
		return Item{}, config.Change{}, nil, ErrNotFound
	}
	it := q.items[i]
	if it.Status != Pending {
		return Item{}, config.Change{}, nil, ErrDecided
	}
	if d.By == Automatic && !q.mayApplyAutomatically(it) {
		return Item{}, config.Change{}, nil, ErrDecided
	}
	v := &Verdict{Action: d.Action, Note: d.Note, By: d.By, At: at.UnixMilli()}
	if it.Evidence != nil {
		v.EvidenceCount = it.Evidence.Count
	}
	var ch config.Change
	switch d.Action {
	case Accept:
		scope := d.Scope
		if scope == "" {
			scope = it.Scope
		}
		p := it.Proposal
		if d.Proposal.Alias != "" {
			p.Alias = d.Proposal.Alias
		}
		if d.Proposal.Target != "" {
			p.Target = d.Proposal.Target
		}
		if d.Proposal.Model != "" {
			p.Model = d.Proposal.Model
		}
		c := q.live.Config()
		if j := q.issueIndex(it.Issue); it.Issue != "" && j >= 0 {
			if err := onIssueProtocol(c, q.issues[j], p); err != nil {
				return Item{}, config.Change{}, nil, err
			}
		}
		edit, what, err := change(c, it, p, scope)
		if err != nil {
			return Item{}, config.Change{}, nil, err
		}
		if guardrails := c.ExtensionsFor(config.HookGuardrail); len(guardrails) > 0 {
			now := &guarded{config: c, guardrails: fmt.Sprint(guardrails), change: extension.Change{
				Item: it.ID, Source: string(it.Source), Application: it.Application, Kind: string(it.Kind), Scope: string(scope),
				What: what, Alias: p.Alias, Target: p.Target, Model: p.Model, By: d.By,
			}}
			if !now.same(judged) {
				return Item{}, config.Change{}, now, nil
			}
		}
		if _, ch, err = q.live.ChangeAs(d.By, what, edit); err != nil {
			if errors.Is(err, config.ErrReadOnly) {
				return Item{}, config.Change{}, nil, err
			}
			return Item{}, config.Change{}, nil, Invalid{err.Error()}
		}
		it.Proposal, it.Status, v.Scope, v.ChangeID = p, Accepted, scope, ch.ID
	case Reject:
		it.Status = Rejected
	case Dismiss:
		it.Status = Dismissed
	default:
		return Item{}, config.Change{}, nil, Invalid{"action must be accept, reject or dismiss"}
	}
	it.Decision, it.UpdatedAt = v, at.UnixMilli()
	q.items[i] = it
	q.save(it)
	q.settledFix(it)
	return it, ch, nil, nil
}

// mayApplyAutomatically reports whether aiport may still apply Suggestion
// it as an automatic fix, checked again in the locked step that applies it
// (the Owner may have changed something since the fix was picked, or while
// a guardrail checked it): adapting and adapting.auto_fix are on, its Issue
// is still open (an automatic fix never overrides the Application's "fixed
// on my side", or anything else that moved the Issue on) and the Owner has
// not refused the fix. Called with the lock held.
func (q *Queue) mayApplyAutomatically(it Item) bool {
	c := q.live.Config()
	j := q.issueIndex(it.Issue)
	return !c.Adapting.Disabled && c.Adapting.AutoFix && it.Issue != "" && j >= 0 &&
		q.issues[j].State == IssueOpen && !neverAuto(q.issues[j], it.Proposal)
}

// index returns the position of item id, or -1.
func (q *Queue) index(id string) int {
	for j := range q.items {
		if q.items[j].ID == id {
			return j
		}
	}
	return -1
}

// change is the configuration edit that applies an accepted item.
// The Scope decides the edit: for this Application an Alias or binding of
// its own; for every Application a shared Alias or the protocol default.
func change(c config.Config, it Item, p Proposal, scope Scope) (config.Edit, string, error) {
	if it.Informational {
		return nil, "", Invalid{"this Suggestion has nothing to apply: pair the Application on the Integrate page, then dismiss it"}
	}
	if scope != ScopeApplication && scope != ScopeEvery {
		return nil, "", Invalid{"scope must be application or every_application"}
	}
	apply := applyKind[it.Kind]
	if apply == nil {
		return nil, "", Invalid{"aiport cannot apply this kind of Change yet"}
	}
	return apply(c, it, p, scope)
}

// applyKind is how each built-in kind of Change is applied: the one table of
// kinds aiport can apply, next to capability.Kinds, which describes them to
// Applications (init checks the two agree).
var applyKind = map[Kind]func(c config.Config, it Item, p Proposal, scope Scope) (config.Edit, string, error){
	KindApplicationAlias: aliasChange,
	KindBindApplication:  bindChange,
}

func init() {
	if len(applyKind) != len(capability.Kinds) {
		panic("decide: applyKind and capability.Kinds list different kinds")
	}
	for _, k := range capability.Kinds {
		if applyKind[Kind(k.Kind)] == nil {
			panic("decide: no way to apply built-in kind " + k.Kind)
		}
	}
}

// aliasChange: for this Application (and any that joined) an Alias of its
// own; for every Application a shared Alias.
func aliasChange(c config.Config, it Item, p Proposal, scope Scope) (config.Edit, string, error) {
	if p.Alias == "" || p.Target == "" || p.Model == "" {
		return nil, "", Invalid{"choose the alias, the Connection and the model the alias should use"}
	}
	if config.IsLevel(p.Alias) {
		return levelChange(c, it, p, scope)
	}
	if scope == ScopeEvery {
		what := "shared alias " + p.Alias + " → " + p.Target + "/" + p.Model + " for every Application"
		return config.AddAlias(config.Alias{Alias: p.Alias, Target: p.Target, Model: p.Model}), what, nil
	}
	apps := requesters(it)
	edits := make([]config.Edit, len(apps))
	for i, app := range apps {
		edits[i] = config.AddOriginAlias(app, config.OriginAlias{Alias: p.Alias, Target: p.Target, Model: p.Model})
	}
	return each(edits), "alias " + p.Alias + " → " + p.Target + "/" + p.Model + " for " + strings.Join(apps, ", "), nil
}

// bindChange: for this Application (and any that joined) a binding to one
// Connection; for every Application that Connection's protocol default.
func bindChange(c config.Config, it Item, p Proposal, scope Scope) (config.Edit, string, error) {
	if p.Target == "" {
		return nil, "", Invalid{"choose the Connection to bind to"}
	}
	if failingTarget(it) == p.Target {
		return nil, "", Invalid{"choose another Connection than the failing one"}
	}
	if scope == ScopeApplication {
		apps := requesters(it)
		edits := make([]config.Edit, len(apps))
		for i, app := range apps {
			edits[i] = config.BindOrigin(app, p.Target)
		}
		return each(edits), "bind " + strings.Join(apps, ", ") + " → " + p.Target, nil
	}
	var t *config.Target
	for i := range c.Targets {
		if c.Targets[i].ID == p.Target {
			t = &c.Targets[i]
		}
	}
	if t == nil {
		return nil, "", Invalid{"no Connection " + p.Target + " in configuration"}
	}
	return config.SetDefaultTarget(t.Protocol, t.ID), "default " + t.Protocol + " Connection → " + t.ID + " for every Application", nil
}

// requesters are the Applications a Change for Scope application is for:
// the one that filed it and every one that joined it (a merged request
// for every Application that the Owner narrowed).
func requesters(it Item) []string {
	apps := []string{it.Application}
	for _, a := range it.AskedBy {
		if !slices.Contains(apps, a.Application) {
			apps = append(apps, a.Application)
		}
	}
	return apps
}

// each applies edits in order as one edit, so they are one Change.
func each(edits []config.Edit) config.Edit {
	return func(root *yaml.Node) error {
		for _, e := range edits {
			if err := e(root); err != nil {
				return err
			}
		}
		return nil
	}
}

func (q *Queue) save(it Item) {
	if q.store != nil {
		_ = q.store.Save(it) // kept for this session even if the database fails
	}
}

// Run calls Detect until stop is closed, waiting every() between passes
// (read again after each pass, so a changed interval applies), with times
// from now. It runs whether or not the loop is on: Detect checks On each
// time, so turning the loop on needs no restart. It is the only place
// detection runs; nothing runs on the forwarding path.
func (q *Queue) Run(stop <-chan struct{}, every func() time.Duration, now func() time.Time) {
	t := time.NewTimer(every())
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			at := now()
			q.Prune(at)
			q.Detect(at)
			t.Reset(every())
		}
	}
}

// Prune forgets decided items (accepted, rejected or dismissed) decided
// longer ago than request history is kept (telemetry.retention_days), so a
// Change request's text is kept no longer than the requests it was about.
// Dismissed items are kept at least QuietFor, so they stay
// quiet as long as promised. Items still waiting, demand for unsupported
// kinds and the Changes themselves (config.Live) are kept. Prune runs
// whether or not the loop is on.
func (q *Queue) Prune(at time.Time) {
	retention := time.Duration(q.live.Config().Telemetry.RetentionDays) * 24 * time.Hour
	if retention <= 0 {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	kept, gone := q.items[:0], []string{}
	for _, it := range q.items {
		keep := retention
		if it.Status == Dismissed {
			keep = max(retention, QuietFor)
		}
		decided := it.Status == Accepted || it.Status == Rejected || it.Status == Dismissed || it.Status == Withdrawn
		if decided && at.UnixMilli()-it.UpdatedAt > keep.Milliseconds() {
			gone = append(gone, it.ID)
			continue
		}
		kept = append(kept, it)
	}
	q.items = kept
	if len(gone) > 0 && q.store != nil {
		_ = q.store.Delete(gone) // forgotten for this session even if the database fails
	}
	q.pruneIssues(at, retention)
}

// Detect runs every detector once over request history up to time at,
// keeps Issues in step (opens, updates, reopens and resolves them) and
// returns the Suggestions it created or updated. Diagnosis (which
// Connections serve a model) runs before the lock is taken. With the loop
// off it does nothing.
func (q *Queue) Detect(at time.Time) []Item {
	if q.history == nil || !q.On() {
		return nil
	}
	c := q.live.Config()
	q.drainCatalog()
	symptoms := unknownModelSymptoms(c, q.history, at)
	q.mu.Lock()
	finder := q.finder
	q.mu.Unlock()
	diagnose(finder, symptoms, c.Adapting.UnknownModel.Count, at)
	failing, succeeded := failingSymptoms(c, q.history, at)
	unmapped := levelUnmappedSymptoms(c, q.history, at)
	q.mu.Lock()
	var found []Item
	if q.failed != nil {
		found = unknownTokens(c, q.failed, at)
	}
	q.revertedFixes(at.UnixMilli())
	q.catalogResolutions(at)
	out := q.track(c, CauseUnknownModel, symptoms, nil, at)
	// Only unknown-model Issues can be fixed automatically: a failing
	// Connection has no diagnosis, so automaticFixes never picks it.
	auto := q.automaticFixes(c, out, at.UnixMilli())
	out = append(out, q.track(c, CauseFailingConnection, failing, succeeded, at)...)
	out = append(out, q.track(c, CauseLevelUnmapped, unmapped, nil, at)...)
	for _, s := range found {
		if q.quiet(s, at) {
			continue
		}
		out = append(out, q.submit(s, at.UnixMilli()))
	}
	q.mu.Unlock()
	// Automatic fixes are decided like the Owner's, outside the lock
	// (guardrail Extensions still judge them).
	for _, j := range auto {
		if it, _, err := q.Decide(out[j].ID, Decision{Action: Accept, By: Automatic}, at); err == nil {
			out[j] = it
		}
	}
	return out
}

// quiet reports whether s was dismissed in the last QuietFor
// and its evidence has not doubled since.
func (q *Queue) quiet(s Item, at time.Time) bool {
	if q.latest(s, Pending) >= 0 {
		return false
	}
	i := q.latest(s)
	if i < 0 {
		return false
	}
	last := q.items[i]
	if last.Status != Dismissed || last.Decision == nil {
		return false
	}
	recent := at.UnixMilli()-last.Decision.At < QuietFor.Milliseconds()
	return recent && s.Evidence.Count < 2*last.Decision.EvidenceCount
}

func hasAlias(o config.Origin, model string) bool {
	return slices.ContainsFunc(o.Aliases, func(a config.OriginAlias) bool { return a.Alias == model })
}

func mostUsed(counts map[string]int) string {
	best, n := "", 0
	for id, c := range counts {
		if c > n || c == n && id < best {
			best, n = id, c
		}
	}
	return best
}

func newID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random source unavailable")
	}
	return hex.EncodeToString(b[:])
}

// unknownTokens suggests pairing when requests with a missing or unknown token
// were refused at least the configured number of times within the window.
// Who sent them is unknown (only counts are kept), so the Suggestion belongs
// to no Application.
func unknownTokens(c config.Config, f FailedAuth, at time.Time) []Item {
	th := c.Adapting.UnknownToken
	n, first, last := f.Count(at.Add(-time.Duration(th.WindowMinutes)*time.Minute), at)
	if n < th.Count {
		return nil
	}
	ev := Evidence{Count: n, FromMS: first, ToMS: last, WindowMinutes: th.WindowMinutes}
	return []Item{{
		Source: FromSuggestion, Kind: KindPairApplication, Subject: "unknown_token", Scope: ScopeApplication,
		Need:          "Requests without a valid token were refused " + strconv.Itoa(n) + " times in the last " + strconv.Itoa(th.WindowMinutes) + " minutes: an Application may be trying to connect without being paired. Pair it on the Integrate page.",
		Evidence:      &ev,
		Informational: true, Opens: OpensApplications,
	}}
}
