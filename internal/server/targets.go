package server

import (
	"net/http"
	"os"
	"strings"

	"aiport/internal/config"
)

// snapshot returns a copy of the live configuration safe to read without locks.
func (h *Handler) snapshot() config.Config {
	h.cfgMu.RLock()
	defer h.cfgMu.RUnlock()
	c := h.config
	c.Targets = make([]config.Target, len(h.config.Targets))
	copy(c.Targets, h.config.Targets)
	c.Origins = append(make([]config.Origin, 0, len(h.config.Origins)), h.config.Origins...)
	c.Aliases = append(make([]config.Alias, 0, len(h.config.Aliases)), h.config.Aliases...)
	return c
}

// applyRouting makes a freshly saved configuration's connections, defaults,
// origins and aliases live, then re-resolves API keys. Listener, auth, proxy
// limits and telemetry still need a restart. Caller holds cfgMu.
func (h *Handler) applyRouting(fresh config.Config) {
	h.config.Targets, h.config.Origins, h.config.Aliases, h.config.Defaults, h.config.Evaluation = fresh.Targets, fresh.Origins, fresh.Aliases, fresh.Defaults, fresh.Evaluation
	h.gateway.Apply(h.config)
	h.refreshKeys()
}

// refreshKeys resolves each connection's key the same way startup does: a key
// saved in the private credential file wins, then the server's environment.
func (h *Handler) refreshKeys() {
	saved := map[string]string{}
	if h.credentialPath != "" {
		if s, err := config.ReadCredentialFile(h.credentialPath); err == nil {
			saved = s
		}
	}
	h.keyMu.Lock()
	defer h.keyMu.Unlock()
	h.keySource, h.secretStatus = map[string]string{}, map[string]bool{}
	for _, t := range h.config.Targets {
		value, source := "", keyMissing
		switch env := os.Getenv(t.APIKeyEnv); {
		case t.APIKeyEnv == "":
			source = keyNotNeeded
		case saved[t.APIKeyEnv] != "":
			value, source = saved[t.APIKeyEnv], keySaved
		case config.ValidCredentialValue(env) && !h.isLocalCredential(env):
			value, source = env, keyEnvironment
		}
		h.gateway.SetKey(t.ID, value)
		h.keySource[t.ID] = source
		h.secretStatus[t.ID] = source != keyMissing
	}
}

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
//	DELETE /api/v1/targets/{id}   remove a connection
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
	if h.configPath == "" {
		fail(503, "config_readonly", "This server cannot edit its configuration")
		return true
	}
	h.cfgMu.Lock()
	defer h.cfgMu.Unlock()
	exists := false
	for _, t := range h.config.Targets {
		exists = exists || t.ID == id
	}
	if id != "" && !exists {
		fail(404, "target_not_found", "No connection with this id")
		return true
	}
	var (
		fresh config.Config
		err   error
	)
	switch {
	case r.URL.Path == "/api/v1/defaults":
		var d config.Defaults
		if decodeJSON(w, r, &d) != nil {
			fail(400, "invalid_request", "Send a JSON body with openai_target and anthropic_target")
			return true
		}
		fresh, err = config.SaveDefaults(h.configPath, config.Defaults{OpenAITarget: strings.TrimSpace(d.OpenAITarget), AnthropicTarget: strings.TrimSpace(d.AnthropicTarget)})
	case r.Method == http.MethodPost:
		var in targetInput
		if decodeJSON(w, r, &in) != nil {
			fail(400, "invalid_request", "Send a JSON body with id and provider")
			return true
		}
		in.trim()
		for _, t := range h.config.Targets {
			if t.ID == in.ID {
				fail(409, "target_exists", "A connection with this id already exists")
				return true
			}
		}
		fresh, err = config.AddTarget(h.configPath, config.Target{ID: in.ID, Provider: in.Provider, Region: in.Region, Endpoint: in.Endpoint, Protocol: in.Protocol, BaseURL: in.BaseURL, APIKeyEnv: in.APIKeyEnv})
	case r.Method == http.MethodPatch:
		var in config.TargetSettings
		if decodeJSON(w, r, &in) != nil {
			fail(400, "invalid_request", "Send a JSON body with region, endpoint and protocol")
			return true
		}
		fresh, err = config.SaveTarget(h.configPath, id, config.TargetSettings{Region: strings.TrimSpace(in.Region), Endpoint: strings.TrimSpace(in.Endpoint), Protocol: strings.TrimSpace(in.Protocol)})
	default:
		fresh, err = config.RemoveTarget(h.configPath, id)
	}
	if err != nil {
		// Validation messages name fields and ids only, never secret values.
		fail(400, "invalid_connection", err.Error())
		return true
	}
	h.applyRouting(fresh)
	writeJSON(w, 200, map[string]any{"ok": true})
	return true
}
