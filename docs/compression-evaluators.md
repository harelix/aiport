# Online Compression Evaluators — Implementation Spec

## 1. Overview

Build four label-free evaluators that measure how much prompt/context compression degrades answer quality on live traffic: shadow agreement, proxy evidence recall, source faithfulness, and synthetic needle probes. All four write to one shared `EvalRecord` and are aggregated per slice.

**In scope**

- Capture of compression requests, sampling, an async shadow runner, the four scorers, the probe generator and runner, aggregation, and alert rules.
- Extractive compression (sentence/chunk selection), token pruning (LLMLingua family), and abstractive compression or compaction (summaries, rolling compaction).

**Non-goals**

- Offline golden-set evaluation, human labeling UI, and ship/block gating. These consume this system's output but are specified separately.
- Changing the compressor or the production answer path. Evaluation never blocks or alters the user-facing response.
- KV-cache compression inside the model (no text-level compressed context to inspect).

**Terminology**

| Term | Meaning |
| --- | --- |
| Full context (`C_full`) | The context that would be sent without compression |
| Compressed context (`C_comp`) | The context actually sent after compression |
| Full answer (`A_full`) | Model answer on `C_full`, produced by the shadow runner |
| Compressed answer (`A_comp`) | Model answer on `C_comp`; the production answer |
| Ratio | `tokens(C_full) / tokens(C_comp)` |
| Method class | `extractive`, `token_pruning`, or `abstractive` |
| Slice | A combination of tags (ratio bucket, method, content type, etc.) used for aggregation |
| Noise floor | Disagreement rate between two independent full-context runs; the baseline that compression disagreement is measured against |
| Claim | An atomic, independently checkable statement extracted from a text |
| Key entity | A number, date, ID, code identifier, file path, URL or proper name extracted from a text |

## 2. Architecture and shared data model

Evaluation runs entirely off the request path: production writes a capture record and returns; everything else is asynchronous workers reading from a queue.

### 2.1 Components

| Component | Responsibility | Runs |
| --- | --- | --- |
| Capture hook | Writes a `CaptureRecord` after every compressed request | In-process, non-blocking (fire-and-forget enqueue) |
| Sampler | Decides which captures get evaluated and by which scorers | Worker, on dequeue |
| Shadow runner | Produces `A_full` (and `A_full_2` for noise-floor samples) | Worker, rate-limited |
| Scorers | Four pure functions: `agreement`, `evidence`, `faithfulness`, `probe` | Worker |
| Probe generator | Builds needle probes from real documents | Scheduled job + on-deploy hook |
| Probe runner | Sends probes through the real compression + model path | Scheduled job |
| Aggregator | Rolls `EvalRecord`s up per slice and window | Scheduled job (e.g. every 15 min) |
| Alerter | Evaluates alert rules on aggregates | After each aggregation |

### 2.2 Data flow

1. Production request compresses `C_full` to `C_comp`, calls the model, returns `A_comp` to the user.
2. Capture hook enqueues a `CaptureRecord` (§2.3). Failure to enqueue is logged and dropped; it never fails the request.
3. Sampler selects the record or discards it (§8.1).
4. Shadow runner calls the same model with the same parameters on `C_full`, temperature 0, producing `A_full`. If the record is chosen for noise-floor estimation, it runs `C_full` a second time to produce `A_full_2`.
5. Applicable scorers run and write their results into one `EvalRecord`.
6. Aggregator computes slice metrics; alerter evaluates rules.
7. Records with `agreement.verdict != agree` (plus a small random sample of agreements) are exported to a review queue for the offline loop.

### 2.3 CaptureRecord

```
CaptureRecord {
  request_id: string            // unique, stable
  ts: timestamp
  app: string                   // calling application (origin id)
  query: string
  context_full_ref: string      // pointer to stored C_full (not inline)
  context_comp_ref: string      // pointer to stored C_comp
  context_full_hash: string     // sha256 of C_full
  method: string                // compressor name + version, e.g. "query_extract@1.4"
  method_class: "extractive" | "token_pruning" | "abstractive"
  compaction_step: int | null   // for rolling compaction, 0-based step index
  tokens_full: int
  tokens_comp: int
  ratio: float                  // tokens_full / tokens_comp
  model: string
  model_params: object          // temperature, max_tokens, system prompt version, etc.
  answer_comp: string
  answer_comp_citations: Citation[] | null
  slice_tags: SliceTags
  conversation_id: string | null  // groups compaction steps (§6.4)
  is_probe: bool                // true for synthetic probe traffic
  probe_id: string | null
}

Citation { doc_id: string, start: int, end: int }   // char offsets into C_full

SliceTags {
  content_type: "code" | "prose" | "table" | "mixed"
  doc_count: int
  hop_estimate: "single" | "multi" | "unknown"
  ratio_bucket: "<2" | "2-4" | "4-8" | "8-16" | ">16"
  context_len_bucket: "<4k" | "4-16k" | "16-64k" | ">64k"
}
```

`content_type` and `hop_estimate` are computed by cheap heuristics at capture time (§3.7). Storing contexts by reference keeps the queue small and lets retention be enforced in one place.

### 2.4 EvalRecord

```
EvalRecord {
  request_id: string
  evaluated_at: timestamp
  method: string
  method_class: string
  ratio: float
  model: string
  slice_tags: SliceTags
  is_probe: bool
  sample_p: float                          // sampling probability (§8.1)

  shadow: {
    answer_full: string
    answer_full_2: string | null      // only for noise-floor samples
    answer_comp_t0: string | null     // only when production temperature > 0 (§4.1)
    truncated_full: bool, truncated_comp: bool   // hit max_tokens (§4.5)
    answer_full_citations: Citation[] | null
    tokens_in: int, latency_ms: int, cost_usd: float
  } | null

  agreement:    AgreementResult | null     // §4
  evidence:     EvidenceResult | null      // §5
  faithfulness: FaithfulnessResult | null  // §6, abstractive only
  probe:        ProbeResult | null         // §7, probe traffic only
  noise_floor:  { verdict: Verdict } | null  // A_full vs A_full_2, same cascade as §4

  errors: ScorerError[]                    // per-scorer failures, never fatal to the record
  versions: { scorer_set: string, nli_model: string, judge_model: string | null }
}

ScorerError { scorer: string, kind: "timeout" | "model_error" | "input_invalid", message: string }
```

