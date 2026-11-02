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
model lists. Host/Origin protections still apply.

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
  provider keys, model names and Aliases, errors and retries,
  `X-Aiport-Request-Id`, and verification.
- The **coding-agent brief** covers the project MCP entry, the `AGENTS.md`
  line, the tools `aiport_capabilities`, `aiport_request_change` and
  `aiport_my_requests` and their HTTP equivalents, how to write a Change
  request (`kind`, `need`, `details`, `why`, `scope`, `scope_case`), waiting
  for the Owner's decision, the permission (`403 change_requests_not_allowed`
  until the Owner allows it) and hints.

When the Owner turns adapting off (`adapting.disabled: true`), the document
marks `change_requests` and `X-Aiport-Hint` with status `off`, the `mcp`
feature says the change tools are off, `adapting` is `false`, and the
coding-agent brief opens with the same one-line reason.

## Hints and the coding-agent brief

Applications the Owner allows to request changes are told how to ask when
they hit a wall a Change could remove: `target_unconfigured` and
`target_disabled` errors gain `error.hint` (the request kind, the
`/v1/aiport/requests` endpoint and the MCP tool `aiport_request_change`), and a
provider `404` for the requested model gains an `X-Aiport-Hint` response
header, with the provider's body passed through unchanged. Status codes and
error codes never change, and other Applications see no hint.

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
