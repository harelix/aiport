# Configuration contract

`aiport.yaml` is authoritative. The dashboard and the Owner API change it as
revertable Changes, applied live without a restart. To edit it by hand: stop,
edit, run `aiport check`, and restart. Failed startup does not edit the file.
The file must be regular, at most 1 MiB, and private (0600 or stricter on POSIX).
Windows file ACL enforcement is not implemented or verified.

YAML has exactly one document. Unknown keys, duplicate mapping keys, duplicate
IDs, missing references, and incompatible default protocols are rejected.

### Providers

Built-in presets keep configuration generic. Name the provider and, where the
provider is regional, a region; aiport derives the endpoint when it loads the file:

```yaml
targets:
  - id: bedrock
    provider: bedrock
    region: eu-west-1      # endpoint: https://bedrock-runtime.eu-west-1.amazonaws.com/openai/v1
  - id: anthropic
    provider: anthropic
  - id: ollama
    provider: ollama
```

| Provider | Group | Protocol | Endpoint template | Key variable |
| --- | --- | --- | --- | --- |
| `openai` | Remote API | openai | `https://api.openai.com/v1` | `OPENAI_API_KEY` |
| `anthropic` | Remote API | anthropic | `https://api.anthropic.com/v1` | `ANTHROPIC_API_KEY` |
| `bedrock` | Remote API | openai (default) or anthropic | see Bedrock below | `AWS_BEARER_TOKEN_BEDROCK` |
| `meta` | Remote API | openai | `https://api.meta.ai/v1` | `MODEL_API_KEY` |
| `ollama` | Local | openai | `http://127.0.0.1:11434/v1` | none |
| `llamacpp` | Local | openai | `http://127.0.0.1:8080/v1` | none |

#### Amazon Bedrock

Bedrock is one provider with three settings: `region`, `endpoint` and `protocol`.
All four combinations use a Bedrock API key (sent as `Authorization: Bearer` for
the OpenAI protocol and `x-api-key` for the Anthropic protocol):

| `endpoint` | `protocol: openai` | `protocol: anthropic` (Claude Code) |
| --- | --- | --- |
| `runtime` (default, AWS-recommended) | `https://bedrock-runtime.{region}.amazonaws.com/openai/v1` | `https://bedrock-runtime.{region}.amazonaws.com/anthropic/v1` |
| `mantle` (server-side tools, Projects) | `https://bedrock-mantle.{region}.api.aws/v1` | `https://bedrock-mantle.{region}.api.aws/anthropic/v1` |

```yaml
  - id: bedrock-claude
    provider: bedrock
    region: eu-central-1
    endpoint: mantle
    protocol: anthropic
```

Connections can also be added, edited, removed and made default in the dashboard
(Settings → Providers, Settings → Routing). The change is written into this
file in place (other keys and comments kept), the whole file is revalidated, and
new requests use it immediately. Several Bedrock connections (different regions,
endpoints or protocols) can coexist under different ids.

Any field can be overridden. A `base_url` may itself use `{region}` (for example
the few Mantle models served under `https://bedrock-mantle.{region}.api.aws/openai/v1`);
`base_url` and `endpoint` cannot be combined. A `region` with
no `{region}` in the URL, or a `{region}` with no region, is rejected rather than
ignored. Other providers are custom: give `protocol` and `base_url`. The dashboard
groups connections as Local (loopback endpoints) or Remote APIs.

