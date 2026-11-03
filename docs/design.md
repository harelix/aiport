# aiport — Design and Implementation Specification

Version: 0.4 · Date: 2026-11-01 · Status: In implementation; expanded scope recorded

Project: **aiport** · Repository/executable: `aiport` · Configuration: `aiport.yaml`

Positioning: **The lightweight, open-source AI gateway.** Naming does not establish repository, domain, or trademark availability.

## Authentication and theme revision — 2026-11-01

The user requested these changes after the dashboard preview. They supersede the
original bearer-only browser login and the original unconditional local-auth rule:

- First-run dashboard access verifies the startup management token, then lets the
  owner set a password and an optional username. Later browser sign-ins use that
  account and an HttpOnly browser session. This remains a single-user local tool.
- A clickable “Forgot username or password?” popup explains how to retrieve the
  management token from the server shell or private onboarding credential file.
  Pasting that token bypasses account login and allows resetting both credentials.
  Agent/provider tokens must never grant management recovery.
- Settings must allow disabling and re-enabling local authentication. Disabling
  applies to both dashboard APIs and inference. Preserve credentials for recovery,
  keep loopback/Host/Origin/CSRF protections, and show the disabled state clearly.
- Authentication changes validate and atomically update the existing YAML, with
  no plaintext stored password. Credential changes revoke prior browser sessions.
- Use Nord throughout the interface, including both modes, setup/recovery/settings,
  charts/statuses and CLI accents: https://www.nordtheme.com/docs/colors-and-palettes/.
  Colors must come from the official Polar Night, Snow Storm, Frost and Aurora
  palettes. Neutral text/status labels accompany color; no other visual theme.

## License decision — 2026-11-01

The maintainer adopted **Apache-2.0** for the source (`LICENSE`, `NOTICE`), with the
aiport name and windsock logo reserved under `TRADEMARKS.md` and contributions
accepted under the DCO (`CONTRIBUTING.md`). This resolves the license item in §11.

## Task queue revision — 2026-11-01

Requested by the maintainer: a fourth dashboard area, **Tasks**, a queue of work
for coding agents. The owner owns intent and priority; agents will own execution
planning (spec sections 2–3 pending). Implemented now: capture with immutable
two-word names (`task-amber-fox`, collision-checked, never reused) separate from an
internal UUID; title, instructions, acceptance criteria, status (pending,
in_progress, blocked, completed, failed, cancelled), priority 0–3 and a bounded
append-only history with actor. Stored in private `<config>.tasks.json` until
SQLite persistence lands. API: `/api/v1/tasks` (management access; mutations need
the CSRF header; `X-Aiport-Actor` attributes agent changes).

## Application connectivity revision — 2026-11-01

Requested by the maintainer: a connectivity standard for applications, exposed by
the gateway itself. `GET /_connect` (public, loopback, no secrets) publishes
**aiport Connect v1** as JSON or Markdown: native endpoints, auth, hint/response
headers, stable error codes, and per-feature status. Model allowlists per
application, effort-based tiers, capabilities and dashboard-approved pairing are
recorded as planned and must be marked available only when implemented.

## Credential handling revision — 2026-11-01

Requested by the maintainer: no manual token exports, and provider keys entered in
the dashboard. Supersedes the environment-variable-only rule for upstream keys.

- Local tokens are created once and kept in the private `<config>.env` (0600);
  `serve` prints the management token only to an interactive terminal.
- Provider API keys may be entered per connection in Settings. They are stored in
  the same private file (never YAML), applied without restart, write-only (never
  returned or displayed), and take precedence over the environment variable.
- A missing provider key no longer blocks startup; requests to that connection
  return 503 `upstream_key_missing`. An OS keychain adapter remains optional later work.

## Provider presets revision — 2026-11-01

Requested by the maintainer: provider configuration must be generic, not a
hard-coded regional URL; Settings groups connections as local vs remote APIs.

- A built-in provider catalog (`internal/config/providers.go`) owns each preset's
  protocol, endpoint template, key variable and location (local or remote).
- A target may name only `provider` (plus `region` for regional providers such as
  Bedrock). `base_url` is an optional override and may contain `{region}`. The
  resolved endpoint is derived at load time and never written back to YAML.
- Unused or missing regions are rejected, not ignored. `aiport init` writes the
  generic form and asks for a region only when a regional provider is chosen.
- Providers with several API surfaces list endpoints, each with per-protocol URL
  templates. Bedrock: `endpoint: runtime|mantle` × `protocol: openai|anthropic`.
- Connections are managed from the dashboard: `POST /api/v1/targets`, `PATCH`/`DELETE
  /api/v1/targets/{id}`, `PUT /api/v1/defaults` (CSRF header required). Removing a
  connection still used by a default, agent or alias is refused with the reason. The YAML is edited node-wise,
  revalidated as a whole, written atomically (0600), then applied live.
- Settings is split into Providers, Routing, Coding agents and Account. Providers
  shows Remote APIs and Local groups, each with configured connections and the
  presets available to add.

