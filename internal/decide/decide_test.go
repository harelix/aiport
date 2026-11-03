package decide_test

import (
	"errors"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/decide"
	"github.com/harelix/aiport/internal/store"
)

// changeRequest is a pending Change request from docs-bot.
func changeRequest(kind decide.Kind, p decide.Proposal) decide.Item {
	return decide.Item{Source: decide.FromRequest, Kind: kind, Application: "docs-bot", Subject: "request:" + string(kind) + p.Alias + p.Target + p.Model, Need: "a need", Proposal: p}
}

func TestSubmitMergesAPendingItemWithTheSameSubject(t *testing.T) {
	f := newFixture(t, "", true)
	ev := &decide.Evidence{Count: 3}
	first, err := f.q.Submit(decide.Item{Source: decide.FromSuggestion, Kind: decide.KindApplicationAlias, Application: "docs-bot", Subject: "quick", Need: "first", Evidence: ev}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == "" || first.Status != decide.Pending || first.Scope != decide.ScopeApplication || first.CreatedAt != t0.UnixMilli() {
		t.Fatalf("submitted: %+v", first)
	}
	// The same subject again: the evidence, proposal and need are replaced.
	later := t0.Add(time.Minute)
	again, _ := f.q.Submit(decide.Item{Source: decide.FromSuggestion, Kind: decide.KindApplicationAlias, Application: "docs-bot", Subject: "quick", Need: "second",
		Evidence: &decide.Evidence{Count: 9}, Proposal: decide.Proposal{Target: "hosted"}}, later)
	if again.ID != first.ID || again.Repeats != 1 || again.Need != "second" || again.Evidence.Count != 9 || again.Proposal.Target != "hosted" ||
		again.CreatedAt != t0.UnixMilli() || again.UpdatedAt != later.UnixMilli() {
		t.Fatalf("merged: %+v", again)
	}
	// What a submitter sets for status, id, times and decision is ignored.
	other, _ := f.q.Submit(decide.Item{ID: "chosen", Status: decide.Accepted, Decision: &decide.Verdict{}, Repeats: 5, Source: decide.FromSuggestion,
		Kind: decide.KindApplicationAlias, Application: "chat-helper", Subject: "quick", Scope: decide.ScopeEvery}, t0)
	if other.ID == "chosen" || other.ID == first.ID || other.Status != decide.Pending || other.Decision != nil || other.Repeats != 0 || other.Scope != decide.ScopeEvery {
		t.Fatalf("other: %+v", other)
	}
	// Two Suggestions fixing different Issues are not merged.
	a, _ := f.q.Submit(decide.Item{Source: decide.FromSuggestion, Kind: decide.KindApplicationAlias, Application: "docs-bot", Subject: "fast", Issue: "issue-a"}, t0)
	b, _ := f.q.Submit(decide.Item{Source: decide.FromSuggestion, Kind: decide.KindApplicationAlias, Application: "docs-bot", Subject: "fast", Issue: "issue-b"}, t0)
	if a.ID == b.ID {
		t.Fatal("fixes for two Issues merged")
	}
	if len(f.store.items) != 4 {
		t.Fatalf("store keeps %d items, want 4", len(f.store.items))
	}
}

func TestListFiltersNewestFirst(t *testing.T) {
	f := newFixture(t, "", false)
	sub := func(src decide.Source, app, subject string) decide.Item {
		it, err := f.q.Submit(decide.Item{Source: src, Kind: decide.KindApplicationAlias, Application: app, Subject: subject}, t0)
		if err != nil {
			t.Fatal(err)
		}
		return it
	}
	a := sub(decide.FromSuggestion, "docs-bot", "one")
	b := sub(decide.FromRequest, "docs-bot", "two")
	c := sub(decide.FromSuggestion, "chat-helper", "three")
	if _, _, err := f.q.Decide(a.ID, decide.Decision{Action: decide.Dismiss}, t0); err != nil {
		t.Fatal(err)
	}
	ids := func(items []decide.Item) []string {
		out := []string{}
		for _, it := range items {
			out = append(out, it.ID)
		}
		return out
	}
	cases := []struct {
		name string
		f    decide.Filter
		want []string
	}{
		{"all", decide.Filter{}, []string{c.ID, b.ID, a.ID}},
		{"one Application", decide.Filter{Application: "docs-bot"}, []string{b.ID, a.ID}},
		{"status", decide.Filter{Status: decide.Pending}, []string{c.ID, b.ID}},
		{"source", decide.Filter{Source: decide.FromSuggestion}, []string{c.ID, a.ID}},
		{"all three", decide.Filter{Application: "docs-bot", Status: decide.Dismissed, Source: decide.FromSuggestion}, []string{a.ID}},
		{"none", decide.Filter{Application: "nobody"}, []string{}},
	}
	for _, tc := range cases {
		if got := ids(f.q.List(tc.f)); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestTheLoopOffRefusesWrites(t *testing.T) {
	f := newFixture(t, "adapting: {disabled: true}\n", false)
	if _, err := f.q.Submit(decide.Item{Source: decide.FromSuggestion, Subject: "x"}, t0); !errors.Is(err, decide.ErrOff) {
		t.Fatalf("Submit: %v", err)
	}
	if _, _, err := f.q.Decide("any", decide.Decision{Action: decide.Accept}, t0); !errors.Is(err, decide.ErrOff) {
		t.Fatalf("Decide: %v", err)
	}
	if _, _, err := f.q.DecideIssue("any", decide.Decision{Action: decide.Dismiss}, t0); !errors.Is(err, decide.ErrOff) {
		t.Fatalf("DecideIssue: %v", err)
	}
	if _, err := f.q.FixedByApplication("docs-bot", "any", t0); !errors.Is(err, decide.ErrOff) {
		t.Fatalf("FixedByApplication: %v", err)
	}
	if _, err := f.q.Explain("any", decide.Explanation{Text: "x"}, t0); !errors.Is(err, decide.ErrOff) {
		t.Fatalf("Explain: %v", err)
	}
	if _, err := f.q.RecordBranch("any", decide.BranchFix{}, t0); !errors.Is(err, decide.ErrOff) {
		t.Fatalf("RecordBranch: %v", err)
	}
	if got := f.q.Detect(t0); got != nil {
		t.Fatalf("Detect: %v", got)
	}
	if got := f.q.Issues(decide.IssueFilter{}); len(got) != 0 {
		t.Fatalf("Issues: %v", got)
	}
	if _, ok := f.q.Issue("any"); ok {
		t.Fatal("Issue found")
	}
}

func TestAcceptAppliesTheChange(t *testing.T) {
	cases := []struct {
		name  string
		item  decide.Item
		d     decide.Decision
		what  string
		check func(t *testing.T, c config.Config)
	}{
		{
			name: "an Alias for the Application",
			item: changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "review-code", Target: "hosted", Model: "model-x"}),
			d:    decide.Decision{Action: decide.Accept},
			what: "alias review-code → hosted/model-x for docs-bot",
			check: func(t *testing.T, c config.Config) {
				o, _ := c.Application("docs-bot")
				if !slices.Contains(o.Aliases, config.OriginAlias{Alias: "review-code", Target: "hosted", Model: "model-x"}) {
					t.Fatalf("docs-bot aliases: %+v", o.Aliases)
				}
				if len(c.Aliases) != 0 {
					t.Fatalf("a shared Alias was added: %+v", c.Aliases)
				}
			},
		},
		{
			name: "the Owner widens it to a shared Alias and fills in the model",
			item: changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "review-code", Target: "hosted"}),
			d:    decide.Decision{Action: decide.Accept, Scope: decide.ScopeEvery, Proposal: decide.Proposal{Model: "model-y"}},
			what: "shared alias review-code → hosted/model-y for every Application",
			check: func(t *testing.T, c config.Config) {
				if !slices.ContainsFunc(c.Aliases, func(a config.Alias) bool {
					return a.Alias == "review-code" && a.Target == "hosted" && a.Model == "model-y"
				}) {
					t.Fatalf("shared aliases: %+v", c.Aliases)
				}
			},
		},
		{
			name: "the Owner's proposal overrides every field",
			item: changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "a", Target: "local", Model: "m"}),
			d:    decide.Decision{Action: decide.Accept, Proposal: decide.Proposal{Alias: "b", Target: "hosted", Model: "n"}},
			what: "alias b → hosted/n for docs-bot",
			check: func(t *testing.T, c config.Config) {
				o, _ := c.Application("docs-bot")
				if len(o.Aliases) != 1 || o.Aliases[0].Alias != "b" {
					t.Fatalf("docs-bot aliases: %+v", o.Aliases)
				}
			},
		},
		{
			name: "a Level for the Application",
			item: changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "fast", Target: "side", Model: "model-z"}),
			d:    decide.Decision{Action: decide.Accept},
			what: "Level fast (Anthropic) → side/model-z for docs-bot",
			check: func(t *testing.T, c config.Config) {
				m, ok := c.Level("docs-bot", "anthropic", "fast")
				if !ok || m.Shared || m.Connection != "side" || m.Model != "model-z" {
					t.Fatalf("level: %+v %v", m, ok)
				}
				if _, ok := c.Level("chat-helper", "anthropic", "fast"); ok {
					t.Fatal("chat-helper got the Level too")
				}
			},
		},
		{
			name: "a shared Level",
			item: changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "best", Target: "local", Model: "model-z"}),
			d:    decide.Decision{Action: decide.Accept, Scope: decide.ScopeEvery},
			what: "Level best (OpenAI) → local/model-z for every Application",
			check: func(t *testing.T, c config.Config) {
				if m, ok := c.Level("chat-helper", "openai", "best"); !ok || !m.Shared {
					t.Fatalf("level: %+v %v", m, ok)
				}
			},
		},
		{
			name: "a binding for the Application",
			item: changeRequest(decide.KindBindApplication, decide.Proposal{Target: "hosted"}),
			d:    decide.Decision{Action: decide.Accept},
			what: "bind docs-bot → hosted",
			check: func(t *testing.T, c config.Config) {
				if o, _ := c.Application("docs-bot"); o.Target != "hosted" {
					t.Fatalf("docs-bot target %q", o.Target)
				}
			},
		},
		{
			name: "a binding for every Application is the protocol default",
			item: changeRequest(decide.KindBindApplication, decide.Proposal{Target: "side"}),
			d:    decide.Decision{Action: decide.Accept, Scope: decide.ScopeEvery},
			what: "default anthropic Connection → side for every Application",
			check: func(t *testing.T, c config.Config) {
				if c.Defaults.AnthropicTarget != "side" || c.Defaults.OpenAITarget != "local" {
					t.Fatalf("defaults: %+v", c.Defaults)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			it, err := f.q.Submit(tc.item, t0)
			if err != nil {
				t.Fatal(err)
			}
			at := t0.Add(time.Minute)
			tc.d.Note = "approved"
			got, ch, err := f.q.Decide(it.ID, tc.d, at)
			if err != nil {
				t.Fatalf("accept: %v", err)
			}
			if ch.ID == "" || ch.What != tc.what || ch.By != config.Owner {
				t.Fatalf("change: id %q what %q by %q", ch.ID, ch.What, ch.By)
			}
			v := got.Decision
			if got.Status != decide.Accepted || v == nil || v.Action != decide.Accept || v.ChangeID != ch.ID || v.By != config.Owner || v.At != at.UnixMilli() || v.Note != "approved" {
				t.Fatalf("decided item: %+v %+v", got, v)
			}
			wantScope := tc.d.Scope
			if wantScope == "" {
				wantScope = decide.ScopeApplication
			}
			if v.Scope != wantScope {
				t.Fatalf("granted scope %q, want %q", v.Scope, wantScope)
			}
			tc.check(t, f.live.Config())
			// The Change is recorded in config.Live, so it can be reverted.
			if changes := f.live.Changes(); len(changes) != 1 || changes[0].ID != ch.ID {
				t.Fatalf("changes: %+v", changes)
			}
		})
	}
}

