# aiport manual

This manual is for the Owner: the person who runs aiport on their machine and
decides what it does. It walks through the dashboard page by page. For exact
configuration keys see [configuration](configuration.md); for the protocol
Applications use see [Connect](connect.md); for policies and guardrails see
[Extensions](extensions.md). Words with a capital letter (Application,
Connection, Alias, Level, Effort, Catalog, Change request, Suggestion, Issue,
Change, Scope, Extension, Guidance version) are
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
./bin/aiport init          # first run: choose providers, coding agents and Levels
./bin/aiport check         # validate the configuration
./bin/aiport serve         # start the gateway
```

`init` writes `aiport.yaml` and a private credential file next to it
(`aiport.yaml.env`, mode 0600) and never overwrites existing files. Provider
keys are never asked for in the terminal: set them in the dashboard, or in
your shell.

`init` also proposes a Level mapping from the providers you chose, per
protocol: a local Connection for `fast` when you chose one, hosted ones for
`balanced` and `best` (whatever you chose when there is no other kind). The
presets know no model ids, so it asks you for each Level's model; an empty
answer leaves that Level unmapped, and nothing is mapped unless you confirm.
Without a terminal, map them with
`--levels fast=local-connection/llama3:8b,best=hosted-connection/some-model`
(`LEVEL=CONNECTION/MODEL`, the Connection being one of the chosen provider
ids; the protocol is that Connection's) or skip them with `--no-levels`; with
neither flag and no input, no Level is mapped and `init` says so.

`serve` prints the dashboard address and the management token
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
dashboard works at the name at once, and each page has its address there
(`http://aiport.localhost:8787/integrate`).

The name is for your browser. Applications and coding agents are always
given the numeric address (`http://127.0.0.1:8787`): the agent instruction,
both briefs, `/_connect`, the `.aiport` lines on a token card and the MCP
entry all use it, even when hostnames are set. A name needs a lookup the
system resolver may not do, resolves to `::1` first (aiport listens on IPv4),
and a coding agent's sandbox sends it through its network proxy, which
refuses it; `127.0.0.1` has none of these problems. aiport still listens on
`127.0.0.1` and still refuses every name it was not given. Names must be
lowercase and end in `.localhost`.

Other commands: `aiport agent codex|claude --model MODEL` launches a coding
agent through aiport with its own token, `aiport task` manages the task
queue, `aiport version` prints the version.

## A tour of the dashboard

The sidebar has three groups:

- **Monitor**: Overview, Activity and Changes, the pages you use every day.
- **Configure**: Integrate, Providers, Routing, Agents and Compression.
- **Workspace** (at the bottom): Tasks and Settings.

Each page has its own address (`/integrate`, `/changes`, `/manual`…), so
reloading stays on the page, Back and Forward move between pages, and a
link opens the page it names.

The toolbar at the top right has the help button (this manual), the
Light / Dark / System appearance switch and Refresh (on Tasks, also the
chime). The dashboard refreshes itself every 10 seconds while it is open;
Refresh fetches everything at once, for example right after you changed the
configuration file by hand. The bottom of the sidebar shows the address aiport
listens on and Sign out (or "Authentication off" when it is turned off).

- **Overview**: request counts, latency, error rate, hourly traffic, provider
  distribution and the latest requests. When something waits for your
  decision, one line here links to Changes.
- **Activity**: every recorded request, filtered by period, Application
  ("Origin"), Connection ("Target"), outcome and the name sent, with a
  details view per request. Only metadata is recorded: time, route, model,
  effort, status, latency.
- **Providers**: your Connections, grouped Local and Remote.
- **Routing**: protocol defaults, the Levels, the Routing map and the
  Applications table.
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
- **Settings**: sign-in, authentication, the default coding agent for Agent
  runs and where their model calls go, the Connection and model that explain Issues, and the configuration
  file.

## Connections and keys

A Connection is a provider aiport can forward to: OpenAI, Anthropic, Amazon
Bedrock, Meta Model API, or a local server such as Ollama or llama.cpp. Each
Connection speaks one protocol, `openai` or `anthropic`. Each provider shows
its logo on its card, in the add list and in the Overview's provider
distribution; llama.cpp and custom providers show a neutral icon.

On **Providers**, Connections are grouped as Remote APIs and Local. Both
groups start collapsed; each heading shows how many Connections it holds and
which need a key or are unhealthy. Click a heading to open it.

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

### Removing a Connection, and when one fails

