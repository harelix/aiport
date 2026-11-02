package server

import (
	"net/http"
	"strings"

	"aiport/internal/config"
)

// originRoute changes one Application's settings for the dashboard:
//
//	PATCH /api/v1/origins/{id}   {may_request_changes}
//
// The change goes through the configuration-change module, so it is live at
// once and recorded as a revertable Change, whose id is returned. The caller
// has already verified management access.
func (h *Handler) originRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	id, ok := strings.CutPrefix(r.URL.Path, "/api/v1/origins/")
	if !ok || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", "PATCH")
		fail(405, "method_not_allowed", "Use PATCH")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change Applications")
		return true
	}
	if !h.live.Writable() {
		fail(503, "config_readonly", "This server cannot edit its configuration")
		return true
	}
	if !h.isApplication(id) {
		fail(404, "origin_not_found", "No Application with this id")
		return true
	}
	var in struct {
		MayRequestChanges *bool `json:"may_request_changes"`
	}
	if decodeJSON(w, r, &in) != nil || in.MayRequestChanges == nil {
		fail(400, "invalid_request", "Send a JSON body with may_request_changes (true or false)")
		return true
	}
	what := "forbid " + id + " to request changes"
	if *in.MayRequestChanges {
		what = "allow " + id + " to request changes"
	}
	_, change, err := h.live.Change(what, config.SetMayRequestChanges(id, *in.MayRequestChanges))
	if err != nil {
		fail(400, "invalid_client", err.Error())
		return true
	}
	writeJSON(w, 200, map[string]any{"ok": true, "change_id": change.ID})
	return true
}