func TestAcceptRefusesWhatCannotBeApplied(t *testing.T) {
	cases := []struct {
		name   string
		item   decide.Item
		d      decide.Decision
		reason string // a part of the Invalid reason
	}{
		{"an Alias without a model", changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "a", Target: "local"}), decide.Decision{}, "choose the alias, the Connection and the model"},
		{"an Alias without an alias", changeRequest(decide.KindApplicationAlias, decide.Proposal{Target: "local", Model: "m"}), decide.Decision{}, "choose the alias"},
		{"an Alias to an unknown Connection", changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "a", Target: "nowhere", Model: "m"}), decide.Decision{}, ""},
		{"a Level on an unknown Connection", changeRequest(decide.KindApplicationAlias, decide.Proposal{Alias: "fast", Target: "nowhere", Model: "m"}), decide.Decision{}, "no Connection nowhere in configuration"},
		{"a binding without a Connection", changeRequest(decide.KindBindApplication, decide.Proposal{}), decide.Decision{}, "choose the Connection to bind to"},
		{"a default to an unknown Connection", changeRequest(decide.KindBindApplication, decide.Proposal{Target: "nowhere"}), decide.Decision{Scope: decide.ScopeEvery}, "no Connection nowhere in configuration"},
		{"a binding to an unknown Connection", changeRequest(decide.KindBindApplication, decide.Proposal{Target: "nowhere"}), decide.Decision{}, ""},
		{"a bad scope", changeRequest(decide.KindBindApplication, decide.Proposal{Target: "hosted"}), decide.Decision{Scope: "team"}, "scope must be application or every_application"},
		{"a kind aiport cannot apply", changeRequest("rate_limit", decide.Proposal{}), decide.Decision{}, "aiport cannot apply this kind of Change yet"},
		{"an informational Suggestion", decide.Item{Source: decide.FromSuggestion, Kind: decide.KindPairApplication, Subject: "unknown_token", Informational: true}, decide.Decision{}, "nothing to apply"},
		{"a binding Suggestion to the failing Connection", decide.Item{Source: decide.FromSuggestion, Kind: decide.KindBindApplication, Application: "docs-bot", Subject: "local"}, decide.Decision{Proposal: decide.Proposal{Target: "local"}}, "choose another Connection than the failing one"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			it, err := f.q.Submit(tc.item, t0)
			if err != nil {
				t.Fatal(err)
			}
			tc.d.Action = decide.Accept
			_, _, err = f.q.Decide(it.ID, tc.d, t0)
			if !isInvalid(err) || !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("want Invalid containing %q, got %v", tc.reason, err)
			}
			// Nothing was applied and the item still waits.
			if len(f.live.Changes()) != 0 {
				t.Fatal("a Change was applied")
			}
			if got := pending(f.q); len(got) != 1 || got[0].ID != it.ID {
				t.Fatalf("pending: %+v", got)
			}
			// It can still be dismissed.
			if _, _, err := f.q.Decide(it.ID, decide.Decision{Action: decide.Dismiss}, t0); err != nil {
				t.Fatalf("dismiss: %v", err)
			}
		})
	}
}

