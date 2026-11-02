# aiport manual

This manual is for the Owner: the person who runs aiport on their machine and
decides what it does. It walks through the dashboard page by page. For exact
configuration keys see [configuration](configuration.md); for the protocol
Applications use see [Connect](connect.md); for policies and guardrails see
[Extensions](extensions.md). Words with a capital letter (Application,
Connection, Alias, Change request, Suggestion, Change, Scope, Extension) are
defined in the [glossary](../GLOSSARY.md).

## What aiport does

aiport is a local AI gateway. Applications on this machine (scripts, services,
coding agents) send their LLM calls to aiport instead of straight to a
provider. aiport forwards each call, in the provider's own format, to one of
your Connections and adds the real provider key on the way out. You get one
place to hold keys, choose where calls go, and see what was called, without
storing prompts or responses.

aiport listens on the loopback address only (`127.0.0.1:8787` by default) and
refuses requests from other sites in your browser.

## Install and run

Build from source (Go is the only prerequisite) and start it:

```sh
make build                 # -> bin/aiport, dashboard included
./bin/aiport init          # first run: choose providers and coding agents
./bin/aiport check         # validate the configuration
./bin/aiport serve         # start the gateway
```

`init` writes `aiport.yaml` and a private credential file next to it
(`aiport.yaml.env`, mode 0600) and never overwrites existing files. Provider
keys are never asked for in the terminal: set them in the dashboard, or in
your shell. `serve` prints the dashboard address and the management token
when started in a terminal; `./bin/aiport auth token` prints the token again
later.

Open the dashboard at `http://127.0.0.1:8787/`. On the first visit, paste the
management token, then set a password (at least 12 characters) and an
optional username. Later visits use those. A browser session lasts 12 hours
or until aiport restarts.

Other commands: `aiport agent codex|claude --model MODEL` launches a coding
agent through aiport with its own token, `aiport task` manages the task
queue, `aiport version` prints the version.

## A tour of the dashboard

The sidebar has three groups:

- **Traffic**: Overview and Activity.
- **Gateway**: Providers, Routing, Integrate, Compression, Agents and Changes.
- **Workspace**: Tasks and Settings.

The header has the help button (this manual), the Light / Dark / System
appearance switch and Refresh. The dashboard refreshes itself every 10
seconds while it is open. The bottom of the sidebar shows the address aiport
listens on and Sign out (or "Authentication off" when it is turned off).

- **Overview**: request counts, latency, error rate, hourly traffic, provider
  distribution and the latest requests. When something waits for your
  decision, one line here links to Changes.
- **Activity**: every recorded request, filtered by period, Application
  ("Origin"), Connection ("Target") and outcome, with a details view per
  request. Only metadata is recorded: time, route, model, status, latency.
- **Providers**: your Connections, grouped Local and Remote.
- **Routing**: protocol defaults and the Applications table.
- **Integrate**: connecting an application, with the instructions aiport
  writes for coding agents.
- **Compression**: opt-in request compression and whether it hurts answers
  (see [compression](compression.md)).
- **Agents**: coding agents launched through `aiport agent`, each with its
  own token.
- **Changes**: what waits for your decision and every configuration Change
  made through aiport, with Revert.
- **Tasks**: a queue of work for coding agents; you own intent and priority.
- **Settings**: sign-in, authentication and the configuration file.

## Connections and keys

A Connection is a provider aiport can forward to: OpenAI, Anthropic, Amazon
Bedrock, Meta Model API, or a local server such as Ollama or llama.cpp. Each
Connection speaks one protocol, `openai` or `anthropic`.

On **Providers**:

- **Add a connection** from a preset, or any other server that speaks the
  OpenAI or Anthropic API natively (remote URLs must use HTTPS); **Remove**
  takes one away. Each is a Change you can revert in Changes.
- Paste a provider API key into a remote Connection's card. It applies at
  once and is stored only in the private `aiport.yaml.env`, never in
  `aiport.yaml`, and is never shown again. Local Connections need no key.