A scorer that does not apply writes `null`, not a default score. A scorer that fails writes `null` and appends a `ScorerError`. Aggregation must distinguish "not applicable", "failed" and "scored".

### 2.5 Scorer interface

Every scorer is a pure function of its inputs plus injected model clients, so it can be unit-tested and re-run on stored records.

```
interface Scorer<R> {
  name: string
  version: string
  applies(rec: CaptureRecord, shadow: Shadow | null): bool
  score(input: ScorerInput, deps: Deps): Promise<R>
}

ScorerInput { capture: CaptureRecord, c_full: string, c_comp: string, shadow: Shadow | null }
Deps { nli: NliClient, judge: JudgeClient | null, embed: EmbedClient, extractor: EntityExtractor, decomposer: ClaimDecomposer }
```

Scorers must be deterministic given the same inputs and model versions: no randomness, model calls at temperature 0, fixed seeds where a library takes one.

### 2.6 Configuration

All thresholds live in one versioned config. Values below are starting defaults, to be recalibrated against the offline labeled set.

| Key | Default | Used in |
| --- | --- | --- |
| `sample.base_rate` | 0.02 | §8.1 |
| `sample.slice_weights` | see §8.1 table | §8.1 |
| `sample.noise_floor_rate` | 0.10 (of sampled records) | §4.2 |
| `sample.faithfulness_rate` | 0.25 (of sampled abstractive records) | §6 |
| `nli.entail_threshold` | 0.80 | §3.4 |
| `nli.contradict_threshold` | 0.80 | §3.4 |
| `agreement.long_answer_tokens` | 60 | §4.3 |
| `agreement.claim_threshold` | 0.90 | §4.3 |
| `evidence.attribution_threshold` | 0.70 | §5.2 |
| `evidence.attribution_top_k` | 5 | §5.2 |
| `evidence.containment_threshold` | 0.90 | §5.3 |
| `evidence.recall_high` | 0.90 | §5.5 |
| `faithfulness.source_top_k` | 3 | §6.2 |
| `shadow.max_concurrency` | 4 | §2.2 |
| `shadow.daily_budget_usd` | set per deployment | §8.1 |
| `scorer.timeout_s` | 60 | all |

## 3. Shared components

The four evaluators share seven building blocks; build and test these first, since every scorer's correctness depends on them.

### 3.1 Text normalization

`normalize(text) -> string`, used for exact-match comparison only (never for display or NLI input):

1. Unicode NFKC; lowercase.
2. Strip leading/trailing whitespace; collapse internal whitespace to one space.
3. Remove trailing punctuation and wrapping quotes or markdown emphasis (`*`, `_`, backticks) around the whole answer.
4. Canonicalize numbers: remove thousands separators, strip trailing zeros after a decimal point (`3.50` → `3.5`), keep sign and units as written.
5. Canonicalize ISO-parsable dates to `YYYY-MM-DD`.
6. Do not remove articles or stopwords; do not stem. Over-normalizing creates false agreements.

### 3.2 Key-entity extraction

`extract_entities(text) -> Entity[]`, where `Entity { kind, value_normalized, surface, span }`.

| Kind | Detection | Normalized value |
| --- | --- | --- |
| `number` | Regex for integers, decimals, percentages, with optional unit | Canonical number string + unit |
| `date` | Regex + date parser | `YYYY-MM-DD` (or `YYYY-MM`) |
| `id` | Regex: UUIDs, hashes (hex ≥ 7 chars), ticket keys (`[A-Z]+-\d+`), version strings | Lowercased |
| `code_ident` | Backtick spans; tokens matching `snake_case`, `camelCase`, `PascalCase`, `a.b.c` call paths | As written (case-sensitive) |
| `path` | Regex for file paths and URLs | As written, trailing slash removed |
| `name` | NER model (persons, orgs, products) | Lowercased |

Requirements:

- Deterministic and local; no LLM call.
- Return a set (dedupe on `kind + value_normalized`).
- Numbers that are part of an ID, path or date must not also be returned as standalone `number` entities.

### 3.3 Claim decomposition

`decompose(text, context_query?) -> Claim[]`, where `Claim { id, text, source_span }`.

- Implemented as an LLM call at temperature 0 with a fixed prompt: split the text into atomic, self-contained claims; resolve pronouns; one fact per claim; do not add information; return JSON.
- Output is validated: each claim must be non-empty and ≤ 40 tokens; malformed output is retried once, then the scorer records `ScorerError{kind: model_error}`.
- Short-circuit: texts ≤ 25 tokens with a single sentence are returned as one claim without a model call.
- Cache by `sha256(text + prompt_version)`.

### 3.4 NLI client and decision rule

`nli(premise, hypothesis) -> {entail, neutral, contradict}` probabilities summing to 1.

- Default model: a local DeBERTa-v3 cross-encoder trained on MNLI-style data, or AlignScore for long premises. The model and version are recorded in `EvalRecord.versions`.
- Premise length: if `premise` exceeds the model's window, chunk it with overlap (window − hypothesis length, 25% overlap) and take, per label, the max over chunks for `entail` and `contradict`.
- Decision rule, with thresholds from config:
  - `entail >= nli.entail_threshold` → `ENTAILS`
  - else `contradict >= nli.contradict_threshold` → `CONTRADICTS`
  - else → `NEUTRAL`
- Batch requests; cache by `sha256(premise + hypothesis + model_version)`.

