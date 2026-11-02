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
	Integration Integration `json:"integration"`
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
			HowToObtain: "Pair: POST /v1/aiport/pair {\"name\":\"<application name>\"} (no token needed), show the returned code to the owner, who approves it in the aiport dashboard, then GET the returned poll path to collect your token once.",
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
			{"X-Aiport-Hint", "one line of text", "On a provider 404 for the requested model, sent only to applications allowed to request changes: which change request (kind and endpoint) could fix it. The provider's body is passed through unchanged.", "available"},
			{"X-Aiport-Source", "gateway or upstream", "Whether an error came from aiport or the provider.", "planned"},
		},
		ErrorFormat: `Gateway errors: {"error":{"code":"...","message":"...","request_id":"...","hint":"..."}}. hint is optional: applications allowed to request changes get one on refusals a change could fix (target_unconfigured, target_disabled), naming the change request kind and endpoint. Status and code never depend on it. Provider errors pass through unchanged (status, body, headers).`,
		Errors: []ErrorCode{
			{401, "unauthorized", "Missing or invalid application token."},
			{403, "invalid_host", "Host header is not the configured local address."},
			{403, "invalid_origin", "Browser cross-origin requests are refused."},
			{404, "unsupported_path", "Path is not part of this protocol."},
			{405, "method_not_allowed", "Use the documented method."},
			{400, "incompatible_protocol", "The selected target speaks the other protocol; aiport does not translate."},
			{400, "unknown_compression", "X-Aiport-Compress names a method aiport does not offer."},
			{400, "invalid_pairing", "Pairing needs the application's name (1–128 characters)."},
			{404, "pairing_not_found", "The pairing request expired, was collected or never existed; pair again."},
			{429, "pairing_busy", "Too many pairing requests are waiting; retry after a minute."},
			{400, "invalid_feedback", "Feedback needs request_id and a score between 0 and 1."},
			{403, "application_token_required", "This endpoint describes one application: send that application's own token, not the owner's."},
			{404, "request_not_found", "No retained request with this id for this application."},
			{403, "change_requests_not_allowed", "The owner has not allowed this application to request changes; the owner can enable it in the dashboard (Routing → Applications)."},
			{400, "invalid_change_request", "The change request is incomplete or too large; error.problems lists every problem as {field, problem}."},
			{403, "change_request_refused", "A policy Extension the owner added refused this change request; error.reason says why and error.extension names it."},
			{429, "change_requests_rate_limited", "Too many change requests from this application in the last hour; retry after Retry-After seconds."},
			{404, "change_request_not_found", "No change request with this id for this application."},
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
			{"model_aliases", "available", "Owner-defined alias -> model, per connection or per application (only the model field is rewritten; an application's own alias may select another connection)."},
			{"model_allowlist", "planned", "Owner restricts which models an application may request (403 model_not_allowed)."},
			{"effort_tiers", "planned", "Virtual model names routed to a model by X-Aiport-Effort or the native reasoning effort field."},
			{"capabilities", "available", "GET /v1/aiport/capabilities with the application's own token: what aiport can do for it right now (protocols it can reach and their connection, its own aliases, its binding, whether it may request changes and which kinds of change: application_alias, bind_application). Answers only about the calling application and reflects applied changes at once."},
			{"change_requests", "available", "POST /v1/aiport/requests with the application's own token asks the owner for a change: {\"kind\":\"application_alias|bind_application|...\",\"need\":\"one line\",\"details\":\"...\",\"why\":\"...\",\"scope\":\"application|every_application\",\"scope_case\":\"the argument for that scope\",\"filed_by\":\"optional, informational\"}; each field at most 2,000 characters, 6,000 in total, plain text never sent to a model. Only for applications the owner allowed (see capabilities). GET /v1/aiport/requests lists this application's own requests and GET /v1/aiport/requests/{id} one of them, with status (pending, accepted, rejected, not_supported), the scope granted and the owner's note. Kinds aiport cannot apply are kept as not_supported. 20 requests per application per hour; refused requests do not count."},
			{"mcp", "available", "Model Context Protocol (Streamable HTTP, JSON responses) at POST /mcp, authenticated with Authorization: Bearer <the application's own token> (from AIPORT_APP_TOKEN; the owner's token is refused). Tools: aiport_capabilities, aiport_request_change and aiport_my_requests, answering exactly as GET /v1/aiport/capabilities and /v1/aiport/requests do. Project-level MCP entry: {\"type\":\"http\",\"url\":\"<this aiport>/mcp\",\"headers\":{\"Authorization\":\"Bearer ${AIPORT_APP_TOKEN}\"}}; never commit the token itself."},
			{"pairing", "available", "POST /v1/aiport/pair {name}: an application asks for its own token; the owner approves the shown code in the dashboard; the application collects the token once from the poll path. Requests expire after 10 minutes."},
		},
		Integration: integration(base + "/v1"),
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
	integrating(w, d)
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

// Integration is the checklist for a coding agent wiring an application to
// this gateway. It is part of the live document, so it always matches the
// gateway being integrated with.
type Integration struct {
	Intro string   `json:"intro"`
	Steps []string `json:"steps"`
	Note  string   `json:"note"`
}

