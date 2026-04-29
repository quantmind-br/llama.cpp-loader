---
title: Multimodal
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal.md
source: git
fetched_at: 2026-04-28T09:49:43.645302009-03:00
rendered_js: false
word_count: 155
summary: This document explains how to configure and run multimodal models, including image and audio support, using llama.cpp's CLI and server tools.
tags:
    - multimodal
    - llama-cpp
    - computer-vision
    - audio-processing
    - gguf
    - cli-tools
    - machine-learning
category: configuration
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# Multimodal

llama.cpp supports multimodal input via `libmtmd` with **image** and **audio** (experimental) input. Two tools support this: [llama-mtmd-cli](../tools/mtmd/README.md) and [llama-server](../tools/server/README.md) (via `/chat/completions`).

## Setup

Two methods to enable multimodal:

| Method | Flags |
|--------|-------|
| Auto-download via `-hf` | Use a supported model. Disable with `--no-mmproj`. Custom projector: `--mmproj local_file.gguf` |
| Local files | `-m model.gguf --mmproj file.gguf` |

Multimodal projector offloads to GPU by default. Disable with `--no-mmproj-offload`.

```sh
# CLI
llama-mtmd-cli -hf ggml-org/gemma-3-4b-it-GGUF

# Server (auto-download)
llama-server -hf ggml-org/gemma-3-4b-it-GGUF

# Server (local files)
llama-server -m gemma-3-4b-it-Q4_K_M.gguf --mmproj mmproj-gemma-3-4b-it-Q4_K_M.gguf

# No GPU offload
llama-server -hf ggml-org/gemma-3-4b-it-GGUF --no-mmproj-offload
```

> [!warning]
> OCR models are trained with specific prompt/input structures. Refer to:
> - [PaddleOCR-VL](https://github.com/ggml-org/llama.cpp/pull/18825) | [GLM-OCR](https://github.com/ggml-org/llama.cpp/pull/19677) | [Deepseek-OCR](https://github.com/ggml-org/llama.cpp/pull/17400) | [Dots.OCR](https://github.com/ggml-org/llama.cpp/pull/17575) | [HunyuanOCR](https://github.com/ggml-org/llama.cpp/pull/21395)

## Pre-quantized Models

Most models come with `Q4_K_M` quantization. Found at [ggml-org multimodal GGUFs](https://huggingface.co/collections/ggml-org/multimodal-ggufs-68244e01ff1f39e5bebeeedc). Replace `(tool_name)` with `llama-mtmd-cli` or `llama-server`.

> [!note]
> Some models require large context windows, e.g., `-c 8192`.

### Vision Models

```sh
# Gemma 3
(tool_name) -hf ggml-org/gemma-3-4b-it-GGUF
(tool_name) -hf ggml-org/gemma-3-12b-it-GGUF
(tool_name) -hf ggml-org/gemma-3-27b-it-GGUF

# SmolVLM
(tool_name) -hf ggml-org/SmolVLM-Instruct-GGUF
(tool_name) -hf ggml-org/SmolVLM-256M-Instruct-GGUF
(tool_name) -hf ggml-org/SmolVLM-500M-Instruct-GGUF
(tool_name) -hf ggml-org/SmolVLM2-2.2B-Instruct-GGUF
(tool_name) -hf ggml-org/SmolVLM2-256M-Video-Instruct-GGUF
(tool_name) -hf ggml-org/SmolVLM2-500M-Video-Instruct-GGUF

# Pixtral 12B
(tool_name) -hf ggml-org/pixtral-12b-GGUF

# Qwen 2 VL
(tool_name) -hf ggml-org/Qwen2-VL-2B-Instruct-GGUF
(tool_name) -hf ggml-org/Qwen2-VL-7B-Instruct-GGUF

# Qwen 2.5 VL
(tool_name) -hf ggml-org/Qwen2.5-VL-3B-Instruct-GGUF
(tool_name) -hf ggml-org/Qwen2.5-VL-7B-Instruct-GGUF
(tool_name) -hf ggml-org/Qwen2.5-VL-32B-Instruct-GGUF
(tool_name) -hf ggml-org/Qwen2.5-VL-72B-Instruct-GGUF

# Mistral Small 3.1 24B (IQ2_M quantization)
(tool_name) -hf ggml-org/Mistral-Small-3.1-24B-Instruct-2503-GGUF

# InternVL 2.5 and 3
(tool_name) -hf ggml-org/InternVL2_5-1B-GGUF
(tool_name) -hf ggml-org/InternVL2_5-4B-GGUF
(tool_name) -hf ggml-org/InternVL3-1B-Instruct-GGUF
(tool_name) -hf ggml-org/InternVL3-2B-Instruct-GGUF
(tool_name) -hf ggml-org/InternVL3-8B-Instruct-GGUF
(tool_name) -hf ggml-org/InternVL3-14B-Instruct-GGUF

# Llama 4 Scout
(tool_name) -hf ggml-org/Llama-4-Scout-17B-16E-Instruct-GGUF

# Moondream2 20250414
(tool_name) -hf ggml-org/moondream2-20250414-GGUF

# Gemma 4
(tool_name) -hf ggml-org/gemma-4-E2B-it-GGUF
(tool_name) -hf ggml-org/gemma-4-E4B-it-GGUF
(tool_name) -hf ggml-org/gemma-4-26B-A4B-it-GGUF
(tool_name) -hf ggml-org/gemma-4-31B-it-GGUF
```

### Audio Models

```sh
# Ultravox 0.5
(tool_name) -hf ggml-org/ultravox-v0_5-llama-3_2-1b-GGUF
(tool_name) -hf ggml-org/ultravox-v0_5-llama-3_1-8b-GGUF

# Qwen2-Audio / SeaLLM-Audio: no pre-quantized GGUF (poor results)
# ref: https://github.com/ggml-org/llama.cpp/pull/13760

# Voxtral
(tool_name) -hf ggml-org/Voxtral-Mini-3B-2507-GGUF

# Qwen3-ASR
(tool_name) -hf ggml-org/Qwen3-ASR-0.6B-GGUF
(tool_name) -hf ggml-org/Qwen3-ASR-1.7B-GGUF
```

### Mixed Modalities

```sh
# Qwen2.5 Omni (audio + vision)
(tool_name) -hf ggml-org/Qwen2.5-Omni-3B-GGUF
(tool_name) -hf ggml-org/Qwen2.5-Omni-7B-GGUF

# Qwen3 Omni (audio + vision)
(tool_name) -hf ggml-org/Qwen3-Omni-30B-A3B-Instruct-GGUF
(tool_name) -hf ggml-org/Qwen3-Omni-30B-A3B-Thinking-GGUF

# Gemma 4 (audio + vision)
(tool_name) -hf ggml-org/gemma-4-E2B-it-GGUF
(tool_name) -hf ggml-org/gemma-4-E4B-it-GGUF
```

## Finding More Models

Browse GGUF vision models on Hugging Face: [image-text-to-text GGUFs](https://huggingface.co/models?pipeline_tag=image-text-to-text&sort=trending&search=gguf)

#multimodal #computer-vision #audio-processing #gguf #cli-tools
