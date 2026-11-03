package decide_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/decide"
	"github.com/harelix/aiport/internal/telemetry"
)

// errorsAgo is where a pass's errors start: well inside the default
// 60-minute window before t0.
var errorsAgo = t0.Add(-50 * time.Minute)

func levelRefusals(app, level, route string, n int, start time.Time) []telemetry.Event {
	out := calls(app, "", level, "rejected", 400, n, start)
	for i := range out {
		out[i].Route, out[i].RefusalCode = route, "level_unmapped"
	}
	return out
}

func TestUnknownModelOpensAnIssueAtTheThreshold(t *testing.T) {
	cases := []struct {
		name   string
		extra  string
		events []telemetry.Event
		issue  bool
		alias  bool // docs-bot has an Alias for the model
	}{
		{"below the threshold", "", notFounds("docs-bot", "quick", 4, errorsAgo), false, false},
		{"at the threshold", "", notFounds("docs-bot", "quick", 5, errorsAgo), true, false},
		{"outside the window", "", notFounds("docs-bot", "quick", 9, t0.Add(-3*time.Hour)), false, false},
		{"an unknown Application", "", notFounds("stranger", "quick", 9, errorsAgo), false, false},
		{"a model the Application has an Alias for", "", notFounds("docs-bot", "quick", 9, errorsAgo), false, true},
		{"not inference", "", func() []telemetry.Event {
			es := notFounds("docs-bot", "quick", 9, errorsAgo)
			for i := range es {
				es[i].Route = "/v1/models"
			}
			return es
		}(), false, false},
		{"other errors are not an unknown model", "", calls("docs-bot", "local", "quick", "upstream_error", 400, 9, errorsAgo), false, false},
		{"a lower threshold", "adapting: {unknown_model: {count: 2, window_minutes: 60}}\n", notFounds("docs-bot", "quick", 2, errorsAgo), true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, tc.extra, false)
			if tc.alias {
				f.setConfig(t, config.AddOriginAlias("docs-bot", config.OriginAlias{Alias: "quick", Target: "hosted", Model: "quick"}))
			}
			f.history.add(tc.events...)
			out := f.q.Detect(t0)
			issues := f.q.Issues(decide.IssueFilter{})
			if (len(issues) == 1) != tc.issue || len(issues) > 1 {
				t.Fatalf("issues: %+v", issues)
			}
			if !tc.issue {
				if len(out) != 0 || len(pending(f.q)) != 0 {
					t.Fatalf("suggestions without an Issue: %+v", out)
				}
				return
			}
			if len(out) != 1 || out[0].Issue != issues[0].ID {
				t.Fatalf("offered: %+v", out)
			}
		})
	}
}

func TestUnknownModelIssueAndItsFix(t *testing.T) {
	f := newFixture(t, "", true)
	errs := notFounds("docs-bot", "quick", 7, errorsAgo)
	errs[6].Target, errs[6].UpstreamErrorType = "hosted", "not_found_error"
	f.history.add(errs...)
	f.history.add(notFounds("chat-helper", "quick", 2, errorsAgo)...) // another Application, too few
	out := f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	if len(f.q.Issues(decide.IssueFilter{Application: "chat-helper"})) != 0 {
		t.Fatal("chat-helper has an Issue")
	}
	wantEv := decide.Evidence{
		Count: 7, FromMS: errs[0].StartedAt, ToMS: errs[6].StartedAt, WindowMinutes: 60, Status: 404, Error: "not_found_error",
		Examples:    []string{errs[6].ID, errs[5].ID, errs[4].ID, errs[3].ID, errs[2].ID},
		Connections: []string{"hosted", "local"},
	}
	if !reflect.DeepEqual(is.Evidence, wantEv) {
		t.Fatalf("evidence:\n got %+v\nwant %+v", is.Evidence, wantEv)
	}
	if is.Cause != decide.CauseUnknownModel || is.Subject != "quick" || is.State != decide.IssueOpen || !is.NeedsOwner || is.CreatedAt != t0.UnixMilli() {
		t.Fatalf("issue: %+v", is)
	}
	if want := `docs-bot asked for model "quick" 7 times in the last 60 minutes and got 404 not_found_error each time. Choose the Connection that serves it: add an Alias.`; is.Summary != want {
		t.Fatalf("summary: %q", is.Summary)
	}
	if len(out) != 1 {
		t.Fatalf("offered: %+v", out)
	}
	s := out[0]
	if is.Fix == nil || is.Fix.Suggestion != s.ID || is.Fix.State != decide.FixProposed || is.Fix.Kind != decide.KindApplicationAlias ||
		is.Fix.Proposal != (decide.Proposal{Alias: "quick", Model: "quick"}) {
		t.Fatalf("fix: %+v", is.Fix)
	}
	if s.Source != decide.FromSuggestion || s.Application != "docs-bot" || s.Subject != "quick" || s.Need != is.Summary || s.Evidence.Count != 7 {
		t.Fatalf("suggestion: %+v", s)
	}
	for _, want := range []string{"aiport Issue " + is.ID, `for model "quick" fail with 404 not_found_error`, "Example request ids: " + errs[6].ID, "Do not retry \"quick\" unchanged", "How to confirm"} {
		if !strings.Contains(is.ApplicationFix, want) {
			t.Errorf("application fix lacks %q:\n%s", want, is.ApplicationFix)
		}
	}
	if f.store.issues[is.ID].ID != is.ID {
		t.Fatal("the Issue was not stored")
	}
	// A second pass over the same history updates the same Issue and
	// Suggestion.
	again := f.q.Detect(t0.Add(time.Minute))
	if len(again) != 1 || again[0].ID != s.ID || again[0].Repeats != 1 || len(pending(f.q)) != 1 {
		t.Fatalf("second pass: %+v", again)
	}
	if is2 := onlyIssue(t, f.q, "docs-bot"); is2.ID != is.ID {
		t.Fatal("a second Issue opened")
	}
}

