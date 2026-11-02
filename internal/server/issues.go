package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"aiport/internal/decide"
)

// modelFinder adapts the gateway's model listings to the decision module's
// question "which enabled Connections serve this model?", adding the models
// Connections of that protocol list (cached listings, nothing generated).
type modelFinder struct{ h *Handler }

func (f modelFinder) Serves(ctx context.Context, protocol, model string) decide.Diagnosis {
	s := f.h.gateway.Serves(ctx, protocol, model)
	d := decide.Diagnosis{Protocol: s.Protocol, Answer: s.Answer, Serving: s.Serving, Unknown: s.Unknown}
	seen := map[string]bool{}
	for _, t := range f.h.snapshot().Targets {
		if t.Disabled || t.Protocol != protocol {
			continue
		}
		for _, id := range f.h.gateway.Models(ctx, t.ID, false).Models {
			if !seen[id] {
				seen[id] = true
				d.Available = append(d.Available, id)
			}
		}
	}
	sort.Strings(d.Available)
	return d
}

// issuesRoute serves Issues to the Owner:
//
//	GET  /api/v1/issues?state=&application=  Issues, newest first; state is open (default: open and resolving), resolving, resolved, dismissed or all
//	POST /api/v1/issues/{id}                 {action: accept|reject|dismiss, note, target, model}: decide the aiport fix
//
// The caller has already verified management access.
func (h *Handler) issuesRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path == "/api/v1/issues" {
		if r.Method != "GET" {
			fail(405, "method_not_allowed", "Use GET")
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
		writeJSON(w, 200, map[string]any{"adapting": h.decisions != nil && h.decisions.On(), "auto_fix": h.live.Config().Adapting.AutoFix, "issues": issues})
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