### 3.5 Judge

`judge(query, answer_a, answer_b) -> {verdict, rationale}`, used only for cases the deterministic steps cannot decide.

- LLM at temperature 0, a fixed prompt version, structured JSON output with `verdict ∈ {same, a_subset_of_b, b_subset_of_a, conflict, different_but_compatible}`.
- Answer order is randomized per call with a seed derived from `request_id`, and the verdict is mapped back, to remove position bias deterministically.
- The judge model must differ from the production model when possible, to reduce self-preference.
- If no judge is configured, the cascade returns `undecided`.

### 3.6 Attribution

`attribute(claim, context, top_k) -> AttributedSpan | null`, where `AttributedSpan { doc_id, start, end, text, score }`.

1. Segment `context` into sentences (code: into logical lines or statements; tables: into rows). Keep char offsets.
2. Embed the claim and all segments; take the top-k segments by cosine similarity.
3. Score each candidate with `nli(segment, claim).entail`; also try adjacent segment pairs (segment i + i+1) to catch claims that span a boundary.
4. Return the highest-scoring candidate if its score ≥ `evidence.attribution_threshold`, else `null`.

When the app already produced citations, those take precedence over attribution (§5.2).

### 3.7 Slice-tag heuristics

Computed at capture time, cheap and deterministic:

- `content_type`: share of `C_full` lines that look like code (indentation + symbol density, fenced blocks) or table rows (pipe or tab delimited). > 60% one kind → that kind; otherwise `mixed`.
- `hop_estimate`: `multi` if the query contains comparison or composition markers ("compare", "both", "difference", "after X, what", two or more named entities); `single` otherwise; `unknown` if the query is under 3 tokens.
- Buckets: fixed edges as listed in `SliceTags`.

## 4. Evaluator 1: Shadow agreement

Measures how often compression changes the answer, by comparing `A_comp` against `A_full` and subtracting the disagreement two full-context runs already have with each other. It does not measure correctness: a wrong `A_full` is still the reference.

### 4.1 Applicability and inputs

- Applies to every sampled non-probe record with a successful shadow run, for all method classes.
- Inputs: `query`, `A_full`, and the compressed answer to compare against.
- Which compressed answer: if the production call ran at temperature 0, use `answer_comp`. Otherwise the shadow runner also generates `A_comp'` on `C_comp` at temperature 0 with otherwise identical params, and the scorer uses `A_comp'`. Store it as `shadow.answer_comp_t0`. This keeps sampling noise out of the comparison.

### 4.2 Noise floor

- For `sample.noise_floor_rate` of sampled records, the shadow runner generates `A_full_2`: a second, independent call on `C_full` with identical params.
- Run the same cascade (§4.3) on `(A_full, A_full_2)` and store the verdict in `EvalRecord.noise_floor`.
- Aggregation reports, per slice: `disagreement_rate`, `noise_floor_rate`, and `excess_disagreement = disagreement_rate − noise_floor_rate`. Alerts fire on `excess_disagreement`, never on the raw rate.
- If a slice has fewer than 30 noise-floor samples in the window, use the global noise floor for that model and flag the slice `noise_floor_source: global`.

### 4.3 Scoring cascade

Let `A = A_full` and `B` = the compressed answer from §4.1. Stop at the first step that returns a verdict.

**Step 0: abstention.** Detect whether each answer declines to answer ("not in the provided context", "I don't know", "cannot determine"). Detection = a fixed phrase list first, then `nli(answer, "The answer says the information is not available.")` → `ENTAILS`.

| A abstains | B abstains | Verdict | `abstention` field |
| --- | --- | --- | --- |
| yes | yes | `agree` | `both` |
| no | yes | `omission` | `compressed_only` |
| yes | no | `addition` | `full_only` |
| no | no | continue | `none` |

**Step 1: exact match.** `normalize(A) == normalize(B)` → `agree`.

**Step 2: key entities.** Compute `E_A = extract_entities(A)`, `E_B = extract_entities(B)`, considering kinds `number, date, id, code_ident, path`.

- `missing = E_A \ E_B`, `extra = E_B \ E_A` (compared on `kind + value_normalized`).
- If, for some kind k, `missing` and `extra` both contain an entity of kind k → `contradict` with `reason: entity_conflict`. Example: A says "timeout is 30s", B says "timeout is 60s".
- Otherwise record `missing` and `extra` and continue. NLI is weak at single-token value differences, which is why this step precedes it.

**Step 3: length switch.** If either answer exceeds `agreement.long_answer_tokens`, go to Step 5 (claim level). Otherwise continue.

**Step 4: bidirectional NLI** (short answers). Prepend the query to each premise so short answers are interpretable:

- `fwd = nli("Q: " + query + "\nA: " + A, B)`
- `bwd = nli("Q: " + query + "\nA: " + B, A)`

| fwd | bwd | Verdict |
| --- | --- | --- |
| ENTAILS | ENTAILS | `agree` |
| ENTAILS | not ENTAILS | `omission` (B says less) |
| not ENTAILS | ENTAILS | `addition` (B says more) |
| CONTRADICTS | any | `contradict` |
| any | CONTRADICTS | `contradict` |
| NEUTRAL | NEUTRAL | go to Step 6 |

If Step 2 recorded non-empty `missing` and Step 4 says `agree`, downgrade to `omission` with `reason: entity_missing`.

**Step 5: claim level** (long answers).

1. `claims_A = decompose(A)`, `claims_B = decompose(B)`.
2. For each `c ∈ claims_A`: `nli(B, c)`. For each `c ∈ claims_B`: `nli(A, c)`. Use the chunking rule of §3.4 when an answer exceeds the NLI window.
3. `claim_recall = |{c ∈ claims_A : ENTAILS}| / |claims_A|`.
4. `claim_precision = |{c ∈ claims_B : ENTAILS}| / |claims_B|`.
5. Verdict, with `t = agreement.claim_threshold` (default 0.9; add to config):
   - any claim in either direction `CONTRADICTS` → `contradict`
   - `recall ≥ t` and `precision ≥ t` → `agree`
   - `recall < t` and `precision ≥ t` → `omission`
   - `recall ≥ t` and `precision < t` → `addition`
   - both `< t` → `divergent`
