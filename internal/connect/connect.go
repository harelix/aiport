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
	// Integration is the Application brief; AgentIntegration the
	// coding-agent brief. Briefs lists where each is served.
	Integration      Integration `json:"integration"`
	AgentIntegration Integration `json:"agent_integration"`
	Briefs           []Brief     `json:"briefs"`
	// Adapting is false when the Owner turned the adapting loop off: no
	// Change requests and no hints then.
	Adapting bool `json:"adapting"`
	// AgentPrompt is the one instruction the Owner pastes to a coding agent
	// in an application's repository; it points at Docs, which holds
	// everything both briefs need.
	AgentPrompt string `json:"agent_prompt"`
}
type Auth struct {
	Required    bool   `json:"required"`
	Credential  string `json:"credential"`
	HowToObtain string `json:"how_to_obtain"`
}

// Describe builds the protocol document for the given configuration.
func Describe(c config.Config) Document {
	base := "http://" + c.Address()
	on := !c.Adapting.Disabled
	has := func(id string) bool {
		for _, t := range c.Targets {
			if t.ID == id && !t.Disabled {
				return true
			}
		}
		return false
	}
	d := Document{
		Protocol:    Protocol,
		Version:     Version,
		Gateway:     "aiport",
		BaseURL:     base + "/v1",
		Docs:        base + Path + "?format=markdown",
		AgentPrompt: AgentPrompt(base + Path + "?format=markdown"),
		Translation: false,
		Protocols: []Wire{
			{Name: "openai", Available: has(c.Defaults.OpenAITarget),
				Endpoints:  []Endpoint{{"POST", "/v1/chat/completions", "OpenAI Chat Completions, native body and streaming"}, {"POST", "/v1/responses", "OpenAI Responses, native body and streaming"}, {"GET", "/v1/models", "The model names this Application can use on this protocol (its Catalog), in the OpenAI list shape"}},
				AuthHeader: "Authorization: Bearer <application token>",
				SDKSetup:   "OpenAI SDK: base_url=" + base + "/v1, api_key=<application token>"},
			{Name: "anthropic", Available: has(c.Defaults.AnthropicTarget),
				Endpoints:  []Endpoint{{"POST", "/v1/messages", "Anthropic Messages, native body and streaming"}, {"GET", "/v1/models", "The model names this Application can use on this protocol (its Catalog), in the Anthropic list shape (send anthropic-version)"}},
				AuthHeader: "x-api-key: <application token>  (or Authorization: Bearer <application token>)",
				SDKSetup:   "Anthropic SDK: base_url=" + base + ", api_key=<application token>"},
		},
		Auth: Auth{
			Required:    !c.Auth.Disabled,
			Credential:  "A per-application aiport token. Never send provider API keys to aiport; it adds them upstream.",
			HowToObtain: "Pair: POST /v1/aiport/pair {\"name\":\"<application name>\",\"repository\":\"<absolute path of the repository's top directory>\"} (no token needed; repository is optional and used only once the Owner confirms it), show the returned code to the owner, who approves it on the Integrate page of the aiport dashboard, then GET the returned poll path to collect your token once.",
		},
		Request: []Header{
			{"X-Aiport-Effort", "low, medium, high", "How hard this call is; picks the per-effort model of the Level or task name sent, if the Owner set one. Wins over the provider's own effort field (reasoning_effort, reasoning.effort, output_config.effort), which aiport also reads but never changes; neither means medium. Never forwarded.", "available"},
			{"X-Aiport-Task", "label, up to 64 chars [A-Za-z0-9._-]", "Task kind for activity grouping and routing rules.", "planned"},
			{"X-Aiport-Session", "opaque id, up to 64 chars", "Groups related calls from one session.", "planned"},
			{"X-Aiport-Compress", "whitespace, json, dedupe or all (comma-separated); off", "Opt-in request compression of message text. Unknown methods return 400 unknown_compression.", "available"},
			{"X-Aiport-Version", Version, "aiport Connect version the client expects.", "planned"},
		},
		Response: []Header{
			{"X-Aiport-Request-Id", "UUID", "Gateway id for this call; quote it when reporting problems.", "available"},
			{"X-Aiport-Catalog", "version (12 hex characters)", "On every answer to a request with an Application token, forwarded calls and errors included: the version of this Application's Catalog. It changes only when its task names or where they go, its binding or the protocol defaults it follows change; when it does, read GET /v1/models or GET /v1/aiport/catalog again.", "available"},
			{"X-Aiport-Compression", "methods; status=applied|no_gain|failed; bytes_before=N; bytes_after=N", "Present when compression was requested. On failed or no_gain the original body was sent.", "available"},
			{"X-Aiport-Model", "model id", "Model actually used after a Level, a task name or per-effort models replaced the name sent (the Owner sees it in Activity today).", "planned"},
			{"X-Aiport-Route", "key=value; ...", "Why that model/target was chosen.", "planned"},
			{"X-Aiport-Hint", "one line of text", "On a failing call that is one of an open Issue's failures (a provider 404 for the Issue's model, or a 5xx, timeout or connection error from the Issue's failing Connection) or that sent a task name a Catalog change removed, sent to any Application: the Issue and where to read it. Otherwise, on a provider 404 for the requested model, sent only to Applications allowed to request changes: which change request (kind and endpoint) could fix it. The provider's body is passed through unchanged.", "available"},
			{"X-Aiport-Source", "gateway or upstream", "Whether an error came from aiport or the provider.", "planned"},
		},
		ErrorFormat: `Gateway errors: {"error":{"code":"...","message":"...","request_id":"...","hint":"..."}}. hint is optional: applications allowed to request changes get one on refusals a change could fix (target_unconfigured, target_disabled), naming the change request kind and endpoint. Status and code never depend on it. Provider errors pass through unchanged (status, body, headers).`,
		Errors: []ErrorCode{
			{401, "unauthorized", "Missing or invalid application token."},
			{403, "invalid_host", "Host header is not the listen address or a configured hostname on the listen port."},
			{403, "invalid_origin", "Browser cross-origin requests are refused."},
			{404, "unsupported_path", "Path is not part of this protocol."},
			{405, "method_not_allowed", "Use the documented method."},
			{400, "incompatible_protocol", "The selected target speaks the other protocol; aiport does not translate."},
			{400, "level_unmapped", "The model is a Level (fast, balanced, best) the Owner has not mapped on this protocol, for this application or shared; the message names the Level and protocol. aiport never picks another model instead."},
			{400, "unknown_compression", "X-Aiport-Compress names a method aiport does not offer."},
			{400, "invalid_pairing", "Pairing needs the application's name (1–128 characters); description at most 500, repository one line of at most 4096."},
			{404, "pairing_not_found", "The pairing request expired, was collected or never existed; pair again."},
			{429, "pairing_busy", "Too many pairing requests are waiting; retry after a minute."},
			{400, "invalid_feedback", "Feedback needs request_id and a score between 0 and 1."},
			{403, "application_token_required", "This endpoint describes one Application: send that Application's own token, not the Owner's."},
			{404, "request_not_found", "No retained request with this id for this application."},
			{403, "change_requests_not_allowed", "The Owner has not allowed this Application to request changes; the Owner can allow it on the Integrate page of the aiport dashboard."},
			{400, "invalid_change_request", "The change request is incomplete or too large; error.problems lists every problem as {field, problem}."},
			{403, "change_request_refused", "A policy Extension the Owner added refused this change request; error.reason says why and error.extension names it."},
			{429, "change_requests_rate_limited", "Too many change requests from this Application in the last hour; retry after Retry-After seconds."},
			{404, "change_request_not_found", "No change request with this id for this Application."},
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
			{"catalog", "available", "An Application's Catalog is what it can use through aiport: its task names (its own Aliases, then the shared ones of the Connections it reaches) and the models those Connections list, with a version. GET /v1/models with the Application's own token (the Owner's token gets 403 application_token_required) lists the names usable on that protocol (OpenAI list shape; Anthropic shape when the request carries anthropic-version): task names first (owned_by \"aiport\", or \"(aiport task name)\" in display_name), then the Connections' models; a Connection whose model list cannot be read adds nothing. Every answer to the Application carries X-Aiport-Catalog: <version>. GET /v1/aiport/catalog (MCP tool aiport_catalog) returns {catalog, changes}: the Catalog (task_names with connection and model, binding or defaults, connections, models, version) and its recent Catalog changes, newest first, each with change_id, at_ms, revert_of (set when it was a revert), before and after (versions), removed, added, moved, routing, connections_removed and connections_added. An Application sees only its own Catalog; never keys, URLs, prompts or tokens. Works with adapting off too."},
			{"model_allowlist", "planned", "Owner restricts which models an application may request (403 model_not_allowed)."},
			{"levels", "available", "Every aiport shares three model names, fast, balanced and best: send one as the model to state the proficiency a call needs without naming a provider. The Owner maps each Level per protocol to a Connection and model (an application may have its own mapping, which comes first); only the model field is replaced, the rest of the body is forwarded unchanged. Mapped Levels are listed in GET /v1/models (owned_by \"aiport\", or \"(aiport Level)\" in display_name) and in the Catalog (task_names entries with level: true and protocol). A Level not mapped on the request's protocol is refused with 400 level_unmapped. A Level or task name may have per-effort models (Catalog effort: {low, medium, high}); the effort of a call (X-Aiport-Effort, else the provider's own effort field, else medium) picks one, and the request is still forwarded unchanged apart from the model name."},
			{"capabilities", "available", "GET /v1/aiport/capabilities with the Application's own token: what aiport can do for it right now (protocols it can reach and their connection, its own aliases, its binding, whether it may request changes and which kinds of change: application_alias, bind_application, and catalog: {version, changed_at_ms}, its Catalog version and when it last changed, null when no Catalog change is kept, and guidance: {current, confirmed}, aiport's guidance version and the one this repository's coding agent last confirmed, null when it never did). Answers only about the calling Application and reflects applied changes at once."},
			{"change_requests", "available", "POST /v1/aiport/requests with the Application's own token asks the Owner for a change: {\"kind\":\"application_alias|bind_application|...\",\"need\":\"one line\",\"details\":\"...\",\"why\":\"...\",\"scope\":\"application|every_application\",\"scope_case\":\"the argument for that scope\",\"filed_by\":\"optional, informational\"}; each field at most 2,000 characters, 6,000 in total, plain text never sent to a model. Only for Applications the Owner allowed (see capabilities). GET /v1/aiport/requests lists this Application's own requests and GET /v1/aiport/requests/{id} one of them, with status (pending, accepted, rejected, not_supported), the scope granted and the Owner's note. Kinds aiport cannot apply are kept as not_supported. 20 requests per Application per hour; refused requests do not count."},
			{"mcp", "available", "Model Context Protocol (Streamable HTTP, JSON responses) at POST /mcp, authenticated with Authorization: Bearer <the Application's own token> (from AIPORT_APP_TOKEN; the Owner's token is refused). Tools: aiport_capabilities, aiport_catalog, aiport_guidance, aiport_guidance_done, aiport_request_change, aiport_my_requests, aiport_my_issues and aiport_issue_fixed, answering exactly as GET /v1/aiport/capabilities, /v1/aiport/catalog, /v1/aiport/guidance (and POST /v1/aiport/guidance/done), /v1/aiport/requests and /v1/aiport/issues do. While this repository's guidance is behind aiport's, the initialize instructions start with a one-line reminder and every other tool answer ends with it, as a second text block. Project-level MCP entry: {\"type\":\"http\",\"url\":\"<this aiport>/mcp\",\"headers\":{\"Authorization\":\"Bearer ${AIPORT_APP_TOKEN}\"}}; never commit the token itself."},
			{"issues", "available", "GET /v1/aiport/issues with the Application's own token: its Issues (problems aiport keeps seeing with its calls), newest first, each with summary, evidence (counts, window, request ids), application_fix (what to change in the Application; it starts with the root cause when the Owner had the Issue explained), explanation (that root cause, who wrote it and when; absent until explained) and aiport_fix (proposed, applied, rejected or reverted); GET /v1/aiport/issues/{id} one of them; POST /v1/aiport/issues/{id}/fixed says it was fixed in the Application (state resolving; resolved as application_fix once its calls succeed again). Capabilities show open_issues; a failing call that is one of an open Issue's failures carries an X-Aiport-Hint naming it. A Change that takes away something the Application used within request history (it sent a removed task name, or a model name that went through the moved binding or protocol default) opens an Issue of cause catalog_changed with catalog_change (what changed and what is still available); it needs no Owner decision and resolves when the Application's calls succeed again, on aiport_issue_fixed, or when the Change is reverted. Check it at the start of a session."},
			{"guidance", "available", "aiport numbers the guidance it gives coding agents (the briefs, the AGENTS.md line and the MCP instructions) and raises the number when agents must act on a change; aiport keeps, per Application, the version its coding agent last confirmed. GET /v1/aiport/guidance (MCP tool aiport_guidance) with the Application's own token returns {current, confirmed, confirmed_at_ms, behind, changes, agents_line, steps}: what changed in every version after the confirmed one (each with version, summary and actions), the current AGENTS.md line and the steps to follow; POST /v1/aiport/guidance/done (MCP tool aiport_guidance_done) records the current version. Capabilities show guidance: {current, confirmed}. While behind, the MCP initialize instructions start with a one-line reminder and every other aiport MCP tool answer ends with it. Check it at the start of a session; an Application only ever sees its own state. Works with adapting off too."},
			{"pairing", "available", "POST /v1/aiport/pair {name, description?, repository?}: an application asks for its own token, optionally reporting the absolute path of its git repository (the Owner confirms or edits it on approval; aiport uses it only for Agent runs the Owner starts); the owner approves the shown code in the dashboard; the application collects the token once from the poll path and keeps it in .aiport at its repository root (gitignored). Requests wait 30 minutes and survive restarts; asking again while one waits returns the same code."},
		},
		Privacy: []string{
			"aiport records request metadata only (time, route, model, status, latency) when telemetry is on; never prompts or responses.",
			"X-Aiport-* request headers are removed before forwarding to providers.",
			"aiport listens on loopback only and refuses cross-site browser requests.",
		},
		Adapting: on,
	}
	if !on {
		adaptingOff(&d)
	}
	d.Integration = applicationBrief(d, base)
	d.AgentIntegration = agentBrief(d, base)
	d.Briefs = briefs(base)
	return d
}

