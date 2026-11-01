# Token compression methods

> Background survey. What Aiport implements is in [compression.md](compression.md):
> opt-in format methods (whitespace, JSON minify, dedupe) with per-request
> measurement and application feedback.

An overview of the ways to reduce the number of tokens a language model has to
read, store or pay for, grouped by where the saving happens.

## 1. Prompt-level compression (text in, shorter text out)

These work with any model, including closed APIs.

- **Token pruning by importance.** A small model scores each token or phrase,
  and the low-information ones are dropped.
  - *Selective Context* scores by self-information.
  - *LLMLingua* uses a small LM's perplexity and works coarse to fine: it drops
    whole demonstrations first, then individual tokens.
  - *LongLLMLingua* is a query-aware version for long contexts and RAG. It also
    reorders documents to work around the "lost in the middle" problem.
  - *LLMLingua-2* trains a token classifier (keep or drop), distilled from GPT-4
    compressions. It is faster and doesn't depend on the task.
  - Typical result: 2–20× compression with modest quality loss.
- **Extractive compression.** Keep only the sentences or chunks relevant to the
  query. Examples are RECOMP's extractive mode and rerank-then-truncate in RAG.
- **Abstractive summarization.** An LLM summarizes earlier context or documents.
  Examples are RECOMP's abstractive mode, rolling summaries and agent
  "compaction". It compresses more than extraction but can lose detail and
  invent details that aren't in the source.
- **Format and serialization tricks.** Minify JSON, drop repeated keys
  (CSV-like formats or TOON instead of JSON), strip HTML and boilerplate,
  shorten IDs and URLs, and remove duplicate content.

## 2. Learned soft compression (text → dense vectors)

These need access to the model's embeddings.

- **Gist tokens** (*Gisting*): the model learns to squeeze a prompt into a few
  special tokens.
- **AutoCompressor and summary vectors:** recursively compress long documents
  into summary embeddings.
- **ICAE** (In-Context Autoencoder) and **500xCompressor:** an encoder turns
  context into memory slots that a frozen LLM decoder reads.
- **xRAG:** squeezes a retrieved document into a single embedding token.
- These reach very high ratios but are tied to one model and don't work through
  a standard API.

## 3. KV-cache compression (inside the model at inference time)

These cut memory and compute, not the number of tokens you are billed for.

- **Eviction:** *H2O* (heavy hitters), *StreamingLLM* (attention sinks plus a
  recent window), *SnapKV*, *PyramidKV*, *Scissorhands*.
- **Quantization:** *KIVI* and *KVQuant* store the cache in 2–4 bits.
- **Architectural:** multi-query attention and grouped-query attention share
  keys and values across heads, multi-head latent attention (DeepSeek)
  compresses them into a low-rank latent, and sliding-window attention limits
  how far back each token looks.
- **Token merging** for vision and multimodal models: *ToMe*, plus pooling or
  pruning of image patch tokens, which are often the largest share of the
  tokens.

## 4. Tokenizer-level

Larger or domain-specific vocabularies mean fewer tokens for the same text,
especially code and non-English text. Byte-level and dynamic-patching
approaches (*BLT*) group bytes adaptively.

## 5. Avoiding resending tokens (cost compression, not true compression)

- **Prompt caching:** reuse a static prefix such as the system prompt, tools or
  documents at a fraction of the price.
- **Agent context management:** clear old tool results, use sub-agents that
  return only conclusions, keep memory files and fetch them only when needed
  (retrieval on demand).

## Rule of thumb

| Goal | Best fit |
|---|---|
| Lower API bills on long documents or RAG | Extractive or query-aware pruning (LongLLMLingua, LLMLingua-2) plus prompt caching |
| Long-running agents | Summarization or compaction plus tool-result clearing |
| Self-hosted inference memory | KV-cache eviction or quantization |
| Maximum compression ratio, own model | Soft prompts (gist tokens, ICAE) |
