package server

import (
	"context"
	"net/http"
	"sync"

	"aiport/internal/catalog"
	"aiport/internal/config"
)

// ModelsPath is where an Application's SDK or Coding agent looks for the
// model names it can use: aiport answers with the Application's Catalog.
const ModelsPath = "/v1/models"

// CatalogHeader carries the calling Application's Catalog version on every
// answer to a request authenticated by an Application token.
const CatalogHeader = "X-Aiport-Catalog"

// setCatalogVersions rebuilds every Application's Catalog version from c, so
// the forwarding path reads a map instead of working out a Catalog.
func (h *Handler) setCatalogVersions(c config.Config) {
	set := map[string]string{}
	for _, o := range c.Origins {
		set[o.ID] = catalog.For(c, o.ID, nil).Version
	}
	h.catalogVersions.Store(&set)
}

// catalogVersion is the Catalog version of Application application, read
// from the versions setCatalogVersions precomputed for every configured
// Application (New and every configuration Change, which also makes new
// Application tokens live). "" for an id it does not know: no Catalog is
// worked out on the forwarding path.
func (h *Handler) catalogVersion(application string) string {
	if set := h.catalogVersions.Load(); set != nil {
		return (*set)[application]
	}
	return ""
}

// catalogFor is the Catalog of application with the model lists of the
// Connections it reaches (cached lists are reused; a list that cannot be
// read contributes nothing).
func (h *Handler) catalogFor(r *http.Request, c config.Config, application string) catalog.Catalog {
	return catalog.For(c, application, h.modelLists(r.Context(), catalog.For(c, application, nil).Connections))
}

// modelLists reads the model lists of Connections ids in parallel (cached
// lists are reused). A list that cannot be read is left out.
func (h *Handler) modelLists(ctx context.Context, ids []string) catalog.Models {
	lists := make([][]string, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l := h.gateway.Models(ctx, id, false); l.Listed {
				lists[i] = l.Models
			}
		}()
	}
	wg.Wait()
	models := catalog.Models{}
	for i, id := range ids {
		if lists[i] != nil {
			models[id] = lists[i]
		}
	}
	return models
}

// listedName is one name an Application's /v1/models lists on a protocol: a
// task name or Level (connection "", owned by aiport) or a model a reached
// Connection listed.
type listedName struct {
	id, connection string
	level          bool
}

// listedOn are the names cat lists on protocol, in /v1/models order: task
// names (Levels included) whose Connection speaks protocol, then the models
// of the reached Connections that speak it. The Routing map is built with
// the same rule (speaks), so the two never disagree.
func listedOn(c config.Config, cat catalog.Catalog, protocol string) []listedName {
	var list []listedName
	for _, t := range cat.TaskNames {
		if speaks(c, t.Connection, protocol) {
			list = append(list, listedName{t.Name, "", t.Level})
		}
	}
	for _, m := range cat.Models {
		if speaks(c, m.Connection, protocol) {
			list = append(list, listedName{m.ID, m.Connection, false})
		}
	}
	return list
}

// speaks reports whether Connection id exists in c and speaks protocol.
func speaks(c config.Config, id, protocol string) bool {
	t, ok := c.Connection(id)
	return ok && t.Protocol == protocol
}

// serveModels answers GET /v1/models with the calling Application's Catalog:
// the OpenAI list shape, or the Anthropic one when the request carries
// anthropic-version. Only names usable on that protocol are listed: task
// names first (owned by aiport), then the models of the Connections the
// Application reaches. Levels are task names too, listed on the protocol
// they are mapped on. Never keys or URLs.
func (h *Handler) serveModels(w http.ResponseWriter, r *http.Request, application string, fail func(int, string, string)) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	if !h.isApplication(application) {
		fail(403, "application_token_required", "A Catalog belongs to one Application: send that Application's own token")
		return
	}
	c := h.snapshot()
	protocol := "openai"
	if r.Header.Get("anthropic-version") != "" {
		protocol = "anthropic"
	}
	list := listedOn(c, h.catalogFor(r, c, application), protocol)
	if protocol == "anthropic" {
		data := []map[string]any{}
		for _, e := range list {
			name := e.id
			switch {
			case e.level:
				name += " (aiport Level)"
			case e.connection == "":
				name += " (aiport task name)"
			}
			data = append(data, map[string]any{"type": "model", "id": e.id, "display_name": name, "created_at": "1970-01-01T00:00:00Z"})
		}
		body := map[string]any{"data": data, "has_more": false, "first_id": nil, "last_id": nil}
		if len(list) > 0 {
			body["first_id"], body["last_id"] = list[0].id, list[len(list)-1].id
		}
		writeJSON(w, 200, body)
		return
	}
	data := []map[string]any{}
	for _, e := range list {
		owner := e.connection
		if owner == "" {
			owner = "aiport" // a task name
		}
		data = append(data, map[string]any{"id": e.id, "object": "model", "created": 0, "owned_by": owner})
	}
	writeJSON(w, 200, map[string]any{"object": "list", "data": data})
}