## Request compression revision — 2026-11-01

Requested by the maintainer: applications may ask for compression per request, and
its effect on responses is evaluated and shown as trends.

- Opt-in only: `X-Aiport-Compress` (whitespace, json, dedupe, all). Without it,
  forwarding stays byte-for-byte. Only message text is rewritten; tools and schemas
  never are. Failures or no gain forward the original body.
- Evaluation without storing content: per request the methods, status, request
  sizes, outcome, latency and an optional application score
  (`POST /v1/aiport/feedback`, 0–1, own requests only).
- Verdicts compare against uncompressed traffic in the same window and abstain
  below 20 requests (feedback needs 10 scores per side).
- Evaluator (2026-11-02, maintainer decision): the owner chooses any configured
  connection and model as the evaluator (`evaluation:` in YAML or the dashboard).
  Judge calls use that connection's protocol and key at temperature 0. Planned
  scorers (shadow agreement with noise floor, needle probes, evidence recall,
  source faithfulness) build on it; records keep counts and kinds, not text.
- Model-based pruning (LLMLingua-style) and shadow A/B comparison are not built;
  shadow testing doubles provider cost and would need explicit owner opt-in.

## User scope revision — 2026-11-01

The maintainer requested the following changes after M0. These requirements
supersede the earlier Bedrock deferral and expand the original onboarding scope.
They are requested capabilities, not shipped support.

- Initial provider set: OpenAI (interpretation of “openapi”), Anthropic, AWS
  Bedrock, Meta Model API, Ollama, and llama.cpp (interpretation of “llamacapp”).
  Do not add other providers in this scope.
- Guided first-run onboarding must configure provider connections and credential
  references, generate separate local credentials for selected coding agents,
  validate configuration, and offer explicit connection checks. Local servers
  may omit upstream credentials. Provider credentials and local gateway tokens
  are different secrets and must never be interchanged.
- Coding-agent integration is part of onboarding: detect or let users select
  installed clients, show the proposed endpoint/model/credential configuration,
  preserve unrelated settings, and apply the selected integration with a backup.
  Test each advertised client/protocol combination. This configures existing
  agents; aiport does not execute or orchestrate agents itself.
- Provide a modern, clean embedded web interface and a terminal onboarding
  command. The interface covers onboarding, Overview, Activity, and Settings;
  use a restrained neutral visual style, clear hierarchy, accessible controls,
  and explicit loading, empty, error, and degraded states. Do not show invented
  activity or imply that an unimplemented provider is usable.
- Keep sequential implementation gates. M1 establishes faithful forwarding;
  onboarding is delivered incrementally with working connections, followed by
  metadata and the embedded dashboard. Bedrock and Meta join M4 compatibility
  gates. Their support requires explicit endpoint/authentication contracts,
  deterministic fixtures, and opt-in live evidence; no implicit translation.

Meta source supplied by the maintainer: https://dev.meta.ai/platform, redirecting
at inspection to https://dev.meta.ai/. The official API overview inspected on
2026-11-01 documents `https://api.meta.ai/v1` and bearer authentication using an
API key (`MODEL_API_KEY`): https://dev.meta.ai/docs/overview. The overview describes
Chat Completions, Responses, and Messages surfaces; verify exact endpoint behavior
and client compatibility during implementation before advertising support.

## 1. Purpose and evidence boundary

Build a lightweight, open-source, local LLM gateway: connect existing API clients and coding agents, forward their calls faithfully, and show useful request metadata in a local dashboard.

This document consolidates the “LLM Gateway Comparison” discussion, including its inline build draft and subsequent OSS naming requirements. The earlier conversation contained an inline draft, **not an actual saved `.md` file**. This file is the concrete handoff artifact.

Source discussion: `chatgpt-conversation://6ae7264f-dfcc-83eb-9b38-abd6bfa604cf`.

**Nothing in this specification establishes shipped support.** No aiport repository, implementation, release, or live integration was inspected for this handoff. All requirements below are planned unless a future compatibility matrix links them to passing tests and a released version. Configuration syntax, defaults, API paths, and schema below are proposed implementation contracts, not existing commands or guarantees.

Use MUST for release requirements, SHOULD for preferred implementation choices, and MAY for explicitly optional work. Complete one milestone before expanding scope.

## 2. Product boundary

The first public MVP MUST provide:

- A Go executable serving both an HTTP gateway and an embedded React/TypeScript UI.
- Provider-native OpenAI Chat Completions and Responses forwarding, plus Anthropic Messages forwarding.
- Configured OpenAI-compatible local targets, initially validated against Ollama and llama.cpp.
- Fixed target selection and optional explicit model aliases; no automatic optimization.
- SQLite metadata observability: origin grouping, timestamps, target/provider/model, timing, usage, estimated cost, and outcomes.
- Bounded asynchronous telemetry that cannot block forwarding on database availability.
- Safe local defaults, documented limitations, deterministic tests, and OSS release/contribution documentation.