| Key | Contract/default |
| --- | --- |
| `version` | Required; exactly `1` |
| `listen` | `127.0.0.1:8787`; literal IPv4/IPv6 loopback and port 1–65535 only |
| `hostnames` | Optional list of local names aiport also answers to on the listen port, for example `[aiport.localhost]`. Each must be lowercase, `localhost` or end in `.localhost`, valid DNS label syntax (letters, digits, inner hyphens), without a port, listed once. aiport accepts `Host`/`Origin` `<name>:<listen port>` next to the listen address and refuses every other name. The dashboard can be opened under these names, but Applications and coding agents are always given the numeric address (`/_connect`, briefs, tokens, `.aiport`): some resolvers and sandbox proxies refuse `*.localhost`. The socket stays on the `listen` IP. A change made through aiport takes effect without a restart; a hand edit needs one |
| `auth.token_env` | Required name of the global management/local bearer token environment variable |
| `targets` | List of `id`, `provider`, and for built-in providers optionally nothing else. `protocol`, `base_url` and `api_key_env` default from the provider preset (required for custom providers). `region` is required by regional providers (Bedrock) and fills `{region}`. `endpoint` picks a provider API surface (Bedrock: `runtime` default, or `mantle`). Optional `disabled` (false) |
| `origins` | List of `id`, `name`, `token_env`, optional `target`, optional `aliases` (`alias`, `target`, `model`, optional `effort`): models only this client can ask for, checked before global aliases and allowed to pick another connection. An entry whose `alias` is a Level is this client's own mapping of that Level on its `target`'s protocol, at most one per Level and protocol |
| `defaults` | Optional `openai_target` and `anthropic_target`; matching target protocols required |
| `aliases` | List of `target`, `alias`, `model`, optional `effort`; unique alias per target; model strings 1–256 bytes. An entry whose `alias` is a Level (`fast`, `balanced`, `best`) is the shared mapping of that Level on its `target`'s protocol, at most one per Level and protocol |
| `aliases[].effort`, `origins[].aliases[].effort` | Optional `{low, medium, high}`, each a model on the same `target`; the model for a call of that effort (`X-Aiport-Effort`, else the provider's own effort field, else `medium`). An effort left out uses `model` |
| `telemetry.enabled` | true; request metadata is saved to `<config>.db` (SQLite) |
| `telemetry.retention_days` | 7; range 1–3650; also how long decided Suggestions and Change requests, and resolved or dismissed Issues, are kept (dismissed ones at least 7 days, the quiet period), and how long Catalog changes are kept; Changes themselves are never pruned |
| `telemetry.queue_capacity` | 4096; range 1–65536 |
| `telemetry.batch_size` | 100; range 1–queue capacity |
| `telemetry.flush_interval_ms` | 250; range 1–60000 |
| `telemetry.shutdown_drain_ms` | 2000; range 1–60000; reserved for M2 telemetry |
| `telemetry.max_event_bytes` | 16384; range 1024–65536 |
| `adapting.disabled` | false; `true` turns the whole adapting loop off: no Suggestions are detected, no Issues are opened and Changes shows nothing new |
| `adapting.interval_seconds` | 60; range 5–86400; how often a background pass looks at request history (never during a request) |
| `adapting.unknown_model.count` | 5; range 1–1000000; upstream 404s for one model from one Application that open an Issue (with an Alias for that Application as the aiport fix); also how many new 404s reopen a resolved one |
| `adapting.unknown_model.window_minutes` | 60; range 1–10080; how far back those 404s are counted |
| `adapting.unknown_token.count` | 10; range 1–1000000; requests refused for a missing or unknown token that lead to an informational Suggestion to pair the Application (only counts per minute are kept, never the token) |
| `adapting.unknown_token.window_minutes` | 60; range 1–10080; how far back those refusals are counted |
| `adapting.failing_connection.count` | 5; range 1–1000000; failed calls (gateway errors, timeouts and upstream 5xx; never cancellations or refusals) from one Application to one Connection that lead to a Suggestion to bind the Application to a working Connection |
| `adapting.failing_connection.window_minutes` | 60; range 1–10080; how far back those failures are counted |
| `adapting.level_unmapped.count` | 5; range 1–1000000; calls from one Application for one Level on one protocol refused with `level_unmapped` that open an Issue (mapping the Level for that Application as the aiport fix) |
| `adapting.level_unmapped.window_minutes` | 60; range 1–10080; how far back those refusals are counted |

A Levels and per-effort example (anonymous names):

```yaml
aliases:
  - {target: local, alias: fast, model: 'llama3:8b'}
  - {target: hosted, alias: balanced, model: mid-model, effort: {low: small-model, high: big-model}}
origins:
  - id: docs-bot
    name: Docs Bot
    token_env: AIPORT_APP_DOCS_BOT_TOKEN
    aliases:
      - {alias: best, target: local, model: 'llama3:70b'}   # its own best
      - {alias: summarize, target: local, model: 'llama3:8b', effort: {high: 'llama3:70b'}}
```

A shared task name normally applies only to calls that already go to its
Connection. `everywhere: true` makes it usable by every Application, whichever
Connection its calls otherwise go to: the name then routes to its own
Connection, like a Level (one `everywhere` entry per name; an Application's
own task name of the same name still comes first). `reasoning: low|medium|high`
makes aiport send that reasoning effort to the provider on every call for the
name whose body sets none of its own (`reasoning_effort`, `reasoning.effort` or
`output_config.effort`), whatever `X-Aiport-Effort` says. The Owner adds,
changes and removes shared task names with `PUT` and `DELETE
/api/v1/task-names/{name}` (`{connection, model, effort?, everywhere?,
reasoning?}`), each a revertable Change.

```yaml
aliases:
  - {target: hosted-cheap, alias: classify, model: small-model, everywhere: true, reasoning: high}
```

Without `reasoning`, an effort stated only with `X-Aiport-Effort` is also sent
to the provider in its own field when aiport chose the model (a Level, task
name or Alias) and the body has none.

Upgrade note: an Alias entry named `fast`, `balanced` or `best`, shared or an
Application's own, is a Level mapping on its Connection's protocol. A shared
one chooses the Connection for every Application's calls for that name on
that protocol (a plain shared Alias only renames a model on the Connection a
call already goes to); an Application's own one applies on that protocol
only. `aiport check` and startup print one notice per such entry; to keep a
plain Alias, rename its `alias` to a name that is not a Level.
| `adapting.auto_fix` | false; `true` lets aiport apply an Issue's aiport fix itself, as a revertable Change made by "aiport (automatic)", only when exactly one Connection serves the model, the error reached its threshold and the fix is an Alias for that one Application. A fix the Owner reverted (or rejected) is never applied automatically again for that Application and cause. Also switchable on the Changes page |
| `adapting.explain` | none; `{connection, model}`: the configured Connection and a model it serves that write an Issue's explanation (Explain) from its evidence alone. Without it Explain is unavailable. Also set in Settings → Explanations |

IDs and provider labels contain 1–64 ASCII alphanumerics, underscores or hyphens,
starting with an alphanumeric. Protocol is `openai` or `anthropic`. Provider is a
configured label, not verified upstream identity. Origin names contain 1–128 bytes.
Disabled targets may remain referenced; eventual forwarding must reject new calls.

Base URLs are fixed HTTP(S) API prefixes, without user information, query strings,
or fragments. HTTP is allowed only for literal loopback IPs or `localhost`.
Remote destinations require HTTPS. Non-loopback serving is unavailable;
remote deployment and TLS termination are not implemented.

Environment references use shell identifier syntax. All configured local tokens
are required, distinct, printable ASCII without spaces, and 32–512 bytes. Length
cannot establish randomness: generate tokens with `aiport token`. Authentication
uses fixed-length SHA-256 digests and constant-time comparisons. Origin tokens do
not grant management access; arbitrary headers never grant identity.

Provider keys can be entered in the dashboard (Settings → Providers).
They are written atomically to the private credential file next to the config
(`aiport.yaml.env`, 0600), apply to new requests immediately, and take precedence
over the same-named environment variable on later starts. Removing a saved key
falls back to the server's environment value, if any. Keys are write-only: the API
reports only `saved`, `environment`, `missing` or `not_needed`.
An enabled target without a key no longer blocks startup; `check`/`serve` warn and
its requests return 503 `upstream_key_missing` until a key is set. Upstream and
local credentials must differ. Raw keys are never stored in YAML. Target/API credentials do not
establish provider compatibility.

The server limits headers to 16 KiB (subject to Go HTTP parser overhead), header
read time to 5 seconds, idle keep-alive connections to 60 seconds, and shutdown
grace to 5 seconds. Inference admission/body/stream limits are defined below. No data directory or SQLite files are created.

## Experimental forwarding and onboarding update

The forwarding path uses these configurable bounds:

| `proxy` key | Default | Allowed range |
| --- | --- | --- |
| `max_concurrent` | 32 | 1–1024 |
| `max_request_bytes` | 8388608 | 1–67108864 |
| `connect_timeout_ms` | 10000 | 1–300000 |
| `response_header_timeout_ms` | 30000 | 1–600000 |
| `stream_idle_timeout_ms` | 120000 | 1–3600000 |
| `request_read_timeout_ms` | 30000 | 1–600000 |

Limits are independent of telemetry. Request bytes are read within the admission
limit before dispatch so oversized bodies return 413 before any upstream call.
Response bodies are copied in bounded chunks and flushed; they are never buffered
in full. Healthy streams have no fixed total timeout. A stalled upstream or slow
downstream is terminated at the idle/write deadline. After response commitment,
transport failures abort the response; no synthetic provider events are appended.
Clients may observe a truncated response. No outcome persistence exists yet.

Selection: a Level (the origin's own mapping on the request's protocol, then
the shared one; neither is 400 `level_unmapped`), then the origin's own alias,
then the origin's target, then the protocol default; per-effort models of the
matched Level or alias then pick the model. Missing targets
return 503, disabled targets 503, protocol mismatch 400, concurrency overflow 429,
upstream connection failure 502, and pre-commit upstream timeout 504. No fallback
or retries. A new upstream connection is used for every request to avoid implicit
transport replay; connection pooling and performance tuning remain unmeasured.

Raw query semantics are retained without logging. The base URL prefix is appended
once (`/v1/chat/completions` appends `/chat/completions`, for example). Native
unknown request fields and body bytes pass unchanged. A client's own alias for the requested model is applied first and may select
another connection; otherwise the client's bound connection (or the protocol
default) and its global aliases apply. Explicit model aliases
replace only the top-level model value while preserving other bytes; ambiguous
model fields are rejected when aliases are configured. No generic required-model
validation is imposed. Native validation/errors remain upstream responsibilities.

