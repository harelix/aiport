package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/harelix/aiport/internal/config"
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
//	PUT    /api/v1/aliases/effort                  {application?, alias, connection?, effort} (serveAliasEffort)
//	PUT    /api/v1/task-names/{name}               {connection, model, effort?, everywhere?, reasoning?} (serveTaskName)
//	DELETE /api/v1/task-names/{name}               {connection}
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
	switch {
	case r.URL.Path == "/api/v1/levels":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		writeJSON(w, 200, levelsOf(h.snapshot()))
		return true
	case strings.HasPrefix(r.URL.Path, ModuleRoutesPath+"/"):
		h.serveModuleRoute(w, r, strings.TrimPrefix(r.URL.Path, ModuleRoutesPath+"/"), fail)
		return true
	case strings.HasPrefix(r.URL.Path, TaskNamesPath+"/"):
		h.serveTaskName(w, r, strings.TrimPrefix(r.URL.Path, TaskNamesPath+"/"), fail)
		return true
	case r.URL.Path == "/api/v1/aliases/effort":
		h.serveAliasEffort(w, r, fail)
		return true
	}
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/levels/")
	if !ok {
		return false
	}
	protocol, level, found := strings.Cut(rest, "/")
	if !found || strings.Contains(level, "/") {
		return false
	}
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "PUT, DELETE")
		fail(405, "method_not_allowed", "Use PUT or DELETE")
		return true
	}
	in, ok := h.levelEditInput(w, r, fail)
	if !ok {
		return true
	}
	if r.Method == http.MethodDelete {
		in.Connection, in.Model, in.Effort = "", "", nil
	} else if in.Connection == "" || in.Model == "" {
		fail(400, "invalid_request", "Send connection and model (DELETE clears a mapping)")
		return true
	}
	edit, err := config.SetLevel(h.snapshot(), config.LevelEdit{Application: in.Application, Protocol: protocol, Level: level, Connection: in.Connection, Model: in.Model, Effort: in.Effort})
	if err != nil {
		fail(400, "invalid_level", err.Error())
		return true
	}
	what := "clear Level " + level + " on " + config.ProtocolName(protocol) + " for " + forWhom(in.Application)
	if in.Connection != "" {
		what = "map Level " + level + " on " + config.ProtocolName(protocol) + " to " + in.Model + " on " + in.Connection + " for " + forWhom(in.Application)
		if in.Effort != nil {
			if e := effortList(*in.Effort); e != "" {
				what += " (" + e + ")"
			} else {
				what += ", no per-effort models"
			}
		}
	}
	h.changeLevels(w, what, edit, fail)
	return true
}

// serveAliasEffort handles PUT /api/v1/aliases/effort: the per-effort
// models of one Alias, Level mapping or task name, the Application's own or
// shared (connection names which entry; required for shared Aliases and
// Levels). An empty effort clears them.
func (h *Handler) serveAliasEffort(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", "PUT")
		fail(405, "method_not_allowed", "Use PUT")
		return
	}
	in, ok := h.levelEditInput(w, r, fail)
	if !ok {
		return
	}
	if in.Effort == nil {
		fail(400, "invalid_request", "Send effort with low, medium and high (empty clears them)")
		return
	}
	edit, err := config.SetEffort(h.snapshot(), in.Application, in.Alias, in.Connection, *in.Effort)
	if err != nil {
		fail(400, "invalid_alias", err.Error())
		return
	}
	what := "per-effort models of " + in.Alias
	if in.Connection != "" {
		what += " on " + in.Connection
	}
	e := effortList(*in.Effort)
	if e == "" {
		e = "none"
	}
	h.changeLevels(w, what+" for "+forWhom(in.Application)+": "+e, edit, fail)
}

// levelEditInput is the body of a Level or per-effort edit, after the
// dashboard and writability checks; ok is false once refused.
func (h *Handler) levelEditInput(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) (in levelEditBody, ok bool) {
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change Levels")
		return in, false
	}
	if !h.live.Writable() {
		fail(503, "config_readonly", "This server cannot edit its configuration")
		return in, false
	}
	if r.ContentLength != 0 || r.Method != http.MethodDelete {
		if decodeJSON(w, r, &in) != nil {
			fail(400, "invalid_request", "Send a JSON body with application, connection, model and effort")
			return in, false
		}
	}
	in.Application, in.Alias, in.Connection, in.Model = strings.TrimSpace(in.Application), strings.TrimSpace(in.Alias), strings.TrimSpace(in.Connection), strings.TrimSpace(in.Model)
	if _, found := h.snapshot().Application(in.Application); in.Application != "" && !found {
		fail(404, "origin_not_found", "No Application with this id")
		return in, false
	}
	return in, true
}

type levelEditBody struct {
	Application string               `json:"application"`
	Alias       string               `json:"alias"`
	Connection  string               `json:"connection"`
	Model       string               `json:"model"`
	Effort      *config.EffortModels `json:"effort"`
}

