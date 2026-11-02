package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"aiport/internal/agentrun"
	"aiport/internal/config"
	"aiport/internal/decide"
	"aiport/internal/explain"
	"aiport/internal/telemetry"
)

// ExplainTimeout bounds one explanation written by a Connection.
const ExplainTimeout = 2 * time.Minute

// explainer is who writes the explanation of an Issue of Application app:
// a Coding agent (agent, an Agent run) or a Connection (w). Without either,
// code and hint say why Explain is unavailable.
//
// This is the one place that chooses the writer: an Application with a
// connected Coding agent gets an Agent run here (spec #21, Explain by a
// Coding agent), every other Application the Connection and model the Owner
// chose in Settings (adapting.explain).
func (h *Handler) explainer(app string) (w explain.Writer, agent *agentWriter, code, hint string) {
	if h.decisions == nil || !h.decisions.On() {
		return nil, nil, "adapting_off", "Explain is unavailable while adapting is turned off in the configuration."
	}
	if a := h.agentExplainer(app, h.codingAgentOf(app)); a != nil {
		return nil, a, "", ""
	}
	c := h.live.Config()
	e := c.Adapting.Explain
	if e.Connection == "" {
		return nil, nil, "explainer_not_chosen", "Choose the Connection and model that write explanations in Settings first."
	}
	protocol := ""
	for _, t := range c.Targets {
		if t.ID == e.Connection {
			protocol = t.Protocol
		}
	}
	return explain.Connection{Gateway: h.gateway, Connection: e.Connection, Model: e.Model, Protocol: protocol, Record: h.recordOwn}, nil, "", ""
}

// recordOwn records one of aiport's own model calls for Activity.
func (h *Handler) recordOwn(e telemetry.Event) {
	e.ID = requestID()
	if h.collector != nil {
		h.collector.Offer(e)
	}
}

// explainStatus is what GET /api/v1/issues says about Explain.
func (h *Handler) explainStatus() map[string]any {
	w, _, _, hint := h.explainer("")
	if w == nil {
		return map[string]any{"available": false, "hint": hint}
	}
	by := w.By()
	return map[string]any{"available": true, "source": by.Source, "connection": by.Connection, "model": by.Model}
}

// explainRoute explains an Issue, on the Owner's click:
//
//	POST /api/v1/issues/{id}/explain
//
// The answer streams as JSON lines (application/x-ndjson), one event each:
//
//	{"type":"start","by":{source, connection, model, agent}}  who writes it
//	{"type":"text","text":"…"}                               the next words, as they are written
//	{"type":"done","issue":{…}}                              saved: the Issue with its explanation
//	{"type":"error","error":{code, message}}                 nothing saved; an earlier explanation stays
//
// Refusals before anything is written are ordinary JSON errors: 404
// issue_not_found, 409 adapting_off, explainer_not_chosen or
// explanation_running (one explanation per Application at a time). The
// caller has already verified management access.
func (h *Handler) explainRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/issues/")
	id, ok2 := strings.CutSuffix(rest, "/explain")
	if !ok || !ok2 || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to explain an Issue")
		return true
	}
	if h.decisions == nil || !h.decisions.On() {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return true
	}
	is, found := h.decisions.Issue(id)
	if !found {
		fail(404, "issue_not_found", "No Issue with this id")
		return true
	}
	writer, agent, code, hint := h.explainer(is.Application)
	if writer == nil && agent == nil {
		fail(409, code, hint)
		return true
	}
	if _, going := h.runner().Active(is.Application); going {
		fail(409, "agent_run_running", agentrun.ErrBusy.Error())
		return true
	}
	if agent != nil {
		h.explainByAgent(w, r, agent, is, fail)
		return true
	}
	if !h.startExplaining(is.Application) {
		fail(409, "explanation_running", "An explanation for this Application is being written; wait for it to finish")
		return true
	}
	defer h.stopExplaining(is.Application)

	send := stream(w)
	send(map[string]any{"type": "start", "by": writer.By()})
	ctx, cancel := context.WithTimeout(r.Context(), ExplainTimeout)
	e, err := writer.Explain(ctx, is, func(text string) { send(map[string]any{"type": "text", "text": text}) })
	cancel()
	if err == nil {
		is, err = h.decisions.Explain(id, e, h.clock())
	}
	h.sendOutcome(send, is, err, ExplainTimeout)
	return true
}