func TestDecideRefusals(t *testing.T) {
	f := newFixture(t, "", false)
	it, _ := f.q.Submit(changeRequest(decide.KindBindApplication, decide.Proposal{Target: "hosted"}), t0)
	cases := []struct {
		name  string
		id    string
		d     decide.Decision
		check func(error) bool
	}{
		{"unknown id", "missing", decide.Decision{Action: decide.Reject}, func(err error) bool { return errors.Is(err, decide.ErrNotFound) }},
		{"note too long", it.ID, decide.Decision{Action: decide.Reject, Note: strings.Repeat("n", 2001)}, isInvalid},
		{"no action", it.ID, decide.Decision{}, isInvalid},
		{"unknown action", it.ID, decide.Decision{Action: "postpone"}, isInvalid},
		{"automatic, for a Suggestion fixing no Issue", it.ID, decide.Decision{Action: decide.Accept, By: decide.Automatic}, func(err error) bool { return errors.Is(err, decide.ErrDecided) }},
	}
	for _, tc := range cases {
		if _, _, err := f.q.Decide(tc.id, tc.d, t0); !tc.check(err) {
			t.Errorf("%s: unexpected %v", tc.name, err)
		}
	}
	// A 2000-character note fits.
	got, ch, err := f.q.Decide(it.ID, decide.Decision{Action: decide.Reject, Note: strings.Repeat("n", 2000), By: "someone"}, t0)
	if err != nil || got.Status != decide.Rejected || got.Decision.By != "someone" || ch.ID != "" {
		t.Fatalf("reject: %+v %+v %v", got, ch, err)
	}
	if len(f.live.Changes()) != 0 {
		t.Fatal("reject changed the configuration")
	}
	if _, _, err := f.q.Decide(it.ID, decide.Decision{Action: decide.Accept}, t0); !errors.Is(err, decide.ErrDecided) {
		t.Fatalf("deciding twice: %v", err)
	}
}

