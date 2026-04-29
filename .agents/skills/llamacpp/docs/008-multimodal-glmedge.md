---
title: GLMV-EDGE
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal/glmedge.md
source: git
fetched_at: 2026-04-28T09:49:33.068380139-03:00
rendered_js: false
word_count: 76
summary: This document provides instructions for building the llama-mtmd-cli binary and converting GLMV-EDGE models from Hugging Face format into GGUF format for use.
tags:
    - glmv-edge
    - gguf-conversion
    - machine-learning
    - model-inference
    - llm
    - cli-usage
category: guide
optimized: true
optimized_at: 2026-04-28T12:00:00Z
---
# GLMV-EDGE

Supports [glm-edge-v-2b](https://huggingface.co/THUDM/glm-edge-v-2b) and [glm-edge-v-5b](https://huggingface.co/THUDM/glm-edge-v-5b).

## Usage

Build the `llama-mtmd-cli` binary, then run:

```sh
./llama-mtmd-cli -m model_path/ggml-model-f16.gguf --mmproj model_path/mmproj-model-f16.gguf
```

> [!tip]
> Use a lower temperature (`--temp 0.1`) for better quality. Use `-ngl` for GPU offloading.

## GGUF conversion

1. Clone a GLMV-EDGE model ([2B](https://huggingface.co/THUDM/glm-edge-v-2b) or [5B](https://huggingface.co/THUDM/glm-edge-v-5b)):

```sh
git clone https://huggingface.co/THUDM/glm-edge-v-5b or https://huggingface.co/THUDM/glm-edge-v-2b
```

2. Split model into LLM and projector using `glmedge-surgery.py`:

```sh
python ./tools/mtmd/glmedge-surgery.py -m ../model_path
```

3. Convert image encoder to GGUF:

```sh
python ./tools/mtmd/glmedge-convert-image-encoder-to-gguf.py -m ../model_path --llava-projector ../model_path/glm.projector --output-dir ../model_path
```

4. Convert LLM part to GGUF:

```sh
python convert_hf_to_gguf.py ../model_path
```

Both LLM and image encoder outputs land in `model_path`.

#glmv-edge #gguf-conversion #multimodal #cli-usage
