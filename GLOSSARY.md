# Glossary

## People and programs

**Owner**
The person who runs aiport on their machine and decides what it does: which
connections exist, which Applications may use it, and which changes are made.

**Application**
Anything on this machine that calls aiport for LLM work with its own token: a
script, a service, a desktop app or a Coding agent. Every call is attributed to
the Application that sent it.
_Avoid_: client, origin (origin is only the configuration key).

**Coding agent**
A program that writes code, such as Claude Code, Codex or Gemini CLI. It may
work on an Application's repository: as the Application's Coding agent it
hears about the Application's Issues, and aiport may start it for an Agent run.
When its own model calls go through aiport (for example started with
`aiport agent`), it is also an Application.

**Pairing**
How an Application gets its own token: it asks aiport, shows the Owner a short
code, and receives the token once the Owner approves that code.

## Routing

**Connection**
A model provider aiport can forward to, local or remote, with its own address
and key.
_Avoid_: target (the configuration key), provider (the kind of service behind
a Connection).

**Alias**
A model name an Application sends that aiport replaces with a real model on a
Connection. An Application can have its own Aliases; they come before the
shared ones.

**Task name**
An Alias named for what the Application does, such as `review-code` or
`summarize`, rather than for a model. The Application sends the task name;
the Owner decides which Connection and model serve it.
_Avoid_: model alias, nickname.

**Level**
A capability name every aiport shares (`fast`, `balanced`, `best`) that the
Owner maps, per protocol, to a model on a Connection. It is a reserved Alias:
an Application's own mapping comes before the shared one.
_Avoid_: tier, `<provider>-<tier>` names.

**Effort**
How hard one call is (`low`, `medium`, `high`), stated by the client (the
`X-Aiport-Effort` header, else its provider's own effort field, else
`medium`). It picks the model when a Level or task name has a model per
effort, and an effort stated only with the header is also written into the
provider's own effort field when aiport chose the model. A task name may fix
its reasoning effort instead. aiport never changes an effort already in the
request.

**Module**
The part of an Application a call serves (`tldr`, `classify`, `insights`),
sent with the `X-Aiport-Module` header. Votes are read per module, and a
module route can send one module's calls for one name to another model.

**Catalog**
What one Application can use through aiport: its task names (its own Aliases
and the shared ones) and the models listed by the Connections it reaches, with
a version. The Application reads it at `GET /v1/models` and sees the version
on every answer.
_Avoid_: model list (that is one Connection's).

**Catalog change**
The difference one Change made to one Application's Catalog: task names
removed, added or moved, and routing that changed. One that takes away
something the Application used is told to it as an Issue.

## Adapting aiport

**Change request**
An Application asking for aiport to work differently for it, or for every
Application, with what it needs, why, and an argument for that scope. It
may advise which provider and model should serve it; the Owner decides,
approves or rejects it.

**Suggestion**
aiport proposing a change itself, from patterns in its request history, with
the evidence that led to it.

**Issue**
A problem aiport keeps seeing for one Application, with its evidence and the
ways to resolve it: a fix in aiport (offered as a Suggestion) and a fix in
the Application, for its Coding agent. An Issue closes itself when the
Application's calls succeed again.

**Agent run**
One run of a Coding agent that aiport starts for an Issue, on the Owner's
click: to Explain it (read-only, in a throwaway worktree) or to Fix it (in its
own worktree on a new branch, ending in a diff the Owner keeps or discards),
with its log, approvals and result.

**Change**
An approved Change request or Suggestion once it is applied to aiport's
configuration. Every Change can be reverted. aiport applies a Change on its
own only when the Owner allowed automatic fixes, and a reverted automatic
fix is never applied automatically again.

**Scope**
Who a Change affects: only the requesting Application, or every Application.
The requester argues for a scope; the Owner decides it.

**Extension**
A program the Owner adds that takes part in deciding about Change requests,
for example a policy that refuses some kinds or a guardrail that checks a
Change before it is applied. Extensions are not part of aiport's core.

**Guidance version**
The number of the guidance aiport gives coding agents (briefs, the AGENTS.md
line, MCP instructions), raised when agents must act on a change, with a
short change log. aiport keeps the version each Application's coding agent
last confirmed; an Application is behind while that is older (or unknown).
