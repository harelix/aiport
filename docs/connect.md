# aiport Connect

aiport Connect is how applications learn to talk to an aiport gateway. A running
gateway publishes it, without authentication, at:

- `GET /_connect`: JSON for programs
- `GET /_connect?format=markdown` (or `Accept: text/markdown`): a guide for people and coding agents. It opens with **Start here**: what a coding agent does, in order, and the one instruction the Owner pastes to it (also `agent_prompt` in the JSON). An agent handed only this URL has everything for both briefs.
- `GET /_connect?format=handover&for=application`: the **Application brief**, for the coding agent changing an application so its LLM calls go through aiport
- `GET /_connect?format=handover&for=agent`: the **coding-agent brief**, for coding agents working in that application's repository later
- `GET /_connect?format=handover`: both briefs, one after the other (any other `for` is `400 invalid_format`)

The document reflects the live gateway: base URL, which native protocols are
configured, whether authentication is required, request/response headers,
error codes and features. It never contains secrets, target URLs, key names or
model lists (an Application reads its own model names from its Catalog, below).
Host/Origin protections still apply.

Addresses in the document (`base_url`, `docs`, `agent_prompt`, briefs and
snippets) use the gateway's numeric listen address, with a wildcard listen
address given as `127.0.0.1` (by default `http://127.0.0.1:8787`). Applications
and their agents get the IP address even when the Owner set `hostnames`: some
resolvers and sandbox proxies do not resolve `*.localhost`, and the name may
resolve to `::1` first while aiport listens on IPv4. The dashboard can still be
opened under a configured hostname. aiport accepts
`Host` and `Origin` for the listen address and each configured hostname on
the listen port, and refuses everything else with `403 invalid_host` /
`invalid_origin`.

Each feature and header has a status. Only `available` items are implemented;
`planned` items (model allowlists, the `X-Aiport-Task`, `X-Aiport-Session`
and `X-Aiport-Version` request headers, and the `X-Aiport-Model`,
`X-Aiport-Route` and `X-Aiport-Source` response headers) are published early
so integrators can prepare.

## The communication standard

`GET /_connect?format=standard` is one markdown document, written for an
application's coding agent, with every rule an integration follows and a
checklist: the token in `.aiport`, the SDK pointed at aiport, Levels and task
names, `X-Aiport-Effort`, `X-Aiport-Module` (the module of the application the
call serves; required, recorded, never forwarded), a thumbs up and down on every
LLM answer shown to a person sent as `POST /v1/aiport/feedback`
`{"request_id": "...", "vote": "up"|"down"}`, the error fixes and the MCP tools.
The JSON document gives its address as `standard` and the prompt that hands it
to a coding agent as `standard_prompt`; Integrate shows both. The source is
`internal/connect/standard.go`.

## Pairing

An application without a token asks for one: `POST /v1/aiport/pair`
`{"name": "...", "description": "..."}` (public on loopback, rate-limited) returns
a short `code` and a secret `poll` path. The owner approves the code in the
dashboard's Integrate page (or the Applications table on Routing), optionally binding a connection; aiport adds
the application as a client, with its own token, without a restart. `GET` on the
poll path returns `pending`, `rejected`, `expired`, or `approved` with the token.
Unapproved requests expire after 30 minutes and survive a gateway restart (the
poll secret is stored only as a hash). One application gets one request: asking
again with the same name or repository returns the same code, and when an
Application with that name or repository already exists the Owner can give it a
new token instead of adding a second one. The token lives in the private
credential file as `AIPORT_APP_<ID>_TOKEN`; the application keeps it as
`AIPORT_APP_TOKEN=...` in a gitignored `.aiport` file at its repository root,
which both the application and its coding agent read. On Integrate the Owner
can also add an Application with a token directly, issue a **New token** (the
old one stops working) and remove an Application (a revertable Change).

Every refusal of an Application's token (`401 unauthorized`,
`403 application_token_required`) carries `error.fix`: `for`, `steps` (check
`.aiport`, else ask the Owner for **New token**), `retry` (retry the same call
once `.aiport` has the token; in Claude Code reconnect the MCP server with
`/mcp`) and `docs`, a link to the "If aiport refuses the token" section of
`/_connect?format=markdown`. The `X-Aiport-Hint` header says the same in one
line. The source of truth is
`internal/connect/connect.go`; bump `Version` on incompatible changes.

## The Catalog

An Application's Catalog is what it can use through aiport: its task names
(its own Aliases, then the shared Aliases of the Connections it reaches, and
shared task names set to `everywhere`, marked `everywhere: true`, which go to
their own Connection), the Levels mapped for it on each protocol, and the
models its Connections list, with a version. A task name may show a fixed
`reasoning` effort. The document describes it
as the `catalog` feature, the `X-Aiport-Catalog` response header and a
`GET /v1/models` endpoint under each protocol:

- `GET /v1/models` with the Application's own token (the management token gets
  `403 application_token_required`): the model names it can
  use on that protocol, in the OpenAI list shape, or the Anthropic one when the
  request carries `anthropic-version`. Task names come first (`owned_by:
  "aiport"`, or "(aiport task name)" in `display_name`), then the models of
  the Connections it reaches; a Connection whose list cannot be read adds
  nothing.
- `X-Aiport-Catalog: <version>` on every answer to a request with an
  Application token, forwarded calls and errors included. The version changes
  only when the Application's task names or their targets, its binding or the
  protocol defaults it follows change.
- `GET /v1/aiport/catalog` (MCP tool `aiport_catalog`): `{catalog, changes}`,
  the Catalog and its recent Catalog changes, newest first, each with
  `change_id`, `at_ms`, `revert_of`, `before`/`after` versions, `removed`,
  `added`, `moved`, `routing`, `connections_removed` and `connections_added`.
  The management token gets `403 application_token_required`.
- `GET /v1/aiport/capabilities` carries `catalog: {version, changed_at_ms}`
  (`changed_at_ms` is `null` when no Catalog change is kept).
- A Change that takes away something the Application used within request
  history (it sent a removed task name, or a model name that went through the
  moved binding or protocol default) opens an Issue of cause
  `catalog_changed` with `catalog_change`; a failing call that sent a removed
  task name carries an `X-Aiport-Hint` naming it.
- The Owner can report one failed request from Activity to the Application:
  an Issue of cause `reported_request` with `report` (request id, name sent,
  module, effort, model, Connection, status, error code, aiport's explanation
  and the Owner's note). Fix what the note and the failure point at, then
  call `aiport_issue_fixed`; the next failing call of the same kind carries an
  `X-Aiport-Hint` naming it.

All of it answers about the calling Application only and holds Application
ids, task names, model names and Connection ids, never keys or URLs. It works
with adapting off; only the `catalog_changed` Issues follow the switch.

## Levels and effort

Every aiport shares three Levels, `fast`, `balanced` and `best` (the `levels`
feature). An Application sends one as its model to say how capable the model
must be; the Owner maps each Level per protocol to a Connection and model,
optionally per Application (its own mapping comes first). Only the `model`
field is replaced; the rest of the body is forwarded unchanged. Mapped Levels
are listed in `GET /v1/models` (`owned_by: "aiport"`, or "(aiport Level)" in
`display_name`) and in the Catalog (`task_names` entries with `level: true`
and `protocol`). A Level not mapped on the request's protocol is refused with
`400 level_unmapped`, naming the Level and protocol; aiport never picks
another model.

`X-Aiport-Effort: low|medium|high` (available) says how hard one call is. A
Level or task name may have per-effort models (`effort: {low, medium, high}`
in the Catalog); the call's effort picks one. Without the header aiport reads
the provider's own field, never changing it: `reasoning_effort` (Chat
Completions), `reasoning.effort` (Responses), `output_config.effort`
(Messages), where `none`, `minimal` and `low` count as low and `high`,
`xhigh` and `max` as high; numeric budgets are ignored. Neither means
`medium`. The header is never forwarded.

Both briefs, the `AGENTS.md` line and the MCP instructions teach the same:
send a Level or a task name, never a provider's model id, and state effort
with `X-Aiport-Effort`.

## Guidance versions

aiport numbers the guidance it gives coding agents (the briefs, the
`AGENTS.md` line from `connect.AgentsLine`, the MCP instructions) in
`connect.GuidanceVersion`, with a change log per version in
`internal/connect/guidance.go` (the `guidance` feature). Raise the version,
and add an entry that says exactly what an agent must change, whenever that
guidance changes in a way agents must act on. Per Application aiport keeps
the version its coding agent last confirmed:

- `GET /v1/aiport/guidance` (MCP tool `aiport_guidance`), with the
  Application's own token: `{current, confirmed, confirmed_at_ms, behind,
  changes, agents_line, steps}`; `changes` lists every version after the
  confirmed one (`version`, `summary`, `actions`), `steps` end with
  confirming.
- `POST /v1/aiport/guidance/done` (MCP tool `aiport_guidance_done`) records
  the current version.
- `GET /v1/aiport/capabilities` carries `guidance: {current, confirmed}`
  (`confirmed` is `null` when the Application never confirmed).
- While behind, the MCP `initialize` instructions start with a one-line
  reminder and every other aiport tool answer ends with it, as a second text
  block.

It works with adapting off, and an Application only ever sees its own state.

## The two briefs

Both briefs are generated from the live gateway: its real listen address and
which protocols are configured. The JSON document carries them as
`integration` (Application brief) and `agent_integration` (coding-agent
brief), each with `for`, `title`, `intro`, `steps`, `snippets` and `note`, and
lists where they are served under `briefs` (`for`, `title`, `audience`,
`url`); the markdown guide links both. A snippet has `id`, `title`,
`language` (`python`, `typescript`, `curl`, `json`, `sh` or `text`), optional
`sdk` and `protocol`, `available` (false for a protocol this gateway has not
configured) and `code`. The dashboard's Integrate page renders the same data.

- The **Application brief** covers pairing for the Application's own token,
  keeping it in `.aiport` (gitignored, loaded at startup) as `AIPORT_BASE_URL`
  and `AIPORT_APP_TOKEN`, what to do when it is lost or revoked, SDK base URLs with snippets for the
  OpenAI and Anthropic SDKs in Python and TypeScript plus curl, removing
  provider keys, model names (a Level or a task name, never a provider's
  model id; read `GET /v1/models`, the `models` snippet; `level_unmapped`),
  effort (`X-Aiport-Effort`, the `level` snippet), watching
  `X-Aiport-Catalog`, errors and retries,
  `X-Aiport-Request-Id`, Issues (including `catalog_changed`), and
  verification.
- The **coding-agent brief** covers the project MCP entry, the `AGENTS.md`
  line, keeping the guidance current (`aiport_guidance_done` after setting
  up), the tools `aiport_capabilities`, `aiport_catalog`, `aiport_guidance`,
  `aiport_guidance_done`, `aiport_request_change`, `aiport_my_requests`,
  `aiport_my_issues` and `aiport_issue_fixed` and their HTTP equivalents,
  checking the Catalog (and re-reading it when `X-Aiport-Catalog` changes),
  the Application's Issues and `aiport_guidance` at the start of a session
  and saying when an Issue is fixed on its side (the `catalog` snippet), how to write a Change
  request (`kind`, `need`, `details`, `why`, `scope`, `scope_case`, and the
  optional `alias`, `prefer_connection` and `prefer_model`), naming the need
  (a Level or task name and an effort) rather than a provider (the Owner
  decides which provider serves an Application; a preference is advice),
  waiting
  for the Owner's decision, the permission (`403 change_requests_not_allowed`
  until the Owner allows it) and hints.

When the Owner turns adapting off (`adapting.disabled: true`), the document
marks `change_requests`, `issues` and `X-Aiport-Hint` with status `off`, the `mcp`
feature says the change tools are off, `adapting` is `false`, and the
coding-agent brief opens with the same one-line reason. The `catalog`,
`levels` and `guidance` features and `X-Aiport-Catalog` stay `available`.

## Hints and the coding-agent brief

Applications the Owner allows to request changes are told how to ask when
they hit a wall a Change could remove: `target_unconfigured` and
`target_disabled` errors gain `error.hint` (the request kind, the
`/v1/aiport/requests` endpoint and the MCP tool `aiport_request_change`), and a
provider `404` for the requested model gains an `X-Aiport-Hint` response
header, with the provider's body passed through unchanged. Status codes and
error codes never change, and other Applications see no hint. While an Issue
is open for the Application, a failing call's `X-Aiport-Hint` names that Issue
(`/v1/aiport/issues/{id}` and the MCP tool `aiport_my_issues`) instead, and
`/v1/aiport/capabilities` shows `open_issues`.

The coding-agent brief tells the coding agent to add a project-level MCP entry
pointing at this gateway's `/mcp` whose `headersHelper` reads the token from
`AIPORT_APP_TOKEN` or, when that is unset, from `.aiport` at the repository root
(never the value itself in `.mcp.json`), and the line from
`connect.AgentsLine` to the application's `AGENTS.md`: send a Level or a task
name, never a provider's model id; state effort with `X-Aiport-Effort`; check
`aiport_catalog`, `aiport_my_issues` and `aiport_guidance` at the start of a
session and call `aiport_guidance_done` after updating. The MCP `initialize`
instructions (`internal/server/mcp.go`) say the same.

## Extensions

The Owner can list Extensions (programs that act as a policy when a Change
request arrives, or as a guardrail before an approved Change is applied) under
`extensions:` in `aiport.yaml`. A refusal or block carries the Extension's
reason (`change_request_refused`, `guardrail_blocked`). The protocol is in
[extensions.md](extensions.md).