6. Store the unmatched claims (`lost_claims` from A, `new_claims` from B).

**Step 6: judge** (only for Step 4 NEUTRAL/NEUTRAL). Call `judge(query, A, B)` and map: `same` → `agree`, `b_subset_of_a` → `omission`, `a_subset_of_b` → `addition`, `conflict` → `contradict`, `different_but_compatible` → `divergent`. No judge configured → `undecided`.

**Code blocks.** If both answers contain fenced code, compare the code separately: whitespace- and comment-insensitive token comparison. Identical → continue the cascade on the prose only; different → `divergent` with `reason: code_differs`, then run the judge if configured to refine to `agree` or `contradict`.

### 4.4 Output

```
AgreementResult {
  verdict: "agree" | "omission" | "addition" | "contradict" | "divergent" | "undecided"
  decided_by: "abstention" | "exact" | "entity" | "nli" | "claims" | "judge" | "code"
  reason: string | null              // e.g. "entity_conflict", "entity_missing", "code_differs"
  abstention: "none" | "both" | "compressed_only" | "full_only"
  entity_missing: Entity[]           // in A, not in B
  entity_extra: Entity[]             // in B, not in A
  nli_fwd: {entail, neutral, contradict} | null
  nli_bwd: {entail, neutral, contradict} | null
  claim_recall: float | null
  claim_precision: float | null
  lost_claims: string[]
  new_claims: string[]
  judge_rationale: string | null
}
```

`agree` is the only verdict that counts as agreement. All others count as disagreement in `disagreement_rate`; `undecided` is reported separately and excluded from both numerator and denominator.

### 4.5 Edge cases

- **Shadow failure** (timeout, provider error): no `AgreementResult`; record `ScorerError` and count it in `shadow_failure_rate` per slice. Do not retry more than once.
- **Empty answer** on either side: treat as abstention.
- **Formatting only** (bullets vs prose, different ordering of the same items): should resolve to `agree` via NLI or claims; add test fixtures for this.
- **Answers in different languages** than each other: `divergent` with `reason: language_mismatch`, no NLI.
- **Truncated answers** (hit `max_tokens`): flag `truncated: true` on the side affected; aggregation excludes truncated pairs from rates and counts them separately.

## 5. Evaluator 2: Proxy evidence recall

Measures whether the evidence the full-context answer relied on is still present in the compressed context. It scores the compressor in isolation, so a drop here points at the compressor rather than at generation.

### 5.1 Applicability

- Applies to sampled non-probe records with a successful shadow run, all method classes. The survival test differs per class (§5.3).
- Not applicable when `A_full` abstains (no evidence was used): write `EvidenceResult{status: not_applicable_abstained}`.

### 5.2 Finding the used spans

The used spans are the parts of `C_full` that support `A_full`.

1. **Citations first.** If `answer_full_citations` is non-empty, each citation is a used span. Validate offsets against `C_full`; drop invalid ones and count them in `invalid_citations`.
2. **Attribution otherwise.**
   - `claims = decompose(A_full)`.
   - For each claim: `attribute(claim, C_full, evidence.attribution_top_k)` (§3.6).
   - Claims with no span above `evidence.attribution_threshold` are recorded as `unattributed_claims`.
3. Merge overlapping or adjacent spans within the same document.
4. If no span was found at all: `status: no_attribution`. This usually means the model answered from parametric knowledge, or attribution failed; it is reported as its own rate, never as recall 0 or 1.

Keep the claim ↔ span mapping: each `UsedSpan` carries the claim ids it supports.

### 5.3 Survival test per method class

**Extractive** (`C_comp` is a selection of whole sentences or chunks):

1. Exact substring check of the span text in `C_comp` (after whitespace normalization). Found → `survived`.
2. Else fuzzy containment: slide a window of `1.5 × len(span_tokens)` over `C_comp` tokens, anchored at occurrences of the span's rarest tokens; containment = `LCS(span_tokens, window_tokens) / len(span_tokens)`. Max ≥ `evidence.containment_threshold` → `survived`, else `dropped`.

**Token pruning** (`C_comp` is `C_full` with tokens removed; LongLLMLingua may also reorder documents):

1. Compute an alignment from `C_comp` tokens back to `C_full` tokens once per record: per document, a monotonic sequence alignment (difflib or equivalent); documents matched to each other first by highest token overlap, to handle reordering.
2. For each span: `token_survival = kept span tokens / span tokens`.
3. `critical_token_recall` = fraction of the span's key entities (§3.2) whose tokens are all kept.
4. `kept_text` = the kept tokens of the span joined in order; `semantic = nli(kept_text, claim)` for each claim the span supports.
5. Span `survived` if `critical_token_recall == 1` and every supported claim gets `ENTAILS`; `damaged` if some kept but either check fails; `dropped` if `token_survival < 0.2`.

**Abstractive** (`C_comp` is a rewrite; spans cannot be located):

1. Test each supported claim instead of each span: `nli(C_comp, claim)` with chunking (§3.4).
2. `ENTAILS` → `survived`; `CONTRADICTS` → `distorted`; `NEUTRAL` → `dropped`.
3. Additionally check key entities of the span verbatim in `C_comp`; record `critical_token_recall` the same way as for token pruning, using presence anywhere in `C_comp`.

### 5.4 Metrics per record

- `span_recall = survived spans / used spans` (for abstractive: survived claims / tested claims).
- `damaged_rate` = damaged spans / used spans (token pruning only).
- `critical_token_recall` = mean across spans.
- `attribution_coverage = attributed claims / claims` (only when attribution, not citations, was used).

