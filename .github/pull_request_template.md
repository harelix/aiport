## What and why

<!-- What this changes, and the issue it closes (Closes #123). -->

## Checks

- [ ] `make test`, `make check` and `make check-offline` pass
- [ ] Dashboard changes: `make web-build`, and the rebuilt `internal/dashboard/dist` is committed
- [ ] Behaviour changes are in `docs/manual.md` (and `docs/configuration.md` for config)
- [ ] New Go dependency: reason given, `go mod vendor` run, license added to `THIRD_PARTY.md`
- [ ] No real keys, tokens or private names; every commit is signed off (`git commit -s`)
