package server

import (
	"net/http"

	"aiport/internal/capability"
)

// CapabilitiesPath answers what aiport can do right now for the calling
// Application.
const CapabilitiesPath = "/v1/aiport/capabilities"

// capabilities answers for the Application with id origin, from the live
// configuration and routing, so an applied Change shows at once. ok is false
// when origin is not a configured Application (the management token, or no
// token while authentication is off). The MCP tool aiport_capabilities must
// answer through this too.
func (h *Handler) capabilities(origin string) (capability.Capabilities, bool) {
	return capability.For(h.snapshot(), origin, h.gateway)
}

// serveCapabilities handles GET /v1/aiport/capabilities for an authenticated
// caller; it only ever answers about the Application whose token was sent.
func (h *Handler) serveCapabilities(w http.ResponseWriter, r *http.Request, originID string, fail func(int, string, string)) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	caps, ok := h.capabilities(originID)
	if !ok {
		fail(403, "application_token_required", "Capabilities describe one Application: send that Application's own token")
		return
	}
	writeJSON(w, 200, caps)
}
