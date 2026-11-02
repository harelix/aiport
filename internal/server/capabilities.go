package server

import (
	"net/http"

	"aiport/internal/capability"
)

// CapabilitiesPath answers what aiport can do right now for the calling
// Application.
const CapabilitiesPath = "/v1/aiport/capabilities"

// capabilities answers for the Application with id applicationID, from the
// live configuration and routing, so an applied Change shows at once. ok is
// false when it is not a configured Application (the management token, or
// no token while authentication is off). The MCP tool aiport_capabilities
// answers through this too.
func (h *Handler) capabilities(applicationID string) (capability.Capabilities, bool) {
	return capability.For(h.snapshot(), applicationID, h.gateway)
}

// isApplication reports whether id is a configured Application, as opposed
// to the management token or no token while authentication is off.
func (h *Handler) isApplication(id string) bool {
	_, ok := h.snapshot().Application(id)
	return ok
}

// serveCapabilities handles GET /v1/aiport/capabilities for an authenticated
// caller; it only ever answers about the Application whose token was sent.
func (h *Handler) serveCapabilities(w http.ResponseWriter, r *http.Request, applicationID string, fail func(int, string, string)) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	caps, ok := h.capabilities(applicationID)
	if !ok {
		fail(403, "application_token_required", "Capabilities describe one Application: send that Application's own token")
		return
	}
	writeJSON(w, 200, caps)
}