- **Check connection** (or **Check all** per group) sends one authenticated
  `GET /models` to the provider. It generates nothing and costs nothing; it
  shows whether the provider is reachable and accepted the key. A passing
  check does not prove every model works.
- **Make OpenAI default** / **Make Anthropic default** sets the Connection a
  protocol uses when nothing else picks one.

Applications never see provider keys. They send their own aiport token, and
aiport adds the key upstream.

## Routing and protocol defaults

aiport speaks the providers' native APIs and does not translate between
them: an OpenAI SDK call needs a Connection that speaks the `openai`
protocol, and an Anthropic SDK call one that speaks `anthropic`.

For each call aiport picks the Connection in this order:

1. the Application's own Alias for the requested model, if it has one (it may
   name another Connection);
2. the Connection the Application is bound to, if any;
3. the protocol default set under **Routing → Protocol defaults**.

A protocol with no Connection answers `503 target_unconfigured`; a disabled
Connection answers `503 target_disabled`.

**Routing** also shows the two base URLs Applications use
(`http://127.0.0.1:8787/v1` for the OpenAI protocol, `http://127.0.0.1:8787`
for the Anthropic protocol), the `/_connect` protocol description, and the
Applications table.

## Applications and pairing

An Application is anything that calls aiport with its own token. Each call
is attributed to the Application that sent it, so Activity shows who called
what. Calls without an Application token show as `unknown`.

An Application gets its token by **pairing**:

1. It asks aiport (`POST /v1/aiport/pair` with its name; no token needed) and
   shows you a short code.
2. The request appears on the **Integrate** page (and in the Applications
   table on Routing) with that code. Check the application shows the same
   code. Optionally choose a Connection to bind it to and tick **May request
   changes**, then **Approve**, or **Reject**.
3. The application collects its token once. aiport adds it to the
   configuration without a restart; the token is stored in the private
   credential file.

Unapproved requests expire after 10 minutes. The Applications table on
Routing lists every Application with its client id, Connection, own Aliases,
and the **May request changes** toggle.

## The Integrate page and the two briefs

The **Integrate** page is where you connect an application. You do not have
to write instructions for the coding agent doing the work: aiport writes
them itself, from this gateway's real address and configured protocols, and
serves them at `/_connect`:

- **Application brief** (`/_connect?format=handover&for=application`): for
  the coding agent changing an application so its LLM calls go through
  aiport. It covers pairing, configuration (`AIPORT_BASE_URL`,
  `AIPORT_APP_TOKEN`), pointing the OpenAI or Anthropic SDK at aiport,
  removing provider keys from the application, model names and Aliases,
  errors and retries, logging `X-Aiport-Request-Id`, and how to verify.
- **Coding-agent brief** (`/_connect?format=handover&for=agent`): for coding
  agents working in that application's repository later. It covers the
  project MCP entry (which reads the token from `${AIPORT_APP_TOKEN}`, never
  the value), the line for the repository's `AGENTS.md`, the tools
  `aiport_capabilities`, `aiport_request_change` and `aiport_my_requests`
  and their HTTP equivalents, how to write a good Change request, waiting
  for your decision, the permission, and hints.

`/_connect?format=handover` returns both, one after the other.

On the page:

- **Checklist**: choose an Application, or "A new application". The steps
  tick themselves as they happen: **Paired** (it is in the configuration),
  **First call seen** (Activity has a call from it in the last 7 days),
  **May request changes** (with the toggle right there), and **MCP entry
  copied** (ticked when you copy the MCP entry, or by hand). Pairing
  requests waiting for you appear at the top with Approve and Reject.
- **Application** and **Coding agent** tabs show each brief. **Copy whole
  brief** copies it as markdown to paste to the coding agent; you can also
  give the agent the brief's address. Each snippet has its own **Copy**.