func TestDiagnosisShapesTheFix(t *testing.T) {
	cases := []struct {
		name    string
		d       decide.Diagnosis
		target  string
		summary string
		diag    string
	}{
		{"one Connection serves it", decide.Diagnosis{Answer: decide.ServedByOne, Serving: []string{"hosted"}, Available: []string{"quick", "slow"}}, "hosted",
			`Only Connection "hosted" serves it: add an Alias for docs-bot to it.`, `Connection "hosted" serves "quick" (from the Connections' free model lists).`},
		{"several serve it", decide.Diagnosis{Answer: decide.ServedBySeveral, Serving: []string{"hosted", "local"}}, "",
			"Connections hosted, local serve it: choose one for an Alias.", `Connections hosted, local serve "quick"`},
		{"none serves it", decide.Diagnosis{Answer: decide.ServedByNone}, "",
			"No enabled Connection serves it", `no enabled Connection of this gateway serves "quick"`},
		{"a listing failed", decide.Diagnosis{Answer: decide.ServedUnknown, Unknown: []string{"local"}}, "",
			"aiport could not list the models of local", "not every Connection could be listed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			fd := &finder{by: map[string]decide.Diagnosis{"quick": tc.d}}
			f.q.Diagnose(fd)
			f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
			f.history.add(notFounds("docs-bot", "rare", 2, errorsAgo)...) // too few to diagnose
			f.q.Detect(t0)
			if !reflect.DeepEqual(fd.asked, []string{"openai/quick"}) {
				t.Fatalf("asked: %v", fd.asked)
			}
			is := onlyIssue(t, f.q, "docs-bot")
			d := is.Diagnosis
			if d == nil || d.At != t0.UnixMilli() || d.Serving == nil || d.Unknown == nil || !strings.Contains(d.Summary, tc.diag) {
				t.Fatalf("diagnosis: %+v", d)
			}
			if is.Fix.Proposal.Target != tc.target || !strings.Contains(is.Summary, tc.summary) {
				t.Fatalf("fix %+v, summary %q", is.Fix.Proposal, is.Summary)
			}
			if !strings.Contains(is.ApplicationFix, "What aiport found: "+d.Summary) {
				t.Fatalf("application fix lacks the diagnosis:\n%s", is.ApplicationFix)
			}
		})
	}
}

func TestDiagnosisAvailableModelsAreBounded(t *testing.T) {
	f := newFixture(t, "", false)
	many := make([]string, 30)
	for i := range many {
		many[i] = "model-" + string(rune('a'+i%26)) + string(rune('a'+i/26))
	}
	f.q.Diagnose(&finder{by: map[string]decide.Diagnosis{"quick": {Answer: decide.ServedByNone, Available: many}}})
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	if len(is.Diagnosis.Available) != 20 || !strings.Contains(is.ApplicationFix, "Models the openai Connections of this gateway list: model-aa") {
		t.Fatalf("available: %v\n%s", is.Diagnosis.Available, is.ApplicationFix)
	}
}

func TestAcceptedFixResolvesTheIssueOnceCallsSucceed(t *testing.T) {
	f := newFixture(t, "", false)
	f.q.Diagnose(&finder{by: map[string]decide.Diagnosis{"quick": {Answer: decide.ServedByOne, Serving: []string{"hosted"}}}})
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")

	got, ch, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != decide.IssueResolving || got.ResolvingBy != decide.ByAiportFix || got.NeedsOwner ||
		got.Fix.State != decide.FixApplied || got.Fix.ChangeID != ch.ID || got.Fix.By != config.Owner || got.Fix.Proposal.Target != "hosted" {
		t.Fatalf("after accept: %+v %+v", got, got.Fix)
	}
	if o, _ := f.live.Config().Application("docs-bot"); !slices.Contains(o.Aliases, config.OriginAlias{Alias: "quick", Target: "hosted", Model: "quick"}) {
		t.Fatalf("aliases: %+v", o.Aliases)
	}
	if !strings.Contains(got.ApplicationFix, `Already fixed in aiport: calls from docs-bot for "quick" now go to Connection "hosted"`) {
		t.Fatalf("application fix:\n%s", got.ApplicationFix)
	}
	if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept}, t0); !errors.Is(err, decide.ErrNoFix) {
		t.Fatalf("accepting again: %v", err)
	}
	// No call yet: the Issue keeps resolving.
	f.q.Detect(t0.Add(time.Minute))
	if is := onlyIssue(t, f.q, "docs-bot"); is.State != decide.IssueResolving {
		t.Fatalf("before any call: %s", is.State)
	}
	f.history.add(successes("docs-bot", "hosted", "quick", 1, t0.Add(2*time.Minute))...)
	f.q.Detect(t0.Add(3 * time.Minute))
	is = onlyIssue(t, f.q, "docs-bot")
	if is.State != decide.IssueResolved || is.Resolution == nil || *is.Resolution != (decide.Resolution{How: decide.ByAiportFix, At: t0.Add(3 * time.Minute).UnixMilli(), ChangeID: ch.ID}) {
		t.Fatalf("resolved: %+v %+v", is, is.Resolution)
	}
	if items := f.q.List(decide.Filter{}); len(items) != 1 || items[0].Status != decide.Accepted {
		t.Fatalf("the Suggestion: %+v", items)
	}
	if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Dismiss}, t0); !errors.Is(err, decide.ErrDecided) {
		t.Fatalf("deciding a resolved Issue: %v", err)
	}
	if _, _, err := f.q.DecideIssue("missing", decide.Decision{Action: decide.Dismiss}, t0); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("an unknown Issue: %v", err)
	}
}

