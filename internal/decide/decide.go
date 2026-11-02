// Package decide is aiport's decision queue: it holds Suggestions and Change
// requests until the Owner decides, applies accepted ones through
// config.Live so they become revertable Changes, and runs the detectors that
// turn patterns in request history into Suggestions.
//
// Its interface is small on purpose; the dashboard API, the detectors and the
// Change request endpoints all use the same five calls: Submit, List, Decide,
// Detect (one pass at a given time) and Run (Detect on a background ticker).
// Nothing here runs on the forwarding path.
package decide

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"aiport/internal/capability"
	"aiport/internal/config"
	"aiport/internal/extension"
	"aiport/internal/telemetry"
)

// Where an item came from.
const (
	FromSuggestion = "suggestion" // aiport noticed a pattern in request history
	FromRequest    = "request"    // an Application asked (Change request)
)

// Statuses. Only Pending items need a decision.
const (
	Pending   = "pending"
	Accepted  = "accepted"
	Rejected  = "rejected"
	Dismissed = "dismissed"
)

// Actions the Owner can take on a pending item.
const (
	Accept  = "accept"
	Reject  = "reject"
	Dismiss = "dismiss"
)

// Scopes: who a Change affects.
const (
	ScopeApplication = "application"
	ScopeEvery       = "every_application"
)

// The built-in kinds of Change (see package capability).
const (
	KindApplicationAlias = capability.ApplicationAlias // an Alias only one Application sees; for every Application a shared Alias
	KindBindApplication  = capability.BindApplication  // all of one Application's calls to one Connection; for every Application the protocol default
)

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
	Action   string `json:"action"`
	Note     string `json:"note,omitempty"`
	Scope    string `json:"scope,omitempty"`
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
	Source      string `json:"source"`
	Kind        string `json:"kind"`
	Application string `json:"application"`
	// Subject is what repeats of the same item share (for an unknown model,
	// the model name).
	Subject string `json:"subject"`
	Status  string `json:"status"`
	Need    string `json:"need"`
	Scope   string `json:"scope"`
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
	Repeats   int      `json:"repeats,omitempty"`
	CreatedAt int64    `json:"created_at_ms"`
	UpdatedAt int64    `json:"updated_at_ms"`
	Decision  *Verdict `json:"decision,omitempty"`
}

// Decision is what the Owner sends: an action, a note for the requester, the
// Scope granted, and the details of the Change where the proposal leaves
// them open (non-empty fields override the proposal).
type Decision struct {
	Action   string
	Note     string
	Scope    string
	Proposal Proposal
	By       string // empty = config.Owner
}

// Filter selects items for List. Zero fields don't filter.
type Filter struct {
	Application, Status, Source string
}

// Store keeps items across restarts (store.Decisions in the database).
type Store interface {
	Load() ([]Item, error)
	Save(Item) error
}

// History is where detectors read request metadata (telemetry.Collector).
type History interface {
	Find(telemetry.Query) []telemetry.Event
}

var (
	ErrNotFound = errors.New("no such item")
	ErrDecided  = errors.New("this item was already decided")
	ErrOff      = errors.New("the adapting loop is turned off in configuration")
)

// Invalid is a decision that cannot be applied as given; the message says why.
type Invalid struct{ Reason string }

func (e Invalid) Error() string { return e.Reason }

