---
title: Speculative Decoding
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/speculative.md
source: git
fetched_at: 2026-04-28T09:49:46.374054321-03:00
rendered_js: false
word_count: 642
summary: This document explains the speculative decoding feature in llama.cpp, detailing how various implementations, including draft models and n-gram based strategies, can be used to accelerate token generation.
tags:
    - speculative-decoding
    - llama-cpp
    - llm-inference
    - performance-optimization
    - ngram-modeling
    - token-generation
category: reference
optimized: true
optimized_at: 2026-04-28T12:00:00Z
---
# Speculative Decoding

Speculative decoding accelerates token generation by predicting multiple tokens ahead of the main model.

[Speculative decoding](https://en.wikipedia.org/wiki/Transformer_(deep_learning)#Speculative_decoding) leverages the fact that computing n tokens in a batch is more efficient than computing n sequentially. Draft tokens are generated quickly, then verified with the target model in a single batch, achieving substantial speedups when draft predictions are frequently correct.

## Implementations

`llama-server` supports several speculative decoding implementations. Implementations with draft models can be mixed with draftless implementations.

### Draft Model (`draft`)

A smaller draft model generates drafts. Most common speculative decoding approach.

### n-gram Cache (`ngram-cache`)

Maintains statistics about short n-gram sequences. Drafts use probabilities derived from these statistics. External statistics can be loaded from files for improved accuracy.

- #5479, #6828, #6848

### n-gram Map Variants

These implementations search token history for patterns and use matching sequences as draft candidates. Require no additional model but rely on patterns already in generated text. Useful for source code rewriting by LLMs.

#### n-gram Simple (`ngram-simple`)

Finds last matching n-gram in history, creates draft using m tokens following it. Simplest self-speculative approach with minimal overhead.

```bash
llama-server [...] --spec-type ngram-simple --draft-max 64
```

#### n-gram Map Key (`ngram-map-k`)

Searches for current n-gram of size n (_key_) in token history. If key is followed by same m tokens (_mgram_) multiple times, creates draft using these m tokens. Requires minimum occurrences (`--spec-ngram-min-hits`, default 1) before generating drafts.

Stores accepted token count for each used n-gram.

```bash
llama-server [...] --spec-type ngram-map-k --draft-max 64
```

#### n-gram Map Key-4-Values (`ngram-map-k4v`)

Experimental. Searches for current n-gram of size n (_key_) in history. Tracks up to four _values_ (n-grams of size m, called _mgrams_) per key. Internal statistic counts mgram occurrences after key n-gram. Uses most frequent mgram as draft if significantly more frequent than others.

Stores accepted token count for each used n-gram.

Best for longer repetitions:

```bash
llama-server [...] --spec-type ngram-map-k4v --spec-ngram-size-n 8 --spec-ngram-size-m 8 --spec-ngram-min-hits 2 --draft-max 64
```

### n-gram Mod (`ngram-mod`)

Basic ngram hasher for speculative decoding:
- Computes hash for each ngram using LCG
- Stores next token for each computed hash
- During speculation, computes rolling hash of last n tokens and picks next token from storage

Characteristics:
- Lightweight (~16 MB)
- Constant memory and complexity
- Variable draft lengths (m not fixed)
- Single hash pool shared across all server slots (different requests benefit each other)

```bash
llama-server ... --spec-type ngram-mod --spec-ngram-size-n 24 --draft-min 48 --draft-max 64
```

Applications:
- Iterating over text/code blocks (e.g., llama.vim)
- Reasoning models (repeating thinking in final answer)
- Summarization

Example Video: #19164

### Differences Between ngram-simple, ngram-map and ngram-mod

- **ngram-simple**: finds previous matching n-gram, inserts following m-gram
- **ngram-map-k**: finds previous matching n-gram, inserts following m-gram, uses internal hash-map in current context window
- **ngram-mod**: uses shared hash pool mapping n-gram hash to next token (not next m-gram)

## Command-Line Options

If draft model combined with draftless decoding, draftless has higher precedence.

```bash
--draft, --draft-n, --draft-max N       number of tokens to draft for speculative decoding (default: 16)
                                        (env: LLAMA_ARG_DRAFT_MAX)
--draft-min, --draft-n-min N            minimum number of draft tokens to use for speculative decoding
                                        (default: 0)
                                        (env: LLAMA_ARG_DRAFT_MIN)
[...]
--spec-type [none|ngram-cache|ngram-simple|ngram-map-k|ngram-map-k4v|ngram-mod]
                                        type of speculative decoding to use when no draft model is provided
                                        (default: none)
--spec-ngram-size-n N                   ngram size N for ngram-simple/ngram-map speculative decoding, length
                                        of lookup n-gram (default: 12)
--spec-ngram-size-m N                   ngram size M for ngram-simple/ngram-map speculative decoding, length
                                        of draft m-gram (default: 48)
--spec-ngram-min-hits N                 minimum hits for ngram-map speculative decoding (default: 1)
```

### `--spec-type TYPE`

Speculative decoding type without draft model.

| Type | Description |
|------|-------------|
| `none` | No speculative decoding (default) |
| `ngram-cache` | n-gram cache lookup |
| `ngram-simple` | Simple n-gram pattern matching |
| `ngram-map-k` | n-gram pattern matching with n-gram-keys |
| `ngram-map-k4v` | n-gram pattern matching with keys and up to four values (experimental) |
| `ngram-mod` | Basic ngram hasher with shared pool |

```bash
./llama-server [...] --spec-type ngram-simple
```

### `--spec-ngram-size-n N`

Lookup n-gram size N for map-based speculative decoding. Determines tokens to look back when searching patterns.

### `--spec-ngram-size-m M`

Draft m-gram size M for map-based speculative decoding. Determines tokens to draft on match. Larger values provide more speedup but may reduce acceptance rate.

### `--spec-ngram-min-hits H`

Minimum key occurrences in token history before use as draft (default 1).

## Statistics

Each implementation prints statistics:

```bash
draft acceptance rate = 0.57576 (  171 accepted /   297 generated)
statistics ngram_simple: #calls = 15, #gen drafts = 5, #acc drafts = 5, #gen tokens = 187, #acc tokens = 73
statistics draft: #calls = 10, #gen drafts = 10, #acc drafts = 10, #gen tokens = 110, #acc tokens = 98
```

```bash
draft acceptance rate = 0.70312 (   90 accepted /   128 generated)
statistics ngram_mod: #calls = 810, #gen drafts = 15, #acc drafts = 15, #gen tokens = 960, #acc tokens = 730, dur(b,g,a) = 0.149, 0.347, 0.005 ms
```

```bash
statistics ngram_map_k: #calls(b,g,a) = 6 1690 26, #gen drafts = 26, #acc drafts = 26, #gen tokens = 1248, #acc tokens = 968, dur(b,g,a) = 2.234, 1.427, 0.016 ms
```

Statistics fields:
- `#calls(b,g,a)`: calls of begin (new prompt), generation and accumulation
- `#gen drafts`: drafts generated
- `#acc drafts`: drafts accepted (partially) by main model
- `#gen tokens`: tokens generated (including rejected)
- `#acc tokens`: tokens accepted by main model
- `dur(b,g,a)`: durations of begin (new prompt), generation and accumulation (process acceptance)