func TestRejectedFixIsOfferedAgainButNeverAutomatically(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	got, ch, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Reject, Note: "not that"}, t0)
	if err != nil || ch.ID != "" {
		t.Fatalf("reject: %v %+v", err, ch)
	}
	if got.State != decide.IssueOpen || got.Fix.State != decide.FixRejected || got.NeedsOwner ||
		!reflect.DeepEqual(got.NeverAuto, []decide.Proposal{{Alias: "quick", Model: "quick"}}) {
		t.Fatalf("after reject: %+v", got)
	}
	// A rejected fix waits for nothing: only dismissing the Issue is left.
	if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept}, t0); !errors.Is(err, decide.ErrNoFix) {
		t.Fatalf("accept after reject: %v", err)
	}
	// While the error sticks the fix is offered again, as a new Suggestion.
	f.history.add(notFounds("docs-bot", "quick", 1, t0)...)
	out := f.q.Detect(t0.Add(time.Minute))
	is = onlyIssue(t, f.q, "docs-bot")
	if len(out) != 1 || is.Fix.State != decide.FixProposed || is.Fix.Suggestion != out[0].ID || len(is.NeverAuto) != 1 {
		t.Fatalf("offered again: %+v %+v", out, is.Fix)
	}
	if len(f.q.List(decide.Filter{})) != 2 {
		t.Fatal("want the rejected and the new Suggestion")
	}
}

func TestDismissingAnIssue(t *testing.T) {
	cases := []struct {
		name string
		// how dismisses the Issue: through its waiting fix, or after its
		// fix was rejected (nothing waits).
		rejectFirst bool
	}{
		{"its fix waits", false},
		{"nothing waits", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
			f.q.Detect(t0)
			is := onlyIssue(t, f.q, "docs-bot")
			if tc.rejectFirst {
				f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Reject}, t0)
				if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Dismiss, Note: strings.Repeat("n", 2001)}, t0); !isInvalid(err) {
					t.Fatalf("a long note: %v", err)
				}
			}
			got, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Dismiss, Note: "known"}, t0)
			if err != nil {
				t.Fatal(err)
			}
			if got.State != decide.IssueDismissed || got.Dismissal == nil || got.Dismissal.EvidenceCount != 5 || got.Dismissal.By != config.Owner || got.Dismissal.Note != "known" {
				t.Fatalf("dismissed: %+v %+v", got, got.Dismissal)
			}
			// The same evidence keeps it quiet; doubled evidence reopens it.
			f.q.Detect(t0.Add(time.Minute))
			if is := onlyIssue(t, f.q, "docs-bot"); is.State != decide.IssueDismissed {
				t.Fatalf("same evidence: %s", is.State)
			}
			f.history.add(notFounds("docs-bot", "quick", 5, t0.Add(-40*time.Minute))...)
			out := f.q.Detect(t0.Add(2 * time.Minute))
			is = onlyIssue(t, f.q, "docs-bot")
			if is.State != decide.IssueOpen || is.Reopened != 1 || is.Fix == nil || is.Fix.State != decide.FixProposed || len(out) != 1 {
				t.Fatalf("doubled: %+v", is)
			}
		})
	}
}

func TestDismissedIssueComesBackAfterQuietFor(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Dismiss}, t0)
	later := t0.Add(decide.QuietFor)
	f.history.add(notFounds("docs-bot", "quick", 5, later.Add(-30*time.Minute))...)
	f.q.Detect(later)
	if is := onlyIssue(t, f.q, "docs-bot"); is.State != decide.IssueOpen || is.Reopened != 1 {
		t.Fatalf("after QuietFor: %+v", is)
	}
}

