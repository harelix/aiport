# aiport manual

This manual is for the Owner: the person who runs aiport on their machine and
decides what it does. It walks through the dashboard page by page. For exact
configuration keys see [configuration](configuration.md); for the protocol
Applications use see [Connect](connect.md); for policies and guardrails see
[Extensions](extensions.md). Words with a capital letter (Application,
Connection, Alias, Change request, Suggestion, Issue, Change, Scope,
Extension) are
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

### Use a name instead of 127.0.0.1

To open aiport at `http://aiport.localhost:8787/` instead of the IP, add to
`aiport.yaml` and restart aiport:

```yaml
hostnames: [aiport.localhost]
```

Browsers resolve any `*.localhost` name to this machine on their own, so the
dashboard works at the name at once. SDKs, scripts and command-line tools
use the system resolver, which may not: add this line to `/etc/hosts` (it
needs `sudo`, so you do it yourself):

```text
127.0.0.1  aiport.localhost
```

aiport still listens on `127.0.0.1` and still refuses every name it was not
given. Applications paired earlier keep working on the IP. Instructions
handed out from now on (the agent instruction, both briefs, `/_connect`, the
base URLs on Routing) use the name, so add the `/etc/hosts` line before
handing them to a coding agent. Names must be lowercase and end in
`.localhost`; the first one listed is the one aiport hands out.
Coding agents started with `aiport agent` keep using the IP.

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
- **Integrate**: your Applications (waiting, approved and rejected),
  onboarding a new one, and the instructions aiport writes for coding
  agents.
- **Compression**: opt-in request compression and whether it hurts answers
  (see [compression](compression.md)).
- **Agents**: coding agents launched through `aiport agent`, each with its
  own token.
- **Changes**: open Issues, what waits for your decision and every
  configuration Change made through aiport, with Revert.
- **Tasks**: a queue of work for coding agents; you own intent and priority.
- **Settings**: sign-in, authentication and the configuration file.

## Connections and keys

A Connection is a provider aiport can forward to: OpenAI, Anthropic, Amazon
Bedrock, Meta Model API, or a local server such as Ollama or llama.cpp. Each
Connection speaks one protocol, `openai` or `anthropic`. Each provider shows
its logo on its card, in the add list and in the Overview's provider
distribution; llama.cpp and custom providers show a neutral icon.

On **Providers**:

- **Add a connection** from a preset, or any other server that speaks the
  OpenAI or Anthropic API natively (remote URLs must use HTTPS); **Remove**
  takes one away. Each is a Change you can revert in Changes.
- Paste a provider API key into a remote Connection's card. It applies at
  once and is stored only in the private `aiport.yaml.env`, never in
  `aiport.yaml`, and is never shown again. Local Connections need no key.
- **Check connection** (or **Check all** per group) sends one authenticated
  `GET /models` to the provider. It generates nothing and costs nothing; it
  shows whether the provider is reachable and accepted the key. A failed
  check says why: the connection error (host name does not resolve,
  connection refused, TLS failed, timed out), or the HTTP status with the
  provider's own error type (for example `HTTP 401, authentication_error`),
  and what to do. A passing check does not prove every model works.
- **Show models** on a card lists the models that Connection serves, using
  the same free `GET /models` listing (OpenAI-, Anthropic- and Ollama-shaped
  answers). After a check, the button shows the count ("Show 12 models").
  aiport keeps only the model ids, in memory, and reuses a list for two
  minutes; **Refresh** lists again, and so does **Check connection**. If the
  listing fails (no key, key rejected, no model list, unreachable), aiport
  treats that Connection's models as unknown, not as "serves nothing".
  Disabled Connections are not listed.
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
2. The request appears at the top of the Applications list on the
   **Integrate** page (and in the Applications table on Routing) with that
   code. Check the application shows the same
   code. Optionally choose a Connection to bind it to and tick **May request
   changes**, then **Approve**, or **Reject**.
3. The application collects its token once. aiport adds it to the
   configuration without a restart; the token is stored in the private
   credential file.

Unapproved requests expire after 10 minutes. The Applications table on
Routing lists every Application with its client id, Connection, own Aliases,
and the **May request changes** toggle.

A request you **Reject** is remembered: the Application learns it was
rejected, and aiport keeps its name, description, code and the time (never a
token), the newest 100, across restarts. See them on the Integrate page with
the **Rejected** filter, where **Forget** removes one. A rejected
Application may ask to pair again.

## The Integrate page and the two briefs

