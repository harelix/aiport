package refusal

import (
	"fmt"
	"strings"
)

// Advice is what the dashboard tells the Owner about a failed request: a
// short summary for the Activity table, a plain-language explanation, what to
// do, and the dashboard page (and section on it) where that is done. Config
// marks a failure the Owner's configuration causes and can fix, as opposed to
// a passing one such as a timeout or a cancellation.
type Advice struct {
	Code        string `json:"code"`
	Summary     string `json:"summary"`
	Explanation string `json:"explanation"`
	Fix         string `json:"fix,omitempty"`
	Page        string `json:"page,omitempty"`
	Section     string `json:"section,omitempty"`
	Config      bool   `json:"config"`
}

// Failure is what request history keeps about why a request failed: fixed
// identifiers only, never message text.
type Failure struct {
	Route             string // decides the protocol named in the advice
	Status            int
	Outcome           string
	RefusalCode       string
	UpstreamErrorType string
	ConnectionError   string
}

// Connection error classes: why a connection to a provider could not be made.
const (
	DNSFailure        = "dns_failure"
	ConnectionRefused = "connection_refused"
	ConnectionReset   = "connection_reset"
	TLSError          = "tls_error"
	ConnectTimeout    = "timeout"
	ConnectionFailed  = "connection_failed"
)

// connection explains each connection error class.
var connection = map[string]Advice{
	DNSFailure: {
		Summary:     "Cannot reach the provider: its host name does not resolve",
		Explanation: "The host name in the Connection's base URL could not be looked up (DNS failure). It may be mistyped, or this machine may be offline.",
		Fix:         "Check the base URL on Providers and this machine's network, then use Check connection.",
		Page:        "Providers", Config: true},
	ConnectionRefused: {
		Summary:     "Cannot reach the provider: connection refused",
		Explanation: "Nothing is listening at the Connection's address. A local model server is probably not running, or the port is wrong.",
		Fix:         "Start the provider (for example Ollama or llama.cpp) or correct the port in the base URL on Providers, then use Check connection.",
		Page:        "Providers", Config: true},
	ConnectionReset: {
		Summary:     "Cannot reach the provider: the connection was closed",
		Explanation: "The provider closed the connection before answering.",
		Fix:         "Retry. If it keeps happening, check the base URL (http or https) on Providers.",
		Page:        "Providers"},
	TLSError: {
		Summary:     "Cannot reach the provider: the secure connection (TLS) failed",
		Explanation: "The provider's certificate was not accepted, or the address does not speak HTTPS.",
		Fix:         "Check that the base URL on Providers uses the right scheme and host name.",
		Page:        "Providers", Config: true},
	ConnectTimeout: {
		Summary:     "Cannot reach the provider: connecting timed out",
		Explanation: "No connection to the provider could be made before aiport's connect deadline.",
		Fix:         "Check the base URL on Providers and this machine's network, then use Check connection.",
		Page:        "Providers"},
	ConnectionFailed: {
		Summary:     "Cannot reach the provider",
		Explanation: "aiport could not connect to the Connection's address.",
		Fix:         "Check that the provider is running and its base URL on Providers is right, then use Check connection.",
		Page:        "Providers"},
}

// ConnectionAdvice explains a connection error class; ok is false for an
// unknown class.
func ConnectionAdvice(class string) (Advice, bool) {
	a, ok := connection[class]
	a.Code = class
	return a, ok
}