func TestIssuesResolveThemselves(t *testing.T) {
	cases := []struct {
		name  string
		after []telemetry.Event
		at    time.Time
	}{
		{"a call for the model succeeds after the last error", successes("docs-bot", "local", "quick", 1, t0.Add(-time.Minute)), t0},
		{"a whole window without the error", nil, errorsAgo.Add(4*time.Minute + time.Hour)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
			out := f.q.Detect(t0.Add(-2 * time.Minute))
			f.history.add(tc.after...)
			f.q.Detect(tc.at)
			is := onlyIssue(t, f.q, "docs-bot")
			if is.State != decide.IssueResolved || is.Resolution.How != decide.Stopped || is.Fix != nil || is.NeedsOwner {
				t.Fatalf("issue: %+v", is)
			}
			// Its fix no longer waits for the Owner.
			items := f.q.List(decide.Filter{})
			if len(items) != 1 || items[0].ID != out[0].ID || items[0].Status != decide.Withdrawn {
				t.Fatalf("suggestion: %+v", items)
			}
			if _, _, err := f.q.Decide(items[0].ID, decide.Decision{Action: decide.Accept}, tc.at); !errors.Is(err, decide.ErrDecided) {
				t.Fatalf("deciding a withdrawn Suggestion: %v", err)
			}
			// A resolved Issue reopens only when the error sticks again
			// after it resolved.
			f.history.add(notFounds("docs-bot", "quick", 4, tc.at.Add(time.Second))...)
			f.q.Detect(tc.at.Add(5 * time.Minute))
			if is := onlyIssue(t, f.q, "docs-bot"); is.State != decide.IssueResolved {
				t.Fatalf("below the threshold: %s", is.State)
			}
			f.history.add(notFounds("docs-bot", "quick", 1, tc.at.Add(5*time.Minute))...)
			f.q.Detect(tc.at.Add(6 * time.Minute))
			if is := onlyIssue(t, f.q, "docs-bot"); is.State != decide.IssueOpen || is.Reopened != 1 || is.Resolution != nil {
				t.Fatalf("reopened: %+v", is)
			}
		})
	}
}

func TestAutomaticFix(t *testing.T) {
	f := newFixture(t, "adapting: {auto_fix: true}\n", false)
	f.q.Diagnose(&finder{by: map[string]decide.Diagnosis{
		"quick": {Answer: decide.ServedByOne, Serving: []string{"hosted"}},
		"vague": {Answer: decide.ServedBySeveral, Serving: []string{"hosted", "local"}},
	}})
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.history.add(notFounds("docs-bot", "vague", 5, errorsAgo)...)
	out := f.q.Detect(t0)
	byModel := map[string]decide.Item{}
	for _, it := range out {
		byModel[it.Subject] = it
	}
	// Certain: applied by aiport itself. Ambiguous: left for the Owner.
	if a := byModel["quick"]; a.Status != decide.Accepted || a.Decision.By != decide.Automatic {
		t.Fatalf("quick: %+v", a)
	}
	if v := byModel["vague"]; v.Status != decide.Pending {
		t.Fatalf("vague: %+v", v)
	}
	var quick decide.Issue
	for _, is := range f.q.Issues(decide.IssueFilter{}) {
		if is.Subject == "quick" {
			quick = is
		}
	}
	if quick.State != decide.IssueResolving || quick.Fix.By != decide.Automatic {
		t.Fatalf("quick Issue: %+v", quick)
	}
	changes := f.live.Changes()
	if len(changes) != 1 || changes[0].By != decide.Automatic {
		t.Fatalf("changes: %+v", changes)
	}
	// The Owner reverts it: the Issue opens again, its fix is offered again
	// and never applied automatically again.
	if _, _, err := f.live.Revert(changes[0].ID); err != nil {
		t.Fatal(err)
	}
	f.q.Detect(t0.Add(time.Minute))
	quick, _ = f.q.Issue(quick.ID)
	if quick.State != decide.IssueOpen || quick.Fix == nil || quick.Fix.State != decide.FixProposed ||
		!reflect.DeepEqual(quick.NeverAuto, []decide.Proposal{{Alias: "quick", Target: "hosted", Model: "quick"}}) {
		t.Fatalf("after revert: %+v %+v", quick, quick.Fix)
	}
	if n := len(f.live.Changes()); n != 2 { // the Change and its revert
		t.Fatalf("changes after revert: %d", n)
	}
	// The Owner may still approve it by hand.
	if _, _, err := f.q.DecideIssue(quick.ID, decide.Decision{Action: decide.Accept}, t0.Add(2*time.Minute)); err != nil {
		t.Fatalf("accepting by hand: %v", err)
	}
}

