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
	// Owner is what the dashboard tells the Owner about this refusal: why
	// it happened and what fixes it (see Explain). Never sent to clients.
	Owner Advice
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
	return "A Change could fix this: ask the Owner with POST " + RequestsPath + " (kind " + strings.Join(kinds, " or ") + ") or the MCP tool " + RequestTool + "; see /_connect."
}

// HintHeader carries a hint on a provider error, whose body aiport passes
// through unchanged.
const HintHeader = "X-Aiport-Hint"

// UnknownModelHint goes with an upstream 404 for a requested model: an Alias
// for this Application could point the name at a model the provider has.
var UnknownModelHint = ChangeHint("application_alias")

// IssuesPath and IssuesTool are where an Application's coding agent reads
// its Issues.
const (
	IssuesPath = "/v1/aiport/issues"
	IssuesTool = "aiport_my_issues"
)

// IssueHint goes with a failing call while Issue id is open for the
// Application: it names the Issue and where its fix is.
func IssueHint(id string) string {
	return "aiport Issue " + id + " is open for this Application: read what is wrong and the fix with GET " + IssuesPath + "/" + id + " or the MCP tool " + IssuesTool + "; see /_connect."
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

func fixable(r Refusal, kinds ...string) Refusal {
	r.Changes = kinds
	return r
}

func advised(r Refusal, a Advice) Refusal {
	a.Code = r.Code
	r.Owner = a
	return r
}

// The catalog. Every entry carries the Owner's advice next to what the client
// is told; {protocol} in advice is the request's protocol (OpenAI, Anthropic).
var (
	// Routing: no usable connection for this client and protocol.
	TargetUnconfigured = advised(fixable(rejected(503, "target_unconfigured", "Configure a target for this protocol or origin"), "bind_application", "application_alias"), Advice{
		Summary:     "Refused: no Connection for the {protocol} protocol",
		Explanation: "No Connection is the {protocol} protocol default and the Application is not bound to a Connection, so aiport had nowhere to send the request.",
		Fix:         "Pick a Connection for the {protocol} protocol under Routing → Protocol defaults, or bind the Application to a Connection on Integrate.",
		Page:        "Routing", Section: "Protocol defaults", Config: true})
	TargetDisabled = advised(fixable(rejected(503, "target_disabled", "The selected target is disabled"), "bind_application", "application_alias"), Advice{
		Summary:     "Refused: the Connection is disabled",
		Explanation: "The Connection this request would go to is switched off.",
		Fix:         "Enable the Connection on Providers, or pick another one under Routing → Protocol defaults.",
		Page:        "Providers", Config: true})
	TargetInvalid = advised(rejected(503, "target_invalid", "Selected target configuration is invalid"), Advice{
		Summary:     "Refused: the Connection's settings are invalid",
		Explanation: "The Connection's base URL cannot be used.",
		Fix:         "Correct the Connection's base URL on Providers, then use Check connection.",
		Page:        "Providers", Config: true})
	IncompatibleProtocol = advised(rejected(400, "incompatible_protocol", "The selected target does not accept this native protocol"), Advice{
		Summary:     "Refused: the Connection does not speak the {protocol} protocol",
		Explanation: "The Connection chosen for this request accepts the other protocol, so aiport cannot forward the request as it is.",
		Fix:         "Pick a Connection of the {protocol} protocol under Routing → Protocol defaults, or bind the Application to one on Integrate.",
		Page:        "Routing", Section: "Protocol defaults", Config: true})
	UpstreamKeyMissing = advised(rejected(503, "upstream_key_missing", "No API key is set for this connection; add it in the dashboard under Providers"), Advice{
		Summary:     "Refused: no API key for the Connection",
		Explanation: "The Connection needs an API key and none is set, so aiport did not send the request.",
		Fix:         "Add the key on the Connection's card under Providers.",
		Page:        "Providers", Section: "API key", Config: true})

	// Admission and request shape.
	ConcurrencyLimit = advised(rejected(429, "concurrency_limit", "Gateway concurrency limit reached"), Advice{
		Summary:     "Refused: too many requests at once",
		Explanation: "aiport was already forwarding as many requests as it allows at the same time.",
		Fix:         "Retry shortly. If it keeps happening, raise proxy.max_concurrent in aiport.yaml.",
		Page:        "Settings"})
	RequestTooLarge = advised(rejected(413, "request_too_large", "Request exceeds the configured admission limit"), Advice{
		Summary:     "Refused: the request is too large",
		Explanation: "The request body is larger than aiport admits.",
		Fix:         "Send a smaller request, or raise proxy.max_request_bytes in aiport.yaml.",
		Page:        "Settings"})
	RequestReadFailed = advised(rejected(400, "request_read_failed", "Could not read request body"), Advice{
		Summary:     "Refused: the request body could not be read",
		Explanation: "The Application stopped sending, or sent too slowly, before the request body was complete.",
		Fix:         "Retry the request from the Application."})
	UnknownCompression = advised(rejected(400, "unknown_compression", "Unknown compression method"), Advice{
		Summary:     "Refused: unknown compression method",
		Explanation: "The Application asked for a compression method aiport does not have.",
		Fix:         "Use a method listed on the Compression page in the X-Aiport-Compression header.",
		Page:        "Compression"})
	AmbiguousModel = advised(rejected(400, "ambiguous_model", "Alias routing requires an unambiguous JSON object"), Advice{
		Summary:     "Refused: the model name is ambiguous",
		Explanation: "An Alias applies, but the request body is not one JSON object with a single model field, so aiport cannot replace the model safely.",
		Fix:         "Send one JSON object with exactly one model field."})

	// Upstream.
	UpstreamRequestFailed = advised(Refusal{Status: 502, Code: "upstream_request_failed", Message: "Could not create upstream request", Outcome: GatewayError}, Advice{
		Summary:     "Gateway error: could not build the provider request",
		Explanation: "aiport could not build a request to the Connection's address.",
		Fix:         "Check the Connection's base URL on Providers.",
		Page:        "Providers"})
	UpstreamUnavailable = advised(Refusal{Status: 502, Code: "upstream_unavailable", Message: "Upstream connection failed", Outcome: GatewayError}, Advice{
		Summary:     "Cannot reach the provider",
		Explanation: "aiport could not connect to the Connection's address.",
		Fix:         "Check that the provider is running and its base URL on Providers is right, then use Check connection.",
		Page:        "Providers"})
	UpstreamTimeout = advised(Refusal{Status: 504, Code: "upstream_timeout", Message: "Upstream response deadline exceeded", Outcome: Timeout}, Advice{
		Summary:     "Timed out: the provider did not answer in time",
		Explanation: "The provider did not start answering before aiport's deadline.",
		Fix:         "Retry. If it keeps happening, check the provider's status, or raise proxy.response_header_timeout_ms in aiport.yaml for slow local models.",
		Page:        "Providers"})
	ClientCancelled = advised(Refusal{Status: 499, Code: "client_cancelled", Message: "Client cancelled request", Outcome: Cancelled}, Advice{
		Summary:     "Cancelled by the Application",
		Explanation: "The Application closed the request before it finished. Nothing is wrong with aiport or the Connection."})
)

// catalog lists every refusal, for looking one up by its code.
var catalog = []Refusal{TargetUnconfigured, TargetDisabled, TargetInvalid, IncompatibleProtocol, UpstreamKeyMissing, ConcurrencyLimit, RequestTooLarge, RequestReadFailed, UnknownCompression, AmbiguousModel, UpstreamRequestFailed, UpstreamUnavailable, UpstreamTimeout, ClientCancelled}