A developer preview MAY ship earlier with a smaller, accurately labeled compatibility matrix. The public MVP is complete only after all MVP gates below pass.

Deferred: HelixDB/graph storage; fallback/retry policies; intelligent routing; payload capture; OS-keychain integration if the environment-variable approach is sufficient.

Non-goals: LiteLLM feature parity, universal API translation, agent execution, MCP gateway, multi-user RBAC, distributed clustering, Redis/PostgreSQL, Kubernetes management, semantic search, embeddings APIs, quality scoring, billing enforcement, budgets, plugin marketplaces, or distributed tracing infrastructure. Do not add speculative abstractions for these.

## 3. Architecture

```text
API client / coding agent
        |
        v
Go net/http listener (loopback by default)
  authenticate -> identify origin -> select fixed target -> native proxy -> upstream
                                                       |
                                         bounded metadata observer
                                                       |
                                         nonblocking bounded queue
                                                       |
                                         batch writer -> SQLite WAL

Browser -> embedded React/TypeScript assets + local management/query API
```

Backend: Go standard `net/http`, a maintained SQLite driver, YAML configuration, structured metadata logging, and `go:embed` for built UI assets. End users MUST NOT need Node.js, a Go toolchain, Docker, or a separate database. Frontend development/builds require the frontend toolchain.

Frontend: React, TypeScript, Vite; Tailwind/shadcn components MAY be used if they keep dependencies small. Use at most one lightweight chart library. Avoid an animation framework.

Keep the forwarding path independent of SQLite and UI queries. Provider code SHOULD be small authentication/path/usage helpers around a common byte-forwarding implementation. Do not introduce SDK-driven response reconstruction or a generic provider plugin framework.

### SQLite now; HelixDB deferred

The immediate queries are time ranges, filters, counts, and grouped aggregates. SQLite supplies local persistence inside the process without another service. Use WAL, migrations, indexes, and batched writes. Do not build a general storage abstraction solely to anticipate replacement.

Revisit HelixDB only when a real feature requires graph relationships among sessions, calls, tools, and dependencies, and a separate design demonstrates value and acceptable operational cost. This decision makes no claim about current HelixDB packaging or feature availability. No graph dependency belongs in the MVP.

## 4. Protocol and compatibility contract

| Interface / target | Planned MVP behavior | Evidence required before “supported” |
|---|---|---|
| OpenAI `POST /v1/chat/completions` | Native streaming/non-streaming pass-through | Fixtures plus recorded live smoke test |
| OpenAI `POST /v1/responses` | Native streaming/non-streaming pass-through | Responses event/tool fixtures plus live smoke test |
| Anthropic `POST /v1/messages` | Native streaming/non-streaming pass-through | Messages events, tools, version/beta-header fixtures plus live smoke test |
| Ollama OpenAI-compatible endpoint | Tested subset of Chat Completions | Pinned server/model versions and explicit feature results |
| llama.cpp OpenAI-compatible server | Tested subset of Chat Completions | Pinned server/model versions and explicit feature results |
| AWS Bedrock | Requested initial provider; not implemented | Verified AWS endpoint/auth contract, native fixtures, pinned client and live evidence |
| Meta Model API | Requested initial provider; not implemented | Verified Meta endpoint/auth contract, native fixtures, pinned client and live evidence |
| Cross-provider protocol translation | Out of scope | No implicit compatibility claim |

“Native” means an OpenAI request goes to a configured OpenAI-protocol target, and an Anthropic request goes to an Anthropic-protocol target. Routing an Anthropic request to an OpenAI-only target MUST fail locally before sending it. OpenAI-compatible does not imply support for Responses, every parameter, every tool mode, or every model feature.

Only the listed inference endpoints are in scope. Files, batches, assistants, realtime/WebSockets, uploads, resource retrieval, and other APIs are not implicitly included. Unsupported local paths return a documented local error. Unknown fields on a supported native route MUST pass through; an upstream may reject them itself.

The coding agent MUST verify current primary provider documentation during implementation, pin tested client/server versions, and record exact endpoint limitations. This design intentionally does not freeze provider event schemas from memory.

### Forwarding fidelity

- Forward request body bytes unchanged in ordinary pass-through. Inspect only bounded metadata. If a configured alias changes `model`, change only that top-level field, preserve unknown fields, and document that reserialization can alter formatting. Ambiguous duplicate routing fields MUST fail before forwarding when alias rewriting is needed.
- Preserve upstream status codes, error bodies, content types, request IDs, native version/beta headers, and relevant end-to-end headers. Remove hop-by-hop headers and headers nominated by `Connection` as required by HTTP.
- Replace incoming authentication with the selected target's configured upstream credentials. Never forward the local gateway token, trusted-origin headers, or arbitrary cookies upstream. Do not place upstream keys in client-visible responses.
- Keep SSE event bytes, order, IDs, multiline data, tool-argument fragments, comments, usage events, and terminal events intact. Do not synthesize text-only streams, add SSE markers, or normalize tool arguments.
- Flush promptly, including heartbeat traffic. TCP/HTTP chunk boundaries may change; byte order and provider event semantics MUST survive. Do not promise identical network chunk boundaries.
- Disable transparent compression/decompression transformations or handle them without changing the body/header contract. Parse only observable encodings; observation failure does not justify modifying the stream.
- Propagate client cancellation upstream. Apply downstream backpressure rather than accumulating unbounded buffers. Limit concurrent requests and reject excess admission with a documented local error.
- No whole-response buffering for streaming. Do not impose a short global timeout that truncates healthy long generations. Use configurable connect, response-header, and stream-idle timeouts, with idle reset on bytes received.
- Before response commitment, local failures return a small redacted structured error with a gateway request ID. After commitment, an upstream disconnect or timeout terminates the stream and records an incomplete outcome; do not append invented provider events or a second HTTP error.
- MVP performs **no automatic retries or fallback**, including after partial output. A paid generation is not generally safe to replay merely because no output reached the client.

