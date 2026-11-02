package server

import (
	"errors"
	"net/http"
	"strings"

	"aiport/internal/decide"
	"aiport/internal/refusal"
	"aiport/internal/telemetry"
)

// IssuesPath is where an Application's Coding agent reads its own Issues
// and says it fixed one on its side:
//
//	GET  /v1/aiport/issues             this Application's Issues (open, being resolved, recently resolved or dismissed), newest first
//	GET  /v1/aiport/issues/{id}        one of them
//	POST /v1/aiport/issues/{id}/fixed  "fixed on my side": the Issue becomes resolving; aiport confirms it from real calls
//
// The MCP tools aiport_my_issues and aiport_issue_fixed answer through this
// handler, so both doors agree.
const IssuesPath = refusal.IssuesPath

// serveIssues handles IssuesPath for an authenticated caller; it only ever
// answers about the Application whose token was sent.
func (h *Handler) serveIssues(w http.ResponseWriter, r *http.Request, applicationID string, fail func(int, string, string)) {
	rest, one := strings.CutPrefix(r.URL.Path, IssuesPath+"/")
	id, fixed := strings.CutSuffix(rest, "/fixed")
	notFound := func() { fail(404, "issue_not_found", "No Issue with this id for this Application") }
	if one && (id == "" || strings.Contains(id, "/")) {
		notFound()
		return
	}
	if !h.isApplication(applicationID) {
		fail(403, "application_token_required", "Issues belong to one Application: send that Application's own token")
		return
	}
	if h.decisions == nil || !h.decisions.On() {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return
	}
	allow := "GET"
	switch {
	case fixed && r.Method == http.MethodPost:
		is, err := h.decisions.FixedByApplication(applicationID, id, h.clock())
		switch {
		case errors.Is(err, decide.ErrNotFound):
			notFound()
		case errors.Is(err, decide.ErrNotOpen):
			fail(409, "issue_not_open", err.Error())
		case errors.Is(err, decide.ErrOff):
			fail(409, "adapting_off", err.Error())
		case err != nil:
			fail(500, "issue_update_failed", "The Issue could not be updated")
		default:
			writeJSON(w, 200, is)
		}
		return
	case fixed:
		allow = "POST"
	case one && r.Method == http.MethodGet:
		is, err := h.decisions.IssueOf(applicationID, id)
		if err != nil {
			notFound()
			return
		}
		writeJSON(w, 200, is)
		return
	case !one && r.Method == http.MethodGet:
		writeJSON(w, 200, map[string]any{"issues": h.decisions.IssuesOf(applicationID)})
		return
	}
	w.Header().Set("Allow", allow)
	fail(405, "method_not_allowed", "Use "+allow)
}

// issueHint is the hint for a failing call from applicationID, e as
// request history will record it: it names the open Issue the call is one
// of the failures of, if any. Empty when there is none or aiport does not
// adapt. Delivering an Issue is not a Change request, so it needs no
// permission to request changes. It reads only snapshots: nothing on the
// forwarding path waits for the decision queue.
func (h *Handler) issueHint(applicationID string, e telemetry.Event) string {
	if h.decisions == nil || h.snapshot().Adapting.Disabled || !h.isApplication(applicationID) {
		return ""
	}
	id, ok := h.decisions.HintIssue(applicationID, e)
	if !ok {
		return ""
	}
	return refusal.IssueHint(id)
}
