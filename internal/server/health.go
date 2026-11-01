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