- The **Python / TypeScript / curl** picker switches the SDK snippets.
  Snippets for a protocol this gateway has not configured are marked "not
  configured".

A typical integration: copy the Application brief to the coding agent
working on the application; approve its pairing code when it appears; wait
for "First call seen"; then, if you want agents in that repository to ask
you for Changes, allow it and have the agent follow the Coding-agent brief.

## Aliases

An Alias is a model name an Application sends that aiport replaces with a
real model on a Connection. Only the `model` field of the request is
rewritten.

- **Shared Aliases** apply to every Application (`aliases:` in `aiport.yaml`,
  one per Connection and name).
- **An Application's own Aliases** come first and may select another
  Connection than its usual one (`origins[].aliases`).

You can write Aliases in the configuration file, or let them arrive as
Changes: approving a Suggestion or Change request of kind
`application_alias` adds one, for that Application or, if you grant
**every Application** as the Scope, as a shared Alias. An Application can
list its own Aliases with `GET /v1/aiport/capabilities`.

## Activity and history

With `telemetry.enabled: true`, aiport records request metadata (never
prompts, responses or headers) in a private SQLite file next to the
configuration (`aiport.yaml.db`, 0600). Recording never slows the
forwarding path; if records are dropped or fail to save, the banner on
Overview and Activity says how many.

**Pruning.** Rows older than `telemetry.retention_days` (default 7) are
deleted at startup and hourly. The same setting decides how long decided
Suggestions and Change requests are kept (dismissed ones at least 7 days,
the quiet period). Applied Changes are never pruned. Deleted rows can linger
in SQLite's files; deletion is not secure erasure.

If the database cannot be opened, aiport says so and keeps the latest 500
records in memory for that session.

Every response carries `X-Aiport-Request-Id`. When an application logs it,
you can find the exact call in Activity.

## Changes

aiport adapts to the Applications that use it, but only with your approval.
**Changes** in the Gateway menu stays quiet: a badge appears only when
something needs a decision.

### What waits for your decision

- **Suggestions** come from aiport itself, from patterns in request history,
  with the evidence (how many requests, over which time window). It
  suggests an Alias when one Application keeps getting a provider `404` for
  the same model; binding an Application to a working Connection when its
  calls to one Connection keep failing; and pairing when requests keep
  arriving with a missing or unknown token (informational: open the
  Integrate page to pair it, or Dismiss). Thresholds are under `adapting:`
  in the configuration.
- **Change requests** come from Applications you allowed to request changes.
  Each shows the kind, the details, why it is needed, the Scope asked for
  and the argument for it, and who filed it (as stated). Identical requests
  from several Applications are merged into one item, each with its own
  argument; one decision answers them all. Request text is plain text from
  the Application: it is shown as is and never sent to a model.

### Scope

Every decision has a Scope: **only** the requesting Application, or **every
Application**. The requester argues for one; you choose. For an Alias,
every Application means a shared Alias; for a binding, it means the
protocol's default Connection.

### Approve, Dismiss, Reject

- **Approve** applies the Change at once. For an Alias you confirm the alias
  name, the Connection and the model; for a binding, the Connection. You
  can add a note.
- **Dismiss** a Suggestion: aiport stays quiet about it for 7 days unless the
  evidence doubles.
- **Reject** a Change request: your note goes back to the Application, which
  sees it when it lists its requests.

Kinds aiport cannot apply are kept as "not supported" and counted under
**Asked for, not supported yet** (kinds and counts only); they need no
decision.

### Applied changes and Revert

**Applied changes** lists every configuration Change made through aiport,
newest first, with who made it: you, or an approved Suggestion or Change
request. The list survives restarts. **Revert** restores the configuration
file exactly as it was before that Change and makes it live at once. If the
file was edited after a Change, revert the newer Changes first; aiport
refuses rather than overwrite them.

### The sound