Bounded metadata observation MUST fail open: invalid JSON, unknown event types, oversized events, or observer overflow make the affected metrics unknown/partial while raw forwarding continues. Define request-body limits separately; an exceeded configured admission limit returns 413 before forwarding, without persisting the body.

## 5. Targets, configuration, and minimal routing

A **target** is a named connection: provider identity, protocol, base URL, and secret reference. Multiple targets may use the same provider (for example two OpenAI accounts or two local servers). **Origin** identifies the calling local application. These concepts MUST remain distinct in UI and storage.

Choose one authoritative configuration store: `aiport.yaml`. Settings updates validate and atomically replace it; SQLite stores request history, not a competing provider configuration. Startup validates unknown keys, duplicate IDs, protocols, URL structure, secret references, and route/alias compatibility. A failed edit leaves the last valid config active. In-flight calls keep their initial configuration snapshot.

Proposed minimal configuration:

```yaml
version: 1
listen: 127.0.0.1:8787
auth:
  token_env: AIPORT_LOCAL_TOKEN
targets:
  - id: openai-main
    provider: openai
    protocol: openai
    base_url: https://api.openai.com/v1
    api_key_env: OPENAI_API_KEY
  - id: anthropic-main
    provider: anthropic
    protocol: anthropic
    base_url: https://api.anthropic.com/v1
    api_key_env: ANTHROPIC_API_KEY
  - id: ollama-local
    provider: ollama
    protocol: openai
    base_url: http://127.0.0.1:11434/v1
origins:
  - id: codex
    name: Codex
    token_env: AIPORT_CODEX_TOKEN
    target: openai-main
defaults:
  openai_target: openai-main
  anthropic_target: anthropic-main
telemetry:
  enabled: true
  retention_days: 7
  queue_capacity: 4096
  batch_size: 100
  flush_interval_ms: 250
  shutdown_drain_ms: 2000
```

The numbers above are conservative proposed defaults, not measured capacity guarantees. Set a configurable maximum encoded event size (initial proposal: 16 KiB), and bound metadata string lengths. Verify compatibility of any chosen SQLite driver with single-executable releases on all advertised platforms before adopting it.

Interpret `base_url` as a fixed API prefix. Append endpoint-relative paths exactly once (`chat/completions`, `responses`, `messages`); test trailing slashes and avoid `/v1/v1`. Preserve endpoint query semantics without logging query values. Do not accept a request-supplied arbitrary upstream URL. Configured loopback HTTP is allowed for local servers; remote credentials require HTTPS.

Fixed selection order:

1. An authenticated origin's configured target, if present.
2. The default target for the incoming protocol.
3. Otherwise reject with a clear local configuration error.

Explicit model aliases MAY map `(selected target, alias)` to an upstream model identifier. They MUST NOT silently switch providers or translate protocols. Without an alias, forward the requested model unchanged. Preserve both requested and forwarded model values in metadata. Native API fields remain subject to upstream validation; do not invent a generic required-model rule for every route.

Settings MUST support listing, adding, editing, disabling, and deleting targets; selecting defaults; configuring origins; and managing aliases. Referenced targets cannot be deleted until references are removed. Disabled targets reject new calls. Historical rows retain attribution after configuration changes/deletion.

Connection testing MUST be explicit and adapter-specific. Prefer a non-generating authenticated probe where supported; otherwise explain that the chosen smoke test may incur cost before running it. A successful probe alone never establishes full compatibility.

## 6. Origin grouping

Attribution precedence:

1. Validated per-origin local credential mapped to a configured origin.
2. Explicit trusted origin header, only when enabled for an authenticated configured caller.
3. Configured identifier on a dedicated local connection.
4. `unknown`.

The global local token permits access but does not itself identify an origin. Arbitrary headers and user-agent strings are not authenticated identities. A supplied session ID is an optional grouping label, not an authorization boundary. Bound its length; never use it as a metric label with unbounded cardinality. Strip internal grouping headers upstream.

Named examples: Codex, Claude Code, Local Script. Names are examples, not verified integrations. Avoid fingerprinting applications from payload content. Display unknown origins explicitly and document how users configure attribution.

