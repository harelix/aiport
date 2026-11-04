package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/decide"
)

// modelFinder adapts the gateway's model listings to the decision module's
// question "which enabled Connections serve this model?", adding the models
// Connections of that protocol list (cached listings, nothing generated).
type modelFinder struct{ h *Handler }

func (f modelFinder) Serves(ctx context.Context, protocol, model string) decide.Diagnosis {
	s := f.h.gateway.Serves(ctx, protocol, model)
	return decide.Diagnosis{Protocol: s.Protocol, Answer: decide.Served(s.Answer), Serving: s.Serving, Unknown: s.Unknown, Available: f.h.gateway.Available(ctx, protocol)}
}

// issuesRoute serves Issues to the Owner:
//
//	GET  /api/v1/issues?state=&application=  Issues, newest first; state is open (default: open and resolving), resolving, resolved, dismissed or all
//	POST /api/v1/issues                      {request_id, note}: report a failed request to its Application's coding agent (an Issue of cause reported_request)
//	POST /api/v1/issues/{id}                 {action: accept|reject|dismiss, note, target, model}: decide the aiport fix
//
// The listing says whether Explain is available ("explain": available,
// and who writes, or a hint); POST /api/v1/issues/{id}/explain is explainRoute.
//
// The caller has already verified management access.
func (h *Handler) issuesRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path == "/api/v1/issues" {
		if r.Method == "POST" {
			h.reportRequest(w, r, fail)
			return true
		}
		if r.Method != "GET" {
			fail(405, "method_not_allowed", "Use GET or POST")
			return true
		}
		f := decide.IssueFilter{Application: r.URL.Query().Get("application")}
		switch st := r.URL.Query().Get("state"); st {
		case "":
			f.States = []decide.IssueState{decide.IssueOpen, decide.IssueResolving}
		case "all":
		case string(decide.IssueOpen), string(decide.IssueResolving), string(decide.IssueResolved), string(decide.IssueDismissed):
			f.States = []decide.IssueState{decide.IssueState(st)}
		default:
			fail(400, "invalid_state", "state must be open, resolving, resolved, dismissed or all")
			return true
		}
		issues := []decide.Issue{}
		if h.decisions != nil {
			issues = h.decisions.Issues(f)
		}
		all := h.codingAgents() // one look for installed agents per request
		writeJSON(w, 200, map[string]any{"adapting": h.decisions != nil && h.decisions.On(), "auto_fix": h.live.Config().Adapting.AutoFix, "explain": h.explainStatus(), "explain_by": h.explainBy(all), "fix_by": h.fixBy(all), "issues": issues})
		return true
	}
	id, ok := strings.CutPrefix(r.URL.Path, "/api/v1/issues/")
	if !ok || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != "POST" {
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to decide")
		return true
	}
	var body struct {
		Action string `json:"action"`
		Note   string `json:"note"`
		Target string `json:"target"`
		Model  string `json:"model"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") || d.Decode(&body) != nil || d.Decode(&struct{}{}) != io.EOF {
		fail(400, "invalid_decision", "Send a JSON object with action and optional note, target and model")
		return true
	}
	if h.decisions == nil {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return true
	}
	// An Issue's aiport fix is always for that Application only.
	is, ch, err := h.decisions.DecideIssue(id, decide.Decision{Action: decide.Action(body.Action), Note: body.Note, Scope: decide.ScopeApplication, Proposal: decide.Proposal{Target: body.Target, Model: body.Model}}, h.clock())
	if err == decide.ErrNotFound {
		fail(404, "issue_not_found", "No Issue with this id")
		return true
	}
	if err == decide.ErrNoFix {
		fail(409, "no_fix_waiting", err.Error())
		return true
	}
	if err != nil {
		decisionFailed(w, err, fail)
		return true
	}
	writeJSON(w, 200, map[string]any{"ok": true, "issue": is, "change_id": ch.ID})
	return true
}

// reportRequest handles POST /api/v1/issues {request_id, note}: the Owner
// reports one failed request from Activity to its Application's coding
// agent. The Issue carries the request's metadata, aiport's explanation and
// the Owner's note; the agent reads it with aiport_my_issues, its next
// failing call of that kind names it in X-Aiport-Hint, and Fix it in
// Changes can hand it to a coding agent aiport runs.
func (h *Handler) reportRequest(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to report a request")
		return
	}
	var body struct {
		RequestID string `json:"request_id"`
		Note      string `json:"note"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
	d.DisallowUnknownFields()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") || d.Decode(&body) != nil || d.Decode(&struct{}{}) != io.EOF || body.RequestID == "" {
		fail(400, "invalid_report", "Send a JSON object with request_id and an optional note")
		return
	}
	if h.decisions == nil {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return
	}
	e, ok := h.collector.Get(body.RequestID)
	if !ok {
		fail(404, "request_not_found", "No request with this id in request history")
		return
	}
	known := false
	for _, o := range h.live.Config().Origins {
		known = known || o.ID == e.Origin
	}
	if !known {
		fail(400, "invalid_report", decide.ErrNoApplication.Reason)
		return
	}
	rep := decide.ReportedFrom(e)
	if a := explained(e).Reason; a != nil {
		rep.Reason, rep.Explain, rep.OwnerFix = a.Summary, a.Explanation, a.Fix
	}
	rep.Note, rep.By = body.Note, config.Owner
	is, err := h.decisions.Report(e.Origin, rep, h.clock())
	if err != nil {
		decisionFailed(w, err, fail)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "issue": is})
}
