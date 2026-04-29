---
title: Llguidance
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/llguidance.md
source: git
fetched_at: 2026-04-28T09:49:30.834634122-03:00
rendered_js: false
word_count: 312
summary: This document explains how to integrate and use the LLGuidance library within llama.cpp to enable high-performance constrained decoding and structured output generation.
tags:
    - llguidance
    - llama-cpp
    - constrained-decoding
    - json-schema
    - structured-outputs
    - grammar-parsing
    - lexer
category: configuration
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# LLGuidance Support in llama.cpp

[LLGuidance](https://github.com/guidance-ai/llguidance) is a constrained-decoding library for LLMs. Originally the backend for [Guidance](https://github.com/guidance-ai/guidance), it works standalone.

- Supports JSON Schemas and arbitrary CFGs written in a [Lark-syntax variant](https://github.com/guidance-ai/llguidance/blob/main/docs/syntax.md).
- [Very fast](https://github.com/guidance-ai/jsonschemabench/tree/main/maskbench) with [excellent](https://github.com/guidance-ai/llguidance/blob/main/docs/json_schema.md) JSON Schema coverage.
- Requires the Rust compiler, which complicates the llama.cpp build.

## Building

Enable with the `LLAMA_LLGUIDANCE` CMake option. Requires [Rust/cargo](https://www.rust-lang.org/tools/install).

```sh
cmake -B build -DLLAMA_LLGUIDANCE=ON
make -C build -j
```

For Windows: `cmake --build build --config Release` instead of `make`.

## Interface

No new CLI arguments or `common_params` changes. When enabled:
- Grammars starting with `%llguidance` → LLGuidance (instead of [current](../grammars/README.md) llama.cpp grammars).
- JSON Schema requests (e.g., `-j` in `llama-cli`) → LLGuidance.

Convert existing GBNF grammars via [gbnf_to_lark.py](https://github.com/guidance-ai/llguidance/blob/main/python/llguidance/gbnf_to_lark.py).

## Performance

For a llama3 tokenizer (128k tokens), computing a token mask averages **50μs** single-core CPU for the [JSON Schema Bench](https://github.com/guidance-ai/jsonschemabench). p99 = 0.5ms, p100 = 20ms. Results come from the lexer/parser split and [optimizations](https://github.com/guidance-ai/llguidance/blob/main/docs/optimizations.md).

## JSON Schema Compliance

Adheres closely to the JSON Schema specification:

- `additionalProperties` defaults to `true` (unlike current grammars). Set `"additionalProperties": false` if needed.
- Any whitespace is allowed.
- Property definition order in `"properties": {}` is maintained regardless of required status (current grammars always put required first).
- Unsupported schemas → error message. No keywords are silently ignored.

## Why Not Reuse GBNF?

GBNF lacks a lexer concept.

Most languages (including JSON) use a two-step process: a lexer (regex-based) converts bytes → lexemes, then a CFG parser processes lexemes. This is faster because lexers are cheaper, and there are ~10x fewer lexemes than bytes. LLM tokens often align with lexemes, so the parser is engaged <0.5% of the time.

The user must distinguish lexemes from CFG symbols (uppercase = lexemes, lowercase = symbols in [Lark](https://github.com/lark-parser/lark)). The [gbnf_to_lark.py script](https://github.com/guidance-ai/llguidance/blob/main/scripts/gbnf_to_lark.py) often handles this automatically. See [syntax docs](https://github.com/guidance-ai/llguidance/blob/main/docs/syntax.md#terminals-vs-rules).

## Error Handling

Errors print to `stderr` and generation continues. Improved error handling may be added later.

#llguidance #constrained-decoding #json-schema #structured-outputs #grammar-parsing