## 7. Observability definitions

Every admitted inference call gets a gateway UUID before upstream dispatch. Record locally rejected calls where possible, with target/model fields nullable. Telemetry-disabled mode does not persist call records.

| Field | Meaning / null behavior |
|---|---|
| `started_at`, `ended_at` | UTC wall-clock timestamps in RFC 3339 at API boundary; store integer epoch milliseconds |
| `latency_ms` | Monotonic elapsed time from handler entry to forwarding termination, excluding telemetry persistence |
| `ttfb_ms` | Time to first upstream response-body byte; distinct from TTFT |
| `ttft_ms` | Streaming time to first observed generated content delta, including tool-argument content; exclude headers, heartbeats, and role-only events |
| `origin_id`, `session_id` | Attribution/grouping labels; origin may be unknown, session nullable |
| `target_id`, `provider` | Selected configured connection and its declared provider; never inferred as verified backend identity |
| `requested_model`, `forwarded_model` | Incoming model and model sent after any explicit alias |
| `reported_model` | Upstream-reported identifier, if observed; otherwise null, not a fabricated “actual model” |
| `route`, `method` | Normalized supported endpoint and HTTP method; no raw query strings |
| `http_status`, `outcome` | Response status if committed; outcome separately describes success, rejection, upstream error, cancellation, timeout, or incomplete stream |
| `input_tokens`, `output_tokens` | Upstream-reported counts, nullable; do not estimate from characters |
| `usage_status` | `reported`, `partial`, or `unknown`, plus source/observer version |
| `estimated_cost_usd`, `cost_status` | Versioned local estimate or null with reason; never equate unknown with zero |
| `error_category` | Bounded sanitized category, not a raw provider error or body |

TTFT is an observation of the first content-bearing event, not an exact tokenizer measurement. Leave it null for non-streamed calls or when event observation is unavailable. Document the protocol-specific event predicate and test it. Do not substitute TTFB for TTFT.

Usage extraction must recognize tested native usage events without changing requests to force providers to emit usage. Some clients/providers omit final usage, especially on cancellation. Do not double-count cumulative stream snapshots. Cache/reasoning categories need explicit semantics; if pricing cannot account for them reliably, return a partial/unknown cost rather than applying a misleading simple formula.

Costs use a small versioned local pricing table, per-million-token rates, exact decimal/integer arithmetic, currency, price version, and calculation timestamp. No background network price fetching in MVP. A known simple case is `(input_tokens × input_rate + output_tokens × output_rate) / 1,000,000`; use only when applicable. Retain the historical estimate when pricing changes. Aggregates show known estimated totals **and coverage** (for example priced calls / total calls). Local models are zero only if explicitly configured as zero-cost; unknown pricing remains unknown.

Dashboard aggregates MUST share the same filter semantics as Activity. Show request volume, origin and provider/model grouping, known token totals, estimated cost coverage, average/p95 latency, error rate, and newest calls. Define error rate as upstream errors + gateway failures + incomplete streams divided by admitted calls in the selected window; display client cancellations and pre-admission rejections separately. Exclude null latency values from percentiles and show sample counts. Use a documented percentile method and cap query ranges/result sizes.

## 8. Bounded telemetry and storage

Proxy handlers MUST never wait for SQLite. Produce one bounded final metadata event and attempt a nonblocking enqueue. If full, drop that event, increment a visible drop counter, and emit rate-limited metadata-only diagnostics. Do not block, spill payloads to disk, or spawn a goroutine per dropped event. Observer parsing uses bounded state/buffers; if it needs a queue, overflow disables observation for that request while forwarding continues.

A single writer batches inserts, with bounded busy timeout/retries. Database lock, disk-full, or write failure increments a separate loss counter; after a bounded retry budget the batch is discarded. No unbounded retry queue. Expose queue depth, dropped events, failed writes, and last successful write in local health/UI. Mark dashboard data as incomplete when drops occur. Loss counters may reset on restart; disclose this.

SQLite MUST use WAL, parameterized SQL, schema migrations, and bounded readers. Run retention deletion in small batches by `started_at`; keep cleanup off the forwarding path. Seven-day retention is the proposed default. Do not vacuum synchronously per request. Document that WAL/backups/free pages can retain deleted metadata; deletion is not a secure-erasure guarantee.

Suggested initial schema (names may follow repository conventions):

```text
schema_migrations(version, applied_at)
requests(
  id PRIMARY KEY, started_at_ms, ended_at_ms,
  origin_id, origin_name_snapshot, session_id,
  target_id, target_name_snapshot, provider, protocol,
  method, route, requested_model, forwarded_model, reported_model,
  http_status, outcome, error_category, streaming,
  latency_ms, ttfb_ms, ttft_ms,
  input_tokens, output_tokens, usage_status, usage_source,
  estimated_cost_usd_decimal, cost_status, price_version,
  cost_calculated_at_ms, observer_version
)
```