func integration(base string) Integration {
	root := strings.TrimSuffix(base, "/v1")
	return Integration{
		Intro: "You are changing an application so its LLM calls go through the aiport gateway on this machine. aiport is a pass-through: keep the provider's native SDK and request format, change only where calls go and which key is sent.",
		Steps: []string{
			"**Get the application its own token by pairing.** If the application has no `AIPORT_APP_TOKEN` yet: `POST " + root + "/v1/aiport/pair` with `{\"name\":\"<application name>\",\"description\":\"<what it uses LLMs for>\"}` (no token needed). Show the returned `code` to the user and tell them to approve it in the aiport dashboard (Routing → Applications). Then `GET` the returned `poll` path every `poll_interval_s` seconds until `status` is `approved` (the reply then contains `token`, shown only once) or `rejected`; requests expire after 10 minutes. Store the token in the application's private, uncommitted configuration as `AIPORT_APP_TOKEN`. Pair once per application, not per run; pair again only if aiport answers `401`.",
			"**Configuration, not code.** Read `AIPORT_BASE_URL` (default `" + root + "`) and `AIPORT_APP_TOKEN` from the application's environment or config. Never hard-code either, and never commit the token.",
			"**Point the SDK at aiport.** OpenAI SDK: `base_url = AIPORT_BASE_URL + \"/v1\"`, `api_key = AIPORT_APP_TOKEN`. Anthropic SDK: `base_url = AIPORT_BASE_URL` (no `/v1`), `api_key = AIPORT_APP_TOKEN`. Plain HTTP: `POST /v1/chat/completions`, `/v1/responses` or `/v1/messages` with `Authorization: Bearer <token>` (Anthropic also accepts `x-api-key`).",
			"**Remove provider keys from the application.** aiport adds the real key upstream. The application must not need `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` for calls routed through aiport.",
			"**Use a protocol the gateway has configured.** `GET " + root + Path + "` lists which protocols are available. A protocol that is not configured returns `503 target_unconfigured`; aiport does not translate between protocols.",
			"**Model names.** Send the model the owner gave you, or an alias the owner set up for this application (`GET " + root + "/v1/aiport/capabilities` with the application's token lists its aliases, binding and reachable protocols). aiport forwards unknown models unchanged; the provider decides.",
			"**Errors.** A body of `{\"error\":{\"code\":...}}` with a code listed at `" + root + Path + "?format=markdown` came from aiport; anything else is the provider's own error, passed through. Retry only `429 concurrency_limit`, `502` and `504`, with backoff. On `401` the token is missing or no longer valid: pair again, do not retry.",
			"**Log `X-Aiport-Request-Id`** from responses next to your own request logs; the owner finds the call in aiport's Activity page with it.",
			"**Optional:** send `X-Aiport-Compress` to shrink large prompts, and `POST /v1/aiport/feedback` with the request id and a 0–1 score so the owner can see whether compression hurts answers. Do not rely on features marked `planned`.",
			"**Let future sessions ask aiport for changes.** Add a project-level MCP entry to the application's repository (Claude Code: `.mcp.json`; other agents: their project MCP config) so coding agents there can reach aiport: `" + mcpEntry(root) + "`. It takes the token from the `AIPORT_APP_TOKEN` environment variable; never write the token itself into the file. Then add this line to the repository's `AGENTS.md`: \"" + AgentsLine + "\" Requesting changes needs the owner's permission: until the owner allows this application in the aiport dashboard (Routing → Applications), aiport answers `403 change_requests_not_allowed`; `GET " + root + "/v1/aiport/capabilities` shows whether it may.",
			"**Verify.** Make one real call, check it succeeds, and ask the owner to confirm it appears under the application's name in aiport's Activity page. If it shows as `unknown`, the token is not the application's own.",
		},
		Note: "aiport stores no prompts or responses, only request metadata, and listens on this machine only. Full protocol reference: " + root + Path + "?format=markdown",
	}
}

// AgentsLine is the line an integrating agent adds to the application's
// AGENTS.md, so future sessions ask aiport instead of working around it.
const AgentsLine = "This app's LLM calls go through aiport. If aiport lacks something it needs, ask with `aiport_request_change` (see /_connect); don't work around it."

// mcpEntry is the project-level MCP server entry for this gateway. It names
// the environment variable holding the token, never the token.
func mcpEntry(root string) string {
	return `{"mcpServers":{"aiport":{"type":"http","url":"` + root + `/mcp","headers":{"Authorization":"Bearer ${AIPORT_APP_TOKEN}"}}}}`
}

func integrating(w func(string, ...any), d Document) {
	w("\n## Integrating an application (for coding agents)\n\n%s\n\n", d.Integration.Intro)
	for i, step := range d.Integration.Steps {
		w("%d. %s\n", i+1, step)
	}
	w("\n%s\n", d.Integration.Note)
}

// Handover renders only the integration instructions, as a self-contained
// brief the owner can paste to a coding agent working in another repository.
func Handover(d Document) string {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	w("# Integrate this application with aiport\n\n%s\n\n", d.Integration.Intro)
	for i, step := range d.Integration.Steps {
		w("%d. %s\n", i+1, step)
	}
	w("\n%s\n", d.Integration.Note)
	return b.String()
}
