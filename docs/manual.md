# aiport manual

aiport gives every application and coding agent on your machine one local
gateway for its LLM calls: one place for provider keys, one place to choose
which model serves what, and one place to see what was called and how well it
went.

**Who this is for:** the Owner, the person who runs aiport on their machine
and decides what it does. Coding agents that integrate an application read the
[communication standard](#the-communication-standard) and the briefs at
`/_connect`; for exact configuration keys see
[configuration](configuration.md), for the wire protocol see
[Connect](connect.md), and for policies and guardrails see
[Extensions](extensions.md). Capitalized words (Application, Connection,
Level, Catalog, Issue, Change and the rest) are defined in [Core
concepts](#core-concepts) and the [glossary](../GLOSSARY.md).

## At a glance

| Concept | In one line |
| --- | --- |
| Connection | A provider aiport forwards to (OpenAI, Anthropic, Bedrock, Meta, Ollama, llama.cpp), with its key kept by aiport. |
| Application | Anything that calls aiport with its own token; every call is attributed to it. |
| Protocol | `openai` or `anthropic`. aiport forwards each call natively and never translates between them. |
| Level | `fast`, `balanced` or `best`: the model name an Application sends to say how capable a model it needs. |
| Task name | A model name for one kind of work (`summarize`, `classify`) that you point at a model. |
| Effort | `low`, `medium` or `high`, sent with `X-Aiport-Effort`: how hard one call is. |
| Module | The part of the Application a call serves, sent with `X-Aiport-Module`; votes are read per module. |
| Catalog | What one Application can use through aiport, with a version on every answer. |
| Issue | A problem aiport keeps seeing for one Application, with evidence and a fix on each side. |
| Change | One configuration change made through aiport. Every Change can be reverted. |
| Suggestion | A change aiport proposes from request history or from votes. You decide. |
| Communication standard | The rules every integration follows, at `/_connect?format=standard`. |

## Quickstart

This takes about five minutes and ends with one real call through aiport,
visible in Activity. You need Go 1.27 or newer; dependencies are vendored and
the dashboard is pre-built, so nothing else is downloaded.

**Step 1: Build aiport.**

```sh
git clone https://github.com/harelix/aiport && cd aiport
make build                 # -> bin/aiport, dashboard included
```

Or install the binary without cloning: `go install
github.com/harelix/aiport/cmd/aiport@latest` (it lands in `$(go env
GOPATH)/bin`).

**Step 2: Create the configuration.**

```sh
./bin/aiport init          # choose providers, coding agents and Levels
./bin/aiport check         # validate the configuration
```

`init` writes `aiport.yaml` and a private credential file next to it,
`aiport.yaml.env` (mode 0600). It never overwrites a file and never asks for
provider keys in the terminal. It also proposes a first Level mapping; see
[Map Levels with init](#map-levels-with-init).

**Step 3: Start the gateway.**

```sh
./bin/aiport serve         # prints the dashboard address and the management token
```

`serve` prints the management token when it runs in a terminal. `./bin/aiport
auth token` prints it again later.

**Step 4: Sign in.** Open the address `serve` printed
(`http://127.0.0.1:8787/` by default). Paste the management token, then set a
password of at least 12 characters and an optional username.

**Step 5: Add a provider key.** On **Providers**, paste the provider's API key
into its card, then click **Check connection**. The check makes one free `GET
/models` call and generates nothing. Start a local provider (Ollama,
llama.cpp) yourself; it needs no key.

**Step 6: Map a Level.** On **Routing**, in the **Levels** panel, click the
pencil for `fast` on the OpenAI protocol and choose a Connection and a model.

**Step 7: Get an Application token.** On **Integrate**, click **Add with a
token**, name it `quickstart` and copy the two `.aiport` lines it shows. Save
them as `.aiport` in an empty directory and load them into your shell:

```sh
set -a; . ./.aiport; set +a
```

**Step 8: Make the first call.**

```sh
curl -s http://127.0.0.1:8787/v1/chat/completions \
  -H "Authorization: Bearer $AIPORT_APP_TOKEN" \
  -H "X-Aiport-Module: quickstart" \
  -H "X-Aiport-Effort: low" \
  -H 'content-type: application/json' \
  -d '{"model":"fast","messages":[{"role":"user","content":"Say hello"}]}'
```

Open **Activity**: the call is there, with the name you sent (`fast`), its
effort, and the Connection and model that served it.

**Next:** connect a real application by handing its coding agent the
communication standard ([Connect an application](#connect-an-application)),
and read [Core concepts](#core-concepts) for the words used in the rest of
this manual.

## Core concepts

Each concept below is one idea aiport is built on, with the reason it exists.

**Connection.** A provider aiport can forward to: OpenAI, Anthropic, Amazon
Bedrock, Meta Model API, or a local server such as Ollama or llama.cpp. It
holds an address and a key. It exists so provider keys live in one place,
never in your applications.

**Application.** Anything that calls aiport with its own token: a script, a
service, a desktop app or a coding agent. It gets its token by pairing, which
you approve. It exists so every call is attributed to someone, and so aiport
can tell one Application about its own problems.

**Protocol.** Each Connection speaks one protocol, `openai` or `anthropic`.
aiport forwards Chat Completions, Responses and Messages calls in the
provider's own format, streaming included, and never translates between
protocols. An OpenAI SDK call needs a Connection that speaks `openai`; an
Anthropic SDK call needs one that speaks `anthropic`. Native forwarding means
nothing is lost in translation.

**Level.** One of three model names every aiport shares: `fast`, `balanced`
and `best`. An Application sends a Level as its model to say how capable a
model it needs, without naming a provider. You map each Level per protocol to
a Connection and a model, and you can give one Application its own mapping. It
exists so you can change models without anyone changing code.

**Task name.** A model name for one kind of work, such as `summarize`,
`review-code` or `classify`, that you point at a Connection and model.
Technically it is an Alias: a name aiport replaces with a real model. A shared
task name marked `everywhere` can be sent by every Application, and one with a
fixed `reasoning` effort always sends that effort to the provider. It exists
so routine work can run on a cheaper model while the Levels stay on stronger
ones.

**Effort.** How hard one call is: `low`, `medium` or `high`, sent with the
`X-Aiport-Effort` header. A Level or task name may have a different model per
effort, and an effort stated only in the header also reaches the provider as
its reasoning effort. It exists so a call can ask for more or less thinking
without choosing another model name.

**Module.** The part of the Application a call serves, such as `tldr`,
`classify` or `insights`, sent with the `X-Aiport-Module` header. aiport
records it with the call and never forwards it. It exists so the votes people
give answers can be read per module, and so one module can be moved to another
model without moving the others.

**Catalog.** What one Application can use through aiport: its Levels, its task
names and the models of the Connections it reaches, with a short version. The
Application reads it at `GET /v1/models` and sees the version on every answer
in `X-Aiport-Catalog`. It exists so an Application never guesses a model name,
and learns at once when you change something.

**Change.** One configuration change made through aiport, by you, by an
approved request, or by an automatic fix you allowed. Every Change is listed
on the Changes page and can be reverted. It exists so you can experiment
safely.

**Issue.** A problem aiport keeps seeing for one Application, such as a model
no Connection serves. It comes with the evidence, a fix in aiport you approve,
and a fix in the Application for its coding agent, and it resolves itself once
calls succeed again. It exists so a failing integration is noticed and
explained instead of retried forever.

**Suggestion.** A change aiport proposes; you always decide. There are two
kinds. Suggestions on the Changes page come from request history (for example,
an unknown token that keeps trying to call). Suggestions from the votes come
from the thumbs people give answers, appear on the Feedback page, and are
applied as module routes.

**Module route.** A route for one Application's calls for one module that send
one name: they go to another Connection and model, ahead of every other route,
while its other modules keep theirs. It exists so a module that does badly can
move without moving the whole Application.

**Change request.** An Application asking for something aiport lacks, such as
a task name, with what it needs and why. Only Applications you allow may send
one, and you decide it on the Changes page.

**Guidance version.** The number of the guidance aiport gives coding agents
(the briefs, the `AGENTS.md` line and the MCP instructions), now version 4.
aiport remembers the version each Application's coding agent last confirmed
and tells it when it is behind. It exists so integrations keep up as aiport
changes.

**Communication standard.** One document, at `/_connect?format=standard`, with
every rule an integration follows and a checklist: the token in `.aiport`, the
SDK pointed at aiport, Levels and task names, `X-Aiport-Effort`,
`X-Aiport-Module`, the thumbs up and down, and the error fixes. It exists so
one paste brings any application in line, new or already connected.

## Guides

Each guide below is one task, start to finish. The page-by-page reference for
the dashboard is in [A tour of the dashboard](#a-tour-of-the-dashboard).

### Connect an application

You connect an application by handing its coding agent one instruction; the
agent wires the LLM calls through aiport and pairs the application, and you
approve the code it shows. The application then keeps its own token, sends its
calls to aiport, and never holds a provider key.

#### The communication standard

The communication standard is one document, served at
`/_connect?format=standard`, with every rule an integration follows and a
checklist at the end. It is written for the coding agent working in the
application's repository, and it is the fastest way to bring any application
in line, new or already connected.

The standard has seven parts, each required:

1. **Identity.** One token per application, in `.aiport` at the repository
   root, listed in `.gitignore`.
2. **Transport.** The OpenAI SDK's base URL is `AIPORT_BASE_URL + "/v1"`, the
   Anthropic SDK's is `AIPORT_BASE_URL`, both with `api_key =
   AIPORT_APP_TOKEN`. No provider keys stay in the application.
3. **Model names.** Send a Level or a task name, never a provider's model id.
   `GET /v1/models` lists the names.
4. **Headers.** `X-Aiport-Module` on every model call (required),
   `X-Aiport-Effort` where calls differ in difficulty. Keep
   `X-Aiport-Request-Id` with every answer, and read `/v1/models` again when
   `X-Aiport-Catalog` changes.
5. **Rating.** A thumbs up and a thumbs down next to every LLM answer the
   application shows a person, built into its own UI by the coding agent, sent
   to `POST /v1/aiport/feedback`.
6. **Errors.** Follow `error.fix` on `401` and `403`; retry with backoff only
   `429 concurrency_limit`, `502` and `504`.
7. **Talking to aiport.** The MCP tools at the start of each session, Change
   requests instead of workarounds, and aiport's line in `AGENTS.md`.

On **Integrate**, the **Communication standard** card at the top of
**Instructions** shows the prompt that hands the standard to a coding agent.
**Copy instruction** copies it, **Copy URL only** copies the address, and
**Read the standard** opens it. The same document is the MCP tool
`aiport_standard`, and the `/_connect` JSON gives it as `standard` with the
prompt as `standard_prompt`.

**Note:** an application without a user interface (a service, batch job or
command-line tool) offers the vote where a person does see the answer. If no
person ever sees it, the coding agent tells you so and skips votes, but still
sends `X-Aiport-Module`.

#### Onboarding an application, step by step

The **Onboard an application** wizard on Integrate walks you through
connecting one application, and each step ticks itself from live state. It is
open while you have no Applications and whenever a pairing request waits;
otherwise it is collapsed, and its header says where things stand ("1 pairing
waiting", or the followed Application and its progress).

The step to do now is highlighted. Select an existing Application in the list
to see how far it got, or click **Onboard a new application** to follow a new
one.

1. **Hand this to the coding agent.** Open the coding agent (Claude Code,
   Codex, Cursor and so on) in the application's repository and paste the
   instruction shown (**Copy instruction**; **Copy URL only** copies just the
   address). It tells the agent to read
   `http://127.0.0.1:8787/_connect?format=markdown` and do both parts in
   order: first wire the LLM calls through aiport (pair, save the token in
   `.aiport` and add it to `.gitignore`, read `.aiport` at startup, point the
   SDK at aiport, remove provider keys, send `X-Aiport-Module`, add the thumbs
   up and down, make one test call), then add the aiport MCP entry and the
   `AGENTS.md` line. It points at the communication standard for the
   checklist, asks the agent to tell you the pairing code and wait, never to
   commit `.aiport`, and to ask you for a **New token** instead of pairing
   again when the token is lost. The exact text, with your gateway's address,
   is also at the top of the markdown document and in the `/_connect` JSON as
   `agent_prompt`.
2. **Approve the pairing.** The agent asks aiport to pair, and the request
   appears at the top of Applications within seconds with a short code. Check
   it matches the code the agent shows you, then **Approve** (optionally
   binding a Connection and allowing change requests). Requests wait 30
   minutes. Once approved, the wizard follows that Application.
3. **First call arrives.** The agent makes one real call; the step ticks when
   Activity has a call from the Application in the last 7 days.
4. **Allow change requests** (optional), with the toggle in place. You still
   decide each request in Changes.
5. **MCP entry in the repository** (optional). The agent adds it in step 1;
   tick it by hand, or copy it from the **Coding agent brief** tile.
6. **Connect a coding agent** (optional). See [Connecting a coding
   agent](#connecting-a-coding-agent). The step ticks when the Application has
   a coding agent (its own or the default) and a repository path.

`/_connect?format=markdown` opens with a **Start here** section and holds
everything for both parts, so an agent handed only that address needs nothing
else.

#### Applications and pairing

An Application is anything that calls aiport with its own token, and pairing
is how it gets that token with your approval. Each call is attributed to the
Application that sent it, so Activity shows who called what; calls without an
Application token show as `unknown`.

Pairing goes like this:

1. The application asks aiport with `POST /v1/aiport/pair` and its name (no
   token needed; public on loopback and rate-limited) and shows you a short
   code. It may also report the path of its git repository on this machine;
   the Application brief asks the integrating coding agent to.
2. The request appears at the top of the Applications list on **Integrate**
   (and in the Applications table on **Routing**) with that code. Check the
   application shows the same code. Optionally choose a Connection to bind it
   to and tick **May request changes**, then **Approve**, or **Reject**.
3. The application collects its token once. aiport adds it to the
   configuration without a restart and stores the token in the private
   credential file as `AIPORT_APP_<ID>_TOKEN`.

A reported repository path is shown in a field under the request, with what is
wrong with it if it is not a git repository here. Confirm it, edit it or clear
it before you approve. aiport keeps only what the field holds when you
approve; a reported path you did not confirm is never used. A path that does
not exist, or is not the top directory of a git repository, is refused with
the reason, and the request keeps waiting.

**Where the token lives: `.aiport`.** Each application keeps its aiport
settings in a file named `.aiport` at the root of its repository, listed in
its `.gitignore` and never committed:

```
AIPORT_BASE_URL=http://127.0.0.1:8787
AIPORT_APP_TOKEN=<its token>
```

The application fills its environment from this file at startup (variables
already set win). Its coding agents read it too: the project MCP entry's
`headersHelper` takes the token from `AIPORT_APP_TOKEN` or else from
`.aiport`, so nothing has to be exported before the agent starts. When a token
is lost or revoked (aiport answers `401`), click **New token** on the
Application's row on Integrate, then **Copy .aiport lines** and paste them
over its `.aiport`; the old token stops working at once. Guidance version 3
told the coding agents of Applications integrated earlier to move their token
there.

**One request per application.** An application that asks again while its
request waits (the same repository, or the same name ignoring case and
punctuation) gets that same request back, with the same code, so the code you
see stays valid; its earlier and newer poll paths both work. A request that
matches an Application you already have is marked **Already connected as …**.
Its main button, **Give … a new token**, approves it as that Application (it
receives a new token and its old one stops working) instead of adding a
duplicate such as `docs-bot-2`. **Add as separate** still adds a new
Application. Over the API, approve with `{"replace": "<id>"}`.

**Waiting and restarts.** Requests wait 30 minutes for you and survive an
aiport restart (the poll secret is stored only as a hash). A waiting request
can still be approved by its code, and an approved application can still
collect its token for 24 hours after a restart. A poll on a request that ran
out answers `expired`, so the application knows to ask again. A poll answers
`pending`, `rejected`, `expired`, or `approved` with the token.

**Rejected requests.** A request you **Reject** is remembered: the application
learns it was rejected, and aiport keeps its name, description, code and the
time (never a token), the newest 100, across restarts. See them on Integrate
with the **Rejected** filter, where **Forget** removes one. A rejected
application may ask to pair again.

The **Applications** table on Routing lists every Application with its client
id, Connection, own Aliases, and the **May request changes** toggle.

#### Give a token yourself, or remove an Application

You can skip pairing and hand an application its token yourself, for example
when its coding agent is not running, or when you would rather paste the token
into its `.aiport` file.

- **Add with a token** (Integrate, next to **Onboard a new application**):
  name the Application and optionally choose a Connection. aiport adds it and
  shows its token once, on a token card with **Copy** and **Copy .aiport
  lines**: `AIPORT_BASE_URL` (this gateway's numeric address) and
  `AIPORT_APP_TOKEN`.
- **New token** (the key button at the end of an Application's row, or **Issue
  new token** after you click its name): after you confirm, aiport issues a
  new token and shows it once the same way. The old token stops working at
  once, whether it came from pairing or from here. Issuing a token changes no
  configuration, so it is not listed in Changes.
- **Remove** (the trash button at the end of the row, or **Remove
  application** after you click its name): names what goes with the
  Application (its own Aliases, its binding, its coding agent settings, its
  permission to request changes), then removes it as one Change and deletes
  its token, which stops working at once. Its activity history is kept.

Reverting a removal in **Changes** brings the Application back **without a
token**: its row shows **No token** and its calls are refused until you click
**Issue token**. `serve` and `aiport check` name such Applications too.

A token set in aiport's own environment when it was started (exported in the
shell rather than kept in the private credential file) cannot be replaced or
deleted from the dashboard: change it where aiport is started, and restart it.

With the management token, the same is `POST /api/v1/origins` with `{name,
target?}`, `POST /api/v1/origins/{id}/token` and `DELETE /api/v1/origins/{id}`
(send `X-Aiport-CSRF: 1` with each). A token is in the response once and is
never shown again.

### Set up the application's coding agent

The coding agent that works in an application's repository talks to aiport
over MCP, with the Application's own token, so it can read the Catalog, its
Issues and new guidance at the start of every session. You set this up once
per repository with an MCP entry and one line in `AGENTS.md`; the integrating
agent adds both in step 1 of onboarding.

#### The MCP entry and the AGENTS.md line

The **Coding agent brief** (`/_connect?format=handover&for=agent`) gives both,
ready to copy.

- **The project MCP entry** (`.mcp.json`) points at this gateway's `/mcp`
  endpoint over HTTP. Its `headersHelper` reads the token from
  `AIPORT_APP_TOKEN` or, when that is unset, from `.aiport` at the repository
  root, so the entry itself never holds the token.
- **The `AGENTS.md` line** tells every later session where the token is (and
  what to do on `401`), to send a Level or task name and never a provider's
  model id, to state effort with `X-Aiport-Effort`, to name the module of
  every call with `X-Aiport-Module`, to add the thumbs up and down where the
  app has a UI, to check `aiport_catalog`, `aiport_my_issues` and
  `aiport_guidance` at the start of a session (calling `aiport_guidance_done`
  after updating), and to ask with `aiport_request_change` instead of working
  around aiport.

The MCP `initialize` instructions an agent gets from aiport say the same as
the `AGENTS.md` line. The MCP endpoint serves one Application: a connection
without a recognized Application token is refused with `403
application_token_required`.

**The MCP tools.** Each tool answers exactly as its HTTP endpoint does, with
the Application's own token.

| Tool | HTTP equivalent | What it does |
| --- | --- | --- |
| `aiport_standard` | `GET /_connect?format=standard` | The communication standard, with its checklist. |
| `aiport_capabilities` | `GET /v1/aiport/capabilities` | Protocols, Connections, own Aliases, binding, `may_request_changes`, `change_kinds`, `open_issues`, `catalog` and `guidance`. |
| `aiport_catalog` | `GET /v1/aiport/catalog` | The Catalog and its recent changes, newest first. |
| `aiport_guidance` | `GET /v1/aiport/guidance` | What changed in the guidance since the confirmed version, the current `AGENTS.md` line and the steps. |
| `aiport_guidance_done` | `POST /v1/aiport/guidance/done` | Confirms the current guidance version. |
| `aiport_request_change` | `POST /v1/aiport/requests` | Files a Change request. |
| `aiport_my_requests` | `GET /v1/aiport/requests` (or `/{id}`) | The Application's Change requests and your answers. |
| `aiport_my_issues` | `GET /v1/aiport/issues` (or `/{id}`) | The Application's Issues. Offered only while adapting is on. |
| `aiport_issue_fixed` | `POST /v1/aiport/issues/{id}/fixed` | Says an Issue is fixed in the Application's code. Offered only while adapting is on. |

**Tip:** Claude Code reads the MCP token when it connects. After the token in
`.aiport` changes, run `/mcp` and reconnect aiport, or start a new session.

#### Guidance versions

aiport numbers the guidance it gives coding agents (the briefs, the
`AGENTS.md` line and the MCP instructions) and raises the number when that
guidance changes in a way agents must act on. Each version has a short change
log, and aiport remembers per Application the version its coding agent last
confirmed, so an agent that is behind learns exactly what to do.

| Version | What it asks agents to do |
| --- | --- |
| 1 | The project setup: the MCP entry, the `AGENTS.md` line, and reading `aiport_catalog` and `aiport_my_issues` at the start of a session. |
| 2 | Levels and effort: send a Level or task name instead of a provider's model id, add `X-Aiport-Effort` where calls differ in difficulty, replace the `AGENTS.md` line, and check `aiport_guidance` at the start of each session. |
| 3 | The token lives in `.aiport`: add it to `.gitignore`, move the token there, load it at startup, replace the MCP entry with the one whose `headersHelper` reads `.aiport`, replace the `AGENTS.md` line, and ask for a **New token** on `401` instead of pairing again. |
| 4 | Votes per module: read the communication standard, give each LLM feature a module name sent as `X-Aiport-Module`, keep `X-Aiport-Request-Id` with each answer, build a thumbs up and down into the UI, and replace the `AGENTS.md` line. |

Each version ends with calling `aiport_guidance_done`. The confirmed version
is kept in the database next to the configuration; Applications integrated
before guidance versions existed have none.

While an Application is behind, its coding agent is told without you:

- the MCP `initialize` instructions start with a one-line reminder;
- every aiport MCP tool answer ends with it, as a second text block (except
  the answers of `aiport_guidance` and `aiport_guidance_done`, which are the
  guidance);
- `aiport_capabilities` shows `guidance: {current, confirmed}` (`confirmed` is
  `null` when it never confirmed).

The agent calls `aiport_guidance` (`GET /v1/aiport/guidance`) for `{current,
confirmed, confirmed_at_ms, behind, changes, agents_line, steps}`, does the
steps, then calls `aiport_guidance_done` (`POST /v1/aiport/guidance/done`).
Both take the Application's own token and only ever show or change that
Application's state. All of this works with adapting off.

You see each Application's state in the **Guidance** column on Integrate, and
**Overview** shows one quiet line while any Application has not confirmed the
current version. An Application that never confirmed any version counts as not
confirmed; the line tells the two apart ("behind": confirmed an older version;
"never confirmed": none yet), as do `behind` and `never_confirmed` in `GET
/api/v1/guidance`. **Copy refresh instruction** in the Guidance column copies
one paragraph to paste to a coding agent that is behind.

### Levels

Every aiport shares three model names, the Levels `fast`, `balanced` and
`best`, and you decide what each means here, per protocol. An Application
sends a Level as its model to say what proficiency it needs without naming a
provider, so you can move it to another model without a code change.

A Level is a reserved Alias name:

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
Application bound to another one or with no protocol default. Only the `model`
field is replaced (and, in one case, the reasoning effort is added; see
[Effort and reasoning](#effort-and-reasoning)). Mapped Levels are listed in
each Application's `GET /v1/models` (on the protocol they are mapped on) and
in its Catalog. Changing a mapping is a Change you can revert and a Catalog
change the Applications are told about. Removing a Connection a Level is
mapped on asks for a replacement Connection of the same protocol.

A Level not mapped on the request's protocol is refused with `400
level_unmapped`, naming the Level and protocol: aiport never quietly uses
another model. When an Application keeps asking for a Level that is not
mapped, aiport opens an Issue for it (see [An unmapped
Level](#an-unmapped-level)).

A request that nothing could route is refused at once, before aiport reads its
body: no binding or protocol default for its protocol, no task name of its own
there, no Level mapped for it, no `everywhere` task name and no module route.
Once any of those exists for an Application, or while its Connection is only
switched off, aiport reads the model first, so a Level is still refused as
`level_unmapped`, naming it.

#### Map Levels with init

`aiport init` proposes a first mapping from the providers you chose, per
protocol: a local Connection for `fast` when you chose one, hosted ones for
`balanced` and `best` (whatever you chose when there is no other kind). The
presets know no model ids, so `init` asks you for each Level's model. An empty
answer leaves that Level unmapped, and nothing is mapped unless you confirm.

Without a terminal, map them with `--levels
fast=local-connection/llama3:8b,best=hosted-connection/some-model`
(`LEVEL=CONNECTION/MODEL`, the Connection being one of the chosen provider
ids; the protocol is that Connection's), or skip them with `--no-levels`. With
neither flag and no input, no Level is mapped and `init` says so.

#### Map Levels in the dashboard

You edit Levels in the dashboard without touching the file. Every save is one
Change: revert it in Changes to get the previous mapping back.

- **Routing → Levels** shows each Level on each protocol. The pencil (or
  **Map**) opens its mapping: a Connection of that protocol, a model (chosen
  from the models the Connection lists), and optional per-effort models.
  **Clear** removes the shared mapping. A cell also says how many Applications
  have their own.
- **Integrate → Applications → Levels** (the toggle under an Application's
  name, next to **Catalog**) shows the same grid for that Application: its own
  mappings, marked **Own**, and the shared ones it falls back to. The pencil
  gives it its own mapping; **Use the shared mapping** removes it again.

The [Routing map](#the-routing-map) shows what each Level reaches for each
Application, and Activity shows which name and effort each call asked for and
what served it.

#### Upgrading: Aliases named like a Level

If your configuration already had an Alias called `fast`, `balanced` or
`best`, it is now that Level's mapping on its Connection's protocol. A shared
one now chooses the Connection for every Application's calls for that name on
that protocol, whatever they are bound to (before, it only renamed a model on
the Connection a call already went to). An Application's own one applies on
its Connection's protocol only.

`aiport check` prints one notice per such entry, naming it and its Connection,
and the startup log sums them up in one line. To keep it a plain Alias, rename
it (its `alias:`) to a name that is not a Level and send the new name. To keep
it as the Level, nothing needs doing.

### Create task names for routine work

A task name is a model name for one kind of work, such as `classify`, `tldr`
or `extract`, that you point at a Connection and model. A typical use is a
cheaper model for routine work at a fixed effort, while the Levels stay on the
models for harder work.

#### Aliases and task names

An Alias is a model name an Application sends that aiport replaces with a real
model on a Connection; a task name is an Alias named for the work. Only the
`model` field of the request is rewritten.

- **Shared Aliases** apply to every Application (`aliases:` in `aiport.yaml`,
  one per Connection and name). Without `everywhere`, a shared Alias applies
  only to calls that already go to its Connection.
- **An Application's own Aliases** come first and may select another
  Connection than its usual one (`origins[].aliases`).

You can write Aliases in the configuration file, or let them arrive as
Changes: approving a Change request of kind `application_alias` (or an Issue's
fix) adds one, for that Application or, if you grant **every Application** as
the Scope, as a shared Alias. An Application lists its own Aliases with `GET
/v1/aiport/capabilities`, and every name it can use with `GET /v1/models` (see
[The Catalog](#the-catalog)).

For example, docs-bot has its own Alias `summarize` and the shared Alias
`review-code`, both on the local Connection it is bound to, which also serves
`llama3:8b`: docs-bot can send `summarize`, `review-code` or `llama3:8b`.

#### Task names for every Application, with a fixed reasoning effort

Set `everywhere: true` on a shared task name and every Application can send
it, whichever Connection its calls otherwise go to: the name routes to its own
Connection, like a Level. Set `reasoning: low`, `medium` or `high` and aiport
sends that reasoning effort to the provider on every call for the name whose
body sets none of its own, whatever `X-Aiport-Effort` says.

```yaml
aliases:
  - {target: local, alias: classify, model: 'llama3:8b', everywhere: true, reasoning: low}
```

Add, change and remove these with the management API: `PUT
/api/v1/task-names/{name}` with `{connection, model, effort?, everywhere?,
reasoning?}`, and `DELETE /api/v1/task-names/{name}` with `{connection}`. Each
is a revertable Change, and each Application's Catalog lists the name with its
`reasoning`.

A task name is 1 to 256 characters without spaces or slashes, and it cannot be
a Level (map those under Levels). One name can route `everywhere` from only
one Connection: remove it there first to move it.

#### Per-effort models of a task name

A task name's per-effort models are edited on its row of the [Routing
map](#the-routing-map) (and in an Application's Catalog on Integrate). The
pencil opens a model for low, medium and high on the task name's own
Connection; an empty field uses its model, and **Clear** removes them all. A
shared task name's per-effort models apply to every Application. Each save is
one Change you can revert.

Over the API it is `PUT /api/v1/aliases/effort` with `{application?, alias,
connection?, effort}`: a shared task name is found by its name and
`connection`, an Application's own by its name. In `aiport.yaml` it is the
`effort:` of the Alias.

### The Catalog

An Application's Catalog is what it can use through aiport: the Levels mapped
for it, its task names (its own Aliases, then the shared Aliases of the
Connections it reaches, and the `everywhere` names) and the models those
Connections list, with a short version. You decide what is in it, and aiport
tells the Application, so it never has to guess a model name or learn from a
provider's `404` that something changed.

The briefs and the communication standard tell every Application and its
coding agent how to use it: send a Level or task name, never a provider's
model id; read `GET /v1/models` for the names; watch the `X-Aiport-Catalog`
header; and check `aiport_catalog`, `aiport_my_issues` and `aiport_guidance`
at the start of each session.

All of the Catalog works with adapting turned off: it is information, not a
Change. Only the Catalog-change Issues follow the adapting switch.

#### What an Application reads

All of these answer only about the Application whose token is sent. Each
Application sees only its own Catalog, and keys, addresses, prompts and tokens
never appear.

- **`GET /v1/models`**, where OpenAI and Anthropic SDKs and coding agents
  already look for models, answers in the OpenAI list shape, or in the
  Anthropic shape when the request carries `anthropic-version` (Anthropic SDKs
  may send the token in `X-Api-Key`). Your management token gets `403
  application_token_required`: a Catalog belongs to one Application. Only
  names usable on that protocol are listed. Mapped Levels and task names come
  first, marked as aiport's (`owned_by: "aiport"`, or "(aiport Level)" and
  "(aiport task name)" in the display name); then the models of the Connection
  the Application is bound to, or of the protocol defaults. A Connection whose
  model list cannot be read adds nothing (lists are the same free, cached ones
  **Show models** uses).
- **`X-Aiport-Catalog: <version>`** (12 hex characters) is on every answer to
  a request with an Application token, forwarded calls and errors included;
  the provider's answer is otherwise unchanged. The version changes when the
  Application's task names or where they go, its binding, or the protocol
  defaults it follows change, and only then: a provider adding or retiring a
  model does not change it. An Application that sees a new version reads
  `/v1/models` again.
- **`GET /v1/aiport/catalog`** (or the MCP tool `aiport_catalog`, same answer)
  returns `catalog` and `changes`. `catalog` holds the task names with their
  Connection and model (Levels among them with `level: true` and their
  `protocol`, per-effort models as `effort`, a fixed `reasoning` where set),
  the binding or the protocol defaults followed, the Connections reached,
  their models and the version. `changes` holds its recent Catalog changes,
  newest first. Your management token gets `403 application_token_required`.
- **The capabilities answer** (`GET /v1/aiport/capabilities`) carries
  `catalog: {version, changed_at_ms}`: the version and the time of the newest
  kept Catalog change (`null` when none is kept).

The same view for one Application, with its recent Catalog changes, is under
**Catalog** on its row on Integrate.

#### The change log

Every configuration Change goes through one place, and on each one aiport
works out what it changed in every Application's Catalog: approving or
reverting an Alias, mapping a Level or changing per-effort models, adding a
task name, binding an Application, changing a protocol default, removing a
Connection. A revert is a Catalog change too, with `revert_of` naming the
Change it undid.

Each Catalog change records the Change id (`change_id`) and time (`at_ms`),
the version `before` and `after`, task names `removed`, `added` or `moved`
(`from` and `to`), the binding or protocol defaults that changed (`routing`),
and the Connections the Application no longer or newly reaches
(`connections_removed`, `connections_added`). An Application the Change did
not affect gets nothing. aiport's own model calls (the built-in Application
`aiport`, which follows `agent_runs.connection`, else the protocol defaults)
get Catalog changes recorded too.

Catalog changes are kept as long as request history
(`telemetry.retention_days`) and survive a restart. They hold Application ids,
task names, model names and Connection ids only, never keys, addresses,
prompts, responses or tokens.

#### When a Change takes something away

Removing a task name an Application used, or moving its binding or a protocol
default it followed, does more than change the version: aiport opens a
Catalog-change Issue (cause `catalog_changed`) for that Application, saying
what changed and what it can still use. "Used" means its calls sent that task
name, or a model name that went through that binding or default, within
request history; an Application that never did gets only the Catalog change.

The Issue needs no decision from you and resolves itself; see [A Catalog
change](#a-catalog-change). Before a removal, the removal dialog shows which
Applications it will affect (see [Removing a Connection, and when one
fails](#removing-a-connection-and-when-one-fails)).

### Effort and reasoning

Effort says how hard one call is, `low`, `medium` or `high`, and lets aiport
pick a stronger or cheaper model for the same name. For example, `balanced`
with high effort can go to a stronger model than `balanced` with low effort.

**Where a call's effort comes from**, in this order:

1. The `X-Aiport-Effort` header (`low`, `medium` or `high`), the one the
   briefs teach. It also accepts the provider words below (`none` and
   `minimal` count as low, `xhigh` and `max` as high). A value that is not an
   effort counts as none.
2. The provider's own effort field: `reasoning_effort` on OpenAI Chat
   Completions, `reasoning.effort` on OpenAI Responses, `output_config.effort`
   on Anthropic Messages. `none`, `minimal` and `low` count as low, `medium`
   as medium, `high`, `xhigh` and `max` as high. Numeric thinking budgets
   (`budget_tokens`) are never read or translated.
3. Otherwise `medium`.

**Per-effort models.** Any Alias, a Level or a task name, shared or an
Application's own, may carry optional per-effort models on its own Connection.
An effort without its own model uses the Alias's `model`.

```yaml
aliases:
  - {target: hosted, alias: balanced, model: mid-model, effort: {low: small-model, high: big-model}}
origins:
  - id: docs-bot
    # ...
    aliases:
      - {alias: summarize, target: local, model: 'llama3:8b', effort: {high: 'llama3:70b'}}
```

**What reaches the provider.** The `X-Aiport-Effort` header itself is never
forwarded; no `X-Aiport-*` request header is. The provider's own effort field,
when the body has one, is never changed. When the body has no effort field and
aiport chose the model (a Level, a task name, an Alias or a module route
replaced the name sent), aiport adds one to the body, so the provider reasons
that hard too:

- a task name with a fixed `reasoning` sends that effort, whatever the header
  says;
- otherwise, an effort that came from the `X-Aiport-Effort` header is sent as
  `low`, `medium` or `high`;
- otherwise nothing is added.

The field is `reasoning_effort` on Chat Completions, `reasoning.effort` on
Responses and `output_config.effort` on Messages. The rest of the body is kept
byte for byte. A model name sent unchanged (no Alias applies) gets nothing
added.

**Editing per-effort models.** A Level's are edited in its mapping on Routing
or Integrate (empty fields use the mapping's model); a task name's on its row
of the Routing map. Over the API, a `PUT` on a Level without `effort` keeps
them while the mapping stays on the same Connection, and drops them when it
moves to another (they name models of the old one). The same holds when a
removal plan moves a Level or a task name to another Connection: it keeps its
model and loses its per-effort models, which you set again on the new
Connection.

Per-effort models are listed in the Catalog (`effort` on a task name), and
changing them is a Catalog change. Activity shows the effort of each call and
where it came from (see [Activity and history](#activity-and-history)).

### Collect feedback and improve routing

People rate the answers an application shows them, and aiport turns those
votes into a per-module report and into Suggestions you can apply. This is how
you learn which model and effort suit each part of each application, and move
the part that does badly without moving the rest.

#### Votes from the application

Every integration names the module of each model call with the
`X-Aiport-Module` header and gives every LLM answer shown to a person a thumbs
up and a thumbs down. The communication standard and guidance version 4 ask
every connected application's coding agent to add both.

A module name is lower case, 1 to 64 characters of `a`–`z`, `0`–`9`, `.`, `_`
and `-` (not starting with a punctuation mark). It is recorded with the call
and never forwarded.

On a vote the application sends:

```sh
curl -s -X POST http://127.0.0.1:8787/v1/aiport/feedback \
  -H "Authorization: Bearer $AIPORT_APP_TOKEN" -H 'content-type: application/json' \
  -d '{"request_id":"<X-Aiport-Request-Id of the answer>","vote":"up"}'
```

- `vote` is `"up"` or `"down"`. Instead, a `score` from 0 to 1 also works; 0.5
  or more counts as up.
- `module` names the module only when the call did not send `X-Aiport-Module`;
  it never replaces one the call sent.
- A changed vote is sent again; the last one counts.
- An application can rate only its own calls, and only calls still retained in
  request history (otherwise `404 request_not_found`). A malformed vote is
  `400 invalid_feedback`.
- Prompts, answers and anything a person typed are never sent: only the
  request id, the vote and the module.

Votes need request history (`telemetry.enabled`) and count only calls still
retained. On Integrate, the **Guidance** column shows two chips per
Application with calls in the last 7 days: **Modules** (its calls name their
module) and **Votes** (its users' votes arrive).

#### Read the report

The [Feedback](#feedback) page shows each Application's votes per module, and
inside each module per name sent, model and effort. With two or more judged
modules it names the best and the weakest. Change the weakest module's Level,
task name or effort under Routing, or apply a Suggestion, and watch its share
of up-votes.

#### Suggestions from the votes

Once a module's calls for one name, on one model and effort, have at least 10
votes, aiport reads them against your Levels and lists what it finds under
**Suggestions from the votes** on the Feedback page. There are three kinds:

| Kind | When | What aiport suggests |
| --- | --- | --- |
| **Better elsewhere** | Below 60% up, and a module of the same name (in another application, or this one) on another model gets at least 20 points more on the same protocol, also over at least 10 votes | That model, on its Connection. The best such evidence wins; ties go to more votes. |
| **One Level up** | Below 60% up, the name sent is `fast` or `balanced`, and there is no such evidence | The next Level's model for that application (`fast` to `balanced`, `balanced` to `best`), untested for this module. |
| **Cheaper may do** | 90% up or more over at least 20 votes, and the name sent is `best` or `balanced` | The Level below's model for that application, untested for this module. |

Evidence comes only from the same module name: aiport compares like with like.
A Level whose model is the one already serving the module is not suggested.
Calls that named no module are never the subject of a Suggestion. There is at
most one Suggestion per application, module and name.

#### Apply a Suggestion as a module route

aiport never moves a module by itself. **Apply** adds a module route: that
application's calls for that module that send that name go to the suggested
model, ahead of every other route, while its other modules keep theirs. It is
one Change, with the reason from the votes included, that you can revert on
Changes.

Each application's panel on the Feedback page lists its module routes with
**Remove**. After applying, watch the module's share of up-votes: the new
model shows as its own row.

- A module route has one model for every effort: it carries no per-effort
  models. A task name's fixed `reasoning` still applies.
- Its Connection must speak the request's protocol; otherwise the call is
  refused with `400 incompatible_protocol`.
- In `aiport.yaml` module routes are `origins[].module_routes`, each `{module,
  name, target, model}`.
- Over the API: `PUT /api/v1/module-routes/{application}/{module}/{name}` with
  `{connection, model, why?}`, and `DELETE` on the same path. `why` (one line,
  at most 300 characters) goes into the Change.

### Issues

An Issue is a problem aiport keeps seeing for one Application, shown at the
top of Changes with its evidence, what aiport found out and the ways to
resolve it. There is one Issue per Application and cause: repeats update its
counts and time window, never a second Issue.

aiport opens four kinds of Issue, each from its own threshold under
`adapting:` in `aiport.yaml`:

| Cause | Opens when | Fix in aiport | Fix in the Application |
| --- | --- | --- | --- |
| Unknown model (`unknown_model`) | A provider keeps answering `404` for a model the Application has no Alias for (`adapting.unknown_model`, default 5 times in 60 minutes) | An Alias for that Application to the Connection that serves the model | Send a model this gateway serves, or ask for the Alias |
| Failing Connection (`failing_connection`) | Its calls to one Connection keep failing (`adapting.failing_connection`, default 5 in 60 minutes) | Bind that Application to a working Connection | Back off and handle `5xx` and timeouts |
| Unmapped Level (`level_unmapped`) | Its calls for one Level keep being refused (`adapting.level_unmapped`, default 5 in 60 minutes) | Map the Level for that Application | Ask for the mapping, or send a mapped name |
| Catalog change (`catalog_changed`) | A Change takes away something it used | None needed | Send a name that still exists |

The badge on Changes counts an Issue only while its fix waits for you.

#### An unknown model

For example, docs-bot asks for `llama3:8b`, it is not bound to a Connection,
and the protocol default is a hosted Connection that does not serve that
model.

- **Evidence**: how many calls failed, the status and the provider's error
  type (such as `404 model_not_found`), the time window, the Connections the
  calls went to and a few request ids to look up in Activity.
- **Diagnosis**: aiport asks each enabled Connection of that protocol for its
  model list (the same free listing the connection check uses, kept for 2
  minutes; nothing is generated and nothing costs money) and says which serve
  the model: exactly one, none, several, or unknown when a listing failed. A
  failed listing never counts as "does not serve".
- **Fix in aiport**: an Alias for that Application only, same model name, to
  the Connection that serves it. With exactly one serving Connection it is
  filled in; with none, several or unknown, choose the Connection (and the
  model, if it has another name there) yourself. **Approve** applies it as an
  ordinary Change you can revert; the Issue becomes **resolving**.
- **Copy fix for the Coding agent**: copies the fix in the Application, a
  ready instruction for the coding agent in the Application's repository;
  paste it as-is. It says what is wrong, the evidence (counts, times, the
  Connections, example request ids), what aiport found, what to change in the
  Application (ask for the Alias with a Change request, or send a model this
  gateway serves, listed for that protocol) and how to confirm the fix (one
  call with the Application's own token answers `200`; aiport then resolves
  the Issue by itself). Once the fix in aiport is applied it says no code
  change is needed. It holds no prompts, keys or tokens. **Read the fix**
  shows the same text first. The button is also on Integrate, under the
  Application's row, and the text is the same one aiport gives the
  Application's coding agent about its Issues.
- **Dismiss**: the Issue stays quiet for 7 days unless the evidence doubles.
  You can dismiss any Issue that is not resolved or dismissed yet, whatever
  its fix stands at (waiting, approved, rejected or reverted); an approved fix
  stays applied until you revert its Change.
- **Reject** (`POST /api/v1/issues/{id}` with `action: reject`): turns the
  aiport fix down and leaves the Issue as it is. The dashboard offers Approve
  and Dismiss; Reject is there for scripts.

**How an Issue resolves.** When the Application's calls for that model succeed
again, aiport marks the Issue **resolved** and records how: `aiport_fix` (the
Change you approved), `application_fix` (the Application's coding agent fixed
it) or `stopped` (the calls succeeded, or stopped failing for a whole window,
without a fix). A fix is credited only when a call shows it works: if the
Application makes no calls for that model for a whole window after you
approved the fix, the Issue resolves as `stopped`, not `aiport_fix`. If you
revert the fix, the Issue opens again and offers the fix again. A resolved
Issue reopens when the error sticks again; only errors after it resolved
count.

Issues are kept across restarts and hold only Application ids, the cause,
model names, Connection ids, counts, request ids and fixed advice text: never
prompts, responses, provider messages, keys or tokens.

#### A failing Connection

Gateway errors and timeouts count toward this Issue; a provider answer below
`500` (an unknown model, a bad request), cancellations and requests aiport
refused never do. For example, docs-bot is bound to a Connection that keeps
answering `502`. There is one Issue per Application and Connection; only
failures to the Connection the Application is bound to now count.

- **Evidence**: how many calls failed, the last status, the time window and a
  few request ids.
- **Fix in aiport**: bind that Application only to a working Connection.
  aiport proposes the configured, enabled Connection with the most successful
  calls in the window; choose another if you like, but never the failing one.
  **Approve** applies it as an ordinary Change you can revert; other
  Applications are not touched. This fix is never applied automatically, even
  with automatic fixes on: moving a whole Application to another Connection is
  always your decision.
- **Fix in the Application**: the request itself is fine. The coding agent is
  told not to retry in a tight loop, to back off and handle `5xx` answers and
  timeouts, and that it can ask for the binding with a Change request of kind
  `bind_application`.

It resolves itself when the Application's calls succeed again, on that
Connection or on the one you bound it to (`aiport_fix` or `stopped`), and
reopens if the failures stick again. Reverting the binding opens it again with
the fix offered again.

There is no automatic fallback. When a Connection is down or keeps answering
with errors, calls to it fail and the Application gets the provider's own
error. aiport never sends calls where you did not choose.

#### An unmapped Level

For example, docs-bot sends `best` on the OpenAI protocol and no `best` is
mapped there, neither for docs-bot nor shared. There is one Issue per
Application, Level and protocol: `best` unmapped on both protocols opens two,
each with its fix on its own protocol.

- **Evidence**: how many calls were refused, the time window and a few request
  ids; the summary names the protocol.
- **Fix in aiport**: map the Level for that Application only, on the protocol
  its calls were refused on. aiport proposes the Connection its other calls on
  that protocol go to (its binding, else the protocol default) and leaves the
  model to you: it never makes up a model id. **Approve** applies it as an
  ordinary Change you can revert (and a Catalog change); a Connection of the
  other protocol is refused. It is never applied automatically.
- **Fix in the Application**: ask for the mapping with a Change request of
  kind `application_alias` whose alias is the Level, or send another Level or
  task name the gateway maps.

It resolves itself when the Application's calls for that Level succeed, and
reopens if the refusals stick again; reverting the mapping opens it again with
the fix offered again. With the adapting loop off, nothing is counted and no
Issue opens.

#### A Catalog change

For example, you remove a Connection and leave docs-bot's Alias `summarize`
without one. aiport opens one Issue of cause `catalog_changed` for that
Application, one per Change, when the Application used what went away: within
request history (`telemetry.retention_days`) it sent the removed task name, or
a model name (not a task name) that went to the Connection its binding or the
moved protocol default pointed at.

An Application that never sent it, and the built-in `aiport` Application, get
the Catalog change recorded but no Issue. Applications the Change took nothing
from get none, and a Change that is an Issue's own aiport fix (or its revert)
opens none: that Issue already says so.

- **What it says**: what changed (task names removed or moved, the binding or
  protocol default and where it goes now), what the Application can still use
  (its task names and the Connections whose models it reaches), and how to
  confirm. Changes and Integrate show it with nothing to approve: it needs no
  decision from you, and it never lights the Changes badge.
- **Delivery**: like every Issue: `GET /v1/aiport/issues`, the MCP tool
  `aiport_my_issues`, the `open_issues` count in the capabilities answer, and
  `X-Aiport-Hint` on a failing call that sent a removed task name.
- **Resolution**: it resolves itself when the Application's calls succeed
  after the change (and after its last failing call with a removed name), at
  once when its coding agent reports it fixed (`aiport_issue_fixed`: there is
  nothing in aiport to confirm), or when you revert the Change. You can also
  dismiss it.

None are opened while adapting is off; the Catalog itself is still disclosed.

#### What the Application's coding agent sees

The Application's coding agent hears about its Issues over the connection it
already has with aiport, with the Application's own token (your management
token gets `403`).

- The MCP tool `aiport_my_issues`, or `GET /v1/aiport/issues` (and `GET
  /v1/aiport/issues/{id}` for one), lists that Application's open and recent
  Issues: what is wrong, the evidence, the fix in the Application
  (`application_fix`, the same text as "Copy fix for the Coding agent"), the
  root cause when you had the Issue explained (`explanation`), and whether the
  aiport fix is proposed, applied, rejected or reverted. An Application never
  sees another Application's Issues (their ids answer `404`), and never your
  notes.
- The capabilities answer (`aiport_capabilities`) shows `open_issues`, the
  number of its Issues that are open or resolving.
- A failing call that is one of an open Issue's failures carries an
  `X-Aiport-Hint` header naming that Issue, so the Application's own logs
  point the agent to it: for an unknown model, a provider `404` for the same
  model; for a failing Connection, a `5xx`, timeout or connection error from
  that Connection; for a Catalog change, a failing call that sent a removed
  task name. Other failing calls, and aiport's own refusals that are not the
  Issue's cause, name no Issue. Issue hints are sent whether or not the
  Application may request changes: telling it about its Issues is not a Change
  request.
- The `AGENTS.md` line tells later sessions to check `aiport_catalog`,
  `aiport_my_issues` and `aiport_guidance` at the start.

When the agent has fixed the Application's code, it says so with
`aiport_issue_fixed` (or `POST /v1/aiport/issues/{id}/fixed`). The Issue
becomes **resolving** by the Application fix, so you see it is being handled;
an aiport fix waiting for you keeps waiting, and you can still approve it.
Only `open` Issues can be marked (`409 issue_not_open`).

aiport does not take the agent's word for it: the Issue is resolved, as
`application_fix`, only once that Application's calls succeed again for the
same cause (for an unknown model, a successful call for that same model; for a
failing Connection, a successful call routed for that Application after it
said so). A successful call for another model does not count, and neither does
silence. If the error goes on, the Issue opens again and offers the aiport fix
again.

### Explain

Explain writes an Issue's root cause in a few sentences, so you know why an
Application fails before you decide on a fix. Without a coding agent connected
to the Application, aiport asks a Connection and model you choose once in
**Settings → Explanations**; with one, the agent reads the Application's code
instead (see [Explain by a Coding agent](#explain-by-a-coding-agent)).

**Explain** is on each Issue on Changes, and under the Application's row on
Integrate. The Connection and model are `adapting.explain` in `aiport.yaml`,
for example a local Connection serving `llama3:8b`. Until you choose one,
Explain is disabled with a hint saying so.

- **What is sent**: only the Issue's evidence: the cause, the status and the
  provider's error type, the counts and time window, example request ids, the
  model name, the Connection ids and aiport's diagnosis (which Connections
  list the model, and the models they list). Never keys, tokens, the
  Application's id or name, prompts or provider messages. The one thing in it
  an Application wrote is the model name it asked for, and only when that name
  looks like a model id (at most 128 letters, digits and `. _ : / @ + -`); any
  other name is replaced by "(a model name that is withheld because it
  contains unusual characters)", and Connection ids, model lists and the error
  type pass the same check. The same holds for what a coding agent is told.
- **Streaming**: the words appear as the model writes them. When it finishes,
  the explanation is saved with the Issue, kept across restarts, and shown
  with who wrote it. **Explain again** replaces it. If writing fails or takes
  longer than 2 minutes, nothing is saved and an earlier explanation stays; an
  explanation cut at its length limit stops there. One explanation per
  Application is written at a time.
- **Carried to the coding agent**: the saved explanation is the "Root cause"
  paragraph at the start of "Copy fix for the Coding agent", and
  `aiport_my_issues` (and `GET /v1/aiport/issues`) give it to the
  Application's coding agent, in that text and as `explanation`.
- **Activity**: the call shows in Activity under the built-in Application
  `aiport` (no configured Application may use that id), with its Connection,
  model, status and latency, like any other call. A remote Connection bills
  explanations to its key; a local one keeps them on this machine.

For scripts: `PUT /api/v1/adapting/explainer` with `{"connection": "local",
"model": "llama3:8b"}` chooses who writes (a revertable Change; both empty
turns Explain off), and `POST /api/v1/issues/{id}/explain` streams JSON lines:
`start` (who writes), `text` (the next words), then `done` (the Issue with its
saved explanation) or `error`.

#### Explain by a Coding agent

When the Application has a coding agent and a repository path, and the agent's
ACP entry point is installed, Explain starts an Agent run instead: the agent
reads the Application's code and writes the root cause, usually with the file
and the change to make. The button's tooltip names the agent; otherwise the
Connection from Settings explains, as above.

See [Connecting a coding agent](#connecting-a-coding-agent) to set one up.

- **How it starts**: only on your click, never by itself (not even with
  automatic fixes on), and never with adapting off. aiport starts the ACP
  entry point it found, by absolute path, never through a shell, with a
  minimal environment: `PATH` (the agent's own directory first, then the
  absolute directories of aiport's `PATH`, so an npm-installed adapter finds
  `node`), `HOME`, `TMPDIR`, `LANG`, `LC_ALL`, `USER`, `LOGNAME`, and, for
  Claude Code, the variables under "Where the run's model calls go" below.
  Never aiport's tokens or provider keys, and no MCP connection to aiport.
- **A throwaway copy, never your working copy**: the agent reads a git
  worktree of the repository that aiport makes for the run, detached at your
  current commit (`HEAD`, no branch), in `aiport.yaml.worktrees/`, and removes
  when the run ends. Your working copy is never where the agent works, so even
  an edit the agent's own settings allow without asking (see "What aiport
  cannot see" below) cannot touch your files. Your uncommitted changes are not
  in that copy; the repository needs at least one commit.
- **What the agent is told**: the Issue's evidence (the same as above) and
  that the run is read-only.
- **Read-only**: aiport answers every permission request itself, without
  asking you. Reading and searching inside the run's copy are allowed; every
  edit, delete, command and anything outside that copy (also your own working
  copy) is denied. aiport only ever answers "once", so nothing is remembered
  in the agent's own settings. A permission request that names a session
  aiport did not open is denied.
- **Which paths count**: every path a request names counts, not only its
  locations: the paths in its raw input (`path`, `file_path`, `cwd`, a move's
  `destination`, the files of a patch, and the base of a search pattern such
  as `/etc/**` or `../*.py`), at any depth. One path outside (or in your home
  directory, `~`) denies the request, and a read or search that names no path
  at all is denied too. This holds for Explain and Fix it alike.
- **Session modes**: when the agent offers modes (Claude Code's adapter
  offers, for example, `default`, `acceptEdits`, `plan` and
  `bypassPermissions`), aiport switches the session to the most restrictive
  suitable one: plan or read-only for Explain, one that asks before acting
  (such as `default`) for Fix it. It never picks a mode that bypasses or
  auto-accepts. If the session starts in such a mode (from your own settings)
  and the agent offers no other, aiport does not run it. The run panel names
  the mode chosen.
- **The run panel** under the Issue shows the agent's name, where its model
  calls go, its steps (thoughts, files read, permissions allowed or denied)
  and its words as they come. **Stop** ends the run at any time (the agent is
  asked to cancel, then its whole process group is ended). Nothing is saved
  from a stopped run. Closing the tab or leaving the page does not stop it:
  the run goes on, saves its explanation when it ends, and the run panel picks
  it up again when you come back.
- **Limits**: one Agent run per Application at a time. An Explain run stops
  after 10 minutes (`agent_runs.explain_minutes`), and nothing is saved.
- **What is kept**: the explanation is saved with the Issue (replacing an
  earlier one, also one a Connection wrote) and carried to the Application's
  coding agent as above. The run's log (its steps and words) is kept in memory
  only while the Issue is open or resolving, never on disk: it is gone when
  the Issue closes or aiport restarts.

**What aiport cannot see.** aiport only sees the permission requests an agent
sends. An agent's own settings can let it read, edit or run commands without
asking, and those actions never reach aiport: it cannot allow, deny or log
them. aiport limits what such settings can do, in layers:

- Every run works in a worktree made for it, never in your working copy
  (Explain's is thrown away when the run ends; Fix it's is shown to you as a
  diff), so an edit nobody was asked about lands only there.
- It picks the most restrictive session mode the agent offers (see Session
  modes above) and does not run a session that only acts without asking.
- Fix it never allows edits to agent or git configuration in the worktree, and
  any edit the agent reports ends "Allow until the next edit".
- Before a run, the run panel warns when the repository's own agent settings
  allow without asking: `.claude/settings.json` or
  `.claude/settings.local.json` with `permissions.allow` entries or a
  `defaultMode` such as `bypassPermissions` or `acceptEdits`;
  `.gemini/settings.json` with `autoAccept` or an approval mode such as
  `yolo`; and `.codex/config.toml` with `approval_policy = "never"` or
  `sandbox_mode = "danger-full-access"`. aiport only reads these files; it
  never changes them.

What it cannot check is your user-level agent settings (for example
`~/.claude/settings.json`, `~/.gemini/settings.json` or
`~/.codex/config.toml`) and any managed settings. A command that those allow
runs without aiport asking you, with your own rights, in the run's worktree
or, if the agent chooses, elsewhere.

**Warning:** do not auto-allow commands or edits in your user-level agent
settings for repositories you use with aiport, and check the warnings in the
run panel.

**Where the run's model calls go.** In **Settings → Default coding agent →
Model calls of Agent runs** you choose an Anthropic-protocol Connection (and,
optionally, a model) for them. Until you choose one, every run uses the
agent's own login and the run panel says so: aiport never sends a run's calls
to the Anthropic protocol default you did not choose for Agent runs.

- **Claude Code** is pointed at aiport by its environment only:
  `ANTHROPIC_BASE_URL` (aiport's address on this machine: the listen port on
  `127.0.0.1`, or the listen address when it names a loopback address),
  `ANTHROPIC_AUTH_TOKEN` (a credential made for this run) and
  `ANTHROPIC_MODEL` when you chose a model. That credential authenticates only
  model calls, as the built-in Application `aiport`; it cannot reach the
  dashboard, the MCP endpoint or any Application's data, is never written down
  or logged, and stops working when the run ends. The calls show in Activity
  under `aiport` with their Connection. Without a Connection chosen for Agent
  runs (or when the chosen one is not reachable), Claude Code uses its own
  login and the run panel says so.
- **Codex** uses its own login: pointing it at aiport would mean editing its
  configuration file, which aiport never does. The run panel says so.
- **Gemini CLI** uses its own login too: aiport does not speak the Gemini API
  yet. The run panel says so.

```yaml
agent_runs:
  default_agent: claude
  connection: claude-hosted   # an Anthropic-protocol Connection
  model: claude-model-id      # optional
  explain_minutes: 10
```

For scripts: `PUT /api/v1/agents/calls` with `{"connection", "model"}` (a
revertable Change), `GET /api/v1/runs` (the runs going), `POST
/api/v1/runs/{id}/stop`, and `GET /api/v1/issues/{id}/run` (the Issue's last
run and its log, while the Issue is open). The explain stream adds `run` (the
run: its id, agent and model calls) and `step` events for an Agent run;
`error` codes include `run_stopped` and `explanation_timed_out`.

#### Fix it

Fix it lets the Application's coding agent change the Application's code on a
branch of its own, and ends with a diff you review. It shows on an open Issue
when the Application has a coding agent and a repository path, as for Explain
by a Coding agent; otherwise a one-line hint under the Issue says what is
missing (for example "Add the repository path of docs-bot").

Fix it is offered only when the cause can be in the Application's code: a
model name the Application asks for that no Connection serves. For a failing
Connection the Issue says instead that the cause is not in the Application's
code and points to the aiport fix (binding the Application to a working
Connection).

- **Its own worktree and branch**: aiport makes a git worktree of the
  repository on a new branch named for the Issue, `aiport/fix-<issue id>`
  (with `-2`, `-3` and so on if that branch exists), starting at your current
  commit (`HEAD`). The worktree lives next to the configuration
  (`aiport.yaml.worktrees/`), not inside your repository. Your working copy,
  its uncommitted changes and your branch are never touched; your uncommitted
  changes are not in the worktree either. The repository needs at least one
  commit.
- **What the agent may do**: aiport answers every permission request itself.
  Reading, searching and editing, creating, deleting or moving files inside
  the worktree are allowed (symbolic links are followed, so a link pointing
  out of the worktree is outside). Never allowed, at any level of the worktree
  and whatever their case: the agents' own configuration (`.claude/`,
  `.codex/`, `.gemini/`, `.mcp.json`, `CLAUDE.md`, `CLAUDE.local.md`,
  `AGENTS.md`, `GEMINI.md`) and git's (`.git/`, git hooks in `.husky/` or
  `.githooks/`, `.gitattributes`, `.gitmodules`), so a kept branch cannot
  change how your agents or git behave. A command inside the worktree waits
  for your answer (see Approving commands below). Anything outside the
  worktree, including your own copy of the repository and a command that says
  it runs elsewhere, is denied without asking. aiport only ever answers "once"
  to the agent, so nothing is remembered in the agent's own settings. The
  agent starts as for Explain: by absolute path, with the minimal environment,
  only on your click.
- **What the agent is told**: the Issue's evidence, the saved explanation if
  there is one, the branch it works on, and that it must not commit, merge or
  push.
- **The run panel** shows its steps and the permissions allowed or denied,
  with **Stop**. The run goes on if you leave the page.
- **Time limit**: a run stops after 30 minutes (`agent_runs.fix_minutes`).
  What it changed by then stays in its worktree and waits for your review with
  **Keep the branch** and **Discard**, as after a finished run (state
  `timed_out`).
- **Nothing left behind**: Stop, a run that fails and a run that changed
  nothing (also one that timed out or went unanswered before changing
  anything) remove the worktree and the branch; the Issue stays open and shows
  why the run ended.
- **Cleanup at startup**: a Fix it run that was going when aiport stopped (or
  crashed) is marked stopped the next time aiport starts, and its worktree and
  branch are removed; so are the throwaway worktrees of Explain runs that were
  going. Only directories inside `aiport.yaml.worktrees/` and branches named
  `aiport/fix-…` are ever removed.
- **What is kept**: the branch name, the commit it started from, the diff
  summary and the changed files are saved with the Issue (and, once kept, the
  commit aiport made), also across restarts. The diff itself is read from the
  worktree or the branch when you look at it. The run's log is kept as for
  Explain.

**Approving commands.** When the agent asks to run a command (tests, a build),
the Issue shows the whole command exactly as the agent sent it, as plain text
(it is the agent's, not aiport's), never cut: a long one scrolls. Characters
that would not show as themselves (control characters, invisible and
direction-changing characters such as U+202E, the right-to-left override,
unusual spaces) are written as escapes, so what you read is what runs; a list
of arguments is shown shell-quoted (`echo 'a b'`). A command longer than 16
KiB is denied without asking, never shown in part. There are three answers:

| Answer | What it allows |
| --- | --- |
| **Allow once** | This one time. |
| **Allow until the next edit** | This time and every time the run asks for the exact same command again (the same list of arguments, or the same string, in the same directory: `echo 'a b'` is not `echo a b`), until the agent changes a file, whether it asked to or not. Another command, or the same one in a later run, asks again. |
| **Deny** | Nothing: the agent is told no and goes on without it. |

An edit can change what a command does, so "Allow until the next edit" asks
again after one. aiport only sees the edits the agent asks about or reports: a
command you allowed that changes files itself (a code generator, a formatter)
does not end the allowance, so choose **Allow once** for commands that write
files.

The agent runs an allowed command itself, in its own process, in the
worktree's directory; aiport only answers the permission, and what the command
changes in the worktree shows in the diff. While a command waits, the Changes
badge in the Monitor menu counts it. A command left unanswered for 5 minutes
stops the run: what it changed so far stays in its worktree, and the Issue
shows it for review with **Keep the branch** and **Discard**, as after a
finished run (state `stopped_waiting`).

**The diff.** When the agent ends its turn, the Issue shows **Review the
change** with the branch, the summary ("2 files changed, 3 insertions(+), 1
deletion(-)") and the whole diff. Nothing is committed yet.

- **Keep the branch**: aiport commits the diff on the new branch, as one
  commit authored by `aiport <aiport@localhost>`, and removes the worktree.
  This is the only commit aiport ever makes, and only when you click Keep: on
  the new branch, never on yours, with hooks off. The branch stays in your
  repository for you to look at, merge or delete; aiport never commits to your
  branch, merges or pushes. The Issue becomes resolving as an Application fix
  and resolves when the Application's calls succeed again, that is after you
  merged and deployed the change. "Copy fix for the Coding agent" names the
  branch.
- **Discard**: aiport removes the worktree and deletes the branch; the Issue
  stays open, and you can Fix it again.

**git, safely.** aiport finds git once, at startup, by absolute path
(`/usr/bin/git` when it is there, otherwise the first `git` in its `PATH`; the
startup output names it). It runs git with a minimal environment, without your
global or the system's git configuration and attributes
(`GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`,
`GIT_ATTR_NOSYSTEM=1`, `core.attributesFile=/dev/null`), and with hooks,
fsmonitor, commit signing and external diff programs turned off. What the
repository itself configures in its own `.git/config` and `.gitattributes`
still applies (for example a clean or smudge filter it defines): it is your
own repository, so check those before using Fix it on a repository you did not
set up. aiport removes a worktree only when its directory is directly inside
`aiport.yaml.worktrees/`; any other path it finds recorded is left alone and
logged.

```yaml
agent_runs:
  fix_minutes: 30
```

For scripts: `POST /api/v1/issues/{id}/fix` starts it (`202` with the run;
`409` with `fix_unavailable` and the hint, `fix_not_for_cause` for a failing
Connection, `fix_waiting` while an earlier diff waits for you,
`agent_run_running`, `issue_not_open` or `worktree_failed`). `GET
/api/v1/issues/{id}/fix` gives `branch_fix` and the `diff`; `POST
/api/v1/issues/{id}/fix/keep` and `…/fix/discard` keep or discard it. `GET
/api/v1/issues` says per Application in `fix_by` whether Fix it is available,
with which agent and time limit and for which `causes` (with `cause_hint` for
the others), or the hint; each Issue carries its `branch_fix`.

Commands waiting for you: `GET /api/v1/approvals` (each with `id`, `run`,
`application`, `issue`, `command`, `expires_at_ms`). Answer with `POST
/api/v1/approvals/{id}` and `{"answer": "allow_once"}`, `"allow_run"` (allow
until the next edit) or `"deny"` (`404 approval_not_found` once answered or
its run ended). `GET /api/v1/runs` gives each run's `waiting` count, and `GET
/api/v1/issues/{id}/run` the Issue's waiting `approvals`. These are management
routes, never open to an Application token; answering needs the dashboard's
CSRF header.

### Connecting a coding agent

Connecting a coding agent to an Application lets aiport go beyond the Issue's
evidence, on your click: Explain reads the code and tells you the root cause,
and Fix it makes the change on a new branch and shows you the diff to keep or
discard. Explain only reads, nothing runs without your click, and aiport never
commits to your branch, merges or pushes.

An Issue tells you that an Application keeps failing and offers the fixes
aiport can see from its side. Without an agent, aiport still explains, more
briefly, from the Issue's evidence alone (with the Connection and model chosen
under Explanations in Settings).

aiport can start a coding agent installed on this machine to work on an
Application's Issues, only when you click (Agent runs). Each Application can
have its own agent, **Claude Code**, **Codex** or **Gemini CLI**, and the path
of its git repository. Set them in step 6 of the onboarding wizard (or from
the **Agent** column on Integrate) and **Save**. Each save is one revertable
Change on the Changes page. **Default coding agent** on the Settings page is
the agent of every Application that has none of its own.

- **What aiport looks for.** aiport talks to agents through the Agent Client
  Protocol (ACP). Gemini CLI speaks it itself (`gemini --acp`); Claude Code
  and Codex need an ACP adapter: `claude-agent-acp` (npm package
  `@agentclientprotocol/claude-agent-acp`) and `codex-acp`
  (`@agentclientprotocol/codex-acp`). aiport finds programs by absolute path,
  in the absolute directories of its own `PATH` and the usual install places
  (`~/.local/bin`, `~/.npm-global/bin`, `/opt/homebrew/bin`, `/usr/local/bin`
  and a few more); it does not run them to look.
- **Nothing is installed for you.** When an agent or its adapter is missing,
  the step shows the exact command to install it, such as `npm install -g
  @agentclientprotocol/codex-acp`; run it yourself, then refresh. An agent
  that is not installed cannot be chosen.
- **The repository path** must be absolute and the top directory of a git
  repository (where `.git` is); otherwise it is refused with the reason.
  aiport stores it with symbolic links resolved. Agent runs never work outside
  it.

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

`GET /api/v1/agents` lists what aiport found (`installed`, `ready`, the ACP
entry point, and `install` when something is missing).

### What waits for your decision

Suggestions and Change requests are the changes others propose to aiport's
configuration, and each one waits on the Changes page until you approve,
dismiss or reject it. Nothing in this list changes aiport until you decide.

- **Suggestions** come from aiport itself, from patterns in request history,
  with the evidence (how many requests, over which time window). aiport
  suggests pairing when requests keep arriving with a missing or unknown token
  (`adapting.unknown_token`); this one is informational: open the Integrate
  page to pair the application, or **Dismiss**. An Alias for a model one
  Application keeps getting `404` for, and binding an Application to a working
  Connection when its calls to one Connection keep failing, come as
  [Issues](#issues) instead. Thresholds are under `adapting:` in the
  configuration. Suggestions from the votes are different: they live on the
  Feedback page (see [Suggestions from the
  votes](#suggestions-from-the-votes)).
- **Change requests** come from Applications you allowed to request changes.
  An Application names a task (for example the alias `review-code`), not a
  provider: you decide which Connection and model serve it. It may add advice
  ("suggests big-model on hosted", with its reasons); the advice pre-fills the
  form, and you can choose otherwise now or change it later
  (`origins[].aliases` in the configuration file, or a new Change request)
  without touching the Application's code.

Each Change request shows what it asks for; **Details and reasons** opens the
full details, why it is needed, the argument for the Scope asked for, and who
filed it (as stated). Identical requests from several Applications are merged
into one item, each with its own argument; one decision answers them all.
Request text is plain text from the Application: it is shown as is and never
sent to a model. Each field is at most 2,000 characters, 6,000 in total.

The **Waiting / Accepted / Rejected** filter on **Suggestions and Change
requests** picks what the list shows. Waiting is what needs you. Accepted and
Rejected show past decisions, most recent first, with who decided, when and
the note; Rejected includes dismissed Suggestions. Decided items are kept as
long as request history (`telemetry.retention_days`); an accepted one's Change
stays in **Applied changes**, where you can revert it.

Kinds aiport cannot apply are kept as "not supported" and counted under
**Asked for, not supported yet** (kinds and counts only); they need no
decision.

#### Scope

Every decision has a Scope: **only** the requesting Application, or **every
Application**. The requester argues for one; you choose. For an Alias, every
Application means a shared Alias; for a binding, it means the protocol's
default Connection.

#### Approve, Dismiss, Reject

- **Approve** applies the Change at once. For an Alias you confirm the alias
  name, the Connection and the model; for a binding, the Connection. You can
  add a note of up to 2,000 characters.
- **Dismiss** a Suggestion: aiport stays quiet about it for 7 days unless the
  evidence doubles.
- **Reject** a Change request: your note goes back to the Application, which
  sees it when it lists its requests.

#### Allowing Change requests

An Application may request changes only after you allow it: tick **May request
changes** when approving its pairing, or use the toggle on Integrate or in the
Applications table on Routing. Until then it gets `403
change_requests_not_allowed`. Allowed Applications also get hints when an
error is one a Change could fix (`error.hint` on `target_unconfigured` and
`target_disabled`, and `X-Aiport-Hint` on a provider `404` for the requested
model). Change requests are limited to 20 per Application per hour (`429
change_requests_rate_limited`); a request a policy refuses spends none of that
limit.

**How an Application writes a good Change request.** The coding-agent brief
teaches the fields: `kind` (from `change_kinds`: `application_alias` or
`bind_application`), `need` (one line), `details`, `why` (the error code, the
`X-Aiport-Request-Id`, what it tried), `scope` (`application` or
`every_application`) and `scope_case`, with the optional `filed_by`, `alias`,
`prefer_connection` and `prefer_model`. A `400 invalid_change_request` lists
every problem in `error.problems`. The answer, read with `aiport_my_requests`,
is `pending`, `accepted`, `rejected` (with your note) or `not_supported`.

### Applied changes and Revert

Applied changes lists every configuration Change made through aiport, newest
first, with who made it, and **Revert** undoes any one of them. The list
survives restarts, so you can always get back to an earlier configuration.

A Change is made by you, by an approved Suggestion or Change request, or by
**aiport (automatic)** for an automatic fix. **Revert** restores the
configuration file exactly as it was before that Change and makes it live at
once. If the file was edited after a Change, revert the newer Changes first;
aiport refuses rather than overwrite them. Applied Changes are never pruned.

#### Automatic fixes

Automatic fixes are off by default. Turn them on with the **Automatic fixes**
switch under Applied changes (or `adapting.auto_fix: true` in `aiport.yaml`);
the switch itself is a Change you can revert. When they are on, aiport applies
an Issue's aiport fix itself, without waiting for you, only when all of this
holds:

- the diagnosis is certain: exactly one enabled Connection serves the model
  (never with none, several or unknown);
- the error is sticky: the Issue's threshold (`adapting.unknown_model`) was
  reached;
- the fix touches only that Application: an Alias for it alone. Binding a
  whole Application to another Connection, Level mappings, shared Aliases,
  protocol defaults or anything for every Application are never automatic;
- you have not refused that fix (below), and the adapting loop is on.

For example, docs-bot keeps getting `404` for `llama3:8b` and only the local
Connection serves it: aiport adds the Alias `llama3:8b` → local for docs-bot
and the Issue becomes resolving. The Change is recorded as made by **aiport
(automatic)**, and Applied changes shows it as "aiport fixed: …" with
**Revert**; the Issue shows "aiport fixed this automatically" with its own
Revert. Guardrail Extensions judge automatic fixes like your approvals.

If you revert an automatic fix, aiport records it and never applies the same
fix automatically again for that Application and cause, also after a restart;
the Issue opens again and its fix waits for your decision, which you can still
approve. A fix you rejected is never applied automatically either, also if you
turn automatic fixes on later; you can still approve it yourself. To keep
these decisions, an Issue that holds a fix you reverted or rejected is kept
beyond request-history retention (`telemetry.retention_days`), even once it is
resolved or dismissed; it holds no more than any other Issue (ids, model
names, counts).

#### The sound

The speaker button next to Applied changes turns on a soft sound for when
something new waits in Changes. It is off by default and remembered in this
browser only.

### Compress requests

An application can ask aiport to shrink a request before it reaches the
provider, and aiport measures whether that hurts the answers. Compression is
off unless the application asks for it, per request, with the
`X-Aiport-Compress` header.

| Method | What it does |
| --- | --- |
| `whitespace` | Removes trailing spaces, normalises line endings, collapses three or more blank lines to one; indentation is kept. |
| `json` | Rewrites text that is entirely JSON (typical tool results) compactly. |
| `dedupe` | When a block of 1 KiB or more repeats, earlier copies become a short note; the latest copy stays. Lossy. |

`all` selects every method; `off` or no header means none. An unknown method
returns `400 unknown_compression`. Only message text is changed; tool
definitions, schemas and metadata are never touched. If compression fails or
saves nothing, the original bytes are sent. Every compressed response carries
`X-Aiport-Compression` with the methods, `status=applied`, `no_gain` or
`failed`, and the bytes before and after.

The [Compression](#compression) page shows the savings and a verdict per
method combination. The full description, including the evaluator, is in
[compression](compression.md).

### Add Extensions

Extensions are programs you add that take part in deciding about Change
requests, so you can enforce your own rules without changing aiport.

- A **policy** runs when a Change request arrives and may refuse it; the
  Application gets `403 change_request_refused` with the reason.
- A **guardrail** runs when you approve, before the Change is applied, and may
  block it (`guardrail_blocked`); Changes then shows which Extension blocked
  it and why, and the item keeps waiting.

List them under `extensions:` in `aiport.yaml` with an absolute path. aiport
runs only listed programs, with a timeout; anything but an explicit allow
denies. The input and output protocol is in [Extensions](extensions.md). The
**Extensions** panel on Changes lists them.

### Suggest to aiport and Share upstream

You can give back to aiport itself in two ways, and both are drafts you check
first; nothing is sent on its own.

- **Suggest to aiport**: under **Asked for, not supported yet**, for a kind of
  Change Applications asked for that aiport cannot apply.
- **Share upstream**: in the **Extensions** panel of Changes, for an Extension
  you listed. The draft names the hook and the input protocol, never the
  Extension's name, path or settings; you add the program yourself after
  removing anything private.

A draft holds kinds, hooks and anonymous counts only, never configuration,
keys, paths, Application names or request text. A kind named by an Application
is its own words, so check it. Both the title and the body are editable, and
every way out uses the text as you edited it: **Open on GitHub** (when
`upstream.repository` is set), **File with gh…** (after you confirm, aiport
runs `gh issue create` with your own gh login), or **Copy text**.

### Turn the adapting loop off

The adapting loop is the part of aiport that watches request history and
proposes fixes: Suggestions, Issues, Change requests and Explain. You can turn
it off and keep everything else.

Set this in `aiport.yaml` and restart aiport:

```yaml
adapting:
  disabled: true
```

With adapting off:

- aiport detects no Suggestions and opens no Issues, and Changes shows nothing
  new;
- Change requests and `/v1/aiport/issues` answer `409 adapting_off`, and the
  Issue tools are not offered over MCP;
- capabilities carry no `open_issues`, no hints are sent, and Explain and
  Agent runs are unavailable;
- `/_connect` and both briefs mark change requests, Issues, their MCP tools
  and `X-Aiport-Hint` as off, with the reason.

Still working: forwarding is unchanged; Revert works on earlier Changes; the
Catalog is still disclosed (`/v1/models`, `X-Aiport-Catalog`,
`/v1/aiport/catalog`, `aiport_catalog`) and Catalog changes are still
recorded, but no Catalog-change Issues are opened; Levels, guidance, the
communication standard and votes keep working. See
[configuration](configuration.md) for the detector thresholds and the check
interval (`adapting.interval_seconds`).

### Use a name instead of 127.0.0.1

You can open the dashboard at `http://aiport.localhost:8787/` instead of the
IP address, because browsers resolve any `*.localhost` name to this machine on
their own. The name is for your browser only; applications and coding agents
keep the numeric address.

Add this to `aiport.yaml` and restart aiport:

```yaml
hostnames: [aiport.localhost]
```

The dashboard works at the name at once, and each page has its address there
(`http://aiport.localhost:8787/integrate`). Names must be lowercase and end in
`.localhost`. aiport still listens on `127.0.0.1` and still refuses every name
it was not given, with `403 invalid_host` or `invalid_origin`.

Applications and coding agents are always given the numeric address
(`http://127.0.0.1:8787`): the agent instruction, both briefs, the
communication standard, `/_connect`, the `.aiport` lines on a token card and
the MCP entry all use it, even when hostnames are set. A name needs a lookup
the system resolver may not do, resolves to `::1` first (aiport listens on
IPv4), and a coding agent's sandbox sends it through its network proxy, which
refuses it; `127.0.0.1` has none of these problems.

## A tour of the dashboard

The dashboard is where you run aiport: it shows the traffic, holds the
Connections and routing, connects Applications, and lists what waits for your
decision. The sidebar has three groups: **Monitor** (Overview, Activity and
Changes, the pages you use every day), **Configure** (Integrate, Providers,
Routing, Agents, Compression and Feedback) and, at the bottom, **Workspace**
(Tasks and Settings).

Each page has its own address (`/integrate`, `/changes`, `/manual` and so on),
so reloading stays on the page, Back and Forward move between pages, and a
link opens the page it names.

The toolbar at the top right has the help button (this manual), the Light /
Dark / System appearance switch and Refresh (on Tasks, also the chime). The
dashboard refreshes itself every 10 seconds while it is open; Refresh fetches
everything at once, for example right after you changed the configuration file
by hand. The ⓘ next to a page title or panel opens the opening of the matching
manual section, with a link to the whole section. The bottom of the sidebar
shows the address aiport listens on and **Sign out** (or "Authentication off"
when it is turned off).

| Page | What it is for |
| --- | --- |
| [Overview](#overview) | Traffic at a glance, and one line for each thing that needs you. |
| [Activity](#activity-and-history) | Every recorded request, with why it failed. |
| [Changes](#changes) | Issues, decisions, and every Change with Revert. |
| [Integrate](#the-integrate-page-and-the-two-briefs) | Your Applications, onboarding and the instructions for coding agents. |
| [Providers](#connections-and-keys) | Your Connections and their keys. |
| [Routing](#routing-and-protocol-defaults) | Protocol defaults, Levels, the Routing map and the Applications table. |
| [Agents](#agents) | Coding agents launched through `aiport agent`. |
| [Compression](#compression) | Opt-in request compression and whether it hurts answers. |
| [Feedback](#feedback) | Votes per Application, module, model and effort, and Suggestions from them. |
| [Tasks](#tasks) | A queue of work for coding agents. |
| [Settings](#settings) | Sign-in, the default coding agent, Explanations and the configuration file. |

### Overview

Overview shows the shape of your AI traffic at a glance: request counts,
average and P95 latency, error rate, hourly traffic, the provider distribution
(with each provider's logo) and the latest requests.

It also shows one line for each thing that needs you: when something waits for
your decision (with **Review in Changes**), when an Application's coding agent
has not confirmed the current guidance version (with **See Guidance in
Integrate**), and when the latest request on a route failed for a reason in
your configuration (with the reason and the fix; it goes away once a request
there succeeds). A banner says when request history is off, kept for this
session only, or incomplete.

### Activity and history

Activity lists every recorded request, so you can see who called what, with
which name and effort, what served it and why it failed. Only metadata is
recorded: time, route, Application, model, effort, module, status and latency,
never prompts, responses or headers.

Filter it by period, Application ("Origin"), Connection ("Target"), outcome
and the name sent; **Load older requests** goes further back, and **Details**
opens one request.

**Turning it on.** With `telemetry.enabled: true`, aiport records request
metadata in a private SQLite file next to the configuration (`aiport.yaml.db`,
mode 0600). Recording never slows the forwarding path; if records are dropped
or fail to save, the banner on Overview and Activity says how many. If the
database cannot be opened, aiport says so and keeps the latest 500 records in
memory for that session.

**Pruning.** Rows older than `telemetry.retention_days` (default 7) are
deleted at startup and hourly. The same setting decides how long decided
Suggestions and Change requests, Catalog changes, and resolved or dismissed
Issues are kept (dismissed ones at least 7 days, the quiet period). Open
Issues are kept until they resolve. Applied Changes are never pruned. Deleted
rows can linger in SQLite's files; deletion is not secure erasure.

**Request ids.** Every response carries `X-Aiport-Request-Id`. When an
application logs it, you can find the exact call in Activity, and votes refer
to it.

**What was asked and what served it.** For each model call Activity keeps the
name the Application sent (a Level, a task name or a model id), its effort and
where that came from (the `X-Aiport-Effort` header, the provider's effort
field, or neither, so `medium`), its module, and the Connection and model
aiport sent it to. The Model column shows `balanced` with `chat/completions →
small-model · low` beneath it. **Details** says it in a sentence, for example
"Asked for balanced at low effort (from the X-Aiport-Effort header); served by
hosted with small-model." A refused call has no serving model. The **Name**
filter shows only calls that sent one name, so you can check a Level or
per-effort mapping works and compare the models behind it. Older records, from
before this was kept, show none of it.

**Why a request failed.** For a failed request, Activity shows the reason next
to the status, for example `503 · Refused: no Connection for the Anthropic
protocol`. **Details** shows the code, what it means, what to do and a link to
the page that fixes it. aiport keeps only short codes for this: its own
refusal code, a connection error class, and the provider's error type (such as
`rate_limit_error`) when the provider sends one. It never keeps the provider's
error message, the request or the response. The reasons and their fixes are in
[Activity reasons and their fixes](#activity-reasons-and-their-fixes).

### Changes

Changes is where aiport adapts to the Applications that use it, and only with
your approval: open Issues, what waits for your decision, and every
configuration Change made through aiport, with Revert. It stays quiet; a badge
appears in the Monitor menu only when something needs a decision.

The page has these panels, top to bottom:

- **Issues**: open Issues with their evidence, Explain, Fix it, and the fix on
  each side. See [Issues](#issues).
- **Suggestions and Change requests**: what waits for your decision, with the
  **Waiting / Accepted / Rejected** filter. See [What waits for your
  decision](#what-waits-for-your-decision).
- **Asked for, not supported yet**: kinds Applications asked for that aiport
  cannot apply, with **Suggest to aiport**. See [Suggest to aiport and Share
  upstream](#suggest-to-aiport-and-share-upstream).
- **Extensions**: the Extensions listed in `aiport.yaml`, with **Share
  upstream**. See [Add Extensions](#add-extensions).
- **Applied changes**: every Change, newest first, with **Revert**, the
  **Automatic fixes** switch and the sound. See [Applied changes and
  Revert](#applied-changes-and-revert).

The badge counts an Issue only while its fix waits for you, and also counts
commands a Fix it run waits for you to approve.

### The Integrate page and the two briefs

Integrate is where you connect an application: your Applications, the
onboarding wizard, and the instructions aiport writes for the coding agents
that do the work. The **Instructions** panel is at the top, collapsed unless
you have no Applications; the Applications list and the wizard are below it.

The two briefs are generated from the live gateway, from its real address and
the protocols it has configured, and served at `/_connect`: the **Application
brief** for the agent wiring the application, and the **Coding agent brief**
for agents working in its repository later. With the communication standard,
they are everything a coding agent needs.

#### Applications

The Applications list on Integrate shows every Application, with what it does
through aiport and what to do next. Pairing requests waiting for your approval
are always at its top.

- **Pairing requests waiting** each show the code, **Approve** (optionally
  binding a Connection and allowing change requests) and **Reject**; see
  [Applications and pairing](#applications-and-pairing).
- **Columns**: name and client id; **Connection** (or the protocol default);
  **Changes**, the **May request changes** toggle; **Last call** (in the last
  7 days, when activity recording is on); **Issues**; **Guidance**; **Agent**;
  **Onboarding** progress, such as `3/6`; and at the end **New token** and the
  trash button.
- **Agent** names the Application's coding agent once it has one and a
  repository path (**Connect** otherwise). Select it to change them in step 6
  of the wizard.
- **Issues** shows how many open or resolving Issues the Application has and a
  short cause, such as "Unknown model llama3:8b" (None when there are none).
  Select it to list them under the row: each with its summary, **Explain**,
  **Open in Changes** (opens that Issue on the Changes page, where you approve
  or dismiss it) and **Copy fix for the Coding agent**.
- **Guidance** shows whether the Application's coding agent follows aiport's
  current guidance (see [Guidance versions](#guidance-versions)): **Up to
  date**, **Behind v1 → v4** (its confirmed and the current version), or
  **Unknown** when it never confirmed one. While it is not up to date, **Copy
  refresh instruction** copies one paragraph to paste to that coding agent.
  "Sends Levels" is only a hint that its code may already use Levels (its
  calls sent Levels in the last 6 hours); it never counts as confirming. Under
  it, the **Modules** and **Votes** chips say whether its calls in the last 7
  days named their module and whether its users' votes arrive; when one is
  missing, hand its coding agent the communication standard.
- **Catalog** (under the Application's name) shows its Catalog: every name it
  can send with the Connection and model that serve it (as on the Routing
  map), the models its Connections list, and its recent Catalog changes,
  newest first, the same changes the Application reads at
  `/v1/aiport/catalog`.
- **Levels** (next to Catalog) shows its Levels grid, with its own mappings
  marked **Own**; see [Map Levels in the
  dashboard](#map-levels-in-the-dashboard).
- Click an Application's name to open its token and removal actions under its
  row (**Issue new token**, **Remove application**); see [Give a token
  yourself, or remove an
  Application](#give-a-token-yourself-or-remove-an-application).
- The **Approved / Rejected / All** filter picks what the list shows. Rejected
  Applications are hidden until you choose **Rejected** or **All**; each shows
  when it was rejected, with **Forget**.
- Select an Application to have the wizard follow it. **Onboard a new
  application** makes the wizard follow a new one, and **Add with a token**
  adds one without pairing.

#### The instructions

You do not have to write instructions for the coding agent doing the work:
aiport writes them itself, from this gateway's real address and configured
protocols, and serves them at `/_connect`. The **Instructions** panel holds
the communication standard card and the two brief tiles.

- **Communication standard** (`/_connect?format=standard`): the card shows the
  prompt that hands it to a coding agent, with **Copy instruction**, **Copy
  URL only** and **Read the standard**. See [The communication
  standard](#the-communication-standard).
- **Application brief** (`/_connect?format=handover&for=application`): for the
  coding agent changing an application so its LLM calls go through aiport. It
  covers pairing, keeping the token in `.aiport` (`AIPORT_BASE_URL`,
  `AIPORT_APP_TOKEN`) and what to do when it is lost, pointing the OpenAI or
  Anthropic SDK at aiport, removing provider keys from the application, model
  names (send a Level or a task name, never a provider's model id; read `GET
  /v1/models` for the names; `level_unmapped`), saying how hard a call is with
  `X-Aiport-Effort`, naming the module with `X-Aiport-Module` and sending
  votes, watching the `X-Aiport-Catalog` header, errors and retries, logging
  `X-Aiport-Request-Id`, Issues (what they are, that the coding agent reads
  them with `aiport_my_issues`, that a Change that takes something away opens
  a `catalog_changed` one, and that aiport may fix routing itself as a
  revertable Change), and how to verify. Its **Python / TypeScript / curl**
  picker switches the SDK snippets.
- **Coding agent brief** (`/_connect?format=handover&for=agent`): for coding
  agents working in that application's repository later. It covers the project
  MCP entry, the line for the repository's `AGENTS.md`, keeping the guidance
  current, the MCP tools and their HTTP equivalents, checking the Catalog,
  Issues and guidance at the start of a session, saying when an Issue is
  fixed, naming the need (a Level or task name and an effort) rather than a
  provider, how to write a good Change request, waiting for your decision, the
  permission, and hints.

Each tile shows who the brief is for, **Copy whole brief** (the brief as
markdown, to paste to the coding agent) and its address; open a tile to see
its steps and snippets, each with its own **Copy**. Snippets for a protocol
this gateway has not configured are marked "not configured".
`/_connect?format=handover` returns both briefs, one after the other, and
`/_connect?format=markdown` holds everything, starting with **Start here**.

### Connections and keys

A Connection is a provider aiport can forward to, and Providers is the page
where you add them, paste their keys and check they work. Applications never
see provider keys: they send their own aiport token, and aiport adds the key
upstream.

aiport supports OpenAI, Anthropic, Amazon Bedrock, Meta Model API, and local
servers such as Ollama or llama.cpp. Each Connection speaks one protocol,
`openai` or `anthropic`. Each provider shows its logo on its card, in the add
list and in the Overview's provider distribution; llama.cpp and custom
providers show a neutral icon.

Connections are grouped as **Remote APIs** and **Local**. Both groups start
collapsed; each heading shows how many Connections it holds and which need a
key or are unhealthy. Click a heading to open it.

- **Add a connection** from a preset, or any other server that speaks the
  OpenAI or Anthropic API natively (remote URLs must use HTTPS). **Remove**
  takes one away. Each is a Change you can revert in Changes.
- **Paste a provider API key** into a remote Connection's card. It applies at
  once, is stored only in the private `aiport.yaml.env`, never in
  `aiport.yaml`, and is never shown again. You can also set the key in the
  shell aiport starts from, under the Connection's key variable (see
  [configuration](configuration.md)); provider keys are never asked for in
  the terminal. Local Connections need no key.
- **Check connection** (or **Check all** per group) sends one authenticated
  `GET /models` to the provider. It generates nothing and costs nothing; it
  shows whether the provider is reachable and accepted the key. A failed check
  says why: the connection error (host name does not resolve, connection
  refused, TLS failed, timed out), or the HTTP status with the provider's own
  error type (for example `HTTP 401, authentication_error`), and what to do. A
  passing check does not prove every model works.
- **Show models** lists the models that Connection serves, using the same free
  `GET /models` listing (OpenAI-, Anthropic- and Ollama-shaped answers). After
  a check, the button shows the count ("Show 12 models"). aiport keeps only
  the model ids, in memory, and reuses a list for two minutes; **Refresh**
  lists again, and so does **Check connection**. If the listing fails (no key,
  key rejected, no model list, unreachable), aiport treats that Connection's
  models as unknown, not as "serves nothing". Disabled Connections are not
  listed.
- **Make OpenAI default** / **Make Anthropic default** sets the Connection a
  protocol uses when nothing else picks one.

Provider presets, endpoints and key variables are in
[configuration](configuration.md).

#### Removing a Connection, and when one fails

You can always remove a Connection, and aiport re-points everything that used
it in the same step. **Remove** on its card opens a dialog that lists
everything still using it: a protocol default, an Application bound to it, an
Alias (shared or an Application's own), Explanations, the model calls of Agent
runs and the compression evaluator. Each item has a Connection picker, already
set to the best match:

1. a Connection of the same protocol that lists the same model (for a binding
   or a protocol default: the one that lists most of the removed Connection's
   models);
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
follow), and updates as you change a picker. aiport itself is among them when
the removal moves what its own model calls (Explanations and Agent runs) use:
`agent_runs.connection`, or the protocol defaults when that is not set. Its
Catalog changes are kept like any Application's, but it gets no Issue: it has
no coding agent to tell.

Confirming applies all the re-pointing and the removal as **one Change**: one
revert in Changes restores the Connection and everything that used it. Each
affected Application is told through its Catalog, and an Application the
removal took something it used from gets a Catalog-change Issue (see [A
Catalog change](#a-catalog-change)).

Without the dialog (a `DELETE /api/v1/targets/{id}` with no plan), aiport
still refuses to remove a Connection something uses (Explanations and Agent
runs' model calls included), and the message names what uses it. The dialog
reads `GET /api/v1/targets/{id}/dependents` and sends the plan as `{"plan":
{"<item>": "<Connection id or empty>"}}`.

There is no automatic fallback. When a Connection is down or keeps answering
with errors, calls to it fail and the Application gets the provider's own
error. When it keeps happening for one Application, aiport opens an Issue that
offers to bind that Application to a working Connection; you approve it (see
[A failing Connection](#a-failing-connection)). Moving an Application to
another Connection is never applied automatically, so aiport never sends calls
where you did not choose.

### Routing and protocol defaults

Routing decides which Connection and model serve each call, and the Routing
page is where you set the protocol defaults, the Levels, and see where every
name goes. aiport speaks the providers' native APIs and does not translate
between them: an OpenAI SDK call needs a Connection that speaks the `openai`
protocol, and an Anthropic SDK call one that speaks `anthropic`.

The page has, top to bottom: **Protocol defaults** (the Connection each
protocol uses when nothing else picks one), the **Levels** panel, the
**Routing map**, **Connect an application** (the two base URLs Applications
use, `http://127.0.0.1:8787/v1` for the OpenAI protocol and
`http://127.0.0.1:8787` for the Anthropic protocol, the `/_connect` protocol
description and the telemetry queue), the **Applications** table, and a
pointer to the instructions on Integrate. Model allowlists are planned, not
built.

#### How aiport picks a Connection and model

For each call aiport picks the Connection and model in this order; the first
that applies wins:

1. **A module route**: the Application has a route for the module the call
   names (`X-Aiport-Module`) and the name it sends. See [Apply a Suggestion as
   a module route](#apply-a-suggestion-as-a-module-route).
2. **A Level** (`fast`, `balanced`, `best`): the Application's own mapping on
   this protocol, else the shared one. A Level mapped on neither is refused
   with `400 level_unmapped`.
3. **The Application's own Alias** for the requested name, if it has one (it
   may name another Connection).
4. **A shared task name marked `everywhere`**, on its own Connection.
5. **The Connection the Application is bound to**, if any, else **the protocol
   default** set under Protocol defaults, with the shared Aliases of that
   Connection.

A Level or an Alias with per-effort models then picks the model for the call's
effort (see [Effort and reasoning](#effort-and-reasoning)). Only the `model`
field is replaced, plus the reasoning effort in the cases that section
describes. A protocol with no Connection answers `503 target_unconfigured`; a
disabled Connection answers `503 target_disabled`; a Connection of the other
protocol answers `400 incompatible_protocol`.

#### The Routing map

The Routing map answers "if this Application sends this name, which Connection
and model serve it?" for every Application at once. It is built from each
Application's Catalog, the same one its `GET /v1/models` answers, so the two
never disagree.

Each Application lists every name it can send: its own task names, the shared
task names of the Connections it reaches (and the `everywhere` ones), the
Levels mapped for it (its own mapping or the shared one, per protocol), and
**Any other model name**, which goes unchanged to its binding or to the
protocol default (with the number of models that Connection lists). Each row
shows the protocol, the Connection with its provider's logo, and the model
aiport sends upstream, with its per-effort models under it when it has any.
For example, docs-bot sending `fast` on the OpenAI protocol reaches the local
Connection with `llama3:8b`. Choose an Application or a provider to narrow the
map.

The same view for one Application, with its recent Catalog changes, is under
**Catalog** on its row on Integrate. The pencil on a task name's row edits its
per-effort models (see [Per-effort models of a task
name](#per-effort-models-of-a-task-name)); a Level's are edited in the Levels
panel.

Scripts can read the map with the management token at `GET /api/v1/routing`
(`?application=<id>` for one Application, with its `changes`): Application
ids, names, task names, model names and Connection ids only, never keys or
addresses.

### Agents

The Agents page lists the coding agents whose own model calls go through
aiport, each with its own token, so they never see your provider keys. It
shows the configured Applications with their Connection (or the protocol
default) and a launch line; when none is configured it says how to create one
with `aiport init --agent codex,claude`.

You launch such an agent from a terminal with `aiport agent codex --model
MODEL` or `aiport agent claude --model MODEL` (add `--dry-run` to see what
would start, and `-- agent-options` to pass options on). It needs an
Application whose id is `codex` or `claude`, which `aiport init --agent`
creates. The settings apply to that launch only, and credentials are never
displayed. This is separate from the Agent runs that [Connecting a coding
agent](#connecting-a-coding-agent) describes.

### Compression

The Compression page shows whether the opt-in request compression saves
anything and whether it hurts answers, per method combination. Compression
runs only for requests that ask for it with `X-Aiport-Compress`; see [Compress
requests](#compress-requests).

It shows the savings, a trend of compressed and uncompressed traffic
(feedback, error rate, size saved), and a verdict per method combination:
**Not enough data** (fewer than 20 requests on either side), **Possible harm**
(feedback more than 0.1 lower, or an error rate more than 5 points higher,
than uncompressed traffic) or **No harm seen**. Token savings are estimated
from bytes (divided by 4).

The **Evaluator** panel chooses a Connection and model for model-based checks;
**Test evaluator** runs two known cases and reports whether the model labels
them correctly. A Connection in use as the evaluator cannot be removed until
another (or none) is chosen, except through the removal dialog. Details are in
[compression](compression.md) and [compression
evaluators](compression-evaluators.md).

### Feedback

The Feedback page shows how the people using each application rate its
answers, per module, model and effort, so you can see which part of which
application does well and move the part that does not. Applications send the
votes; see [Collect feedback and improve
routing](#collect-feedback-and-improve-routing) for how.

Choose the last 7, 30 or 90 days (30 by default). For each Application the
page shows its thumbs up and down, its share of up-votes, and how many of its
calls were rated. Its table lists each module, and inside each module each
combination of name sent (a Level, task name or model id), model and effort:
calls, votes, and the share of up-votes as a bar, green from 75%, amber from
50%, red below. A module or combination with fewer than 5 votes is marked "few
votes": too few to judge. Calls that named no module are grouped as "No module
named"; guidance version 4 asks the application's coding agent to add
`X-Aiport-Module`.

With two or more judged modules, the page names the best and the weakest
module; try another model or effort for the weakest under Routing, and watch
its share.

**Suggestions from the votes** lists what aiport found in the votes, each
marked **Better elsewhere**, **One Level up** or **Cheaper may do**, with the
move (from which model to which) and the reason, and an **Apply** button; see
[Suggestions from the votes](#suggestions-from-the-votes). Each Application's
panel lists its **Module routes**, each with **Remove**.

With request history off, the page says no calls or votes are kept. With calls
but no votes, it says applications start sending them once their coding agent
follows guidance version 4. The same report is `GET /api/v1/feedback?hours=N`
(1 to 2160, default 720), with `applications`, `suggestions` and
`module_routes`.

### Tasks

Tasks is a queue of work for coding agents: you capture tasks and own their
intent and priority, and coding agents pick them up by name. Each task gets a
memorable name, such as `task-amber-fox`, that you can hand to a coding agent;
click it to copy it.

A task has a title (up to 200 characters), instructions, acceptance criteria,
a priority (None, Low, Medium or High) and a status: pending, in progress,
blocked, completed, failed or cancelled. Its history keeps who changed what
and when. The filters above the queue show the open tasks (the default), all
of them, or one status, and the chime in the toolbar plays when no pending or
in-progress tasks remain.

The queue is stored next to the configuration in `aiport.yaml.tasks.json`.
From a terminal:

```sh
aiport task list [--all]
aiport task show NAME
aiport task set NAME --status completed --priority 2 --note "tests pass"
```

### Settings

Settings holds the parts of this aiport install that are not routing: sign-in,
the default coding agent for Agent runs, who writes Issue explanations, and
the configuration file.

- **Authentication**: the username and password you sign in with, and whether
  authentication is on. Use **Forgot username or password?** on the sign-in
  screen with the management token to set a new password here.
- **Default coding agent**: the agent of every Application that has none of
  its own, and **Model calls of Agent runs**, the Anthropic-protocol
  Connection (and optional model) their calls go through. See [Connecting a
  coding agent](#connecting-a-coding-agent) and [Explain by a Coding
  agent](#explain-by-a-coding-agent).
- **Explanations**: the Connection and model that explain an Issue when no
  coding agent does. See [Explain](#explain).
- **Configuration**: one file, `aiport.yaml`. Providers, routing, Levels and
  per-effort models and each Application's coding agent are managed from their
  pages; task names arrive as approved Changes or are edited, with the listen
  address, in the file. Validate hand edits with `aiport check`, then restart.

## Reference

Exact commands, endpoints, headers and codes, for when you know what you are
looking for.

### Command line

Every command takes `--config PATH` (default `aiport.yaml`).

| Command | What it does |
| --- | --- |
| `aiport init` | First run: choose providers, coding agents and Levels; writes `aiport.yaml` and `aiport.yaml.env` and never overwrites them. Flags: `--provider` (comma-separated: `openai`, `anthropic`, `bedrock`, `meta`, `ollama`, `llamacpp`), `--agent codex,claude` or `none`, `--listen HOST:PORT` (default `127.0.0.1:8787`), `--region`, `--base-url`, `--levels LEVEL=CONNECTION/MODEL,...` or `--no-levels`. |
| `aiport check` | Validates the configuration, names Applications without a token, and prints one notice per Alias named like a Level. |
| `aiport serve` | Starts the gateway; prints the dashboard address and, in a terminal, the management token. |
| `aiport auth token` | Prints the management token again. |
| `aiport agent codex` or `aiport agent claude` | Launches that coding agent through aiport with its own token. Needs `--model MODEL`; `--dry-run` shows what would start; options after `--` go to the agent. |
| `aiport task list`, `show NAME`, `set NAME` | The task queue. `list --all` includes closed tasks; `set` takes `--status`, `--priority 0-3`, `--note` and `--actor`. |
| `aiport token` | Prints a new random token, for example for a token variable you set yourself. |
| `aiport version` | Prints the version. |

### Endpoints applications use

All of these are on the gateway's address (`http://127.0.0.1:8787` by
default). Except for pairing and `/_connect`, they take the Application's own
token as `Authorization: Bearer <token>` (Anthropic SDKs may use `x-api-key`).

| Endpoint | What it is |
| --- | --- |
| `POST /v1/chat/completions`, `POST /v1/responses` | OpenAI-protocol model calls, forwarded natively. |
| `POST /v1/messages` | Anthropic-protocol model calls, forwarded natively. |
| `GET /v1/models` | The Application's Catalog in the SDK's own list shape. |
| `GET /_connect` | The protocol description as JSON; `?format=markdown` for the guide, `?format=standard` for the communication standard, `?format=handover&for=application` or `&for=agent` for a brief. No token needed. |
| `POST /v1/aiport/pair` | Asks for pairing with `{name, description?, repository?}`; returns a `code` and a `poll` path. No token needed. |
| `GET /v1/aiport/capabilities` | What the Application can use and ask for right now. |
| `GET /v1/aiport/catalog` | The Catalog and its recent changes. |
| `POST /v1/aiport/feedback` | A vote on an answer: `{request_id, vote}` or `{request_id, score}`, optional `module`. |
| `GET /v1/aiport/guidance`, `POST /v1/aiport/guidance/done` | The guidance changes since the confirmed version, and confirming the current one. |
| `POST /v1/aiport/requests`, `GET /v1/aiport/requests` (or `/{id}`) | File a Change request, and read the answers. |
| `GET /v1/aiport/issues` (or `/{id}`), `POST /v1/aiport/issues/{id}/fixed` | The Application's Issues, and saying one is fixed. |
| `/mcp` | The MCP endpoint for the Application's coding agent (HTTP transport). |

### Owner API

The dashboard uses the management API under `/api/v1/`, and scripts can too,
with the management token as `Authorization: Bearer <token>`. Every request
that changes something also sends `X-Aiport-CSRF: 1`. Application tokens never
open these routes. Each edit below that changes the configuration is one
revertable Change.

| Endpoint | What it does |
| --- | --- |
| `POST /api/v1/origins` | Adds an Application with a token: `{name, target?}`. |
| `POST /api/v1/origins/{id}/token` | Issues a new token for an Application; the old one stops working. |
| `DELETE /api/v1/origins/{id}` | Removes an Application. |
| `GET /api/v1/pairings?status=rejected` | Remembered rejected pairing requests. |
| `DELETE /api/v1/targets/{id}` | Removes a Connection; with `{"plan": {...}}`, re-points what used it. |
| `GET /api/v1/targets/{id}/dependents` | What uses a Connection, with the suggested replacements. |
| `GET /api/v1/routing` | The Routing map; `?application=<id>` for one Application, with its `changes`. |
| `GET /api/v1/levels` | All Level mappings, shared and per Application, and the Connections. |
| `PUT /api/v1/levels/{protocol}/{level}` | Maps a Level: `{application?, connection, model, effort?}`. |
| `DELETE /api/v1/levels/{protocol}/{level}` | Clears a mapping: `{application?}`. |
| `PUT /api/v1/aliases/effort` | Per-effort models of any Alias: `{application?, alias, connection?, effort}`. |
| `PUT /api/v1/task-names/{name}` | Sets a shared task name: `{connection, model, effort?, everywhere?, reasoning?}`. |
| `DELETE /api/v1/task-names/{name}` | Removes a shared task name: `{connection}`. |
| `PUT /api/v1/module-routes/{application}/{module}/{name}` | Sets a module route: `{connection, model, why?}`. |
| `DELETE /api/v1/module-routes/{application}/{module}/{name}` | Removes a module route. |
| `GET /api/v1/feedback?hours=N` | The votes report, Suggestions from the votes and module routes. |
| `GET /api/v1/guidance` | Each Application's guidance state, with `behind` and `never_confirmed`. |
| `GET /api/v1/issues` | Issues, with `fix_by` per Application. |
| `POST /api/v1/issues/{id}` | Acts on an Issue, for example `{"action": "reject"}`. |
| `POST /api/v1/issues/{id}/explain` | Streams an explanation as JSON lines. |
| `PUT /api/v1/adapting/explainer` | Chooses who writes explanations: `{connection, model}`. |
| `POST /api/v1/issues/{id}/fix`, `GET /api/v1/issues/{id}/fix` | Starts Fix it; reads its `branch_fix` and `diff`. |
| `POST /api/v1/issues/{id}/fix/keep`, `.../fix/discard` | Keeps or discards a Fix it diff. |
| `GET /api/v1/issues/{id}/run` | The Issue's last Agent run, its log and waiting approvals. |
| `GET /api/v1/runs`, `POST /api/v1/runs/{id}/stop` | The Agent runs going, and stopping one. |
| `GET /api/v1/approvals`, `POST /api/v1/approvals/{id}` | Commands waiting for you, and answering one. |
| `GET /api/v1/agents`, `PUT /api/v1/agents/calls` | The coding agents found, and where Agent runs' model calls go. |

### Headers

aiport removes every `X-Aiport-*` request header before forwarding.

| Header | Direction | What it means |
| --- | --- | --- |
| `Authorization: Bearer <token>` or `x-api-key` | Request | The Application's own token. aiport replaces it with the provider key upstream. |
| `X-Aiport-Module` | Request | The module of the application the call serves; required on every model call. Lower case `a`–`z`, `0`–`9`, `.`, `_`, `-`, at most 64. |
| `X-Aiport-Effort` | Request | `low`, `medium` or `high`: how hard the call is. Picks a per-effort model, and is sent to the provider as its reasoning effort when the body sets none. |
| `X-Aiport-Compress` | Request | `whitespace`, `json`, `dedupe` or `all` (comma-separated), or `off`. |
| `X-Aiport-CSRF: 1` | Request | Required on management API requests that change something. |
| `X-Aiport-Request-Id` | Response | aiport's id for the call; keep it with the answer for votes and logs. |
| `X-Aiport-Catalog` | Response | The version of the Application's Catalog (12 hex characters). |
| `X-Aiport-Compression` | Response | The methods, `status` (`applied`, `no_gain` or `failed`) and the bytes before and after. |
| `X-Aiport-Hint` | Response | One line: the open Issue a failing call belongs to, the Change request that could fix it, or what to do about a refused token. |

`X-Aiport-Task`, `X-Aiport-Session` and `X-Aiport-Version` (requests) and
`X-Aiport-Model`, `X-Aiport-Route` and `X-Aiport-Source` (responses) are
published in `/_connect` as planned, not built.

### Error codes

An aiport error is `{"error": {"code", "message", "request_id", ...}}`; any
other body is the provider's own, passed through unchanged with its status and
headers. Status and code never depend on the optional `hint`.

**The `fix` block.** Every refusal of an Application's token, `401
unauthorized` on a model or Application endpoint and `403
application_token_required`, carries `error.fix`, written for the
application's coding agent, and an `X-Aiport-Hint` header that says the same
in one line:

```json
{"error": {"code": "unauthorized", "message": "...", "request_id": "...",
  "fix": {"for": "the application's coding agent",
          "steps": ["Check that .aiport exists at the repository root ...", "If .aiport is missing ..., ask the Owner ... New token ..."],
          "retry": "Once .aiport has the token, retry the same call ... Claude Code reads the MCP token when it connects: run /mcp and reconnect aiport, or start a new session.",
          "docs": "http://127.0.0.1:8787/_connect?format=markdown#if-aiport-refuses-the-token"}}}
```

The steps say: check `.aiport` has the `AIPORT_APP_TOKEN=` line (an
`AIPORT_APP_TOKEN` in the environment wins over the file, so unset a stale
one); if it is missing, lost or revoked, ask the Owner for **New token**
(**Issue token** when there is none) on Integrate and paste the copied lines
into `.aiport`; never pair again unless the Owner asks, and never look for a
token anywhere else. Then retry the same call, restarting the application if
it reads `.aiport` only at startup.

| Status and code | Meaning |
| --- | --- |
| `401 unauthorized` | Missing or invalid Application token; follow `error.fix`. |
| `403 application_token_required` | The endpoint serves one Application and no Application token was recognized (the management token is not one); follow `error.fix`. |
| `403 invalid_host`, `403 invalid_origin` | The `Host` is not the listen address or a configured hostname; browser cross-site requests are refused. |
| `404 unsupported_path`, `405 method_not_allowed` | Not part of this protocol, or the wrong method. |
| `400 incompatible_protocol` | The chosen Connection speaks the other protocol; aiport does not translate. |
| `400 level_unmapped` | The Level is not mapped on this protocol, for this Application or shared; the message names both. |
| `400 ambiguous_model` | The body must be one JSON object with one `model` field when a name is rewritten. |
| `400 unknown_compression` | `X-Aiport-Compress` names a method aiport does not offer. |
| `413 request_too_large` | The body exceeds the admission limit (`proxy.max_request_bytes`). |
| `429 concurrency_limit` | Too many requests at once; retry with backoff (`proxy.max_concurrent`). |
| `503 target_unconfigured` | No Connection for this protocol or Application. |
| `503 target_disabled` | The chosen Connection is disabled. |
| `503 upstream_key_missing` | The Owner has not set the provider key yet. |
| `502 upstream_unavailable` | Could not connect to the provider. |
| `504 upstream_timeout` | The provider did not answer in time. |
| `400 invalid_pairing`, `404 pairing_not_found`, `429 pairing_busy` | Pairing needs a name (1 to 128 characters); the request expired or never existed; too many waiting, retry after a minute. |
| `400 invalid_feedback`, `404 request_not_found` | A malformed vote; no retained request with that id for this Application. |
| `403 change_requests_not_allowed` | The Owner has not allowed this Application to request changes. |
| `400 invalid_change_request` | Incomplete or too large; `error.problems` lists every problem. |
| `403 change_request_refused` | A policy Extension refused it; `error.reason` says why. |
| `429 change_requests_rate_limited` | More than 20 requests in the last hour; retry after `Retry-After`. |
| `404 change_request_not_found`, `404 issue_not_found`, `409 issue_not_open` | Not this Application's request or Issue, or the Issue is not open. |
| `409 adapting_off` | The adapting loop is turned off in the configuration. |

The full list is at `http://127.0.0.1:8787/_connect?format=markdown#errors`.

### Configuration

`aiport.yaml` is the configuration; [configuration](configuration.md)
documents every key, and [aiport.example.yaml](../aiport.example.yaml) is a
starting point. The keys this manual mentions:

| Key | What it sets |
| --- | --- |
| `listen`, `hostnames` | The loopback address, and extra `*.localhost` names for the browser. |
| `targets` | The Connections. |
| `defaults` | The protocol defaults. |
| `aliases` | Shared Level mappings and task names (`effort`, `everywhere`, `reasoning`). |
| `origins` | Applications: `token_env`, `target`, `aliases`, `may_request_changes`, `agent`, `repository`, `module_routes`. |
| `telemetry.enabled`, `telemetry.retention_days` | Request history and how long it, decided items and Catalog changes are kept (default 7 days). |
| `adapting` | `disabled`, `auto_fix`, `explain`, the thresholds `unknown_model`, `failing_connection`, `level_unmapped` and `unknown_token`, and `interval_seconds`. |
| `agent_runs` | `default_agent`, `connection`, `model`, `explain_minutes`, `fix_minutes`. |
| `evaluation` | The compression evaluator's Connection and model. |
| `extensions` | Policy and guardrail programs. |
| `upstream.repository` | Where Suggest to aiport and Share upstream open drafts. |
| `proxy` | Limits: `max_concurrent`, `max_request_bytes`, `response_header_timeout_ms`, `stream_idle_timeout_ms`. |

Changes made through the dashboard are written into the file in place, other
keys and comments kept, and apply at once. Hand edits need `aiport check` and
a restart.

### Data and privacy

aiport keeps request metadata, never prompts or responses, and everything
stays on this machine. These files live next to the configuration:

| File | What it holds |
| --- | --- |
| `aiport.yaml` | The configuration. Never holds a key or token. |
| `aiport.yaml.env` | Private credentials (mode 0600): provider keys you paste, Application tokens, the management token. |
| `aiport.yaml.db` | Private SQLite database (mode 0600): request history, Changes, decisions, Issues, Catalog changes, guidance confirmations, rejected pairings. |
| `aiport.yaml.tasks.json` | The task queue. |
| `aiport.yaml.worktrees/` | Worktrees of Agent runs while they run. |

**Stored per call:** time, route, Application, the name sent, effort and its
source, module, Connection, model, status, latency, short error codes,
compression sizes and the vote. **Never stored:** prompts, responses, request
or response headers, provider error messages, keys and tokens (tokens only in
the credential file). Issues, Catalog changes and explanations hold only ids,
model names, counts, request ids and fixed advice text.

**Retention:** request history, decided Suggestions and Change requests,
Catalog changes, and resolved or dismissed Issues are kept for
`telemetry.retention_days` (default 7); open Issues until they resolve;
Applied Changes and Issues that hold a fix you reverted or rejected are kept
beyond that. Deletion is not secure erasure. Model lists are kept in memory
only, for two minutes; Agent run logs in memory only, while their Issue is
open.

### Security

aiport is built to be safe to run on a developer machine. It is experimental:
read [SECURITY.md](../SECURITY.md) to report a vulnerability privately.

- **Loopback only.** aiport listens on `127.0.0.1:8787` by default, accepts
  only its listen address and configured `*.localhost` names as `Host`, and
  refuses cross-site browser requests.
- **One token per Application.** Applications never see provider keys; each
  sends its own token, and aiport adds the key upstream. Calls go only to the
  Connections you configured.
- **Dashboard sign-in.** The management token, then a password of at least 12
  characters. A browser session lasts 12 hours or until aiport restarts.
  Management requests that change something need `X-Aiport-CSRF: 1`.
- **Nothing changes without you.** Every Change is listed and revertable;
  automatic fixes are off by default and limited to one Application's Alias.
- **Agent runs** start only on your click, in a worktree, with a minimal
  environment and no aiport tokens or provider keys; commands wait for your
  answer, and aiport never commits to your branch, merges or pushes. Read
  [What aiport cannot see](#explain-by-a-coding-agent) before using them.
- **Extensions** run only when listed with an absolute path, with a timeout,
  and deny unless they explicitly allow.

## Troubleshooting

Find the symptom, then the cause and the fix. For any failed request,
Activity's **Details** names the reason and links to the page that fixes it.

| Symptom | Cause | Fix |
| --- | --- | --- |
| The dashboard does not open | aiport is not running, or the address differs | Check `aiport serve` is running and use the exact address it prints. aiport refuses other host names and other sites; to use a name, see [Use a name instead of 127.0.0.1](#use-a-name-instead-of-127001). |
| Forgot the password | — | Use **Forgot username or password?** on the sign-in screen, paste the token from `aiport auth token`, then set a new password in Settings → Authentication. |
| An application gets `401 unauthorized` or `403 application_token_required` | Its token is missing, lost, revoked or not loaded | Follow `error.fix`: check `.aiport`, else click **New token** on its row on Integrate and paste the lines into `.aiport`; restart the application if it reads `.aiport` only at startup. Do not pair again. |
| A `401` never shows in Activity | It has no Application to attribute it to | As above; pair the application on Integrate if it never had a token. |
| The aiport MCP server does not connect in Claude Code | The MCP entry read an old or missing token when it connected | Fix `.aiport`, then run `/mcp` and reconnect aiport, or start a new session. A `403 application_token_required` means no Application token was recognized. |
| An SDK, script or coding agent cannot reach `aiport.localhost` | Its resolver does not know the name, it resolves to `::1`, or a sandbox proxy refuses it | Use `http://127.0.0.1:8787`, which aiport always hands out; or add `127.0.0.1  aiport.localhost` to `/etc/hosts` for a script. |
| Calls show as `unknown` in Activity | The application is not sending its own token (for example a shared or old one) | Pair it, or give it a token with **Add with a token**. |
| `400 level_unmapped` | The Level is not mapped on that protocol | Routing → Levels: map it, or give the Application its own mapping on its Integrate row. When it keeps happening aiport opens an Issue. |
| `503 target_unconfigured` | No Connection for that protocol | Routing → Protocol defaults: pick one, or bind the Application. aiport does not translate between protocols. |
| The Feedback page shows no votes | Request history is off, the application's coding agent has not followed guidance version 4, or nobody voted yet | Turn on `telemetry.enabled`; check the **Modules** and **Votes** chips on Integrate and hand the coding agent the communication standard. |
| A vote answers `404 request_not_found` | The request id is not this Application's, or the call is no longer retained, or history is off | Send the `X-Aiport-Request-Id` of that answer, soon after the call. |
| No Suggestions from the votes | Fewer than 10 votes on one module, name, model and effort, or the votes are between 60% and 90% up | Wait for more votes; Suggestions need clear evidence. |
| `403 change_requests_not_allowed` | You have not allowed the Application to request changes | Allow it on Integrate if you want its requests. |
| `409 adapting_off` | The adapting loop is turned off in the configuration | Remove `adapting.disabled: true` and restart, if you want it back. |
| Explain is disabled | No Connection chosen for Explanations, or adapting is off | Settings → Explanations: choose a Connection and model. |
| Revert refuses with a conflict | The configuration changed after that Change | Revert the newer Changes first. |
| Activity is empty | Request history is off or not persisting | Check `telemetry.enabled: true` and that the banner does not say history is kept for this session only. |

### Activity reasons and their fixes

Activity's **Details** names the reason for every failed request. The reasons,
and where each is fixed:

| Activity says | Code | What to do |
| --- | --- | --- |
| Refused: no Connection for the Anthropic (or OpenAI) protocol | `503 target_unconfigured` | Routing → Protocol defaults: pick a Connection for that protocol, or bind the Application to one on Integrate. aiport does not translate between protocols. |
| Refused: the Connection is disabled | `503 target_disabled` | Enable it on Providers, or pick another default on Routing. |
| Refused: the Connection's settings are invalid | `503 target_invalid` | Correct its base URL on Providers. |
| Refused: the Connection does not speak the … protocol | `400 incompatible_protocol` | Routing → Protocol defaults: pick a Connection of that protocol. |
| Refused: the Level is not mapped for the … protocol | `400 level_unmapped` | Routing → Levels: map the Level on that protocol, or give the Application its own mapping on its Integrate row. When it keeps happening aiport opens an Issue (see [An unmapped Level](#an-unmapped-level)). |
| Refused: no API key for the Connection | `503 upstream_key_missing` | Paste the key on the Connection's card on Providers. |
| Refused: too many requests at once | `429 concurrency_limit` | Retry; raise `proxy.max_concurrent` if it persists. |
| Refused: the request is too large | `413 request_too_large` | Send less, or raise `proxy.max_request_bytes`. |
| Refused: the request body could not be read | `400 request_read_failed` | The Application stopped sending; retry. |
| Refused: unknown compression method | `400 unknown_compression` | Use a method from the Compression page. |
| Refused: the model name is ambiguous | `400 ambiguous_model` | Send one JSON object with one `model` field. |
| Cannot reach the provider: its host name does not resolve | `502 upstream_unavailable`, `dns_failure` | Check the base URL on Providers and the network. |
| Cannot reach the provider: connection refused | `502 upstream_unavailable`, `connection_refused` | Start the local server (Ollama, llama.cpp) or fix the port; **Check connection**. |
| Cannot reach the provider: the secure connection (TLS) failed | `502 upstream_unavailable`, `tls_error` | Check the scheme and host name of the base URL. |
| Cannot reach the provider: connecting timed out, or the connection was closed | `502 upstream_unavailable`, `timeout` or `connection_reset` | Check the network and the base URL; retry. |
| Timed out: the provider did not answer in time | `504 upstream_timeout` | Retry; for slow local models raise `proxy.response_header_timeout_ms`. |
| Provider error: the key was rejected | provider `401` or `403` (`authentication_error`, `invalid_api_key`, `permission_error`) | Replace the key on Providers, then **Check connection**. |
| Provider error: unknown model | provider `404` (`not_found_error`, `model_not_found`) | Use a model the provider lists (the Application sees them with `GET /v1/models`), or add an Alias. When it keeps happening aiport opens an Issue in Changes that says which Connection serves the model. |
| Provider error: rate-limited | provider `429` (`rate_limit_error`) | Wait, or ask the provider for a higher limit. |
| Provider error: out of quota or credit | `insufficient_quota`, `billing_error` | Check the plan and billing with the provider. |
| Provider error: the request was not accepted | provider `400` (`invalid_request_error`) | Check the model name and parameters the Application sends. |
| Provider error: the provider failed | provider `5xx` (`overloaded_error`, `api_error`) | Retry later; check the provider's status page. |
| The response broke off | `incomplete` | Retry; raise `proxy.stream_idle_timeout_ms` for long answers. |
| Cancelled by the Application | `499 client_cancelled` | Nothing to fix. |

## What's new

The most recent changes first. The commit history has every detail.

**2026-11-04**

- **Votes and Suggestions per module.** Applications name the module of each
  call with `X-Aiport-Module` and send a thumbs up or down per answer to
  `/v1/aiport/feedback`. The new [Feedback](#feedback) page shows votes per
  Application, module, model and effort, and turns them into Suggestions you
  apply as module routes.
- **The communication standard** at `/_connect?format=standard`, also the MCP
  tool `aiport_standard`, with every rule and a checklist; guidance version 4
  asks coding agents to follow it, and to build the thumbs into the
  application's UI.
- **Integrate shows Modules and Votes chips** per Application in the Guidance
  column.
- **Task names everywhere, with a fixed reasoning effort**, and an Owner API
  for them (`/api/v1/task-names`).
- **`X-Aiport-Effort` alone also sets the provider's reasoning effort** when
  the body sets none and aiport chose the model.
- **Token refusals tell the coding agent what to do** with an `error.fix`
  block and an `X-Aiport-Hint`.

**2026-11-03**

- **Levels** (`fast`, `balanced`, `best`), mapped per protocol and per
  Application, edited in the dashboard; `init` proposes a mapping; an Issue
  for unmapped Levels.
- **Effort picks the model**, with per-effort models for Levels and task
  names; Activity records the name, effort and model of each call.
- **The Catalog**: `GET /v1/models` per Application, `X-Aiport-Catalog`,
  `/v1/aiport/catalog`, Catalog-change Issues, and Connection removal with a
  re-pointing plan.
- **The Routing map** of every Application's names and what serves them.
- **Guidance versions**, so coding agents learn when to refresh; Integrate
  shows who is behind.
- **Tokens live in `.aiport`**; issue a token yourself, add an Application
  without pairing, or remove one. Pairing survives restarts and repeat
  requests.
- **Applications and agents are given `127.0.0.1`**, not the hostname.
- **The module path is `github.com/harelix/aiport`**, so `go install` works.

**2026-11-02**

- **Issues**, start to finish: unknown models and failing Connections, Copy
  fix for the Coding agent, automatic fixes if you allow them, and the
  Application's coding agent hearing about its Issues.
- **Explain** an Issue with a Connection, or with a coding agent reading the
  code; **Fix it** on a branch with a diff and command approvals.
- **The Integrate page** with the onboarding wizard and both briefs; pairing;
  Change requests; Extensions; Suggest to aiport and Share upstream; the MCP
  endpoint.
