package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"aiport/internal/config"
	"aiport/internal/decide"
	"aiport/internal/extension"
	"aiport/internal/telemetry"
)

// openDecisions sets up the decision queue, keeping its items in the same
// database as Changes (this session only without one).
func (h *Handler) openDecisions() {
	var history decide.History
	if h.collector != nil {
		history = h.collector
	}
	var kept decide.Store
	if h.changesDB != nil {
		kept = decide.SQLStore{DB: h.changesDB}
	}
	q, err := decide.New(h.live, history, kept)
	if err != nil {
		// The store's errors name what failed, never tokens or request text.
		log.Printf("aiport: %v; decisions are kept for this session only", err)
		q, _ = decide.New(h.live, history, nil)
	}
	h.failedAuth = telemetry.NewFailedAuth()
	q.WatchFailedAuth(h.failedAuth)
	q.Diagnose(modelFinder{h})
	h.decisions = q
}

// Detect runs one background pass at time at, as the ticker does: it
// forgets decided items older than request history, then detects over
// request history.
func (h *Handler) Detect(at time.Time) {
	if h.decisions != nil {
		h.decisions.Prune(at)
		h.decisions.Detect(at)
		h.forgetClosedRuns()
	}
}

// StartAdapting runs detection on a background ticker until stop is closed,
// at the configured interval. The ticker runs even while the adapting loop is
// off and each pass checks whether it is on, so turning it on in aiport.yaml
// needs no restart. It does nothing when the handler has no decision queue.
// Detection never runs during a request.
func (h *Handler) StartAdapting(stop <-chan struct{}) {
	if h.decisions == nil {
		return
	}
	every := func() time.Duration {
		if h.adaptEvery > 0 {
			return h.adaptEvery
		}
		return time.Duration(h.snapshot().Adapting.IntervalSeconds) * time.Second
	}
	go h.decisions.Run(stop, every, h.clock)
}

// decisionsRoute serves the items waiting for the Owner in Changes:
//
//	GET  /api/v1/decisions       pending Suggestions and Change requests, newest first, and demand for unsupported kinds
//	POST /api/v1/decisions/{id}  {action: accept|reject|dismiss, note, scope, target, model}
//
// The caller has already verified management access.
func (h *Handler) decisionsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path == "/api/v1/decisions" {
		if r.Method != "GET" {
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		items, demand := []decide.Item{}, []decide.Demand{}
		if h.decisions != nil {
			items, demand = h.decisions.List(decide.Filter{Status: decide.Pending}), h.decisions.Demand()
		}
		writeJSON(w, 200, map[string]any{"adapting": h.decisions != nil && h.decisions.On(), "items": items, "demand": demand})
		return true
	}
	id, ok := strings.CutPrefix(r.URL.Path, "/api/v1/decisions/")
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
		Scope  string `json:"scope"`
		Alias  string `json:"alias"`
		Target string `json:"target"`
		Model  string `json:"model"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") || d.Decode(&body) != nil || d.Decode(&struct{}{}) != io.EOF {
		fail(400, "invalid_decision", "Send a JSON object with action and optional note, scope, target and model")
		return true
	}
	if h.decisions == nil {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return true
	}
	it, ch, err := h.decisions.Decide(id, decide.Decision{Action: decide.Action(body.Action), Note: body.Note, Scope: decide.Scope(body.Scope), Proposal: decide.Proposal{Alias: body.Alias, Target: body.Target, Model: body.Model}}, h.clock())
	if err != nil {
		decisionFailed(w, err, fail)
		return true
	}
	writeJSON(w, 200, map[string]any{"ok": true, "item": it, "change_id": ch.ID})
	return true
}

// decisionFailed answers an Owner decision the queue refused.
func decisionFailed(w http.ResponseWriter, err error, fail func(int, string, string)) {
	var invalid decide.Invalid
	var blocked extension.Denied
	switch {
	case errors.As(err, &blocked):
		writeError(w, 409, "guardrail_blocked", err.Error(), map[string]any{"extension": blocked.Extension, "reason": blocked.Reason})
	case errors.Is(err, decide.ErrOff):
		fail(409, "adapting_off", err.Error())
	case errors.Is(err, decide.ErrChanged):
		fail(409, "decision_conflict", err.Error())
	case errors.Is(err, decide.ErrNotFound):
		fail(404, "decision_not_found", "Nothing waits for a decision with this id")
	case errors.Is(err, decide.ErrDecided):
		fail(409, "already_decided", err.Error())
	case errors.Is(err, config.ErrReadOnly):
		fail(503, "config_readonly", "This server cannot edit its configuration")
	case errors.As(err, &invalid):
		fail(400, "invalid_decision", invalid.Reason)
	default:
		fail(400, "invalid_decision", err.Error())
	}
}
