# Extensions

An **Extension** is a program the Owner adds that takes part in deciding about
Change requests. aiport ships none. There are two hook points:

| Hook | Runs | A deny… |
| --- | --- | --- |
| `policy` | when a Change request arrives, after aiport's own checks (permission, rate limit, complete fields) | refuses the request; the requesting Application gets the reason |
| `guardrail` | when a Suggestion or Change request is approved (by the Owner, through an Issue's aiport fix, or automatically with `adapting.auto_fix`), before the Change is applied | blocks Approve; nothing changes, the item still waits, and the dashboard shows the reason |

Reject and Dismiss are not Changes and run no Extension.

## Listing Extensions

Only programs listed in `aiport.yaml` ever run. Nothing an Application sends,
and nothing in the dashboard or the API, can add one. A Change request asking
for an Extension is kept as `not_supported`.

```yaml
extensions:
  - name: ticket-policy           # unique; letters, digits, - and _
    hook: policy                  # policy or guardrail
    path: /usr/local/libexec/aiport/ticket-policy   # absolute path to the executable
    args: ["--strict"]            # optional, passed as given
    timeout_ms: 2000              # optional; default 5000, at most 60000
  - name: freeze
    hook: guardrail
    path: /usr/local/libexec/aiport/freeze
```

If several Extensions are listed for one hook, they run in the listed order
and the first deny wins.

## How aiport runs an Extension

- It starts `path` directly with `args`. There is no shell and no `PATH` lookup,
  so `path` must be absolute. The working directory is the program's own
  directory.
- The environment holds only `PATH`, `HOME`, `TMPDIR`, `LANG` and `LC_ALL`
  (Windows: `SystemRoot`, `PATH`, `TEMP`, `TMP`, `USERPROFILE`, `PATHEXT`).
  Provider keys and tokens in aiport's environment never reach it.
- It writes one JSON object (the input) to standard input and closes it.
- It reads one JSON object (the output) from standard output, at most 64 KiB.
  Standard error is ignored.
- After `timeout_ms` the Extension and every process it started are stopped.

**Anything but a clean exit with `{"decision":"allow"}` denies:** a deny, a
timeout, a crash, a non-zero exit status, output that is not JSON, too much
output, or a decision other than `allow` or `deny`. aiport then gives its own
reason, such as `no decision within 2000 ms` or `the Extension failed (exit
status 3)`. A broken Extension never lets anything through.

## Input

Every input has `protocol` (currently `1`), `hook` and `extension` (the listed
name). Text fields come from Applications as plain text: treat them as data,
never as commands, and never pass them to a shell.

A policy gets `request`:

```json
{
  "protocol": 1,
  "hook": "policy",
  "extension": "ticket-policy",
  "request": {
    "application": "helper",
    "kind": "application_alias",
    "supported": true,
    "need": "An alias fast for a small local model",
    "details": "Requests for model fast should go to llama3 on the local Connection.",
    "why": "The nightly summary job is cheap.",
    "scope": "application",
    "scope_case": "Only this job needs it.",
    "filed_by": "nightly agent",
    "alias": "fast",
    "prefer_connection": "local",
    "prefer_model": "llama3"
  }
}
```

`alias`, `prefer_connection` and `prefer_model` are present only when the
Application sent them: its task name and its advice on what should serve it.
`supported` is false for a kind aiport cannot apply (it would be kept as
`not_supported`); a policy can refuse those too.

A guardrail gets `change`, with the Scope the Owner granted and the Change as
it will be applied:

```json
{
  "protocol": 1,
  "hook": "guardrail",
  "extension": "freeze",
  "change": {
    "item": "82cb9bfa9ff2d630",
    "source": "request",
    "application": "helper",
    "kind": "application_alias",
    "scope": "every_application",
    "what": "shared alias fast → local/llama3 for every Application",
    "alias": "fast",
    "target": "local",
    "model": "llama3",
    "by": "Owner"
  }
}
```

`source` is `request` (a Change request) or `suggestion`. `by` is `Owner`,
or `aiport (automatic)` for an automatic fix. For
`bind_application`, `target` is the Connection; `alias` and `model` are empty.

## Output

```json
{"decision": "allow"}
{"decision": "deny", "reason": "no new aliases during the freeze"}
```

`reason` is shown as plain text, on one line, at most 500 characters. A deny
without a reason reads `no reason given`.

## What callers see

A policy refusal of `POST /v1/aiport/requests` (and the MCP tool
`aiport_request_change`) is `403` with code `change_request_refused`; the
error has the `extension` name and its `reason`, and the message reads
`refused by policy ticket-policy: …`. The request is not kept and does not
count against the hourly rate limit.

A guardrail block of `POST /api/v1/decisions/{id}` (or `POST
/api/v1/issues/{id}` for an Issue's aiport fix) is `409` with code
`guardrail_blocked`, with `extension` and `reason`. The configuration is not
touched, no Change is recorded, and the item stays pending.

## A minimal policy

```sh
#!/bin/sh
# Refuse bindings for every Application; allow everything else.
in=$(cat)
case "$in" in
  *'"kind":"bind_application"'*'"scope":"every_application"'*)
    echo '{"decision":"deny","reason":"ask for your own Application only"}' ;;
  *) echo '{"decision":"allow"}' ;;
esac
```

A real Extension should parse the JSON (for example with `jq` or a small
program) rather than match text.

## Sharing an Extension upstream

The Changes page lists the Extensions in `aiport.yaml` with **Share upstream**.
It prefills a GitHub contribution for aiport that names only the hook and the
input protocol, never the Extension's name, path, arguments or timeout; you
paste the program yourself after removing anything private. aiport opens it in
your browser, or runs `gh issue create` after you confirm. See
[Upstream](configuration.md#upstream).