// changeLevels makes edit as one Change described by what and answers its id.
func (h *Handler) changeLevels(w http.ResponseWriter, what string, edit config.Edit, fail func(int, string, string)) {
	_, change, err := h.live.Change(what, edit)
	if err != nil {
		fail(400, "invalid_level", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "change_id": change.ID})
}

// forWhom names who a mapping is for in a Change: one Application, or
// every Application for a shared one.
func forWhom(application string) string {
	if application == "" {
		return "every Application"
	}
	return application
}

// effortList lists per-effort models for the Changes list ("" when none).
func effortList(e config.EffortModels) string {
	var parts []string
	for _, p := range [][2]string{{"low", e.Low}, {"medium", e.Medium}, {"high", e.High}} {
		if v := strings.TrimSpace(p[1]); v != "" {
			parts = append(parts, p[0]+" "+v)
		}
	}
	return strings.Join(parts, ", ")
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

// TaskNamesPath is where the Owner adds, changes and removes shared task
// names.
const TaskNamesPath = "/api/v1/task-names"

// serveTaskName handles PUT and DELETE /api/v1/task-names/{name}: a shared
// task name on one Connection, optionally usable by every Application
// (everywhere) and with a fixed reasoning effort. Each edit is one revertable
// Change.
func (h *Handler) serveTaskName(w http.ResponseWriter, r *http.Request, name string, fail func(int, string, string)) {
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "PUT, DELETE")
		fail(405, "method_not_allowed", "Use PUT or DELETE")
		return
	}
	var in struct {
		Connection string               `json:"connection"`
		Model      string               `json:"model"`
		Effort     *config.EffortModels `json:"effort"`
		Everywhere bool                 `json:"everywhere"`
		Reasoning  string               `json:"reasoning"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&in); err != nil {
		fail(400, "invalid_request", "Send {connection, model, effort?, everywhere?, reasoning?} as JSON")
		return
	}
	if r.Method == http.MethodDelete {
		in.Model = ""
	} else if in.Connection == "" || in.Model == "" {
		fail(400, "invalid_request", "Send connection and model (DELETE removes a task name)")
		return
	}
	edit, err := config.SetTaskName(h.snapshot(), config.TaskNameEdit{Name: name, Connection: in.Connection, Model: in.Model, Effort: in.Effort, Everywhere: in.Everywhere, Reasoning: in.Reasoning})
	if err != nil {
		fail(400, "invalid_task_name", err.Error())
		return
	}
	what := "remove task name " + name + " from " + in.Connection
	if in.Model != "" {
		what = "task name " + name + " to " + in.Model + " on " + in.Connection
		if in.Everywhere {
			what += ", for every Application"
		}
		if in.Reasoning != "" {
			what += ", reasoning always " + in.Reasoning
		}
	}
	_, change, err := h.live.Change(what, edit)
	if err != nil {
		fail(400, "invalid_task_name", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "change_id": change.ID})
}

// ModuleRoutesPath is where the Owner sends one Application module's calls
// for one name to another model.
const ModuleRoutesPath = "/api/v1/module-routes"

// serveModuleRoute handles PUT and DELETE
// /api/v1/module-routes/{application}/{module}/{name} with {connection,
// model, why?}: why (one line, such as the votes that suggested it) goes into
// the Change. Each edit is one revertable Change.
func (h *Handler) serveModuleRoute(w http.ResponseWriter, r *http.Request, rest string, fail func(int, string, string)) {
	parts := strings.SplitN(rest, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		fail(404, "not_found", "Use /api/v1/module-routes/{application}/{module}/{name}")
		return
	}
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "PUT, DELETE")
		fail(405, "method_not_allowed", "Use PUT or DELETE")
		return
	}
	var in struct {
		Connection string `json:"connection"`
		Model      string `json:"model"`
		Why        string `json:"why"`
	}
	if r.ContentLength != 0 {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&in); err != nil {
			fail(400, "invalid_request", "Send {connection, model, why?} as JSON")
			return
		}
	}
	app, module, name := parts[0], parts[1], parts[2]
	if r.Method == http.MethodDelete {
		in.Model = ""
	} else if in.Connection == "" || in.Model == "" {
		fail(400, "invalid_request", "Send connection and model (DELETE removes the route)")
		return
	}
	edit, err := config.SetModuleRoute(h.snapshot(), app, module, name, in.Connection, in.Model)
	if err != nil {
		fail(400, "invalid_module_route", err.Error())
		return
	}
	what := "stop routing module " + module + " of " + app + " for " + name + " separately"
	if in.Model != "" {
		what = "route module " + module + " of " + app + " for " + name + " to " + in.Model + " on " + in.Connection
	}
	if why := strings.Join(strings.Fields(in.Why), " "); why != "" {
		if len(why) > 300 {
			why = strings.ToValidUTF8(why[:300], "")
		}
		what += " (" + why + ")"
	}
	_, change, err := h.live.Change(what, edit)
	if err != nil {
		fail(400, "invalid_module_route", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "change_id": change.ID})
}
