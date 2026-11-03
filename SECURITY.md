# Security policy

aiport holds your provider keys and decides which application may call which
model, so security reports are welcome and handled first.

## Reporting a vulnerability

Report privately through GitHub: open the repository's **Security** tab and choose
**Report a vulnerability**. Please do not open a public issue, pull request or
discussion for a vulnerability.

Include what you found, the aiport version (`aiport version`) or commit, and the
steps to reproduce it. Never include a real provider key or token; a made-up one
shows the problem just as well.

You will get an acknowledgement within a week. Once a fix is released, the
advisory credits you unless you ask not to be named.

## Supported versions

aiport is pre-1.0: only the latest release and `main` receive fixes.

## Scope

In scope: anything that lets a caller reach a provider, a model, a key or another
Application's data it should not; tokens or keys written to logs, the request
history or the dashboard; bypassing the dashboard login, the management token, the
CSRF check or the Host and Origin checks; and an agent run acting outside its git
worktree or without the approval it asks for.

Out of scope: problems that need an attacker already running code as your user
on your machine, and what a provider or a coding agent does with a call aiport was
allowed to send (see [what aiport cannot see](docs/manual.md#explain-by-a-coding-agent)).
