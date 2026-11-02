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
An Application that writes code, such as Codex or Claude Code, usually started
with `aiport agent`.

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

## Adapting aiport

**Change request**
An Application asking for aiport to work differently for it, or for every
Application, with what it needs, why, and an argument for that scope. The
Owner approves or rejects it.

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
click: to Explain it (read-only) or to Fix it (in its own branch, ending in a
diff the Owner keeps or discards), with its log, approvals and result.

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
