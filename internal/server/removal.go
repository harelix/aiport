package server

import (
	"context"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"

	"aiport/internal/catalog"
	"aiport/internal/config"
)

// removalPlan is the body of DELETE /api/v1/targets/{id} and of the
// dependents preview: each dependent's replacement Connection ("" = none).
// A nil Plan means none was sent.
type removalPlan struct {
	Plan *map[string]string `json:"plan"`
}

// dependent is one item using a Connection with its suggested replacement.
type dependent struct {
	config.Dependent
	Suggestion string `json:"suggestion"`
}

// affectedApplication is one Application whose Catalog a removal changes.
type affectedApplication struct {
	catalog.Change
	Name string `json:"name"`
}

// targetDependentsRoute lists what uses a Connection before it is removed:
//
//	GET  /api/v1/targets/{id}/dependents          every dependent with a suggested replacement, and the Applications the suggested plan affects
//	POST /api/v1/targets/{id}/dependents {plan}   the same, with the Applications the given plan affects
//
// Nothing is changed. Only ids, task names and model names are returned,
// never keys or URLs. Management access is already verified.
func (h *Handler) targetDependentsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/targets/")
	if !ok {
		return false
	}
	id, ok := strings.CutSuffix(rest, "/dependents")
	if !ok || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		fail(405, "method_not_allowed", "Use GET or POST")
		return true
	}
	if r.Method == http.MethodPost && r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to plan a removal")
		return true
	}
	current := h.snapshot()
	if _, ok := current.Connection(id); !ok {
		fail(404, "target_not_found", "No connection with this id")
		return true
	}
	suggested := h.suggestReplacements(r.Context(), current, id)
	deps := []dependent{}
	for _, d := range config.Dependents(current, id) {
		deps = append(deps, dependent{d, suggested[d.ID]})
	}
	plan := suggested
	if r.Method == http.MethodPost {
		var in removalPlan
		if decodeJSON(w, r, &in) != nil || in.Plan == nil {
			fail(400, "invalid_request", "Send a JSON body with plan: {dependent id: Connection id or \"\"}")
			return true
		}
		plan = *in.Plan
	}
	answer := map[string]any{"connection": id, "dependents": deps}
	apps, err := h.removalAffects(current, id, plan)
	if err != nil {
		answer["applications"], answer["problem"] = []affectedApplication{}, err.Error()
	} else {
		answer["applications"] = apps
	}
	writeJSON(w, 200, answer)
	return true
}

// removeTarget removes Connection id. With a plan, everything using it is
// re-pointed (or left without one) and the Connection removed as one Change;
// without one, the removal is refused while anything uses it.
func (h *Handler) removeTarget(w http.ResponseWriter, r *http.Request, current config.Config, id string, fail func(int, string, string)) {
	var in removalPlan
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
	if err != nil {
		fail(413, "request_too_large", "A removal body must be at most 64 KiB")
		return
	}
	if strings.TrimSpace(string(body)) != "" {
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		if decodeJSON(w, r, &in) != nil {
			fail(400, "invalid_request", "Send no body, or a JSON body with plan: {dependent id: Connection id or \"\"}")
			return
		}
	}
	what := "remove connection " + id
	if in.Plan == nil {
		edit, err := config.RemoveTarget(current, id)
		if err == nil {
			_, _, err = h.live.Change(what, edit)
		}
		if err != nil {
			fail(400, "invalid_connection", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
		return
	}
	edit, err := config.RemoveTargetWith(current, id, *in.Plan)
	if err != nil {
		fail(400, "invalid_removal_plan", err.Error())
		return
	}
	_, ch, err := h.live.Change(what, edit)
	if err != nil {
		// Validation messages name fields and ids only, never secret values.
		fail(400, "invalid_removal_plan", err.Error())
		return
	}
	// The Applications affected are worked out from the configuration the
	// Change started from, not the snapshot the plan was checked against.
	answer := map[string]any{"ok": true, "change_id": ch.ID, "applications": []affectedApplication{}}
	if before, after, err := ch.Configs(); err == nil {
		answer["applications"] = affectedBy(before, after)
	}
	writeJSON(w, 200, answer)
}

// removalAffects is the Applications whose Catalog removing id with plan
// would change, worked out without changing anything.
func (h *Handler) removalAffects(current config.Config, id string, plan map[string]string) ([]affectedApplication, error) {
	edit, err := config.RemoveTargetWith(current, id, plan)
	if err != nil {
		return nil, err
	}
	before, after, err := h.live.Try(edit)
	if err != nil {
		return nil, err
	}
	return affectedBy(before, after), nil
}

func affectedBy(before, after config.Config) []affectedApplication {
	out := []affectedApplication{}
	for app, ch := range catalog.Diff(before, after) {
		name := config.BuiltinApplication // aiport's own model calls
		if o, ok := after.Application(app); ok {
			name = o.Name
		}
		out = append(out, affectedApplication{ch, name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Application < out[j].Application })
	return out
}

// suggestReplacements reads the model lists of the enabled Connections
// (cached lists are reused; one that cannot be read lists nothing) and
// suggests a replacement for each item using id.
func (h *Handler) suggestReplacements(ctx context.Context, c config.Config, id string) map[string]string {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	models := catalog.Models{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, t := range c.Targets {
		if t.Disabled {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l := h.gateway.Models(ctx, t.ID, false); l.Listed {
				mu.Lock()
				models[t.ID] = l.Models
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return catalog.Suggest(c, id, models)
}