### 5.5 Diagnosis

When the same record also has an `AgreementResult`, compute `diagnosis`. Recall counts as high when `span_recall ≥ 0.9` (`evidence.recall_high`, add to config).

| Recall | Agreement | `diagnosis` | Meaning |
| --- | --- | --- | --- |
| high | agree | `healthy` | Evidence kept, answer unchanged |
| low | not agree | `compressor_dropped_evidence` | Fix the compressor |
| high | not agree | `generation_shift` | Evidence kept, answer changed: ordering, position, damaged spans |
| low | agree | `context_not_needed` | Model didn't need the context, or attribution is wrong |
| any | undecided | `unknown` |  |

### 5.6 Output

```
EvidenceResult {
  status: "scored" | "not_applicable_abstained" | "no_attribution"
  span_source: "citations" | "attribution"
  used_spans: UsedSpan[]
  span_recall: float | null
  damaged_rate: float | null
  critical_token_recall: float | null
  attribution_coverage: float | null
  unattributed_claims: string[]
  invalid_citations: int
  diagnosis: "healthy" | "compressor_dropped_evidence" | "generation_shift" | "context_not_needed" | "unknown" | null
}

UsedSpan {
  doc_id: string, start: int, end: int, text: string
  claim_ids: string[]
  outcome: "survived" | "damaged" | "dropped" | "distorted"
  containment: float | null        // extractive
  token_survival: float | null     // token pruning
  critical_token_recall: float
  missing_entities: Entity[]
}
```

### 5.7 Edge cases

- **Spans larger than the NLI window** (long code blocks, tables): split into sub-spans at segment boundaries; a span survives only if all sub-spans needed by its claims survive.
- **Duplicate content** in `C_full` (same sentence in two documents): a span survives if any copy survives.
- **Compressor adds connective text** (some extractive pipelines insert headers or separators): strip known separator patterns from `C_comp` before matching.
- **Very short spans** (< 5 tokens): skip fuzzy matching; exact match or entity presence only.

## 6. Evaluator 3: Source faithfulness

Measures whether abstractive compression invents or distorts content, by checking every claim in the compressed context, and in the answer built on it, against the original source. Extractive and token-pruning methods cannot add claims, so this evaluator does not run for them.

### 6.1 Applicability and inputs

- Applies only when `method_class == abstractive`, to `sample.faithfulness_rate` of sampled abstractive records (it is the most expensive scorer).
- Does not require a shadow run for check A; check B uses `A_full` when available (§6.3).
- **Source requirement for rolling compaction:** `context_full_ref` must point to the original uncompressed material covered by the compaction (raw turns and tool results), not to the previous summary. The capture hook must preserve that material for every step. Scoring against the previous summary hides compounding drift and is not allowed.

### 6.2 Check A: compressed context vs source

1. Chunk `C_full` into segments that fit the NLI window (sentence-aligned, 25% overlap); embed all chunks once per record.
2. `claims = decompose(C_comp)`. For long compactions, decompose per paragraph and concatenate.
3. For each claim:
   1. Retrieve the top `faithfulness.source_top_k` chunks by embedding similarity.
   2. Run `nli(chunk, claim)` on each; take `max entail` and `max contradict`.
   3. Label: `max entail ≥ nli.entail_threshold` → `supported`; else `max contradict ≥ nli.contradict_threshold` → `contradicted`; else `unsupported`.
   4. **Retrieval-miss guard:** if `unsupported`, retry once with `3 × source_top_k` chunks before finalizing.
4. Entity check per claim: key entities (§3.2) in the claim that appear nowhere in `C_full` are recorded as `invented_entities`; any such entity forces the label to `unsupported` even if NLI says `supported`.
5. Rates: `supported_rate`, `unsupported_rate`, `contradicted_rate` over all claims.

### 6.3 Check B: compressed answer vs source

Catches invented content that propagated into the answer the user saw.

1. `claims = decompose(A_comp)` (use `A_comp'` if the shadow runner produced one, per §4.1).
2. Label each claim against `C_full` exactly as in §6.2.
3. For each `unsupported` or `contradicted` claim, if `A_full` is available, check `nli(A_full, claim)`:
   - `ENTAILS` → mark `also_in_full_answer: true`. The model adds this content even with the full context (general knowledge, not compression), so it is excluded from the compression-induced rate.
   - otherwise `also_in_full_answer: false`.
4. Rates: the same three rates as check A, plus `compression_induced_unsupported_rate` = unsupported or contradicted claims with `also_in_full_answer == false`, divided by all claims. Without `A_full`, this rate is `null`.

### 6.4 Rolling-compaction drift

- Every compaction step is scored individually, with `step_index = compaction_step`.
- The aggregator groups records by conversation id and computes, per method, `unsupported_rate` by step index (mean across conversations) and the slope of a linear fit over steps 0..N.
- A positive slope means errors compound across steps; it is reported as `drift_per_step`.
- Conversation id: add `conversation_id: string | null` to `CaptureRecord`.

### 6.5 Output

```
FaithfulnessResult {
  step_index: int | null
  context_check: {
    claims_total: int
    supported_rate: float
    unsupported_rate: float
    contradicted_rate: float
    bad_claims: FaithClaim[]          // unsupported + contradicted only
  }
  answer_check: {
    claims_total: int
    supported_rate: float
    unsupported_rate: float
    contradicted_rate: float
    compression_induced_unsupported_rate: float | null
    bad_claims: FaithClaim[]
  } | null
}

FaithClaim {
  text: string
  label: "unsupported" | "contradicted"
  best_chunk: { doc_id, start, end, entail: float, contradict: float } | null
  invented_entities: Entity[]
  also_in_full_answer: bool | null   // answer_check only
}
```

### 6.6 Edge cases