The **Integrate** page is where you connect an application. On a wide
screen it has two columns: your **Applications** and the onboarding wizard
on the left, and a narrow strip with the **Instructions** on the right (on a
narrow screen the strip moves below).

### Applications

- **Pairing requests waiting** are always at the top, each with its code,
  **Approve** (optionally binding a Connection and allowing change requests)
  and **Reject**.
- The list shows each approved Application: name, client id, Connection (or
  the protocol default), the **May request changes** toggle, its last call
  (in the last 7 days, when activity recording is on), its open **Issues**
  and its onboarding progress, such as `3/5`.
- **Issues** shows how many open or resolving Issues the Application has
  and a short cause, such as "Unknown model llama3:8b" (None when there are
  none). Select it to list them under the row: each with its summary, **Open
  in Changes** (opens that Issue on the Changes page, where you approve or
  dismiss it) and **Copy fix for the coding agent** (see Changes, Issues).
- The **Approved / Rejected / All** filter picks what the list shows.
  Rejected Applications are hidden until you choose **Rejected** or **All**;
  each shows when it was rejected, with **Forget**.
- Select an Application to have the wizard follow it. **Onboard a new
  application** makes the wizard follow a new one.

### The instructions

You do not have to write instructions for the coding agent doing the work:
aiport writes them itself, from this gateway's real address and configured
protocols, and serves them at `/_connect`. The strip shows two tiles,
collapsed at first; each shows who the brief is for, **Copy whole brief**
(the brief as markdown, to paste to the coding agent) and its address. Open
a tile to see its steps and snippets, each with its own **Copy**.

- **Application brief** (`/_connect?format=handover&for=application`): for
  the coding agent changing an application so its LLM calls go through
  aiport. It covers pairing, configuration (`AIPORT_BASE_URL`,
  `AIPORT_APP_TOKEN`), pointing the OpenAI or Anthropic SDK at aiport,
  removing provider keys from the application, model names and Aliases,
  errors and retries, logging `X-Aiport-Request-Id`, and how to verify. Its
  **Python / TypeScript / curl** picker switches the SDK snippets. Snippets
  for a protocol this gateway has not configured are marked "not
  configured".
- **Coding agent brief** (`/_connect?format=handover&for=agent`): for coding
  agents working in that application's repository later. It covers the
  project MCP entry (which reads the token from `${AIPORT_APP_TOKEN}`, never
  the value), the line for the repository's `AGENTS.md` (which tells later
  sessions to check `aiport_my_issues` at the start), the tools
  `aiport_capabilities`, `aiport_request_change`, `aiport_my_requests`,
  `aiport_my_issues` and `aiport_issue_fixed` and their HTTP equivalents,
  how to write a good Change request, waiting for your decision, the
  permission, and hints.

`/_connect?format=handover` returns both, one after the other.

### Onboarding an application, step by step

The quickest way: give the coding agent **one instruction** that points at
`/_connect?format=markdown`. That document opens with a **Start here**
section and holds everything for both parts, so the agent needs nothing
else. The **Onboard an application** wizard below the Applications list
walks you through it. It is open while you have no Applications and
whenever a pairing request waits; otherwise it is collapsed, and its header
says where things stand ("1 pairing waiting", or the followed Application
and its progress). The step to do now is highlighted, and each step ticks
itself from live state:

1. **Hand this to the coding agent.** Open the coding agent (Claude Code,
   Codex, Cursor…) in the application's repository and paste the instruction
   shown (**Copy instruction**; **Copy URL only** copies just the address).
   It reads, in substance: "Read `http://127.0.0.1:8787/_connect?format=markdown`
   and follow it to connect this application to aiport: first wire its LLM
   calls through aiport, then add the aiport MCP entry and the `AGENTS.md`
   line; tell me the pairing code and wait while I approve it." The exact
   text, with your gateway's address, is also at the top of the markdown
   document and in the `/_connect` JSON as `agent_prompt`.
2. **Approve the pairing.** The agent asks aiport to pair; the request shows
   up at the top of Applications within seconds with a short code. Check it
   matches the code the agent shows you, then **Approve** (optionally
   binding a Connection and allowing change requests). Requests expire after
   10 minutes. Once approved, the wizard follows that application.
3. **First call arrives.** The agent makes one real call; the step ticks
   when Activity has a call from the application (in the last 7 days).
4. **Allow change requests** (optional), with the toggle in place. You still
   decide each request in Changes.
5. **MCP entry in the repository** (optional): the agent adds it in step 1;
   tick it by hand, or copy it from the **Coding agent brief** tile.

