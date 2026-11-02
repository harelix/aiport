package server

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"aiport/internal/decide"
	"aiport/internal/extension"
)

// RequestsPath is where an Application files and reads its own Change
// requests:
//
//	POST /v1/aiport/requests       file one {kind, need, details, why, scope, scope_case, filed_by?}
//	GET  /v1/aiport/requests       list this Application's requests, newest first
//	GET  /v1/aiport/requests/{id}  one of them, with the Owner's decision
//
// The MCP tools aiport_request_change and aiport_my_requests answer through
// the same decide.Queue calls (File, Requests, RequestOf) and the same
// refusal codes (requestRefusal).
const RequestsPath = "/v1/aiport/requests"

// serveRequests handles RequestsPath for an authenticated caller; it only
// ever files for, and answers about, the Application whose token was sent.
func (h *Handler) serveRequests(w http.ResponseWriter, r *http.Request, originID string, fail func(int, string, string)) {
	id, one := strings.CutPrefix(r.URL.Path, RequestsPath+"/")
	if one && (id == "" || strings.Contains(id, "/")) {
		fail(404, "change_request_not_found", "No change request with this id for this Application")
		return
	}
	if _, ok := h.capabilities(originID); !ok {
		fail(403, "application_token_required", decide.ErrNotApplication.Error())
		return
	}
	if h.decisions == nil {
		fail(409, "adapting_off", decide.ErrOff.Error())
		return
	}
	switch {
	case one && r.Method == http.MethodGet:
		a, err := h.decisions.RequestOf(originID, id)
		if err != nil {
			fail(404, "change_request_not_found", "No change request with this id for this Application")
			return
		}
		writeJSON(w, 200, a)
	case !one && r.Method == http.MethodGet:
		writeJSON(w, 200, map[string]any{"requests": h.decisions.Requests(originID)})
	case !one && r.Method == http.MethodPost:
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, decide.MaxBody))
		if err != nil {
			fail(413, "request_too_large", "A change request body must be at most 64 KiB")
			return
		}
		a, err := h.decisions.File(originID, body, h.clock())
		if err != nil {
			h.requestRefusal(w, err, fail)
			return
		}
		writeJSON(w, 201, a)
	default:
		allow := "GET"
		if !one {
			allow = "GET, POST"
		}
		w.Header().Set("Allow", allow)
		fail(405, "method_not_allowed", "Use "+allow)
	}
}

// requestRefusal answers a refused Change request. Both doors (HTTP and MCP)
// use these codes.
func (h *Handler) requestRefusal(w http.ResponseWriter, err error, fail func(int, string, string)) {
	var incomplete decide.Incomplete
	var limited decide.RateLimited
	var denied extension.Denied
	switch {
	case errors.Is(err, decide.ErrOff):
		fail(409, "adapting_off", err.Error())
	case errors.Is(err, decide.ErrNotApplication):
		fail(403, "application_token_required", err.Error())
	case errors.Is(err, decide.ErrNotPermitted):
		fail(403, "change_requests_not_allowed", err.Error())
	case errors.As(err, &limited):
		w.Header().Set("Retry-After", strconv.Itoa(int(limited.RetryAfter.Seconds())+1))
		fail(429, "change_requests_rate_limited", err.Error())
	case errors.As(err, &denied):
		writeJSON(w, 403, map[string]any{"error": map[string]any{"code": "change_request_refused", "message": err.Error(), "request_id": w.Header().Get("X-Aiport-Request-Id"), "extension": denied.Extension, "reason": denied.Reason}})
	case errors.As(err, &incomplete):
		writeJSON(w, 400, map[string]any{"error": map[string]any{"code": "invalid_change_request", "message": err.Error(), "request_id": w.Header().Get("X-Aiport-Request-Id"), "problems": incomplete.Problems}})
	default:
		fail(500, "change_request_failed", "The change request could not be filed")
	}
}
