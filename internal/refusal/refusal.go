// Package refusal is the catalog of requests aiport declines to serve on the
// inference path. Each refusal carries everything about it in one place: the
// HTTP status, the stable error code clients match on, the message, how the
// request is recorded in Activity (its outcome), and an optional hint telling
// the client what would fix it.
package refusal

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
}

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

var (
	// Routing: no usable connection for this client and protocol.
	TargetUnconfigured   = rejected(503, "target_unconfigured", "Configure a target for this protocol or origin")
	TargetDisabled       = rejected(503, "target_disabled", "The selected target is disabled")
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