Select an existing application in the list to see how far it got.

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
Suggestions and Change requests, and resolved or dismissed Issues, are kept
(dismissed ones at least 7 days, the quiet period). Open Issues are kept
until they resolve. Applied Changes are never pruned. Deleted rows can linger
in SQLite's files; deletion is not secure erasure.

If the database cannot be opened, aiport says so and keeps the latest 500
records in memory for that session.

Every response carries `X-Aiport-Request-Id`. When an application logs it,
you can find the exact call in Activity.

**Why a request failed.** For a failed request, Activity shows the reason
next to the status, for example `503 · Refused: no Connection for the
Anthropic protocol`. **Details** shows the code, what it means, what to do
and a link to the page that fixes it. aiport keeps only short codes for
this: its own refusal code, a connection error class, and the provider's
error type (such as `rate_limit_error`) when the provider sends one. It
never keeps the provider's error message, the request or the response.
When the latest request on a route failed for a reason in your
configuration, Overview shows one line with the reason and the fix; it goes
away once a request there succeeds.

## Changes

aiport adapts to the Applications that use it, but only with your approval.
**Changes** in the Gateway menu stays quiet: a badge appears only when
something needs a decision.

### Issues

An **Issue** is a problem aiport keeps seeing for one Application, shown at
the top of Changes with its evidence, what aiport found out and the ways to
resolve it. There is one Issue per Application and cause: repeats update
its counts and time window, never a second Issue.

aiport opens an Issue when one Application keeps getting a provider
`404` for the same model it has no Alias for (`adapting.unknown_model`,
default 5 times in 60 minutes). For example, docs-bot asks for
`llama3:8b`, it is not bound to a Connection, and the protocol default is a
hosted Connection that does not serve that model.

- **Evidence**: how many calls failed, the status and the provider's error
  type (such as `404 model_not_found`), the time window, the Connections the
  calls went to and a few request ids to look up in Activity.
- **Diagnosis**: aiport asks each enabled Connection of that protocol for its
  model list (the same free listing the connection check uses, kept for 2
  minutes; nothing is generated and nothing costs money) and says which
  serve the model: exactly one, none, several, or unknown when a listing
  failed (a failed listing never counts as "does not serve").
- **Fix in aiport**: an Alias for that Application only, same model name, to
  the Connection that serves it. With exactly one serving Connection it is
  filled in; with none, several or unknown, choose the Connection (and the
  model, if it has another name there) yourself. **Approve** applies it as an
  ordinary Change you can revert; the Issue becomes **resolving**.
- **Copy fix for the coding agent**: copies the fix in the Application, a
  ready instruction for the coding agent in the Application's repository:
  paste it as-is. It says what is wrong, the evidence (counts, times, the
  Connections, example request ids), what aiport found, what to change in
  the Application (ask for the Alias with a Change request, or send a model
  this gateway serves, listed for that protocol) and how to confirm the fix
  (one call with the Application's own token answers `200`; aiport then
  resolves the Issue by itself). Once the fix in aiport is applied it says
  no code change is needed. It holds no prompts, keys or tokens. **Read the
  fix** shows the same text first. The button is also on the Integrate page,
  under the Application's row, and the text is the same one aiport gives the
  Application's coding agent about its Issues.
- **Dismiss**: the Issue stays quiet for 7 days unless the evidence doubles.

An Issue resolves itself: when the Application's calls for that model
succeed again, aiport marks it **resolved** and records how: `aiport_fix`
(the Change you approved), `application_fix` (the Application's coding
agent fixed it, see below) or `stopped` (the calls succeeded, or stopped
failing for a whole window, without a fix). If you revert the fix, the Issue
opens again and offers the fix again. A resolved Issue reopens when the
error sticks again; only errors after it resolved count. Issues are kept
across restarts and hold only Application ids, the cause, model names,
Connection ids, counts, request ids and fixed advice text: never prompts,
responses, provider messages, keys or tokens.

#### A failing Connection

aiport also opens an Issue when one Application's calls to one Connection
keep failing (`adapting.failing_connection`, default 5 times in 60
minutes): gateway errors and timeouts count; a provider answer below `500`
(an unknown model, a bad request), cancellations and requests aiport
refused never do. For example, docs-bot is bound to a Connection that keeps
answering `502`. There is one Issue per Application and Connection; only
failures to the Connection the Application is bound to now count.

- **Evidence**: how many calls failed, the last status, the time window and
  a few request ids.