- **Meta-statements in summaries** ("The user asked about X", "Earlier, the agent ran tests"): the decomposition prompt must rewrite them as content claims ("The tests were run") or drop pure framing. Framing claims that cannot be checked are labeled `unsupported` but tagged `meta: true` and excluded from rates.
- **Aggregate claims** ("Most of the errors were timeouts"): NLI against a single chunk often fails. If unsupported after the retrieval-miss guard, run the judge with the top chunks concatenated, if a judge is configured.
- **Very long sources** (> 64k tokens): embedding retrieval is mandatory; never run NLI against the full source.

## 7. Evaluator 4: Synthetic needle probes

Plants facts the model cannot know into copies of real documents, sends them through the real compression and model path, and checks whether each fact survived compression and was answered. It is the only online evaluator with exact labels.

### 7.1 Isolation

- Probes run on copies of documents; user originals are never modified.
- Probe requests carry `is_probe: true` end to end. They are excluded from production analytics and user-visible history, and their `EvalRecord`s are aggregated separately.
- Probe documents and answers are stored in a separate namespace with the same retention rules as captured contexts.

### 7.2 Document pool

- Source: `C_full` from captured records, sampled stratified by `content_type × context_len_bucket`.
- Exclude documents that already contain probe markers, or that are shorter than 512 tokens.
- Snapshot the pool per probe-set version so regression runs use identical documents.

### 7.3 Needle generation

1. **Invented entities.** Generate pseudo-words (pronounceable, 6–10 letters, seeded) for names, and random values for numbers, dates and IDs. Reject any entity that appears in the document pool.
2. **Fact text.** Fill a template for the probe type (§7.5).
3. **Style matching.** Rewrite the fact in the host document's register with an LLM at temperature 0, given the surrounding 3 sentences. Content-type rules:
   - code → a comment or a constant/config assignment
   - table → a new row with the table's columns
   - prose → a sentence in the paragraph's style
4. **Value preservation check.** After rewriting, every planted entity and value must appear verbatim in the needle text; otherwise regenerate (max 3 attempts, then drop).
5. **Parametric check.** Ask the production model the probe query with no context. If it produces the expected value, discard the probe (the fact is guessable).

Perplexity-based pruners keep high-surprisal text, so an out-of-place needle survives more often than real content and inflates retention. That is why step 3 is mandatory, not optional.

### 7.4 Placement

- Depths: `[0.0, 0.25, 0.5, 0.75, 1.0]` of the context's token length, snapped to the nearest sentence, line or row boundary.
- Multi-document contexts: depth is computed over the concatenated context; also record `doc_index`.
- Record the needle's char span in the probe context for scoring.

### 7.5 Probe types

| Type | Planted | Query asks | Expected |
| --- | --- | --- | --- |
| `single_fact` | One fact about an invented entity | The fact's value | The value |
| `numeric` | A number with unit ("the Orvane batch limit is 4,812 rows") | The number | Number + unit, entity-matched |
| `identifier` | An ID, version or ticket key | The ID | Exact ID |
| `code_ident` | A function/constant name and its value or role | The name or value | Exact identifier |
| `two_hop` | Fact 1: A → B; fact 2: B → C, at depths 0.25 and 0.75 (or reversed) | A → ? via B | C |
| `distractor` | The target fact plus a near-miss (same entity, different qualifier and value) at a different depth | The target qualifier | Target value, and not the distractor value |
| `negative` | Nothing | A fact about an invented entity | Abstention |

### 7.6 Queries

Each probe gets two query variants, run as separate probes:

- `lexical`: shares the entity name and key noun with the needle.
- `paraphrased`: same intent, with the key noun replaced by a synonym or description; the entity name is kept.

The gap between the two variants measures how much a query-aware compressor depends on lexical overlap.

### 7.7 Execution

1. Send `(query, probe_context)` through the production compression + model path with the production config.
2. **Control run:** also send `(query, probe_context)` uncompressed. If the control fails, the miss is a model or long-context limit, not compression.
3. Store the resulting `CaptureRecord` with `is_probe: true, probe_id`.

### 7.8 Scoring

- `survived_compression`: the needle survived in `C_comp`, using the survival test for the method class (§5.3), with the needle span as the used span and the needle fact as the claim. For `two_hop`, both needles must survive (`survived_partial` if only one does).
- `answered_correctly`: key-entity match (§3.2) of the expected value in the answer. For `distractor`, the answer must contain the target value and not the distractor value. For `negative`, it is `null`.
- `abstained_correctly`: for `negative` only, the abstention detection from §4.3 Step 0.
- `control_correct`: the same checks applied to the control answer.
- `outcome`:

| survived | answered | control | `outcome` |
| --- | --- | --- | --- |
| yes | yes | any | `pass` |
| no | no | yes | `compressor_drop` |
| yes | no | yes | `generation_miss` |
| any | no | no | `model_limit` |
| no | yes | any | `answered_without_evidence` (investigate: leak or guess) |

Rows are evaluated top to bottom; the first match wins. Distractor probes that return the distractor value get `outcome: distractor_confusion`. Negative probes: `pass` if abstained, `false_answer` otherwise.

### 7.9 Probe sets and schedule

- **Regression set:** fixed seed, fixed document snapshot. Full grid = 7 types × 5 depths × 2 query variants × each content type present. Versioned; changes only by explicit version bump.
- **Rotating set:** same grid, regenerated weekly with a new seed and new documents.
- **Runs:** regression set on every deploy that changes the compressor, its config, or the model; both sets daily.
- Comparisons across runs are only valid within the same regression-set version.

### 7.10 Output

