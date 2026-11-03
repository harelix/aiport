package decide_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/decide"
)

// request builds a Change request body from a valid one, with fields
// replaced (a nil value removes the field).
func request(t *testing.T, change map[string]any) []byte {
	t.Helper()
	r := map[string]any{
		"kind": "application_alias", "need": "a name for code review", "details": "send review-code to a capable model",
		"why": "reviews time out", "scope": "application", "scope_case": "only docs-bot reviews code",
	}
	for k, v := range change {
		if v == nil {
			delete(r, k)
			continue
		}
		r[k] = v
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestFileListsEveryProblemInFieldOrder(t *testing.T) {
	type pr = decide.Problem
	cases := []struct {
		name string
		body []byte
		want []pr
	}{
		{"not JSON", []byte("not json"), []pr{{Field: "", Problem: "send a JSON object with kind, need, details, why, scope and scope_case (filed_by, alias, prefer_connection and prefer_model are optional)"}}},
		{"null", []byte("null"), []pr{{Field: "", Problem: "send a JSON object with kind, need, details, why, scope and scope_case (filed_by, alias, prefer_connection and prefer_model are optional)"}}},
		{"an array", []byte("[]"), []pr{{Field: "", Problem: "send a JSON object with kind, need, details, why, scope and scope_case (filed_by, alias, prefer_connection and prefer_model are optional)"}}},
		{"empty object: every required field missing", []byte("{}"), []pr{
			{"kind", "missing; the kind of Change, such as application_alias or bind_application (see /v1/aiport/capabilities)"},
			{"need", "missing; one line saying what is needed"},
			{"details", "missing; the details the Owner needs to make the Change"},
			{"why", "missing; why the Application needs it"},
			{"scope", "missing; application (only this Application) or every_application"},
			{"scope_case", "missing; the argument for that Scope"},
		}},
		{"whitespace only is missing", request(t, map[string]any{"why": "   \t"}), []pr{{"why", "missing; why the Application needs it"}}},
		{"a number is not a string (and not also missing)", request(t, map[string]any{"need": 7}), []pr{{"need", "must be a string"}}},
		{"unknown fields sort after known ones, by name", request(t, map[string]any{"zeta": "x", "alpha": "y", "kind": "Bad Kind"}), []pr{
			{"kind", "must be a short name of lowercase letters, digits and underscores, such as application_alias"},
			{"alpha", "unknown field; the fields are kind, need, details, why, scope, scope_case, filed_by, alias, prefer_connection, prefer_model"},
			{"zeta", "unknown field; the fields are kind, need, details, why, scope, scope_case, filed_by, alias, prefer_connection, prefer_model"},
		}},
		{"need must be one line", request(t, map[string]any{"need": "one\ntwo"}), []pr{{"need", "must be one line"}}},
		{"kind longer than 64", request(t, map[string]any{"kind": strings.Repeat("k", 65)}), []pr{{"kind", "must be a short name of lowercase letters, digits and underscores, such as application_alias"}}},
		{"bad scope", request(t, map[string]any{"scope": "everyone"}), []pr{{"scope", "must be application or every_application"}}},
		{"names have no spaces", request(t, map[string]any{"alias": "review code", "prefer_model": "a b"}), []pr{
			{"alias", "must be one name without spaces"}, {"prefer_model", "must be one name without spaces"},
		}},
		{"names at most 256", request(t, map[string]any{"prefer_connection": strings.Repeat("c", 257)}), []pr{{"prefer_connection", "too long: at most 256 characters"}}},
		{"a field at most 2000 characters", request(t, map[string]any{"details": strings.Repeat("d", 2001)}), []pr{{"details", "too long: at most 2,000 characters"}}},
		{"all fields at most 6000 characters", request(t, map[string]any{"details": strings.Repeat("d", 2000), "why": strings.Repeat("w", 2000), "scope_case": strings.Repeat("s", 2000)}), []pr{
			{"", "too long: all fields together at most 6,000 characters"},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "", false)
			_, err := f.q.File("docs-bot", c.body, t0)
			var inc decide.Incomplete
			if !errors.As(err, &inc) {
				t.Fatalf("want Incomplete, got %v", err)
			}
			if !reflect.DeepEqual(inc.Problems, c.want) {
				t.Fatalf("problems:\n got %+v\nwant %+v", inc.Problems, c.want)
			}
			if len(f.q.Requests("docs-bot")) != 0 {
				t.Fatal("a refused request was kept")
			}
		})
	}
}

// Lengths count characters, not bytes: 2000 multibyte characters fit.
func TestFileCountsCharactersNotBytes(t *testing.T) {
	f := newFixture(t, "", false)
	a, err := f.q.File("docs-bot", request(t, map[string]any{"details": strings.Repeat("é", 2000)}), t0)
	if err != nil {
		t.Fatalf("2000 two-byte characters: %v", err)
	}
	if a.Status != decide.Pending {
		t.Fatalf("status %q", a.Status)
	}
}

func TestFileRefusesInOrder(t *testing.T) {
	cases := []struct {
		name  string
		extra string
		app   string
		want  error
	}{
		{"loop off", "adapting: {disabled: true}\n", "docs-bot", decide.ErrOff},
		{"not an Application", "", "nobody", decide.ErrNotApplication},
		{"the management token", "", "", decide.ErrNotApplication},
		{"not permitted", "", "quiet-app", decide.ErrNotPermitted},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, c.extra, false)
			// An invalid body: the earlier refusal wins.
			if _, err := f.q.File(c.app, []byte("{}"), t0); !errors.Is(err, c.want) {
				t.Fatalf("got %v, want %v", err, c.want)
			}
		})
	}
}