Index `started_at_ms`, `(origin_id, started_at_ms)`, and `(target_id, started_at_ms)`. Add provider/model indexes only if measured queries need them. Origin/target IDs can be stable configuration IDs; history snapshots prevent rename/deletion from destroying attribution. No foreign-key cascade should erase request history on target deletion.

Persist final records only in MVP; a crash may lose in-flight calls and queued telemetry. Do not imply audit-grade completeness. Startup must distinguish malformed config (fatal) from observability unavailable (proxy runs degraded with a visible warning). Newer unsupported database schemas disable telemetry rather than corrupting data.

Shutdown: stop admission, allow existing requests a configurable grace period, cancel remaining upstream work, then drain telemetry within the bounded drain deadline and close SQLite. Report abandoned telemetry counts. Do not wait indefinitely.

## 9. Privacy and security

- Bind to `127.0.0.1` by default. Any non-loopback binding requires explicit configuration, authentication, and documented TLS termination; this remains a local single-user tool, not a hardened multi-tenant service.
- Authentication is enabled by default. Native API clients use bearer/per-origin tokens; browser management uses username/password sessions or management-token recovery. The owner may disable local authentication in Settings as specified in the user revision. Tokens must have enough randomness, compare safely, and never appear in URLs. Never forward local credentials to providers.
- UI users verify the startup management token for first-run setup, then use an optional username and password for subsequent sign-ins. Use HttpOnly SameSite browser sessions; recovery tokens are submitted only to the same-origin recovery endpoint and cleared from form state. No query-string bootstrap token, localStorage credential persistence, or third-party analytics. Serve assets with restrictive content security policy and same-origin API access.
- Validate Host and browser Origin against configured local addresses; disable cross-origin access by default. Reject cross-site state-changing requests. Test DNS rebinding and CSRF defenses; loopback alone does not make management APIs safe.
- Metadata only: no prompts, response text, tool arguments/results, raw headers, bodies, raw query values, or raw provider error messages in SQLite or logs. Parsing in memory for routing/usage is not permission to persist payloads. Payload capture is deferred, including debug mode.
- Upstream secrets use environment-variable references for MVP; never embed raw secrets in YAML, generated examples, SQLite, URLs, logs, or UI responses. Return only secret-reference names and configured/missing status. A later keychain adapter is optional, not a prerequisite.
- Restrict config/data/database permissions to the current OS user (0700 directory / 0600 files on POSIX; equivalent ACLs on Windows). Protect backups similarly. No claim of SQLite encryption at rest; rely on OS account and disk protections and document the limitation.
- Restrict upstream destinations to configured target URLs. Disable automatic upstream redirects to prevent credential leakage. Allow explicit loopback HTTP for local providers, preserve HTTPS certificate validation, and do not offer an insecure “skip verification” default.
- Bound request admission, request size, metadata strings, stream observation buffers, concurrency, timeouts, and query limits. Do not expose credentials in health endpoints or raw error stacks. Render all metadata as text in the UI.

The local OS account and configured upstream providers are trusted. aiport cannot protect secrets from malware running as the same account, or guarantee privacy after forwarding to a provider. Origin/model/session metadata itself may be sensitive; retention applies to it too.

## 10. Embedded dashboard and local API

Build only three primary areas:

1. **Overview:** compact metrics, one traffic chart, provider distribution, recent requests, and telemetry health/coverage indicator.
2. **Activity:** cursor-paginated newest-first table; filters for time, origin, target, provider, model, status/outcome; metadata detail view with timing and routing decisions.
3. **Settings:** target configuration, defaults/aliases, origins, secret-reference status, local endpoint instructions, retention, and explicit connection test.

Use a compact neutral UI, dark default with light option, accessible focus/contrast, keyboard navigation, responsive tables, and clear empty/error/degraded states. Request details never imply that prompts are available. Display timestamps in the user's selected/local timezone, with UTC available; store UTC.

Proposed management API prefix: `/api/v1/`, separate from native `/v1/` inference routes. Provide requests list/detail, aggregate stats, redacted config read/update, target tests, and health. Authenticate all metadata/config endpoints. Request list uses a stable `(started_at_ms, id)` cursor; cap page size (proposed 100). Detail returns 404 for unavailable/dropped records. Validate filter values and use parameterized queries. No WebSocket/live dashboard infrastructure: modest polling with cancellation is enough.

## 11. Repository and OSS distribution

```text
cmd/aiport/        # the binary: init, serve, agent, task and the other commands
internal/          # every package is private to the module
  config/          # aiport.yaml, credentials, Levels, live Changes
  server/          # HTTP API, dashboard, MCP endpoint, pairing
  proxy/           # forwarding to providers per protocol
  decide/          # Issues, Suggestions and Change requests
  catalog/         # what each Application can use
  store/           # SQLite request history and state
  telemetry/       # request metadata collection
  ...              # one package per concern; see `go list ./...`
  */testdata/      # synthetic, secret-free fixtures next to the package that reads them
web/               # React/TypeScript dashboard sources (pnpm, Vite)
internal/dashboard/dist/  # the built dashboard, committed and embedded in the binary
vendor/            # vendored Go dependencies, for offline builds
docs/              # manual, configuration, connect, this design; docs/images/ for pictures
```

