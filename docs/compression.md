# Request compression

Applications can ask aiport to shrink a request before it reaches the provider,
and aiport measures whether that hurts the answers. Compression is **off unless
the application asks for it**, per request, on any chat, responses or messages
call to any connection.

## Turning it on

```http
POST /v1/chat/completions
X-Aiport-Compress: whitespace, json
```

| Method | What it does | Meaning |
| --- | --- | --- |
| `whitespace` | Removes trailing spaces, normalises line endings, collapses 3+ blank lines to one. Indentation is kept. | Preserved |
| `json` | Rewrites text that is entirely JSON (typical tool results) compactly. | Preserved |
| `dedupe` | When a block of 1 KiB or more repeats, earlier copies become a short note; the latest copy stays. | Lossy |

`all` selects every method; `off`, `none` or no header means none. An unknown method
returns `400 unknown_compression`.

Only message text is changed (`content`, `text`, `system`, `output`, `input`).
Tool definitions, schemas, `tool_choice`, `response_format` and metadata are never
touched. If compression fails or saves nothing, the original bytes are sent.

Every response to a request that asked for compression carries
`X-Aiport-Compression: whitespace,json; status=applied; bytes_before=…; bytes_after=…`,
where `status` is `applied`, `no_gain` or `failed` (on the last two the
original body was sent).

## Measuring whether it hurts

For each request aiport records the methods, status, request size before and
after, outcome and latency. Requests without the header are the baseline.

Applications can score the answer they got, using the response's
`X-Aiport-Request-Id`:

```http
POST /v1/aiport/feedback
{"request_id": "…", "score": 0.9}
```

Scores run from 0 (useless) to 1 (perfect). A person's thumbs can be sent as
`{"request_id": "…", "vote": "up"}` (or `"down"`), which counts as 1 (or 0), with
an optional `module`. An application can only score its own requests.

The dashboard's **Compression** page (left menu, Configure) shows savings, a trend of
compressed vs uncompressed traffic (feedback, error rate, size saved) and a verdict
per method combination:

- **Not enough data:** fewer than 20 requests on either side.
- **Possible harm:** feedback is more than 0.1 lower (with at least 10 scored
  requests on each side), or the error rate is more
  than 5 points higher, than uncompressed traffic. Errors are counted the same
  way as on Overview: upstream, gateway, incomplete and timed-out requests;
  cancellations and requests aiport refused are not errors.
- **No harm seen:** neither of those.

Token savings are estimated from bytes (÷4) until provider-reported usage is
recorded. Prompts and responses are never stored. Results are saved with the rest
of the request history in SQLite and kept for `telemetry.retention_days`.

## Evaluator

Model-based checks (NLI, claim checks, LLM judge) run on a connection you choose:
any configured provider, local or remote. Set it on the Compression page
(Evaluator) or in YAML:

```yaml
evaluation:
  connection: ollama-local   # any connection id
  model: llama3.1
```

Requests use that connection's own protocol and key at temperature 0. **Test
evaluator** runs two known NLI cases (one entailment, one contradiction) and
reports whether the model labels them correctly; a model that fails is too weak
to judge. A remote evaluator receives question and answer text and bills your
key; a local one keeps everything on your machine. A connection in use as the
evaluator cannot be removed until another (or none) is chosen.