```
ProbeResult {
  probe_id: string
  probe_set: "regression" | "rotating"
  probe_set_version: string
  probe_type: "single_fact" | "numeric" | "identifier" | "code_ident" | "two_hop" | "distractor" | "negative"
  query_variant: "lexical" | "paraphrased"
  depth: float                        // two_hop: depth of the first needle
  depth_2: float | null               // two_hop only
  doc_index: int
  content_type: string
  survived_compression: "yes" | "no" | "partial" | null   // null for negative
  answered_correctly: bool | null
  abstained_correctly: bool | null
  control_correct: bool
  outcome: "pass" | "compressor_drop" | "generation_miss" | "model_limit" | "answered_without_evidence" | "distractor_confusion" | "false_answer"
  expected: string | null
  answer: string
}
```

## 8. Sampling, aggregation, metrics, alerting

Sampling is stratified toward failure-prone slices, aggregation re-weights by inverse sampling probability so global numbers stay unbiased, and alerts fire per slice on excess over the noise floor or over the slice's own history.

### 8.1 Sampler

- **Deterministic:** `u = hash(request_id + sampler_version) / 2^64`; the record is sampled if `u < p`. Re-running the sampler on stored captures gives the same selection.
- **Probability:** `p = min(1, sample.base_rate × w)`, where `w` is the product of slice weights (defaults; add to config):

| Tag | Weight |
| --- | --- |
| `ratio_bucket` 8-16 or >16 | 3 |
| `content_type` code or table | 2 |
| `hop_estimate` multi | 2 |
| `context_len_bucket` >64k | 2 |
| otherwise | 1 |

- Store `sample_p` on the `EvalRecord`.
- **Budget:** the shadow runner tracks spend against `shadow.daily_budget_usd`. When the budget is exhausted, stop shadow runs for the rest of the day; captures continue, and scorers that need no shadow (faithfulness check A) still run. Log `budget_exhausted_at`.
- **Slice floor:** if a slice with traffic has fewer than 30 evaluated records in the trailing 24h, its weight is doubled for the next day (max weight 16). This keeps rare slices measurable.
- **Sub-sampling:** noise-floor runs (`sample.noise_floor_rate`) and faithfulness (`sample.faithfulness_rate`) are drawn from already-sampled records with an independent hash.

### 8.2 Aggregation

- **Windows:** 1h, 24h, 7d, rolling.
- **Group-by:** `method × model × ratio_bucket × content_type × hop_estimate × context_len_bucket`, plus every single-dimension marginal and a global total.
- **Weighting:** slice-level numbers are unweighted within the slice; marginals and global use weights `1 / sample_p` (Horvitz–Thompson style).
- **Minimum n:** a metric with fewer than 30 contributing records is reported as `insufficient_n`, never as a number.
- **Intervals:** Wilson 95% interval for rates; bootstrap (1,000 resamples, fixed seed) for means.
- Probe records are aggregated separately, grouped by `probe_set_version × probe_type × depth × query_variant × content_type × method`.

**Metrics**

| Metric | Definition | Source |
| --- | --- | --- |
| `disagreement_rate` | Share of decided agreement verdicts ≠ `agree` | §4 |
| `noise_floor_rate` | Same, on noise-floor pairs | §4.2 |
| `excess_disagreement` | `disagreement_rate − noise_floor_rate` | §4.2 |
| `verdict_mix` | Share of each verdict (omission, addition, contradict, divergent) | §4 |
| `undecided_rate` | Share of `undecided` | §4 |
| `abstention_induced_rate` | Share with `abstention == compressed_only` | §4 |
| `span_recall_mean` | Mean `span_recall` over `scored` records | §5 |
| `critical_token_recall_mean` | Mean `critical_token_recall` | §5 |
| `no_attribution_rate` | Share of evidence records with `status == no_attribution` | §5 |
| `diagnosis_mix` | Share of each diagnosis | §5.5 |
| `ctx_unsupported_rate`, `ctx_contradicted_rate` | Mean of check-A rates | §6.2 |
| `answer_compression_induced_rate` | Mean of check-B compression-induced rate | §6.3 |
| `drift_per_step` | Slope of unsupported rate over compaction steps | §6.4 |
| `probe_pass_rate` | Share of probes with `outcome == pass` | §7.8 |
| `probe_outcome_mix` | Share of each probe outcome | §7.8 |
| `lexical_gap` | `probe_pass_rate(lexical) − probe_pass_rate(paraphrased)` | §7.6 |
| `shadow_failure_rate` | Failed shadow runs / attempted | §4.5 |
| `eval_cost_usd` | Shadow + judge + decomposition spend | §2.4 |

### 8.3 Alert rules

All rules require the metric to be reported (not `insufficient_n`). Thresholds are starting defaults.

| Rule | Fires when | Window |
| --- | --- | --- |
| Excess disagreement spike | Slice `excess_disagreement` > 2× its 7d median, and the Wilson lower bound > the 7d median | 24h |
| Contradictions | Slice `verdict_mix.contradict` > 3% | 24h |
| Evidence loss | Slice `span_recall_mean` drops > 0.05 below its 7d mean | 24h |
| Induced abstention | Slice `abstention_induced_rate` > 2× its 7d median | 24h |
| Faithfulness | `answer_compression_induced_rate` > 2%, or `ctx_contradicted_rate` > 1% | 24h |
| Compaction drift | `drift_per_step` > 0.01 (1 point per step) | 7d |
| Probe regression | Regression-set `probe_pass_rate` drops vs the previous run of the same version, McNemar p < 0.05 on paired probe ids | Per run |
| Shadow health | `shadow_failure_rate` > 10% | 1h |
| Attribution health | `no_attribution_rate` > 2× its 7d median | 24h |

Alerts carry the slice key, metric value, interval, baseline, and links to 5 example records.

### 8.4 Review export

Records go to the offline review queue when either:

- `agreement.verdict ∈ {omission, addition, contradict, divergent}`, or
- they fall in a 1% random sample of `agree` records (needed to estimate how often agreement hides a shared error).

The export includes the query, both answers, used spans with outcomes, and faithfulness bad claims, with references to contexts rather than inline text.

