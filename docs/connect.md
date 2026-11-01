# Aiport Connect

Aiport Connect is how applications learn to talk to an Aiport gateway. A running
gateway publishes it, without authentication, at:

- `GET /_connect`: JSON for programs
- `GET /_connect?format=markdown` (or `Accept: text/markdown`): a guide for people and coding agents

The document reflects the live gateway: base URL, which native protocols are
configured, whether authentication is required, request/response headers,
error codes and features. It never contains secrets, target URLs, key names or
model lists. Host/Origin protections still apply.

Each feature and header has a status. Only `available` items are implemented;
`planned` items (model allowlists, effort tiers, capabilities, pairing, routing
headers) are published early so integrators can prepare. The source of truth is
`internal/connect/connect.go`; bump `Version` on incompatible changes.
