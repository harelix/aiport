# Compatibility: experimental preview

No provider/client combination has completed the live support gate. There is no
supporting release yet. Configuration presets are not compatibility evidence.

| Connection | Implemented contract | Evidence / remaining limits |
| --- | --- | --- |
| OpenAI | Native Chat and Responses POST byte forwarding; bearer key | Deterministic native fixtures only; no live smoke |
| Anthropic | Native Messages POST byte forwarding; x-api-key upstream; version/beta pass-through | Deterministic native fixtures only; no live smoke |
| Meta Model API | OpenAI-protocol preset at `https://api.meta.ai/v1`, `MODEL_API_KEY` | No live model/stream/tools validation |
| Bedrock | `provider: bedrock` + `region`, `endpoint: runtime|mantle`, `protocol: openai|anthropic`; `AWS_BEARER_TOKEN_BEDROCK` | API-key mode only; no SigV4, profiles, Converse or InvokeModel; no live region/model validation; a few Mantle OpenAI models need the `/openai/v1` prefix via `base_url` |
| Ollama | OpenAI-protocol preset at `http://127.0.0.1:11434/v1` | No server/model versions tested; endpoint subset must be checked |
| llama.cpp | OpenAI-protocol preset at `http://127.0.0.1:8080/v1` | No server/model versions tested; endpoint subset must be checked |
| Codex CLI | Per-launch Responses provider; separate local token, no client retry, no WebSocket | Installed version observed: 0.166.0; launch-config tests only; no live session |
| Claude Code | Per-launch Messages base URL and bearer local token | Installed version observed: 2.1.316; launch-config tests only; no live session |
| Local CLI/config/handler | Built/run CLI on macOS arm64 | In-memory HTTP tests; sandbox denies socket fixtures |
| SQLite persistence | Request metadata, decisions and Catalog changes in `<config>.db` (pure-Go `modernc.org/sqlite`, WAL, 0600) | Retention pruning and migration tests; session-only fallback when the database cannot be opened |
| Embedded dashboard/session activity | Preview implemented | Typecheck/build and API tests; browser QA unavailable |
| Other operating systems | Planned | No runtime/platform release testing |

Implemented native paths are exactly POST `/v1/chat/completions`, `/v1/responses`,
`/v1/messages`. `GET /v1/models` is answered by aiport itself from the calling
Application's Catalog, not forwarded. Files, response retrieval/compaction,
batches and WebSockets are unavailable. Agents using additional endpoints can fail even when
a basic generation works. A provider may reject unsupported models/fields/routes.
There is no protocol translation, automatic retry, fallback or usage observation.

Fixture evidence: `internal/proxy/proxy_test.go` covers byte preservation, query
semantics, native errors, tool-fragment/heartbeat stream bytes, flushing before
completion, aliases, origin routing, credential/header stripping, overload, size,
cancellation and idle/aborted streams. This does not establish live TLS, phase
network deadlines, platform packaging or performance. No latency guarantee or
benchmark has been measured.

## Primary documentation inspected on 2026-11-01

- [OpenAI Chat](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create)
- [OpenAI Responses](https://developers.openai.com/api/docs/guides/migrate-to-responses)
- [Codex provider configuration](https://learn.chatgpt.com/docs/config-file/config-reference)
- [Anthropic Messages](https://platform.claude.com/docs/en/api/messages/create)
- [Claude Code gateway connection](https://code.claude.com/docs/en/llm-gateway-connect)
- [Meta Model API overview](https://dev.meta.ai/docs/overview)
- [Bedrock OpenAI-compatible Responses endpoints/auth](https://docs.aws.amazon.com/bedrock/latest/userguide/inference-responses-api.html)
- [Ollama OpenAI compatibility](https://docs.ollama.com/api/openai-compatibility)
- [llama.cpp server](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md)

## Dashboard preview update

Embedded React dashboard and authenticated session Activity/Stats/config APIs now
exist. Request history is kept in SQLite (see above), and Connections, defaults,
Levels and other settings are edited from the dashboard as revertable Changes.
Frontend typecheck/build and deterministic API/queue tests are evidence only;
browser visual/interaction checks could not run because the browser host was
unavailable. No shipped native/live compatibility claim changes from this UI work.