Local bearer auth is replaced with the selected upstream secret: bearer for
OpenAI-protocol targets, `x-api-key` for Anthropic-protocol targets. Messages also
accept local tokens in `x-api-key` when Authorization is absent. Native Anthropic
version/beta headers are forwarded without inventing a version value. Hop headers,
cookies, incoming auth, `x-aiport-*`, forwarded-client and browser-origin headers
are stripped. Upstream redirects are returned without following them; cookies and
authentication response headers are stripped. Compression is not transformed.

`aiport init` supports comma-separated providers/agents, private exclusive setup
files and a companion shell `.env` file with generated local credentials only.
It proposes a Level mapping from the chosen providers and asks for each
model, mapping nothing unless confirmed; `--levels
fast=local/llama3:8b,best=hosted/big-model` (`LEVEL=CONNECTION/MODEL`, the
Connection one of the chosen provider ids) maps them without a terminal, and
`--no-levels` skips the step.
It configures endpoints and references; it does not contact providers or test keys.
`aiport agent` performs per-launch configuration without changing global client
settings. Agent target selection follows the configured origin; edit YAML to
change it. `--dry-run` checks the launch setup without starting an agent.

## Request metadata history and embedded dashboard

When telemetry is enabled, each completed inference handler offers one metadata
record to a queue of `queue_capacity` without waiting. A single consumer keeps the
latest 500 records in memory and writes every record to `<config>.db` (SQLite, WAL,
0600) in transactions of up to `batch_size`, at least every `flush_interval_ms`.
Rows older than `retention_days` are deleted in small batches at startup and hourly,
off the forwarding path. The queue drains and the last batch is written within
`shutdown_drain_ms`. Dropped records and failed writes are counted in health.
`max_event_bytes` is reserved and not enforced yet.

