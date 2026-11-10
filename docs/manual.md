# aiport manual

aiport is a small gateway that runs on your own machine. Your apps and coding
agents send their LLM calls to aiport instead of straight to OpenAI, Anthropic
and the rest, and aiport sends each call on to the right provider.

That gives you three things in one place:

- **Your provider keys.** They live in aiport, not in every app.
- **Your model choices.** You decide which model serves what, and change it
  without touching any app's code.
- **A record of every call.** You see what was called, by whom, and how well
  it went.

**Who this is for:** you, the Owner: the person who runs aiport on their
machine and decides what it does. This manual is written for you.

Other readers have their own documents:

- Coding agents that connect an app read the
  [communication standard](#the-communication-standard) and the briefs at
  `/_connect`.
- Exact configuration keys are in [configuration](configuration.md).
- The wire protocol is in [Connect](connect.md).
- Policies and guardrails are in [Extensions](extensions.md).

Capitalised words (Application, Connection, Level, Catalog, Issue, Change and
the rest) are explained in [Core concepts](#core-concepts) and the
[glossary](../GLOSSARY.md).

## At a glance

| Concept | In one line |
| --- | --- |
| Connection | A provider aiport sends calls to (OpenAI, Anthropic, Bedrock, Meta, Ollama, llama.cpp), with its key kept by aiport. |
| Application | Anything that calls aiport with its own token. Every call is recorded against it. |
| Protocol | `openai` or `anthropic`. aiport passes each call on in the provider's own format and never translates between them. |
| Level | `fast`, `balanced` or `best`: a model name an Application sends to say how capable a model it needs. |
| Task name | A model name for one kind of work (`summarize`, `classify`) that you point at a model. |
| Effort | `low`, `medium` or `high`, sent with `X-Aiport-Effort`: how hard one call is. |
| Module | The part of the Application a call serves, sent with `X-Aiport-Module`. Votes are read per module. |
| Catalog | What one Application can use through aiport, with a version on every answer. |
| Issue | A problem aiport keeps seeing for one Application, with evidence and a fix on each side. |
| Change | One configuration change made through aiport. Every Change can be reverted. |
| Suggestion | A change aiport proposes from request history or from votes. You decide. |
| Communication standard | The rules every integration follows, at `/_connect?format=standard`. |

## Quickstart

In about five minutes you will build aiport, connect one provider, and make
one real call through it that you can see in **Activity**.

You need Go 1.27 or newer. Nothing else is downloaded: the dependencies are
vendored and the dashboard is pre-built.

1. **Build aiport.**

   ```sh
   git clone https://github.com/harelix/aiport && cd aiport
   make build                 # -> bin/aiport, dashboard included
   ```

   Or install the binary without cloning:
   `go install github.com/harelix/aiport/cmd/aiport@latest`. It lands in
   `$(go env GOPATH)/bin`.

2. **Create the configuration.**

   ```sh
   ./bin/aiport init          # choose providers, coding agents and Levels
   ./bin/aiport check         # validate the configuration
   ```

   `init` writes `aiport.yaml`, plus a private file for credentials next to
   it, `aiport.yaml.env` (mode 0600). It never overwrites a file, and it never
   asks for provider keys in the terminal. It also proposes a first mapping
   for the Levels; see [Map Levels with init](#map-levels-with-init).

3. **Start aiport.**

   ```sh
   ./bin/aiport serve         # prints the dashboard address and the management token
   ```

   `serve` prints the management token when it runs in a terminal. To see the
   token again later, run `./bin/aiport auth token`.

4. **Sign in.** Open the address `serve` printed (`http://127.0.0.1:8787/` by
   default). Paste the management token. Then set a password of at least 12
   characters and, if you like, a username.

5. **Add a provider key.** On **Providers**, paste the provider's API key into
   its card and click **Check connection**. The check makes one free
   `GET /models` call and generates nothing. A local provider (Ollama,
   llama.cpp) needs no key, but you start it yourself.

6. **Map a Level.** A Level is a shared model name such as `fast`. On
   **Routing**, in the **Levels** panel, find `fast` on the OpenAI protocol.
   Click **Map** (or **Edit**, if `init` already mapped it) and choose a
   Connection and a model.

7. **Get an Application token.** On **Integrate**, click **Add with a token**,
   name it `quickstart`, and copy the two `.aiport` lines it shows. Save them
   as a file named `.aiport` in an empty directory, then load them into your
   shell from that directory:

   ```sh
   set -a; . ./.aiport; set +a
   ```

8. **Make the first call.**

   ```sh
   curl -s http://127.0.0.1:8787/v1/chat/completions \
     -H "Authorization: Bearer $AIPORT_APP_TOKEN" \
     -H "X-Aiport-Module: quickstart" \
     -H "X-Aiport-Effort: low" \
     -H 'content-type: application/json' \
     -d '{"model":"fast","messages":[{"role":"user","content":"Say hello"}]}'
   ```

   You get an answer from the model you mapped to `fast`. Open **Activity**:
   the call is listed with the name you sent (`fast`), its effort, and the
   Connection and model that served it.

**Next:** connect a real app by handing its coding agent the communication
standard (see [Connect an application](#connect-an-application)). Read
[Core concepts](#core-concepts) for the words used in the rest of this manual.

## Core concepts

These are the ideas the rest of the manual builds on. Each one says what the
thing is, then why it is there.

**Connection.** A provider aiport can send calls to: OpenAI, Anthropic, Amazon
Bedrock, Meta Model API, or a local server such as Ollama or llama.cpp. It
holds an address and a key. Think of it as an entry in your address book that
also keeps the key. Your provider keys live here, in one place, and never in
your apps.

**Application.** Anything that calls aiport with its own token: a script, a
service, a desktop app or a coding agent. It gets its token by pairing, which
you approve, much like letting a new device join your Wi-Fi. Because each app
has its own token, every call is recorded against someone, and aiport can tell
each app about its own problems.

**Protocol.** Each Connection speaks one protocol: `openai` or `anthropic`.
aiport passes Chat Completions, Responses and Messages calls on in the
provider's own format, streaming included. It never translates from one
protocol to the other. So a call from the OpenAI SDK needs a Connection that
speaks `openai`, and a call from the Anthropic SDK needs one that speaks
`anthropic`. Because nothing is translated, nothing is lost in translation.

**Level.** One of three model names every aiport shares: `fast`, `balanced`
and `best`. An app sends a Level as its model name to say how capable a model
it needs, without naming a provider. It is like ordering "small, medium or
large" and letting the kitchen pick the cup. You map each Level, per
protocol, to a Connection and a model, and you can give one app its own
mapping. You can then change models without anyone changing code.

**Task name.** A model name for one kind of work, such as `summarize`,
`review-code` or `classify`, that you point at a Connection and a model.
Under the hood it is an Alias: a name aiport swaps for a real model. A shared
task name marked `everywhere` can be sent by every app. A task name with a
fixed `reasoning` effort always sends that effort to the provider. Task names
let routine work run on a cheaper model while the Levels stay on stronger
ones.

**Effort.** How hard one call is: `low`, `medium` or `high`, sent with the
`X-Aiport-Effort` header. A Level or task name can have a different model for
each effort. When the call's body sets no effort of its own, the effort from
the header also reaches the provider as its reasoning effort. Effort lets a call ask for
more or less thinking without picking another model name.

**Module.** The part of the app a call serves, such as `tldr`, `classify` or
`insights`, sent with the `X-Aiport-Module` header. aiport records it with the
call and never sends it on to the provider. Modules let you read the votes
people give answers one module at a time, and move one module to another
model without moving the others.

**Catalog.** The menu for one app: everything it can use through aiport. That
is its Levels, its task names and the models of the Connections it can reach,
with a short version. The app reads it at `GET /v1/models` and sees the
version on every answer in the `X-Aiport-Catalog` header. The app never has to
guess a model name, and it learns at once when you change something.

**Change.** One configuration change made through aiport: by you, by a request
you approved, or by an automatic fix you allowed. Every Change is listed on
the Changes page and can be reverted, like an undo history. That makes it safe
to experiment.

**Issue.** A problem aiport keeps seeing for one app, such as a model that no
Connection serves. It comes with the evidence and two fixes: one in aiport,
which you approve, and one in the app, for its coding agent. It resolves
itself once calls succeed again. Issues make sure a failing integration is
noticed and explained, instead of being retried forever.

**Suggestion.** A change aiport proposes; you always decide. There are two
kinds:

- Suggestions on the Changes page come from request history. For example, an
  unknown token that keeps trying to call.
- Suggestions from the votes come from the thumbs up and down people give
  answers. They appear on the Feedback page and are applied as module routes.

**Module route.** A route for one module of one app. Calls from that module
that send a given name go to another Connection and model, ahead of every
other route. The app's other modules keep their routes. A module that does
badly can move without moving the whole app.

**Change request.** An app asking for something aiport lacks, such as a task
name, with what it needs and why. Only apps you allow may send one. You decide
it on the Changes page.

**Guidance version.** The version number of the guidance aiport gives coding
agents: the briefs, the `AGENTS.md` line and the MCP instructions. It is now
version 4. aiport remembers the version each app's coding agent last
confirmed and tells it when it is behind, so integrations keep up as aiport
changes.

**Communication standard.** One document, at `/_connect?format=standard`, with
every rule an integration follows and a checklist: the token in `.aiport`, the
SDK pointed at aiport, Levels and task names, `X-Aiport-Effort`,
`X-Aiport-Module`, the thumbs up and down, and the fixes for errors. One paste
brings any app in line, whether it is new or already connected.

## Guides

Each guide walks you through one task from start to finish. For a page-by-page
description of the dashboard, see [A tour of the dashboard](#a-tour-of-the-dashboard).

### Connect an application

To connect an application, you give its coding agent one instruction and then
approve the code it shows you. The agent sends the app's LLM calls through
aiport and pairs the app. From then on the app has its own aiport token, sends
its calls to aiport, and never holds a provider key.

The quickest way:

1. On **Integrate**, open the **Onboard an application** wizard and click
   **Copy instruction**.
2. Open the coding agent in the application's repository and paste it.
3. When the agent shows you a pairing code, check that the same code appears on
   Integrate, then click **Approve**.

The rest of this page explains each part.

#### The communication standard

The communication standard is one document that lists every rule an
integration follows, with a checklist at the end. It is written for the coding
agent working in the application's repository, and it is the fastest way to
bring any application in line, whether it is new or already connected.

To hand it to a coding agent:

1. Go to **Integrate** and find the **Communication standard** card at the top
   of **Instructions**.
2. Click **Copy instruction** and paste the prompt to the agent.

**Copy URL only** copies just the address, and **Read the standard** opens the
document.

The standard has seven parts. All of them are required.

1. **Identity.** One token per application, kept in `.aiport` at the
   repository root and listed in `.gitignore`.
2. **Transport.** Point the OpenAI SDK at `AIPORT_BASE_URL + "/v1"` and the
   Anthropic SDK at `AIPORT_BASE_URL`. Both use `api_key = AIPORT_APP_TOKEN`.
   No provider keys stay in the application.
3. **Model names.** Send a Level or a task name, never a provider's model id.
   `GET /v1/models` lists the names.
4. **Headers.** Send `X-Aiport-Module` on every model call (required). Send
   `X-Aiport-Effort` where calls differ in difficulty. Keep
   `X-Aiport-Request-Id` with every answer, together with `X-Aiport-Model` and
   `X-Aiport-Route`, which name the model that answered and how it was chosen.
   Read `/v1/models` again when `X-Aiport-Catalog` changes.
5. **Rating.** Put a thumbs up and a thumbs down next to every LLM answer the
   application shows a person. The coding agent builds them into the app's own
   UI, and they are sent to `POST /v1/aiport/feedback`.
6. **Errors.** On `401` and `403`, follow `error.fix`. Retry with backoff only
   on `429 concurrency_limit`, `502` and `504`.
7. **Talking to aiport.** Use the MCP tools at the start of each session, file
   Change requests instead of building workarounds, and add aiport's line to
   `AGENTS.md`.

**Good to know**

- An application without a user interface (a service, a batch job or a
  command-line tool) offers the vote wherever a person does see the answer. If
  no person ever sees it, the coding agent tells you so and skips the votes,
  but it still sends `X-Aiport-Module`.

**For scripts:** the document is served at `/_connect?format=standard`. The
same document is the MCP tool `aiport_standard`. The `/_connect` JSON gives it
as `standard`, and the prompt as `standard_prompt`.

#### Onboarding an application, step by step

The **Onboard an application** wizard on Integrate walks you through connecting
one application. Each step ticks itself off as aiport sees it happen, and the
step to do now is highlighted.

The wizard is open while you have no Applications and whenever a pairing
request is waiting. Otherwise it is collapsed, and its header tells you where
things stand, such as "1 pairing waiting" or the Application it is following
and how far it got. To follow an app, select an existing Application in the
list, or click **Onboard a new application** for a new one.

1. **Hand this to the coding agent.** Open the coding agent (Claude Code,
   Codex, Cursor and so on) in the application's repository. Click **Copy
   instruction** and paste it (**Copy URL only** copies just the address). See
   [What the instruction asks](#what-the-instruction-asks) below.
2. **Approve the pairing.** The agent asks aiport to pair. Within seconds the
   request appears at the top of Applications with a short code. Check that it
   matches the code the agent shows you, then click **Approve**. You can also
   bind a Connection (a provider aiport can call, with its key) and allow
   change requests while you approve. Requests wait 30 minutes. Once you
   approve, the wizard follows that Application.
3. **First call arrives.** The agent makes one real call. The step ticks when
   Activity has a call from the Application in the last 7 days.
4. **Allow change requests** (optional). Use the toggle in the step. You still
   decide each request yourself in Changes.
5. **MCP entry in the repository** (optional). The agent already adds it in
   step 1. Tick the step by hand, or copy the entry from the **Coding agent
   brief** tile.
6. **Connect a coding agent** (optional). See [Connecting a coding
   agent](#connecting-a-coding-agent). The step ticks when the Application has
   a coding agent (its own or the default) and a repository path.

##### What the instruction asks

The instruction tells the agent to read
`http://127.0.0.1:8787/_connect?format=markdown` and to do two parts, in
order:

1. Wire the LLM calls through aiport: pair, save the token in `.aiport` and
   add that file to `.gitignore`, read `.aiport` at startup, point the SDK at
   aiport, remove provider keys, send `X-Aiport-Module`, add the thumbs up and
   down, and make one test call.
2. Add the aiport MCP entry and the `AGENTS.md` line.

It also points the agent at the communication standard for the checklist. It
asks the agent to tell you the pairing code and wait, never to commit
`.aiport`, and to ask you for a **New token** instead of pairing again if the
token is lost.

The markdown document opens with a **Start here** section and holds everything
for both parts, so an agent given only that address needs nothing else.

**For scripts:** the exact instruction text, with your gateway's address, is at
the top of the markdown document and in the `/_connect` JSON as
`agent_prompt`.

#### Applications and pairing

An Application is anything that calls aiport with its own token, and pairing is
how it gets that token with your approval. aiport credits each call to the
Application that sent it, so Activity shows who called what. Calls without an
Application token show as `unknown`.

How pairing works:

1. The application asks aiport to pair, giving its name, and shows you a short
   code. It may also report the path of its git repository on this machine
   (the Application brief asks the integrating coding agent to do this).
2. The request appears at the top of the Applications list on **Integrate**,
   and in the Applications table on **Routing**, with that code. Check that the
   application shows the same code.
3. If you want, choose a Connection to bind it to and tick **May request
   changes**.
4. Click **Approve**, or **Reject**.
5. The application collects its token, once. aiport adds the Application to
   the configuration without a restart and stores the token in the private
   credential file as `AIPORT_APP_<ID>_TOKEN`.

##### The repository path

If the application reported a repository path, it appears in a field under the
request. If it is not a git repository on this machine, the field says what is
wrong with it. Confirm, edit or clear the path before you approve.

aiport keeps only what the field holds when you approve. A reported path you
did not confirm is never used. A path that does not exist, or is not the top
directory of a git repository, is refused with the reason, and the request
keeps waiting.

##### Where the token lives: `.aiport`

Each application keeps its aiport settings in a file named `.aiport` at the
root of its repository. The file is listed in `.gitignore` and never
committed:

```
AIPORT_BASE_URL=http://127.0.0.1:8787
AIPORT_APP_TOKEN=<its token>
```

The application loads these into its environment at startup. Variables that
are already set win.

Its coding agents read the file too. The project MCP entry's `headersHelper`
takes the token from `AIPORT_APP_TOKEN`, or else from `.aiport`, so nothing
has to be exported before the agent starts.

If a token is lost or revoked, aiport answers `401`. To fix it:

1. On Integrate, click **New token** on the Application's row.
2. Click **Copy .aiport lines** and paste them over the app's `.aiport`.

The old token stops working at once. Guidance version 3 told the coding agents
of Applications integrated earlier to move their token into `.aiport`.

##### One request per application

If an application asks again while its request is still waiting, it gets the
same request back, with the same code, so the code you see stays valid. aiport
treats it as the same application when it has the same repository, or the same
name ignoring case and punctuation. Both its earlier and its newer poll paths
keep working.

If a request matches an Application you already have, it is marked **Already
connected as …**. Its main button, **Give … a new token**, approves it as that
existing Application: it receives a new token and its old one stops working.
This avoids a duplicate such as `docs-bot-2`. **Add as separate** still adds a
new Application.

##### Waiting and restarts

Requests wait 30 minutes for you and survive an aiport restart, because the
poll secret is stored only as a hash. After a restart you can still approve a
waiting request by its code, and an approved application can still collect its
token for 24 hours. When a request has run out, the application's poll answers
`expired`, so it knows to ask again.

##### Rejected requests

aiport remembers a request you **Reject**. The application learns it was
rejected. aiport keeps its name, description, code and the time (never a
token) for the newest 100 rejections, across restarts. To see them, use the
**Rejected** filter on Integrate. **Forget** removes one. A rejected
application may ask to pair again.

##### The Applications table

The **Applications** table on Routing lists every Application with its client
id, its Connection, its own Aliases (model names that only it uses), and the
**May request changes** toggle.

**For scripts:**

- The application starts pairing with `POST /v1/aiport/pair` and its name. No
  token is needed. The route is public on loopback and rate-limited.
- A poll answers `pending`, `rejected`, `expired`, or `approved` with the
  token.
- To approve a request as an existing Application over the API, send
  `{"replace": "<id>"}`.

#### Give a token yourself, or remove an Application

You can skip pairing and give an application its token yourself. This helps
when its coding agent is not running, or when you would rather paste the token
into its `.aiport` file.

- **Add with a token** (on Integrate, next to **Onboard a new application**).
  Name the Application and, if you want, choose a Connection (a provider aiport
  can call, with its key). aiport adds it and shows its token once, on a token
  card with **Copy** and **Copy .aiport lines**. The lines are
  `AIPORT_BASE_URL` (this gateway's numeric address) and `AIPORT_APP_TOKEN`.
- **New token** (the key button at the end of an Application's row, or **Issue
  new token** after you click its name). After you confirm, aiport issues a new
  token and shows it once, the same way. The old token stops working at once,
  whether it came from pairing or from here. Issuing a token changes no
  configuration, so it does not appear in Changes.
- **Remove** (the trash button at the end of the row, or **Remove
  application** after you click its name). aiport first lists what goes with
  the Application: its own Aliases, its binding, its coding agent settings and
  its permission to request changes. It then removes all of it as one Change
  and deletes the token, which stops working at once. Its activity history is
  kept.

**Good to know**

- If you revert a removal in **Changes**, the Application comes back **without
  a token**. Its row shows **No token**, and its calls are refused until you
  click **Issue token**. `serve` and `aiport check` also name such
  Applications.
- A token set in aiport's own environment when it started (exported in the
  shell instead of kept in the private credential file) cannot be replaced or
  deleted from the dashboard. Change it where aiport is started, and restart
  aiport.

**For scripts:** with the management token, use `POST /api/v1/origins` with
`{name, target?}`, `POST /api/v1/origins/{id}/token` and `DELETE
/api/v1/origins/{id}`. Send `X-Aiport-CSRF: 1` with each. A token appears in
the response once and is never shown again.

### Set up the application's coding agent

The coding agent that works in an application's repository talks to aiport
over MCP, using the Application's own token. That way, at the start of every
session it can read the Catalog (the model names the app may use), its Issues
(problems aiport found with its calls) and any new guidance. You set this up
once per repository, with an MCP entry and one line in `AGENTS.md`. If you
onboarded the app with the wizard, the integrating agent already added both in
step 1.

To add them yourself:

1. Open the **Coding agent brief** on Integrate. It has both pieces, ready to
   copy.
2. Add the MCP entry to the repository's `.mcp.json`.
3. Add the line to the repository's `AGENTS.md`.

#### The MCP entry and the AGENTS.md line

**The project MCP entry** (`.mcp.json`) points at this gateway's `/mcp`
endpoint over HTTP. Its `headersHelper` reads the token from
`AIPORT_APP_TOKEN`, or from `.aiport` at the repository root when that
variable is unset. The entry itself never holds the token.

**The `AGENTS.md` line** tells every later session to:

- find the token in `.aiport`, and what to do on `401`;
- send a Level or task name, never a provider's model id;
- state effort with `X-Aiport-Effort`;
- name the module of every call with `X-Aiport-Module`;
- add the thumbs up and down where the app has a UI;
- check `aiport_catalog`, `aiport_my_issues` and `aiport_guidance` at the
  start of a session, and call `aiport_guidance_done` after updating;
- ask with `aiport_request_change` instead of working around aiport.

The instructions aiport sends an agent when it connects over MCP (`initialize`)
say the same as the `AGENTS.md` line.

The MCP endpoint serves one Application at a time. A connection without a
recognized Application token is refused with `403
application_token_required`.

**Tip:** Claude Code reads the MCP token when it connects. After the token in
`.aiport` changes, run `/mcp` and reconnect aiport, or start a new session.

##### The MCP tools

Each tool answers exactly as its HTTP endpoint does, with the Application's own
token.

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

**For scripts:** the Coding agent brief is served at
`/_connect?format=handover&for=agent`.

#### Guidance versions

aiport gives coding agents guidance: the briefs, the `AGENTS.md` line and the
MCP instructions. It numbers that guidance and raises the number whenever it
changes in a way agents must act on. Each version has a short change log.
aiport remembers, per Application, the last version its coding agent
confirmed, so an agent that is behind learns exactly what to do.

| Version | What it asks agents to do |
| --- | --- |
| 1 | The project setup: the MCP entry, the `AGENTS.md` line, and reading `aiport_catalog` and `aiport_my_issues` at the start of a session. |
| 2 | Levels and effort: send a Level or task name instead of a provider's model id, add `X-Aiport-Effort` where calls differ in difficulty, replace the `AGENTS.md` line, and check `aiport_guidance` at the start of each session. |
| 3 | The token lives in `.aiport`: add it to `.gitignore`, move the token there, load it at startup, replace the MCP entry with the one whose `headersHelper` reads `.aiport`, replace the `AGENTS.md` line, and ask for a **New token** on `401` instead of pairing again. |
| 4 | Votes per module: read the communication standard, give each LLM feature a module name sent as `X-Aiport-Module`, keep `X-Aiport-Request-Id` with each answer, build a thumbs up and down into the UI, and replace the `AGENTS.md` line. |

Each version ends with the agent calling `aiport_guidance_done`. The confirmed
version is kept in the database next to the configuration. Applications
integrated before guidance versions existed have none.

##### How a coding agent that is behind finds out

You do not have to tell it. While an Application is behind:

- the MCP `initialize` instructions start with a one-line reminder;
- every aiport MCP tool answer ends with the same reminder, as a second text
  block. The exceptions are `aiport_guidance` and `aiport_guidance_done`,
  whose answers are the guidance itself;
- `aiport_capabilities` shows `guidance: {current, confirmed}`. `confirmed` is
  `null` if the agent never confirmed.

The agent then calls `aiport_guidance`, does the steps, and calls
`aiport_guidance_done`. Both use the Application's own token and only show or
change that Application's state. All of this works with adapting off.

##### Where you see it

- The **Guidance** column on Integrate shows each Application's state. An
  Application that is behind shows **Behind v… → v…**, or **Unknown** if it
  never confirmed. A copy button next to it copies one paragraph to paste to
  its coding agent.
- **Overview** shows one quiet line while any Application has not confirmed
  the current version. An Application that never confirmed any version counts
  as not confirmed. The line tells the two cases apart: "behind" means it
  confirmed an older version, and "never confirmed" means none yet.

**For scripts:** `GET /v1/aiport/guidance` (the `aiport_guidance` tool) returns
`{current, confirmed, confirmed_at_ms, behind, changes, agents_line, steps}`.
`POST /v1/aiport/guidance/done` confirms. On the management side, `GET
/api/v1/guidance` has `behind` and `never_confirmed`.

### Levels

Levels are three model names every aiport shares: `fast`, `balanced` and
`best`. You decide which model each one means on your machine, per protocol.
An Application sends a Level as its model to say how capable a model it needs
without naming a provider, so you can move it to another model without
changing its code.

To map a Level:

1. Go to **Routing → Levels**.
2. Click **Edit** on the Level and protocol (or **Map**, if it is not mapped
   yet).
3. Choose a Connection (a provider aiport can call, with its key) and a model.
   Click **Save**.

You can also map Levels with `aiport init` or in `aiport.yaml`. Details for
each are below.

#### How a Level mapping works

A Level is a reserved Alias name (an Alias is a model name that aiport replaces
with a real model on a Connection). There are two kinds of mapping:

- **The shared mapping** is an entry under `aliases:` whose `alias` is the
  Level, on a Connection of the protocol it serves. Each Level is mapped at
  most once per protocol.
- **An Application's own mapping** is the same kind of entry under its
  `origins[].aliases`. It comes before the shared one, on that protocol only.

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

A Level mapping chooses the Connection. That is the difference from a plain
shared Alias, which only renames a model on the Connection the request is
already going to. A Level mapping chooses the Connection even for an
Application bound to another one, or one with no protocol default.

Only the request's `model` field is replaced. In one case the reasoning effort
is added too; see [Effort and reasoning](#effort-and-reasoning).

Mapped Levels are listed in each Application's `GET /v1/models`, on the
protocol they are mapped on, and in its Catalog. Changing a mapping is a
Change you can revert, and a Catalog change the Applications are told about.
If you remove a Connection that a Level is mapped on, aiport asks you for a
replacement Connection of the same protocol.

#### When a Level is not mapped

aiport never quietly uses another model. A Level that is not mapped on the
request's protocol is refused with `400 level_unmapped`, naming the Level and
the protocol. If an Application keeps asking for an unmapped Level, aiport
opens an Issue for it (see [An unmapped Level](#an-unmapped-level)).

Some requests have nowhere at all to go. aiport refuses those at once, before
reading the body. That happens when, for the request's protocol, the
Application has none of these:

- a binding or a protocol default;
- a task name of its own there;
- a Level mapped for it;
- an `everywhere` task name;
- a module route.

Once any of those exists for the Application, or while its Connection is only
switched off, aiport reads the model first. So a Level is still refused as
`level_unmapped`, with its name.

#### Map Levels with init

`aiport init` suggests a first mapping per protocol, based on the providers
you chose:

- `fast` goes to a local Connection, if you chose one;
- `balanced` and `best` go to hosted ones;
- if you chose only one kind, it uses whatever you chose.

The presets do not know any model ids, so `init` asks you for each Level's
model. An empty answer leaves that Level unmapped. Nothing is mapped unless
you confirm.

Without a terminal, pass the mapping as a flag:

```
--levels fast=local-connection/llama3:8b,best=hosted-connection/some-model
```

Each item is `LEVEL=CONNECTION/MODEL`. The Connection is one of the provider
ids you chose, and the protocol is that Connection's. To skip Levels, pass
`--no-levels`. With neither flag and no input, no Level is mapped and `init`
says so.

#### Map Levels in the dashboard

You can edit Levels in the dashboard without touching the file. Every save is
one Change, so you can revert it in Changes to get the previous mapping back.

- **Routing → Levels** shows each Level on each protocol. **Edit** (or
  **Map**, when the Level is not mapped) opens its mapping: a Connection of
  that protocol, a model chosen from the models the Connection lists, and
  optional per-effort models. **Clear** removes the shared mapping. Each cell
  also says how many Applications have their own mapping.
- **Integrate → Applications → Levels** (the toggle under an Application's
  name, next to **Catalog**) shows the same grid for one Application: its own
  mappings, marked **Own**, and the shared ones it falls back to. **Override**
  gives it its own mapping, and **Use the shared mapping** removes it again.

To check the result, the [Routing map](#the-routing-map) shows what each Level
reaches for each Application. Activity shows which name and effort each call
asked for, and what served it.

#### Upgrading: Aliases named like a Level

If your configuration already had an Alias called `fast`, `balanced` or
`best`, it is now that Level's mapping on its Connection's protocol.

- A shared one now chooses the Connection for every Application's calls for
  that name on that protocol, whatever they are bound to. Before, it only
  renamed a model on the Connection a call was already going to.
- An Application's own one applies on its Connection's protocol only.

`aiport check` prints one notice per such entry, naming it and its Connection.
The startup log sums them up in one line.

- To keep it as the Level, do nothing.
- To keep it as a plain Alias, rename it (its `alias:`) to a name that is not
  a Level, and send the new name.

### Create task names for routine work

A task name is a model name for one kind of work, such as `classify`, `tldr`
or `extract`, that you point at a Connection (a provider aiport can call, with
its key) and a model. A typical use is a cheaper model for routine work at a
fixed effort, while the Levels stay on the models for harder work.

For example, to send all classification work, from every Application, to a
small local model with low reasoning effort, add this to `aiport.yaml`:

```yaml
aliases:
  - {target: local, alias: classify, model: 'llama3:8b', everywhere: true, reasoning: low}
```

Any Application can now send `classify` as its model.

#### Aliases and task names

An Alias is a model name an Application sends that aiport replaces with a real
model on a Connection. A task name is simply an Alias named for the work. Only
the `model` field of the request is rewritten.

- **Shared Aliases** apply to every Application. They live under `aliases:` in
  `aiport.yaml`, one per Connection and name. Without `everywhere`, a shared
  Alias applies only to calls that already go to its Connection.
- **An Application's own Aliases** come first, and may pick a different
  Connection from its usual one. They live under `origins[].aliases`.

For example, docs-bot has its own Alias `summarize` and the shared Alias
`review-code`, both on the local Connection it is bound to. That Connection
also serves `llama3:8b`. So docs-bot can send `summarize`, `review-code` or
`llama3:8b`.

You can write Aliases in the configuration file, or let them arrive as
Changes. Approving a Change request of kind `application_alias`, or an Issue's
fix, adds one. It is for that Application, or a shared Alias if you choose
**every Application** as the Scope.

**For scripts:** an Application lists its own Aliases with `GET
/v1/aiport/capabilities`, and every name it can use with `GET /v1/models` (see
[The Catalog](#the-catalog)).

#### Task names for every Application, with a fixed reasoning effort

Two settings on a shared task name make it work like a Level:

- **`everywhere: true`** lets every Application send it, whichever Connection
  its calls otherwise go to. The name routes to its own Connection.
- **`reasoning: low`, `medium` or `high`** makes aiport send that reasoning
  effort to the provider on every call for the name, whatever
  `X-Aiport-Effort` says. A call whose body already sets its own reasoning
  effort keeps it.

The example at the top of this page uses both. Each Application's Catalog
lists the name with its `reasoning`.

**Good to know**

- A task name is 1 to 256 characters, with no spaces or slashes.
- It cannot be a Level. Map those under [Levels](#levels).
- One name can route `everywhere` from only one Connection. To move it,
  remove it there first.

**For scripts:** add or change a task name with `PUT
/api/v1/task-names/{name}` and `{connection, model, effort?, everywhere?,
reasoning?}`. Remove it with `DELETE /api/v1/task-names/{name}` and
`{connection}`. Each is a Change you can revert.

#### Per-effort models of a task name

A task name can use a different model for each effort (low, medium, high), all
on its own Connection.

1. Find the task name's row on the [Routing map](#the-routing-map), or in an
   Application's Catalog on Integrate.
2. Click the pencil.
3. Fill in a model for low, medium and high. An empty field uses the task
   name's usual model. **Clear** removes them all.

A shared task name's per-effort models apply to every Application. Each save is
one Change you can revert.

**For scripts:** use `PUT /api/v1/aliases/effort` with `{application?, alias,
connection?, effort}`. A shared task name is found by its name and
`connection`, an Application's own by its name. In `aiport.yaml` the
per-effort models are the Alias's `effort:` field.

### The Catalog

The Catalog is the list of names an Application (an app or agent connected to aiport) can use through aiport. You decide what is in it, and aiport tells the app, so the app never has to guess a model name or learn from a provider's `404` that something changed.

An Application's Catalog holds:

- the Levels mapped for it;
- its task names: its own Aliases first, then the shared Aliases of the
  Connections it reaches, then the `everywhere` names;
- the models those Connections list;
- a short version code, which changes when the Catalog changes.

The briefs and the communication standard teach every app and its coding agent
how to use it:

- send a Level or a task name, never a provider's model id;
- read `GET /v1/models` to find the names;
- watch the `X-Aiport-Catalog` header;
- check `aiport_catalog`, `aiport_my_issues` and `aiport_guidance` at the
  start of each session.

The Catalog works with adapting turned off. It is information, not a Change.
Only the Catalog-change Issues follow the adapting switch.

#### What an Application reads

To see an app's Catalog yourself, open Integrate and click **Catalog** on the
app's row. It shows the same view the app gets, with its recent Catalog
changes.

An app reads its Catalog with its own token, and every answer is about that
app only. It never sees another app's Catalog. Keys, addresses, prompts and
tokens never appear.

**For scripts:**

- **`GET /v1/models`** is where OpenAI and Anthropic SDKs and coding agents
  already look for models.
  - It answers in the OpenAI list shape. When the request carries
    `anthropic-version`, it answers in the Anthropic shape (Anthropic SDKs may
    send the token in `X-Api-Key`).
  - Only names usable on that protocol are listed.
  - Mapped Levels and task names come first, marked as aiport's:
    `owned_by: "aiport"`, or "(aiport Level)" and "(aiport task name)" in the
    display name.
  - Then come the models of the Connection the app is bound to, or of the
    protocol defaults. A Connection whose model list cannot be read adds
    nothing. These are the same free, cached lists **Show models** uses.
  - Your management token gets `403 application_token_required`: a Catalog
    belongs to one Application.
- **`X-Aiport-Catalog: <version>`** (12 hex characters) is on every answer to
  a request made with an Application token, forwarded calls and errors
  included. The provider's answer is otherwise unchanged.
  - The version changes only when one of these changes: the app's task names
    or where they go, its binding, or the protocol defaults it follows.
  - A provider adding or retiring a model does not change it.
  - An app that sees a new version reads `/v1/models` again.
- **`GET /v1/aiport/catalog`** (or the MCP tool `aiport_catalog`, which gives
  the same answer) returns `catalog` and `changes`.
  - `catalog` holds the task names with their Connection and model. Levels are
    among them with `level: true` and their `protocol`. Per-effort models
    appear as `effort`, and a fixed `reasoning` where one is set. It also holds
    the binding or the protocol defaults followed, the Connections reached,
    their models, and the version.
  - `changes` holds the app's recent Catalog changes, newest first.
  - Your management token gets `403 application_token_required`.
- **The capabilities answer** (`GET /v1/aiport/capabilities`) carries
  `catalog: {version, changed_at_ms}`: the version, and the time of the newest
  kept Catalog change (`null` when none is kept).

#### The change log

Every configuration Change goes through one place. On each one, aiport works
out what it changed in every app's Catalog. Changes that can do this:

- approving or reverting an Alias;
- mapping a Level, or changing its per-effort models;
- adding a task name;
- binding an Application;
- changing a protocol default;
- removing a Connection.

A revert is a Catalog change too, with `revert_of` naming the Change it undid.
An app the Change did not affect gets nothing.

aiport's own model calls get Catalog changes recorded too. They belong to the
built-in Application `aiport`, which follows `agent_runs.connection`, else the
protocol defaults.

Catalog changes are kept as long as request history
(`telemetry.retention_days`) and survive a restart. They hold Application ids,
task names, model names and Connection ids only. They never hold keys,
addresses, prompts, responses or tokens.

**For scripts:** each Catalog change records:

- the Change id (`change_id`) and time (`at_ms`);
- the version `before` and `after`;
- task names `removed`, `added` or `moved` (with `from` and `to`);
- the binding or protocol defaults that changed (`routing`);
- the Connections the app no longer reaches or newly reaches
  (`connections_removed`, `connections_added`).

#### When a Change takes something away

Sometimes a Change removes something an app was using. Examples: removing a
task name it used, or moving its binding or a protocol default it followed.
Then aiport does more than change the version. It opens a Catalog-change Issue
(cause `catalog_changed`) for that app, saying what changed and what the app
can still use.

"Used" means that, within request history, the app's calls sent that task
name, or sent a model name that went through that binding or default. An app
that never did gets only the Catalog change.

The Issue needs no decision from you and resolves itself; see [A Catalog
change](#a-catalog-change). Before you remove a Connection, the removal dialog
shows which apps it will affect (see [Removing a Connection, and when one
fails](#removing-a-connection-and-when-one-fails)).

### Effort and reasoning

Effort says how hard one call is: `low`, `medium` or `high`. It lets aiport
pick a stronger or a cheaper model for the same name. For example, `balanced`
with high effort can go to a stronger model than `balanced` with low effort.

#### Where a call's effort comes from

aiport looks in this order and takes the first it finds:

1. The `X-Aiport-Effort` header (`low`, `medium` or `high`). This is the one
   the briefs teach. It also accepts the provider words below: `none` and
   `minimal` count as low, `xhigh` and `max` as high. A value that is not an
   effort counts as no header.
2. The provider's own effort field:
   - `reasoning_effort` on OpenAI Chat Completions;
   - `reasoning.effort` on OpenAI Responses;
   - `output_config.effort` on Anthropic Messages.

   `none`, `minimal` and `low` count as low, `medium` as medium, and `high`,
   `xhigh` and `max` as high. Numeric thinking budgets (`budget_tokens`) are
   never read or translated.
3. Otherwise, `medium`.

#### Per-effort models

Any Alias (a name aiport turns into a model on a Connection) can have its own
model for each effort, on its own Connection. That covers Levels and task
names, shared or an app's own. An effort with no model of its own uses the
Alias's `model`.

```yaml
aliases:
  - {target: hosted, alias: balanced, model: mid-model, effort: {low: small-model, high: big-model}}
origins:
  - id: docs-bot
    # ...
    aliases:
      - {alias: summarize, target: local, model: 'llama3:8b', effort: {high: 'llama3:70b'}}
```

Here `balanced` goes to `small-model` at low effort, `mid-model` at medium and
`big-model` at high. docs-bot's `summarize` goes to `llama3:70b` at high effort
and `llama3:8b` otherwise.

To edit them:

- A Level's per-effort models are in its mapping on Routing or Integrate.
  Empty fields use the mapping's model.
- A task name's are on its row of the Routing map.

Per-effort models are listed in the Catalog (`effort` on a task name), and
changing them is a Catalog change. Activity shows the effort of each call and
where it came from (see [Activity and history](#activity-and-history)).

#### What reaches the provider

The `X-Aiport-Effort` header itself is never forwarded; no `X-Aiport-*`
request header is. If the body already has the provider's effort field, aiport
never changes it.

If the body has no effort field and aiport chose the model, aiport adds one,
so the provider reasons that hard too. "aiport chose the model" means a Level,
a task name, an Alias or a module route replaced the name sent. What it adds:

- for a task name with a fixed `reasoning`: that effort, whatever the header
  says;
- otherwise, if the effort came from the `X-Aiport-Effort` header: `low`,
  `medium` or `high`;
- otherwise: nothing.

The field is `reasoning_effort` on Chat Completions, `reasoning.effort` on
Responses and `output_config.effort` on Messages. The rest of the body is kept
byte for byte. A model name sent unchanged (no Alias applies) gets nothing
added.

**Good to know:**

- Over the API, a `PUT` on a Level without `effort` keeps its per-effort
  models while the mapping stays on the same Connection. If the mapping moves
  to another Connection, they are dropped, because they name models of the old
  one.
- The same happens when a removal plan moves a Level or a task name to another
  Connection: it keeps its model and loses its per-effort models. Set them
  again on the new Connection.

### Collect feedback and improve routing

People rate the answers an app shows them, and aiport turns those votes into a
report per module and into Suggestions you can apply. This shows you which
model and effort suit each part of each app, so you can move the part that does
badly without moving the rest.

#### Votes from the application

A module is one part of an app, such as its summarizer or its search. For
feedback to work, every integration does two things:

1. It names the module of each model call with the `X-Aiport-Module` header.
2. It gives every LLM answer shown to a person a thumbs up and a thumbs down.

The communication standard and guidance version 4 ask every connected app's
coding agent to add both.

A module name is lower case, 1 to 64 characters of `a`–`z`, `0`–`9`, `.`, `_`
and `-`, and does not start with a punctuation mark. It is recorded with the
call and never forwarded.

Votes need request history (`telemetry.enabled`) and count only calls still
retained. On Integrate, the **Guidance** column shows two chips for each
Application with calls in the last 7 days:

- **Modules**: its calls name their module.
- **Votes**: its users' votes arrive.

Prompts, answers and anything a person typed are never sent with a vote: only
the request id, the vote and the module.

**For scripts:** on a vote the app sends:

```sh
curl -s -X POST http://127.0.0.1:8787/v1/aiport/feedback \
  -H "Authorization: Bearer $AIPORT_APP_TOKEN" -H 'content-type: application/json' \
  -d '{"request_id":"<X-Aiport-Request-Id of the answer>","vote":"up"}'
```

- `vote` is `"up"` or `"down"`. A `score` from 0 to 1 works instead; 0.5 or
  more counts as up.
- `module` names the module only when the call did not send
  `X-Aiport-Module`. It never replaces one the call sent.
- To change a vote, send it again; the last one counts.
- An app can rate only its own calls, and only calls still kept in request
  history. Otherwise the answer is `404 request_not_found`.
- A malformed vote gets `400 invalid_feedback`.

#### Read the report

The [Feedback](#feedback) page shows each Application's votes per module. Inside
each module it splits them by name sent, model and effort. With two or more
judged modules, it names the best and the weakest.

To improve the weakest module:

1. Change its Level, task name or effort under Routing, or apply a Suggestion.
2. Watch its share of up-votes on the Feedback page.

#### Suggestions from the votes

A Suggestion is a proposed change of model for one module, based on its votes.
aiport looks at a group of calls once it has at least 10 votes. A group is one
module's calls for one name, on one model and effort. aiport reads the votes
against your Levels and lists what it finds under **Suggestions from the
votes** on the Feedback page.

There are three kinds:

| Kind | When | What aiport suggests |
| --- | --- | --- |
| **Better elsewhere** | Below 60% up, and a module of the same name (in another application, or this one) on another model gets at least 20 points more on the same protocol, also over at least 10 votes | That model, on its Connection. The best such evidence wins; ties go to more votes. |
| **One Level up** | Below 60% up, the name sent is `fast` or `balanced`, and there is no such evidence | The next Level's model for that application (`fast` to `balanced`, `balanced` to `best`), untested for this module. |
| **Cheaper may do** | 90% up or more over at least 20 votes, and the name sent is `best` or `balanced` | The Level below's model for that application, untested for this module. |

**Good to know:**

- Evidence comes only from the same module name, so aiport compares like with
  like.
- A Level whose model already serves the module is not suggested.
- Calls that named no module never get a Suggestion.
- There is at most one Suggestion per application, module and name.

#### Apply a Suggestion as a module route

aiport never moves a module by itself. When you click **Apply**, aiport adds a
module route: that app's calls for that module that send that name go to the
suggested model. The route comes ahead of every other route, and the app's
other modules keep theirs.

Applying is one Change, with the reason from the votes included, and you can
revert it on Changes. Each app's panel on the Feedback page lists its module
routes with **Remove**. After applying, watch the module's share of up-votes:
the new model shows as its own row.

**Good to know:**

- A module route has one model for every effort: it has no per-effort models.
  A task name's fixed `reasoning` still applies.
- Its Connection must speak the request's protocol. Otherwise the call is
  refused with `400 incompatible_protocol`.
- In `aiport.yaml`, module routes are `origins[].module_routes`, each
  `{module, name, target, model}`.

**For scripts:** `PUT /api/v1/module-routes/{application}/{module}/{name}`
with `{connection, model, why?}` adds one, and `DELETE` on the same path
removes it. `why` goes into the Change as one line, cut at 300 bytes.

### Issues

An Issue is a problem aiport keeps seeing for one Application (an app or agent connected to aiport), such as calls to a model no provider serves. It sits at the top of Changes with the evidence, what aiport found out, and the ways to fix it: approve a fix in aiport, or hand a fix to the app's coding agent.

What you do with an Issue:

1. Read the evidence and what aiport found out.
2. Choose a fix. **Approve** the fix in aiport, or click **Copy fix for the
   Coding agent** and paste it to the agent in the app's repository. Or
   **Dismiss** the Issue.
3. Wait. aiport resolves the Issue by itself once the app's calls work again.

There is one Issue per Application and cause. When the problem repeats, aiport
updates that Issue's counts and time window; it never opens a second one.

#### The kinds of Issue

- **Unknown model**: the app asks for a model the provider does not have.
- **Failing Connection**: a provider the app uses keeps failing.
- **Unmapped Level**: the app asks for a Level that is not set up for it.
- **Catalog change**: one of your Changes took away something the app used.
  Nothing to approve; it is a notice.
- **Reported request**: you report a failed request yourself from Activity.

aiport opens the first four by itself. The first three open when a threshold
under `adapting:` in `aiport.yaml` is reached. You open the fifth from
Activity.

| Cause | Opens when | Fix in aiport | Fix in the Application |
| --- | --- | --- | --- |
| Unknown model (`unknown_model`) | A provider keeps answering `404` for a model the Application has no Alias for (`adapting.unknown_model`, default 5 times in 60 minutes) | An Alias for that Application to the Connection that serves the model | Send a model this gateway serves, or ask for the Alias |
| Failing Connection (`failing_connection`) | Its calls to one Connection keep failing (`adapting.failing_connection`, default 5 in 60 minutes) | Bind that Application to a working Connection | Back off and handle `5xx` and timeouts |
| Unmapped Level (`level_unmapped`) | Its calls for one Level keep being refused (`adapting.level_unmapped`, default 5 in 60 minutes) | Map the Level for that Application | Ask for the mapping, or send a mapped name |
| Catalog change (`catalog_changed`) | A Change takes away something it used | None needed | Send a name that still exists |
| Reported request (`reported_request`) | You report a failed request from Activity's **Details** | None | Whatever the failure and your note call for; the agent says it fixed it and the next call of that name and module confirms it |

The badge on Changes counts an Issue only while its fix waits for you.

#### What you can do with an Issue

- **Approve** applies the fix in aiport as an ordinary Change you can revert.
  The Issue becomes **resolving**.
- **Copy fix for the Coding agent** copies the fix in the Application: a ready
  instruction for the coding agent in the app's repository. Paste it as-is.
  **Read the fix** shows the same text first. The button is also on Integrate,
  under the app's row. The text is the same one aiport gives the app's coding
  agent about its Issues, and it holds no prompts, keys or tokens.
- **Dismiss** keeps the Issue quiet for 7 days, unless the evidence doubles.
  You can dismiss any Issue that is not resolved or dismissed yet, wherever
  its fix stands (waiting, approved, rejected or reverted). An approved fix
  stays applied until you revert its Change.
- **Reject** turns the aiport fix down and leaves the Issue as it is. The
  dashboard offers only Approve and Dismiss; Reject is for scripts.

Issues are kept across restarts. They hold only Application ids, the cause,
model names, Connection ids, counts, request ids and fixed advice text. They
never hold prompts, responses, provider messages, keys or tokens.

**For scripts:** Reject is `POST /api/v1/issues/{id}` with `action: reject`.

#### An unknown model

The app asks for a model that the provider it reaches does not serve. For
example, docs-bot asks for `llama3:8b`. It is not bound to a Connection (a
provider aiport can call, with its key), and the protocol default is a hosted
Connection that does not serve that model.

What the Issue shows:

- **Evidence**: how many calls failed, the status and the provider's error
  type (such as `404 model_not_found`), the time window, the Connections the
  calls went to, and a few request ids to look up in Activity.
- **Diagnosis**: which enabled Connections of that protocol serve the model:
  exactly one, none, several, or unknown when a listing failed. aiport finds
  out by asking each one for its model list. That is the same free listing the
  connection check uses, kept for 2 minutes; nothing is generated and nothing
  costs money. A failed listing never counts as "does not serve".

The fixes:

- **Fix in aiport**: an Alias for that app only, with the same model name, to
  the Connection that serves it. With exactly one serving Connection, the
  Connection is filled in. With none, several or unknown, choose the
  Connection yourself (and the model, if it has another name there).
  **Approve** applies it; the Issue becomes **resolving**.
- **Copy fix for the Coding agent**: the instruction says what is wrong, the
  evidence (counts, times, the Connections, example request ids), and what
  aiport found. It says what to change in the app: ask for the Alias with a
  Change request, or send a model this gateway serves, listed for that
  protocol. It says how to confirm the fix: one call with the app's own token
  answers `200`, and aiport then resolves the Issue by itself. Once the fix in
  aiport is applied, it says no code change is needed.

**How it resolves.** When the app's calls for that model succeed again, aiport
marks the Issue **resolved** and records how:

- `aiport_fix`: the Change you approved fixed it;
- `application_fix`: the app's coding agent fixed it;
- `stopped`: the calls succeeded, or stopped failing for a whole window,
  without a fix.

**Good to know:**

- A fix is credited only when a call shows it works. If the app makes no calls
  for that model for a whole window after you approved the fix, the Issue
  resolves as `stopped`, not `aiport_fix`.
- If you revert the fix, the Issue opens again and offers the fix again.
- A resolved Issue reopens when the error sticks again. Only errors after it
  resolved count.

#### A failing Connection

The app's calls to one Connection (a provider aiport can call, with its key)
keep failing. For example, docs-bot is bound to a Connection that keeps
answering `502`.

What counts as a failure: gateway errors and timeouts. These never count: a
provider answer below `500` (an unknown model, a bad request), cancellations,
and requests aiport refused. There is one Issue per Application and
Connection, and only failures to the Connection the app is bound to now count.

- **Evidence**: how many calls failed, the last status, the time window and a
  few request ids.
- **Fix in aiport**: bind that app only to a working Connection. aiport
  proposes the configured, enabled Connection with the most successful calls
  in the window. You can choose another, but never the failing one.
  **Approve** applies it as an ordinary Change you can revert; other apps are
  not touched.
- **Fix in the Application**: the request itself is fine. The coding agent is
  told not to retry in a tight loop, to back off and handle `5xx` answers and
  timeouts, and that it can ask for the binding with a Change request of kind
  `bind_application`.

It resolves itself when the app's calls succeed again, on that Connection or
on the one you bound it to (`aiport_fix` or `stopped`). It reopens if the
failures stick again. Reverting the binding opens it again with the fix
offered again.

**Good to know:**

- This fix is never applied automatically, even with automatic fixes on.
  Moving a whole app to another Connection is always your decision.
- There is no automatic fallback. When a Connection is down or keeps answering
  with errors, calls to it fail and the app gets the provider's own error.
  aiport never sends calls where you did not choose.

#### An unmapped Level

The app keeps asking for a Level (a name like `fast`, `balanced` or `best`
that stands for a model) that is not mapped to a model for it. For example,
docs-bot sends `best` on the OpenAI protocol, and no `best` is mapped there,
neither for docs-bot nor shared.

There is one Issue per Application, Level and protocol. If `best` is unmapped
on both protocols, two Issues open, each with its fix on its own protocol.

- **Evidence**: how many calls were refused, the time window and a few request
  ids. The summary names the protocol.
- **Fix in aiport**: map the Level for that app only, on the protocol its calls
  were refused on. aiport proposes the Connection the app's other calls on
  that protocol go to (its binding, else the protocol default). It leaves the
  model to you: it never makes up a model id. **Approve** applies it as an
  ordinary Change you can revert, and as a Catalog change. A Connection of the
  other protocol is refused. This fix is never applied automatically.
- **Fix in the Application**: ask for the mapping with a Change request of kind
  `application_alias` whose alias is the Level, or send another Level or task
  name the gateway maps.

It resolves itself when the app's calls for that Level succeed, and reopens if
the refusals stick again. Reverting the mapping opens it again with the fix
offered again. With the adapting loop off, nothing is counted and no Issue
opens.

#### A Catalog change

One of your Changes took away something the app was using, so the app should
switch to a name that still exists. For example, you remove a Connection and
leave docs-bot's Alias `summarize` without one. The Issue tells the app; it
needs no decision from you.

aiport opens one Issue of cause `catalog_changed` per Change, for each app that
used what went away. "Used" means that, within request history
(`telemetry.retention_days`), the app sent the removed task name, or sent a
model name (not a task name) that went to the Connection its binding or the
moved protocol default pointed at.

- **What it says**: what changed (task names removed or moved, the binding or
  protocol default and where it goes now), what the app can still use (its
  task names and the Connections whose models it reaches), and how to confirm.
  Changes and Integrate show it with nothing to approve, and it never lights
  the Changes badge.
- **How the app hears of it**: like every Issue, through `GET
  /v1/aiport/issues`, the MCP tool `aiport_my_issues`, the `open_issues` count
  in the capabilities answer, and `X-Aiport-Hint` on a failing call that sent
  a removed task name.
- **How it resolves**: by itself when the app's calls succeed after the change
  (and after its last failing call with a removed name). Or at once when its
  coding agent reports it fixed (`aiport_issue_fixed`), since there is nothing
  in aiport to confirm. Or when you revert the Change. You can also dismiss
  it.

**Good to know:**

- An app that never sent what went away, and the built-in `aiport`
  Application, get the Catalog change recorded but no Issue.
- Apps the Change took nothing from get nothing.
- A Change that is an Issue's own aiport fix (or its revert) opens no Catalog
  Issue: that Issue already says so.
- None are opened while adapting is off. The Catalog itself is still shown to
  apps.

#### What the Application's coding agent sees

The app's coding agent hears about its Issues over the connection it already
has with aiport, using the app's own token. It sees what is wrong, the
evidence and the fix to make, and it reports back when it has fixed the code.
Your management token gets `403`.

Where the agent finds its Issues:

- The MCP tool `aiport_my_issues`, or `GET /v1/aiport/issues` (and
  `GET /v1/aiport/issues/{id}` for one), lists the app's open and recent
  Issues. Each has what is wrong, the evidence, the fix in the Application
  (`application_fix`, the same text as **Copy fix for the Coding agent**), the
  root cause when you had the Issue explained (`explanation`), and whether the
  aiport fix is proposed, applied, rejected or reverted. An app never sees
  another app's Issues (their ids answer `404`), and never sees your notes.
- The capabilities answer (`aiport_capabilities`) shows `open_issues`: the
  number of its Issues that are open or resolving.
- A failing call that is one of an open Issue's failures carries an
  `X-Aiport-Hint` header naming that Issue, so the app's own logs point the
  agent to it. That is a provider `404` for the same model (unknown model); a
  `5xx`, timeout or connection error from that Connection (failing
  Connection); or a failing call that sent a removed task name (Catalog
  change). Other failing calls, and aiport's own refusals that are not the
  Issue's cause, name no Issue. Issue hints are sent whether or not the app may
  request changes: telling it about its Issues is not a Change request.
- The `AGENTS.md` line tells later sessions to check `aiport_catalog`,
  `aiport_my_issues` and `aiport_guidance` at the start.

When the agent has fixed the app's code, it says so with `aiport_issue_fixed`
(or `POST /v1/aiport/issues/{id}/fixed`). The Issue becomes **resolving** by
the Application fix, so you can see it is being handled. An aiport fix waiting
for you keeps waiting, and you can still approve it. Only `open` Issues can be
marked; others get `409 issue_not_open`.

aiport does not take the agent's word for it. The Issue resolves as
`application_fix` only once the app's calls succeed again for the same cause:

- for an unknown model, a successful call for that same model;
- for a failing Connection, a successful call routed for that app after the
  agent said so.

A successful call for another model does not count, and neither does silence.
If the error goes on, the Issue opens again and offers the aiport fix again.

### Explain

Explain tells you, in a few sentences, why an Application keeps failing, so you
can pick the right fix. Click **Explain** on an Issue (a problem aiport noticed
in an app's calls). A model you choose writes the answer, or the app's coding
agent reads its code and writes it.

You find **Explain** on each Issue on the Changes page, and under the
Application's row on Integrate.

Who writes the explanation:

- **No coding agent connected to the Application:** a Connection (a provider
  aiport can call, with its key) and model that you choose once in
  **Settings → Explanations**. For example, a local Connection serving
  `llama3:8b`. Until you choose one, Explain is disabled, with a hint saying
  so.
- **A coding agent connected:** the agent reads the Application's code instead.
  See [Explain by a Coding agent](#explain-by-a-coding-agent).

In `aiport.yaml` the Connection and model are `adapting.explain`.

#### What is sent to the model

Only the Issue's evidence:

- the cause, the status and the provider's error type;
- the counts and the time window;
- example request ids;
- the model name;
- the Connection ids;
- aiport's diagnosis: which Connections list the model, and the models they
  list.

Never sent: keys, tokens, the Application's id or name, prompts, or provider
messages.

The only text in it that an Application wrote is the model name it asked for.
That name is sent only when it looks like a model id: at most 128 characters,
all letters, digits or `. _ : / @ + -`. Any other name is replaced by "(a model
name that is withheld because it contains unusual characters)". Connection ids,
model lists and the error type pass the same check. A coding agent is told
exactly the same.

#### While it writes, and after

- The words appear as the model writes them.
- When it finishes, the explanation is saved with the Issue and kept across
  restarts. It shows who wrote it.
- **Explain again** writes a new one that replaces it.
- If writing fails or takes longer than 2 minutes, nothing is saved and any
  earlier explanation stays.
- If the explanation hits its length limit, it stops there.
- Only one explanation per Application is written at a time.

#### Where the explanation goes next

The saved explanation is the "Root cause" paragraph at the start of "Copy fix
for the Coding agent". The Application's coding agent also gets it from
`aiport_my_issues` (and `GET /v1/aiport/issues`), both in that text and as the
`explanation` field.

#### In Activity, and what it costs

The call shows in Activity under the built-in Application `aiport`, with its
Connection, model, status and latency, like any other call. No configured
Application may use the id `aiport`. A remote Connection bills explanations to
its key; a local one keeps them on this machine.

**For scripts:**

- `PUT /api/v1/adapting/explainer` with `{"connection": "local", "model":
  "llama3:8b"}` chooses who writes. It is a revertable Change; both empty turns
  Explain off.
- `POST /api/v1/issues/{id}/explain` streams JSON lines: `start` (who writes),
  `text` (the next words), then `done` (the Issue with its saved explanation)
  or `error`.

#### Explain by a Coding agent

When an Application has a coding agent connected, Explain lets that agent read
the app's code and write the root cause, usually naming the file and the change
to make. The agent works read-only in a throwaway copy of your repository, and
only when you click.

This needs three things: the Application has a coding agent, it has a
repository path, and the agent's ACP entry point (the program aiport talks to)
is installed. Then the **Explain** button's tooltip names the agent, and
clicking it starts an Agent run (a coding agent that aiport starts on your
click). Otherwise the Connection from Settings explains, as in
[Explain](#explain). To set an agent up, see
[Connecting a coding agent](#connecting-a-coding-agent).

##### What aiport cannot see

Read this before you use Agent runs.

aiport only sees the permission requests an agent sends it. An agent's own
settings can let it read, edit or run commands without asking. Those actions
never reach aiport: it cannot allow them, deny them or log them.

aiport limits what such settings can do, in layers:

- **Every run works in a worktree made for it, never in your working copy.**
  Explain's worktree is thrown away when the run ends; Fix it's is shown to you
  as a diff. An edit nobody was asked about lands only there.
- **It picks the most restrictive session mode the agent offers** (see Session
  modes below), and does not run a session that can only act without asking.
- **Fix it never allows edits to agent or git configuration** in the worktree,
  and any edit the agent reports ends "Allow until the next edit".
- **Before a run, the run panel warns you** when the repository's own agent
  settings allow actions without asking. It looks for:
  - `.claude/settings.json` or `.claude/settings.local.json` with
    `permissions.allow` entries, or a `defaultMode` such as
    `bypassPermissions` or `acceptEdits`;
  - `.gemini/settings.json` with `autoAccept`, or an approval mode such as
    `yolo`;
  - `.codex/config.toml` with `approval_policy = "never"` or
    `sandbox_mode = "danger-full-access"`.

  aiport only reads these files. It never changes them.

What aiport cannot check: your user-level agent settings (for example
`~/.claude/settings.json`, `~/.gemini/settings.json` or `~/.codex/config.toml`)
and any managed settings. A command those settings allow runs without aiport
asking you. It runs with your own rights, in the run's worktree, or elsewhere
if the agent chooses.

**Warning:** do not auto-allow commands or edits in your user-level agent
settings for repositories you use with aiport. Check the warnings in the run
panel.

##### How a run starts

- Only when you click. Never by itself, not even with automatic fixes on, and
  never with adapting off.
- aiport starts the ACP entry point it found by its absolute path, never
  through a shell.
- The agent gets a minimal environment: `PATH` (the agent's own directory
  first, then the absolute directories of aiport's `PATH`, so an npm-installed
  adapter finds `node`), `HOME`, `TMPDIR`, `LANG`, `LC_ALL`, `USER`, `LOGNAME`
  and, for Claude Code, the variables in
  [Where the run's model calls go](#where-the-runs-model-calls-go).
- It never gets aiport's tokens or provider keys, and has no MCP connection to
  aiport.

##### A throwaway copy, never your working copy

aiport makes a git worktree of the repository for the run, in
`aiport.yaml.worktrees/`. It is detached at your current commit (`HEAD`, no
branch), and aiport removes it when the run ends. The agent reads that copy,
never your working copy. So even an edit that the agent's own settings allow
without asking cannot touch your files.

Your uncommitted changes are not in that copy. The repository needs at least
one commit.

##### What the agent is told

The Issue's evidence (the same as in [Explain](#explain)), and that the run is
read-only.

##### Read-only: what the agent may do

aiport answers every permission request itself, without asking you:

- Reading and searching inside the run's copy: allowed.
- Every edit, delete or command: denied.
- Anything outside the run's copy, including your own working copy: denied.
- A permission request that names a session aiport did not open: denied.

aiport only ever answers "once", so nothing is remembered in the agent's own
settings.

**Which paths count.** Every path a request names counts, not only its
locations. That includes the paths anywhere in its raw input: `path`,
`file_path`, `cwd`, a move's `destination`, the files of a patch, and the base
of a search pattern such as `/etc/**` or `../*.py`. One path outside the copy
(or in your home directory, `~`) denies the whole request. A read or search
that names no path at all is denied too. The same rules apply to Explain and
Fix it.

##### Session modes

Some agents offer modes. Claude Code's adapter, for example, offers `default`,
`acceptEdits`, `plan` and `bypassPermissions`. aiport switches the session to
the most restrictive suitable mode:

- for Explain, plan or read-only;
- for Fix it, one that asks before acting, such as `default`.

It never picks a mode that bypasses permissions or auto-accepts. If the session
starts in such a mode (from your own settings) and the agent offers no other,
aiport does not run it. The run panel names the mode chosen.

##### The run panel

The run panel under the Issue shows the agent's name, where its model calls
go, its steps (thoughts, files read, permissions allowed or denied) and its
words as they come.

- **Stop** ends the run at any time. The agent is asked to cancel, then its
  whole process group is ended. Nothing is saved from a stopped run.
- Closing the tab or leaving the page does not stop the run. It goes on, saves
  its explanation when it ends, and the run panel picks it up again when you
  come back.

##### Limits

- One Agent run per Application at a time.
- An Explain run stops after 10 minutes (`agent_runs.explain_minutes`), and
  nothing is saved.

##### What is kept

- The explanation is saved with the Issue. It replaces any earlier one, also
  one a Connection wrote. It is passed to the Application's coding agent as in
  [Explain](#explain).
- The run's log (its steps and words) is kept in memory only, while the Issue
  is open or resolving. It is never written to disk, and it is gone when the
  Issue closes or aiport restarts.

##### Where the run's model calls go

The agent itself calls a model while it works. You choose where those calls go
in **Settings → Default coding agent → Model calls of Agent runs**: an
Anthropic-protocol Connection and, optionally, a model.

Until you choose one, every run uses the agent's own login, and the run panel
says so. aiport never sends a run's calls to the Anthropic protocol default
unless you chose it for Agent runs.

How each agent behaves:

- **Claude Code** is pointed at aiport through its environment only:
  - `ANTHROPIC_BASE_URL`: aiport's address on this machine (the listen port on
    `127.0.0.1`, or the listen address when that is a loopback address);
  - `ANTHROPIC_AUTH_TOKEN`: a credential made for this run;
  - `ANTHROPIC_MODEL`, when you chose a model.

  That credential only authenticates model calls, as the built-in Application
  `aiport`. It cannot reach the dashboard, the MCP endpoint or any
  Application's data. It is never written down or logged, and it stops working
  when the run ends. The calls show in Activity under `aiport` with their
  Connection. With no Connection chosen for Agent runs, or when the chosen one
  is not reachable, Claude Code uses its own login and the run panel says so.
- **Codex** uses its own login. Pointing it at aiport would mean editing its
  configuration file, which aiport never does. The run panel says so.
- **Gemini CLI** uses its own login too, because aiport does not speak the
  Gemini API yet. The run panel says so.

```yaml
agent_runs:
  default_agent: claude
  connection: claude-hosted   # an Anthropic-protocol Connection
  model: claude-model-id      # optional
  explain_minutes: 10
```

**For scripts:**

- `PUT /api/v1/agents/calls` with `{"connection", "model"}` sets where model
  calls go (a revertable Change).
- `GET /api/v1/runs` lists the runs going; `POST /api/v1/runs/{id}/stop` stops
  one.
- `GET /api/v1/issues/{id}/run` gives the Issue's last run and its log, while
  the Issue is open.
- For an Agent run, the explain stream adds `run` events (the run: its id,
  agent and model calls) and `step` events. `error` codes include
  `run_stopped` and `explanation_timed_out`.

#### Fix it

Fix it lets the Application's coding agent change the app's code on a branch of
its own. It ends with a diff that you review, then keep or discard. Nothing is
committed until you click **Keep the branch**.

Fix it shows on an open Issue when the Application has a coding agent and a
repository path, as for [Explain by a Coding agent](#explain-by-a-coding-agent).
Otherwise a one-line hint under the Issue says what is missing, for example
"Add the repository path of docs-bot".

It is offered only when the cause can be in the app's code: the app asks for a
model name that no Connection serves. When a Connection is failing, the Issue
says instead that the cause is not in the app's code, and points to the aiport
fix (binding the Application to a working Connection).

##### Its own worktree and branch

- aiport makes a git worktree of the repository on a new branch named for the
  Issue: `aiport/fix-<issue id>`, with `-2`, `-3` and so on if that branch
  exists. It starts at your current commit (`HEAD`).
- The worktree lives next to the configuration, in `aiport.yaml.worktrees/`,
  not inside your repository.
- Your working copy, its uncommitted changes and your branch are never
  touched. Your uncommitted changes are not in the worktree either.
- The repository needs at least one commit.

##### What the agent may do

aiport answers every permission request itself:

- **Allowed:** reading, searching, and editing, creating, deleting or moving
  files inside the worktree. Symbolic links are followed, so a link that points
  out of the worktree counts as outside.
- **Never allowed**, at any level of the worktree and in any letter case:
  - the agents' own configuration: `.claude/`, `.codex/`, `.gemini/`,
    `.mcp.json`, `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `GEMINI.md`;
  - git's configuration: `.git/`, git hooks in `.husky/` or `.githooks/`,
    `.gitattributes`, `.gitmodules`.

  So a branch you keep cannot change how your agents or git behave.
- **Asks you:** a command inside the worktree waits for your answer (see
  [Approving commands](#approving-commands)).
- **Denied without asking:** anything outside the worktree, including your own
  copy of the repository and a command that says it runs elsewhere.

aiport only ever answers "once" to the agent, so nothing is remembered in the
agent's own settings. The agent starts as for Explain: by absolute path, with
the minimal environment, only on your click.

##### What the agent is told

The Issue's evidence, the saved explanation if there is one, the branch it
works on, and that it must not commit, merge or push.

##### While it runs

- The run panel shows its steps and the permissions allowed or denied, with
  **Stop**. The run goes on if you leave the page.
- **Time limit:** a run stops after 30 minutes (`agent_runs.fix_minutes`).
  What it changed by then stays in its worktree and waits for your review with
  **Keep the branch** and **Discard**, as after a finished run (state
  `timed_out`).
- **Nothing left behind:** these runs remove the worktree and the branch:
  Stop, a run that fails, and a run that changed nothing (also one that timed
  out or went unanswered before changing anything). The Issue stays open and
  shows why the run ended.
- **Cleanup at startup:** if aiport stopped or crashed while a Fix it run was
  going, the run is marked stopped the next time aiport starts, and its
  worktree and branch are removed. So are the throwaway worktrees of Explain
  runs that were going. Only directories inside `aiport.yaml.worktrees/` and
  branches named `aiport/fix-…` are ever removed.

##### What is kept

The branch name, the commit it started from, the diff summary and the changed
files are saved with the Issue, across restarts. Once you keep the branch, so
is the commit aiport made. The diff itself is read from the worktree or the
branch when you look at it. The run's log is kept as for Explain: in memory
only.

##### Approving commands

When the agent asks to run a command (tests, a build), the Issue shows it and
waits for your answer.

What you see is exactly what runs:

- The whole command, exactly as the agent sent it, as plain text. It is the
  agent's command, not aiport's.
- Never cut: a long command scrolls.
- Characters that would not show as themselves are written as escapes:
  control characters, invisible and direction-changing characters (such as
  U+202E, the right-to-left override), and unusual spaces.
- A list of arguments is shown shell-quoted (`echo 'a b'`).
- A command longer than 16 KiB is denied without asking, never shown in part.

There are three answers:

| Answer | What it allows |
| --- | --- |
| **Allow once** | This one time. |
| **Allow until the next edit** | This time, and every time the run asks for the exact same command again, until the agent changes a file (whether it asked to or not). "The same" means the same list of arguments, or the same string, in the same directory: `echo 'a b'` is not `echo a b`. Another command, or the same one in a later run, asks again. |
| **Deny** | Nothing. The agent is told no and goes on without it. |

An edit can change what a command does, which is why "Allow until the next
edit" asks again after one. But aiport only sees the edits the agent asks about
or reports. A command you allowed that changes files itself (a code generator,
a formatter) does not end the allowance. So choose **Allow once** for commands
that write files.

The agent runs an allowed command itself, in its own process, in the
worktree's directory. aiport only answers the permission. Whatever the command
changes in the worktree shows in the diff.

While a command waits, the Changes badge in the Monitor menu counts it. A
command left unanswered for 5 minutes stops the run. What it changed so far
stays in its worktree, and the Issue shows it for review with **Keep the
branch** and **Discard**, as after a finished run (state `stopped_waiting`).

##### Reviewing the diff

When the agent ends its turn, the Issue shows **Review the change**: the
branch, the summary (for example "2 files changed, 3 insertions(+), 1
deletion(-)") and the whole diff. Nothing is committed yet.

- **Keep the branch:** aiport commits the diff on the new branch as one commit,
  authored by `aiport <aiport@localhost>`, and removes the worktree.
  - This is the only commit aiport ever makes, and only when you click Keep. It
    goes on the new branch, never on yours, with hooks off.
  - The branch stays in your repository for you to look at, merge or delete.
    aiport never commits to your branch, merges or pushes.
  - The Issue becomes resolving, as an Application fix. It resolves when the
    Application's calls succeed again, that is, after you have merged and
    deployed the change.
  - "Copy fix for the Coding agent" names the branch.
- **Discard:** aiport removes the worktree and deletes the branch. The Issue
  stays open, and you can Fix it again.

##### git, safely

- aiport finds git once, at startup, by absolute path: `/usr/bin/git` when it
  is there, otherwise the first `git` in its `PATH`. The startup output names
  it.
- It runs git with a minimal environment, without your global or the system's
  git configuration and attributes (`GIT_CONFIG_GLOBAL=/dev/null`,
  `GIT_CONFIG_NOSYSTEM=1`, `GIT_ATTR_NOSYSTEM=1`,
  `core.attributesFile=/dev/null`).
- Hooks, fsmonitor, commit signing and external diff programs are turned off.
- What the repository itself sets in its own `.git/config` and
  `.gitattributes` still applies, for example a clean or smudge filter it
  defines. It is your own repository, so check those files before you use Fix
  it on a repository you did not set up.
- aiport removes a worktree only when its directory is directly inside
  `aiport.yaml.worktrees/`. Any other recorded path it finds is left alone and
  logged.

```yaml
agent_runs:
  fix_minutes: 30
```

**For scripts:**

- `POST /api/v1/issues/{id}/fix` starts a run and answers `202` with the run.
  It answers `409` with `fix_unavailable` and the hint, `fix_not_for_cause`
  for any cause but an unknown model, `fix_waiting` while an earlier diff
  waits for you, `agent_run_running`, `issue_not_open` or `worktree_failed`.
- `GET /api/v1/issues/{id}/fix` gives `branch_fix` and the `diff`. `POST
  /api/v1/issues/{id}/fix/keep` and `…/fix/discard` keep or discard it.
- `GET /api/v1/issues` says per Application, in `fix_by`, whether Fix it is
  available: with which agent and time limit, and for which `causes` (with
  `cause_hint` for the others), or the hint. Each Issue carries its
  `branch_fix`.
- Commands waiting for you: `GET /api/v1/approvals`, each with `id`, `run`,
  `application`, `issue`, `command` and `expires_at_ms`.
- Answer one with `POST /api/v1/approvals/{id}` and `{"answer": "allow_once"}`,
  `"allow_run"` (allow until the next edit) or `"deny"`. Once it is answered
  or its run has ended, this gives `404 approval_not_found`.
- `GET /api/v1/runs` gives each run's `waiting` count, and `GET
  /api/v1/issues/{id}/run` the Issue's waiting `approvals`.
- These are management routes, never open to an Application token. Answering
  needs the dashboard's CSRF header.

### Connecting a coding agent

Connect a coding agent to an Application so aiport can look at the app's code,
on your click. Explain reads the code and tells you the root cause, and Fix it
makes the change on a new branch and shows you the diff to keep or discard.
Explain only reads, nothing runs without your click, and aiport never commits
to your branch, merges or pushes.

Without an agent, an Issue (a problem aiport noticed in an app's calls) still
offers the fixes aiport can see from its side. Explain still works too, more
briefly: it uses only the Issue's evidence, with the Connection (a provider
aiport can call) and model chosen under **Explanations** in Settings.

#### Set it up

aiport can start a coding agent installed on this machine to work on an
Application's Issues, only when you click. These are called Agent runs. Each
Application can have its own agent (**Claude Code**, **Codex** or **Gemini
CLI**) and the path of its git repository.

1. Open step 6 of the onboarding wizard, or the **Agent** column on Integrate.
2. Choose the agent. If it is missing, install it first (see below).
3. Enter the repository path.
4. Click **Save**.

Each save is one revertable Change on the Changes page.

**Default coding agent** on the Settings page is the agent for every
Application that has none of its own.

#### What aiport looks for

aiport talks to agents through the Agent Client Protocol (ACP).

| Agent | What aiport needs |
| --- | --- |
| Gemini CLI | Nothing extra: it speaks ACP itself (`gemini --acp`). |
| Claude Code | The adapter `claude-agent-acp` (npm package `@agentclientprotocol/claude-agent-acp`). |
| Codex | The adapter `codex-acp` (npm package `@agentclientprotocol/codex-acp`). |

aiport finds programs by absolute path. It looks in the absolute directories of
its own `PATH` and in the usual install places (`~/.local/bin`,
`~/.npm-global/bin`, `/opt/homebrew/bin`, `/usr/local/bin` and a few more). It
does not run them to look.

#### Nothing is installed for you

When an agent or its adapter is missing, the step shows the exact command to
install it, such as `npm install -g @agentclientprotocol/codex-acp`.

1. Run the command yourself.
2. Refresh the page.

An agent that is not installed cannot be chosen.

#### The repository path

The path must be absolute and must be the top directory of a git repository
(where `.git` is). Otherwise it is refused, with the reason. aiport stores it
with symbolic links resolved. Agent runs never work outside it.

#### In the configuration file

In `aiport.yaml` these are `agent` and `repository` on the Application under
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

**For scripts:** `GET /api/v1/agents` lists what aiport found: `installed`,
`ready`, the ACP entry point, and `install` when something is missing.

### What waits for your decision

Suggestions and Change requests are changes to aiport's configuration that
someone else proposes. Each one waits on the Changes page until you approve,
dismiss or reject it, and nothing in this list changes aiport until you decide.

There are two kinds.

**Suggestions** come from aiport itself, from patterns in request history. Each
shows its evidence: how many requests, over which time window.

- aiport suggests pairing when requests keep arriving with a missing or unknown
  token (`adapting.unknown_token`). This one is informational: open the
  Integrate page to pair the application, or click **Dismiss**.
- Two other fixes come as [Issues](#issues) instead: an Alias (another name for
  a model) for a model one Application keeps getting `404` for, and binding an
  Application to a working Connection (a provider aiport can call) when its
  calls to one Connection keep failing.
- The thresholds are under `adapting:` in the configuration.
- Suggestions from the votes are different. They live on the Feedback page
  (see [Suggestions from the votes](#suggestions-from-the-votes)).

**Change requests** come from Applications (the apps and agents that call
aiport) that you allowed to request changes.

- An Application names a task, not a provider, for example the alias
  `review-code`. You decide which Connection and model serve it.
- It may add advice, such as "suggests big-model on hosted", with its reasons.
  The advice pre-fills the form. You can choose otherwise now, or change it
  later without touching the Application's code: in `origins[].aliases` in the
  configuration file, or with a new Change request.

#### Reading a Change request

Each Change request shows what it asks for. **Details and reasons** opens the
rest: the full details, why it is needed, the argument for the Scope asked for,
and who filed it (as the request states).

When several Applications send identical requests, they are merged into one
item, each with its own argument. One decision answers them all.

Request text is plain text from the Application. It is shown as is and never
sent to a model. Each field is at most 2,000 characters, and 6,000 in total.

#### Waiting, Accepted, Rejected

The **Waiting / Accepted / Rejected** filter on **Suggestions and Change
requests** picks what the list shows.

- **Waiting** is what needs you.
- **Accepted** and **Rejected** show past decisions, most recent first, with
  who decided, when, and the note. Rejected includes dismissed Suggestions.

Decided items are kept as long as request history
(`telemetry.retention_days`). An accepted item's Change stays in **Applied
changes**, where you can revert it.

Kinds of request aiport cannot apply are kept as "not supported" and counted
under **Asked for, not supported yet** (kinds and counts only). They need no
decision.

#### Scope

Every decision has a Scope: it applies **only** to the requesting Application,
or to **every Application**. The requester argues for one; you choose. For an
Alias, every Application means a shared Alias. For a binding, it means the
protocol's default Connection.

#### Approve, Dismiss, Reject

- **Approve** applies the Change at once. For an Alias you confirm the alias
  name, the Connection and the model; for a binding, the Connection. You can
  add a note of up to 2,000 characters.
- **Dismiss** a Suggestion: aiport stays quiet about it for 7 days, unless the
  evidence doubles.
- **Reject** a Change request: your note goes back to the Application, which
  sees it when it lists its requests.

#### Allowing Change requests

An Application may request changes only after you allow it. Do one of these:

- tick **May request changes** when you approve its pairing;
- use the toggle on Integrate;
- use the toggle in the Applications table on Routing.

Until then it gets `403 change_requests_not_allowed`.

Good to know:

- Allowed Applications also get hints when an error is one a Change could fix:
  `error.hint` on `target_unconfigured` and `target_disabled`, and the
  `X-Aiport-Hint` header on a provider `404` for the requested model.
- Change requests are limited to 20 per Application per hour (`429
  change_requests_rate_limited`). A request that a policy refuses does not
  count toward that limit.

**How an Application writes a good Change request.** The coding-agent brief
teaches it these fields:

- `kind`, from `change_kinds`: `application_alias` or `bind_application`;
- `need`: one line;
- `details`;
- `why`: the error code, the `X-Aiport-Request-Id`, what it tried;
- `scope`: `application` or `every_application`, and `scope_case`;
- optional: `filed_by`, `alias`, `prefer_connection` and `prefer_model`.

A `400 invalid_change_request` lists every problem in `error.problems`. The
Application reads the answer with `aiport_my_requests`: `pending`, `accepted`,
`rejected` (with your note) or `not_supported`.

### Applied changes and Revert

Applied changes lists every configuration Change made through aiport, newest
first, with who made it. **Revert** undoes any one of them, and the list
survives restarts, so you can always get back to an earlier configuration.

A Change is made by you, by an approved Suggestion or Change request, or by
**aiport (automatic)** for an automatic fix.

**Revert** restores the configuration file exactly as it was before that
Change, and makes it live at once. If the file was changed after that Change,
revert the newer Changes first: aiport refuses rather than overwrite them.

Applied Changes are never pruned.

#### Automatic fixes

Automatic fixes let aiport apply an Issue's aiport fix itself, without waiting
for you, in a few safe cases. They are off by default.

To turn them on, use the **Automatic fixes** switch under Applied changes, or
set `adapting.auto_fix: true` in `aiport.yaml`. The switch itself is a Change
you can revert.

aiport applies a fix automatically only when all of these hold:

- **The diagnosis is certain:** exactly one enabled Connection serves the
  model. Never when none, several or an unknown number do.
- **The error is sticky:** the Issue's threshold (`adapting.unknown_model`) was
  reached.
- **The fix touches only that Application:** an Alias for it alone. These are
  never automatic: binding a whole Application to another Connection, Level
  mappings, shared Aliases, protocol defaults, or anything for every
  Application.
- **You have not refused that fix** (see below), and the adapting loop is on.

For example, docs-bot keeps getting `404` for `llama3:8b`, and only the local
Connection serves it. aiport adds the Alias `llama3:8b` → local for docs-bot,
and the Issue becomes resolving.

- The Change is recorded as made by **aiport (automatic)**.
- Applied changes shows it as "aiport fixed: …" with **Revert**.
- The Issue shows "aiport fixed this automatically", with its own Revert.
- Guardrail Extensions judge automatic fixes just as they judge your
  approvals.

**When you say no.** If you revert an automatic fix, aiport records that. It
never applies the same fix automatically again for that Application and cause,
also after a restart. The Issue opens again and its fix waits for your
decision; you can still approve it.

A fix you rejected is never applied automatically either, even if you turn
automatic fixes on later. You can still approve it yourself.

To remember these decisions, an Issue that holds a fix you reverted or rejected
is kept beyond request-history retention (`telemetry.retention_days`), even
once it is resolved or dismissed. It holds no more than any other Issue: ids,
model names, counts.

#### The sound

The speaker button next to **Applied changes** plays a soft sound when
something new waits in Changes. It is off by default, and the setting is kept
in this browser only.

### Compress requests

An application can ask aiport to shrink a request before it goes to the
provider, and aiport measures whether that makes the answers worse.
Compression is off unless the application asks for it, request by request,
with the `X-Aiport-Compress` header.

To turn it on for a request, the application sends the methods it wants, for
example:

```
X-Aiport-Compress: whitespace, json
```

| Method | What it does |
| --- | --- |
| `whitespace` | Removes trailing spaces, normalises line endings, and turns two or more blank lines in a row into one. Indentation is kept. |
| `json` | Rewrites text that is entirely JSON (typical tool results) in compact form. |
| `dedupe` | When a block of 1 KiB or more repeats, earlier copies become a short note and the latest copy stays. Lossy. |

`all` picks every method. `off`, or no header, picks none.

To see whether it is worth it, open the [Compression](#compression) page: it
shows the savings and a verdict for each combination of methods. The full
description, including the evaluator, is in [compression](compression.md).

Good to know:

- Only message text is changed. Tool definitions, schemas and metadata are
  never touched.
- If compression fails or saves nothing, aiport sends the original bytes.

**For scripts:**

- An unknown method returns `400 unknown_compression`.
- Every compressed response carries `X-Aiport-Compression` with the methods,
  the status (`status=applied`, `no_gain` or `failed`), and the bytes before
  and after.

### Add Extensions

An Extension is a program of your own that helps decide on Change requests
(changes an Application asks aiport to make). Use one to enforce your own
rules without changing aiport.

There are two kinds:

- A **policy** runs when a Change request arrives and may refuse it. The
  Application then gets `403 change_request_refused` with the reason.
- A **guardrail** runs when you approve, before the Change is applied, and may
  block it (`guardrail_blocked`). Changes then shows which Extension blocked
  it and why, and the item keeps waiting.

To add one:

1. List it under `extensions:` in `aiport.yaml`, with an absolute path.
2. Restart aiport.
3. Check that it appears in the **Extensions** panel on Changes.

aiport runs only programs you listed, each with a timeout. Anything other
than an explicit allow counts as a deny. The input and output protocol is in
[Extensions](extensions.md).

### Suggest to aiport and Share upstream

You can send ideas back to the aiport project in two ways. Both give you a
draft to check first; nothing is ever sent on its own.

- **Suggest to aiport** sits under **Asked for, not supported yet**. Use it
  for a kind of Change that Applications asked for and aiport cannot apply.
- **Share upstream** sits in the **Extensions** panel of Changes. Use it to
  offer an Extension (a program you added that judges Change requests) you
  listed. The draft names the hook and the input protocol, never the
  Extension's name, path or settings. You add the program yourself, after
  removing anything private.

What a draft contains: kinds, hooks and anonymous counts only. It never holds
configuration, keys, paths, Application names or request text. One thing to
check: a kind named by an Application is in that Application's own words.

You can edit both the title and the body. Every way of sending uses your
edited text:

- **Open on GitHub**, when `upstream.repository` is set.
- **File with gh…**: after you confirm, aiport runs `gh issue create` with
  your own gh login.
- **Copy text**.

### Turn the adapting loop off

The adapting loop is the part of aiport that watches request history and
proposes fixes: Suggestions, Issues, Change requests and Explain. You can turn
it off and keep everything else working.

1. Add this to `aiport.yaml`:

   ```yaml
   adapting:
     disabled: true
   ```

2. Restart aiport.

What stops:

- aiport finds no Suggestions and opens no Issues, so Changes shows nothing
  new.
- Change requests and `/v1/aiport/issues` answer `409 adapting_off`, and the
  Issue tools are not offered over MCP.
- Capabilities carry no `open_issues`, no hints are sent, and Explain and
  Agent runs are unavailable.
- `/_connect` and both briefs mark change requests, Issues, their MCP tools
  and `X-Aiport-Hint` as off, and say why.

What keeps working:

- Calls are sent to providers exactly as before.
- Revert works on earlier Changes.
- The Catalog (the list of names each Application can send) is still shown
  to Applications: `/v1/models`, `X-Aiport-Catalog`, `/v1/aiport/catalog` and
  `aiport_catalog`. Catalog changes are still recorded, but no Catalog-change
  Issues are opened.
- Levels, guidance, the communication standard and votes.

The detector thresholds and the check interval (`adapting.interval_seconds`)
are in [configuration](configuration.md).

### Use a name instead of 127.0.0.1

You can open the dashboard at `http://aiport.localhost:8787/` instead of the
IP address, because browsers send any `*.localhost` name to this machine by
themselves. The name is for your browser only; applications and coding agents
keep using the numeric address.

1. Add this to `aiport.yaml`:

   ```yaml
   hostnames: [aiport.localhost]
   ```

2. Restart aiport.

The dashboard then works at the name straight away, and every page has its
address there, for example `http://aiport.localhost:8787/integrate`.

Rules for names:

- Lowercase, with no port.
- Either `localhost` or ending in `.localhost`.

aiport still listens on `127.0.0.1`. It still refuses every name you did not
list, with `403 invalid_host` or `invalid_origin`.

**Why apps keep the number.** Applications and coding agents are always given
`http://127.0.0.1:8787`, even when hostnames are set. That covers the agent
instruction, both briefs, the communication standard, `/_connect`, the
`.aiport` lines on a token card and the MCP entry. A name causes three
problems that the number does not:

- it needs a lookup the system resolver may not do;
- it resolves to `::1` first, and aiport listens on IPv4;
- a coding agent's sandbox sends it through its network proxy, which refuses
  it.

## A tour of the dashboard

The dashboard is where you run aiport day to day: it shows your traffic, holds
your Connections and routing, connects your Applications, and lists what
waits for your decision. The Overview page shows the traffic at a glance and
one line for each thing that needs you.

The sidebar has three groups:

- **Monitor**, the pages you use every day: Overview, Activity and Changes.
- **Configure**: Integrate, Providers, Routing, Agents, Compression and
  Feedback.
- **Workspace**, at the bottom: Tasks and Settings.

The bottom of the sidebar also shows the address aiport listens on, and
**Sign out** (or "Authentication off" when sign-in is turned off).

The toolbar at the top right has:

- the help button, which opens this manual;
- the **Light** / **Dark** / **System** appearance switch;
- Refresh (on Tasks, also the chime).

The dashboard refreshes itself every 10 seconds while it is open. Refresh
fetches everything at once, which helps right after you edited the
configuration file by hand.

The ⓘ next to a page title or panel opens the start of the matching manual
section, with a link to the whole section.

Each page has its own address (`/integrate`, `/changes`, `/manual` and so on).
Reloading stays on the page, Back and Forward move between pages, and a link
opens the page it names.

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

Overview shows the shape of your AI traffic at a glance, and tells you when
something needs you.

The traffic part shows request counts, average and P95 latency, error rate,
hourly traffic, the provider distribution (with each provider's logo) and the
latest requests.

Below that, you get one line for each thing that needs you:

- Something waits for your decision. Use **Review in Changes**.
- An Application's coding agent has not confirmed the current guidance
  version. Use **See Guidance in Integrate**.
- The latest request on a route failed because of something in your
  configuration. The line gives the reason and the fix, and goes away once a
  request there succeeds.

A banner says when request history is off, kept for this session only, or
incomplete.

### Activity and history

Activity lists every recorded request, so you can see who called what, with
which name and effort, what served it and why it failed. aiport records only
metadata: time, route, Application, model, effort, module, status and
latency. It never records prompts, responses or headers.

How to use it:

- Filter by period, Application (**Origin**), Connection (**Target**),
  outcome, and the name sent (**Name**).
- **Load older requests** goes further back.
- **Details** opens one request right under its row. It shows, in order: why
  it failed and what fixes it; what the Application asked for (name, module,
  effort); what served it (Connection, model, provider, route); the result
  (outcome, latency, the user's vote); and its id and time.

#### Hand a failure to the Application's coding agent

Under a failed request from an Application, **Report to the Application's
coding agent** opens an Issue (a problem aiport tracks for one Application)
with the cause `reported_request`. The Issue holds the request's metadata,
aiport's explanation and your note.

1. Open **Details** on the failed request.
2. Paste the error the application showed into the note. Request history
   keeps no provider messages, so this is the only place the error text comes
   from.
3. Select **Report to the Application's coding agent**.

What happens next:

- The coding agent reads the Issue at the start of its next session
  (`aiport_my_issues`).
- Its next failing call of the same kind (same status, error, name and
  module) names the Issue in `X-Aiport-Hint`.
- Reporting another failure of that kind adds it to the same Issue.
- The Issue resolves after the agent says it fixed it and the next call with
  that name and module succeeds. Another failure of that kind reopens it.

aiport cannot talk to a coding agent session that is already running. To have
aiport run a coding agent on the Issue now, open it in Changes and use **Fix
it**.

#### The live gauge

The gauge at the top of every page shows, live, whether aiport answers and
which Applications are calling. It reads request metadata only.

- **Listening** is green while the gateway answers the dashboard. It reads
  **No reply** when it does not, and **Signed out** when your session has
  ended.
- One light per Application: green when it called in the last two minutes,
  red when its call in the last two minutes failed, faint when it called in
  the last hour.
- A level meter lights up with each call as it passes, red at the top for a
  failure.

#### Turning request history on

Set `telemetry.enabled: true`. aiport then records request metadata in a
private SQLite file next to the configuration (`aiport.yaml.db`, mode 0600).

- Recording never slows down calls to providers.
- If records are dropped or fail to save, the banner on Overview and Activity
  says how many.
- If the database cannot be opened, aiport says so and keeps the latest 500
  records in memory for that session.

#### How long history is kept

Rows older than `telemetry.retention_days` (default 7) are deleted at startup
and then every hour.

The same setting decides how long these are kept:

- decided Suggestions and Change requests;
- Catalog changes;
- resolved or dismissed Issues (dismissed ones at least 7 days, the quiet
  period).

Open Issues are kept until they resolve. Applied Changes are never deleted.

Deleted rows can linger in SQLite's files: deletion is not secure erasure.

#### Request ids

Every response carries `X-Aiport-Request-Id`. If an application logs it, you
can find the exact call in Activity. Votes refer to it too.

#### What was asked and what served it

For each model call, Activity keeps both sides:

- **What the Application asked for**: the name it sent (a Level, a task name
  or a model id), its effort and where the effort came from (the
  `X-Aiport-Effort` header, the provider's effort field, or neither, which
  means `medium`), and its module.
- **What served it**: the Connection and model aiport sent it to.

The Model column shows the name sent with the result beneath it, for example
`balanced` with `chat/completions → small-model · low` under it. **Details**
says it in a sentence, for example "Asked for balanced at low effort (from
the X-Aiport-Effort header); served by hosted with small-model."

The **Name** filter shows only calls that sent one name. Use it to check that
a Level or a per-effort mapping works, and to compare the models behind it.

Good to know:

- A refused call has no serving model.
- Older records, from before this was kept, show none of it.

#### Why a request failed

For a failed request, Activity shows the reason next to the status, for
example `503 · Refused: no Connection for the Anthropic protocol`.
**Details** shows the code, what it means, what to do, and a link to the page
that fixes it.

aiport keeps only short codes for this: its own refusal code, a connection
error class, and the provider's error type (such as `rate_limit_error`) when
the provider sends one. It never keeps the provider's error message, the
request or the response.

All reasons and their fixes are in [Activity reasons and their
fixes](#activity-reasons-and-their-fixes).

### Changes

Changes is where aiport adapts to the Applications that use it, always with
your approval: it holds open Issues, what waits for your decision, and every
configuration Change made through aiport, with Revert. It stays quiet; a
badge appears in the Monitor menu only when something needs a decision.

The panels, top to bottom:

- **Issues**: open Issues (problems aiport keeps seeing for one Application)
  with their evidence, Explain, Fix it, and the fix on each side. See
  [Issues](#issues).
- **Suggestions and Change requests**: what waits for your decision, with the
  **Waiting / Accepted / Rejected** filter. See [What waits for your
  decision](#what-waits-for-your-decision).
- **Asked for, not supported yet**: kinds of Change that Applications asked
  for and aiport cannot apply, with **Suggest to aiport**. See [Suggest to
  aiport and Share upstream](#suggest-to-aiport-and-share-upstream).
- **Extensions**: the Extensions (your own programs that judge Change
  requests) listed in `aiport.yaml`, with **Share upstream**. See [Add
  Extensions](#add-extensions).
- **Applied changes**: every Change, newest first, with **Revert**, the
  automatic fixes switch (**Fix Issues automatically when the fix is clear**)
  and the sound. See [Applied changes and
  Revert](#applied-changes-and-revert).

What the badge counts:

- an Issue, only while its fix waits for you;
- commands a Fix it run waits for you to approve.

### The Integrate page and the two briefs

Integrate is where you connect an application to aiport: it holds your
Applications, the onboarding wizard, and the instructions aiport writes for
the coding agents that do the work. The **Instructions** panel is at the top
(collapsed unless you have no Applications), with the Applications list and
the wizard below it.

The two briefs are written by aiport from the running gateway, with its real
address and the protocols it has configured. They are served at `/_connect`:

- the **Application brief**, for the coding agent wiring the application to
  aiport;
- the **Coding agent brief**, for coding agents working in that application's
  repository later.

Together with the communication standard, they are everything a coding agent
needs.

#### Applications

The Applications list shows every Application (an app or agent that calls
aiport with its own token), what it does through aiport, and what to do next.
Pairing requests waiting for your approval are always at the top.

**Pairing requests.** Each shows its code, **Approve** (where you can also
bind a Connection and allow change requests) and **Reject**. See
[Applications and pairing](#applications-and-pairing).

**Columns**, left to right:

- name and client id;
- **Connection**, or the protocol default;
- **Changes**: the **May request changes** toggle;
- **Last call**, in the last 7 days (when activity recording is on);
- **Issues**;
- **Guidance**;
- **Agent**;
- **Onboarding** progress, such as `3/6`;
- **New token** and the trash button.

**Agent** names the Application's coding agent once it has one and a
repository path, and shows **Connect** otherwise. Select it to change them in
step 6 of the wizard.

**Issues** shows how many open or resolving Issues the Application has, with
a short cause such as "Unknown model llama3:8b" (or None). Select it to list
them under the row. Each Issue has its summary and:

- **Explain**;
- **Open in Changes**, which opens that Issue on the Changes page, where you
  approve or dismiss it;
- **Copy fix for the Coding agent**.

**Guidance** shows whether the Application's coding agent follows aiport's
current guidance (see [Guidance versions](#guidance-versions)):

- **Up to date**;
- **Behind v1 → v4**: the version it confirmed, and the current one;
- **Unknown**: it never confirmed one.

While it is not up to date, the copy button next to it copies the refresh
instruction: one paragraph to paste to that coding agent.

Under Guidance you may also see:

- "Sends Levels": a hint that its code may already use Levels, because its
  calls sent Levels in the last 6 hours. It never counts as confirming.
- The **Modules** and **Votes** chips: whether its calls in the last 7 days
  named their module, and whether its users' votes arrive. When one is
  missing, hand its coding agent the communication standard.

**Under the Application's name:**

- **Catalog** shows every name the Application can send, with the Connection
  and model that serve it (as on the Routing map), the models its Connections
  list, and its recent Catalog changes, newest first. These are the same
  changes the Application reads at `/v1/aiport/catalog`.
- **Levels**, next to Catalog, shows its Levels grid (`fast`, `balanced`,
  `best`), with its own mappings marked **Own**. See [Map Levels in the
  dashboard](#map-levels-in-the-dashboard).
- Click the name itself to open **Issue new token** and **Remove
  application** under its row. See [Give a token yourself, or remove an
  Application](#give-a-token-yourself-or-remove-an-application).

**Above the list:**

- The **Approved / Rejected / All** filter picks what the list shows.
  Rejected Applications are hidden until you choose **Rejected** or **All**;
  each shows when it was rejected, with **Forget**.
- **Onboard a new application** makes the wizard follow a new one. Selecting
  an existing Application makes the wizard follow that one.
- **Add with a token** adds an Application without pairing.

#### The instructions

You do not have to write instructions for the coding agent doing the work:
aiport writes them itself, from this gateway's real address and configured
protocols, and serves them at `/_connect`. The **Instructions** panel holds
the communication standard card and the two brief tiles.

**Communication standard** (`/_connect?format=standard`). The card shows the
prompt that hands the standard to a coding agent, with **Copy instruction**,
**Copy URL only** and **Read the standard**. See [The communication
standard](#the-communication-standard).

**Application brief** (`/_connect?format=handover&for=application`). For the
coding agent changing an application so its LLM calls go through aiport. Its
**Python / TypeScript / curl** picker switches the SDK snippets. It covers:

- pairing;
- keeping the token in `.aiport` (`AIPORT_BASE_URL`, `AIPORT_APP_TOKEN`), and
  what to do when it is lost;
- pointing the OpenAI or Anthropic SDK at aiport;
- removing provider keys from the application;
- model names: send a Level or a task name, never a provider's model id; read
  `GET /v1/models` for the names; `level_unmapped`;
- saying how hard a call is with `X-Aiport-Effort`;
- naming the module with `X-Aiport-Module`, and sending votes;
- watching the `X-Aiport-Catalog` header;
- errors and retries;
- logging `X-Aiport-Request-Id`;
- Issues: what they are, that the coding agent reads them with
  `aiport_my_issues`, that a Change that takes something away opens a
  `catalog_changed` one, and that aiport may fix routing itself as a
  revertable Change;
- how to verify.

**Coding agent brief** (`/_connect?format=handover&for=agent`). For coding
agents working in that application's repository later. It covers:

- the project MCP entry, and the line for the repository's `AGENTS.md`;
- keeping the guidance current;
- the MCP tools and their HTTP equivalents;
- checking the Catalog, Issues and guidance at the start of a session;
- saying when an Issue is fixed;
- naming the need (a Level or task name, and an effort) rather than a
  provider;
- how to write a good Change request, and waiting for your decision;
- the permission;
- hints.

**Using a tile.** Each tile shows who the brief is for, **Copy whole brief**
(the brief as markdown, to paste to the coding agent) and its address. Open a
tile to see its steps and snippets, each with its own **Copy**. Snippets for
a protocol this gateway has not configured are marked "not configured".

**For scripts:** `/_connect?format=handover` returns both briefs, one after
the other. `/_connect?format=markdown` holds everything, starting with
**Start here**.

### Connections and keys

A Connection is a provider aiport can send calls to, and Providers is the page
where you add Connections, paste their keys and check they work. Applications
never see provider keys: they send their own aiport token, and aiport adds the
provider key on the way out.

aiport supports OpenAI, Anthropic, Amazon Bedrock, Meta Model API, and local
servers such as Ollama or llama.cpp. Each Connection speaks one protocol,
`openai` or `anthropic`. Each provider shows its logo on its card, in the add
list and in the Overview's provider distribution; llama.cpp and custom
providers show a neutral icon.

Connections are grouped as **Remote APIs** and **Local**. Both groups start
collapsed. Each heading shows how many Connections it holds and which need a
key or are unhealthy; click it to open the group.

**Add or remove.** **Add a connection** from a preset, or any other server
that speaks the OpenAI or Anthropic API as is (remote URLs must use HTTPS).
**Remove** takes one away (see below). Each is a Change you can revert in
Changes.

**Paste a provider API key** into a remote Connection's card.

- It applies at once.
- It is stored only in the private `aiport.yaml.env`, never in `aiport.yaml`.
- It is never shown again.

You can also set the key in the shell aiport starts from, under the
Connection's key variable (see [configuration](configuration.md)). Provider
keys are never asked for in the terminal. Local Connections need no key.

**Check connection** (or **Check all** for a group) sends one authenticated
`GET /models` to the provider. It generates nothing and costs nothing. It
shows whether the provider is reachable and accepted the key. A failed check
says why, and what to do:

- a connection error: host name does not resolve, connection refused, TLS
  failed, or timed out;
- or the HTTP status with the provider's own error type, for example
  `HTTP 401, authentication_error`.

A passing check does not prove every model works.

**Show models** lists the models that Connection serves, from the same free
`GET /models` listing (OpenAI-, Anthropic- and Ollama-shaped answers). After a
check, the button shows the count ("Show 12 models").

- aiport keeps only the model ids, in memory, and reuses a list for two
  minutes. **Refresh** lists again, and so does **Check connection**.
- If the listing fails (no key, key rejected, no model list, unreachable),
  aiport treats that Connection's models as unknown, not as "serves nothing".
- Disabled Connections are not listed.

**Make OpenAI default** / **Make Anthropic default** sets the Connection a
protocol uses when nothing else picks one.

Provider presets, endpoints and key variables are in
[configuration](configuration.md).

#### Removing a Connection, and when one fails

You can always remove a Connection: aiport moves everything that used it to
another Connection in the same step.

1. Select **Remove** on the Connection's card. A dialog lists everything still
   using it: a protocol default, an Application bound to it, an Alias (shared
   or an Application's own), Explanations, the model calls of Agent runs, and
   the compression evaluator.
2. Check the Connection picker on each item. It is already set to the best
   match:
   1. a Connection of the same protocol that lists the same model (for a
      binding or a protocol default: the one that lists most of the removed
      Connection's models);
   2. otherwise the protocol default (the new one, if you are removing the
      default);
   3. otherwise none.
3. Check the list of affected Applications below (see next).
4. Confirm.

**Choosing None** leaves an item without a Connection on purpose:

- a protocol default is unset;
- a binding is removed, and the Application follows the protocol defaults
  again;
- an Alias is removed, and its task name is retired;
- Explanations, Agent runs' model calls or the evaluator are turned off
  (Agent runs then use the coding agent's own login).

**Who is affected.** Below the list, the dialog names the Applications whose
Catalog (the names they can send) the removal changes: task names removed or
moved, or a binding or protocol default they follow. It updates as you change
a picker. aiport itself is among them when the removal moves what its own
model calls (Explanations and Agent runs) use: `agent_runs.connection`, or the
protocol defaults when that is not set. aiport's Catalog changes are kept like
any Application's, but it gets no Issue, because it has no coding agent to
tell.

**One Change.** Confirming applies all the moves and the removal as **one
Change**, so one revert in Changes brings back the Connection and everything
that used it. Each affected Application is told through its Catalog. An
Application that lost something it used gets a Catalog-change Issue (see [A
Catalog change](#a-catalog-change)).

**When a Connection fails.** There is no automatic fallback. When a
Connection is down or keeps answering with errors, calls to it fail and the
Application gets the provider's own error. When this keeps happening for one
Application, aiport opens an Issue that offers to bind that Application to a
working Connection, and you approve it (see [A failing
Connection](#a-failing-connection)). aiport never moves an Application to
another Connection by itself, so calls never go where you did not choose.

**For scripts:**

- Without the dialog (a `DELETE /api/v1/targets/{id}` with no plan), aiport
  refuses to remove a Connection that something uses (Explanations and Agent
  runs' model calls included). The message names what uses it.
- The dialog reads `GET /api/v1/targets/{id}/dependents` and sends the plan as
  `{"plan": {"<item>": "<Connection id or empty>"}}`.

### Routing and protocol defaults

Routing decides which Connection and model serve each call, and the Routing
page is where you set the protocol defaults and the Levels, and see where
every name goes. aiport speaks each provider's own API and does not translate
between them: an OpenAI SDK call needs a Connection that speaks the `openai`
protocol, and an Anthropic SDK call needs one that speaks `anthropic`.

The page, top to bottom:

- **Protocol defaults**: the Connection each protocol uses when nothing else
  picks one.
- The **Levels** panel: what `fast`, `balanced` and `best` mean here.
- The **Routing map**: where every name goes.
- **Connect an application**: the two base URLs Applications use
  (`http://127.0.0.1:8787/v1` for the OpenAI protocol, `http://127.0.0.1:8787`
  for the Anthropic protocol), the `/_connect` protocol description, and the
  telemetry queue.
- The **Applications** table.
- A pointer to the instructions on Integrate.

Model allowlists are planned, not built.

#### How aiport picks a Connection and model

For each call, aiport goes down this list and uses the first rule that
applies:

1. **A module route**: the Application has a route for the module the call
   names (`X-Aiport-Module`) and the name it sends. See [Apply a Suggestion as
   a module route](#apply-a-suggestion-as-a-module-route).
2. **A Level** (`fast`, `balanced`, `best`): the Application's own mapping on
   this protocol, else the shared one. A Level mapped on neither is refused
   with `400 level_unmapped`.
3. **The Application's own Alias** (a task name it alone uses) for the
   requested name, if it has one. It may name another Connection.
4. **A shared task name marked `everywhere`**, on its own Connection.
5. **The Connection the Application is bound to**, if any, else **the protocol
   default** set under Protocol defaults, with the shared Aliases of that
   Connection.

If the Level or Alias has per-effort models, aiport then picks the model for
the call's effort (see [Effort and reasoning](#effort-and-reasoning)). Only
the `model` field is replaced, plus the reasoning effort in the cases that
section describes.

When no Connection can serve the call:

| Situation | Answer |
| --- | --- |
| The protocol has no Connection | `503 target_unconfigured` |
| The Connection is disabled | `503 target_disabled` |
| The Connection speaks the other protocol | `400 incompatible_protocol` |

#### The Routing map

The Routing map answers "if this Application sends this name, which Connection
and model serve it?" for every Application at once. It is built from each
Application's Catalog, the same list its `GET /v1/models` answers, so the two
never disagree.

For each Application, the map lists every name it can send:

- its own task names;
- the shared task names of the Connections it reaches, and the `everywhere`
  ones;
- the Levels mapped for it (its own mapping or the shared one, per protocol);
- **Any other model name**, which goes unchanged to its binding or to the
  protocol default (with the number of models that Connection lists).

Each row shows the protocol, the Connection with its provider's logo, and the
model aiport sends to the provider, with its per-effort models under it when
it has any. For example, docs-bot sending `fast` on the OpenAI protocol
reaches the local Connection with `llama3:8b`.

Choose an Application or a provider to narrow the map.

Good to know:

- The same view for one Application, with its recent Catalog changes, is
  under **Catalog** on its row on Integrate.
- The pencil on a task name's row edits its per-effort models (see [Per-effort
  models of a task name](#per-effort-models-of-a-task-name)). A Level's
  per-effort models are edited in the Levels panel.

**For scripts:** read the map with the management token at
`GET /api/v1/routing` (`?application=<id>` for one Application, with its
`changes`). It holds Application ids, names, task names, model names and
Connection ids only, never keys or addresses.

### Agents

The Agents page lists the coding agents whose own model calls go through
aiport, each with its own token, so they never see your provider keys. For
each, it shows the Application with its Connection (or the protocol default)
and the line that launches it.

To set one up and launch it:

1. Create the Application with `aiport init --agent codex,claude`. The agent
   needs an Application whose id is `codex` or `claude`. When none exists,
   the page tells you to run this.
2. In a terminal, run `aiport agent codex --model MODEL` or
   `aiport agent claude --model MODEL`.

Good to know:

- Add `--dry-run` to see what would start.
- Put options for the agent after `--`, as in `-- agent-options`.
- The settings apply to that launch only.
- Credentials are never displayed.
- This is separate from the Agent runs (aiport running a coding agent on an
  Issue) that [Connecting a coding agent](#connecting-a-coding-agent)
  describes.

### Compression

The Compression page shows whether opt-in request compression saves anything
and whether it makes answers worse, for each combination of methods.
Compression only runs on requests that ask for it with `X-Aiport-Compress`;
see [Compress requests](#compress-requests).

The page shows the savings and a trend of compressed against uncompressed
traffic (feedback, error rate, size saved). Token savings are estimated from
bytes, divided by 4.

Each method combination gets a verdict:

| Verdict | When |
| --- | --- |
| **Not enough data** | Fewer than 20 requests on either side. |
| **Possible harm** | Feedback more than 0.1 lower, or an error rate more than 5 points higher, than uncompressed traffic. |
| **No harm seen** | Otherwise. |

**The evaluator.** The **Evaluator** panel picks a Connection and model for
model-based checks. **Test evaluator** runs two known cases and reports
whether the model labels them correctly. A Connection in use as the evaluator
cannot be removed until you choose another one (or none), except through the
removal dialog.

Details are in [compression](compression.md) and [compression
evaluators](compression-evaluators.md).

### Feedback

The Feedback page shows how the people using each application rate its
answers, per module, model and effort, so you can see which part of which
application does well and move the part that does not. Applications send the
votes; [Collect feedback and improve
routing](#collect-feedback-and-improve-routing) explains how.

Pick the last 7, 30 or 90 days (30 by default).

**Per Application**, the page shows its thumbs up and down, its share of
up-votes, and how many of its calls were rated.

**Its table** lists each module (a named part of the application, sent as
`X-Aiport-Module`). Inside each module, one row per combination of name sent
(a Level, task name or model id), model and effort, with calls, votes and the
share of up-votes as a bar:

- green from 75%;
- amber from 50%;
- red below 50%.

A module or combination with fewer than 5 votes is marked "few votes": too
few to judge. Calls that named no module are grouped as "No module named";
guidance version 4 asks the application's coding agent to add
`X-Aiport-Module`.

**Best and weakest.** With two or more judged modules, the page names the
best and the weakest module. Try another model or effort for the weakest
under Routing, and watch its share.

**Suggestions from the votes** lists what aiport found in the votes. Each one
is marked **Better elsewhere**, **One Level up** or **Cheaper may do**, shows
the move (from which model to which) and the reason, and has an **Apply**
button. See [Suggestions from the votes](#suggestions-from-the-votes).

Each Application's panel also lists its **Module routes**, each with
**Remove**.

When there is nothing to show:

- With request history off, the page says no calls or votes are kept.
- With calls but no votes, it says applications start sending them once
  their coding agent follows guidance version 4.

**For scripts:** the same report is `GET /api/v1/feedback?hours=N` (1 to
2160, default 720), with `applications`, `suggestions` and `module_routes`.

### Tasks

Tasks is a queue of work for coding agents: you write down tasks and decide
their intent and priority, and coding agents pick them up by name. Each task
gets a memorable name, such as `task-amber-fox`, that you can hand to a coding
agent; click the name to copy it.

A task has:

- a title, up to 200 characters;
- instructions;
- acceptance criteria;
- a priority: None, Low, Medium or High;
- a status: pending, in progress, blocked, completed, failed or cancelled.

Its history keeps who changed what, and when.

The filters above the queue show the open tasks (pending, in progress or
blocked; the default), all of them, or one status. The chime in the toolbar
plays when no pending or in-progress tasks remain.

The queue is stored next to the configuration, in `aiport.yaml.tasks.json`.
From a terminal (`--priority` takes 0 to 3, None to High):

```sh
aiport task list [--all]
aiport task show NAME
aiport task set NAME --status completed --priority 2 --note "tests pass"
```

### Settings

Settings holds everything about this aiport install that is not routing:
sign-in, the default coding agent for Agent runs, who writes Issue
explanations, and the configuration file.

- **Authentication**: the username and password you sign in with, and whether
  authentication is on. If you forget them, use **Forgot username or
  password?** on the sign-in screen with the management token, then set a new
  password here.
- **Default coding agent**: the coding agent of every Application that has
  none of its own. The same panel holds **Model calls of Agent runs**: the
  Anthropic-protocol Connection (and optional model) that those runs' calls go
  through. See [Connecting a coding agent](#connecting-a-coding-agent) and
  [Explain by a Coding agent](#explain-by-a-coding-agent).
- **Explanations**: the Connection and model that explain an Issue when no
  coding agent does. See [Explain](#explain).
- **Configuration**: one file, `aiport.yaml`. Providers, routing, Levels and
  per-effort models and each Application's coding agent are managed from their
  pages; task names arrive as approved Changes or are edited, with the listen
  address, in the file. Validate hand edits with `aiport check`, then restart.

## Reference

Lookup pages: the exact commands, endpoints, headers, error codes and settings,
for when you already know what you are looking for.

### Command line

These are the `aiport` commands you type in a terminal. Every command takes
`--config PATH` to use a different configuration file (default `aiport.yaml`).

| Command | What it does |
| --- | --- |
| `aiport init` | First-run setup. Writes `aiport.yaml` and `aiport.yaml.env`; never overwrites them. Flags below. |
| `aiport check` | Checks the configuration. Lists Connections with no API key and Applications with no token, and warns once for each Alias named like a Level. |
| `aiport serve` | Starts aiport. Prints the dashboard address and, in a terminal, the management token. |
| `aiport auth token` | Prints the management token again. |
| `aiport agent codex`, `aiport agent claude` | Starts that coding agent through aiport, with its own token. Needs `--model MODEL`. |
| `aiport task list`, `show NAME`, `set NAME` | Reads and changes the task queue. |
| `aiport token` | Prints a new random token, for example for a token variable you set yourself. |
| `aiport version` | Prints the version. |

**`aiport init` flags.** It asks you to choose providers, coding agents and
Levels (the names `fast`, `balanced`, `best` that apps use instead of a model
name). You can answer with flags instead:

- `--provider`: comma-separated, from `openai`, `anthropic`, `bedrock`,
  `meta`, `ollama`, `llamacpp`.
- `--agent codex,claude`, or `--agent none`.
- `--listen HOST:PORT` (default `127.0.0.1:8787`).
- `--region` and `--base-url`.
- `--levels LEVEL=CONNECTION/MODEL,...`, or `--no-levels` to skip them.

**Other flags.**

- `aiport agent`: `--dry-run` shows what would start without starting it.
  Anything after `--` goes to the agent.
- `aiport task list --all` includes closed tasks.
- `aiport task set` takes `--status`, `--priority 0-3`, `--note` and
  `--actor`.

### Endpoints applications use

These are the addresses an Application (an app or script you connected to
aiport) calls. They are all on aiport's address, `http://127.0.0.1:8787` by
default.

Every endpoint needs the Application's own token as
`Authorization: Bearer <token>`. Anthropic SDKs may send it as `x-api-key`
instead. Pairing and `/_connect` need no token.

| Endpoint | What it is |
| --- | --- |
| `POST /v1/chat/completions`, `POST /v1/responses` | Model calls in the OpenAI protocol, sent to the provider as they are. |
| `POST /v1/messages` | Model calls in the Anthropic protocol, sent to the provider as they are. |
| `GET /v1/models` | The Application's Catalog (the model names it may use), in the SDK's own list format. |
| `GET /_connect` | How to connect, as JSON. `?format=markdown` gives the guide, `?format=standard` the communication standard, `?format=handover&for=application` or `&for=agent` a brief. No token. |
| `POST /v1/aiport/pair` | Asks to pair. Send `{name, description?, repository?}`; you get a `code` and a `poll` path. No token. |
| `GET /v1/aiport/pair/{secret}` | The `poll` path. Its `status` is `pending`, `approved` (with the `token`, shown once), `rejected` or `expired`. No token. |
| `GET /v1/aiport/capabilities` | What the Application can use and ask for right now. |
| `GET /v1/aiport/catalog` | The Catalog and its recent changes. |
| `POST /v1/aiport/feedback` | A vote on an answer: `{request_id, vote}` or `{request_id, score}`, with an optional `module`. |
| `GET /v1/aiport/guidance`, `POST /v1/aiport/guidance/done` | Guidance changes since the version last confirmed; confirm the current one. |
| `POST /v1/aiport/requests`, `GET /v1/aiport/requests` (or `/{id}`) | File a Change request (ask the Owner to change a setting); read the answers. |
| `GET /v1/aiport/issues` (or `/{id}`), `POST /v1/aiport/issues/{id}/fixed` | The Application's Issues (problems aiport found); say one is fixed. |
| `POST /mcp` | The MCP endpoint for the Application's coding agent. Streamable HTTP, JSON-RPC sent as `application/json`. |

### Owner API

This is the API the dashboard itself uses, under `/api/v1/`. Your own scripts
can call it too, to do anything the dashboard does.

**Signing in.** Send the management token as `Authorization: Bearer <token>`.
The dashboard uses its sign-in session instead. Application tokens never work
here.

**The CSRF header.** Every request that changes something must also send
`X-Aiport-CSRF: 1`. So must the two requests that ask providers for their
model lists (`.../models` and `/serving`).

Every edit here that changes the configuration is recorded as one Change you
can revert.

| Endpoint | What it does |
| --- | --- |
| `GET /api/v1/auth/status` | Whether a password is set, sign-in is on, and this browser is signed in. |
| `POST /api/v1/auth/login`, `POST /api/v1/auth/logout` | Sign in with `{username, password}`; sign out. |
| `POST /api/v1/auth/recover` | Starts a session with the management token. Used by **Forgot username or password?**. |
| `POST /api/v1/auth/settings` | Sets `{username, password, disabled}`. |
| `GET /api/v1/health` | aiport is up, and the state of request history. |
| `GET /api/v1/config` | The configuration as the dashboard sees it, with which keys are set (never the keys themselves). |
| `GET /api/v1/requests` (or `/{id}`), `GET /api/v1/stats` | Request history, one record, and the summary on Overview. Filters: `hours` (1 to 168), `origin`, `target`, `provider`, `outcome`, `model`, `from_ms`/`to_ms`, `limit` (1 to 100), `cursor`. |
| `POST /api/v1/targets` | Adds a Connection: `{id, provider, ...}`. |
| `PATCH /api/v1/targets/{id}`, `DELETE /api/v1/targets/{id}` | Edits a Connection's `region`, `endpoint` or `protocol`; removes one. Send `{"plan": {...}}` with the delete to re-point what used it. |
| `PUT /api/v1/targets/{id}/api-key`, `DELETE .../api-key` | Saves a provider key `{api_key}` to the private credential file; removes it. |
| `POST /api/v1/targets/{id}/check` | **Check connection**: one signed-in `GET /models` to the provider. |
| `GET /api/v1/targets/{id}/models` | **Show models**: the model ids the Connection lists. |
| `GET /api/v1/targets/{id}/dependents`, `POST .../dependents` | What uses a Connection, with suggested replacements. `POST {plan}` previews a plan. |
| `GET /api/v1/serving?protocol=P&model=M` | Which Connections of a protocol list a model. `&refresh=1` asks again. |
| `PUT /api/v1/defaults` | Sets the protocol defaults: `{openai_target, anthropic_target}`. |
| `GET /api/v1/pairings` | Pairing requests waiting. `?status=rejected` lists the remembered rejected ones. |
| `POST /api/v1/pairings/{code}/approve`, `.../reject` | Approves or rejects a pairing request. Approve optionally takes `{id, target, may_request_changes, agent, repository, replace}`. |
| `DELETE /api/v1/pairings/rejected/{id}` | **Forget** a rejected Application. |
| `POST /api/v1/origins` | Adds an Application with a token: `{name, id?, target?, may_request_changes?, agent?, repository?}`. |
| `PATCH /api/v1/origins/{id}` | Sets an Application's `may_request_changes`, `agent` or `repository`. |
| `POST /api/v1/origins/{id}/token` | Issues a new token for an Application. The old one stops working. |
| `DELETE /api/v1/origins/{id}` | Removes an Application. |
| `GET /api/v1/routing` | The Routing map. `?application=<id>` for one Application, with its `changes`. |
| `GET /api/v1/levels` | Every Level mapping, shared and per Application, and the Connections. |
| `PUT /api/v1/levels/{protocol}/{level}` | Maps a Level: `{application?, connection, model, effort?}`. |
| `DELETE /api/v1/levels/{protocol}/{level}` | Clears a mapping: `{application?}`. |
| `PUT /api/v1/aliases/effort` | Sets the per-effort models of any Alias: `{application?, alias, connection?, effort}`. |
| `PUT /api/v1/task-names/{name}` | Sets a shared task name: `{connection, model, effort?, everywhere?, reasoning?}`. |
| `DELETE /api/v1/task-names/{name}` | Removes a shared task name: `{connection}`. |
| `PUT /api/v1/module-routes/{application}/{module}/{name}` | Sets a module route: `{connection, model, why?}`. |
| `DELETE /api/v1/module-routes/{application}/{module}/{name}` | Removes a module route. |
| `GET /api/v1/feedback?hours=N` | The votes report, Suggestions from the votes, and module routes. |
| `GET /api/v1/compression?hours=N` | The compression report. 1 to 2160 hours, default 168. |
| `PUT /api/v1/evaluation`, `POST /api/v1/evaluation/test` | Chooses the compression evaluator `{connection, model}`; **Test evaluator**. |
| `GET /api/v1/guidance` | Each Application's guidance state, with `behind` and `never_confirmed`. |
| `GET /api/v1/changes`, `POST /api/v1/changes/{id}/revert` | Every Change, newest first; **Revert** one. |
| `GET /api/v1/decisions?status=S`, `POST /api/v1/decisions/{id}` | Suggestions and Change requests (`pending`, `accepted` or `rejected`); decide one with `{action: accept, reject or dismiss, note?, ...}`. |
| `GET /api/v1/issues` | Issues, filtered by `?state=` and `?application=`, with `fix_by` per Application. |
| `POST /api/v1/issues` | Reports a failed request to its Application's coding agent: `{request_id, note?}`. |
| `POST /api/v1/issues/{id}` | Acts on an Issue, for example `{"action": "reject"}`. |
| `POST /api/v1/issues/{id}/explain` | Streams an explanation as JSON lines. |
| `PUT /api/v1/adapting/explainer` | Chooses who writes explanations: `{connection, model}`. |
| `PUT /api/v1/adapting/auto-fix` | Turns automatic fixes on or off: `{enabled}`. |
| `POST /api/v1/issues/{id}/fix`, `GET /api/v1/issues/{id}/fix` | Starts **Fix it**; reads its `branch_fix` and `diff`. |
| `POST /api/v1/issues/{id}/fix/keep`, `.../fix/discard` | Keeps or discards a Fix it diff. |
| `GET /api/v1/issues/{id}/run` | The Issue's last Agent run, its log and approvals waiting. |
| `GET /api/v1/runs`, `POST /api/v1/runs/{id}/stop` | The Agent runs going now; stop one. |
| `GET /api/v1/approvals`, `POST /api/v1/approvals/{id}` | Commands waiting for you; answer one with `{answer: allow_once, allow_run or deny}`. |
| `GET /api/v1/agents`, `PUT /api/v1/agents/calls` | The coding agents found; where Agent runs send their model calls. |
| `PUT /api/v1/agents/default` | The default coding agent: `{agent}` (`claude`, `codex`, `gemini` or empty). |
| `GET /api/v1/upstream`, `POST /api/v1/upstream/issues` | Drafts for **Suggest to aiport** and **Share upstream**; file one after you confirm. |
| `GET /api/v1/tasks`, `POST /api/v1/tasks` | The task queue; add a task. |
| `GET /api/v1/tasks/{name}`, `PATCH /api/v1/tasks/{name}` | One task; change it. `X-Aiport-Actor` names who made the change. |

### Headers

These are the extra HTTP headers aiport reads on requests and adds to
responses. aiport removes every `X-Aiport-*` request header before it sends the
call on to the provider.

| Header | Direction | What it means |
| --- | --- | --- |
| `Authorization: Bearer <token>` or `x-api-key` | Request | The Application's own token. aiport swaps it for the provider key before sending the call on. |
| `X-Aiport-Module` | Request | Which part of the app the call is for. Required on every model call. Lower-case `a`–`z`, `0`–`9`, `.`, `_`, `-`; at most 64 characters. |
| `X-Aiport-Effort` | Request | `low`, `medium` or `high`: how hard the call is. Picks a per-effort model. If the body sets no reasoning effort, aiport also sends this to the provider as the reasoning effort. |
| `X-Aiport-Compress` | Request | `whitespace`, `json`, `dedupe` or `all` (comma-separated), or `off`. |
| `X-Aiport-CSRF: 1` | Request | Required on Owner API requests that change something. |
| `X-Aiport-Actor` | Request | Task API only (`/api/v1/tasks`): a short name for who made the change, kept in the task's history. Default `owner`. |
| `X-Aiport-Request-Id` | Response | aiport's id for the call. Keep it with the answer, for votes and logs. |
| `X-Aiport-Model` | Response | On every call sent on: the model that actually answered, after a Level, task name, per-effort model or module route replaced the name sent. A vote for this request id counts for this model. |
| `X-Aiport-Route` | Response | On every call sent on: `connection=…; name=…; effort=…; via=…`. `via` is `level`, `task_name`, `module_route` or `model`. |
| `X-Aiport-Catalog` | Response | The version of the Application's Catalog (12 hex characters). |
| `X-Aiport-Compression` | Response | The compression methods used, the `status` (`applied`, `no_gain` or `failed`), and the size in bytes before and after. |
| `X-Aiport-Hint` | Response | One line of help: the open Issue a failing call belongs to, the Change request that could fix it, or what to do about a refused token. |
| `X-Aiport-Connect` | Response | On `/_connect`: the version of the connection description (now `1`). |

**Good to know:** `/_connect` also lists `X-Aiport-Task`, `X-Aiport-Session`
and `X-Aiport-Version` (requests) and `X-Aiport-Source` (responses). They are
planned, not built yet.

### Error codes

When aiport refuses or fails a call, the body is
`{"error": {"code", "message", "request_id", ...}}`. Any other error body comes
from the provider; aiport passes it through unchanged, with its status and
headers.

The status and code never depend on the optional `hint`.

#### When aiport refuses an Application's token

A refused token gets either `401 unauthorized` (on a model or Application
endpoint) or `403 application_token_required`. Both answers carry an
`error.fix` block, written for the application's coding agent, and an
`X-Aiport-Hint` header that says the same in one line:

```json
{"error": {"code": "unauthorized", "message": "...", "request_id": "...",
  "fix": {"for": "the application's coding agent",
          "steps": ["Check that .aiport exists at the repository root ...", "If .aiport is missing ..., ask the Owner ... New token ..."],
          "retry": "Once .aiport has the token, retry the same call ... Claude Code reads the MCP token when it connects: run /mcp and reconnect aiport, or start a new session.",
          "docs": "http://127.0.0.1:8787/_connect?format=markdown#if-aiport-refuses-the-token"}}}
```

The steps tell the coding agent to:

1. Check that `.aiport` has the `AIPORT_APP_TOKEN=` line. An
   `AIPORT_APP_TOKEN` set in the environment wins over the file, so unset an
   old one.
2. If the token is missing, lost or revoked, ask the Owner to click
   **New token** on Integrate (**Issue token** when there is none yet), and
   paste the copied lines into `.aiport`.
3. Not pair again unless the Owner asks, and not look for a token anywhere
   else.
4. Retry the same call. If the application reads `.aiport` only at startup,
   restart it first.

#### Codes for applications

| Status and code | Meaning |
| --- | --- |
| `401 unauthorized` | The Application token is missing or wrong. Follow `error.fix`. |
| `403 application_token_required` | This endpoint serves one Application, and no Application token was recognized. The management token does not count. Follow `error.fix`. |
| `403 invalid_host`, `403 invalid_origin`, `403 cross_site_request` | The `Host` is not the listen address or a configured hostname, or a browser sent a cross-origin or cross-site request. All are refused. |
| `403 run_credential` | An Agent run's credential was used for something other than its model calls. |
| `404 unsupported_path`, `405 method_not_allowed` | Not part of aiport's protocol, or the wrong HTTP method. |
| `400 incompatible_protocol` | The chosen Connection speaks the other protocol. aiport does not translate between them. |
| `400 level_unmapped` | The Level is not mapped on this protocol, neither for this Application nor shared. The message names both. |
| `400 ambiguous_model` | aiport needs to replace the model name, so the body must be one JSON object with one `model` field. |
| `400 request_read_failed` | aiport could not read the request body (the Application stopped sending). Retry. |
| `400 unknown_compression` | `X-Aiport-Compress` names a method aiport does not offer. |
| `413 request_too_large` | The body is over the size limit (`proxy.max_request_bytes`). |
| `429 concurrency_limit` | Too many requests at once (`proxy.max_concurrent`). Retry with backoff. |
| `503 target_unconfigured` | No Connection for this protocol or Application. |
| `503 target_disabled` | The chosen Connection is turned off. |
| `503 target_invalid` | The chosen Connection's settings (its base URL) cannot be used. |
| `503 upstream_key_missing` | The Owner has not set the provider key yet. |
| `502 upstream_unavailable` | aiport could not connect to the provider. |
| `502 upstream_request_failed` | aiport could not build the request to the Connection's address. |
| `504 upstream_timeout` | The provider did not answer in time. |
| `499 client_cancelled` | The Application hung up first. Recorded in Activity, never sent. |
| `400 invalid_format` | `/_connect` got a `format` other than `json`, `markdown`, `handover` or `standard`, or a `for` other than `application` or `agent`. |
| `415 unsupported_media_type`, `400 unsupported_protocol_version` | `/mcp` takes JSON-RPC as `application/json`, in an MCP protocol version aiport supports. |
| `400 invalid_pairing`, `404 pairing_not_found`, `429 pairing_busy` | Pairing needs a name of 1 to 128 characters; the request expired or never existed; too many are waiting, retry after a minute. |
| `400 invalid_feedback`, `404 request_not_found` | The vote is malformed; or aiport has no kept request with that id for this Application. |
| `403 change_requests_not_allowed` | The Owner has not allowed this Application to request changes. |
| `400 invalid_change_request` | The request is incomplete or too large. `error.problems` lists every problem. |
| `403 change_request_refused` | A policy Extension (a program you added to check requests) refused it. `error.reason` says why. |
| `429 change_requests_rate_limited` | More than 20 requests in the last hour. Retry after `Retry-After`. |
| `404 change_request_not_found`, `404 issue_not_found`, `409 issue_not_open` | Not this Application's request or Issue, or the Issue is no longer open. |
| `409 adapting_off` | The adapting loop is turned off in the configuration. |
| `500 change_request_failed`, `500 issue_update_failed`, `500 guidance_not_saved` | aiport could not save the Change request, the Issue or the guidance confirmation. Retry. |

The full list is at `http://127.0.0.1:8787/_connect?format=markdown#errors`.

#### Codes from the Owner API

The Owner API answers with the same error body. The common codes:

- `401 unauthorized`: sign in again.
- `403 csrf_required`: send `X-Aiport-CSRF: 1`.
- `503 config_readonly`: aiport cannot write its configuration or credential
  file.
- `409 revert_conflict` and `409 already_reverted`: from **Revert**.
- `409 token_from_environment`: the token is set in the environment, so aiport
  cannot replace it.
- A `404` whose code ends in `_not_found`: the id is unknown.
- A `400` whose code starts with `invalid_`: the message says what to send.

### Configuration

All of aiport's settings live in one file, `aiport.yaml`.
[configuration](configuration.md) explains every key, and
[aiport.example.yaml](../aiport.example.yaml) is a starting point to copy.

When you change something in the dashboard, aiport writes it into the file in
place, keeps your other keys and comments, and applies it at once. If you edit
the file by hand, run `aiport check` and then restart aiport.

The top-level keys:

| Key | What it sets |
| --- | --- |
| `version` | The configuration format version. |
| `listen`, `hostnames` | The loopback address aiport listens on, and extra `*.localhost` names for the browser. |
| `auth` | Dashboard sign-in: `username`, `password_hash` (set it from Settings, never by hand), `disabled`, and `token_env`, the variable that holds the management token. |
| `targets` | The Connections (providers aiport can call). |
| `defaults` | The protocol defaults. |
| `aliases` | Shared Level mappings and task names (`effort`, `everywhere`, `reasoning`). |
| `origins` | The Applications: `token_env`, `target`, `aliases`, `may_request_changes`, `agent`, `repository`, `module_routes`. |
| `telemetry.enabled`, `telemetry.retention_days` | Request history on or off, and how long it, decided items and Catalog changes are kept (default 7 days). |
| other `telemetry` keys | `queue_capacity`, `batch_size`, `flush_interval_ms`, `shutdown_drain_ms`, `max_event_bytes` tune the recorder. |
| `adapting` | `disabled`, `auto_fix`, `explain`, the thresholds `unknown_model`, `failing_connection`, `level_unmapped` and `unknown_token`, and `interval_seconds`. |
| `agent_runs` | `default_agent`, `connection`, `model`, `explain_minutes`, `fix_minutes`. |
| `evaluation` | The compression evaluator's Connection and model. |
| `extensions` | Policy and guardrail programs. |
| `upstream.repository` | Where **Suggest to aiport** and **Share upstream** open drafts. |
| `proxy` | Limits: `max_concurrent`, `max_request_bytes`, `connect_timeout_ms`, `request_read_timeout_ms`, `response_header_timeout_ms`, `stream_idle_timeout_ms`. |

### Data and privacy

aiport keeps a record of each call, such as which model answered and how long
it took. It never keeps what you asked or what the model answered. Everything
stays on your machine.

#### The files

aiport keeps its files next to the configuration file:

| File | What it holds |
| --- | --- |
| `aiport.yaml` | The configuration. Never holds a key or token. |
| `aiport.yaml.env` | The private credentials, readable only by you (mode 0600): provider keys you paste, Application tokens, the management token. |
| `aiport.yaml.db` | A private SQLite database (mode 0600): request history, Changes, decisions, Issues, Catalog changes, guidance confirmations and rejected pairings. |
| `aiport.yaml.tasks.json` | The task queue (mode 0600). |
| `aiport.yaml.worktrees/` | The working copies Agent runs use while they run. |

#### What aiport records about each call

For each call aiport records the time, the route, the Application, the model
name it sent, the effort and where that effort came from, the module, the
Connection, the model that answered, the status, how long it took, short error
codes, the compression sizes and the vote.

aiport never records prompts, responses, request or response headers, or the
provider's error messages. It never stores keys or tokens there either: tokens
live only in the credential file.

Issues, Catalog changes and explanations hold only ids, model names, counts,
request ids and fixed advice text.

#### How long it is kept

Request history, decided Suggestions and Change requests, Catalog changes,
and resolved or dismissed Issues are kept for `telemetry.retention_days`
(7 days by default). Open Issues are kept until they are resolved.

Some things are kept longer than that: applied Changes, and Issues that hold a
fix you reverted or rejected.

Some things are never written to disk. Model lists stay in memory for two
minutes. Agent run logs stay in memory while their Issue is open.

When aiport deletes something, it is not securely erased from the disk.

### Security

aiport is built to be safe to run on your own development machine. It is
still experimental. To report a vulnerability privately, see
[SECURITY.md](../SECURITY.md).

**Only your machine can reach it.** aiport listens on `127.0.0.1:8787` by
default, which other computers cannot reach. It only answers requests
addressed to its own listen address or to the `*.localhost` names you
configured, and it refuses browser requests that come from other sites.

**Apps never see your provider keys.** Each Application (an app you
connected) has its own aiport token. aiport swaps that token for the real
provider key when it sends the call on. Calls only ever go to the Connections
you set up.

**The dashboard needs a sign-in.** The first time, you sign in with the
management token, then set a password of at least 12 characters. A browser
session lasts 12 hours, or until aiport restarts. Requests that change
something also need the `X-Aiport-CSRF: 1` header, so another web page cannot
make them for you.

**Nothing changes without you.** Every Change is listed and can be reverted.
Automatic fixes are off by default, and even when on they can only change one
Application's Alias.

**Agent runs stay boxed in.** An Agent run (a coding agent aiport starts to
explain or fix an Issue) starts only when you click. It works in a separate
copy of the repository, with a minimal environment and no aiport tokens or
provider keys. Each command it wants to run waits for your answer. aiport
never commits to your branch, merges or pushes. Read
[What aiport cannot see](#explain-by-a-coding-agent) before you use them.

**Extensions are opt-in.** An Extension (a policy or guardrail program) runs
only if you list it with an absolute path. It has a timeout, and it refuses
unless it explicitly allows.

## Troubleshooting

Find what you see below, then follow the fix. For any failed request, open it
in Activity and click **Details**: it names the reason and links to the page
that fixes it. The exact reasons are listed in
[Activity reasons and their fixes](#activity-reasons-and-their-fixes).

#### The dashboard does not open

Either aiport is not running, or you are using a different address. Check that
`aiport serve` is running, and open the exact address it prints. aiport refuses
other host names and requests from other sites. To use a name instead of the
number, see [Use a name instead of 127.0.0.1](#use-a-name-instead-of-127001).

#### I forgot the password

On the sign-in screen, click **Forgot username or password?**. Paste the token
that `aiport auth token` prints. Then set a new password in Settings →
Authentication.

#### An application gets `401 unauthorized` or `403 application_token_required`

Its token is missing, lost, revoked or not loaded. The error has an
`error.fix` block that tells the app's coding agent what to do:

1. Check the token in `.aiport`.
2. If it is not there or does not work, click **New token** on the app's row
   on Integrate, and paste the lines into `.aiport`.
3. Restart the application if it reads `.aiport` only at startup.

Do not pair the app again.

#### A `401` never shows up in Activity

aiport cannot tell which Application sent it, so it has nowhere to list it.
Fix it as above. If the application never had a token, pair it on Integrate.

#### The aiport MCP server does not connect in Claude Code

The MCP entry read an old or missing token when it connected. Fix `.aiport`,
then run `/mcp` and reconnect aiport, or start a new session. A
`403 application_token_required` means no Application token was recognized.

#### An SDK, script or coding agent cannot reach `aiport.localhost`

Its resolver does not know the name, the name resolves to `::1`, or a sandbox
proxy refuses it. Use `http://127.0.0.1:8787` instead; aiport always hands
that address out. For a script, you can also add `127.0.0.1  aiport.localhost`
to `/etc/hosts`.

#### Calls show as `unknown` in Activity

The application is not sending its own token, for example it uses a shared or
old one. Pair it, or give it a token with **Add with a token**.

#### Calls fail with `400 level_unmapped`

The Level (a name like `fast` that stands for a model) is not mapped on that
protocol. Go to Routing → Levels and map it, or give the Application its own
mapping on its Integrate row. If it keeps happening, aiport opens an Issue.

#### Calls fail with `503 target_unconfigured`

There is no Connection (a provider aiport can call) for that protocol. Go to
Routing → Protocol defaults and pick one, or bind the Application to a
Connection. aiport does not translate between protocols.

#### The Feedback page shows no votes

One of three things:

- Request history is off. Turn on `telemetry.enabled`.
- The application's coding agent has not followed guidance version 4. Check
  the **Modules** and **Votes** chips on Integrate, and hand the coding agent
  the communication standard.
- Nobody has voted yet.

#### A vote gets `404 request_not_found`

The request id is not this Application's, the call is no longer kept, or
history is off. Send the `X-Aiport-Request-Id` of that answer, soon after the
call.

#### No Suggestions come from the votes

aiport only suggests a change when the evidence is clear. It needs at least 10
votes on the same module, name, model and effort, and the votes must not sit
between 60% and 90% thumbs up. Wait for more votes.

#### An application gets `403 change_requests_not_allowed`

You have not allowed that Application to request changes. Allow it on
Integrate if you want its requests.

#### An application gets `409 adapting_off`

The adapting loop is turned off in the configuration. If you want it back,
remove `adapting.disabled: true` and restart.

#### Explain is greyed out

No Connection is chosen for explanations, or adapting is off. Go to Settings →
Explanations and choose a Connection and model.

#### Revert refuses with a conflict

The configuration changed after that Change. Revert the newer Changes first.

#### Activity is empty

Request history is off, or it is not being saved. Check that
`telemetry.enabled: true` is set, and that the banner does not say history is
kept for this session only.

### Activity reasons and their fixes

When a request fails, Activity's **Details** shows a reason in plain words.
Find that text in the left column to see what to do.

| Activity says | Code | What to do |
| --- | --- | --- |
| Refused: no Connection for the Anthropic (or OpenAI) protocol | `503 target_unconfigured` | Routing → Protocol defaults: pick a Connection for that protocol, or bind the Application to one on Integrate. aiport does not translate between protocols. |
| Refused: the Connection is disabled | `503 target_disabled` | Turn it on in Providers, or pick another default on Routing. |
| Refused: the Connection's settings are invalid | `503 target_invalid` | Correct its base URL in Providers. |
| Refused: the Connection does not speak the … protocol | `400 incompatible_protocol` | Routing → Protocol defaults: pick a Connection of that protocol. |
| Refused: the Level is not mapped for the … protocol | `400 level_unmapped` | Routing → Levels: map the Level on that protocol, or give the Application its own mapping on its Integrate row. If it keeps happening, aiport opens an Issue (see [An unmapped Level](#an-unmapped-level)). |
| Refused: no API key for the Connection | `503 upstream_key_missing` | Paste the key on the Connection's card in Providers. |
| Refused: too many requests at once | `429 concurrency_limit` | Retry. If it keeps happening, raise `proxy.max_concurrent`. |
| Refused: the request is too large | `413 request_too_large` | Send less, or raise `proxy.max_request_bytes`. |
| Refused: the request body could not be read | `400 request_read_failed` | The Application stopped sending. Retry. |
| Refused: unknown compression method | `400 unknown_compression` | Use a method listed on the Compression page. |
| Refused: the model name is ambiguous | `400 ambiguous_model` | Send one JSON object with one `model` field. |
| Gateway error: could not build the provider request | `502 upstream_request_failed` | Check the Connection's base URL in Providers. |
| Cannot reach the provider: its host name does not resolve | `502 upstream_unavailable`, `dns_failure` | Check the base URL in Providers, and the network. |
| Cannot reach the provider: connection refused | `502 upstream_unavailable`, `connection_refused` | Start the local server (Ollama, llama.cpp) or fix the port, then **Check connection**. |
| Cannot reach the provider: the secure connection (TLS) failed | `502 upstream_unavailable`, `tls_error` | Check the scheme (`http` or `https`) and host name in the base URL. |
| Cannot reach the provider: connecting timed out, or the connection was closed | `502 upstream_unavailable`, `timeout` or `connection_reset` | Check the network and the base URL, then retry. |
| Timed out: the provider did not answer in time | `504 upstream_timeout` | Retry. For slow local models, raise `proxy.response_header_timeout_ms`. |
| Provider error: the key was rejected | provider `401` or `403` (`authentication_error`, `invalid_api_key`, `permission_error`) | Replace the key in Providers, then **Check connection**. |
| Provider error: unknown model | provider `404` (`not_found_error`, `model_not_found`) | Use a model the provider lists (the Application sees them with `GET /v1/models`), or add an Alias. If it keeps happening, aiport opens an Issue in Changes that says which Connection serves the model. |
| Provider error: rate-limited | provider `429` (`rate_limit_error`) | Wait, or ask the provider for a higher limit. |
| Provider error: out of quota or credit | `insufficient_quota`, `billing_error` | Check your plan and billing with the provider. |
| Provider error: the request was not accepted | provider `400` (`invalid_request_error`) | Check the model name and parameters the Application sends. |
| Provider error: the request is too large | provider `413` (`request_too_large`) | Send a smaller request or a shorter context. |
| Provider error: the provider failed | provider `5xx` (`overloaded_error`, `api_error`) | Retry later, and check the provider's status page. |
| The response broke off | `incomplete` | Retry. For long answers, raise `proxy.stream_idle_timeout_ms`. |
| Cancelled by the Application | `499 client_cancelled` | Nothing to fix. |

## What's new

Newest first. The commit history has every detail.

**2026-11-04**

- Every call aiport sends on now says which model answered
  (`X-Aiport-Model`) and how aiport chose it (`X-Aiport-Route`).
- Activity details open right under their row. You can report a failed
  request to the Application's coding agent as an Issue.
- A live gauge at the top of every page: Listening, one light per
  Application, and a meter of calls as they pass.
- The Catalog marks task names that route everywhere (`everywhere: true`).
- This manual was rewritten: quickstart, concepts, guides, a tour of the
  dashboard, reference and troubleshooting.
- Votes and Suggestions per module. Apps name the module of each call with
  `X-Aiport-Module`, and send a thumbs up or down per answer to
  `/v1/aiport/feedback`. The new [Feedback](#feedback) page shows the votes
  per Application, module, model and effort, and turns them into Suggestions
  you apply as module routes.
- The communication standard, at `/_connect?format=standard` and as the MCP
  tool `aiport_standard`: every rule, plus a checklist. Guidance version 4
  asks coding agents to follow it and to add the thumbs to the app's screens.
- Integrate shows **Modules** and **Votes** chips for each Application, in the
  Guidance column.
- Task names can apply everywhere and carry a fixed reasoning effort. New
  Owner API for them: `/api/v1/task-names`.
- `X-Aiport-Effort` on its own now also sets the provider's reasoning effort,
  when the body sets none and aiport chose the model.
- When aiport refuses a token, it tells the coding agent what to do, with an
  `error.fix` block and an `X-Aiport-Hint` header.

**2026-11-03**

- Levels (`fast`, `balanced`, `best`): mapped per protocol and per
  Application, and edited in the dashboard. `init` proposes a mapping. An
  Issue opens for unmapped Levels.
- Effort picks the model, with per-effort models for Levels and task names.
  Activity records the name, effort and model of each call.
- The Catalog: `GET /v1/models` per Application, `X-Aiport-Catalog`,
  `/v1/aiport/catalog`, Issues when the Catalog changes, and removing a
  Connection with a plan to re-point what used it.
- The Routing map: every Application's names and what serves them.
- Guidance versions, so coding agents know when to refresh. Integrate shows
  who is behind.
- Tokens live in `.aiport`. You can issue a token yourself, add an Application
  without pairing, or remove one. Pairing survives restarts and repeated
  requests.
- Applications and agents are given `127.0.0.1`, not the hostname.
- The module path is `github.com/harelix/aiport`, so `go install` works.

**2026-11-02**

- Issues from start to finish: unknown models and failing Connections,
  **Copy fix for the Coding agent**, automatic fixes if you allow them, and
  the Application's coding agent hearing about its Issues.
- Explain an Issue with a Connection, or with a coding agent that reads the
  code. **Fix it** on a branch, with a diff and approval for each command.
- The Integrate page, with the onboarding wizard and both briefs. Also:
  pairing, Change requests, Extensions, **Suggest to aiport** and **Share
  upstream**, and the MCP endpoint.
