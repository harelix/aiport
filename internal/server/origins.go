package server

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"aiport/internal/agents"
	"aiport/internal/config"

	"go.yaml.in/yaml/v3"
)

// originRoute manages Applications for the dashboard:
//
//	POST   /api/v1/origins             {name, id?, target?, may_request_changes?, agent?, repository?}  add one with a token, without pairing
//	PATCH  /api/v1/origins/{id}        {may_request_changes?, agent?, repository?}
//	DELETE /api/v1/origins/{id}        remove it and delete its token
//	POST   /api/v1/origins/{id}/token  issue a new token; the old one stops working
//
// A token is in the response once and never again; aiport keeps it only in
// the private credential file.
//
// agent is its Coding agent (claude, codex, gemini; "" = the default one)
// and repository the absolute path of its git repository ("" = none); a
// field left out stays as it is. The change goes through the
// configuration-change module, so it is live at once and recorded as one
// revertable Change, whose id is returned. The caller has already verified
// management access.
func (h *Handler) originRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path == "/api/v1/origins" {
		h.addOrigin(w, r, fail)
		return true
	}
	id, ok := strings.CutPrefix(r.URL.Path, "/api/v1/origins/")
	if !ok || id == "" {
		return false
	}
	if app, ok := strings.CutSuffix(id, "/token"); ok && app != "" && !strings.Contains(app, "/") {
		h.issueToken(w, r, app, fail)
		return true
	}
	if strings.Contains(id, "/") {
		return false
	}
	if r.Method == http.MethodDelete {
		h.removeOrigin(w, r, id, fail)
		return true
	}
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", "PATCH, DELETE")
		fail(405, "method_not_allowed", "Use PATCH or DELETE")
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

// tokenAccess checks what every token-writing Application route needs: a
// same-origin dashboard request and a configuration and credential file
// aiport may write.
func (h *Handler) tokenAccess(w http.ResponseWriter, r *http.Request, method string, fail func(int, string, string)) bool {
	if r.Method != method {
		w.Header().Set("Allow", method)
		fail(405, "method_not_allowed", "Use "+method)
		return false
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change Applications")
		return false
	}
	if !h.live.Writable() || h.keys.File == "" {
		fail(503, "config_readonly", "This server cannot edit its configuration or credentials")
		return false
	}
	return true
}

// addOrigin adds an Application the Owner names, with its token, without
// pairing: for an application whose coding agent is not running, or when the
// Owner would rather paste the token into its .aiport file.
func (h *Handler) addOrigin(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	if !h.tokenAccess(w, r, http.MethodPost, fail) {
		return
	}
	var in struct {
		Name              string `json:"name"`
		ID                string `json:"id"`
		Target            string `json:"target"`
		MayRequestChanges bool   `json:"may_request_changes"`
		Agent             string `json:"agent"`
		Repository        string `json:"repository"`
	}
	if decodeJSON(w, r, &in) != nil {
		fail(400, "invalid_request", "Send JSON with name and optional id, target, may_request_changes, agent and repository")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 128 || strings.ContainsAny(in.Name, "\r\n\x00") {
		fail(400, "invalid_request", "name must be 1–128 characters on one line")
		return
	}
	agent, repository, code, msg := h.codingAgentSettings(strings.TrimSpace(in.Agent), strings.TrimSpace(in.Repository))
	if code != "" {
		fail(400, code, msg)
		return
	}
	id, token, status, code, msg := h.createOrigin("add application", config.Origin{ID: strings.TrimSpace(in.ID), Name: in.Name, Target: strings.TrimSpace(in.Target), MayRequestChanges: in.MayRequestChanges, Agent: agent, Repository: repository})
	if code != "" {
		fail(status, code, msg)
		return
	}
	writeJSON(w, 201, map[string]any{"client_id": id, "token": token, "note": "Save it as AIPORT_APP_TOKEN in .aiport at the application's repository root (in .gitignore, never committed). It is shown only once."})
}

// issueToken gives an Application a new token under its existing reference;
// the old token stops working at once. Not a configuration Change: nothing in
// aiport.yaml changes, and a token is never kept for a revert.
func (h *Handler) issueToken(w http.ResponseWriter, r *http.Request, id string, fail func(int, string, string)) {
	if !h.tokenAccess(w, r, http.MethodPost, fail) {
		return
	}
	app, found := h.snapshot().Application(id)
	if !found {
		fail(404, "origin_not_found", "No Application with this id")
		return
	}
	token := randomString(32)
	if err := h.keys.SetLocal(app.TokenEnv, token); err != nil {
		if errors.Is(err, config.ErrTokenFromEnvironment) {
			fail(409, "token_from_environment", app.TokenEnv+": "+err.Error())
			return
		}
		fail(500, "token_save_failed", "Could not save the new token")
		return
	}
	h.resolveTokens(h.snapshot(), true)
	writeJSON(w, 200, map[string]any{"client_id": id, "token": token, "note": "Save it as AIPORT_APP_TOKEN in .aiport at the application's repository root (in .gitignore, never committed). It is shown only once; the previous token no longer works."})
}

// removeOrigin takes an Application out of the configuration (one Change,
// with its own Aliases, binding and Coding agent settings) and deletes its
// token. Reverting the Change brings the Application back without a token:
// it is listed as needing one until the Owner issues it.
func (h *Handler) removeOrigin(w http.ResponseWriter, r *http.Request, id string, fail func(int, string, string)) {
	if !h.tokenAccess(w, r, http.MethodDelete, fail) {
		return
	}
	app, found := h.snapshot().Application(id)
	if !found {
		fail(404, "origin_not_found", "No Application with this id")
		return
	}
	// Refuse before changing anything when the token cannot be deleted.
	if current := h.keys.Env; current != nil && current(app.TokenEnv) != "" {
		if saved, _ := config.ReadCredentialFile(h.keys.File); saved[app.TokenEnv] != current(app.TokenEnv) {
			fail(409, "token_from_environment", app.TokenEnv+": "+config.ErrTokenFromEnvironment.Error())
			return
		}
	}
	_, change, err := h.live.Change("remove application "+id, config.RemoveOrigin(id))
	if err != nil {
		fail(400, "invalid_client", err.Error())
		return
	}
	if err := h.keys.SetLocal(app.TokenEnv, ""); err != nil {
		log.Printf("aiport: removed Application %s but could not delete its token %s: %v", id, app.TokenEnv, err)
	}
	h.resolveTokens(h.snapshot(), true)
	writeJSON(w, 200, map[string]any{"removed": id, "change_id": change.ID})
}