If the database cannot be opened, startup prints the reason and history is kept
for that session only (the banner says so). If it was written by a newer aiport,
recording is turned off instead. WAL, backups and free pages can keep deleted
metadata; deletion is not a secure-erasure guarantee.

Records contain gateway UUID, UTC start, origin, selected target/provider/protocol,
normalized method/route, bounded requested model, its effort and the effort's
source (`header`, `field` or `default`), the model sent upstream
(`resolved_model`, empty when the call was not forwarded), status/outcome,
monotonic latency and first observed body-byte time. No prompts, body fragments, tool arguments,
raw headers, query values or provider errors are retained. Usage and pricing are
unobserved/unknown. There is no TTFT metric; TTFB is not relabeled as TTFT.

Authenticated GET management APIs: `/api/v1/config` (redacted references/status,
read-only), `/api/v1/requests`, `/api/v1/requests/{id}`, `/api/v1/stats`, and health.
List/stats share `hours` (1–168), origin/target/provider/outcome/model filters
(`model` is the name the Application sent), and optional
exact UTC `from_ms`/`to_ms` bounds capped at seven days. Stable descending
`(started_at_ms, id)` cursors are base64url encoded; list size defaults to 50, max
100. Detail returns 404 if absent/evicted. Counts cover retained records only.
P95 uses nearest rank; latency sample counts are returned. Error rate excludes
pre-admission rejections and counts upstream/gateway/incomplete/timeout outcomes;
cancellations and rejections are shown separately. Cost/token totals are null.