// Upstream advice: the provider answered, with an error status.
var (
	keyRejected = Advice{Code: "key_rejected", Summary: "Provider error: the key was rejected",
		Explanation: "The provider refused the Connection's API key: it is wrong, revoked, or lacks access to this model.",
		Fix:         "Replace the key on the Connection's card under Providers, then use Check connection.",
		Page:        "Providers", Section: "API key", Config: true}
	unknownModel = Advice{Code: "model_not_found", Summary: "Provider error: unknown model",
		Explanation: "The provider does not have the model the Application asked for (or the path does not exist there).",
		Fix:         "Use a model the provider lists, or give the name an Alias (origins[].aliases in aiport.yaml) that points at one it has."}
	rateLimited = Advice{Code: "rate_limited", Summary: "Provider error: rate-limited",
		Explanation: "The provider is limiting how many requests this key may send right now.",
		Fix:         "Wait and retry, or ask the provider for a higher limit."}
	noQuota = Advice{Code: "quota_exceeded", Summary: "Provider error: out of quota or credit",
		Explanation: "The provider's account behind this key has no quota or credit left.",
		Fix:         "Check the plan and billing with the provider, or use another Connection."}
	badRequest = Advice{Code: "invalid_request", Summary: "Provider error: the request was not accepted",
		Explanation: "The provider rejected what the Application sent, for example a parameter or model this provider does not accept.",
		Fix:         "Check the model name and parameters the Application sends."}
	tooLarge = Advice{Code: "request_too_large", Summary: "Provider error: the request is too large",
		Explanation: "The provider rejected the request as too large.",
		Fix:         "Send a smaller request or a shorter context."}
	providerDown = Advice{Code: "provider_error", Summary: "Provider error: the provider failed",
		Explanation: "The provider had a problem of its own, or is overloaded.",
		Fix:         "Retry later, and check the provider's status page."}
	providerOther = Advice{Code: "provider_error", Summary: "Provider error",
		Explanation: "The provider answered with an error."}
	incomplete = Advice{Code: "incomplete", Summary: "The response broke off",
		Explanation: "The response stopped before it was complete: the provider or the Application closed the stream.",
		Fix:         "Retry. If it keeps happening with long answers, raise proxy.stream_idle_timeout_ms in aiport.yaml."}
)

// upstreamTypes maps the providers' own error types and codes (OpenAI's code
// or type, Anthropic's type) to advice; others fall back to the status.
var upstreamTypes = map[string]Advice{
	"authentication_error": keyRejected, "invalid_api_key": keyRejected, "permission_error": keyRejected, "permission_denied": keyRejected,
	"not_found_error": unknownModel, "model_not_found": unknownModel,
	"rate_limit_error": rateLimited, "rate_limit_exceeded": rateLimited,
	"insufficient_quota": noQuota, "billing_error": noQuota,
	"invalid_request_error": badRequest, "request_too_large": tooLarge,
	"overloaded_error": providerDown, "api_error": providerDown, "server_error": providerDown,
}

func byStatus(status int) Advice {
	switch {
	case status == 401 || status == 403:
		return keyRejected
	case status == 404:
		return unknownModel
	case status == 429:
		return rateLimited
	case status == 413:
		return tooLarge
	case status >= 500:
		return providerDown
	case status >= 400:
		return badRequest
	}
	return providerOther
}

// ByCode returns a catalog refusal by its code.
func ByCode(code string) (Refusal, bool) {
	for _, r := range catalog {
		if r.Code == code {
			return r, true
		}
	}
	return Refusal{}, false
}

// Explain tells the Owner why a request failed and what fixes it, from what
// history keeps about it; nil for a request that did not fail.
func Explain(f Failure) *Advice {
	var a Advice
	switch {
	case f.ConnectionError != "":
		c, ok := ConnectionAdvice(f.ConnectionError)
		if !ok {
			c, _ = ConnectionAdvice(ConnectionFailed)
		}
		a = c
	case f.RefusalCode != "":
		r, ok := ByCode(f.RefusalCode)
		if !ok {
			a = Advice{Code: f.RefusalCode, Summary: "Refused: " + strings.ReplaceAll(f.RefusalCode, "_", " "), Explanation: "aiport refused the request."}
			break
		}
		a = r.Owner
	case f.Outcome == "incomplete":
		a = incomplete
	case f.Outcome == "cancelled":
		a = ClientCancelled.Owner
	case f.Status >= 400:
		var ok bool
		if a, ok = upstreamTypes[f.UpstreamErrorType]; !ok {
			a = byStatus(f.Status)
		}
		if f.UpstreamErrorType != "" {
			a.Summary += " (" + f.UpstreamErrorType + ", HTTP " + fmt.Sprint(f.Status) + ")"
		} else {
			a.Summary += " (HTTP " + fmt.Sprint(f.Status) + ")"
		}
	default:
		return nil
	}
	protocol := "OpenAI"
	if f.Route == "/v1/messages" {
		protocol = "Anthropic"
	}
	for _, s := range []*string{&a.Summary, &a.Explanation, &a.Fix} {
		*s = strings.ReplaceAll(*s, "{protocol}", protocol)
	}
	return &a
}
