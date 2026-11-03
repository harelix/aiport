package server

import (
	"net/http"

	"aiport/internal/catalog"
	"aiport/internal/config"
)

// RoutingMapPath answers the Owner's Routing map: for every Application,
// each name it can send and the Connection and model serving it.
const RoutingMapPath = "/api/v1/routing"

// Route is one row of the Routing map: a name an Application can send on a
// protocol and the Connection and model serving it. Kind is "task_name" (an
// Alias, the Application's own or shared), "level" (fast, balanced, best)
// or "default" (any other model name, sent unchanged to the binding or the
// protocol default; Name and Model are empty). The Catalog's task name is
// embedded, so whatever a task name carries (such as per-effort models)
// reaches the map unchanged.
type Route struct {
	catalog.TaskName
	Kind     string `json:"kind"`
	Protocol string `json:"protocol"`
}

// RoutingEntry is one Application in the Routing map: its Catalog (exactly
// what its /v1/models lists), its name, its routes per protocol and, for one
// Application asked for alone, its recent Catalog changes.
type RoutingEntry struct {
	catalog.Catalog
	Name    string           `json:"name"`
	Routes  []Route          `json:"routes"`
	Changes []catalog.Record `json:"changes,omitempty"`
}

// routesOf are cat's routes in /v1/models order: each task name and Level
// on the protocol its Connection speaks (the rule listedOn uses), then the
// Connection any other name goes to on each protocol it can be served on.
func routesOf(c config.Config, cat catalog.Catalog) []Route {
	routes := []Route{}
	for _, t := range cat.TaskNames {
		for _, p := range config.Protocols {
			if !speaks(c, t.Connection, p) {
				continue
			}
			kind := "task_name"
			if t.Level {
				kind = "level"
			}
			routes = append(routes, Route{TaskName: t, Kind: kind, Protocol: p})
		}
	}
	for _, p := range config.Protocols {
		id := cat.Binding
		if id == "" {
			id = cat.Defaults[p]
		}
		if id != "" && speaks(c, id, p) {
			routes = append(routes, Route{TaskName: catalog.TaskName{Connection: id}, Kind: "default", Protocol: p})
		}
	}
	return routes
}

// serveRoutingMap handles GET /api/v1/routing (management token only, as
// every /api/ route): every configured Application's Catalog with its
// routes, built by catalog.For with the same model lists /v1/models reads,
// so the two never disagree. ?application=<id> answers that one Application
// with its recent Catalog changes (the Integrate view). Application ids,
// names, task names, model names and Connection ids only: never keys, URLs
// or tokens. The built-in Application is left out: it is aiport's own.
func (h *Handler) serveRoutingMap(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	c := h.snapshot()
	origins := c.Origins
	one := r.URL.Query().Get("application")
	if one != "" {
		o, ok := c.Application(one)
		if !ok || one == config.BuiltinApplication {
			fail(404, "unknown_application", "No Application has this id")
			return
		}
		origins = []config.Origin{o}
	}
	reached := []string{}
	for _, o := range origins {
		for _, id := range catalog.For(c, o.ID, nil).Connections {
			if !contains(reached, id) {
				reached = append(reached, id)
			}
		}
	}
	models := h.modelLists(r.Context(), reached)
	out := []RoutingEntry{}
	for _, o := range origins {
		cat := catalog.For(c, o.ID, models)
		e := RoutingEntry{Catalog: cat, Name: o.Name, Routes: routesOf(c, cat)}
		if one != "" {
			e.Changes = h.recentCatalogChanges(o.ID, catalogChangesLimit)
		}
		out = append(out, e)
	}
	writeJSON(w, 200, map[string]any{"applications": out})
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