func TestFileRateLimitsEachApplication(t *testing.T) {
	f := newFixture(t, "", false)
	for i := 0; i < decide.RequestsPerHour; i++ {
		// Refused bodies never spend the allowance.
		if _, err := f.q.File("docs-bot", []byte("{}"), t0); err == nil {
			t.Fatal("an empty request was filed")
		}
		at := t0.Add(time.Duration(i) * time.Minute)
		if _, err := f.q.File("docs-bot", request(t, map[string]any{"need": "need " + string(rune('a'+i))}), at); err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
	}
	// The 21st, even with a bad body, is rate limited; RetryAfter is when
	// the oldest one leaves the hour.
	at := t0.Add(30 * time.Minute)
	_, err := f.q.File("docs-bot", []byte("{}"), at)
	var rl decide.RateLimited
	if !errors.As(err, &rl) {
		t.Fatalf("want RateLimited, got %v", err)
	}
	if rl.RetryAfter != 30*time.Minute {
		t.Fatalf("RetryAfter %v, want 30m", rl.RetryAfter)
	}
	// Another Application has its own allowance.
	if _, err := f.q.File("chat-helper", request(t, nil), at); err != nil {
		t.Fatalf("chat-helper: %v", err)
	}
	// One hour after the first, one more fits.
	if _, err := f.q.File("docs-bot", request(t, map[string]any{"need": "later"}), t0.Add(time.Hour)); err != nil {
		t.Fatalf("an hour later: %v", err)
	}
	if _, err := f.q.File("docs-bot", request(t, map[string]any{"need": "later still"}), t0.Add(time.Hour)); !errors.As(err, &rl) {
		t.Fatalf("want RateLimited again, got %v", err)
	}
}

func TestFileKeepsTheRequestAsTheApplicationSentIt(t *testing.T) {
	f := newFixture(t, "", false)
	body := request(t, map[string]any{"need": "  a name for code review  ", "filed_by": "agent-7", "alias": "review-code", "prefer_connection": "hosted", "prefer_model": "model-x"})
	a, err := f.q.File("docs-bot", body, t0)
	if err != nil {
		t.Fatal(err)
	}
	want := decide.Answer{
		ID: a.ID, Kind: decide.KindApplicationAlias, Need: "a name for code review", Details: "send review-code to a capable model",
		Why: "reviews time out", Scope: decide.ScopeApplication, ScopeCase: "only docs-bot reviews code", FiledBy: "agent-7",
		Alias: "review-code", PreferConnection: "hosted", PreferModel: "model-x", Status: decide.Pending,
		CreatedAt: t0.UnixMilli(), UpdatedAt: t0.UnixMilli(),
	}
	if a.ID == "" || !reflect.DeepEqual(a, want) {
		t.Fatalf("answer:\n got %+v\nwant %+v", a, want)
	}
	// The Owner sees it as a pending Change request with the preferences
	// as its proposal.
	items := f.q.List(decide.Filter{Source: decide.FromRequest})
	if len(items) != 1 || items[0].Proposal != (decide.Proposal{Alias: "review-code", Target: "hosted", Model: "model-x"}) || items[0].Application != "docs-bot" {
		t.Fatalf("items: %+v", items)
	}
	if got, err := f.q.RequestOf("docs-bot", a.ID); err != nil || got != a {
		t.Fatalf("RequestOf: %+v %v", got, err)
	}
	if _, err := f.q.RequestOf("chat-helper", a.ID); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("another Application's request: %v", err)
	}
	if got := f.q.Requests("chat-helper"); len(got) != 0 {
		t.Fatalf("chat-helper sees %+v", got)
	}
}

