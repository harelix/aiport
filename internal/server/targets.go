package server

import (
	"net/http"
	"strings"

	"github.com/harelix/aiport/internal/config"
)

type targetInput struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
	Protocol  string `json:"protocol"`
	BaseURL   string `json:"base_url"`
	APIKeyEnv string `json:"api_key_env"`
}

func (in *targetInput) trim() {
	for _, f := range []*string{&in.ID, &in.Provider, &in.Region, &in.Endpoint, &in.Protocol, &in.BaseURL, &in.APIKeyEnv} {
		*f = strings.TrimSpace(*f)
	}
}

// targetRoute handles connection management for the dashboard:
//
//	POST   /api/v1/targets        add a connection
//	PATCH  /api/v1/targets/{id}   change region, endpoint, protocol
//	DELETE /api/v1/targets/{id}   remove a connection (with a plan: re-point what uses it first, as one Change)
//	PUT    /api/v1/defaults       choose the default connection per protocol
//
// Each edits the YAML file in place (other keys and comments kept), validates
// the whole file, writes it atomically and applies it to new requests. The
// caller has already verified management access.
func (h *Handler) targetRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	var id string
	switch {
	case r.URL.Path == "/api/v1/targets", r.URL.Path == "/api/v1/defaults":
	default:
		rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/targets/")
		if !ok || rest == "" || strings.Contains(rest, "/") {
			return false
		}
		id = rest
	}
	allowed := map[string]string{"/api/v1/targets": "POST", "/api/v1/defaults": "PUT"}[r.URL.Path]
	if id != "" {
		allowed = "PATCH, DELETE"
	}
	if !strings.Contains(allowed, r.Method) || r.Method == "" {
		w.Header().Set("Allow", allowed)
		fail(405, "method_not_allowed", "Use "+allowed)
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change connections")
		return true
	}
	if !h.live.Writable() {
		fail(503, "config_readonly", "This server cannot edit its configuration")
		return true
	}
	current := h.snapshot()
	if _, exists := current.Connection(id); id != "" && !exists {
		fail(404, "target_not_found", "No connection with this id")
		return true
	}
	var (
		edit config.Edit
		what string
		err  error
	)
	switch {
	case r.URL.Path == "/api/v1/defaults":
		var d config.Defaults
		if decodeJSON(w, r, &d) != nil {
			fail(400, "invalid_request", "Send a JSON body with openai_target and anthropic_target")
			return true
		}
		what, edit = "protocol defaults", config.SetDefaults(config.Defaults{OpenAITarget: strings.TrimSpace(d.OpenAITarget), AnthropicTarget: strings.TrimSpace(d.AnthropicTarget)})
	case r.Method == http.MethodPost:
		var in targetInput
		if decodeJSON(w, r, &in) != nil {
			fail(400, "invalid_request", "Send a JSON body with id and provider")
			return true
		}
		in.trim()
		for _, t := range current.Targets {
			if t.ID == in.ID {
				fail(409, "target_exists", "A connection with this id already exists")
				return true
			}
		}
		what, edit = "add connection "+in.ID, config.AddTarget(config.Target{ID: in.ID, Provider: in.Provider, Region: in.Region, Endpoint: in.Endpoint, Protocol: in.Protocol, BaseURL: in.BaseURL, APIKeyEnv: in.APIKeyEnv})
	case r.Method == http.MethodPatch:
		var in config.TargetSettings
		if decodeJSON(w, r, &in) != nil {
			fail(400, "invalid_request", "Send a JSON body with region, endpoint and protocol")
			return true
		}
		what, edit = "edit connection "+id, config.EditTarget(id, config.TargetSettings{Region: strings.TrimSpace(in.Region), Endpoint: strings.TrimSpace(in.Endpoint), Protocol: strings.TrimSpace(in.Protocol)})
	default:
		h.removeTarget(w, r, current, id, fail)
		return true
	}
	if err == nil {
		_, _, err = h.live.Change(what, edit)
	}
	if err != nil {
		// Validation messages name fields and ids only, never secret values.
		fail(400, "invalid_connection", err.Error())
		return true
	}
	writeJSON(w, 200, map[string]any{"ok": true})
	return true
}
