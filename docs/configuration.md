# Configuration contract

`aiport.yaml` is authoritative. No settings write API or reload exists yet; stop,
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
| `auth.token_env` | Required name of the global management/local bearer token environment variable |
| `targets` | List of `id`, `provider`, and for built-in providers optionally nothing else. `protocol`, `base_url` and `api_key_env` default from the provider preset (required for custom providers). `region` is required by regional providers (Bedrock) and fills `{region}`. `endpoint` picks a provider API surface (Bedrock: `runtime` default, or `mantle`). Optional `disabled` (false) |
| `origins` | List of `id`, `name`, `token_env`, optional `target`, optional `aliases` (`alias`, `target`, `model`): models only this client can ask for, checked before global aliases and allowed to pick another connection |
| `defaults` | Optional `openai_target` and `anthropic_target`; matching target protocols required |
| `aliases` | List of `target`, `alias`, `model`; unique alias per target; model strings 1–256 bytes |
| `telemetry.enabled` | true; request metadata is saved to `<config>.db` (SQLite) |
| `telemetry.retention_days` | 7; range 1–3650; also how long decided Suggestions and Change requests are kept (dismissed ones at least 7 days, the quiet period); Changes themselves are never pruned |
| `telemetry.queue_capacity` | 4096; range 1–65536 |
| `telemetry.batch_size` | 100; range 1–queue capacity |
| `telemetry.flush_interval_ms` | 250; range 1–60000 |
| `telemetry.shutdown_drain_ms` | 2000; range 1–60000; reserved for M2 telemetry |
| `telemetry.max_event_bytes` | 16384; range 1024–65536 |
| `adapting.disabled` | false; `true` turns the whole adapting loop off: no Suggestions are detected and Changes shows nothing new |
| `adapting.interval_seconds` | 60; range 5–86400; how often a background pass looks at request history (never during a request) |
| `adapting.unknown_model.count` | 5; range 1–1000000; upstream 404s for one model from one Application that lead to a Suggestion to add an Alias |
| `adapting.unknown_model.window_minutes` | 60; range 1–10080; how far back those 404s are counted |
| `adapting.unknown_token.count` | 10; range 1–1000000; requests refused for a missing or unknown token that lead to an informational Suggestion to pair the Application (only counts per minute are kept, never the token) |
| `adapting.unknown_token.window_minutes` | 60; range 1–10080; how far back those refusals are counted |
| `adapting.failing_connection.count` | 5; range 1–1000000; failed calls (gateway errors, timeouts and upstream 5xx; never cancellations or refusals) from one Application to one Connection that lead to a Suggestion to bind the Application to a working Connection |
| `adapting.failing_connection.window_minutes` | 60; range 1–10080; how far back those failures are counted |

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

Selection: authenticated origin target, then protocol default; missing targets
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
normalized method/route, bounded requested model, status/outcome, monotonic latency
and first observed body-byte time. No prompts, body fragments, tool arguments,
raw headers, query values or provider errors are retained. Usage and pricing are
unobserved/unknown. There is no TTFT metric; TTFB is not relabeled as TTFT.

Authenticated GET management APIs: `/api/v1/config` (redacted references/status,
read-only), `/api/v1/requests`, `/api/v1/requests/{id}`, `/api/v1/stats`, and health.
List/stats share `hours` (1–168), origin/target/provider/outcome filters, and optional
exact UTC `from_ms`/`to_ms` bounds capped at seven days. Stable descending
`(started_at_ms, id)` cursors are base64url encoded; list size defaults to 50, max
100. Detail returns 404 if absent/evicted. Counts cover retained records only.
P95 uses nearest rank; latency sample counts are returned. Error rate excludes
pre-admission rejections and counts upstream/gateway/incomplete/timeout outcomes;
cancellations and rejections are shown separately. Cost/token totals are null.

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
