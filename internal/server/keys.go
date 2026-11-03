package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/harelix/aiport/internal/config"
)

// Key sources reported to the dashboard. Key values are never returned.
const (
	keyNotNeeded   = config.KeyNotNeeded
	keySaved       = config.KeySaved
	keyEnvironment = config.KeyEnvironment
	keyMissing     = config.KeyMissing
)

// initKeySources reports key sources for a bare handler built from resolved
// startup secrets (tests and embedding), where no credential file is known.
func (h *Handler) initKeySources(c config.Config, s config.Secrets) {
	h.keyMu.Lock()
	defer h.keyMu.Unlock()
	h.keySource = map[string]string{}
	for _, t := range c.Targets {
		switch {
		case t.APIKeyEnv == "":
			h.keySource[t.ID] = keyNotNeeded
		case s.Upstream[t.ID] != "":
			h.keySource[t.ID] = keyEnvironment
		default:
			h.keySource[t.ID] = keyMissing
		}
		h.secretStatus[t.ID] = h.keySource[t.ID] != keyMissing
	}
}

// refreshKeys gives every connection the key config.Keys resolves for it and
// records where it came from. A provider key equal to a local token is
// refused, so an aiport token never leaves for a provider.
func (h *Handler) refreshKeys(c config.Config) {
	h.keyMu.Lock()
	defer h.keyMu.Unlock()
	h.keySource, h.secretStatus = map[string]string{}, map[string]bool{}
	for _, t := range c.Targets {
		value, source := h.keys.Provider(t)
		if value != "" && h.isLocalCredential(value) {
			value, source = "", keyMissing
		}
		h.gateway.SetKey(t.ID, value)
		h.keySource[t.ID] = source
		h.secretStatus[t.ID] = source != keyMissing
	}
}

func (h *Handler) keyStatus() (map[string]bool, map[string]string) {
	h.keyMu.Lock()
	defer h.keyMu.Unlock()
	status, source := map[string]bool{}, map[string]string{}
	for k, v := range h.secretStatus {
		status[k] = v
	}
	for k, v := range h.keySource {
		source[k] = v
	}
	return status, source
}

// isLocalCredential prevents a provider key from equalling a local token.
func (h *Handler) isLocalCredential(value string) bool {
	digest := sha256.Sum256([]byte(value))
	match := 0
	for _, c := range h.creds() {
		match |= subtle.ConstantTimeCompare(digest[:], c.Digest[:])
	}
	return match == 1
}

// targetKeyRoute handles PUT/DELETE /api/v1/targets/{id}/api-key.
// The caller has already verified management access.
func (h *Handler) targetKeyRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/targets/")
	if !ok {
		return false
	}
	id, ok := strings.CutSuffix(rest, "/api-key")
	if !ok || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "PUT, DELETE")
		fail(405, "method_not_allowed", "Use PUT or DELETE")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change API keys")
		return true
	}
	if h.keys.File == "" {
		fail(503, "keys_unavailable", "API keys cannot be saved by this server")
		return true
	}
	live := h.snapshot()
	var target config.Target
	for _, t := range live.Targets {
		if t.ID == id {
			target = t
		}
	}
	if target.ID == "" {
		fail(404, "target_not_found", "No connection with this id")
		return true
	}
	if target.APIKeyEnv == "" {
		fail(400, "key_not_used", "This connection does not use an API key")
		return true
	}
	value, source := "", keyMissing
	if r.Method == http.MethodPut {
		var input struct {
			APIKey string `json:"api_key"`
		}
		if err := decodeAuth(w, r, &input); err != nil {
			fail(400, "invalid_request", "Send a JSON body with api_key")
			return true
		}
		value = strings.TrimSpace(input.APIKey)
		if !config.ValidCredentialValue(value) {
			fail(400, "invalid_api_key", "API key must be printable text without spaces or quotes")
			return true
		}
		if h.isLocalCredential(value) {
			fail(400, "invalid_api_key", "That is an aiport token, not a provider API key")
			return true
		}
		source = keySaved
	}
	if err := config.SetCredential(h.keys.File, target.APIKeyEnv, value); err != nil {
		fail(500, "key_save_failed", "Could not update the private credential file")
		return true
	}
	if r.Method == http.MethodDelete {
		// Fall back to whatever else provides a key: the server's environment.
		value, source = h.keys.Provider(target)
		if value != "" && h.isLocalCredential(value) {
			value, source = "", keyMissing
		}
	}
	h.keyMu.Lock()
	// Targets sharing the same variable share the key.
	for _, t := range live.Targets {
		if t.APIKeyEnv == target.APIKeyEnv {
			h.gateway.SetKey(t.ID, value)
			h.keySource[t.ID] = source
			h.secretStatus[t.ID] = source != keyMissing
		}
	}
	h.keyMu.Unlock()
	status, sources := h.keyStatus()
	writeJSON(w, 200, map[string]any{"secret_status": status, "secret_source": sources})
	return true
}
