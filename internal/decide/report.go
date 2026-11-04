package decide

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/harelix/aiport/internal/telemetry"
)

// A reported request is an Issue the Owner opens from one failed request in
// Activity ("Report to the Application's coding agent"). aiport detects
// some failures by itself (an unknown model, a failing Connection, an
// unmapped Level); many others only the Owner sees, such as a provider
// refusing a request body. Reporting one hands the failure, what aiport
// knows about it and the Owner's note to the Application's coding agent the
// way every Issue reaches it: aiport_my_issues and GET /v1/aiport/issues at
// the start of its session, X-Aiport-Hint on its next failing call of the
// same kind, and the Owner's Fix it in Changes, where aiport runs a coding
// agent on the Application's repository.
//
// Failures of the same kind (the same Application, HTTP status, error,
// name sent and module) share one Issue: reporting another adds it to the
// evidence. The Issue resolves when the Application says it fixed it and
// its next call of that name and module succeeds, or when the Owner
// dismisses it.

// CauseReportedRequest: the Owner reported a failed request to its
// Application's coding agent. Subject is the failure's kind (Reported.kind).
const CauseReportedRequest Cause = "reported_request"

// MaxReportNote is the longest note the Owner can add, in characters.
const MaxReportNote = 2000

// Reported is what a reported-request Issue keeps about the failure: ids,
// names, codes and aiport's own explanation, plus the Owner's note (which
// may carry the error text the Application showed: request history never
// keeps provider messages).
type Reported struct {
	RequestID  string `json:"request_id"`
	At         int64  `json:"at_ms"`
	Route      string `json:"route"`
	Name       string `json:"name,omitempty"`        // the model name the Application sent
	Model      string `json:"model,omitempty"`       // the model aiport chose
	Connection string `json:"connection,omitempty"`  // the Connection it went to
	Module     string `json:"module,omitempty"`      // X-Aiport-Module
	Effort     string `json:"effort,omitempty"`      // the effort of the call
	Status     int    `json:"http_status"`           // the answer's HTTP status
	Outcome    string `json:"outcome"`               // request history's outcome
	Error      string `json:"error,omitempty"`       // refusal code, provider error type or connection error
	Reason     string `json:"reason,omitempty"`      // aiport's summary of the failure
	Explain    string `json:"explanation,omitempty"` // aiport's explanation
	OwnerFix   string `json:"owner_fix,omitempty"`   // what aiport tells the Owner to do, if anything
	Note       string `json:"note,omitempty"`        // the Owner's note for the coding agent
	By         string `json:"by,omitempty"`          // who reported it
}

