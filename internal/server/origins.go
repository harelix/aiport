package server

import (
	"net/http"
	"strings"

	"aiport/internal/agents"
	"aiport/internal/config"

	"go.yaml.in/yaml/v3"
)

// originRoute changes one Application's settings for the dashboard:
//
//	PATCH /api/v1/origins/{id}   {may_request_changes?, agent?, repository?}
//
// agent is its Coding agent (claude, codex, gemini; "" = the default one)
// and repository the absolute path of its git repository ("" = none); a
// field left out stays as it is. The change goes through the
// configuration-change module, so it is live at once and recorded as one
// revertable Change, whose id is returned. The caller has already verified
// management access.
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
	app, found := h.snapshot().Application(id)
	if !found {
		fail(404, "origin_not_found", "No Application with this id")
		return true
	}
	var in struct {
		MayRequestChanges *bool   `json:"may_request_changes"`
		Agent             *string `json:"agent"`
		Repository        *string `json:"repository"`
	}
	if decodeJSON(w, r, &in) != nil || in.MayRequestChanges == nil && in.Agent == nil && in.Repository == nil {
		fail(400, "invalid_request", "Send a JSON body with may_request_changes (true or false), agent or repository")
		return true
	}
	var edits []config.Edit
	var what []string
	if in.MayRequestChanges != nil {
		edits = append(edits, config.SetMayRequestChanges(id, *in.MayRequestChanges))
		if *in.MayRequestChanges {
			what = append(what, "allow "+id+" to request changes")
		} else {
			what = append(what, "forbid "+id+" to request changes")
		}
	}
	if in.Agent != nil || in.Repository != nil {
		agent, repository := app.Agent, app.Repository
		if in.Agent != nil {
			agent = strings.TrimSpace(*in.Agent)
		}
		if in.Repository != nil {
			repository = strings.TrimSpace(*in.Repository)
		}
		agent, repository, code, msg := h.codingAgentSettings(agent, repository)
		if code != "" {
			fail(400, code, msg)
			return true
		}
		edits = append(edits, config.SetCodingAgent(id, agent, repository))
		what = append(what, codingAgentWhat(id, agent, repository))
	}
	_, change, err := h.live.Change(strings.Join(what, "; "), func(root *yaml.Node) error {
		for _, e := range edits {
			if err := e(root); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fail(400, "invalid_client", err.Error())
		return true
	}
	writeJSON(w, 200, map[string]any{"ok": true, "change_id": change.ID})
	return true
}

// codingAgentSettings checks an Application's Coding agent and repository
// path as the Owner gave them, and returns them as they are stored (the
// repository with symbolic links resolved), or an error code and message.
func (h *Handler) codingAgentSettings(agent, repository string) (string, string, string, string) {
	if code, msg := h.checkCodingAgent(agent); code != "" {
		return "", "", code, msg
	}
	if repository != "" {
		real, err := agents.Repository(repository)
		if err != nil {
			return "", "", "invalid_repository", err.Error()
		}
		repository = real
	}
	return agent, repository, "", ""
}

// codingAgentWhat describes a coding agent setting for the Changes list.
func codingAgentWhat(id, agent, repository string) string {
	name := "the default coding agent"
	if agent != "" {
		name = "coding agent " + agent
	}
	if repository == "" {
		return id + ": " + name + ", no repository"
	}
	return id + ": " + name + " in " + repository
}