// explainByAgent starts Explain by a Coding agent and streams its run until
// it ends or the request does: the run goes on without the request (the
// dashboard follows it again with GET /api/v1/issues/{id}/run) and saves its
// explanation when it ends.
func (h *Handler) explainByAgent(w http.ResponseWriter, r *http.Request, agent *agentWriter, is decide.Issue, fail func(int, string, string)) {
	run, err := agent.start(is)
	switch {
	case errors.Is(err, agentrun.ErrBusy):
		fail(409, "agent_run_running", err.Error())
		return
	case err != nil:
		fail(409, "explanation_failed", "No explanation: "+err.Error())
		return
	}
	send := stream(w)
	send(map[string]any{"type": "start", "by": agent.By()})
	send(map[string]any{"type": "run", "run": run.Info()})
	if run.Follow(r.Context(), func(u agentrun.Update) {
		switch u.Kind {
		case agentrun.UpdateText:
			send(map[string]any{"type": "text", "text": u.Text})
		case agentrun.UpdateStatus:
		default:
			send(map[string]any{"type": "step", "step": u})
		}
	}) != nil {
		return // the request ended; the run goes on
	}
	<-run.Done()
	h.sendOutcome(send, run.issue, run.err, agent.limit)
}

// stream starts a JSON-lines answer and returns how to send one event.
func stream(w http.ResponseWriter) func(any) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(200)
	rc := http.NewResponseController(w)
	enc := json.NewEncoder(w)
	return func(v any) {
		_ = enc.Encode(v)
		_ = rc.Flush()
	}
}

// sendOutcome sends the last event of an explain stream: done with the
// saved Issue, or the error.
func (h *Handler) sendOutcome(send func(any), is decide.Issue, err error, limit time.Duration) {
	if err != nil {
		code, message := runRefusal(err, limit)
		send(map[string]any{"type": "error", "error": map[string]string{"code": code, "message": message}})
		return
	}
	send(map[string]any{"type": "done", "issue": is})
}

func (h *Handler) startExplaining(app string) bool {
	h.explainMu.Lock()
	defer h.explainMu.Unlock()
	if h.explaining == nil {
		h.explaining = map[string]bool{}
	}
	if h.explaining[app] {
		return false
	}
	h.explaining[app] = true
	return true
}

func (h *Handler) stopExplaining(app string) {
	h.explainMu.Lock()
	defer h.explainMu.Unlock()
	delete(h.explaining, app)
}

// explainerRoute is the Settings choice of who writes explanations:
//
//	PUT /api/v1/adapting/explainer  {connection, model}: sets adapting.explain as a revertable Change; both empty turns Explain off
//
// The caller has already verified management access.
func (h *Handler) explainerRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path != "/api/v1/adapting/explainer" {
		return false
	}
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		fail(405, "method_not_allowed", "Use PUT")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to choose who writes explanations")
		return true
	}
	var in config.Explainer
	if decodeJSON(w, r, &in) != nil {
		fail(400, "invalid_request", "Send a JSON body with connection and model")
		return true
	}
	in.Connection, in.Model = strings.TrimSpace(in.Connection), strings.TrimSpace(in.Model)
	what := "explanations off"
	if in.Connection != "" {
		what = "explanations written by " + in.Connection + " (" + in.Model + ")"
	}
	fresh, ch, err := h.live.Change(what, config.SetExplainer(in))
	switch {
	case errors.Is(err, config.ErrReadOnly):
		fail(503, "config_readonly", "This server cannot edit its configuration")
	case err != nil:
		fail(400, "invalid_explainer", err.Error())
	default:
		writeJSON(w, 200, map[string]any{"explainer": fresh.Adapting.Explain, "change_id": ch.ID})
	}
	return true
}
