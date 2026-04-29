---
title: Preset
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/preset.md
source: git
fetched_at: 2026-04-28T09:49:45.267596353-03:00
rendered_js: false
word_count: 176
summary: This document explains how to use INI preset files to manage and share reusable parameter configurations for llama.cpp models, including local and remote Hugging Face implementations.
tags:
    - llama-cpp
    - ini-presets
    - configuration-management
    - hugging-face
    - model-parameters
    - cli-tools
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# llama.cpp INI Presets

INI presets ([PR#17859](https://github.com/ggml-org/llama.cpp/pull/17859)) create reusable, shareable parameter configurations for llama.cpp.

## Using presets with the server

When running multiple models (router mode), INI preset files configure model-specific parameters. See the [server documentation](../tools/server/README.md) for details.

## Using a remote preset

> [!note]
> Remote presets are only supported via the `-hf` option.

For GGUF models on Hugging Face, include a `preset.ini` in the repository root:

```ini
hf-repo-draft = username/my-draft-model-GGUF
temp = 0.5
top-k = 20
top-p = 0.95
```

Only certain options are allowed for security — see [preset.cpp](../common/preset.cpp) for the complete list.

Usage example with repo `username/my-model-with-preset` containing the above `preset.ini`:

```sh
llama-cli -hf username/my-model-with-preset

# Equivalent to:
llama-cli -hf username/my-model-with-preset \
  --hf-repo-draft username/my-draft-model-GGUF \
  --temp 0.5 \
  --top-k 20 \
  --top-p 0.95
```

Override preset values by specifying them on the command line:

```sh
llama-cli -hf username/my-model-with-preset --temp 0.1
```

### Multiple presets via separate HF repos

Create a blank HF repo per preset, each with a `preset.ini` referencing actual models:

```ini
hf-repo = user/my-model-main
hf-repo-draft = user/my-model-draft
temp = 0.8
ctx-size = 1024
```

### Named presets

Define multiple named configurations in a single `preset.ini` with `[*]` for defaults and `[section-name]` for named presets:

```ini
[*]
mmap = 1

[gpt-oss-20b-hf]
hf          = ggml-org/gpt-oss-20b-GGUF
batch-size  = 2048
ubatch-size = 2048
top-p       = 1.0
top-k       = 0
min-p       = 0.01
temp        = 1.0
chat-template-kwargs = {"reasoning_effort": "high"}

[gpt-oss-120b-hf]
hf          = ggml-org/gpt-oss-120b-GGUF
batch-size  = 2048
ubatch-size = 2048
top-p       = 1.0
top-k       = 0
min-p       = 0.01
temp        = 1.0
chat-template-kwargs = {"reasoning_effort": "high"}
```

Select a named preset with the tag syntax:

```sh
llama-server -hf user/repo:gpt-oss-120b-hf
```

> [!warning]
> Each child preset must have the correct `hf-repo`. Otherwise you'll get: `The specified tag is not a valid quantization scheme.`

#ini-presets #configuration-management #hugging-face