You can always remove a Connection. **Remove** on its card opens a dialog
that lists everything still using it: a protocol default, an Application
bound to it, an Alias (shared or an Application's own), Explanations, the
model calls of Agent runs and the compression evaluator. Each item has a
Connection picker, already set to the best match:

1. a Connection of the same protocol that lists the same model (for a
   binding or a protocol default: the one that lists most of the removed
   Connection's models);
2. otherwise the protocol default (the new one, if you are removing the
   default);
3. otherwise none.

Choose **None** to leave an item without a Connection on purpose: a protocol
default is unset, a binding is removed (the Application follows the protocol
defaults again), an Alias is removed (its task name is retired), and
Explanations, Agent runs' model calls or the evaluator are turned off (Agent
runs then use the coding agent's own login).

Below the list the dialog names the Applications whose Catalog the removal
changes (task names removed or moved, a binding or a protocol default they
follow), and updates as you change a picker. aiport itself is among them
when the removal moves what its own model calls (Explanations and Agent
runs) use: `agent_runs.connection`, or the protocol defaults when that is
not set. Its Catalog changes are kept like any Application's, but it gets
no Issue: it has no coding agent to tell. Confirming applies all the
re-pointing and the removal as **one Change**: one revert in **Changes**
restores the Connection and everything that used it. Each affected
Application is told through its Catalog, and an Application the removal
took something it used from gets a Catalog-change Issue (see The Catalog).

Without the dialog (a `DELETE /api/v1/targets/{id}` with no plan), aiport
still refuses to remove a Connection something uses (Explanations and Agent
runs' model calls included), and the message names
what uses it. The dialog reads `GET /api/v1/targets/{id}/dependents` and
sends the plan as `{"plan": {"<item>": "<Connection id or empty>"}}`.

There is no automatic fallback. When a Connection is down or keeps
answering with errors, calls to it fail and the Application gets the
provider's own error. When it keeps happening for one Application, aiport
opens an Issue that offers to bind that Application to a working
Connection; you approve it. Moving an Application to another Connection is
never applied automatically, so aiport never sends calls where you did not
choose.

## Routing and protocol defaults

aiport speaks the providers' native APIs and does not translate between
them: an OpenAI SDK call needs a Connection that speaks the `openai`
protocol, and an Anthropic SDK call one that speaks `anthropic`.

For each call aiport picks the Connection and model in this order:

1. a Level (`fast`, `balanced`, `best`): the Application's own mapping on
   this protocol, else the shared one; a Level mapped on neither is refused
   with `400 level_unmapped` (see [Levels](#levels));
2. the Application's own Alias for the requested model, if it has one (it may
   name another Connection);
3. the Connection the Application is bound to, if any;
4. the protocol default set under **Routing → Protocol defaults**, with the
   shared Aliases of the Connection chosen in 3 or 4.

A Level or an Alias with per-effort models then picks the model for the
call's effort (see [Effort](#effort)). Only the `model` field is replaced.
A protocol with no Connection answers `503 target_unconfigured`; a disabled
Connection answers `503 target_disabled`.

**Routing** also shows the two base URLs Applications use
(`http://127.0.0.1:8787/v1` for the OpenAI protocol, `http://127.0.0.1:8787`
for the Anthropic protocol), the `/_connect` protocol description, and the
Applications table.

### The Routing map

The **Routing map** panel on Routing answers "if this Application sends this
name, which Connection and model serve it?" for every Application at once.
Each Application lists every name it can send: its own task names, the
shared task names of the Connections it reaches, the Levels mapped for it
(its own mapping or the shared one, per protocol), and **Any other model
name**, which goes unchanged to its binding or to the protocol default
(with the number of models that Connection lists). Each row shows the
protocol, the Connection with its provider's logo, and the model aiport
sends upstream, with its per-effort models under it when it has any. For
example, docs-bot sending `fast` on the OpenAI protocol
reaches the local Connection with `llama3:8b`. Choose an Application or a
provider to narrow the map.

The map is built from each Application's Catalog, the same one its
`GET /v1/models` answers, so the two never disagree. The same view for one
Application, with its recent Catalog changes, is under **Catalog** on its
row on Integrate. The pencil on a task name's row edits its per-effort
models (see [Effort](#effort)); a Level's are edited in the Levels panel.
Scripts can read it with the management token at
`GET /api/v1/routing` (`?application=<id>` for one Application, with its
`changes`): Application ids, names, task names, model names and Connection
ids only, never keys or addresses.

## Applications and pairing

An Application is anything that calls aiport with its own token. Each call
is attributed to the Application that sent it, so Activity shows who called
what. Calls without an Application token show as `unknown`.

An Application gets its token by **pairing**:

1. It asks aiport (`POST /v1/aiport/pair` with its name; no token needed) and
   shows you a short code. It may also report the path of its git
   repository on this machine (the Application brief asks the integrating
   coding agent to).
2. The request appears at the top of the Applications list on the
   **Integrate** page (and in the Applications table on Routing) with that
   code. Check the application shows the same
   code. Optionally choose a Connection to bind it to and tick **May request
   changes**, then **Approve**, or **Reject**. A reported repository path
   is shown in a field under the request, with what is wrong with it if it
   is not a git repository here: confirm it, edit it or clear it before you
   approve. aiport keeps only what the field holds when you approve; a
   reported path you did not confirm is never used. A path that does not
   exist or is not the top directory of a git repository is refused with
   the reason, and the request keeps waiting.
3. The application collects its token once. aiport adds it to the
   configuration without a restart; the token is stored in the private
   credential file.

**Where the token lives: `.aiport`.** Each application keeps its aiport
settings in a file named `.aiport` at the root of its repository, listed in
its `.gitignore` and never committed:

```
AIPORT_BASE_URL=http://aiport.localhost:8787
AIPORT_APP_TOKEN=<its token>
```

The application fills its environment from it at startup (variables already
set win), and its coding agents read it too: the project MCP entry's
`headersHelper` takes the token from `AIPORT_APP_TOKEN` or else `.aiport`, so
nothing has to be exported before the agent starts. When a token is lost or
revoked (aiport answers `401`), click **New token** on the application's row
on Integrate, **Copy .aiport lines** and paste them over its `.aiport`; the
old token stops working at once. This is guidance version 3: coding agents
of applications integrated earlier are told to move their token there.

**One request per application.** An application that asks again while its
request waits (the same repository, or the same name, ignoring case and
punctuation) gets that same request back, with the same code, so the code
you see stays valid; its earlier and newer poll paths both work. A request
that matches an Application you already have is marked **Already connected
as …**; its main button, **Give … a new token**, approves it as that
Application (it receives a new token and its old one stops working) instead
of adding a duplicate such as `docs-bot-2`. **Add as separate** still adds a
new Application. Over the API, approve with `{"replace": "<id>"}`.

Requests wait 30 minutes for you, and survive an aiport restart: a waiting
request can still be approved by its code, and an approved application can
still collect its token (for 24 hours) after a restart. A poll on a request
that ran out answers `expired`, so the application knows to ask again. The Applications table on
Routing lists every Application with its client id, Connection, own Aliases,
and the **May request changes** toggle.

### Issuing a token yourself, and removing an Application

You can also skip pairing and hand an application its token yourself, for
example when its coding agent is not running, or you would rather paste the
token into the application's `.aiport` file:

- **Add with a token** (Integrate, next to *Onboard a new application*): name
  the Application and optionally choose a Connection. aiport adds it and shows
  its token once, as the two lines to paste into its `.aiport`:
  `AIPORT_BASE_URL` (this gateway's address) and `AIPORT_APP_TOKEN`.
- **New token** (the Token column of an Application's row): after you confirm,
  aiport issues a new token and shows it once the same way. The old token
  stops working at once, whether it came from pairing or from here. Issuing a
  token changes no configuration, so it is not listed in Changes.
- **Remove** (same column): names what goes with the Application (its own
  Aliases, its binding, its coding agent settings, its permission to request
  changes), then removes it as one Change and deletes its token, which stops
  working at once. Its activity history is kept. Reverting that Change in
  **Changes** brings the Application back **without a token**: its row shows
  *No token* and its calls are refused until you **Issue token**. `serve` and
  `aiport check` name such Applications too.

A token set in aiport's own environment when it was started (exported in
the shell rather than kept in the private credential file) cannot be
replaced or deleted from the dashboard: change it where aiport is started and
restart it.

The same is available with the management token: `POST /api/v1/origins`
with `{name, target?}`, `POST /api/v1/origins/{id}/token` and
`DELETE /api/v1/origins/{id}` (send `X-Aiport-CSRF: 1` with each).
A token is in the response once and is never shown again.

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
  (in the last 7 days, when activity recording is on), its open **Issues**,
  its coding **Agent** and its onboarding progress, such as `3/6`.
- **Agent** names the Application's coding agent once it has one and a
  repository path (**Connect** otherwise). Select it to change them in step
  6 of the wizard.
- **Issues** shows how many open or resolving Issues the Application has
  and a short cause, such as "Unknown model llama3:8b" (None when there are
  none). Select it to list them under the row: each with its summary, **Open
  in Changes** (opens that Issue on the Changes page, where you approve or
  dismiss it) and **Copy fix for the Coding agent** (see Changes, Issues).
- **Guidance** shows whether the Application's coding agent follows
  aiport's current guidance (see Guidance versions below): **Up to date**,
  **Behind v1 → v2**, or **Unknown** when it never confirmed one (every
  Application integrated before guidance versions existed). While it is not
  up to date, **Copy refresh instruction** copies one paragraph to paste to
  that coding agent. "Sends Levels" is only a hint that its code may already
  use Levels; it never counts as confirming.
- **Catalog** shows, under the row, the Application's Catalog: every name it
  can send with the Connection and model that serve it (as on the Routing
  map), the models its Connections list, and its recent Catalog changes
  (newest first), the same changes the Application reads at
  `/v1/aiport/catalog`.
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
  removing provider keys from the application, model names (send a Level or
  a task name, never a provider's model id; read `GET /v1/models` for the
  names; `level_unmapped`), saying how hard a call is with the
  `X-Aiport-Effort` header (and that the provider's own effort field is read,
  never changed), watching the `X-Aiport-Catalog` header, errors and retries, logging `X-Aiport-Request-Id`, Issues (what
  they are, that the Coding agent reads them with `aiport_my_issues`, that a
  Change that takes something away opens a `catalog_changed` one, and that
  aiport may fix routing itself as a revertable Change), and how to verify. Its
  **Python / TypeScript / curl** picker switches the SDK snippets. Snippets
  for a protocol this gateway has not configured are marked "not
  configured".
- **Coding agent brief** (`/_connect?format=handover&for=agent`): for coding
  agents working in that application's repository later. It covers the
  project MCP entry (which reads the token from `${AIPORT_APP_TOKEN}`, never
  the value), the line for the repository's `AGENTS.md` (which tells later
  sessions to send a Level or task name, never a provider's model id, to
  state effort with `X-Aiport-Effort`, and to check `aiport_catalog`,
  `aiport_my_issues` and `aiport_guidance` at the start, calling
  `aiport_guidance_done` after updating),
  the tools `aiport_capabilities`, `aiport_catalog`, `aiport_guidance`,
  `aiport_guidance_done`, `aiport_request_change`, `aiport_my_requests`, `aiport_my_issues` and
  `aiport_issue_fixed` and their HTTP equivalents, checking the Catalog
  (and re-reading it when `X-Aiport-Catalog` changes), Issues and guidance
  at the start of a session, naming the need (a Level or task name and an
  effort) rather than a provider, how to write a good Change request, waiting for your decision, the permission, and
  hints.

`/_connect?format=handover` returns both, one after the other. The MCP
`initialize` instructions an agent gets from aiport say the same as the
`AGENTS.md` line.


**Communication standard.** The first card in Instructions is
`/_connect?format=standard`: one document with every rule an integration
follows, and a checklist. It covers the token in `.aiport`, the SDK pointed at
aiport, Levels and task names instead of provider model ids,
`X-Aiport-Effort`, `X-Aiport-Module` naming the module of the application each
call serves, a thumbs up and down on every LLM answer shown to a person sent to
`/v1/aiport/feedback`, the error fixes, and the MCP tools. **Copy instruction**
gives the prompt that hands it to a coding agent, for a new application or one
already connected.
### Guidance versions

aiport numbers the guidance it gives coding agents (the briefs, the
`AGENTS.md` line and the MCP instructions) and raises the number when that
guidance changes in a way agents must act on; each version has a short
change log. Version 1 is the project setup (the MCP entry, the `AGENTS.md`
line, the Catalog and Issues). Version 2 is Levels and effort: send a Level or
task name instead of a provider's model id, add `X-Aiport-Effort` where calls
differ in difficulty, replace the `AGENTS.md` line with the new one (it adds
Levels, effort and `aiport_guidance`), and check `aiport_guidance` at the
start of each session. For each Application aiport
remembers the version its coding agent last confirmed (in the database
next to the configuration; none for Applications integrated earlier).

While an Application is behind, its coding agent is told without you: the
MCP `initialize` instructions start with a one-line reminder, every aiport
MCP tool answer ends with it (except the answers of `aiport_guidance` and
`aiport_guidance_done` themselves, which are the guidance), and `aiport_capabilities` shows
`guidance: {current, confirmed}`. The agent calls `aiport_guidance`
(`GET /v1/aiport/guidance`) for what changed since its version, the current
`AGENTS.md` line and the steps, does them, then calls
`aiport_guidance_done` (`POST /v1/aiport/guidance/done`). Both take the
Application's own token and only ever show or change that Application's
state. The Integrate page shows each Application's state in the
**Guidance** column, and **Overview** shows one quiet line while any
Application has not confirmed the current version. An Application that never
confirmed any version counts as not confirmed; the line tells the two apart
("behind": confirmed an older version; "never confirmed": none yet), as does
`never_confirmed` next to `behind` in `GET /api/v1/guidance`.

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
   binding a Connection and allowing change requests). Requests wait 30
   minutes. Once approved, the wizard follows that application.
3. **First call arrives.** The agent makes one real call; the step ticks
   when Activity has a call from the application (in the last 7 days).
4. **Allow change requests** (optional), with the toggle in place. You still
   decide each request in Changes.
5. **MCP entry in the repository** (optional): the agent adds it in step 1;
   tick it by hand, or copy it from the **Coding agent brief** tile.
6. **Connect a coding agent** (optional): see below. The step ticks when
   the application has a coding agent (its own or the default) and a
   repository path.

Select an existing application in the list to see how far it got.

### Connecting a coding agent

**Why connect one.** An Issue tells you *that* an Application keeps
failing and offers the fixes aiport can see from its side. A coding agent
connected to the Application's repository lets aiport go further, on your
click: **Explain** reads the code and tells you the root cause, and
**Fix it** makes the change on a new branch and shows you the diff to keep
or discard. Explain only reads; nothing runs without your click; aiport
never commits to your branch, merges or pushes. Without an agent, aiport
still explains, more briefly, from the Issue's evidence alone (with the
Connection and model chosen under Explanations in Settings).

aiport can start a Coding agent installed on this machine to work on an
Application's Issues, only when you click (Agent runs). Each Application
can have its own agent, **Claude Code**, **Codex** or **Gemini CLI**, and
the path of its git repository; set them in step 6 of the wizard (or from
the **Agent** column) and **Save**. Each save is one revertable Change on
the Changes page. **Default coding agent** on the Settings page is the
agent of every Application that has none of its own.

- **What aiport looks for.** aiport talks to agents through the Agent
  Client Protocol (ACP). Gemini CLI speaks it itself (`gemini --acp`);
  Claude Code and Codex need an ACP adapter: `claude-agent-acp` (npm
  package `@agentclientprotocol/claude-agent-acp`) and `codex-acp`
  (`@agentclientprotocol/codex-acp`). aiport finds programs by absolute
  path, in the absolute directories of its own `PATH` and the usual
  install places (`~/.local/bin`, `~/.npm-global/bin`, `/opt/homebrew/bin`,
  `/usr/local/bin` and a few more); it does not run them to look.
- **Nothing is installed for you.** When an agent or its adapter is
  missing, the step shows the exact command to install it, such as
  `npm install -g @agentclientprotocol/codex-acp`; run it yourself, then
  refresh. An agent that is not installed cannot be chosen.
- **The repository path** must be absolute and the top directory of a git
  repository (where `.git` is); otherwise it is refused with the reason.
  aiport stores it with symbolic links resolved. Agent runs never work
  outside it.

In `aiport.yaml` this is `agent` and `repository` on the Application under
`origins`, and `agent_runs.default_agent`:

```yaml
agent_runs:
  default_agent: claude
origins:
  - id: docs-bot
    name: Docs Bot
    token_env: AIPORT_APP_DOCS_BOT_TOKEN
    agent: gemini
    repository: /home/you/src/docs-bot
```

`GET /api/v1/agents` lists what aiport found (`installed`, `ready`, the
ACP entry point, and `install` when something is missing).

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
list its own Aliases with `GET /v1/aiport/capabilities`, and every task name
it can use with `GET /v1/models` (see The Catalog).

### Task names for every Application, with a fixed reasoning effort

A shared Alias normally applies only to calls that already go to its
Connection. Set `everywhere: true` on it and every Application can send the
name, whichever Connection its calls otherwise go to: the name routes to its
own Connection, like a Level. Set `reasoning: low`, `medium` or `high` and
aiport sends that reasoning effort to the provider on every call for the name
whose body sets none of its own, whatever `X-Aiport-Effort` says. A typical use
is a cheaper model for routine work (`classify`, `tldr`, `extract`) at a fixed
effort, while the Levels stay on the models for harder work. Add, change and
remove these with `PUT` and `DELETE /api/v1/task-names/{name}`
(`{connection, model, effort?, everywhere?, reasoning?}`); each is a
revertable Change, and each Application's Catalog lists the name with its
`reasoning`.

## Levels

Every aiport shares three model names, the **Levels** `fast`, `balanced` and
`best`. An Application sends one as its model to say what proficiency it
needs, without naming a provider; you decide what each Level means here, per
protocol. A Level is a reserved Alias name:

- **The shared mapping** is an entry under `aliases:` whose `alias` is the
  Level, on a Connection of the protocol it serves. Each Level is mapped at
  most once per protocol.
- **An Application's own mapping** is the same entry under its
  `origins[].aliases`, and comes before the shared one, on that protocol only.

```yaml
aliases:
  - {target: hosted, alias: best, model: big-model}         # OpenAI protocol
  - {target: local, alias: fast, model: 'llama3:8b'}        # OpenAI protocol
  - {target: team, alias: balanced, model: team-model}      # Anthropic protocol
origins:
  - id: docs-bot
    # ...
    aliases:
      - {alias: best, target: local, model: 'llama3:70b'}   # its own best, OpenAI only
```

Unlike a plain shared Alias, which only renames a model on the Connection a
request already goes to, a Level mapping chooses the Connection, also for an
Application bound to another one or with no protocol default. Only the
`model` field is replaced; the rest of the request goes to the provider
unchanged. A Level not mapped on the request's protocol is refused with
`400 level_unmapped`, naming the Level and protocol: aiport never quietly
uses another model. Mapped Levels are listed in each Application's
`GET /v1/models` (on the protocol they are mapped on) and in its Catalog, and
changing a mapping is a Change you can revert and a Catalog change the
Applications are told about. Removing a Connection a Level is mapped on asks
for a replacement Connection of the same protocol.

A request that nothing could route (no binding or protocol default for its
protocol, no task name of its own there and no Level mapped for it) is
refused at once, before aiport reads its body. Once any Level is mapped for
an Application, or while its Connection is only switched off, aiport reads
the model first, so a Level is still refused as `level_unmapped`, naming it.

**Upgrading: Aliases named like a Level.** If your configuration already had
an Alias called `fast`, `balanced` or `best`, it is now that Level's mapping
on its Connection's protocol. A shared one now chooses the Connection for
every Application's calls for that name on that protocol, whatever they are
bound to (before, it only renamed a model on the Connection a call already
went to); an Application's own one applies on its Connection's protocol
only. `aiport check` and the startup log print one notice per such entry,
naming it and its Connection. To keep it a plain Alias, rename it (its
`alias:`) to a name that is not a Level and send the new name; to keep it as
the Level, nothing needs doing.

`aiport init` proposes a first mapping (see [Install and run](#install-and-run));
the briefs and the `AGENTS.md` line tell every Application's coding agent to
send a Level or a task name, never a provider's model id. The
[Routing map](#the-routing-map) shows what each Level reaches for each
Application, and Activity shows which name and effort each call asked for and
what served it.

You edit Levels in the dashboard, without touching the file:

- **Routing → Levels** shows each Level on each protocol. The pencil opens
  its mapping: a Connection of that protocol, a model, and optional
  per-effort models (see Effort below). **Clear** removes the shared mapping.
  A cell also says how many Applications have their own.
- **Integrate → Applications → Levels** (the toggle under an Application's
  name, next to **Catalog**) shows the same grid for that Application: its own mappings, marked *Own*, and the
  shared ones it falls back to. The pencil gives it its own mapping; **Use
  the shared mapping** removes it again.

Every save is one Change: revert it in Changes to get the previous mapping
back. The same edits are in the management API, with the management token
and the `X-Aiport-CSRF: 1` header: `PUT /api/v1/levels/{protocol}/{level}`
with `{application?, connection, model, effort?}`, `DELETE` on the same path
with `{application?}`, `PUT /api/v1/aliases/effort` with
`{application?, alias, connection?, effort}` for the per-effort models of any
Alias, and `GET /api/v1/levels` to read them all. Leaving `application` out
edits the shared mapping.

When an Application keeps asking for a Level that is not mapped, aiport
opens an Issue for it (see [An unmapped Level](#an-unmapped-level)).

### Effort

An Application may also say how hard one call is: `low`, `medium` or
`high`. aiport uses effort only to pick the model, so `balanced` with high
effort can go to a stronger model than `balanced` with low effort. Any Alias,
a Level or a task name, shared or an Application's own, may carry optional
per-effort models on its own Connection:

```yaml
aliases:
  - {target: hosted, alias: balanced, model: mid-model, effort: {low: small-model, high: big-model}}
origins:
  - id: docs-bot
    # ...
    aliases:
      - {alias: summarize, target: local, model: 'llama3:8b', effort: {high: 'llama3:70b'}}
```

An effort without its own model uses the Alias's `model`. The effort of a
call comes from, in this order:

1. The `X-Aiport-Effort` header (`low`, `medium` or `high`), the one the
   briefs teach. It also accepts the provider words below (`none` and
   `minimal` count as low, `xhigh` and `max` as high). It is never sent on to
   the provider. A value that is not an effort counts as none.
2. The provider's own effort field, which aiport reads but never changes:
   `reasoning_effort` on OpenAI Chat Completions, `reasoning.effort` on OpenAI
   Responses, `output_config.effort` on Anthropic Messages. `none`, `minimal`
   and `low` count as low, `medium` as medium, `high`, `xhigh` and `max` as
   high. Numeric thinking budgets (`budget_tokens`) are never translated.
3. Otherwise `medium`.

As with every Alias, only the `model` field is replaced: the request,
including its effort field, reaches the provider unchanged. Per-effort models
are listed in the Catalog (`effort` on a task name), and changing them is a
Catalog change. Activity shows the effort of each call and where it came
from (see below).

A Level's per-effort models are edited in its mapping on Routing or
Integrate (empty fields use the mapping's model). Over the API, a `PUT` on a
Level without `effort` keeps them while the mapping stays on the same
Connection and drops them when it moves to another (they name models of
the old one). The same holds when a removal plan moves a Level or a task
name to another Connection: it keeps its model and loses its per-effort
models, which you set again on the new Connection.

A task name's per-effort models are edited on its row of the
[Routing map](#the-routing-map) (and in an Application's Catalog on
Integrate): the pencil opens a model for low, medium and high on the task
name's own Connection; an empty field uses its model, and **Clear** removes
them all. A shared task name's per-effort models apply to every Application.
Each save is one Change you can revert. Over the API it is
`PUT /api/v1/aliases/effort` (a shared task name is found by its name and
`connection`, an Application's own by its name), and in `aiport.yaml` the
`effort:` of the Alias.

## The Catalog

An Application's **Catalog** is what it can use through aiport: the Levels
mapped for it, its task names (its own Aliases, then the shared Aliases of
the Connections it reaches) and the models those Connections list, with a
short **version**.
You decide what is in it; aiport tells the Application, so it never has to
guess a model name or learn from a provider's `404` that something changed.
For example, docs-bot has its own Alias `summarize` and shared Alias
`review-code`, both on the local Connection it is bound to, which also serves
`llama3:8b`: its
Catalog lists `summarize`, `review-code` and `llama3:8b`.

The briefs tell every Application and its coding agent how to use it (see **The instructions**
under the Integrate page): send a Level or task name, never a provider's
model id, read `GET /v1/models` for the names, watch the `X-Aiport-Catalog`
header, and check `aiport_catalog`, `aiport_my_issues` and `aiport_guidance`
at the start of each session.

### What an Application reads

All of these answer only about the Application whose token is sent: each
Application sees only its own Catalog, and keys, addresses, prompts and
tokens never appear.

- **`GET /v1/models`**, where OpenAI and Anthropic SDKs and coding agents
  already look for models, answers in the OpenAI list shape, or in the
  Anthropic shape when the request carries `anthropic-version` (Anthropic
  SDKs may send the token in `X-Api-Key`). Your management token gets `403
  application_token_required`: a Catalog belongs to one Application. Only
  names usable on that protocol are listed. Mapped Levels and task names come
  first, marked as aiport's (`owned_by: "aiport"`, or "(aiport Level)" and
  "(aiport task name)" in the display name); then
  the models of the Connection the Application is bound to, or of the
  protocol defaults. A Connection whose model list cannot be read adds
  nothing (lists are the same free, cached ones **Show models** uses).
- **`X-Aiport-Catalog: <version>`** is on every answer to a request with an
  Application token, forwarded calls and errors included; the provider's
  answer is otherwise unchanged. The version changes when the Application's
  task names or where they go, its binding, or the protocol defaults it
  follows change, and only then (a provider adding or retiring a model does
  not change it). An Application that sees a new version reads
  `/v1/models` again.
- **`GET /v1/aiport/catalog`** (or the MCP tool `aiport_catalog`, same
  answer) returns `catalog` (task names with their Connection and model,
  Levels among them with `level: true` and their `protocol`, and per-effort
  models as `effort`,
  the binding or the protocol defaults followed, the Connections reached,
  their models and the version) and `changes`, its recent Catalog changes,
  newest first (see below). Your management token gets `403
  application_token_required`.
- **The capabilities answer** (`GET /v1/aiport/capabilities`) carries
  `catalog: {version, changed_at_ms}`: the version and the time of the
  newest kept Catalog change (`null` when none is kept).

### The change log

Every configuration Change goes through one place, and on each one aiport
works out what it changed in every Application's Catalog: approving or
reverting an Alias, mapping a Level or changing per-effort models, binding
an Application, changing a protocol default, removing a Connection. A revert is a Catalog change too (with `revert_of`
naming the Change it undid). Each **Catalog change** records the Change id
and time, the version before and after, task names `removed`, `added` or
`moved` (`from` → `to`), the binding or protocol defaults that changed
(`routing`), and the Connections the Application no longer or newly reaches
(`connections_removed`, `connections_added`). An Application the Change did
not affect gets nothing. aiport's own model calls (the built-in Application
`aiport`, which follows `agent_runs.connection`, else the protocol defaults)
get Catalog changes recorded too.

Catalog changes are kept as long as request history
(`telemetry.retention_days`) and survive a restart. They hold Application
ids, task names, model names and Connection ids only, never keys,
addresses, prompts, responses or tokens.

### When a Change takes something away

Removing a task name an Application used, or moving its binding or a
protocol default it followed, does more than change the version: aiport
opens a **Catalog-change Issue** (cause `catalog_changed`) for that
Application, saying what changed and what it can still use. "Used" means
its calls sent that task name, or a model name that went through that
binding or default, within request history; an Application that never did
gets only the Catalog change. It needs no
decision from you and resolves itself; see
**A Catalog change** under Issues. Before a removal, the
removal dialog shows which Applications it will affect (see
**Removing a Connection** under Connections and keys).

All of the Catalog works with adapting turned off: it is information, not a
Change. Only the Catalog-change Issues follow the adapting switch.

## Feedback

Every integration names the module of the application each model call is
for with the `X-Aiport-Module` header (`tldr`, `classify`, `analysis`,
`insights`...) and gives every LLM answer shown to a person a thumbs up and a
thumbs down. On a vote the application sends `POST /v1/aiport/feedback` with
`{"request_id": "<X-Aiport-Request-Id of the answer>", "vote": "up"}` or
`"down"` (or a `score` from 0 to 1; `module` names it when the call did
not). The communication standard (below) and guidance version 4 asks every connected application's coding agent to add
both; prompts and answers are never sent.

The **Feedback** page shows, for the last 7, 30 or 90 days, each
Application's votes per module, and inside each module per name sent (a
Level, task name or model id), model and effort: calls, votes, and the share
of up-votes as a bar (green from 75%, amber from 50%, red below). A module with
fewer than 5 votes is marked "few votes": too few to judge. With two or more
judged modules the page names the best and the weakest; change the weakest
module's Level, task name or effort under Routing and watch its share. The
same report is `GET /api/v1/feedback?hours=N`.

**Suggestions from the votes.** Once a module's calls for one name have at
least 10 votes, aiport reads them against your Levels, and lists what it finds
under **Suggestions from the votes**:

- **Better elsewhere**: the module is below 60% up, and a module of the same
  name (in another application, or on another model) gets at least 20 points
  more on the same protocol. aiport suggests that model.
- **One Level up**: below 60% up with no such evidence; aiport suggests the
  next Level's model (`fast` to `balanced`, `balanced` to `best`), untested for
  this module.
- **Cheaper may do**: 90% up or more over at least 20 votes on `best` or
  `balanced`; aiport suggests the Level below, untested for this module.

aiport never moves a module by itself. **Apply** adds a **module route**: that
application's calls for that module that send that name go to the suggested
model, ahead of every other route, while its other modules keep theirs. It is
one Change (reason included) that you can revert on Changes, and each
application's panel lists its module routes with **Remove**. Then watch the
module's share of up-votes: the new model shows as its own row. Module routes
are `origins[].module_routes` in `aiport.yaml`, and `PUT`/`DELETE
/api/v1/module-routes/{application}/{module}/{name}` (`{connection, model,
why?}`). Votes need request history
(`telemetry.enabled`) and count only calls still retained.

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

**What was asked and what served it.** For each model call Activity keeps
the name the Application sent (a Level, a task name or a model id), its
effort and where that came from (the `X-Aiport-Effort` header, the
provider's effort field, or neither, so `medium`), and the Connection and
model aiport sent it to. The Model column shows `balanced` with
`chat/completions → small-model · low` beneath it; **Details** says it in a
sentence, for example "Asked for balanced at low effort (from the
X-Aiport-Effort header); served by hosted with small-model." A refused call
has no serving model. The **Name** filter shows only calls that sent one
name, so you can check a Level or per-effort mapping works and compare the
models behind it. Older records, from before this was kept, show none of it.

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
**Changes** in the Monitor menu stays quiet: a badge appears only when
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
- **Copy fix for the Coding agent**: copies the fix in the Application, a
  ready instruction for the Coding agent in the Application's repository:
  paste it as-is. It says what is wrong, the evidence (counts, times, the
  Connections, example request ids), what aiport found, what to change in
  the Application (ask for the Alias with a Change request, or send a model
  this gateway serves, listed for that protocol) and how to confirm the fix
  (one call with the Application's own token answers `200`; aiport then
  resolves the Issue by itself). Once the fix in aiport is applied it says
  no code change is needed. It holds no prompts, keys or tokens. **Read the
  fix** shows the same text first. The button is also on the Integrate page,
  under the Application's row, and the text is the same one aiport gives the
  Application's Coding agent about its Issues.
- **Dismiss**: the Issue stays quiet for 7 days unless the evidence doubles.
  You can dismiss any Issue that is not resolved or dismissed yet, whatever
  its fix stands at (waiting, approved, rejected or reverted); an approved
  fix stays applied until you revert its Change.
- **Reject** (`POST /api/v1/issues/{id}` with `action: reject`): turns the
  aiport fix down and leaves the Issue as it is. The dashboard offers
  Approve and Dismiss; Reject is there for scripts.

An Issue resolves itself: when the Application's calls for that model
succeed again, aiport marks it **resolved** and records how: `aiport_fix`
(the Change you approved), `application_fix` (the Application's Coding
agent fixed it, see below) or `stopped` (the calls succeeded, or stopped
failing for a whole window, without a fix). A fix is credited only when a
call shows it works: if the Application makes no calls for that model for a
whole window after you approved the fix, the Issue resolves as `stopped`,
not `aiport_fix`. If you revert the fix, the Issue
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
- **Fix in the Application**: the request itself is fine; the Coding agent
  is told not to retry in a tight loop, to back off and handle `5xx`
  answers and timeouts, and that it can ask for the binding with a Change
  request of kind `bind_application`.

It resolves itself when the Application's calls succeed again, on that
Connection or on the one you bound it to (`aiport_fix` or `stopped`), and
reopens if the failures stick again. Reverting the binding opens it again
with the fix offered again.

#### An unmapped Level

aiport opens an Issue when one Application's calls for one Level keep being
refused with `level_unmapped` (`adapting.level_unmapped`, default 5 times
in 60 minutes). For example, docs-bot sends `best` on the OpenAI protocol
and no `best` is mapped there, neither for docs-bot nor shared. There is one
Issue per Application, Level and protocol: `best` unmapped on both protocols
opens two, each with its fix on its own protocol.

- **Evidence**: how many calls were refused, the time window and a few
  request ids; the summary names the protocol.
- **Fix in aiport**: map the Level for that Application only, on the
  protocol its calls were refused on. aiport proposes the Connection its
  other calls on that protocol go to (its binding, else the protocol
  default) and leaves the model to you: it never makes up a model id.
  **Approve** applies it as an ordinary Change you can revert (and a Catalog
  change); a Connection of the other protocol is refused. It is never
  applied automatically.
- **Fix in the Application**: ask for the mapping with a Change request of
  kind `application_alias` whose alias is the Level, or send another Level
  or task name the gateway maps.

It resolves itself when the Application's calls for that Level succeed, and
reopens if the refusals stick again; reverting the mapping opens it again
with the fix offered again. With the adapting loop off, nothing is counted
and no Issue opens.

#### A Catalog change

When a Change takes something away from an Application's Catalog (see The
Catalog: a task name it had, or the binding or a protocol default it
follows moves; for
example you remove a Connection and leave docs-bot's Alias `summarize`
without one), aiport opens one Issue of cause `catalog_changed` for that
Application, one per Change, when the Application used what went away: within
request history (`telemetry.retention_days`) it sent the removed task name,
or a model name (not a task name) that went to the Connection its binding or
the moved protocol default pointed at. An Application that never sent it,
and the built-in `aiport` Application, get the Catalog change recorded but
no Issue. Applications the Change took nothing from get none, and a Change that is an Issue's own aiport fix (or its revert) opens
none: that Issue already says so.

- **What it says**: what changed (task names removed or moved, the binding
  or protocol default and where it goes now), what the Application can still
  use (its task names and the Connections whose models it reaches), and how
  to confirm. Changes and Integrate show it with nothing to approve: it needs
  no decision from you, and it never lights the Changes badge.
- **Delivery**: like every Issue: `GET /v1/aiport/issues`, the MCP tool
  `aiport_my_issues`, the `open_issues` count in the capabilities answer, and
  `X-Aiport-Hint` on a failing call that sent a removed task name.
- **Resolution**: it resolves itself when the Application's calls succeed
  after the change (and after its last failing call with a removed name),
  at once when its Coding agent reports it fixed (`aiport_issue_fixed`:
  there is nothing in aiport to confirm), or when you revert the Change.
  You can also dismiss it.

None are opened while adapting is off; the Catalog itself is still disclosed.

The badge on Changes counts an Issue only while its fix waits for you.

#### Explain

**Explain** on an Issue (on Changes, and under the Application's row on
Integrate) writes its root cause in a few sentences. Without a Coding agent
connected to the Application, aiport asks a Connection and model you choose
once in **Settings → Explanations** (`adapting.explain` in `aiport.yaml`,
for example a local Connection serving `llama3:8b`). Until you choose one,
Explain is disabled with a hint saying so.

- **What is sent**: only the Issue's evidence: the cause, the status and the
  provider's error type, the counts and time window, example request ids,
  the model name, the Connection ids and aiport's diagnosis (which
  Connections list the model, and the models they list). Never keys,
  tokens, the Application's id or name, prompts or provider messages. The
  one thing in it an Application wrote is the model name it asked for,
  and only when that name looks like a model id (at most 128 letters,
  digits and `. _ : / @ + -`); any other name is replaced by "(a model
  name that is withheld because it contains unusual characters)", and
  Connection ids, model lists and the error type pass the same check. The
  same holds for what a Coding agent is told.
- **Streaming**: the words appear as the model writes them. When it
  finishes, the explanation is saved with the Issue, kept across restarts,
  and shown with who wrote it. **Explain again** replaces it; if writing
  fails or takes longer than 2 minutes, nothing is saved and an earlier
  explanation stays. One explanation per Application is written at a time.
- **Carried to the Coding agent**: the saved explanation is the "Root
  cause" paragraph at the start of "Copy fix for the Coding agent", and
  `aiport_my_issues` (and `GET /v1/aiport/issues`) give it to the
  Application's Coding agent, in that text and as `explanation`.
- **Activity**: the call shows in Activity under the built-in Application
  `aiport` (no configured Application may use that id), with its
  Connection, model, status and latency, like any other call. A remote
  Connection bills explanations to its key; a local one keeps them on this
  machine.

For scripts: `PUT /api/v1/adapting/explainer` with `{"connection": "local",
"model": "llama3:8b"}` chooses who writes (a revertable Change; both empty
turns Explain off), and `POST /api/v1/issues/{id}/explain` streams JSON
lines: `start` (who writes), `text` (the next words), then `done` (the
Issue with its saved explanation) or `error`.

#### Explain by a Coding agent

When the Application has a Coding agent and a repository path (see
"Connecting a coding agent") and the agent's ACP entry point is installed,
**Explain** starts an **Agent run** instead: the agent reads the
Application's code and writes the root cause, usually with the file and
the change to make. The button's tooltip names the agent; otherwise the
Connection from Settings explains, as above.

- **How it starts**: only on your click, never by itself (not even with
  automatic fixes on), and never with adapting off. aiport starts the ACP
  entry point it found, by absolute path, never through a shell, with a
  minimal environment: `PATH` (the agent's own
  directory first, then the absolute directories of aiport's `PATH`, so an
  npm-installed adapter finds `node`), `HOME`, `TMPDIR`, `LANG`, `LC_ALL`,
  `USER`, `LOGNAME`, and, for Claude Code, the variables below. Never
  aiport's tokens or provider keys, and no MCP connection to aiport.
- **A throwaway copy, never your working copy**: the agent reads a git
  worktree of the repository that aiport makes for the run, detached at
  your current commit (`HEAD`, no branch), in `aiport.yaml.worktrees/`,
  and removes when the run ends. Your working copy is never where the
  agent works, so even an edit the agent's own settings allow without
  asking (see "What aiport cannot see" below) cannot touch your files.
  Your uncommitted changes are not in that copy; the repository needs at
  least one commit.
- **What the agent is told**: the Issue's evidence (the same as above) and
  that the run is read-only.
- **Read-only**: aiport answers every permission request itself, without
  asking you. Reading and searching inside the run's copy are allowed;
  every edit, delete, command and anything outside that copy (also your
  own working copy) is denied. aiport only ever answers "once", so nothing
  is remembered in the agent's own settings. A permission request that
  names a session aiport did not open is denied.
- **Which paths count**: every path a request names counts, not only its
  locations: the paths in its raw input (`path`, `file_path`, `cwd`, a
  move's `destination`, the files of a patch, and the base of a search
  pattern such as `/etc/**` or `../*.py`), at any depth. One path outside
  (or in your home directory, `~`) denies the request, and a read or
  search that names no path at all is denied too. This holds for Explain
  and Fix it alike.
- **Session modes**: when the agent offers modes (Claude Code's adapter
  offers, for example, `default`, `acceptEdits`, `plan` and
  `bypassPermissions`), aiport switches the session to the most
  restrictive suitable one: plan or read-only for Explain, one that asks
  before acting (such as `default`) for Fix it. It never picks a mode that
  bypasses or auto-accepts. If the session starts in such a mode (from
  your own settings) and the agent offers no other, aiport does not run
  it. The run panel names the mode chosen.
- **The run panel** under the Issue shows the agent's name, where its
  model calls go, its steps (thoughts, files read, permissions allowed or
  denied) and its words as they come. **Stop** ends the run at any time
  (the agent is asked to cancel, then its whole process group is ended).
  Nothing is saved from a stopped run. Closing the tab or leaving the page
  does not stop it: the run goes on, saves its explanation when it ends,
  and the run panel picks it up again when you come back.
- **Limits**: one Agent run per Application at a time; an Explain run
  stops after 10 minutes (`agent_runs.explain_minutes`), and nothing is
  saved.
- **What is kept**: the explanation is saved with the Issue (replacing an
  earlier one, also one a Connection wrote) and carried to the
  Application's Coding agent as above. The run's log (its steps and words)
  is kept in memory only while the Issue is open or resolving, never on
  disk: it is gone when the Issue closes or aiport restarts.

**What aiport cannot see.** aiport only sees the permission requests an
agent sends. An agent's own settings can let it read, edit or run commands
without asking, and those actions never reach aiport: it cannot allow,
deny or log them. aiport limits what such settings can do, in layers:

- Every run works in a worktree made for it, never in your working copy
  (Explain's is thrown away when the run ends; Fix it's is shown to you as
  a diff), so an edit nobody was asked about lands only there.
- It picks the most restrictive session mode the agent offers (see
  Session modes above) and does not run a session that only acts without
  asking.
- Fix it never allows edits to agent or git configuration in the worktree,
  and any edit the agent reports ends "Allow until the next edit".
- Before a run, the run panel warns when the repository's own agent
  settings allow without asking: `.claude/settings.json` or
  `.claude/settings.local.json` with `permissions.allow` entries or a
  `defaultMode` such as `bypassPermissions` or `acceptEdits`,
  `.gemini/settings.json` with `autoAccept` or an approval mode such as
  `yolo`, and `.codex/config.toml` with `approval_policy = "never"` or
  `sandbox_mode = "danger-full-access"`. aiport only reads these files; it
  never changes them.

What it cannot check is your user-level agent settings (for example
`~/.claude/settings.json`, `~/.gemini/settings.json` or
`~/.codex/config.toml`) and any managed settings. A command that those
allow runs without aiport asking you, with your own rights, in the run's
worktree or, if the agent chooses, elsewhere. **Do not auto-allow commands
or edits** in your user-level settings for repositories you use with
aiport, and check the warnings in the run panel.

**Where the run's model calls go.** In **Settings → Default coding agent →
Model calls of Agent runs** you choose an Anthropic-protocol Connection
(and, optionally, a model) for them. Until you choose one, every run uses
the agent's own login and the run panel says so: aiport never sends a
run's calls to the Anthropic protocol default you did not choose for Agent
runs.

- **Claude Code** is pointed at aiport by its environment only:
  `ANTHROPIC_BASE_URL` (aiport's address on this machine: the listen
  port on `127.0.0.1`, or the listen address when it names a loopback
  address), `ANTHROPIC_AUTH_TOKEN`
  (a credential made for this run) and `ANTHROPIC_MODEL` when you chose a
  model. That credential authenticates only model calls, as the built-in
  Application `aiport`; it cannot reach the dashboard, the MCP endpoint or
  any Application's data, is never written down or logged, and stops
  working when the run ends. The calls show in Activity under `aiport`
  with their Connection. Without a Connection chosen for Agent runs (or
  when the chosen one is not reachable), Claude Code uses its own login
  and the run panel says so.
- **Codex** uses its own login: pointing it at aiport would mean editing
  its configuration file, which aiport never does. The run panel says so.
- **Gemini CLI** uses its own login too: aiport does not speak the Gemini
  API yet. The run panel says so.

```yaml
agent_runs:
  default_agent: claude
  connection: claude-hosted   # an Anthropic-protocol Connection
  model: claude-model-id      # optional
  explain_minutes: 10
```

For scripts: `PUT /api/v1/agents/calls` with `{"connection", "model"}`
(a revertable Change), `GET /api/v1/runs` (the runs going),
`POST /api/v1/runs/{id}/stop`, and `GET /api/v1/issues/{id}/run` (the
Issue's last run and its log, while the Issue is open). The explain
stream adds `run` (the run: its id, agent and model calls) and `step`
events for an Agent run; `error` codes include `run_stopped` and
`explanation_timed_out`.

#### Fix it

**Fix it** on an open Issue lets the Application's Coding agent change the
Application's code, on a branch of its own, and ends with a diff you
review. It shows when the Application has a Coding agent and a repository
path, as for Explain by a Coding agent; otherwise a one-line hint under
the Issue says what is missing (for example "Add the repository path of
docs-bot").

Fix it is offered only when the cause can be in the Application's code:
a model name the Application asks for that no Connection serves. For a
failing Connection the Issue says instead that the cause is not in the
Application's code and points to the aiport fix (binding the Application
to a working Connection).

- **Its own worktree and branch**: aiport makes a git worktree of the
  repository on a new branch named for the Issue, `aiport/fix-<issue id>`
  (`-2`, `-3`… if that branch exists), starting at your current commit
  (`HEAD`). The worktree lives next to the configuration
  (`aiport.yaml.worktrees/`), not inside your repository. Your working
  copy, its uncommitted changes and your branch are never touched; your
  uncommitted changes are not in the worktree either. The repository needs
  at least one commit.
- **What the agent may do**: aiport answers every permission request
  itself. Reading, searching and editing, creating, deleting or moving
  files inside the worktree are allowed (symbolic links are followed, so a
  link pointing out of the worktree is outside). Never allowed, at any
  level of the worktree and whatever their case: the agents' own
  configuration (`.claude/`, `.codex/`, `.gemini/`, `.mcp.json`,
  `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `GEMINI.md`) and git's
  (`.git/`, git hooks in `.husky/` or `.githooks/`, `.gitattributes`,
  `.gitmodules`), so a kept branch cannot change how your agents or git
  behave. A command inside the worktree waits for
  your answer (see Approving commands below). Anything outside the
  worktree, including your own copy of the repository and a command that
  says it runs elsewhere, is denied without asking. aiport only ever
  answers "once" to the agent, so nothing is remembered in the agent's own
  settings. The agent starts as for Explain: by absolute path, with the
  minimal environment, only on your click.
- **Approving commands**: when the agent asks to run a command (tests, a
  build), the Issue shows the whole command exactly as the agent sent it
  (as plain text: it is the agent's, not aiport's), never cut: a long one
  scrolls. Characters that would not show as themselves (control
  characters, invisible and direction-changing characters such as
  `‮`, unusual spaces) are written as such escapes, so what you read
  is what runs; a list of arguments is shown shell-quoted (`echo 'a b'`).
  A command longer than 16 KiB is denied without asking, never shown in
  part. There are three answers:
  - **Allow once**: this one time.
  - **Allow until the next edit**: this time and every time the run asks
    for the exact same command again (the same list of arguments, or the
    same string, in the same directory: `echo 'a b'` is not `echo a b`),
    until the agent changes a file, whether it asked to or not: an edit
    can change what the command does, so it asks again after one. Another
    command, or the same one in a later run, asks again. aiport only sees
    the edits the agent asks about or reports: a command you allowed that
    changes files itself (a code generator, a formatter) does not end the
    allowance, so choose Allow once for commands that write files.
  - **Deny**: the agent is told no and goes on without it.

  The agent runs an allowed command itself, in its own process, in the
  worktree's directory; aiport only answers the permission, and what the
  command changes in the worktree shows in the diff. While a command
  waits, the Changes badge in the Monitor menu counts it. A command left
  unanswered for 5 minutes stops the run: what it changed so far stays in
  its worktree, and the Issue shows it for review with Keep the branch and
  Discard, as after a finished run (state `stopped_waiting`).
- **What the agent is told**: the Issue's evidence, the saved explanation
  if there is one, the branch it works on, and that it must not commit,
  merge or push.
- **The run panel** shows its steps and the permissions allowed or denied,
  with **Stop**. The run goes on if you leave the page.
- **The diff**: when the agent ends its turn, the Issue shows **Review the
  change** with the branch, the summary ("2 files changed, 3
  insertions(+), 1 deletion(-)") and the whole diff. Nothing is committed
  yet.
  - **Keep the branch**: aiport commits the diff on the new branch, as one
    commit authored by `aiport <aiport@localhost>`, and removes the
    worktree. This is the only commit aiport ever makes, and only when
    you click Keep: on the new branch, never on yours, with hooks off. The
    branch stays in your repository for you to look at, merge or delete;
    aiport never commits to your branch, merges or pushes. The Issue becomes resolving as an Application fix and resolves
    when the Application's calls succeed again, that is after you merged
    and deployed the change. "Copy fix for the Coding agent" names the
    branch.
  - **Discard**: aiport removes the worktree and deletes the branch; the
    Issue stays open, and you can Fix it again.
- **Time limit**: a run stops after 30 minutes (`agent_runs.fix_minutes`).
  What it changed by then stays in its worktree and waits for your review
  with Keep the branch and Discard, as after a finished run (state
  `timed_out`).
- **Nothing left behind**: Stop, a run that fails and a run that changed
  nothing (also one that timed out or went unanswered before changing
  anything) remove the worktree and the branch; the Issue stays open and
  shows why the run ended.
- **Cleanup at startup**: a Fix it run that was going when aiport stopped
  (or crashed) is marked stopped the next time aiport starts, and its
  worktree and branch are removed; so are the throwaway worktrees of
  Explain runs that were going. Only directories inside
  `aiport.yaml.worktrees/` and branches named `aiport/fix-…` are ever
  removed.
- **What is kept**: the branch name, the commit it started from, the diff
  summary and the changed files are saved with the Issue (and, once kept,
  the commit aiport made), also across restarts. The diff itself is read
  from the worktree or the branch when you look at it. The run's log is
  kept as for Explain.
- **git, safely**: aiport finds git once, at startup, by absolute path
  (`/usr/bin/git` when it is there, otherwise the first `git` in its
  `PATH`; the startup output names it), and runs it with a minimal
  environment, without your global or the system's git configuration and
  attributes (`GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`,
  `GIT_ATTR_NOSYSTEM=1`, `core.attributesFile=/dev/null`), and with hooks,
  fsmonitor, commit signing and external diff programs turned off. What
  the repository itself configures in its own `.git/config` and
  `.gitattributes` still applies (for example a clean or smudge filter it
  defines): it is your own repository, so check those before using Fix it
  on a repository you did not set up.
- **Only its own worktrees**: aiport removes a worktree only when its
  directory is directly inside `aiport.yaml.worktrees/`; any other path it
  finds recorded is left alone and logged.

```yaml
agent_runs:
  fix_minutes: 30
```

For scripts: `POST /api/v1/issues/{id}/fix` starts it (`202` with the run;
`409` with `fix_unavailable` and the hint, `fix_not_for_cause` for a
failing Connection, `fix_waiting` while an earlier
diff waits for you, `agent_run_running`, `issue_not_open` or
`worktree_failed`); `GET /api/v1/issues/{id}/fix` gives `branch_fix` and
the `diff`; `POST /api/v1/issues/{id}/fix/keep` and `…/fix/discard`.
`GET /api/v1/issues` says per Application in `fix_by` whether Fix it is
available, with which agent and time limit and for which `causes` (with
`cause_hint` for the others), or the hint; each Issue carries its
`branch_fix`. Commands waiting for you: `GET /api/v1/approvals`
(each with `id`, `run`, `application`, `issue`, `command`, `expires_at_ms`);
answer with `POST /api/v1/approvals/{id}` and `{"answer": "allow_once"}`,
`"allow_run"` (allow until the next edit) or `"deny"` (`404 approval_not_found` once answered or its run
ended). `GET /api/v1/runs` gives each run's `waiting` count, and
`GET /api/v1/issues/{id}/run` the Issue's waiting `approvals`. These are
management routes, never open to an Application token; answering needs
the dashboard's CSRF header.

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
you can still approve. A fix you rejected is never applied automatically
either, also if you turn automatic fixes on later; you can still approve it
yourself. To keep these decisions, an Issue that holds a fix you reverted or
rejected is kept beyond request-history retention
(`telemetry.retention_days`), even once it is resolved or dismissed; it
holds no more than any other Issue (ids, model names, counts).

### What the Application's Coding agent sees

The Application's Coding agent hears about its Issues over the connection
it already has with aiport, with the Application's own token (your
management token gets `403`):

- The MCP tool `aiport_my_issues`, or `GET /v1/aiport/issues` (and
  `GET /v1/aiport/issues/{id}` for one), lists that Application's open and
  recent Issues: what is wrong, the evidence, the fix in the Application
  (the same text as "Copy fix for the Coding agent"), the root cause when
  you had the Issue explained (`explanation`) and whether the aiport
  fix is proposed, applied, rejected or reverted. An Application never sees
  another Application's Issues (their ids answer `404`), and never your
  notes.
- The capabilities answer (`aiport_capabilities`) shows `open_issues`, the
  number of its Issues that are open or resolving.
- A failing call that is one of an open Issue's failures carries an
  `X-Aiport-Hint` header naming that Issue, so the Application's own logs
  point the agent to it: for an unknown model, a provider `404` for the
  same model; for a failing Connection, a `5xx`, timeout or connection
  error from that Connection; for a Catalog change, a failing call that
  sent a removed task name. Other failing calls, and aiport's own
  refusals that are not the Issue's cause, name no Issue. Issue hints are
  sent whether or not the Application may request changes: telling it about
  its Issues is not a Change request.
- The `AGENTS.md` line from the Coding agent brief tells later sessions to
  check `aiport_catalog`, `aiport_my_issues` and `aiport_guidance` at the
  start.

When the agent has fixed the Application's code, it says so with `aiport_issue_fixed` (or
`POST /v1/aiport/issues/{id}/fixed`). The Issue becomes **resolving** by the
Application fix, so you see it is being handled; an aiport fix waiting for
you keeps waiting, and you can still approve it. aiport does not
take the agent's word for it: the Issue is resolved, as `application_fix`,
only once that Application's calls succeed again for the same cause (for an
unknown model, a successful call for that same model; for a failing
Connection, a successful call routed for that Application after it said
so). A successful call for another model does not count, and neither does
silence. If the error goes on, the
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
  An Application names a task (for example the alias `review-code`), not a
  provider: you decide which Connection and model serve it. It may add
  advice ("suggests big-model on hosted", with its reasons); the
  advice pre-fills the form, and you can choose otherwise now or change it
  later (`origins[].aliases` in the configuration file, or a new Change
  request) without touching the Application's code.
  Each shows what it asks for; **Details and reasons** opens the full
  details, why it is needed, the argument for the Scope asked for, and who
  filed it (as stated). Identical requests
  from several Applications are merged into one item, each with its own
  argument; one decision answers them all. Request text is plain text from
  the Application: it is shown as is and never sent to a model.

The **Waiting / Accepted / Rejected** filter on **Suggestions and Change
requests** picks what the list shows. Waiting is what needs you. Accepted and
Rejected show past decisions, most recent first, with who decided, when and
the note; Rejected includes dismissed Suggestions. Decided items are kept as
long as request history (`retention_days`); an accepted one's Change stays in
**Applied changes**, where you can revert it.

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
no hints are sent, Explain is unavailable, and `/_connect` and both briefs mark change requests,
Issues, their MCP tools and `X-Aiport-Hint` as off with the reason. The
Catalog is still disclosed (`/v1/models`, `X-Aiport-Catalog`,
`/v1/aiport/catalog`, `aiport_catalog`) and Catalog changes are still
recorded, but no Catalog-change Issues are opened. Forwarding is unchanged, and Revert
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
| Refused: the Level is not mapped for the … protocol | `400 level_unmapped` | Routing → Levels: map the Level on that protocol, or give the Application its own mapping on its Integrate row. When it keeps happening aiport opens an Issue (see An unmapped Level). |
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
| Provider error: unknown model | provider `404` (`not_found_error`, `model_not_found`) | Use a model the provider lists (the Application sees them with `GET /v1/models`), or add an Alias. When it keeps happening aiport opens an Issue in Changes that says which Connection serves the model. |
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
