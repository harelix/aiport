# Contributing to aiport

Thanks for helping. aiport is a small, local, single-binary AI gateway; please read
`aiport_DESIGN.md` before proposing larger changes, and open an issue first for
anything that changes the product scope.

## Setup

- Go 1.27+ is the only requirement to build: `make build`, `make test`, `make check`,
  `make check-offline`.
- Dashboard changes also need Node.js and pnpm: `pnpm --dir web install`, then
  `make web-build`, and commit the rebuilt `internal/dashboard/dist`.
- Go dependencies are vendored. Adding one requires a reason in the PR, `go mod vendor`,
  and an entry in `THIRD_PARTY.md` with its license (compatible with Apache-2.0).

## Live development

```sh
pnpm --dir web install   # once, for the dashboard dev server
make dev
```

- **Go:** saving any `.go` file (or the config) rebuilds and restarts the gateway,
  about 1–2 seconds. Go has no in-process hot reload, so a restart is the
  reliable equivalent. A failed build keeps the watcher running; fix and save.
- **Dashboard:** open http://127.0.0.1:5173. Vite reloads the page on every save
  and updates CSS in place, and proxies `/api` to the gateway on
  `127.0.0.1:8787`. (Component state is not preserved across edits; that needs
  `@vitejs/plugin-react`, which is not a dependency.)
- The dev proxy rewrites `Origin` only for the dev page itself, so the gateway's
  cross-site protections still apply to every other site.
- `make dev` uses `bin/aiport-dev`. Run `make web-build` before committing UI
  changes so the embedded dashboard is current.

## Pull requests

- Keep them small and focused; include tests for behaviour changes.
- Never commit secrets, real provider responses or personal data; use synthetic fixtures.
- Update docs (`README.md`, `docs/`) when user-visible behaviour changes.

## Sign your commits (DCO)

Every commit must carry a `Signed-off-by` line, certifying the
[Developer Certificate of Origin](https://developercertificate.org/): that you
wrote the change or otherwise have the right to submit it under the project license.

```sh
git commit -s -m "Describe your change"
```

## License

By contributing, you agree that your contributions are licensed under the
Apache License 2.0 (see `LICENSE`). The aiport name and logo are covered by
`TRADEMARKS.md`.