- **Fix in aiport**: bind that Application only to a working Connection.
  aiport proposes the configured, enabled Connection with the most
  successful calls in the window; choose another if you like, but never the
  failing one. **Approve** applies it as an ordinary Change you can revert;
  other Applications are not touched. This fix is never applied
  automatically, even with automatic fixes on: moving a whole Application
  to another Connection is always your decision.
- **Fix in the Application**: the request itself is fine; the coding agent
  is told not to retry in a tight loop, to back off and handle `5xx`
  answers and timeouts, and that it can ask for the binding with a Change
  request of kind `bind_application`.

It resolves itself when the Application's calls succeed again, on that
Connection or on the one you bound it to (`aiport_fix` or `stopped`), and
reopens if the failures stick again. Reverting the binding opens it again
with the fix offered again.

The badge on Changes counts an Issue only while its fix waits for you.

### Automatic fixes

Automatic fixes are off by default. Turn them on with the **Automatic
fixes** switch under Applied changes (or `adapting.auto_fix: true` in
aiport.yaml); the switch itself is a Change you can revert. When they are
on, aiport applies an Issue's aiport fix itself, without waiting for you,
only when all of this holds:

- the diagnosis is certain: exactly one enabled Connection serves the model
  (never with none, several or unknown);
- the error is sticky: the Issue's threshold (`adapting.unknown_model`) was
  reached;
- the fix touches only that Application: an Alias for it alone. Binding a
  whole Application to another Connection, shared Aliases, protocol
  defaults or anything for every Application are never automatic;
- you have not refused that fix (below), and the adapting loop is on.

For example, docs-bot keeps getting `404` for `llama3:8b` and only the
local Connection serves it: aiport adds the Alias `llama3:8b` → local for
docs-bot and the Issue becomes resolving. The Change is recorded as made by
**aiport (automatic)**, and Applied changes shows it as "aiport fixed: …"
with **Revert**; the Issue shows "aiport fixed this automatically" with its
own Revert. Guardrail Extensions judge automatic fixes like your approvals.

If you revert an automatic fix, aiport records it and never applies the same
fix automatically again for that Application and cause, also after a
restart; the Issue opens again and its fix waits for your decision, which
you can still approve. A fix you rejected is not applied automatically
either.

### What the Application's coding agent sees

The Application's coding agent hears about its Issues over the connection
it already has with aiport, with the Application's own token (your
management token gets `403`):

- The MCP tool `aiport_my_issues`, or `GET /v1/aiport/issues` (and
  `GET /v1/aiport/issues/{id}` for one), lists that Application's open and
  recent Issues: what is wrong, the evidence, the fix in the Application
  (the same text as "Copy fix for the coding agent") and whether the aiport
  fix is proposed, applied, rejected or reverted. An Application never sees
  another Application's Issues (their ids answer `404`), and never your
  notes.
- The capabilities answer (`aiport_capabilities`) shows `open_issues`, the
  number of its Issues that are open or resolving.
- While an Issue is open, a failing call carries an `X-Aiport-Hint` header
  naming it, so the Application's own logs point the agent to it.
- The `AGENTS.md` line from the Coding agent brief tells later sessions to
  check `aiport_my_issues` at the start.

When the agent has fixed the Application's code (for docs-bot: it now sends
a model this gateway serves), it says so with `aiport_issue_fixed` (or
`POST /v1/aiport/issues/{id}/fixed`). The Issue becomes **resolving** by the
Application fix, and the aiport fix stops waiting for you. aiport does not
take the agent's word for it: the Issue is resolved, as `application_fix`,
only once that Application's calls succeed again. If the error goes on, the
Issue opens again and offers the aiport fix again.

### What waits for your decision

- **Suggestions** come from aiport itself, from patterns in request history,
  with the evidence (how many requests, over which time window). An Alias
  for a model one Application keeps getting `404` for, and binding an
  Application to a working Connection when its calls to one Connection keep
  failing, come as Issues (above). It also suggests pairing when requests keep
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
request, or **aiport (automatic)** for an automatic fix. The list survives restarts. **Revert** restores the configuration
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

With adapting off, aiport detects no Suggestions and opens no Issues, Changes
shows nothing new,
Change requests and `/v1/aiport/issues` answer `409 adapting_off`, the
Issue tools are not offered over MCP, capabilities carry no `open_issues`,
no hints are sent, and `/_connect` and both briefs mark change requests,
Issues, their MCP tools and `X-Aiport-Hint` as off with the reason. Forwarding is unchanged, and Revert
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
  sites; to use a name, see "Use a name instead of 127.0.0.1".
