# Routing alias naming review

> Design note, 4 November 2026. A review of a proposal for routing aliases,
> written while Levels and task names were being designed. The names aiport
> uses today (`fast`, `balanced`, `best`, task names, `X-Aiport-Effort`) are
> in the [manual](../manual.md); where this note differs, the manual wins.

## Verdict

Adopt most of the routing-alias proposal: task and tier as separate inputs, the renames, the column names, keeping protocol separate from model choice, and filtering before optimizing. Change four parts before locking it in. The most important is the word "tier", which already means something else in both upstream APIs aiport translates.

## Adopt as written

- **Task and tier as separate inputs.** This matches the `task` + `level` split in the aiport intent envelope. "Fast extraction" and "high-quality summarization" are the real use cases.
- **Renames:** `tldr` → `summarize`, `best` → `quality`, "Shared task name" → "Task alias". Keep the old names as compatibility aliases.
- **Column renames**, including "Client API protocol" vs "Upstream provider" and "Upstream model ID".
- **"Explicit model route"** instead of "protocol default", with a clear error for unknown IDs.
- **Filter on capabilities first, then optimize for the tier using measured performance.** A task alias states intent but doesn't prove which model is best, which is why the thumbs feedback loop exists.

## Change before locking in

### 1. Call the axis "profile", not "tier"

Three different axes hide under fast/quality:

| Axis | What it controls | Where it already exists |
| --- | --- | --- |
| Which model | fast / balanced / quality model choice | aiport (this proposal) |
| How hard one model thinks | reasoning budget within a single model | OpenAI `reasoning_effort`, Anthropic `effort` |
| Capacity class | pricing and latency priority, not quality | `service_tier` in both OpenAI and Anthropic |

A client could send `service_tier: "flex"` together with `routing.tier: "fast"`, and those mean opposite things. Name the aiport axis **`profile`** (`profile/fast`, `profile/balanced`, `profile/quality`). Pass `reasoning_effort` / `effort` and `service_tier` through as separate native fields, so a client can ask for "Sonnet-class model, low effort".

### 2. Let one model string carry both axes

With task in `model` and tier in an extension field, clients that can't add fields need a header. A client sending `model: "tier/fast"` plus `routing.tier: "quality"` also needs a precedence rule. OpenRouter already uses a suffix convention for routing variants (`:nitro`, `:floor`), so copy it: `task/extract:fast` works from any OpenAI or Anthropic SDK with no extra fields. If the extension field is also present, the field wins and the receipt shows the reason code `OVERRIDE:profile`.

### 3. Tighten the task list

- **`task/reason`** describes difficulty, not a kind of task. Express it as `profile/quality` plus `requires.reasoning`; otherwise "reason, fast" contradicts itself.
- **`task/generate`** will catch everything. Acceptable, but expect the learner to have little signal in that bucket.
- **`auto`** needs a classifier call on every request, which adds latency and cost. The receipt should state the inferred task, so thumbs feedback can separate a wrong inference from a wrong model. The existing classify route on Kimi K3 could serve as the classifier.
- **Tool use is a capability** (`requires.tools`), not a task. **Embeddings** are a separate operation and endpoint, not a task alias.

### 4. Keep fallback and escalation as two columns

The proposal merges "low / high alternatives" into one "Fallback / escalation policy" column. Escalation means the model was too weak or overkill: a learned quality judgment. Fallback means the provider failed: availability. Merged, a Sonnet timeout gets recorded as evidence that Sonnet is bad.

Minor: name the body extension `aiport` rather than `routing`. A generic key risks colliding with a field an upstream provider adds later.

## Revised routing table

The task-alias routes have no failure fallback yet; that chain needs to be defined.

| Routing alias | Route type | Candidates (primary first) | Escalation set | Failure fallback |
| --- | --- | --- | --- | --- |
| `task/classify`, `task/extract`, `task/summarize` | Task alias | `us.moonshotai.kimi-k3` (Bedrock) | — | Undefined: needs a chain |
| `profile/fast` | Profile | `claude-haiku-5-5`, `muse-spark-1.3` | — | The other candidate |
| `profile/balanced` | Profile | `claude-sonnet-5-5`, `muse-spark-1.3` | Down: Haiku · Up: Opus / `muse-wonder-1.0` | The other candidate |
| `profile/quality` | Profile | `claude-opus-5-5`, `muse-wonder-1.0` | Down: Sonnet / `muse-spark-1.3` | The other candidate |
| Any listed model ID | Explicit model route | As sent | None | None (pinned) |
| `auto` | Inferred | Classifier → task alias, then the requested profile | Per resolved route | Per resolved route |

## Open decision

- [ ] Is protocol-based model mapping deliberate?

Today the protocol decides the model: OpenAI-protocol `fast` goes to Muse Spark, Anthropic-protocol `fast` goes to Haiku. If that is deliberate (for example, so Anthropic-SDK clients consistently get Claude-style output), keep it as a soft **provider affinity** that ranks candidates, not a hard mapping. If not, drop it, and each profile uses the merged candidate list above, ranked by measurements and thumbs.