Use small packages. Add interfaces only when a real consumer/test seam requires one. Choose a SQLite driver after testing cross-platform packaging; document CGO/native-library consequences. Do not call a binary self-contained if it requires an undeclared runtime library.

aiport is licensed Apache-2.0 (`LICENSE`, `NOTICE`); dependency licenses are listed in `THIRD_PARTY.md` and bundled under `licenses/`. Contributions are signed off under the DCO (see `CONTRIBUTING.md`); there is no CLA.

Required repository documents:

- `README.md`: purpose, install/run quickstart, one verified coding-agent setup, local provider example, supported-vs-planned table, privacy defaults, troubleshooting.
- `CONTRIBUTING.md`: Go/frontend setup, checks, fixture conventions, small provider contribution guide, PR expectations.
- `SECURITY.md`: private vulnerability reporting route, supported release policy, threat model, and secret-handling guidance. Do not invent contact addresses.
- `docs/compatibility.md`: exact endpoint/client/provider/server/model versions, streaming/tools/usage results, test date, limitations, evidence, first supporting release.
- `docs/configuration.md`: authoritative schema, defaults/limits, origin attribution, base URL rules, authentication, secret references, data/retention paths.
- `docs/releases.md` and `CHANGELOG.md`: version/tag process, build inputs, verification/checksums, upgrade/migration/rollback notes.

Initial release target matrix: macOS arm64/amd64, Linux arm64/amd64, Windows amd64. An unbuilt or untested platform remains planned. Tagged CI builds MUST pin dependencies/toolchains, run applicable checks, embed UI assets and version information, publish checksums, and document build reproduction. Report whether bit-for-bit reproducibility was actually verified. Signing SHOULD be added where feasible, without claiming signatures that do not exist.

Standalone archives are required. Homebrew installation and a small Docker image are optional later distribution conveniences; Docker is never required to run the MVP. No automatic updates or remote telemetry in MVP.

## 12. Implementation milestones and exit gates

| Milestone | Scope | Exit gate |
|---|---|---|
| M0 — Repository foundation | Inspect existing code/license; config contract; CLI; secret handling; synthetic fixtures; dependency/build choice | `aiport serve` loads validated config, binds loopback, authenticates, fails clearly on invalid configuration |
| M1 — Faithful proxy | OpenAI Chat streaming/non-streaming, fixed target, cancellation, bounded limits, native errors/tools | Direct-vs-proxy fixtures and one recorded real client integration pass; no telemetry dependency |
| M2 — Metadata pipeline | Origin attribution, timing/usage, bounded queue, SQLite WAL/migrations/retention, query API | Queue/database fault tests show uninterrupted forwarding and visible loss; metadata privacy tests pass |
| M3 — Embedded UI | Overview, Activity, Settings, protected management API, atomic config edits | A single executable serves usable UI and gateway with no runtime frontend/database service |
| M4 — MVP compatibility | Native Responses and Anthropic Messages; Ollama and llama.cpp tested Chat subset; Meta Model API and Bedrock | Per-combination fixtures + explicit live smoke evidence; every advertised feature has evidence and limitations |
| M5 — OSS release | License/contributor/security/release docs, platform archives, checksums, measured resource/performance report | Clean-install test and all applicable acceptance gates pass; matrix names exact released support |

M1–M3 may be labeled a developer preview with only verified combinations. The user revision brings Meta Model API and Bedrock into M4. Complete forwarding correctness before extending the protocol matrix; do not label configuration presets as tested support.

## 13. Acceptance test specification

Use a deterministic local upstream fixture server for protocol/failure tests; paid live tests are opt-in and never required for routine contributor CI. Live smoke evidence is required before publishing a provider combination as supported. Use synthetic payloads; sanitize any retained live evidence.