// An automatic decision on a Suggestion whose Issue is no longer open is
// refused, so it never overrides what moved the Issue on.
func TestAutomaticDecisionNeedsAnOpenIssue(t *testing.T) {
	f := newFixture(t, "adapting: {auto_fix: true}\n", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.q.Detect(t0) // not diagnosed: nothing automatic
	is := onlyIssue(t, f.q, "docs-bot")
	if is.State != decide.IssueOpen {
		t.Fatalf("state %s", is.State)
	}
	if _, err := f.q.FixedByApplication("docs-bot", is.ID, t0); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.q.Decide(is.Fix.Suggestion, decide.Decision{Action: decide.Accept, By: decide.Automatic, Proposal: decide.Proposal{Target: "hosted"}}, t0); !errors.Is(err, decide.ErrDecided) {
		t.Fatalf("automatic on a resolving Issue: %v", err)
	}
	if len(f.live.Changes()) != 0 {
		t.Fatal("a Change was applied")
	}
}

func TestFixedByApplication(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")

	for _, app := range []string{"chat-helper", ""} {
		if _, err := f.q.FixedByApplication(app, is.ID, t0); !errors.Is(err, decide.ErrNotFound) {
			t.Fatalf("by %q: %v", app, err)
		}
	}
	if _, err := f.q.FixedByApplication("docs-bot", "missing", t0); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("an unknown Issue: %v", err)
	}
	a, err := f.q.FixedByApplication("docs-bot", is.ID, t0)
	if err != nil {
		t.Fatal(err)
	}
	if a.State != decide.IssueResolving || a.ResolvingBy != decide.ByApplicationFix || a.SinceMS != t0.UnixMilli() || a.AiportFix == nil || a.AiportFix.State != decide.FixProposed {
		t.Fatalf("agent view: %+v", a)
	}
	// Saying it twice is fine.
	if again, err := f.q.FixedByApplication("docs-bot", is.ID, t0.Add(time.Minute)); err != nil || again.SinceMS != t0.UnixMilli() {
		t.Fatalf("again: %+v %v", again, err)
	}
	// The aiport fix still waits for the Owner.
	if is := onlyIssue(t, f.q, "docs-bot"); !is.NeedsOwner || len(pending(f.q)) != 1 {
		t.Fatalf("the fix stopped waiting: %+v", is)
	}
	// Silence never confirms the Application's word.
	f.q.Detect(t0.Add(2 * time.Hour))
	if is := onlyIssue(t, f.q, "docs-bot"); is.State != decide.IssueResolving {
		t.Fatalf("after a quiet window: %s", is.State)
	}
	// A successful call does.
	f.history.add(successes("docs-bot", "local", "quick", 1, t0.Add(2*time.Hour))...)
	f.q.Detect(t0.Add(2*time.Hour + time.Minute))
	is = onlyIssue(t, f.q, "docs-bot")
	if is.State != decide.IssueResolved || is.Resolution.How != decide.ByApplicationFix {
		t.Fatalf("after a success: %+v", is)
	}
	if _, err := f.q.FixedByApplication("docs-bot", is.ID, t0); !errors.Is(err, decide.ErrNotOpen) {
		t.Fatalf("a resolved Issue: %v", err)
	}
}

// When the Application said it fixed an Issue but the error sticks again,
// the Issue opens again and its waiting fix is updated, not duplicated.
func TestApplicationFixThatDidNotWork(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	first := f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	f.q.FixedByApplication("docs-bot", is.ID, t0)
	// Errors from before it said so don't count.
	f.q.Detect(t0.Add(time.Minute))
	if is := onlyIssue(t, f.q, "docs-bot"); is.State != decide.IssueResolving {
		t.Fatalf("old errors moved it: %s", is.State)
	}
	f.history.add(notFounds("docs-bot", "quick", 5, t0.Add(2*time.Minute))...)
	out := f.q.Detect(t0.Add(10 * time.Minute))
	is = onlyIssue(t, f.q, "docs-bot")
	if is.State != decide.IssueOpen || is.ResolvingBy != "" {
		t.Fatalf("after new errors: %+v", is)
	}
	if p := pending(f.q); len(p) != 1 || p[0].ID != first[0].ID || len(out) != 1 || out[0].ID != first[0].ID {
		t.Fatalf("pending: %+v", p)
	}
}

func TestAgentViews(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.history.add(notFounds("chat-helper", "slow", 5, errorsAgo)...)
	f.q.Detect(t0)
	docs := onlyIssue(t, f.q, "docs-bot")
	chat := onlyIssue(t, f.q, "chat-helper")

	if got := f.q.IssuesOf("docs-bot"); len(got) != 1 || got[0].ID != docs.ID || got[0].ApplicationFix != docs.ApplicationFix || got[0].AiportFix.Kind != decide.KindApplicationAlias {
		t.Fatalf("IssuesOf: %+v", got)
	}
	if got := f.q.IssuesOf(""); len(got) != 0 {
		t.Fatalf("IssuesOf(\"\"): %+v", got)
	}
	if _, err := f.q.IssueOf("docs-bot", chat.ID); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("another Application's Issue: %v", err)
	}
	if _, err := f.q.IssueOf("", docs.ID); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("no Application: %v", err)
	}
	if a, err := f.q.IssueOf("docs-bot", docs.ID); err != nil || a.ID != docs.ID || a.Summary != docs.Summary {
		t.Fatalf("IssueOf: %+v %v", a, err)
	}
	if n := f.q.OpenIssues("docs-bot"); n != 1 {
		t.Fatalf("OpenIssues: %d", n)
	}
	f.q.DecideIssue(docs.ID, decide.Decision{Action: decide.Dismiss}, t0)
	if n := f.q.OpenIssues("docs-bot"); n != 0 {
		t.Fatalf("OpenIssues after dismiss: %d", n)
	}
	// States filter.
	if got := f.q.Issues(decide.IssueFilter{States: []decide.IssueState{decide.IssueDismissed}}); len(got) != 1 || got[0].ID != docs.ID {
		t.Fatalf("dismissed: %+v", got)
	}
	// Newest update first.
	if got := f.q.Issues(decide.IssueFilter{}); len(got) != 2 {
		t.Fatalf("all: %+v", got)
	}
}