// AdaptingOff is the one-line reason Change requests, their MCP tools and
// hints are off.
const AdaptingOff = "Off: the Owner has turned adapting off on this gateway (adapting.disabled), so Change requests answer 409 adapting_off and no hints are sent."

// adaptingOff marks what the adapting loop provides as off, with the reason.
func adaptingOff(d *Document) {
	for i, h := range d.Response {
		if h.Name == "X-Aiport-Hint" {
			d.Response[i].Status, d.Response[i].Description = "off", AdaptingOff
		}
	}
	for i, f := range d.Features {
		switch f.Name {
		case "change_requests", "issues":
			d.Features[i].Status, d.Features[i].Description = "off", AdaptingOff
		case "mcp":
			d.Features[i].Description = "Model Context Protocol (Streamable HTTP, JSON responses) at POST /mcp, authenticated with Authorization: Bearer <the Application's own token> (from AIPORT_APP_TOKEN). aiport_capabilities, aiport_catalog, aiport_guidance and aiport_guidance_done work; aiport_my_issues and aiport_issue_fixed are not offered; aiport_request_change and aiport_my_requests are " + strings.ToLower(AdaptingOff[:1]) + AdaptingOff[1:]
		}
	}
	d.ErrorFormat += " With adapting off no hint is sent."
}