## 9. Testing and acceptance criteria

The build is accepted when every item below passes. The scorer-validity checks (§9.2) matter most, because the evaluators are themselves models whose errors would otherwise go unmeasured.

### 9.1 Unit tests (fixtures, no network)

| Component | Required fixtures |
| --- | --- |
| `normalize` | Number formats, dates, quotes/markdown wrappers; must not equate "30s" and "60s" |
| `extract_entities` | Each kind; numbers inside IDs/paths not double-counted; dedupe |
| NLI decision rule | Threshold boundaries; chunking over long premises (max over chunks) |
| Agreement cascade | One fixture per verdict and per `decided_by`; formatting-only differences → `agree`; value change → `contradict` via entity step; abstention table rows |
| Evidence, extractive | Exact hit, fuzzy hit above/below threshold, duplicate content |
| Evidence, token pruning | Alignment with reordered documents; entity token dropped → `damaged`; < 20% kept → `dropped` |
| Evidence, abstractive | Claim survived / distorted / dropped |
| Faithfulness | Supported, contradicted, unsupported; invented entity forces unsupported; retrieval-miss guard; `also_in_full_answer` exclusion |
| Probe generator | Value preservation check rejects lossy rewrites; parametric check discards guessable probes; placement snaps to boundaries |
| Probe scoring | Every row of the outcome table, including distractor and negative cases |
| Sampler | Determinism; weights; slice-floor doubling; budget cut-off |
| Aggregator | Inverse-probability weighting on a synthetic dataset with known true rate; `insufficient_n`; Wilson interval values |

Model clients (NLI, judge, decomposer, embeddings) are injected, so unit tests use stubs with scripted outputs.

### 9.2 Scorer validity (meta-evaluation, with real models)

Build a small hand-labeled set once, from real captured pairs, and keep it versioned in the repo:

- **Agreement:** 200 `(query, A, B)` pairs labeled with the verdict. Target: cascade verdict matches the human label on ≥ 90% of decided pairs; `contradict` precision ≥ 90%; `undecided` ≤ 10%.
- **Attribution:** 100 claims labeled with their supporting span. Target: correct span in top-1 ≥ 80%, and `null` returned for unsupported claims ≥ 80%.
- **Faithfulness:** 150 claims labeled supported / contradicted / unsupported against their source. Target: macro-F1 ≥ 0.80; `contradicted` precision ≥ 85%.
- **Probes:** manual review of 50 generated probes. Target: ≥ 95% read naturally in context and have exactly one correct answer.

These thresholds are proposals; if a scorer misses its target, report the measured value rather than tuning on the test set. Re-run the meta-evaluation whenever the NLI, judge or decomposer model or prompt version changes.

### 9.3 Integration and system tests

- **End to end:** with a fake compressor that deterministically drops a known sentence, the pipeline produces `omission` or `contradict`, `compressor_dropped_evidence`, and a probe `compressor_drop` for a needle placed in that sentence.
- **Determinism:** re-scoring 100 stored records with the same versions yields byte-identical `EvalRecord`s, excluding timestamps and cost.
- **Non-blocking:** capture hook adds ≤ 2 ms p99 latency to the production path; with the queue unavailable, production requests still succeed and the hook logs the drop.
- **Isolation:** probe traffic never appears in production metrics or user history.
- **Budget:** with a tiny budget, shadow runs stop at the cap and resume the next day.

### 9.4 Acceptance checklist

- [ ] All unit fixtures in §9.1 pass
- [ ] Scorer-validity targets in §9.2 measured and reported (met or explicitly waived)
- [ ] Integration tests in §9.3 pass
- [ ] All thresholds read from one versioned config; no hard-coded thresholds in scorer code
- [ ] `EvalRecord.versions` populated on every record
- [ ] Dashboard shows every metric in §8.2 per slice, with intervals and `insufficient_n`
- [ ] Every alert rule in §8.3 fires on a seeded synthetic regression
- [ ] Review export (§8.4) produces records the offline loop can ingest

## 10. Open decisions and known limitations

Five decisions must be made before implementation starts; the limitations are inherent to label-free evaluation and should be stated wherever the numbers are shown.

### 10.1 Open decisions

| Decision | Options | Why it matters |
| --- | --- | --- |
| Language and service placement | Inside the gateway process vs a separate eval worker | Determines the queue, storage and model-client libraries |
| Context storage and retention | Raw text with TTL vs hashes + references to source documents | Shadow runs, attribution and probes all need `C_full`; retention limits on user documents may forbid raw storage |
| NLI model | English DeBERTa-v3 vs a multilingual model (e.g. mDeBERTa) | If any traffic is non-English or mixed language, an English-only NLI model will mislabel it |
| Judge | None (cascade ends at `undecided`) vs a configured model | Cost and the size of the `undecided` bucket |
| Citations in production prompts | Ask the model to cite spans vs rely on attribution | Citations make §5 cheaper and more reliable, but change the production prompt |

### 10.2 Known limitations

- **Agreement is not correctness.** A shared error on both sides scores as `agree`. The random `agree` sample in the review export (§8.4) is the only way to estimate how often that happens.
- **Faithfulness is not completeness.** A compaction can be fully supported and still have dropped what the next query needs; §5 and §7 cover that, §6 does not.
- **Scorer error compounds.** NLI, attribution and decomposition each have their own error rates; slice-level rates inherit them. §9.2 measures them, but they are not corrected for in the dashboard.
- **Probes are unnatural even after style matching.** Treat probe pass rates as a regression signal across versions, not as an absolute retention rate.
- **Shadow runs cost money.** Every sampled request pays for a full-context call, which reduces the savings being measured; `eval_cost_usd` should be shown next to the compression savings.
- **Temperature-0 nondeterminism.** Providers do not guarantee identical outputs at temperature 0, which is why the noise floor exists; it must be re-estimated after any model change.