func TestHintIssueNamesTheIssueACallIsOneOf(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.history.add(calls("docs-bot", "hosted", "any", "gateway_error", 502, 5, errorsAgo)...)
	f.history.add(levelRefusals("docs-bot", "fast", "/v1/messages", 5, errorsAgo)...)
	f.q.Detect(t0)
	ids := map[decide.Cause]string{}
	for _, is := range f.q.Issues(decide.IssueFilter{}) {
		ids[is.Cause] = is.ID
	}
	if len(ids) != 3 {
		t.Fatalf("issues: %v", ids)
	}
	one := func(e []telemetry.Event) telemetry.Event { return e[0] }
	cases := []struct {
		name string
		app  string
		e    telemetry.Event
		want string
	}{
		{"an unknown model", "docs-bot", one(notFounds("docs-bot", "quick", 1, t0)), ids[decide.CauseUnknownModel]},
		{"another model", "docs-bot", one(notFounds("docs-bot", "other", 1, t0)), ""},
		{"another Application", "chat-helper", one(notFounds("chat-helper", "quick", 1, t0)), ""},
		{"no Application", "", one(notFounds("docs-bot", "quick", 1, t0)), ""},
		{"a success", "docs-bot", one(successes("docs-bot", "local", "quick", 1, t0)), ""},
		{"a failing Connection", "docs-bot", one(calls("docs-bot", "hosted", "x", "timeout", 0, 1, t0)), ids[decide.CauseFailingConnection]},
		{"a request error on that Connection", "docs-bot", one(calls("docs-bot", "hosted", "x", "upstream_error", 400, 1, t0)), ""},
		{"an unmapped Level on its protocol", "docs-bot", one(levelRefusals("docs-bot", "fast", "/v1/messages", 1, t0)), ids[decide.CauseLevelUnmapped]},
		{"an unmapped Level on another protocol", "docs-bot", one(levelRefusals("docs-bot", "fast", "/v1/chat/completions", 1, t0)), ""},
	}
	for _, tc := range cases {
		id, ok := f.q.HintIssue(tc.app, tc.e)
		if id != tc.want || ok != (tc.want != "") {
			t.Errorf("%s: got %q %v, want %q", tc.name, id, ok, tc.want)
		}
	}
}

func TestFailingConnection(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(calls("docs-bot", "local", "m", "gateway_error", 502, 3, errorsAgo)...)
	f.history.add(calls("docs-bot", "local", "m", "timeout", 0, 2, errorsAgo.Add(3*time.Minute))...)
	f.history.add(calls("docs-bot", "local", "m", "upstream_error", 404, 9, errorsAgo)...) // about the request, not the Connection
	f.history.add(successes("chat-helper", "hosted", "m", 3, errorsAgo)...)
	f.history.add(successes("chat-helper", "side", "m", 1, errorsAgo)...)
	f.history.add(calls("docs-bot", "gone", "m", "gateway_error", 502, 9, errorsAgo)...) // not a configured Connection
	f.q.Detect(t0)
	var is decide.Issue
	for _, x := range f.q.Issues(decide.IssueFilter{Application: "docs-bot"}) {
		if x.Cause == decide.CauseFailingConnection {
			is = x
		}
	}
	if is.ID == "" || is.Subject != "local" || is.Evidence.Count != 5 {
		t.Fatalf("issue: %+v", is)
	}
	if is.Fix == nil || is.Fix.Kind != decide.KindBindApplication || is.Fix.Proposal != (decide.Proposal{Target: "hosted"}) {
		t.Fatalf("fix: %+v", is.Fix)
	}
	if !strings.Contains(is.Summary, `Connection "hosted" works: bind docs-bot to it.`) || !strings.Contains(is.ApplicationFix, `To move docs-bot to Connection "hosted"`) {
		t.Fatalf("summary %q\n%s", is.Summary, is.ApplicationFix)
	}
	// The failing Connection is refused as the fix.
	if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept, Proposal: decide.Proposal{Target: "local"}}, t0); !isInvalid(err) {
		t.Fatalf("binding to the failing Connection: %v", err)
	}
	got, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if o, _ := f.live.Config().Application("docs-bot"); o.Target != "hosted" || got.State != decide.IssueResolving {
		t.Fatalf("bound to %q, state %s", o.Target, got.State)
	}
	if !strings.Contains(got.ApplicationFix, `Already fixed in aiport: docs-bot is now bound to Connection "hosted"`) {
		t.Fatalf("application fix:\n%s", got.ApplicationFix)
	}
	// A success on the Connection it is bound to now confirms the fix.
	f.history.add(successes("docs-bot", "hosted", "m", 1, t0.Add(time.Minute))...)
	f.q.Detect(t0.Add(2 * time.Minute))
	if is, _ := f.q.Issue(is.ID); is.State != decide.IssueResolved || is.Resolution.How != decide.ByAiportFix {
		t.Fatalf("after a success: %+v", is)
	}
}

func TestFailingConnectionWithNothingWorking(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(calls("docs-bot", "local", "m", "gateway_error", 503, 5, errorsAgo)...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	if is.Fix.Proposal.Target != "" || !strings.Contains(is.Summary, "Choose a working Connection to bind docs-bot to.") {
		t.Fatalf("issue: %+v", is)
	}
	if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept}, t0); !isInvalid(err) {
		t.Fatalf("accepting without a Connection: %v", err)
	}
	if _, _, err := f.q.DecideIssue(is.ID, decide.Decision{Action: decide.Accept, Proposal: decide.Proposal{Target: "side"}}, t0); err != nil {
		t.Fatalf("the Owner chooses: %v", err)
	}
}

