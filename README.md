<p align="center">
  <img src="visuals/logo-256.png" alt="aiport logo" width="128" height="128">
</p>

# aiport

A local AI gateway, under development from [the design](aiport_DESIGN.md).

**Current build: experimental native forwarding and terminal onboarding.** Chat
Completions, Responses and Messages share a byte-forwarding path. No provider or
coding-agent combination has passed the live compatibility gate yet. The embedded
dashboard shows request metadata saved in SQLite and edits providers and routing.
Web-first onboarding is still planned.

## Run the existing configuration

```sh
./bin/aiport check
./bin/aiport serve
```

The original example points to Ollama at `http://127.0.0.1:11434/v1`; start your
local provider separately. Use your configured upstream model identifier.

No token export is needed. On first start, `serve`/`check`/`agent` create any missing
local tokens and save them to the private file next to the config
(`aiport.yaml.env`, mode 0600). Every later start, in any terminal, reuses them.
A token already exported in your shell takes precedence. When started in an
interactive terminal, `serve` prints the dashboard URL and management token;
redirected output gets a hint instead. Retrieve it later with
`./bin/aiport auth token`.

Stop an older running aiport with Ctrl+C before starting the rebuilt binary.

## First-run onboarding

For this workspace, where `aiport.yaml` already exists, use a new config path:

```sh
./bin/aiport init --config aiport-onboard.yaml
```

The terminal wizard selects one or more providers and coding agents, detects
installed clients, and creates private config and local-credential files.
For a fresh installation, omit `--config` to create `aiport.yaml`.
Existing files are never overwritten. Example without prompts:

```sh
./bin/aiport init --config aiport-onboard.yaml \
  --provider openai,anthropic --agent codex,claude
source aiport-onboard.yaml.env
# Set OPENAI_API_KEY and ANTHROPIC_API_KEY in your shell to your own provider keys.
./bin/aiport check --config aiport-onboard.yaml
./bin/aiport serve --config aiport-onboard.yaml
```

The wizard prints commands using `aiport`; use `./bin/aiport` in this checkout,
or put the built executable on your PATH. Upstream keys are referenced through
environment variables, never pasted into the wizard, printed, or stored in YAML.
The generated `.env` file contains only random local credentials. Keep it private
and out of version control. Source it in each terminal that needs those tokens.

Preset IDs: `openai`, `anthropic`, `bedrock`, `meta`, `ollama`, `llamacpp`.
Use comma-separated IDs for multiple targets. `--region REGION` selects the
Bedrock region (default `us-east-1`). `--base-url URL` overrides one selected
provider's API prefix, including a different local port or AWS endpoint.
The first selected target of each protocol becomes its default. Inspect/edit
the generated YAML to choose a different per-agent target before starting.

## Launch a coding agent

In another terminal, source the generated credential file and launch your client:

```sh
source aiport-onboard.yaml.env
./bin/aiport agent codex --config aiport-onboard.yaml --model YOUR_OPENAI_MODEL
# Or, when the config includes an Anthropic target and claude origin:
./bin/aiport agent claude --config aiport-onboard.yaml --model YOUR_ANTHROPIC_MODEL
```

Add `--dry-run` to inspect the endpoint/model without launching. Extra native
agent options go after `--`. Settings apply to this launch; global client settings
are not edited. The child receives its own local token, not configured upstream
keys or the management credential. Codex uses Responses over HTTP, with client
request/stream retries disabled. Claude Code uses native Messages. The selected
target must actually support the client's protocol; aiport does no translation.
Ollama/llama.cpp Chat compatibility alone does not establish Codex compatibility.

## Connect an application

Any running gateway describes how to talk to it at `/_connect`, without a token:

```sh
curl http://127.0.0.1:8787/_connect                    # JSON for programs
curl 'http://127.0.0.1:8787/_connect?format=markdown'  # guide for people/agents
```

See [docs/connect.md](docs/connect.md).

**Integrate one of your applications.** Open the dashboard's **Routing** page and
use **Copy instructions** under *Hand over to a coding agent*: paste them to the
coding agent working on that application (or into its `AGENTS.md`). The agent
switches the application's SDK to aiport and **pairs** it: the application asks
for its own token and shows a short code, you approve the code under
*Applications* on the same page, and the application collects its token once.
Its calls then appear under its own name in Activity. The same instructions are
served at `/_connect?format=handover`.

## Check health

```sh
curl -H "Authorization: Bearer $AIPORT_LOCAL_TOKEN" \
  http://127.0.0.1:8787/api/v1/health
```

All API paths require local credentials. Management health requires the global
token. Origin tokens only grant inference access. Host must match the exact
configured IP and port. Foreign browser Origin/Host and cross-site requests are
rejected. No CORS or remote serving. Unknown paths return structured 404 errors;
missing/disabled routes and admission failures are redacted local errors with a
UUID. Native upstream statuses, bodies, content types and request IDs pass through.

## Build from source