func TestAcceptOnAReadOnlyConfiguration(t *testing.T) {
	f := newFixture(t, "", false)
	q, err := decide.New(config.NewLive("", f.live.Config()), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	it, _ := q.Submit(changeRequest(decide.KindBindApplication, decide.Proposal{Target: "hosted"}), t0)
	if _, _, err := q.Decide(it.ID, decide.Decision{Action: decide.Accept}, t0); !errors.Is(err, config.ErrReadOnly) {
		t.Fatalf("want ErrReadOnly, got %v", err)
	}
	if got := pending(q); len(got) != 1 {
		t.Fatal("the item no longer waits")
	}
	// Without history nothing is detected.
	if got := q.Detect(t0); got != nil {
		t.Fatalf("Detect without history: %v", got)
	}
}

// A request for every Application from two Applications, narrowed by the
// Owner to Scope application, gives each of them the Alias.
func TestNarrowedRequestGoesToEveryRequester(t *testing.T) {
	f := newFixture(t, "", false)
	body := request(t, map[string]any{"scope": "every_application", "alias": "review-code", "prefer_connection": "hosted", "prefer_model": "model-x"})
	a, err := f.q.File("docs-bot", body, t0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.q.File("chat-helper", body, t0); err != nil {
		t.Fatal(err)
	}
	_, ch, err := f.q.Decide(a.ID, decide.Decision{Action: decide.Accept, Scope: decide.ScopeApplication, Note: "only you two"}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if ch.What != "alias review-code → hosted/model-x for docs-bot, chat-helper" {
		t.Fatalf("what: %q", ch.What)
	}
	c := f.live.Config()
	for _, app := range []string{"docs-bot", "chat-helper"} {
		if o, _ := c.Application(app); len(o.Aliases) != 1 {
			t.Fatalf("%s aliases: %+v", app, o.Aliases)
		}
	}
	if o, _ := c.Application("quiet-app"); len(o.Aliases) != 0 {
		t.Fatal("quiet-app got the Alias")
	}
	// Both see the granted scope and the note.
	for _, app := range []string{"docs-bot", "chat-helper"} {
		ans, _ := f.q.RequestOf(app, a.ID)
		if ans.Status != decide.Accepted || ans.GrantedScope != decide.ScopeApplication || ans.Note != "only you two" || ans.ChangeID != ch.ID {
			t.Fatalf("%s sees %+v", app, ans)
		}
	}
}

func TestPruneForgetsDecidedItemsAfterRetention(t *testing.T) {
	f := newFixture(t, "telemetry: {retention_days: 1}\n", true)
	sub := func(subject string) decide.Item {
		it, _ := f.q.Submit(decide.Item{Source: decide.FromSuggestion, Kind: decide.KindBindApplication, Application: "docs-bot", Subject: subject}, t0)
		return it
	}
	waiting := sub("waiting")
	rejected := sub("rejected")
	dismissed := sub("dismissed")
	f.q.Decide(rejected.ID, decide.Decision{Action: decide.Reject}, t0)
	f.q.Decide(dismissed.ID, decide.Decision{Action: decide.Dismiss}, t0)
	notSupported, _ := f.q.Submit(decide.Item{Source: decide.FromRequest, Kind: "rate_limit", Application: "docs-bot", Subject: "ns", Status: decide.NotSupported}, t0)

	has := func(id string) bool {
		return slices.ContainsFunc(f.q.List(decide.Filter{}), func(it decide.Item) bool { return it.ID == id })
	}
	cases := []struct {
		at   time.Duration
		kept map[string]bool
	}{
		{23 * time.Hour, map[string]bool{waiting.ID: true, rejected.ID: true, dismissed.ID: true, notSupported.ID: true}},
		{25 * time.Hour, map[string]bool{waiting.ID: true, rejected.ID: false, dismissed.ID: true, notSupported.ID: true}},
		// Dismissed items are kept at least QuietFor.
		{decide.QuietFor + time.Hour, map[string]bool{waiting.ID: true, dismissed.ID: false, notSupported.ID: true}},
	}
	for _, c := range cases {
		f.q.Prune(t0.Add(c.at))
		for id, want := range c.kept {
			if has(id) != want {
				t.Errorf("after %v: item %s kept=%v, want %v", c.at, id, has(id), want)
			}
		}
	}
	if !reflect.DeepEqual(f.store.deleted, []string{rejected.ID, dismissed.ID}) {
		t.Fatalf("deleted from the store: %v", f.store.deleted)
	}
}

func TestNewLoadsWhatTheStoreKept(t *testing.T) {
	f := newFixture(t, "", true)
	it, _ := f.q.Submit(changeRequest(decide.KindBindApplication, decide.Proposal{Target: "hosted"}), t0)
	f.store.issues["kept-issue"] = decide.Issue{ID: "kept-issue", Application: "docs-bot", Cause: decide.CauseUnknownModel, Subject: "m", State: decide.IssueOpen}

	q, err := decide.New(f.live, nil, f.store)
	if err != nil {
		t.Fatal(err)
	}
	if got := q.List(decide.Filter{}); len(got) != 1 || got[0].ID != it.ID {
		t.Fatalf("items: %+v", got)
	}
	if _, ok := q.Issue("kept-issue"); !ok {
		t.Fatal("the kept Issue is missing")
	}
	// A Store that keeps no Issues still keeps items.
	q2, err := decide.New(f.live, nil, itemsOnly{f.store})
	if err != nil || len(q2.List(decide.Filter{})) != 1 || len(q2.Issues(decide.IssueFilter{})) != 0 {
		t.Fatalf("items-only store: %v", err)
	}

	f.store.issuesErr = errors.New("issues unreadable")
	if _, err := decide.New(f.live, nil, f.store); err == nil {
		t.Fatal("an Issues load error was ignored")
	}
	f.store.loadErr = errors.New("items unreadable")
	if _, err := decide.New(f.live, nil, f.store); err == nil {
		t.Fatal("an items load error was ignored")
	}
}

func TestSQLStoreRoundTrip(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "aiport.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	s := decide.SQLStore{DB: db}
	a := decide.Item{ID: "a", Source: decide.FromRequest, Kind: decide.KindBindApplication, Application: "docs-bot", Subject: "s", Status: decide.Pending, CreatedAt: 2}
	b := decide.Item{ID: "b", Source: decide.FromSuggestion, Kind: decide.KindApplicationAlias, Application: "chat-helper", Subject: "t", Status: decide.Pending, CreatedAt: 1,
		Evidence: &decide.Evidence{Count: 4, Examples: []string{"req-1"}}}
	for _, it := range []decide.Item{a, b} {
		if err := s.Save(it); err != nil {
			t.Fatal(err)
		}
	}
	a.Status, a.Decision = decide.Rejected, &decide.Verdict{Action: decide.Reject, By: "owner", At: 5}
	if err := s.Save(a); err != nil { // replaces by ID
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil || !reflect.DeepEqual(got, []decide.Item{b, a}) {
		t.Fatalf("load (oldest first): %+v %v", got, err)
	}
	if err := s.Delete([]string{"b"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Load(); len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("after delete: %+v", got)
	}

	is := decide.Issue{ID: "i", Application: "docs-bot", Cause: decide.CauseUnknownModel, Subject: "m", State: decide.IssueOpen, CreatedAt: 3,
		Evidence: decide.Evidence{Count: 5}, NeverAuto: []decide.Proposal{{Alias: "m", Target: "local", Model: "m"}}}
	if err := s.SaveIssue(is); err != nil {
		t.Fatal(err)
	}
	is.State = decide.IssueResolved
	if err := s.SaveIssue(is); err != nil {
		t.Fatal(err)
	}
	issues, err := s.LoadIssues()
	if err != nil || !reflect.DeepEqual(issues, []decide.Issue{is}) {
		t.Fatalf("issues: %+v %v", issues, err)
	}
	if err := s.DeleteIssues([]string{"i"}); err != nil {
		t.Fatal(err)
	}
	if issues, _ := s.LoadIssues(); len(issues) != 0 {
		t.Fatalf("after delete: %+v", issues)
	}
	// A closed database fails without detail.
	db.Close()
	if _, err := s.Load(); err == nil {
		t.Fatal("Load on a closed database")
	}
	if err := s.Save(a); err == nil {
		t.Fatal("Save on a closed database")
	}
	if _, err := s.LoadIssues(); err == nil {
		t.Fatal("LoadIssues on a closed database")
	}
	if err := s.SaveIssue(is); err == nil {
		t.Fatal("SaveIssue on a closed database")
	}
	if err := s.Delete([]string{"a"}); err == nil {
		t.Fatal("Delete on a closed database")
	}
	if err := s.DeleteIssues([]string{"a"}); err == nil {
		t.Fatal("DeleteIssues on a closed database")
	}
}

func TestUnknownTokensSuggestPairing(t *testing.T) {
	f := newFixture(t, "", false)
	f.q.WatchFailedAuth(failedAuth{n: 9, first: 1, last: 2})
	if got := f.q.Detect(t0); len(got) != 0 {
		t.Fatalf("below the threshold: %+v", got)
	}
	f.q.WatchFailedAuth(failedAuth{n: 10, first: 1, last: 2})
	got := f.q.Detect(t0)
	if len(got) != 1 {
		t.Fatalf("want one Suggestion, got %+v", got)
	}
	s := got[0]
	if s.Kind != decide.KindPairApplication || !s.Informational || s.Opens != decide.OpensApplications || s.Application != "" ||
		s.Evidence == nil || !reflect.DeepEqual(*s.Evidence, decide.Evidence{Count: 10, FromMS: 1, ToMS: 2, WindowMinutes: 60}) || !strings.Contains(s.Need, "refused 10 times in the last 60 minutes") {
		t.Fatalf("suggestion: %+v", s)
	}
	// Another pass updates the same one.
	if again := f.q.Detect(t0.Add(time.Minute)); len(again) != 1 || again[0].ID != s.ID || again[0].Repeats != 1 {
		t.Fatalf("again: %+v", again)
	}
	if _, _, err := f.q.Decide(s.ID, decide.Decision{Action: decide.Dismiss}, t0); err != nil {
		t.Fatal(err)
	}
	// Dismissed: quiet unless the evidence doubles or QuietFor passes.
	cases := []struct {
		name  string
		n     int
		after time.Duration
		back  bool
	}{
		{"same evidence", 10, time.Hour, false},
		{"almost doubled", 19, time.Hour, false},
		{"doubled", 20, time.Hour, true},
	}
	for _, c := range cases {
		f.q.WatchFailedAuth(failedAuth{n: c.n})
		got := f.q.Detect(t0.Add(c.after))
		if (len(got) == 1) != c.back {
			t.Fatalf("%s: %+v", c.name, got)
		}
	}
	// The doubled one is pending again; dismiss it and wait out QuietFor.
	p := pending(f.q)
	if len(p) != 1 {
		t.Fatalf("pending: %+v", p)
	}
	f.q.Decide(p[0].ID, decide.Decision{Action: decide.Dismiss}, t0.Add(time.Hour))
	f.q.WatchFailedAuth(failedAuth{n: 20})
	if got := f.q.Detect(t0.Add(time.Hour + decide.QuietFor - time.Minute)); len(got) != 0 {
		t.Fatalf("within QuietFor: %+v", got)
	}
	if got := f.q.Detect(t0.Add(time.Hour + decide.QuietFor)); len(got) != 1 {
		t.Fatalf("after QuietFor: %+v", got)
	}
}

func TestRunDetectsUntilStopped(t *testing.T) {
	f := newFixture(t, "", false)
	f.q.WatchFailedAuth(failedAuth{n: 10})
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		f.q.Run(stop, func() time.Duration { return time.Millisecond }, func() time.Time { return t0 })
		close(done)
	}()
	deadline := time.Now().Add(5 * time.Second)
	for len(pending(f.q)) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("Run never detected")
		}
		time.Sleep(time.Millisecond)
	}
	close(stop)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop")
	}
}

// The refusal says a note may have 2,000 characters, and a Change request's
// own fields are counted in characters, but Decide (and DecideIssue's
// dismissal) count the note's bytes: a 2,000-character note in a script
// outside ASCII is refused.
func TestANoteOf2000CharactersFits(t *testing.T) {
	t.Skip("BUG: Decide and DecideIssue check len(d.Note) > 2000 (bytes) while the refusal says \"at most 2,000 characters\"; a 2,000-character non-ASCII note is refused")
	f := newFixture(t, "", false)
	it, _ := f.q.Submit(changeRequest(decide.KindBindApplication, decide.Proposal{Target: "hosted"}), t0)
	if _, _, err := f.q.Decide(it.ID, decide.Decision{Action: decide.Reject, Note: strings.Repeat("é", 2000)}, t0); err != nil {
		t.Fatalf("a 2,000-character note: %v", err)
	}
}
