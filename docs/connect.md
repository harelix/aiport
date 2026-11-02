# aiport Connect

aiport Connect is how applications learn to talk to an aiport gateway. A running
gateway publishes it, without authentication, at:

- `GET /_connect`: JSON for programs
- `GET /_connect?format=markdown` (or `Accept: text/markdown`): a guide for people and coding agents
- `GET /_connect?format=handover`: only the integration instructions, to paste to a coding agent working on an application

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
dashboard (Routing → Applications), optionally binding a connection; aiport adds
the application as a client, with its own token, without a restart. `GET` on the
poll path returns `pending`, `rejected`, or `approved` with the token, which is
delivered exactly once. Unapproved requests expire after 10 minutes. The token
lives in the private credential file as `AIPORT_APP_<ID>_TOKEN`. The source of truth is
`internal/connect/connect.go`; bump `Version` on incompatible changes.

## Extensions

The Owner can list Extensions (programs that act as a policy when a Change
request arrives, or as a guardrail before an approved Change is applied) under
`extensions:` in `aiport.yaml`. A refusal or block carries the Extension's
reason (`change_request_refused`, `guardrail_blocked`). The protocol is in
[extensions.md](extensions.md).
