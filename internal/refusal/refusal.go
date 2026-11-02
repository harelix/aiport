// Package refusal is the catalog of requests aiport declines to serve on the
// inference path. Each refusal carries everything about it in one place: the
// HTTP status, the stable error code clients match on, the message, how the
// request is recorded in Activity (its outcome), and an optional hint telling
// the client what would fix it.
package refusal

import "strings"

// Outcomes recorded in request history for a refused request.
const (
	Rejected     = "rejected"
	Timeout      = "timeout"
	GatewayError = "gateway_error"
	Cancelled    = "cancelled"
)

// Refusal is one way aiport declines a request.
type Refusal struct {
	Status  int
	Code    string
	Message string
	Outcome string
	Hint    string
	// Changes names the kinds of Change request that could remove this
	// refusal. Only Applications allowed to request changes are told (see
	// ForRequester); the status and code never depend on it.
	Changes []string
}

// ForRequester returns the refusal as an Application allowed to request
// changes sees it: with a hint naming the request kinds and the endpoint,
// when a Change could fix it.
func (r Refusal) ForRequester() Refusal {
	if len(r.Changes) > 0 {
		r.Hint = ChangeHint(r.Changes...)
	}
	return r
}

// RequestsPath and RequestTool are where an Application asks for a Change.
const (
	RequestsPath = "/v1/aiport/requests"
	RequestTool  = "aiport_request_change"
)

// ChangeHint tells an Application how to ask the Owner for a Change of one
// of the given kinds.
func ChangeHint(kinds ...string) string {
	return "A Change could fix this: ask the owner with POST " + RequestsPath + " (kind " + strings.Join(kinds, " or ") + ") or the MCP tool " + RequestTool + "; see /_connect."
}

// HintHeader carries a hint on a provider error, whose body aiport passes
// through unchanged.
const HintHeader = "X-Aiport-Hint"

// UnknownModelHint goes with an upstream 404 for a requested model: an Alias
// for this Application could point the name at a model the provider has.
var UnknownModelHint = ChangeHint("application_alias")

// WithMessage returns the refusal with a more specific message.
func (r Refusal) WithMessage(m string) Refusal {
	r.Message = m
	return r
}

// Body is the JSON error object clients receive.
func (r Refusal) Body(requestID string) map[string]any {
	e := map[string]string{"code": r.Code, "message": r.Message, "request_id": requestID}
	if r.Hint != "" {
		e["hint"] = r.Hint
	}
	return map[string]any{"error": e}
}

func rejected(status int, code, message string) Refusal {
	return Refusal{Status: status, Code: code, Message: message, Outcome: Rejected}
}

func fixable(r Refusal, kinds ...string) Refusal {
	r.Changes = kinds
	return r
}

var (
	// Routing: no usable connection for this client and protocol.
	TargetUnconfigured   = fixable(rejected(503, "target_unconfigured", "Configure a target for this protocol or origin"), "bind_application", "application_alias")
	TargetDisabled       = fixable(rejected(503, "target_disabled", "The selected target is disabled"), "bind_application", "application_alias")
	TargetInvalid        = rejected(503, "target_invalid", "Selected target configuration is invalid")
	IncompatibleProtocol = rejected(400, "incompatible_protocol", "The selected target does not accept this native protocol")
	UpstreamKeyMissing   = rejected(503, "upstream_key_missing", "No API key is set for this connection; add it in the dashboard under Settings")

	// Admission and request shape.
	ConcurrencyLimit   = rejected(429, "concurrency_limit", "Gateway concurrency limit reached")
	RequestTooLarge    = rejected(413, "request_too_large", "Request exceeds the configured admission limit")
	RequestReadFailed  = rejected(400, "request_read_failed", "Could not read request body")
	UnknownCompression = rejected(400, "unknown_compression", "Unknown compression method")
	AmbiguousModel     = rejected(400, "ambiguous_model", "Alias routing requires an unambiguous JSON object")

	// Upstream.
	UpstreamRequestFailed = Refusal{Status: 502, Code: "upstream_request_failed", Message: "Could not create upstream request", Outcome: GatewayError}
	UpstreamUnavailable   = Refusal{Status: 502, Code: "upstream_unavailable", Message: "Upstream connection failed", Outcome: GatewayError}
	UpstreamTimeout       = Refusal{Status: 504, Code: "upstream_timeout", Message: "Upstream response deadline exceeded", Outcome: Timeout}
	ClientCancelled       = Refusal{Status: 499, Code: "client_cancelled", Message: "Client cancelled request", Outcome: Cancelled}
)