| ID | Test | Pass condition |
|---|---|---|
| A01 | Normal native requests and unknown fields | Upstream receives exact request bytes absent alias/auth/HTTP-required changes; client receives equivalent status, relevant headers, and body |
| A02 | SSE split across arbitrary reads, comments, multiline data, terminal events | Event bytes/order remain intact; first content arrives before upstream completion; no full-stream buffering |
| A03 | Fragmented/parallel tool calls and Responses/Anthropic event types | All fragments/IDs/native events survive; client can reconstruct the same calls as direct upstream |
| A04 | Provider 4xx/5xx and error body | Original status/body/relevant headers preserved; no key/body leakage in local logs or DB |
| A05 | Client cancellation and slow/disconnected clients | Upstream canceled promptly; bounded buffers/goroutines; other requests continue; cancellation outcome recorded if telemetry succeeds |
| A06 | Slow headers, idle stream, healthy long stream | Configured phase/idle deadlines work; ongoing traffic resets idle timeout; no fixed short total timeout truncates valid work |
| A07 | Malformed/unknown/oversized stream event | Raw forwarding unaffected; observer state bounded; unavailable usage/TTFT marked unknown/partial; premature EOF recorded incomplete where detectable |
| A08 | Locked/unavailable DB, disk-full simulation | Proxy succeeds independently; bounded retries and visible failed-write counters; no unbounded backlog |
| A09 | Saturated telemetry queue | Drop policy occurs immediately; loss counter/degraded UI visible; request path does not wait for writer |
| A10 | Concurrent streams | No cross-request byte/metadata mixing; bounded concurrency; overload rejected predictably |
| A11 | Origin precedence/spoofing | Valid origin credential wins; arbitrary headers cannot impersonate trusted origin; unknown is explicit; internal headers stripped upstream |
| A12 | Usage/TTFT/cost correctness | Heartbeats/roles excluded from TTFT; cumulative usage not summed twice; missing usage/prices remain null; pricing version and aggregate coverage present |
| A13 | Security and privacy | Canary secrets/payloads absent from logs/DB/API; browser cross-origin/Host attacks rejected; gateway token never reaches upstream; redirects do not leak credentials |
| A14 | Fixed routing/aliases/config edits | Deterministic target choice; incompatible protocol rejected before dispatch; alias changes only model semantics; failed edits preserve valid config; no retry/fallback |
| A15 | Retention, migration, shutdown/restart | Expired rows deleted in bounded batches; migration preserves data; unsupported schema degrades safely; bounded drain; dropped/in-flight crash limits documented |
| A16 | UI filters and aggregation | List/detail/group totals use same UTC range/filter logic; pagination stable; null metrics/empty states/telemetry loss displayed accurately |
| A17 | Clean installation and embedded UI | Advertised archive runs without Go/Node/Docker/database server; serves assets offline; secret references work; permissions and version output correct |
| A18 | Advertised live combinations | Each listed provider/client/version passes streaming, tools where claimed, native errors, cancellation and usage checks; absent features labeled untested/unsupported |

Protocol comparison allows HTTP-required hop/header changes, credentials replacement, explicit model aliases, and transport chunk resegmentation only. Never accept changed tool semantics or missing native events as a cosmetic difference.

Performance target: **p95 added time-to-first-byte overhead under 10 ms** against a direct local fixture upstream on a documented reference machine, excluding upstream service time. This is a target, not an achieved guarantee. Record client count, payload/event sizes, OS/hardware, build, sample count, and database state. Compare direct upstream, proxy with telemetry off, and proxy with telemetry on; report p50/p95/p99, CPU, RSS, throughput, and binary size. Use matched upstream behavior and enough samples to make the percentile meaningful. Streamed TTFT gets a separate measurement.

Demonstrate bounded queue/observer memory under sustained overload and no persistent goroutine growth after cancellation. Establish baseline memory/binary budgets from a measured representative build rather than claiming arbitrary numbers. If the latency target fails, document evidence and resolve or explicitly revise the gate before release; do not advertise unmeasured performance.

## 14. Optional work after MVP

Each item needs its own approved scope and acceptance criteria:

AWS Bedrock has moved into the requested initial provider set. Its implementation
requires a verified choice of AWS endpoint and supported authentication, explicit
region/model identity, native streaming and cancellation tests, and usage evidence.
Do not assume a request format or authentication mechanism from another provider.
- **Fallback/retries:** explicit opt-in policies, eligible failure classification, replay/idempotency analysis, strict attempt limits/backoff, per-attempt attribution, and no replay after response commitment without an established safety contract.
- **Routing rules:** start with explicit user rules. Cost/latency/availability do not establish model quality; quality-based optimization needs evaluation evidence.
- **HelixDB:** only after demonstrated graph queries justify storage and deployment complexity.
- **Keychain, Homebrew, Docker, payload capture:** independent proposals; payload capture additionally needs explicit consent, redaction, limits, and retention design.

Do not implement a generic policy engine, plugin system, or second database now to prepare for these possibilities.

## 15. Instructions to the coding agent

1. Inspect the actual repository and its instructions, current license, toolchain, and tests. Report which requirements already exist with evidence; this document alone proves none.
2. Preserve the slim product boundary. Resolve routine implementation choices within it; raise scope-changing contradictions before building them.
3. Implement M0 then M1, validating streaming/tool fidelity before telemetry and UI. Keep commits/reviews small if repository workflow permits.
4. Keep request forwarding independent of observation, persistence, pricing, and dashboard reads. Bound every observer/queue/retry path.
5. Maintain a checklist against milestone gates and A01–A18. Label each compatibility claim planned, experimental, or supported with test/release evidence.
6. Update configuration/docs alongside actual behavior; do not present proposed sample commands as tested until run.
7. Stop at the agreed milestone. Do not add deferred providers, smarter routing, unrelated cleanup, or optional release channels automatically.

**Done condition:** a clean-install, single-executable local gateway with the advertised native protocols and tested local-provider subset, faithful streaming, safe metadata-only observability, a usable embedded UI, documented OSS contribution/release paths, and passing evidence for every advertised MVP capability.