The only prerequisite is **Go 1.27+** (`brew install go`, or https://go.dev/dl).
Everything else is in the repository: Go dependencies are vendored, and the
dashboard is committed pre-built in `internal/dashboard/dist`, so no Node.js,
network access, C compiler or database server is needed.

```sh
git clone <repo> aiport && cd aiport
make build            # -> bin/aiport (UI embedded, pure Go, CGO disabled)
./bin/aiport serve
```

What you do **not** need to run: `go get`, `go mod tidy` or `go mod vendor`. All
dependencies, including the pure-Go SQLite engine, are already committed under
`vendor/`; those commands are only for maintainers changing dependencies (see
CONTRIBUTING.md).

Checks: `make test`, `make check`, and `make check-offline` (proves the build needs
no network). `make release VERSION=x.y.z` cross-builds self-contained archives for
macOS, Linux and Windows plus SHA-256 checksums into `dist-release/`.

Changing the dashboard itself additionally needs Node.js and pnpm
(`pnpm --dir web install && make web-build`); commit the rebuilt `dist` with your change.

Deterministic fixtures use in-memory HTTP transports because this execution
sandbox denies listening sockets. Live provider and agent workflows remain
unverified. The user has separately observed the M0 listener start on macOS arm64.
No archives or supported platform release matrix exist yet.

See [configuration](docs/configuration.md), [implementation](docs/implementation.md),
[compatibility](docs/compatibility.md), and [dependency notices](THIRD_PARTY.md).
aiport is licensed under [Apache-2.0](LICENSE); the name and logo are covered by
[TRADEMARKS.md](TRADEMARKS.md). See [CONTRIBUTING.md](CONTRIBUTING.md) (DCO sign-off required).

## Embedded dashboard preview

After restarting the rebuilt binary, open **http://127.0.0.1:8787/** (or your
configured literal loopback address). For first-run access, verify the startup `AIPORT_LOCAL_TOKEN`, then set a password
and an optional username. Later visits use those credentials. An HttpOnly browser
session survives reloads and expires after 12 hours or a server restart. Login
assets and minimal sign-in status are public on the local listener.

Overview shows retained request counts, latency, errors and provider distribution.
Activity has shared time/origin/target/outcome filters, cursor pagination and
metadata details. Settings shows connections, agent routing, and working authentication controls.
Paste provider API keys into each connection card; they apply immediately and are
stored only in the private `aiport.yaml.env` (never YAML, never shown again).
Adding or removing connections still means editing YAML, checking, and restarting.
Cost and token metrics remain unknown. **Check connection** on each provider card
(or **Check all** per group) sends one authenticated `GET /models` to that
connection: it generates nothing and costs nothing, and shows whether the provider
is reachable and accepted the key. It runs only when you click it, and a passing
check does not prove every model or feature works.

Activity is saved to a private SQLite file next to your configuration
(`aiport.yaml.db`, 0600) when `telemetry.enabled: true`. Request handlers only
offer a metadata record to a nonblocking bounded queue; one writer saves records
in batches and deletes rows older than `retention_days` (default 7). History
survives restarts. Dropped records and failed writes are counted and shown.
If the database cannot be opened, aiport says so at startup and keeps the latest
500 records in memory for that session instead. A database written by a newer
aiport turns recording off rather than risk it. Generated onboarding configs
enable telemetry; existing configs are preserved, so set `telemetry.enabled: true`
in yours and restart if you want it. When disabled, no request metadata is recorded.
Deleted rows can linger in SQLite's WAL and free pages; deletion is not secure erasure.

The React/TypeScript UI is embedded in `bin/aiport`; running it needs no frontend
service or network downloads. To rebuild its assets as a contributor:

```sh
pnpm --dir web install
pnpm --dir web build
go build -trimpath -o bin/aiport ./cmd/aiport
```

The current frontend compiled and typechecked with local installed packages.
Registry installation and browser visual/interaction QA remain unverified because
network/browser tooling was unavailable during this work block.

## Sign-in, recovery, and optional authentication

Click **Forgot username or password?** on the login screen for the recovery popup.
It explains how to retrieve and paste the management token to bypass account login.
After recovery, open **Settings → Authentication** to replace username/password.
Agent credentials and upstream provider API keys cannot recover access.

```sh
./bin/aiport auth token --config aiport.yaml
```

Use your actual config path. This explicit command prints the configured token
from the current shell, or from the matching private `aiport.yaml.env` file.
If you changed the token only in the server's environment, use that same shell:
`printf '%s\n' "$AIPORT_LOCAL_TOKEN"` (or your configured variable name).
`aiport token` generates a new token; it does not retrieve a running server's token.
Never paste a recovery token into provider-key fields.

The Authentication section can disable or re-enable local authentication and
reset the account. Disabling removes credential requirements for both management
and inference on the loopback listener; Host/Origin/cross-site protections remain.
Credentials are retained. Re-enabling requires a saved or new password. Source
your private token file before restarting with authentication enabled.

Passwords require at least 12 characters and are stored only as salted
PBKDF2-HMAC-SHA256 hashes in the private YAML. Changes take effect immediately,
atomically persist, and revoke previous browser sessions. Provider settings and
secret references are preserved. No browser account/session token is forwarded
upstream or stored in localStorage.

The entire interface now uses the official **Nord** palette in dark/light modes.
Terminal onboarding/startup accents use Nord Frost when writing to a TTY;
`NO_COLOR`, redirected output, and secret/token output remain plain.
