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
snippets) use the gateway's advertised address: the first entry of
`hostnames` on the listen port when the Owner set one (for example
`http://aiport.localhost:8787`), otherwise the listen address. aiport accepts
`Host` and `Origin` for the listen address and each configured hostname on
the listen port, and refuses everything else with `403 invalid_host` /
`invalid_origin`.

Each feature and header has a status. Only `available` items are implemented;
`planned` items (model allowlists, effort tiers, capabilities, routing headers)
are published early so integrators can prepare.

## Pairing

An application without a token asks for one: `POST /v1/aiport/pair`
`{"name": "...", "description": "..."}` (public on loopback, rate-limited) returns
a short `code` and a secret `poll` path. The owner approves the code in the
dashboard's Integrate page (or the Applications table on Routing), optionally binding a connection; aiport adds
the application as a client, with its own token, without a restart. `GET` on the
poll path returns `pending`, `rejected`, or `approved` with the token, which is
delivered exactly once. Unapproved requests expire after 10 minutes. The token
lives in the private credential file as `AIPORT_APP_<ID>_TOKEN`. The source of truth is
`internal/connect/connect.go`; bump `Version` on incompatible changes.

## The Catalog

An Application's Catalog is what it can use through aiport: its task names
(its own Aliases, then the shared Aliases of the Connections it reaches) and
the models those Connections list, with a version. The document describes it
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

All of it answers about the calling Application only and holds Application
ids, task names, model names and Connection ids, never keys or URLs. It works
with adapting off; only the `catalog_changed` Issues follow the switch.

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
  `AIPORT_BASE_URL` and `AIPORT_APP_TOKEN`, SDK base URLs with snippets for the
  OpenAI and Anthropic SDKs in Python and TypeScript plus curl, removing
  provider keys, model names (read `GET /v1/models`, prefer task names; the
  `models` snippet), watching `X-Aiport-Catalog`, errors and retries,
  `X-Aiport-Request-Id`, Issues (including `catalog_changed`), and
  verification.
- The **coding-agent brief** covers the project MCP entry, the `AGENTS.md`
  line, the tools `aiport_capabilities`, `aiport_catalog`,
  `aiport_request_change`, `aiport_my_requests`, `aiport_my_issues` and
  `aiport_issue_fixed` and their HTTP equivalents, checking the Catalog (and
  re-reading it when `X-Aiport-Catalog` changes) and the Application's Issues
  at the start of a session and saying when one is fixed on its side (the
  `catalog` snippet), how to write a Change
  request (`kind`, `need`, `details`, `why`, `scope`, `scope_case`, and the
  optional `alias`, `prefer_connection` and `prefer_model`), naming tasks
  rather than providers (the Owner decides which provider serves an
  Application; a preference is advice), waiting
  for the Owner's decision, the permission (`403 change_requests_not_allowed`
  until the Owner allows it) and hints.

When the Owner turns adapting off (`adapting.disabled: true`), the document
marks `change_requests`, `issues` and `X-Aiport-Hint` with status `off`, the `mcp`
feature says the change tools are off, `adapting` is `false`, and the
coding-agent brief opens with the same one-line reason. The `catalog` feature
and `X-Aiport-Catalog` stay `available`.

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
pointing at this gateway's `/mcp`, with the token taken from
`${AIPORT_APP_TOKEN}` (never the value), and the line from
`connect.AgentsLine` to the application's `AGENTS.md`.

## Extensions

The Owner can list Extensions (programs that act as a policy when a Change
request arrives, or as a guardrail before an approved Change is applied) under
`extensions:` in `aiport.yaml`. A refusal or block carries the Extension's
reason (`change_request_refused`, `guardrail_blocked`). The protocol is in
[extensions.md](extensions.md).
