// Package connect describes the aiport Connect protocol: how an application
// talks to this gateway. It is served unauthenticated at GET /_connect and
// contains no secrets, target URLs, model names or key status.
//
// Every capability carries a status. Only "available" features are
// implemented; "planned" ones are published so integrators can prepare, and
// must not be relied on yet.
package connect

import (
	"fmt"
	"strings"

	"aiport/internal/config"
)

const (
	Protocol = "aiport-connect"
	Version  = "1"
	Path     = "/_connect"
)

type Endpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Use    string `json:"use"`
}
type Wire struct {
	Name       string     `json:"name"`
	Available  bool       `json:"available"`
	Endpoints  []Endpoint `json:"endpoints"`
	AuthHeader string     `json:"auth_header"`
	SDKSetup   string     `json:"sdk_setup"`
}
type Header struct {
	Name        string `json:"name"`
	Values      string `json:"values"`
	Description string `json:"description"`
	Status      string `json:"status"`
}
type ErrorCode struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Meaning string `json:"meaning"`
}
type Feature struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type Document struct {
	Protocol    string      `json:"protocol"`
	Version     string      `json:"version"`
	Gateway     string      `json:"gateway"`
	BaseURL     string      `json:"base_url"`
	Docs        string      `json:"docs"`
	Translation bool        `json:"protocol_translation"`
	Protocols   []Wire      `json:"protocols"`
	Auth        Auth        `json:"auth"`
	Request     []Header    `json:"request_headers"`
	Response    []Header    `json:"response_headers"`
	ErrorFormat string      `json:"error_format"`
	Errors      []ErrorCode `json:"errors"`
	Features    []Feature   `json:"features"`
	Privacy     []string    `json:"privacy"`
}
type Auth struct {
	Required    bool   `json:"required"`
	Credential  string `json:"credential"`
	HowToObtain string `json:"how_to_obtain"`
}

// Describe builds the protocol document for the given configuration.
func Describe(c config.Config) Document {
	base := "http://" + c.Listen
	has := func(id string) bool {
		for _, t := range c.Targets {
			if t.ID == id && !t.Disabled {
				return true
			}
		}
		return false
	}
	return Document{
		Protocol:    Protocol,
		Version:     Version,
		Gateway:     "aiport",
		BaseURL:     base + "/v1",
		Docs:        base + Path + "?format=markdown",
		Translation: false,
		Protocols: []Wire{
			{Name: "openai", Available: has(c.Defaults.OpenAITarget),
				Endpoints:  []Endpoint{{"POST", "/v1/chat/completions", "OpenAI Chat Completions, native body and streaming"}, {"POST", "/v1/responses", "OpenAI Responses, native body and streaming"}},
				AuthHeader: "Authorization: Bearer <application token>",
				SDKSetup:   "OpenAI SDK: base_url=" + base + "/v1, api_key=<application token>"},
			{Name: "anthropic", Available: has(c.Defaults.AnthropicTarget),
				Endpoints:  []Endpoint{{"POST", "/v1/messages", "Anthropic Messages, native body and streaming"}},
				AuthHeader: "x-api-key: <application token>  (or Authorization: Bearer <application token>)",
				SDKSetup:   "Anthropic SDK: base_url=" + base + ", api_key=<application token>"},
		},
		Auth: Auth{
			Required:    !c.Auth.Disabled,
			Credential:  "A per-application aiport token. Never send provider API keys to aiport; it adds them upstream.",
			HowToObtain: "Ask the aiport owner for an application (origin) token. Self-service pairing is planned.",
		},
		Request: []Header{
			{"X-Aiport-Effort", "low, medium, high", "Task difficulty used to pick a model within a tier.", "planned"},
			{"X-Aiport-Task", "label, up to 64 chars [A-Za-z0-9._-]", "Task kind for activity grouping and routing rules.", "planned"},
			{"X-Aiport-Session", "opaque id, up to 64 chars", "Groups related calls from one session.", "planned"},
			{"X-Aiport-Compress", "whitespace, json, dedupe or all (comma-separated); off", "Opt-in request compression of message text. Unknown methods return 400 unknown_compression.", "available"},
			{"X-Aiport-Version", Version, "aiport Connect version the client expects.", "planned"},
		},
		Response: []Header{
			{"X-Aiport-Request-Id", "UUID", "Gateway id for this call; quote it when reporting problems.", "available"},
			{"X-Aiport-Compression", "methods; status=applied|no_gain|failed; bytes_before=N; bytes_after=N", "Present when compression was requested. On failed or no_gain the original body was sent.", "available"},
			{"X-Aiport-Model", "model id", "Model actually used after aliases or tier routing.", "planned"},
			{"X-Aiport-Route", "key=value; ...", "Why that model/target was chosen.", "planned"},
			{"X-Aiport-Source", "gateway or upstream", "Whether an error came from aiport or the provider.", "planned"},
		},
		ErrorFormat: `Gateway errors: {"error":{"code":"...","message":"...","request_id":"..."}}. Provider errors pass through unchanged (status, body, headers).`,
		Errors: []ErrorCode{
			{401, "unauthorized", "Missing or invalid application token."},
			{403, "invalid_host", "Host header is not the configured local address."},
			{403, "invalid_origin", "Browser cross-origin requests are refused."},
			{404, "unsupported_path", "Path is not part of this protocol."},
			{405, "method_not_allowed", "Use the documented method."},
			{400, "incompatible_protocol", "The selected target speaks the other protocol; aiport does not translate."},
			{400, "unknown_compression", "X-Aiport-Compress names a method aiport does not offer."},
			{400, "invalid_feedback", "Feedback needs request_id and a score between 0 and 1."},
			{404, "request_not_found", "No retained request with this id for this application."},
			{400, "ambiguous_model", "Body must be one JSON object with a single model field when aliases apply."},
			{413, "request_too_large", "Request body exceeds the admission limit."},
			{429, "concurrency_limit", "Too many concurrent requests; retry with backoff."},
			{503, "target_unconfigured", "No target is configured for this protocol or application."},
			{503, "target_disabled", "The selected target is disabled."},
			{503, "upstream_key_missing", "The owner has not set a provider API key for the target yet."},
			{502, "upstream_unavailable", "Could not connect to the provider."},
			{504, "upstream_timeout", "The provider did not respond in time."},
		},
		Features: []Feature{
			{"native_forwarding", "available", "Requests and streams are forwarded byte-for-byte to the selected provider, unless the application asks for compression."},
			{"compression", "available", "X-Aiport-Compress shrinks message text (tools and schemas untouched). Sizes and outcomes are recorded per request to compare against uncompressed traffic."},
			{"feedback", "available", "POST /v1/aiport/feedback {\"request_id\":\"<X-Aiport-Request-Id>\",\"score\":0..1} scores a response, so the owner can see whether compression hurts quality."},
			{"model_aliases", "available", "Owner-defined alias -> model per target (only the model field is rewritten)."},
			{"model_allowlist", "planned", "Owner restricts which models an application may request (403 model_not_allowed)."},
			{"effort_tiers", "planned", "Virtual model names routed to a model by X-Aiport-Effort or the native reasoning effort field."},
			{"capabilities", "planned", "GET /aiport/v1/capabilities: models, tiers and hints allowed for the calling application."},
			{"pairing", "planned", "POST /aiport/v1/connect: request a token, approved by the owner in the dashboard."},
		},
		Privacy: []string{
			"aiport records request metadata only (time, route, model, status, latency) when telemetry is on; never prompts or responses.",
			"X-Aiport-* request headers are removed before forwarding to providers.",
			"aiport listens on loopback only and refuses cross-site browser requests.",
		},
	}
}

