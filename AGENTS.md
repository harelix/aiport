# AGENTS.md

## Commands

- Build: `make build` (-> `bin/aiport`; pure Go, `CGO_ENABLED=0`, `GOFLAGS=-mod=vendor`)
- Test: `make test` · Vet: `make check` · No-network build: `make check-offline`
- Dashboard: `pnpm --dir web install`, then `make web-build`
- Live development: `make dev`
- Docs site: `pnpm --dir site install`, then `make docs` (-> `site/dist`) or
  `make docs-dev`; published to GitHub Pages from `main` by `.github/workflows/docs.yml`

## Rules

- Dependencies are vendored. Do not run `go get`; adding one needs
  `go mod vendor` and an entry in `THIRD_PARTY.md`.
- After changing `web/`, run `make web-build` and commit the rebuilt
  `internal/dashboard/dist` (it is embedded and committed on purpose).
- Sign off every commit (`git commit -s`; for merges `git merge --signoff`,
  since `-s` there means strategy).
- Icons come only from Lucide (`lucide-react`), and colours only from the
  windsock logo: orange `#FE5234` (`#C93A1F` for readable text) and white.
- Always write "aiport" in lowercase. `X-Aiport-*` headers and `AIPORT_*`
  environment variables keep their case.
- Update `docs/manual.md` whenever user-visible behaviour changes. The docs
  site is generated from `docs/` and `GLOSSARY.md` (`site/scripts/sync-docs.mjs`
  splits the manual into one page per `###` section): never edit
  `site/src/content/docs`. A new `##` section in the manual needs a place in
  `MANUAL.sections` there; `make docs` fails on a broken link or anchor.
- Published specs, tickets and docs use made-up Application and model names.
- Never commit `aiport.yaml`, `*.env`, `*.db` or `aiport.log`; they are
  local and ignored.
- Secrets and private words are checked before every commit (`make hooks`,
  gitleaks and `.git/info/private-patterns`); never bypass it with `--no-verify`.
  `make leaks` scans all history.
- Tests drive the server Handler over HTTP with a temporary config and
  SQLite, a fixed clock and fake upstreams; assert on responses, not internals.

## Agent skills

### Issue tracker

Issues and specs live in GitHub Issues on harelix/aiport, managed with the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

The five default labels: needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` (created when the first ADR is written) at the repo root. See `docs/agents/domain.md`.
