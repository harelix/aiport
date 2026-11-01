# Implementation evidence and remaining gates

Current state: experimental native forwarding and terminal onboarding, built on
an authenticated loopback Go server. No provider/client is released as supported.

Implemented:

- `serve`, `check`, `token`, `version`, guided `init`, and per-launch `agent` CLI.
- Strict bounded YAML, environment credential references, private config files,
  distinct local credentials, fixed config snapshots and redacted UUID errors.
- Native Chat, Responses and Messages forwarding, protocol/default/origin routing,
  byte-preserving model aliases, upstream credential replacement, hop/cookie/internal
  header removal, no redirects/retries, bounded admission and phase/idle deadlines.
- A bounded request buffer before dispatch ensures 413 precedes upstream access;
  responses copy and flush incrementally without whole-response buffering.
- Six onboarding presets; comma-separated provider/agent selections; private local
  credential companion file; exclusive creation preserves existing user setup.
- Codex Responses and Claude Code Messages launch configuration, with configured
  upstream and management secrets removed from the child environment. No global
  client files are changed. Both clients were detected locally; no session was run.

Fresh upstream connections prevent Go transport replay on stale pooled connections.
This trades connection reuse for a straightforward no-retry contract. Performance,
resource usage, pooling and p95 overhead remain unmeasured. SQLite and its packaging
choice remain deferred to M2; forwarding imports no persistence or UI code.

Validation: `go test ./...`, `go test -race ./...`, `go vet ./...`, and a CGO-disabled
standalone build. A binary smoke check creates all six presets plus both origin
credentials, validates synthetic environment references, exercises both agent
`--dry-run` paths, checks secret-safe output, and rejects existing-file overwrite.
Proxy fixtures cover native bytes, unknown fields, query semantics, error statuses,
tool/heartbeat SSE, early flush, aliases, protocol routing, cancellation, overload,
size, idle deadline and incomplete-stream abort. Configuration and HTTP-boundary
tests cover malformed settings, credentials, Host, Origin and privilege separation.

Runtime socket fixtures cannot run in this sandbox (EPERM on bind). The user
separately supplied output showing the M0 listener starting successfully on macOS.
No network-phase deadline, live TLS, paid generation, client session, cross-platform
archive or clean-install release claim follows from the in-memory tests.

| Milestone | State / remaining gate |
| --- | --- |
| M0 | Foundation implemented; license/release decisions remain open |
| M1 | Experimental forwarding and fixtures implemented; real-client integration and full failure/load evidence pending |
| M2 | Planned: SQLite driver/package tests, bounded metadata queue, persistence/retention/query API and fault/privacy evidence |
| M3 | Planned: modern embedded React dashboard, web onboarding/settings, protected atomic config edits |
| M4 | Partial native forwarding only: six-provider live compatibility, supported endpoint subsets, model/client versions and tools/usage evidence pending |
| M5 | Planned: maintainer license decision, contribution/security/release docs, platform archives, clean install and measurements |

A01–A07/A10/A11/A13/A14 have focused partial fixture coverage, not complete
acceptance. A08/A09/A12/A15/A16 await metadata/UI work. A17/A18 await release/live
validation. See `docs/compatibility.md` for limitations and primary sources.

## Dashboard/session metadata block

Implemented embedded React/TypeScript assets built with Vite, token-in-memory
login/lock, Overview, Activity with stable cursor pagination/shared filters,
request metadata details, and read-only Settings. Includes dark/light themes,
responsive table containers and explicit disabled/session-only/error/empty states.
All UI queries use authenticated local APIs; assets alone are public. No fake
activity ships. Usage, TTFT and cost are unavailable rather than fabricated.

Added bounded nonblocking session metadata queue plus a 500-record ring and
visible loss/eviction counters, with no SQLite dependency on forwarding. This is
an explicit preview deviation, not completion of M2. Dependency downloads and
browser access failed in this environment. SQLite persistence, disk fault/migration
and retention gates, visual/keyboard browser QA, web config writes/onboarding,
provider probes and release packaging remain open. TypeScript, Vite builds and
Go/race tests pass; rendered appearance has not been inspected.


## Authentication and Nord update

Implemented first-run management-token verification followed by password/optional
username setup, normal password login, HttpOnly 12-hour browser sessions, sign-out,
management-token bypass/recovery popup and CLI token lookup. Settings can reset
credentials and disable/re-enable local authentication, saved atomically to YAML.
Prior sessions are revoked on changes; provider settings remain intact. A disabled
mode retains Host/Origin/CSRF protections. Added flow/restart/recovery tests and
PBKDF2/session tests. Nord official color variables now cover light/dark UI and
interactive CLI accents; checked key text/control contrasts. Supplied original
`visuals/logo-circle-with-sky` is embedded for login/sidebar/recovery and favicon,
preserving its original artwork. Browser visual verification remains pending.