func TestFileMergesAnIdenticalWaitingRequest(t *testing.T) {
	f := newFixture(t, "", false)
	first, err := f.q.File("docs-bot", request(t, nil), t0)
	if err != nil {
		t.Fatal(err)
	}
	again, err := f.q.File("docs-bot", request(t, nil), t0.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != first.ID || again.Repeats != 1 || again.UpdatedAt != t0.Add(time.Minute).UnixMilli() || again.CreatedAt != first.CreatedAt {
		t.Fatalf("merged: %+v", again)
	}
	// Any difference is another request; so is the same from another
	// Application.
	other, _ := f.q.File("docs-bot", request(t, map[string]any{"why": "a different reason"}), t0)
	fromChat, _ := f.q.File("chat-helper", request(t, nil), t0)
	if other.ID == first.ID || fromChat.ID == first.ID {
		t.Fatal("different requests were merged")
	}
	// Once decided, the same request is a new one.
	if _, _, err := f.q.Decide(first.ID, decide.Decision{Action: decide.Reject, Note: "no"}, t0.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	fresh, _ := f.q.File("docs-bot", request(t, nil), t0.Add(3*time.Minute))
	if fresh.ID == first.ID || fresh.Repeats != 0 || fresh.Status != decide.Pending {
		t.Fatalf("after a decision: %+v", fresh)
	}
	decided, _ := f.q.RequestOf("docs-bot", first.ID)
	if decided.Status != decide.Rejected || decided.Note != "no" || decided.DecidedAt != t0.Add(2*time.Minute).UnixMilli() {
		t.Fatalf("decided answer: %+v", decided)
	}
}

func TestFileForEveryApplicationJoinsAMatchingRequest(t *testing.T) {
	f := newFixture(t, "", false)
	every := func(details, why string) []byte {
		return request(t, map[string]any{"scope": "every_application", "details": details, "why": why, "scope_case": "everyone reviews"})
	}
	a, err := f.q.File("docs-bot", every("Send review-code to a capable model", "docs-bot reason"), t0)
	if err != nil {
		t.Fatal(err)
	}
	// Same kind and details, ignoring case and spacing, from another
	// Application: one item asked by both.
	b, err := f.q.File("chat-helper", every("  send   REVIEW-CODE to a capable model ", "chat-helper reason"), t0.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != a.ID {
		t.Fatalf("not joined: %s vs %s", a.ID, b.ID)
	}
	items := f.q.List(decide.Filter{})
	if len(items) != 1 || len(items[0].AskedBy) != 2 {
		t.Fatalf("items: %+v", items)
	}
	// docs-bot asks again: its own repeat count goes up.
	if _, err := f.q.File("docs-bot", every("send review-code to a capable model", "docs-bot again"), t0.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	// Each Application sees its own argument and count, never the other's.
	docs, _ := f.q.RequestOf("docs-bot", a.ID)
	chat, _ := f.q.RequestOf("chat-helper", a.ID)
	if docs.Why != "docs-bot again" || docs.Repeats != 1 || docs.CreatedAt != t0.UnixMilli() {
		t.Fatalf("docs-bot sees %+v", docs)
	}
	if chat.Why != "chat-helper reason" || chat.Repeats != 0 || chat.CreatedAt != t0.Add(time.Minute).UnixMilli() {
		t.Fatalf("chat-helper sees %+v", chat)
	}
	// Both find it in their own lists, and an Application filter on the
	// queue matches either.
	if len(f.q.Requests("docs-bot")) != 1 || len(f.q.Requests("chat-helper")) != 1 {
		t.Fatal("a joined request is missing from a list")
	}
	if len(f.q.List(decide.Filter{Application: "chat-helper"})) != 1 {
		t.Fatal("the Application filter misses a joined request")
	}
	// Different details are another request.
	c, _ := f.q.File("chat-helper", every("something else", "x"), t0.Add(3*time.Minute))
	if c.ID == a.ID {
		t.Fatal("different details joined")
	}
}

func TestUnsupportedKindsAreDemand(t *testing.T) {
	f := newFixture(t, "", false)
	file := func(app, kind, need string) decide.Answer {
		t.Helper()
		a, err := f.q.File(app, request(t, map[string]any{"kind": kind, "need": need}), t0)
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	a := file("docs-bot", "rate_limit", "fewer calls")
	if a.Status != decide.NotSupported {
		t.Fatalf("status %q", a.Status)
	}
	file("docs-bot", "rate_limit", "fewer calls") // a repeat
	file("chat-helper", "rate_limit", "slower")
	file("docs-bot", "cache_responses", "cache")
	file("chat-helper", "budget_cap", "cap")
	file("docs-bot", "application_alias", "supported, not demand")

	if got := f.q.List(decide.Filter{Status: decide.Pending}); len(got) != 1 {
		t.Fatalf("want one pending (the supported one), got %+v", got)
	}
	want := []decide.Demand{
		{Kind: "rate_limit", Requests: 3, Applications: 2, LastAt: t0.UnixMilli()},
		{Kind: "budget_cap", Requests: 1, Applications: 1, LastAt: t0.UnixMilli()},
		{Kind: "cache_responses", Requests: 1, Applications: 1, LastAt: t0.UnixMilli()},
	}
	if got := f.q.Demand(); !reflect.DeepEqual(got, want) {
		t.Fatalf("demand:\n got %+v\nwant %+v", got, want)
	}
	// A not-supported request needs no decision.
	if _, _, err := f.q.Decide(a.ID, decide.Decision{Action: decide.Reject}, t0); !errors.Is(err, decide.ErrDecided) {
		t.Fatalf("deciding a not-supported request: %v", err)
	}
}

func TestDemandCountsEachAskerOfAnUnsupportedRequestForEveryApplication(t *testing.T) {
	f := newFixture(t, "", false)
	body := request(t, map[string]any{"kind": "rate_limit", "scope": "every_application"})
	for _, app := range []string{"docs-bot", "chat-helper", "docs-bot"} {
		if _, err := f.q.File(app, body, t0); err != nil {
			t.Fatal(err)
		}
	}
	want := []decide.Demand{{Kind: "rate_limit", Requests: 3, Applications: 2, LastAt: t0.UnixMilli()}}
	if got := f.q.Demand(); !reflect.DeepEqual(got, want) {
		t.Fatalf("demand: %+v", got)
	}
}

func TestReadsWithTheLoopOffShowNothing(t *testing.T) {
	f := newFixture(t, "adapting: {disabled: true}\n", false)
	if f.q.On() {
		t.Fatal("the loop is on")
	}
	if got := f.q.Requests("docs-bot"); got == nil || len(got) != 0 {
		t.Fatalf("Requests: %#v", got)
	}
	if got := f.q.Demand(); got == nil || len(got) != 0 {
		t.Fatalf("Demand: %#v", got)
	}
	if got := f.q.List(decide.Filter{}); got == nil || len(got) != 0 {
		t.Fatalf("List: %#v", got)
	}
	if _, err := f.q.RequestOf("docs-bot", "any"); !errors.Is(err, decide.ErrNotFound) {
		t.Fatalf("RequestOf: %v", err)
	}
}

func TestErrorTexts(t *testing.T) {
	if (decide.RateLimited{RetryAfter: time.Minute}).Error() == "" || (decide.Incomplete{}).Error() == "" || (decide.Invalid{Reason: "why"}).Error() != "why" {
		t.Fatal("an error has no text")
	}
}