Model listings (management access and the `X-Aiport-CSRF: 1` header; both call
providers with the free `GET {base}/models`, never a generating request):
`GET /api/v1/targets/{id}/models[?refresh=1]` returns `{connection, listed,
state, models, listed_at_ms}` (model ids only; `listed: false` means unknown);
`GET /api/v1/serving?protocol=openai|anthropic&model=…[&refresh=1]` returns
`{model, protocol, answer, serving, unknown}` over the enabled Connections of
that protocol, where `answer` is `one`, `none`, `several` or `unknown` (a
Connection whose listing failed is in `unknown`). Lists are kept in memory
for two minutes; a failed listing is not kept.

Catalog (Application token; the management token gets `403
application_token_required`; works with `adapting.disabled`): `GET /v1/models`
lists the calling Application's task names, then the models its reached
Connections list, from the same cached lists (OpenAI list shape; Anthropic
shape with `anthropic-version`). Every answer to an Application token carries
`X-Aiport-Catalog: <version>` (12 hex characters of SHA-256 over its binding,
followed protocol defaults, task names and their targets; model lists do not
change it). `GET /v1/aiport/catalog` returns `{catalog, changes}`; the capabilities answer
carries `catalog: {version, changed_at_ms}`, where `changed_at_ms` is `null`
when no Catalog change is kept for the Application. Catalog
changes are worked out from every configuration Change and revert, for every
Application and the built-in `aiport` one, and kept in the `catalog_changes`
table of `<config>.db` for `telemetry.retention_days` (pruned at startup and
hourly; without a database they are kept in memory for this session, also no
longer than that). A `catalog_changed` Issue opens only for an Application
that used what the Change took away within request history.

Levels and per-effort models (management access; writes need the
`X-Aiport-CSRF: 1` header; each write is one revertable Change and a Catalog
change): `GET /api/v1/levels` returns `{shared, applications, connections}`;
`PUT /api/v1/levels/{protocol}/{level}` with `{application?, connection,
model, effort?}` maps a Level (`application` left out: the shared mapping;
`effort` left out keeps the per-effort models while the Connection stays the
same and drops them when it changes); `DELETE` on the same path with
`{application?}` removes the mapping; `PUT /api/v1/aliases/effort` with
`{application?, alias, connection?, effort}` sets the per-effort models of
any Alias (`connection` required for shared Aliases and Levels).

Routing map (management access): `GET /api/v1/routing` returns, per
Application, every name it can send (task names, Levels, any other model
name) with the Connection and model that serve it, built from the same
Catalog as its `GET /v1/models`; `?application=<id>` returns one Application
with its recent Catalog `changes`.

Guidance versions: `GET /v1/aiport/guidance` and `POST
/v1/aiport/guidance/done` (Application token; MCP `aiport_guidance` and
`aiport_guidance_done`; work with `adapting.disabled`) read and confirm the
guidance version of the calling Application; `GET /api/v1/guidance`
(management access) lists every Application's `confirmed`,
`confirmed_at_ms`, `behind` and `sends_levels` (a hint only), with the
counts `behind` (every Application not on the current version) and
`never_confirmed` (those of them that never confirmed any) and the refresh
instruction. Confirmations are kept in the `guidance` table of
`<config>.db` (in memory for the session without a database).

Removing a Connection (management access): `GET /api/v1/targets/{id}/dependents`
(or `POST` with `X-Aiport-CSRF: 1` and a draft plan) returns `{connection, dependents, applications,
problem?}`, each dependent with its suggested replacement and `applications`
the Applications whose Catalog the plan changes; `DELETE /api/v1/targets/{id}`
with `{"plan": {"<dependent id>": "<Connection id or empty>"}}` re-points every
dependent and removes the Connection as one Change (`{ok, change_id,
applications}`). Dependent ids are `default:<protocol>`, `binding:<app>`,
`alias:<name>`, `application_alias:<app>:<name>`, `level:<level>:<protocol>`,
`application_level:<app>:<level>:<protocol>`, `explain`, `agent_runs` and
`evaluation`; an empty Connection leaves the item without one (default unset,
binding or Alias removed, the feature turned off). An Alias or Level mapping
moved to another Connection keeps its model and drops its per-effort models
(`effort`), which name models of the removed Connection. Without a plan, a
Connection something uses is refused (`400 invalid_connection`, naming what
uses it).

The public local root serves embedded login assets under a restrictive same-origin
CSP. The UI sends its in-memory management token as bearer auth; no localStorage,
query bootstrap token, external font, CDN or analytics is used. Config editing,
web-first provider onboarding, and connection testing remain future work.


## Username/password, recovery and authentication settings

