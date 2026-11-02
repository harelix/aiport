package server

import (
	"errors"
	"net/http"

	"aiport/internal/config"
)

// autoFixRoute is the Changes page switch for automatic fixes:
//
//	PUT /api/v1/adapting/auto-fix  {enabled: bool}: sets adapting.auto_fix as a revertable Change
//
// The caller has already verified management access.
func (h *Handler) autoFixRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path != "/api/v1/adapting/auto-fix" {
		return false
	}
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		fail(405, "method_not_allowed", "Use PUT")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to switch automatic fixes")
		return true
	}
	var in struct {
		Enabled *bool `json:"enabled"`
	}
	if decodeJSON(w, r, &in) != nil || in.Enabled == nil {
		fail(400, "invalid_request", "Send a JSON body with enabled: true or false")
		return true
	}
	what := "automatic fixes off"
	if *in.Enabled {
		what = "automatic fixes on"
	}
	fresh, ch, err := h.live.Change(what, config.SetAutoFix(*in.Enabled))
	switch {
	case errors.Is(err, config.ErrReadOnly):
		fail(503, "config_readonly", "This server cannot edit its configuration")
	case err != nil:
		fail(400, "invalid_change", err.Error())
	default:
		writeJSON(w, 200, map[string]any{"auto_fix": fresh.Adapting.AutoFix, "change_id": ch.ID})
	}
	return true
}
