package decide_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/catalog"
	"github.com/harelix/aiport/internal/decide"
	"github.com/harelix/aiport/internal/telemetry"
)

// removal is a Catalog change for docs-bot that removed the task name
// review-code.
func removal(changeID string) decide.CatalogNotice {
	return decide.CatalogNotice{
		Record: catalog.Record{ChangeID: changeID, At: t0.UnixMilli(), Change: catalog.Change{
			Application: "docs-bot",
			Removed:     []catalog.TaskName{{Name: "review-code", Connection: "hosted", Model: "model-x"}},
		}},
		Available:   []catalog.TaskName{{Name: "summarize", Connection: "local", Model: "model-s"}},
		Connections: []string{"local"},
	}
}

// moved is a Catalog change for docs-bot whose routing of kind of moved
// from before to after.
func moved(of catalog.Routed, before, after string) decide.CatalogNotice {
	return decide.CatalogNotice{
		Record: catalog.Record{ChangeID: "chg-route", At: t0.UnixMilli(), Change: catalog.Change{
			Application: "docs-bot",
			Routing:     []catalog.Switch{{Of: of, Before: before, After: after}},
		}},
		Available: []catalog.TaskName{{Name: "summarize", Connection: "local", Model: "model-s"}},
	}
}

func withProtocol(es []telemetry.Event, p string) []telemetry.Event {
	for i := range es {
		es[i].Protocol = p
	}
	return es
}

func TestCatalogChangeOpensAnIssueOnlyWhenTheApplicationUsedWhatWentAway(t *testing.T) {
	cases := []struct {
		name   string
		notice decide.CatalogNotice
		events []telemetry.Event
		issue  bool
	}{
		{"it sent the removed task name", removal("chg-1"), successes("docs-bot", "hosted", "review-code", 1, t0.Add(-time.Hour)), true},
		{"it never sent it", removal("chg-1"), successes("docs-bot", "local", "summarize", 3, t0.Add(-time.Hour)), false},
		{"another Application sent it", removal("chg-1"), successes("chat-helper", "hosted", "review-code", 3, t0.Add(-time.Hour)), false},
		{"longer ago than request history", removal("chg-1"), successes("docs-bot", "hosted", "review-code", 1, t0.Add(-8*24*time.Hour)), false},
		{"after the change", removal("chg-1"), successes("docs-bot", "hosted", "review-code", 1, t0.Add(time.Minute)), false},
		{"its binding moved and it used it", moved(catalog.RoutedBinding, "local", "hosted"), successes("docs-bot", "local", "some-model", 1, t0.Add(-time.Hour)), true},
		{"its binding moved, but it only sent task names", moved(catalog.RoutedBinding, "local", "hosted"), successes("docs-bot", "local", "summarize", 1, t0.Add(-time.Hour)), false},
		{"a protocol default moved and it used it", moved(catalog.RoutedOpenAI, "local", "hosted"), withProtocol(successes("docs-bot", "local", "some-model", 1, t0.Add(-time.Hour)), "openai"), true},
		{"a protocol default moved, on another protocol", moved(catalog.RoutedOpenAI, "local", "hosted"), withProtocol(successes("docs-bot", "local", "some-model", 1, t0.Add(-time.Hour)), "anthropic"), false},
		{"nothing was taken away (a new binding)", moved(catalog.RoutedBinding, "", "hosted"), successes("docs-bot", "hosted", "some-model", 1, t0.Add(-time.Hour)), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			f.history.add(tc.events...)
			f.q.CatalogChanged([]decide.CatalogNotice{tc.notice})
			issues := f.q.Issues(decide.IssueFilter{})
			if (len(issues) == 1) != tc.issue || len(issues) > 1 {
				t.Fatalf("issues: %+v", issues)
			}
			if tc.issue {
				is := issues[0]
				if is.Cause != decide.CauseCatalogChanged || is.Subject != tc.notice.ChangeID || is.State != decide.IssueOpen || is.Fix != nil || is.NeedsOwner || is.Catalog == nil {
					t.Fatalf("issue: %+v", is)
				}
				if len(pending(f.q)) != 0 {
					t.Fatal("a Catalog change waits for the Owner")
				}
			}
		})
	}
}

func TestCatalogIssueTexts(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(successes("docs-bot", "hosted", "review-code", 1, t0.Add(-time.Hour))...)
	n := removal("chg-1")
	n.Moved = []catalog.Move{{Name: "summarize", From: catalog.TaskName{Connection: "hosted", Model: "old"}, To: catalog.TaskName{Connection: "local", Model: "model-s"}}}
	n.Routing = []catalog.Switch{{Of: catalog.RoutedBinding, Before: "local", After: ""}, {Of: catalog.RoutedAnthropic, Before: "", After: "side"}}
	n.Connections = []string{"local", "hosted"}
	f.q.CatalogChanged([]decide.CatalogNotice{n})
	is := onlyIssue(t, f.q, "docs-bot")
	wantSummary := `A Change (chg-1) changed what docs-bot can use: removed task name "review-code"; the binding (every model name without an Alias) moved from Connection "local" to no Connection. The Application was told; nothing waits for the Owner.`
	if is.Summary != wantSummary {
		t.Fatalf("summary:\n got %q\nwant %q", is.Summary, wantSummary)
	}
	for _, want := range []string{
		"aiport Issue " + is.ID + ": the Owner changed what the Application docs-bot can use through aiport (Change chg-1 at ",
		`- Task name "review-code" was removed (it went to model "model-x" on Connection "hosted"); nothing replaces it.`,
		`- Task name "summarize" now goes to model "model-s" on Connection "local" (was "old" on "hosted").`,
		`- The binding (every model name without an Alias) now goes to no Connection (was Connection "local").`,
		`- The anthropic protocol default (model names without an Alias) now goes to Connection "side".`,
		`What docs-bot can still use: task names "summarize", and the models Connections local, hosted list.`,
		`stop sending "review-code"; send one of the names above instead.`,
		"/" + is.ID + "/fixed",
	} {
		if !strings.Contains(is.ApplicationFix, want) {
			t.Errorf("application fix lacks %q:\n%s", want, is.ApplicationFix)
		}
	}
	// Nothing for the Owner to decide but dismissing it.
	if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept}, t0); !errors.Is(err, decide.ErrNoFix) {
		t.Fatalf("accept: %v", err)
	}
	if got := decide.ApplicationFix(decide.Issue{Cause: decide.CauseCatalogChanged}); got != "" {
		t.Fatalf("a Catalog Issue without its change: %q", got)
	}
}