New `auth` keys: `username` (optional, at most 64 bytes), `password_hash` (generated
by the server, not returned by JSON APIs), and `disabled` (false by default).
Keep `token_env`; it identifies the management/recovery credential. Never manually
put a plaintext password in YAML. Hash encoding is
`pbkdf2-sha256$600000$<16-byte-salt-base64>$<32-byte-key-base64>`.

First-run browser setup requires the management token before account creation.
A password and optional username are then saved using atomic YAML replacement;
unrelated config keys/comments remain. Configuration symlinks are not editable
through this auth API. Save/validation failure leaves runtime authentication
unchanged. Normal password sign-in uses the saved account after server restarts.

Public GET `/api/v1/auth/status` exposes only setup/authentication flags and the
recovery environment reference name. POST routes under `/api/v1/auth/` are
`login`, `recover`, `logout`, and `settings`. JSON mutation requests require
`X-Aiport-CSRF: 1`, with Host/Origin/fetch-site checks still enforced. Recovery
requires a valid management bearer; agent credentials never authorize it.

Browser cookies are opaque random 256-bit identifiers, stored by the server only
as digests, scoped to `/api/v1/`, HttpOnly and SameSite Strict. Lifetime is 12 hours;
server restart, sign-out, and auth changes invalidate sessions. Cookies use Secure
when served over TLS; the current default listener is loopback HTTP. Browser
cookies authorize management only, never native inference. Password verification
has a global 10-attempt/minute bound and at most two expensive auth operations at
a time; at most 128 active browser sessions are retained.

`disabled: true` permits unauthenticated local management/inference. It retains
saved credentials and existing origin attribution where valid tokens are supplied.
Missing local token envs are tolerated while disabled; upstream secrets are still
required for configured enabled providers. Ensure local references are available
again before restarting with auth enabled. Turning auth back on requires a saved
or newly set password. All browser cross-site protections stay active.

Settings auth changes apply immediately. Provider settings remain read-only.
The source references for hashing/session decisions are Go `crypto/pbkdf2`,
OWASP Password Storage and Session Management Cheat Sheets. Nord source palette:
https://www.nordtheme.com/docs/colors-and-palettes/. Exact checked pairs include
Nord6/Nord1 8.73:1, Nord4/Nord1 7.45:1, Nord0/Nord8 6.24:1, Nord3/Nord6 6.40:1,
and input-border Nord9/Nord1 3.74:1. These checks do not establish full UI
accessibility or visual/browser verification.

## Health check, errors and terminal output

```sh
curl -H "Authorization: Bearer $AIPORT_LOCAL_TOKEN" http://127.0.0.1:8787/api/v1/health
```

All API paths require local credentials (unless auth is disabled). Management
routes, health included, need the management token; Application tokens only
grant inference. Host must match a configured address. Unknown paths return
structured 404 errors; missing or disabled routes and admission failures
return redacted local errors with a UUID. Upstream statuses, bodies, content
types and request ids pass through unchanged.

A token already exported in your shell takes precedence over the one in the
private `.env` file next to the config. Terminal output uses colour only on a
TTY; `NO_COLOR`, redirected output and token output stay plain.

## Extensions

`extensions:` lists the programs (policies and guardrails) that take part in
deciding about Change requests: `name`, `hook` (`policy` or `guardrail`), an
absolute `path`, optional `args` and `timeout_ms` (default 5000, at most 60000).
Only listed programs run; failure, a crash or a timeout denies. See
[extensions.md](extensions.md).

## Upstream

`upstream.repository` (`owner/name` on GitHub) is where "Suggest to aiport"
and "Share upstream" drafts open. Empty means the repository aiport was built
with (`make build UPSTREAM=owner/name`), if any; with neither, the dashboard
still shows each draft to copy.

- `GET /api/v1/upstream` lists the drafts: `suggestions` (one per kind in
  demand, from `demand` in `GET /api/v1/decisions`) and `shares` (one per
  listed Extension), each with `title`, `body` and, when the repository is
  known, a GitHub `url` that opens the prefilled issue. `gh` says whether the
  gh CLI can file them.
- `POST /api/v1/upstream/issues` `{"suggest": kind}` or `{"share": extension}`
  with `"confirm": true` runs `gh issue create` with the Owner's own gh login.
  Without `confirm` it is `400 confirmation_required` and nothing runs.

Drafts carry kinds, hooks and anonymous counts only: never configuration
values, tokens, paths, Application names, request text or Extension settings.
aiport stores no GitHub credential and sends nothing on its own; gh gets only
`PATH`, `HOME` and the variables it needs to find its own login.
