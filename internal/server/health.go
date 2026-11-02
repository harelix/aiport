package server

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const probeTimeout = 15 * time.Second

// targetCheckRoute runs an explicit, non-generating health check:
//
//	POST /api/v1/targets/{id}/check
//
// The caller has already verified management access. Nothing runs in the
// background: checks happen only when the owner asks.
func (h *Handler) targetCheckRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/targets/")
	if !ok {
		return false
	}
	id, ok := strings.CutSuffix(rest, "/check")
	if !ok || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to check connections")
		return true
	}
	found := false
	for _, t := range h.snapshot().Targets {
		found = found || t.ID == id
	}
	if !found {
		fail(404, "target_not_found", "No connection with this id")
		return true
	}
	ctx, cancel := context.WithTimeout(r.Context(), probeTimeout)
	defer cancel()
	writeJSON(w, 200, h.gateway.Probe(ctx, id))
	return true
}

// targetModelsRoute lists the models one Connection serves, from the same
// free GET /models listing the check uses:
//
//	GET /api/v1/targets/{id}/models[?refresh=1]
//
// A recent list is reused unless refresh is set. Only model ids are
// returned, never provider messages or keys.
func (h *Handler) targetModelsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/targets/")
	if !ok {
		return false
	}
	id, ok := strings.CutSuffix(rest, "/models")
	if !ok || id == "" || strings.Contains(id, "/") {
		return false
	}
	if !modelsRequest(w, r, fail) {
		return true
	}
	found := false
	for _, t := range h.snapshot().Targets {
		found = found || t.ID == id
	}
	if !found {
		fail(404, "target_not_found", "No connection with this id")
		return true
	}
	ctx, cancel := context.WithTimeout(r.Context(), probeTimeout)
	defer cancel()
	writeJSON(w, 200, h.gateway.Models(ctx, id, r.URL.Query().Get("refresh") == "1"))
	return true
}

// servingRoute answers which enabled Connections of a protocol serve a
// model: one, none, several or unknown.
//
//	GET /api/v1/serving?protocol=openai&model=llama3:8b[&refresh=1]
func (h *Handler) servingRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path != "/api/v1/serving" {
		return false
	}
	if !modelsRequest(w, r, fail) {
		return true
	}
	q := r.URL.Query()
	protocol, model := q.Get("protocol"), q.Get("model")
	if protocol != "openai" && protocol != "anthropic" {
		fail(400, "invalid_protocol", "Give protocol openai or anthropic")
		return true
	}
	if model == "" || len(model) > 256 {
		fail(400, "invalid_model", "Give the model id to look for")
		return true
	}
	ctx, cancel := context.WithTimeout(r.Context(), probeTimeout)
	defer cancel()
	if q.Get("refresh") == "1" {
		for _, t := range h.snapshot().Targets {
			if !t.Disabled && t.Protocol == protocol {
				h.gateway.Models(ctx, t.ID, true)
			}
		}
	}
	writeJSON(w, 200, h.gateway.Serves(ctx, protocol, model))
	return true
}

// modelsRequest admits a GET from the same-origin dashboard: listing calls
// providers, so a cross-site page must not trigger it.
func modelsRequest(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return false
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to list models")
		return false
	}
	return true
}