func TestCatalogIssueTextWithNothingLeft(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(successes("docs-bot", "local", "some-model", 1, t0.Add(-time.Hour))...)
	n := moved(catalog.RoutedBinding, "local", "hosted")
	n.Available, n.Connections = nil, []string{"hosted"}
	f.q.CatalogChanged([]decide.CatalogNotice{n})
	is := onlyIssue(t, f.q, "docs-bot")
	for _, want := range []string{
		"What docs-bot can still use: no task names, and the models Connection hosted list.",
		"stop sending model names the new routing does not serve;",
	} {
		if !strings.Contains(is.ApplicationFix, want) {
			t.Errorf("application fix lacks %q:\n%s", want, is.ApplicationFix)
		}
	}
}

func TestCatalogIssueLifecycle(t *testing.T) {
	cases := []struct {
		name    string
		resolve func(t *testing.T, f *fixture, is decide.Issue)
		how     decide.How
	}{
		{"the Change is reverted", func(t *testing.T, f *fixture, is decide.Issue) {
			r := removal("chg-2")
			r.RevertOf = "chg-1"
			f.q.CatalogChanged([]decide.CatalogNotice{r})
		}, decide.Stopped},
		{"the Application says it fixed it", func(t *testing.T, f *fixture, is decide.Issue) {
			if _, err := f.q.FixedByApplication("docs-bot", is.ID, t0.Add(time.Minute)); err != nil {
				t.Fatal(err)
			}
		}, decide.ByApplicationFix},
		{"its calls succeed after the last one that sent the removed name", func(t *testing.T, f *fixture, is decide.Issue) {
			f.history.add(calls("docs-bot", "hosted", "review-code", "upstream_error", 404, 1, t0.Add(time.Minute))...)
			f.history.add(successes("docs-bot", "local", "summarize", 1, t0.Add(30*time.Second))...) // before that failure
			f.q.Detect(t0.Add(2 * time.Minute))
			if is, _ := f.q.Issue(is.ID); is.State != decide.IssueOpen {
				t.Fatal("resolved by a success before the last failure")
			}
			f.history.add(successes("docs-bot", "local", "summarize", 1, t0.Add(3*time.Minute))...)
			f.q.Detect(t0.Add(4 * time.Minute))
		}, decide.Stopped},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			f.history.add(successes("docs-bot", "hosted", "review-code", 1, t0.Add(-time.Hour))...)
			f.q.CatalogChanged([]decide.CatalogNotice{removal("chg-1")})
			is := onlyIssue(t, f.q, "docs-bot")
			// The same Change told again opens nothing new.
			f.q.CatalogChanged([]decide.CatalogNotice{removal("chg-1")})
			onlyIssue(t, f.q, "docs-bot")
			// A failing call that sends the removed name is one of its errors.
			hint := calls("docs-bot", "hosted", "review-code", "upstream_error", 404, 1, t0)[0]
			if id, ok := f.q.HintIssue("docs-bot", hint); !ok || id != is.ID {
				t.Fatalf("hint: %q %v", id, ok)
			}
			tc.resolve(t, f, is)
			got := onlyIssue(t, f.q, "docs-bot")
			if got.State != decide.IssueResolved || got.Resolution == nil || got.Resolution.How != tc.how {
				t.Fatalf("issue: %+v %+v", got, got.Resolution)
			}
			if _, ok := f.q.HintIssue("docs-bot", hint); ok {
				t.Fatal("a resolved Issue is still hinted")
			}
		})
	}
}

// A Change that applied an Issue's aiport fix, or its revert, is told by
// that Issue: no Catalog Issue opens for it.
func TestCatalogChangeOfAnAiportFixOpensNoSecondIssue(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.history.add(successes("docs-bot", "hosted", "review-code", 1, t0.Add(-time.Hour))...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	_, ch, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept, Proposal: decide.Proposal{Target: "hosted"}}, t0)
	if err != nil {
		t.Fatal(err)
	}
	f.q.CatalogChanged([]decide.CatalogNotice{removal(ch.ID)})
	rev := removal("chg-revert")
	rev.RevertOf = ch.ID
	f.q.CatalogChanged([]decide.CatalogNotice{rev})
	f.q.CatalogChanged(nil) // nothing to tell
	onlyIssue(t, f.q, "docs-bot")
}