// Markdown renders the document as a guide for people and coding agents.
func Markdown(d Document) string {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	w("# aiport Connect v%s\n\n", d.Version)
	startHere(w, d)
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
	w("\n## Integration briefs\n\nSelf-contained instructions generated from this gateway, to paste to a coding agent:\n\n")
	for _, br := range d.Briefs {
		w("- **%s** (`%s`): %s\n", br.Title, br.URL, br.Audience)
	}
	w("\n`%s?format=handover` returns both.\n", Path)
	render(w, d.Integration, "##")
	render(w, d.AgentIntegration, "##")
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

// AgentPrompt is the instruction an Owner gives a coding agent working in an
// application's repository. docs is this document's markdown URL, which holds
// everything the agent needs for both parts.
func AgentPrompt(docs string) string {
	return "Read " + docs + " and follow it to connect this application to the aiport gateway on this machine. " +
		"Do both parts in order: first the section \"Integrate this application with aiport\" (pair the application for its own token, save it in .aiport at the repository root with .aiport added to .gitignore, make the application read .aiport at startup, point its OpenAI or Anthropic SDK at aiport, remove provider keys, make one test call), " +
		"then the section \"Ask aiport for changes from this repository\" (add the aiport MCP entry, which reads the token from .aiport, and the AGENTS.md line so later sessions can ask aiport for changes). " +
		"When pairing shows a code, tell me the code and wait while I approve it in the aiport dashboard. Never commit .aiport or write the token into any committed file. If the token is ever lost or refused (401), ask me for a New token from the Integrate page into .aiport instead of pairing again."
}

// startHere opens the markdown document with what a coding agent should do,
// so an agent handed only this URL knows where to begin.
func startHere(w func(string, ...any), d Document) {
	w("## Start here\n\n")
	w("**If you are a coding agent** asked to connect an application to aiport: everything you need is in this document. Do these in order:\n\n")
	w("1. **%s** below: give the application its own token by pairing, then point its SDK at aiport.\n", d.Integration.Title)
	w("2. **%s** below: add the aiport MCP entry and the `AGENTS.md` line to the repository, so later sessions can ask aiport for changes.\n", d.AgentIntegration.Title)
	w("3. Tell the user the pairing code when it appears and wait for them to approve it in the aiport dashboard; never commit the token.\n\n")
	w("**If you are the Owner**, paste this to the coding agent working in the application's repository:\n\n> %s\n\n", d.AgentPrompt)
}
