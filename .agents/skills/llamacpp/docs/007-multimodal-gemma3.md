---
title: Gemma 3 vision
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal/gemma3.md
source: git
fetched_at: 2026-04-28T09:49:32.6417814-03:00
rendered_js: false
word_count: 61
summary: This document provides instructions on how to build and execute the Gemma 3 vision model using the llama.cpp framework, including steps for model conversion and CLI usage.
tags:
    - gemma-3
    - vision-model
    - gguf
    - llama-cpp
    - model-conversion
    - multimodal
category: tutorial
optimized: true
optimized_at: 2026-04-28T12:00:00Z
---
# Gemma 3 vision

> [!warning]
> This is very experimental, only used for demo purpose.

## Quick started

Use pre-quantized models from [ggml-org](https://huggingface.co/ggml-org):

```bash
# build
cmake -B build
cmake --build build --target llama-mtmd-cli

# alternatively, install from brew (MacOS)
brew install llama.cpp

# run it
llama-mtmd-cli -hf ggml-org/gemma-3-4b-it-GGUF
llama-mtmd-cli -hf ggml-org/gemma-3-12b-it-GGUF
llama-mtmd-cli -hf ggml-org/gemma-3-27b-it-GGUF

# note: 1B model does not support vision
```

## How to get mmproj.gguf?

Add `--mmproj` when converting via `convert_hf_to_gguf.py`:

```bash
cd gemma-3-4b-it
python ../llama.cpp/convert_hf_to_gguf.py --outfile model.gguf --outtype f16 --mmproj .
# output file: mmproj-model.gguf
```

## How to run it?

Required:
- Text model GGUF (convert via `convert_hf_to_gguf.py`)
- mmproj file from above
- An image file

```bash
# build
cmake -B build
cmake --build build --target llama-mtmd-cli

# run it
./build/bin/llama-mtmd-cli -m {text_model}.gguf --mmproj mmproj.gguf --image your_image.jpg
```

#gemma-3 #vision-model #llama-cpp #multimodal