- **An SDK or script cannot reach `aiport.localhost`.** Its resolver does not
  know the name: add `127.0.0.1  aiport.localhost` to `/etc/hosts`, or use
  `127.0.0.1`.
- **Forgot the password.** Use **Forgot username or password?** on the
  sign-in screen, paste the token from `./bin/aiport auth token`, then set a
  new password in Settings → Authentication.
- **An application gets `401 unauthorized`.** Its token is missing or no
  longer valid: it should pair again.
- **Calls show as `unknown` in Activity.** The application is not sending its
  own token (for example a shared or old one). Pair it.
Activity's **Details** names the reason for every failed request. The
reasons, and where each is fixed:

| Activity says | Code | What to do |
| --- | --- | --- |
| Refused: no Connection for the Anthropic (or OpenAI) protocol | `503 target_unconfigured` | Routing → Protocol defaults: pick a Connection for that protocol, or bind the Application to one on Integrate. aiport does not translate between protocols. |
| Refused: the Connection is disabled | `503 target_disabled` | Enable it on Providers, or pick another default on Routing. |
| Refused: the Connection's settings are invalid | `503 target_invalid` | Correct its base URL on Providers. |
| Refused: the Connection does not speak the … protocol | `400 incompatible_protocol` | Routing → Protocol defaults: pick a Connection of that protocol. |
| Refused: no API key for the Connection | `503 upstream_key_missing` | Paste the key on the Connection's card on Providers. |
| Refused: too many requests at once | `429 concurrency_limit` | Retry; raise `proxy.max_concurrent` if it persists. |
| Refused: the request is too large | `413 request_too_large` | Send less, or raise `proxy.max_request_bytes`. |
| Refused: the request body could not be read | `400 request_read_failed` | The Application stopped sending; retry. |
| Refused: unknown compression method | `400 unknown_compression` | Use a method from the Compression page. |
| Refused: the model name is ambiguous | `400 ambiguous_model` | Send one JSON object with one `model` field. |
| Cannot reach the provider: its host name does not resolve | `502 upstream_unavailable`, `dns_failure` | Check the base URL on Providers and the network. |
| Cannot reach the provider: connection refused | `502 upstream_unavailable`, `connection_refused` | Start the local server (Ollama, llama.cpp) or fix the port; **Check connection**. |
| Cannot reach the provider: the secure connection (TLS) failed | `502 upstream_unavailable`, `tls_error` | Check the scheme and host name of the base URL. |
| Cannot reach the provider: connecting timed out / the connection was closed | `502 upstream_unavailable`, `timeout` / `connection_reset` | Check the network and the base URL; retry. |
| Timed out: the provider did not answer in time | `504 upstream_timeout` | Retry; for slow local models raise `proxy.response_header_timeout_ms`. |
| Provider error: the key was rejected | provider `401`/`403` (`authentication_error`, `invalid_api_key`, `permission_error`) | Replace the key on Providers, then **Check connection**. |
| Provider error: unknown model | provider `404` (`not_found_error`, `model_not_found`) | Use a model the provider lists, or add an Alias. When it keeps happening aiport opens an Issue in Changes that says which Connection serves the model. |
| Provider error: rate-limited | provider `429` (`rate_limit_error`) | Wait, or ask the provider for a higher limit. |
| Provider error: out of quota or credit | `insufficient_quota`, `billing_error` | Check the plan and billing with the provider. |
| Provider error: the request was not accepted | provider `400` (`invalid_request_error`) | Check the model name and parameters the Application sends. |
| Provider error: the provider failed | provider `5xx` (`overloaded_error`, `api_error`) | Retry later; check the provider's status page. |
| The response broke off | `incomplete` | Retry; raise `proxy.stream_idle_timeout_ms` for long answers. |
| Cancelled by the Application | `499 client_cancelled` | Nothing to fix. |

A request refused with `401 unauthorized` never reaches Activity (it has no
Application to attribute it to): see the entry above, and pair the
Application on Integrate.

- **`403 change_requests_not_allowed`.** Allow the Application on the
  Integrate page if you want its requests.
- **`409 adapting_off`.** The adapting loop is turned off in the
  configuration.
- **Revert refuses with a conflict.** The configuration changed after that
  Change: revert the newer Changes first.
- **Activity is empty.** Check `telemetry.enabled: true` and that the banner
  does not say history is kept for this session only.

The full list of error codes is at `http://127.0.0.1:8787/_connect?format=markdown`.