// Markdown renders the document as a guide for people and coding agents.
func Markdown(d Document) string {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	w("# aiport Connect v%s\n\n", d.Version)
	w("How to talk to this aiport gateway. Machine-readable: `GET %s` (JSON).\n\n", Path)
	w("Base URL: `%s`. aiport speaks the providers' **native** APIs; existing OpenAI/Anthropic SDKs work by changing the base URL and API key. There is no protocol translation.\n\n", d.BaseURL)
	w("## Authentication\n\n")
	if d.Auth.Required {
		w("Required. %s\n\n%s\n\n", d.Auth.Credential, d.Auth.HowToObtain)
	} else {
		w("Currently **disabled** on this gateway (development). Send your application token anyway, so your calls are attributed to your application and keep working when authentication is re-enabled.\n\n")
	}
	w("## Protocols\n\n")
	for _, p := range d.Protocols {
		state := "available"
		if !p.Available {
			state = "not configured on this gateway"
		}
		w("### %s (%s)\n\n", p.Name, state)
		for _, e := range p.Endpoints {
			w("- `%s %s`: %s\n", e.Method, e.Path, e.Use)
		}
		w("- Auth: `%s`\n- Setup: %s\n\n", p.AuthHeader, p.SDKSetup)
	}
	w("Example:\n\n```sh\ncurl %s/messages \\\n  -H \"x-api-key: $AIPORT_APP_TOKEN\" -H \"anthropic-version: 2023-06-01\" \\\n  -H \"content-type: application/json\" \\\n  -d '{\"model\":\"MODEL\",\"max_tokens\":64,\"messages\":[{\"role\":\"user\",\"content\":\"ping\"}]}'\n```\n\n", d.BaseURL)
	header := func(title string, hs []Header) {
		w("## %s\n\n| Header | Values | Status | Meaning |\n|---|---|---|---|\n", title)
		for _, h := range hs {
			w("| `%s` | %s | %s | %s |\n", h.Name, h.Values, h.Status, h.Description)
		}
		w("\n")
	}
	header("Request headers (optional hints)", d.Request)
	header("Response headers", d.Response)
	w("## Errors\n\n%s\n\n| Status | Code | Meaning |\n|---|---|---|\n", d.ErrorFormat)
	for _, e := range d.Errors {
		w("| %d | `%s` | %s |\n", e.Status, e.Code, e.Meaning)
	}
	w("\n## Features\n\n| Feature | Status | Description |\n|---|---|---|\n")
	for _, f := range d.Features {
		w("| %s | %s | %s |\n", f.Name, f.Status, f.Description)
	}
	w("\nOnly `available` features are implemented. `planned` features are published so you can prepare; do not rely on them yet.\n\n## Privacy\n\n")
	for _, p := range d.Privacy {
		w("- %s\n", p)
	}
	return b.String()
}
