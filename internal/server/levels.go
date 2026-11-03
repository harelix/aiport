package server

import (
	"net/http"
	"strings"

	"aiport/internal/config"

	"go.yaml.in/yaml/v3"
)

// levelMapping is one Level mapping as the dashboard reads it.
type levelMapping struct {
	Level      string              `json:"level"`
	Protocol   string              `json:"protocol"`
	Connection string              `json:"connection"`
	Model      string              `json:"model"`
	Effort     config.EffortModels `json:"effort"`
}

// levelsRoute lets the Owner edit Levels and per-effort models:
//
//	GET    /api/v1/levels                          the shared and per-Application mappings, and the Connections
//	PUT    /api/v1/levels/{protocol}/{level}       {application?, connection, model, effort?}
//	DELETE /api/v1/levels/{protocol}/{level}       {application?} (body optional)
//	PUT    /api/v1/aliases/effort                  {application?, alias, connection?, effort}
//
// application "" (or left out) is the shared mapping, otherwise that
// Application's own, which comes before the shared one. PUT on a Level sets
// or changes its mapping; effort left out keeps the per-effort models while
// the mapping stays on the same Connection, and drops them otherwise.
// aliases/effort sets the per-effort models of any Alias, a Level mapping
// or a task name (an empty effort clears them). Each edit is one revertable
// Change, whose id is returned; the Catalog changes it causes are recorded
// like any other. The caller has already verified management access.
func (h *Handler) levelsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	var protocol, level string
	switch {
	case r.URL.Path == "/api/v1/levels":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		writeJSON(w, 200, levelsOf(h.snapshot()))
		return true
	case r.URL.Path == "/api/v1/aliases/effort":
		if r.Method != http.MethodPut {
			w.Header().Set("Allow", "PUT")
			fail(405, "method_not_allowed", "Use PUT")
			return true
		}
	default:
		rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/levels/")
		if !ok {
			return false
		}
		var found bool
		protocol, level, found = strings.Cut(rest, "/")
		if !found || strings.Contains(level, "/") {
			return false
		}
		if r.Method != http.MethodPut && r.Method != http.MethodDelete {
			w.Header().Set("Allow", "PUT, DELETE")
			fail(405, "method_not_allowed", "Use PUT or DELETE")
			return true
		}
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change Levels")
		return true
	}
	if !h.live.Writable() {
		fail(503, "config_readonly", "This server cannot edit its configuration")
		return true
	}
	var in struct {
		Application string               `json:"application"`
		Alias       string               `json:"alias"`
		Connection  string               `json:"connection"`
		Model       string               `json:"model"`
		Effort      *config.EffortModels `json:"effort"`
	}
	if r.ContentLength != 0 || r.Method != http.MethodDelete {
		if decodeJSON(w, r, &in) != nil {
			fail(400, "invalid_request", "Send a JSON body with application, connection, model and effort")
			return true
		}
	}
	in.Application, in.Alias, in.Connection, in.Model = strings.TrimSpace(in.Application), strings.TrimSpace(in.Alias), strings.TrimSpace(in.Connection), strings.TrimSpace(in.Model)
	current := h.snapshot()
	if _, ok := current.Application(in.Application); in.Application != "" && !ok {
		fail(404, "origin_not_found", "No Application with this id")
		return true
	}
	who := "shared"
	if in.Application != "" {
		who = in.Application + ":"
	}
	var (
		edits []config.Edit
		what  string
	)
	if level == "" { // aliases/effort
		if in.Effort == nil {
			fail(400, "invalid_request", "Send effort with low, medium and high (empty clears them)")
			return true
		}
		edit, err := config.SetEffort(current, in.Application, in.Alias, in.Connection, *in.Effort)
		if err != nil {
			fail(400, "invalid_alias", err.Error())
			return true
		}
		edits, what = append(edits, edit), who+" per-effort models of "+in.Alias+effortWhat(*in.Effort)
	} else {
		if r.Method == http.MethodDelete {
			in.Connection, in.Model, in.Effort = "", "", nil
		} else if in.Connection == "" || in.Model == "" {
			fail(400, "invalid_request", "Send connection and model (DELETE clears a mapping)")
			return true
		}
		edit, err := config.SetLevel(current, in.Application, protocol, level, in.Connection, in.Model)
		if err != nil {
			fail(400, "invalid_level", err.Error())
			return true
		}
		edits = append(edits, edit)
		if in.Connection == "" {
			what = who + " clear Level " + level + " on " + protocol
		} else {
			what = who + " map Level " + level + " on " + protocol + " to " + in.Model + " on " + in.Connection
		}
		if in.Effort != nil {
			// The mapping is in place once the first edit ran; the
			// per-effort models go on it in the same Change.
			effort := *in.Effort
			edits = append(edits, config.EffortEdit(in.Application, level, in.Connection, effort))
			what += effortWhat(effort)
		}
	}
	what = strings.Replace(what, "shared ", "", 1)
	_, change, err := h.live.Change(what, func(root *yaml.Node) error {
		for _, e := range edits {
			if err := e(root); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fail(400, "invalid_level", err.Error())
		return true
	}
	writeJSON(w, 200, map[string]any{"ok": true, "change_id": change.ID})
	return true
}

// effortWhat describes per-effort models for the Changes list.
func effortWhat(e config.EffortModels) string {
	if e.IsZero() {
		return ", no per-effort models"
	}
	var parts []string
	for _, p := range [][2]string{{"low", e.Low}, {"medium", e.Medium}, {"high", e.High}} {
		if p[1] != "" {
			parts = append(parts, p[0]+" "+p[1])
		}
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

// levelsOf lists every Level mapping, the shared ones and each
// Application's own, with the Connections a mapping may use.
func levelsOf(c config.Config) map[string]any {
	list := func(aliases []config.Alias) []levelMapping {
		out := []levelMapping{}
		for _, a := range aliases {
			if t, ok := c.Connection(a.Target); ok && config.IsLevel(a.Alias) {
				out = append(out, levelMapping{Level: a.Alias, Protocol: t.Protocol, Connection: a.Target, Model: a.Model, Effort: a.Effort})
			}
		}
		return out
	}
	applications := map[string][]levelMapping{}
	for _, o := range c.Origins {
		own := make([]config.Alias, 0, len(o.Aliases))
		for _, a := range o.Aliases {
			own = append(own, config.Alias{Target: a.Target, Alias: a.Alias, Model: a.Model, Effort: a.Effort})
		}
		applications[o.ID] = list(own)
	}
	type connection struct {
		ID       string          `json:"id"`
		Provider string          `json:"provider"`
		Protocol string          `json:"protocol"`
		Location config.Location `json:"location"`
		Disabled bool            `json:"disabled"`
	}
	connections := []connection{}
	for _, t := range c.Targets {
		connections = append(connections, connection{ID: t.ID, Provider: t.Provider, Protocol: t.Protocol, Location: t.Location, Disabled: t.Disabled})
	}
	return map[string]any{"levels": config.Levels, "protocols": config.Protocols, "efforts": config.Efforts, "shared": list(c.Aliases), "applications": applications, "connections": connections}
}