The speaker button next to Applied changes turns on a soft sound for when
something new waits in Changes. It is off by default and remembered in this
browser only.

### Allowing Change requests

An Application may request changes only after you allow it: tick **May
request changes** when approving its pairing, or use the toggle on the
Integrate page or in the Applications table on Routing. Until then it gets
`403 change_requests_not_allowed`. Allowed Applications also get hints when
an error is one a Change could fix. Change requests are limited to 20 per
Application per hour.

## Turning the adapting loop off

Set this in `aiport.yaml` and restart aiport:

```yaml
adapting:
  disabled: true
```

With adapting off, aiport detects no Suggestions, Changes shows nothing new,
Change requests answer `409 adapting_off`, no hints are sent, and
`/_connect` and both briefs mark change requests, their MCP tools and
`X-Aiport-Hint` as off with the reason. Forwarding is unchanged, and Revert
still works on earlier Changes. See [configuration](configuration.md) for
the detector thresholds and the check interval.

## Extensions

Extensions are programs you add that take part in deciding about Change
requests:

- a **policy** runs when a Change request arrives and may refuse it; the
  Application gets `403 change_request_refused` with the reason;
- a **guardrail** runs when you approve, before the Change is applied, and
  may block it; Changes then shows which Extension blocked it and why, and
  the item keeps waiting.

List them under `extensions:` in `aiport.yaml` with an absolute path. aiport
runs only listed programs, with a timeout; anything but an explicit allow
denies. The input and output protocol is in [Extensions](extensions.md).

## Suggest to aiport and Share upstream

Two ways to give back to aiport itself, both drafts you check first:

- **Suggest to aiport**: under **Asked for, not supported yet**, for a kind
  of Change Applications asked for that aiport cannot apply.
- **Share upstream**: in the **Extensions** panel of Changes, for an
  Extension you listed. The draft names the hook and the input protocol,
  never the Extension's name, path or settings; you add the program
  yourself after removing anything private.

A draft holds kinds, hooks and anonymous counts only, never configuration,
keys, paths, Application names or request text. A kind named by an
Application is its own words, so check it. Both the title and the body are
editable, and every way out uses the text as you edited it: **Open on
GitHub** (when `upstream.repository` is set), **File with gh…** (after you
confirm, aiport runs `gh issue create` with your own gh login), or **Copy
text**. Nothing is sent on its own.

## Troubleshooting

- **The dashboard does not open.** Check `aiport serve` is running and use
  the exact address it prints. aiport refuses other host names and other
  sites.
- **Forgot the password.** Use **Forgot username or password?** on the
  sign-in screen, paste the token from `./bin/aiport auth token`, then set a
  new password in Settings → Authentication.
- **An application gets `401 unauthorized`.** Its token is missing or no
  longer valid: it should pair again.
- **Calls show as `unknown` in Activity.** The application is not sending its
  own token (for example a shared or old one). Pair it.
- **`503 target_unconfigured`.** No Connection serves that protocol for this
  Application: set a protocol default on Routing, or use an SDK for a
  protocol that is configured. aiport does not translate.
- **`503 target_disabled` or `upstream_key_missing`.** Enable the Connection,
  or paste its key on Providers.
- **`502 upstream_unavailable` / `504 upstream_timeout`.** The provider could
  not be reached in time; **Check connection** on Providers.
- **A provider `404` for a model.** The model name is not one the provider
  knows. Add an Alias, or wait for the Suggestion aiport makes when it keeps
  happening.
- **`403 change_requests_not_allowed`.** Allow the Application on the
  Integrate page if you want its requests.
- **`409 adapting_off`.** The adapting loop is turned off in the
  configuration.
- **Revert refuses with a conflict.** The configuration changed after that
  Change: revert the newer Changes first.
- **Activity is empty.** Check `telemetry.enabled: true` and that the banner
  does not say history is kept for this session only.

The full list of error codes is at `http://127.0.0.1:8787/_connect?format=markdown`.
