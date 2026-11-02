# AGENTS.md

## Commands

- Build: `make build` (-> `bin/aiport`; pure Go, `CGO_ENABLED=0`, `GOFLAGS=-mod=vendor`)
- Test: `make test` · Vet: `make check` · No-network build: `make check-offline`
- Dashboard: `pnpm --dir web install`, then `make web-build`
- Live development: `make dev`

## Rules

- Dependencies are vendored. Do not run `go get`; adding one needs
  `go mod vendor` and an entry in `THIRD_PARTY.md`.
- After changing `web/`, run `make web-build` and commit the rebuilt
  `internal/dashboard/dist` (it is embedded and committed on purpose).
- Sign off every commit (`git commit -s`; for merges `git merge --signoff`,
  since `-s` there means strategy).
- Always write "aiport" in lowercase. `X-Aiport-*` headers and `AIPORT_*`
  environment variables keep their case.
- Update `docs/manual.md` whenever user-visible behaviour changes.
- Published specs, tickets and docs use made-up Application and model names.
- Never commit `aiport.yaml`, `*.env`, `*.db` or `aiport.log`; they are
  local and ignored.
- Tests drive the server Handler over HTTP with a temporary config and
  SQLite, a fixed clock and fake upstreams; assert on responses, not internals.

## Agent skills

### Issue tracker

Issues and specs live in GitHub Issues on harelix/aiport, managed with the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

The five default labels: needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` (created when the first ADR is written) at the repo root. See `docs/agents/domain.md`.