// kind is the Subject of a reported request: failures of the same kind
// share an Issue.
func (r Reported) kind() string {
	return strconv.Itoa(r.Status) + " " + orDash(r.Error) + " " + orDash(r.Name) + " " + orDash(r.Module)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// sameKind reports whether call e failed the way the reported request did.
func (r Reported) sameKind(e telemetry.Event) bool { return r.like(e) && failureCode(e) == r.Error }

// mayBeKind is sameKind for a call still being answered (X-Aiport-Hint):
// the provider's error type is read from the body after the headers are
// written, so a call that has none yet may still be of the kind.
func (r Reported) mayBeKind(e telemetry.Event) bool {
	c := failureCode(e)
	return r.like(e) && (c == r.Error || c == "")
}

// like reports whether call e failed with the reported status, name and
// module.
func (r Reported) like(e telemetry.Event) bool {
	if !(telemetry.IsError(e.Outcome) || e.Status >= 400) || e.Status != r.Status || e.Module != r.Module {
		return false
	}
	name := ""
	if e.RequestedModel != nil {
		name = *e.RequestedModel
	}
	return name == r.Name
}

// failureCode is the most specific error identifier history keeps for e.
func failureCode(e telemetry.Event) string {
	switch {
	case e.RefusalCode != "":
		return e.RefusalCode
	case e.ConnectionError != "":
		return e.ConnectionError
	}
	return e.UpstreamErrorType
}

// ReportedFrom is the Reported for request e, before aiport's explanation
// and the Owner's note are added.
func ReportedFrom(e telemetry.Event) Reported {
	r := Reported{RequestID: e.ID, At: e.StartedAt, Route: e.Route, Model: e.ResolvedModel, Connection: e.Target, Module: e.Module, Effort: e.Effort, Status: e.Status, Outcome: e.Outcome, Error: failureCode(e)}
	if e.RequestedModel != nil {
		r.Name = *e.RequestedModel
	}
	return r
}

// ErrNotFailed: only a failed request can be reported.
var ErrNotFailed = Invalid{"only a failed request can be reported to the Application's coding agent"}

// ErrNoApplication: the request came from no known Application.
var ErrNoApplication = Invalid{"this request came from no Application, so there is no coding agent to report it to"}

// Report opens (or adds to) the reported-request Issue of Application app
// for failure r. A second report of a request already in the Issue's
// evidence changes only the note.
func (q *Queue) Report(app string, r Reported, at time.Time) (Issue, error) {
	if !q.On() {
		return Issue{}, ErrOff
	}
	if app == "" {
		return Issue{}, ErrNoApplication
	}
	if r.Status < 400 && !telemetry.IsError(r.Outcome) {
		return Issue{}, ErrNotFailed
	}
	r.Note = strings.TrimSpace(r.Note)
	if utf8.RuneCountInString(r.Note) > MaxReportNote {
		return Issue{}, Invalid{"the note must be at most 2,000 characters"}
	}
	q.drainCatalog()
	q.mu.Lock()
	defer q.mu.Unlock()
	now := at.UnixMilli()
	subject := r.kind()
	if i := q.issueFor(app, CauseReportedRequest, subject, ""); i >= 0 && undecided(q.issues[i].State) {
		is := &q.issues[i]
		if !containsString(is.Evidence.Examples, r.RequestID) {
			is.Evidence.Count++
			is.Evidence.Examples = append([]string{r.RequestID}, is.Evidence.Examples...)
			if len(is.Evidence.Examples) > 5 {
				is.Evidence.Examples = is.Evidence.Examples[:5]
			}
			is.Evidence.FromMS = min(is.Evidence.FromMS, r.At)
			is.Evidence.ToMS = max(is.Evidence.ToMS, r.At)
		}
		if r.Note == "" && is.Report != nil {
			r.Note = is.Report.Note // a report without a note keeps the earlier one
		}
		is.Report = &r
		if is.State == IssueResolving { // it failed again: back to the agent
			is.State, is.ResolvingBy, is.SinceMS = IssueOpen, "", now
		}
		is.UpdatedAt = now
		is.Summary = reportedSummary(*is)
		q.saveIssue(is)
		return *is, nil
	}
	ev := Evidence{Count: 1, FromMS: r.At, ToMS: r.At, Status: r.Status, Error: r.Error, Examples: []string{r.RequestID}}
	if r.Connection != "" {
		ev.Connections = []string{r.Connection}
	}
	q.issues = append(q.issues, Issue{
		ID: newID(), Application: app, Cause: CauseReportedRequest, Subject: subject, State: IssueOpen,
		Evidence: ev, Report: &r, SinceMS: now, CreatedAt: now, UpdatedAt: now,
	})
	is := &q.issues[len(q.issues)-1]
	is.Summary = reportedSummary(*is)
	q.saveIssue(is)
	return *is, nil
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// reportResolutions resolves reported-request Issues the Application said
// it fixed once a call of the same name and module succeeds after its last
// failure of that kind. An open one stays open: one success elsewhere in a
// module does not show the reported failure is gone. Called with the lock
// held.
func (q *Queue) reportResolutions(at time.Time) {
	if q.history == nil {
		return
	}
	now := at.UnixMilli()
	for i := range q.issues {
		is := &q.issues[i]
		if is.Cause != CauseReportedRequest || is.State != IssueResolving || is.Report == nil {
			continue
		}
		r := *is.Report
		after, success := is.SinceMS, int64(0)
		for _, e := range q.history.Find(telemetry.Query{Origin: is.Application, From: is.SinceMS, To: now, InferenceOnly: true}) {
			name := ""
			if e.RequestedModel != nil {
				name = *e.RequestedModel
			}
			switch {
			case r.sameKind(e):
				after = max(after, e.StartedAt)
			case telemetry.Succeeded(e.Outcome) && name == r.Name && e.Module == r.Module:
				success = max(success, e.StartedAt)
			}
		}
		if success > after {
			q.resolve(is, ByApplicationFix, now)
		}
	}
}

// callOf names the reported call: "calls for balanced (module visualize)".
func callOf(r Reported) string {
	s := "calls"
	if r.Name != "" {
		s += " for " + strconv.Quote(r.Name)
	}
	if r.Module != "" {
		s += " (module " + strconv.Quote(r.Module) + ")"
	}
	return s
}

func failureOf(r Reported) string {
	s := "HTTP " + strconv.Itoa(r.Status)
	if r.Error != "" {
		s += " (" + r.Error + ")"
	}
	return s
}

func reportedSummary(is Issue) string {
	r := is.Report
	if r == nil {
		return ""
	}
	n := "1 failed request"
	if is.Evidence.Count > 1 {
		n = strconv.Itoa(is.Evidence.Count) + " failed requests"
	}
	s := "The Owner reported " + n + " to " + is.Application + "'s coding agent: " + callOf(*r) + " failed with " + failureOf(*r)
	if r.Reason != "" {
		s += ". " + strings.TrimSuffix(r.Reason, ".")
	}
	return s + "."
}

// reportedFix is the Application fix of a reported request: what failed,
// what aiport knows about it, the Owner's note and what to do.
func reportedFix(is Issue) string {
	r := is.Report
	if r == nil {
		return ""
	}
	app := is.Application
	var b strings.Builder
	b.WriteString("aiport Issue " + is.ID + ": the Owner reported a failed request from the Application " + app + " to its coding agent: " + callOf(*r) + " failed with " + failureOf(*r) + ".\n\n")
	b.WriteString("The request: " + r.RequestID + " at " + utc(r.At) + ", route " + r.Route)
	if r.Effort != "" {
		b.WriteString(", effort " + r.Effort)
	}
	if r.Model != "" {
		b.WriteString(", served by model " + strconv.Quote(r.Model))
	}
	if r.Connection != "" {
		b.WriteString(" on Connection " + strconv.Quote(r.Connection))
	}
	b.WriteString(".")
	if ev := is.Evidence; ev.Count > 1 {
		b.WriteString(" Reported " + strconv.Itoa(ev.Count) + " times between " + utc(ev.FromMS) + " and " + utc(ev.ToMS) + "; example request ids: " + strings.Join(ev.Examples, ", ") + ".")
	}
	b.WriteString("\n\n")
	if r.Reason != "" || r.Explain != "" {
		b.WriteString("What aiport knows: " + strings.TrimSpace(strings.TrimSuffix(r.Reason, ".")+". "+r.Explain) + "\n")
		b.WriteString("aiport keeps request metadata only, never the provider's error message or the request body.\n\n")
	}
	if r.Note != "" {
		b.WriteString("The Owner's note:\n" + r.Note + "\n\n")
	}
	b.WriteString("What to do in " + app + ":\n")
	b.WriteString("1. Find the code that sends " + callOf(*r) + " and reproduce the failure")
	if r.Note == "" {
		b.WriteString(" (the Application's own error message says what the provider rejected)")
	}
	b.WriteString(".\n")
	b.WriteString("2. Fix it in " + app + "; if the fix needs something from aiport (a task name, a route, a Level), ask with a Change request instead of working around it.\n")
	b.WriteString("3. Say you fixed it: aiport_issue_fixed (or POST /v1/aiport/issues/" + is.ID + "/fixed).\n\n")
	b.WriteString("How to confirm: make the same call through aiport with " + app + "'s own token and check it no longer answers " + strconv.Itoa(r.Status) + ". aiport marks this Issue resolved after you say it is fixed and the next " + callOf(*r) + " succeed; another failure of the same kind reopens it.")
	return b.String()
}