func TestLevelUnmapped(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(levelRefusals("docs-bot", "fast", "/v1/chat/completions", 5, errorsAgo)...)
	f.history.add(levelRefusals("docs-bot", "fast", "/v1/messages", 5, errorsAgo)...)
	f.history.add(levelRefusals("docs-bot", "quick", "/v1/chat/completions", 9, errorsAgo)...) // not a Level
	out := f.q.Detect(t0)
	issues := f.q.Issues(decide.IssueFilter{Application: "docs-bot"})
	if len(issues) != 2 || len(out) != 2 {
		t.Fatalf("want one Issue and fix per protocol: %+v", issues)
	}
	by := map[string]decide.Issue{}
	for _, is := range issues {
		if is.Cause != decide.CauseLevelUnmapped || is.Subject != "fast" {
			t.Fatalf("issue: %+v", is)
		}
		by[is.Protocol] = is
	}
	// The OpenAI default is proposed; there is no Anthropic default.
	if by["openai"].Fix.Proposal != (decide.Proposal{Alias: "fast", Target: "local"}) || by["anthropic"].Fix.Proposal != (decide.Proposal{Alias: "fast"}) {
		t.Fatalf("fixes: %+v / %+v", by["openai"].Fix, by["anthropic"].Fix)
	}
	if !strings.Contains(by["anthropic"].Summary, "Anthropic Connection and model") {
		t.Fatalf("summary: %q", by["anthropic"].Summary)
	}
	if !strings.Contains(by["openai"].ApplicationFix, "refused with 400 level_unmapped") {
		t.Fatalf("application fix:\n%s", by["openai"].ApplicationFix)
	}
	// A Connection of the other protocol would not reach the refused calls.
	_, _, err := f.q.DecideIssue(by["openai"].ID, decide.Decision{Action: decide.Accept, Proposal: decide.Proposal{Target: "side", Model: "model-z"}}, t0)
	if !isInvalid(err) || !strings.Contains(err.Error(), "OpenAI") {
		t.Fatalf("another protocol: %v", err)
	}
	// aiport never invents the model: the Owner must choose one.
	if _, _, err := f.q.DecideIssue(by["openai"].ID, decide.Decision{Action: decide.Accept}, t0); !isInvalid(err) {
		t.Fatalf("no model: %v", err)
	}
	got, _, err := f.q.DecideIssue(by["openai"].ID, decide.Decision{Action: decide.Accept, Proposal: decide.Proposal{Model: "model-z"}}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if m, ok := f.live.Config().Level("docs-bot", "openai", "fast"); !ok || m.Connection != "local" || m.Model != "model-z" {
		t.Fatalf("level: %+v %v", m, ok)
	}
	if !strings.Contains(got.ApplicationFix, `Already fixed in aiport: the Level "fast" now goes to Connection "local"`) {
		t.Fatalf("application fix:\n%s", got.ApplicationFix)
	}
	// The other protocol's Issue is untouched.
	if is, _ := f.q.Issue(by["anthropic"].ID); is.State != decide.IssueOpen {
		t.Fatalf("anthropic Issue: %s", is.State)
	}
}

func TestExplain(t *testing.T) {
	long := strings.Repeat("é", decide.MaxExplanation) // 2 bytes each
	cases := []struct {
		name string
		e    decide.Explanation
		err  func(error) bool
		text string
	}{
		{"empty", decide.Explanation{Text: ""}, isInvalid, ""},
		{"whitespace", decide.Explanation{Text: " \n\t"}, isInvalid, ""},
		{"trimmed", decide.Explanation{Text: "  the model was renamed \n", Source: decide.ByConnection, Connection: "hosted", Model: "m"}, nil, "the model was renamed"},
		{"cut at the limit, on a character boundary", decide.Explanation{Text: long, Source: decide.ByCodingAgent, Agent: "helper-agent"}, nil, strings.Repeat("é", decide.MaxExplanation/2)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
			f.q.Detect(t0)
			is := onlyIssue(t, f.q, "docs-bot")
			got, err := f.q.Explain(is.ID, tc.e, t0.Add(time.Minute))
			if tc.err != nil {
				if !tc.err(err) {
					t.Fatalf("got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Explanation == nil || got.Explanation.Text != tc.text || got.Explanation.At != t0.Add(time.Minute).UnixMilli() || got.UpdatedAt != t0.Add(time.Minute).UnixMilli() {
				t.Fatalf("explanation: %+v", got.Explanation)
			}
			// The Application fix carries it right after its first paragraph.
			parts := strings.SplitN(got.ApplicationFix, "\n\n", 3)
			if len(parts) < 3 || !strings.HasPrefix(parts[1], "Root cause, as explained by ") || !strings.HasSuffix(parts[1], tc.text) {
				t.Fatalf("application fix:\n%s", got.ApplicationFix)
			}
		})
	}
	f := newFixture(t, "", false)
	if _, err := f.q.Explain("missing", decide.Explanation{Text: "x"}, t0); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("an unknown Issue: %v", err)
	}
}

func TestExplanationAuthors(t *testing.T) {
	cases := []struct {
		e    decide.Explanation
		want string
	}{
		{decide.Explanation{Text: "x", Source: decide.ByConnection, Connection: "hosted", Model: "m"}, `Root cause, as explained by the Connection "hosted" (model "m"), from the evidence below only:` + "\nx"},
		{decide.Explanation{Text: "x", Source: decide.ByCodingAgent, Agent: "helper-agent"}, `Root cause, as explained by the Coding agent "helper-agent", after reading the Application's code:` + "\nx"},
		{decide.Explanation{Text: "x"}, "Root cause, as explained by aiport:\nx"},
	}
	for _, tc := range cases {
		is := decide.Issue{ID: "i", Application: "docs-bot", Cause: decide.CauseUnknownModel, Subject: "m", Explanation: &tc.e}
		if parts := strings.SplitN(decide.ApplicationFix(is), "\n\n", 3); len(parts) < 3 || parts[1] != tc.want {
			t.Errorf("source %q:\n%s", tc.e.Source, decide.ApplicationFix(is))
		}
	}
	if got := decide.ApplicationFix(decide.Issue{Cause: "no_such_cause"}); got != "" {
		t.Fatalf("an unknown cause: %q", got)
	}
}

func TestRecordBranch(t *testing.T) {
	f := newFixture(t, "", false)
	f.history.add(notFounds("docs-bot", "quick", 5, errorsAgo)...)
	f.q.Detect(t0)
	is := onlyIssue(t, f.q, "docs-bot")
	if _, err := f.q.RecordBranch("missing", decide.BranchFix{}, t0); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("an unknown Issue: %v", err)
	}
	files := make([]string, decide.MaxBranchFiles+5)
	for i := range files {
		files[i] = "file.go"
	}
	ready := decide.BranchFix{Branch: "aiport/fix-" + is.ID, Agent: "helper-agent", State: decide.BranchReady, Files: files, Summary: "1 file changed"}
	got, err := f.q.RecordBranch(is.ID, ready, t0.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if got.State != decide.IssueOpen || len(got.BranchFix.Files) != decide.MaxBranchFiles || got.BranchFix.At != t0.Add(time.Minute).UnixMilli() {
		t.Fatalf("ready: %+v", got.BranchFix)
	}
	if strings.Contains(got.ApplicationFix, "A fix is waiting on branch") {
		t.Fatal("a branch not yet kept is in the Application fix")
	}
	f.q.Explain(is.ID, decide.Explanation{Text: "renamed"}, t0)
	kept := ready
	kept.State, kept.Commit = decide.BranchKept, "0123456789abcdef0123"
	got, err = f.q.RecordBranch(is.ID, kept, t0.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if got.State != decide.IssueResolving || got.ResolvingBy != decide.ByApplicationFix || got.SinceMS != t0.Add(2*time.Minute).UnixMilli() {
		t.Fatalf("kept: %+v", got)
	}
	// The kept branch comes first after the opening paragraph, then the
	// explanation.
	parts := strings.SplitN(got.ApplicationFix, "\n\n", 4)
	wantBranch := `A fix is waiting on branch "aiport/fix-` + is.ID + `" of the Application's repository (commit 0123456789ab), made by the Coding agent "helper-agent" and kept by the Owner: 1 file changed. aiport does not merge it: merge it, deploy, and aiport confirms the fix from the next calls.`
	if len(parts) < 4 || parts[1] != wantBranch || !strings.HasPrefix(parts[2], "Root cause") {
		t.Fatalf("application fix:\n%s", got.ApplicationFix)
	}
	// A kept branch on an Issue that is not open leaves its state alone.
	again, _ := f.q.RecordBranch(is.ID, kept, t0.Add(3*time.Minute))
	if again.SinceMS != t0.Add(2*time.Minute).UnixMilli() {
		t.Fatalf("since moved: %+v", again)
	}
}

func TestBranchReviewable(t *testing.T) {
	cases := map[decide.BranchState]bool{
		decide.BranchRunning: false, decide.BranchReady: true, decide.BranchEmpty: false, decide.BranchKept: false,
		decide.BranchDiscarded: false, decide.BranchStopped: false, decide.BranchStoppedWaiting: true, decide.BranchTimedOut: true,
	}
	for s, want := range cases {
		if got := (decide.BranchFix{State: s}).Reviewable(); got != want {
			t.Errorf("%s: %v", s, got)
		}
	}
}

func TestPruneIssues(t *testing.T) {
	f := newFixture(t, "telemetry: {retention_days: 1}\n", true)
	f.history.add(notFounds("docs-bot", "resolves", 5, errorsAgo)...)
	f.history.add(notFounds("docs-bot", "refused", 5, errorsAgo)...)
	f.history.add(notFounds("docs-bot", "sticks", 5, errorsAgo)...)
	f.q.Detect(t0)
	ids := map[string]string{}
	for _, is := range f.q.Issues(decide.IssueFilter{}) {
		ids[is.Subject] = is.ID
	}
	// One rejected fix (kept forever), one left open, one resolving itself.
	f.q.DecideIssue(ids["refused"], decide.Decision{Action: decide.Reject}, t0)
	f.history.add(successes("docs-bot", "local", "resolves", 1, t0)...)
	f.history.add(successes("docs-bot", "local", "refused", 1, t0)...)
	f.history.add(notFounds("docs-bot", "sticks", 1, t0)...)
	f.q.Detect(t0.Add(time.Minute))
	f.q.Prune(t0.Add(48 * time.Hour))
	for subject, kept := range map[string]bool{"resolves": false, "refused": true} {
		if _, ok := f.q.Issue(ids[subject]); ok != kept {
			t.Errorf("%s kept=%v, want %v", subject, ok, kept)
		}
	}
	if _, ok := f.q.Issue(ids["sticks"]); !ok {
		t.Errorf("the open Issue was forgotten")
	}
	if _, stored := f.store.issues[ids["resolves"]]; stored {
		t.Fatal("the pruned Issue is still stored")
	}
}