// Queue is the decision queue. It is safe for concurrent use.
type Queue struct {
	mu      sync.Mutex
	live    *config.Live
	history History
	store   Store
	items   []Item
	filed   map[string][]int64 // Application → when it filed Change requests, for the rate limit
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
	}
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
		cur.Evidence, cur.Proposal, cur.Need, cur.UpdatedAt = it.Evidence, it.Proposal, it.Need, at
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
// Application and subject) with one of statuses (any status when none), or -1.
func (q *Queue) latest(it Item, statuses ...string) int {
	for i := len(q.items) - 1; i >= 0; i-- {
		c := q.items[i]
		if c.Source != it.Source || c.Kind != it.Kind || c.Application != it.Application || c.Subject != it.Subject {
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
		if (f.Application == "" || it.Application == f.Application) && (f.Status == "" || it.Status == f.Status) && (f.Source == "" || it.Source == f.Source) {
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
func (q *Queue) Decide(id string, d Decision, at time.Time) (Item, config.Change, error) {
	if !q.On() {
		return Item{}, config.Change{}, ErrOff
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	i := q.index(id)
	if i < 0 {
		return Item{}, config.Change{}, ErrNotFound
	}
	it := q.items[i]
	if it.Status != Pending {
		return Item{}, config.Change{}, ErrDecided
	}
	if len(d.Note) > 2000 {
		return Item{}, config.Change{}, Invalid{"the note must be at most 2,000 characters"}
	}
	by := d.By
	if by == "" {
		by = config.Owner
	}
	v := &Verdict{Action: d.Action, Note: d.Note, By: by, At: at.UnixMilli()}
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
		edit, what, err := change(c, it, p, scope)
		if err != nil {
			return Item{}, config.Change{}, err
		}
		// Guardrail Extensions run before the Change is applied, outside the
		// lock so a slow one never holds up the queue; a block leaves the
		// item waiting.
		if len(c.ExtensionsFor(config.HookGuardrail)) > 0 {
			q.mu.Unlock()
			err := extension.Check(c, config.HookGuardrail, extension.Input{Change: &extension.Change{
				Item: it.ID, Source: it.Source, Application: it.Application, Kind: it.Kind, Scope: scope,
				What: what, Alias: p.Alias, Target: p.Target, Model: p.Model, By: by,
			}})
			q.mu.Lock()
			if err != nil {
				return Item{}, config.Change{}, err
			}
			if i = q.index(id); i < 0 {
				return Item{}, config.Change{}, ErrNotFound
			}
			if q.items[i].Status != Pending {
				return Item{}, config.Change{}, ErrDecided
			}
			it = q.items[i]
		}
		if _, ch, err = q.live.ChangeAs(by, what, edit); err != nil {
			if errors.Is(err, config.ErrReadOnly) {
				return Item{}, config.Change{}, err
			}
			return Item{}, config.Change{}, Invalid{err.Error()}
		}
		it.Proposal, it.Status, v.Scope, v.ChangeID = p, Accepted, scope, ch.ID
	case Reject:
		it.Status = Rejected
	case Dismiss:
		it.Status = Dismissed
	default:
		return Item{}, config.Change{}, Invalid{"action must be accept, reject or dismiss"}
	}
	it.Decision, it.UpdatedAt = v, at.UnixMilli()
	q.items[i] = it
	q.save(it)
	return it, ch, nil
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
func change(c config.Config, it Item, p Proposal, scope string) (config.Edit, string, error) {
	if scope != ScopeApplication && scope != ScopeEvery {
		return nil, "", Invalid{"scope must be application or every_application"}
	}
	switch it.Kind {
	case KindApplicationAlias:
		if p.Alias == "" || p.Target == "" || p.Model == "" {
			return nil, "", Invalid{"choose the alias, the Connection and the model the alias should use"}
		}
		if scope == ScopeEvery {
			what := "shared alias " + p.Alias + " → " + p.Target + "/" + p.Model + " for every Application"
			return config.AddAlias(config.Alias{Alias: p.Alias, Target: p.Target, Model: p.Model}), what, nil
		}
		what := "alias " + p.Alias + " → " + p.Target + "/" + p.Model + " for " + it.Application
		return config.AddOriginAlias(it.Application, config.OriginAlias{Alias: p.Alias, Target: p.Target, Model: p.Model}), what, nil
	case KindBindApplication:
		if p.Target == "" {
			return nil, "", Invalid{"choose the Connection to bind to"}
		}
		if scope == ScopeApplication {
			return config.BindOrigin(it.Application, p.Target), "bind " + it.Application + " → " + p.Target, nil
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
		d := c.Defaults
		if t.Protocol == "anthropic" {
			d.AnthropicTarget = t.ID
		} else {
			d.OpenAITarget = t.ID
		}
		return config.SetDefaults(d), "default " + t.Protocol + " Connection → " + t.ID + " for every Application", nil
	}
	return nil, "", Invalid{"aiport cannot apply this kind of Change yet"}
}

func (q *Queue) save(it Item) {
	if q.store != nil {
		_ = q.store.Save(it) // kept for this session even if the database fails
	}
}

// Run calls Detect every interval until stop is closed, with times from now.
// It is the only place detection runs; nothing runs on the forwarding path.
func (q *Queue) Run(stop <-chan struct{}, interval time.Duration, now func() time.Time) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			q.Detect(now())
		}
	}
}

// Detect runs every detector once over request history up to time at and
// returns the Suggestions it created or updated. With the loop off it does
// nothing.
func (q *Queue) Detect(at time.Time) []Item {
	if q.history == nil || !q.On() {
		return nil
	}
	c := q.live.Config()
	found := unknownModels(c, q.history, at)
	q.mu.Lock()
	defer q.mu.Unlock()
	var out []Item
	for _, s := range found {
		if q.quiet(s, at) {
			continue
		}
		out = append(out, q.submit(s, at.UnixMilli()))
	}
	return out
}

// quiet reports whether s was dismissed (or rejected) in the last QuietFor
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
	if last.Status != Dismissed && last.Status != Rejected || last.Decision == nil {
		return false
	}
	recent := at.UnixMilli()-last.Decision.At < QuietFor.Milliseconds()
	return recent && s.Evidence.Count < 2*last.Decision.EvidenceCount
}

// unknownModels finds Applications that repeatedly got upstream 404s for the
// same model within the configured window, for models they have no Alias for.
func unknownModels(c config.Config, history History, at time.Time) []Item {
	th := c.Adapting.UnknownModel
	from := at.Add(-time.Duration(th.WindowMinutes) * time.Minute)
	apps := map[string]config.Origin{}
	for _, o := range c.Origins {
		apps[o.ID] = o
	}
	type group struct {
		app, model string
		ev         Evidence
		targets    map[string]int
	}
	groups := map[[2]string]*group{}
	for _, e := range history.Find(telemetry.Query{From: from.UnixMilli(), To: at.UnixMilli(), Outcome: "upstream_error", InferenceOnly: true}) {
		o, known := apps[e.Origin]
		if e.Status != 404 || e.RequestedModel == nil || *e.RequestedModel == "" || !known || hasAlias(o, *e.RequestedModel) {
			continue
		}
		key := [2]string{e.Origin, *e.RequestedModel}
		g := groups[key]
		if g == nil {
			g = &group{app: e.Origin, model: *e.RequestedModel, ev: Evidence{FromMS: e.StartedAt, ToMS: e.StartedAt, WindowMinutes: th.WindowMinutes}, targets: map[string]int{}}
			groups[key] = g
		}
		g.ev.Count++
		g.ev.FromMS, g.ev.ToMS = min(g.ev.FromMS, e.StartedAt), max(g.ev.ToMS, e.StartedAt)
		g.targets[e.Target]++
	}
	var out []Item
	for _, g := range groups {
		if g.ev.Count < th.Count {
			continue
		}
		ev := g.ev
		out = append(out, Item{
			Source: FromSuggestion, Kind: KindApplicationAlias, Application: g.app, Subject: g.model, Scope: ScopeApplication,
			Need:     g.app + " asked for model \"" + g.model + "\" " + strconv.Itoa(ev.Count) + " times in the last " + strconv.Itoa(ev.WindowMinutes) + " minutes and got 404 each time: add an Alias",
			Evidence: &ev, Proposal: Proposal{Alias: g.model, Target: mostUsed(g.targets)},
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Application+"\x00"+out[i].Subject < out[j].Application+"\x00"+out[j].Subject
	})
	return out
}

func hasAlias(o config.Origin, model string) bool {
	for _, a := range o.Aliases {
		if a.Alias == model {
			return true
		}
	}
	return false
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
