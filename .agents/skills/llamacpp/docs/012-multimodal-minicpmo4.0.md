---
title: MiniCPM-o 4
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal/minicpmo4.0.md
source: git
fetched_at: 2026-04-28T09:49:38.602345782-03:00
rendered_js: false
word_count: 59
summary: This document provides instructions on building the llama.cpp environment, converting MiniCPM-o 4 PyTorch models to GGUF format, and performing inference.
tags:
    - minicpm-o
    - llama-cpp
    - gguf-conversion
    - llm-quantization
    - model-inference
category: guide
optimized: true
optimized_at: 2026-04-28T12:00:00Z
---
## MiniCPM-o 4

### Prepare models and code

Download [MiniCPM-o-4](https://huggingface.co/openbmb/MiniCPM-o-4) PyTorch model from Hugging Face to `MiniCPM-o-4` folder.

### Build llama.cpp

Readme modification time: 20250206

For differences in usage refer to the official [[027-build#|build documentation]].

```bash
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
cmake -B build
cmake --build build --config Release
```

### Usage of MiniCPM-o 4

Convert PyTorch model to GGUF files (pre-converted [gguf](https://huggingface.co/openbmb/MiniCPM-o-4-gguf) also available):

```bash
python ./tools/mtmd/legacy-models/minicpmv-surgery.py -m ../MiniCPM-o-4
python ./tools/mtmd/legacy-models/minicpmv-convert-image-encoder-to-gguf.py -m ../MiniCPM-o-4 --minicpmv-projector ../MiniCPM-o-4/minicpmv.projector --output-dir ../MiniCPM-o-4/ --minicpmv_version 6
python ./convert_hf_to_gguf.py ../MiniCPM-o-4/model

# quantize int4 version
./build/bin/llama-quantize ../MiniCPM-o-4/model/ggml-model-f16.gguf ../MiniCPM-o-4/model/ggml-model-Q4_K_M.gguf Q4_K_M
```

Inference on Linux or Mac:

```bash
# run in single-turn mode
./build/bin/llama-mtmd-cli -m ../MiniCPM-o-4/model/ggml-model-f16.gguf --mmproj ../MiniCPM-o-4/mmproj-model-f16.gguf -c 4096 --temp 0.7 --top-p 0.8 --top-k 100 --repeat-penalty 1.05 --image xx.jpg -p "What is in the image?"

# run in conversation mode
./build/bin/llama-mtmd-cli -m ../MiniCPM-o-4/model/ggml-model-Q4_K_M.gguf --mmproj ../MiniCPM-o-4/mmproj-model-f16.gguf
```

#minicpm-o #llama-cpp #gguf-conversion #model-inference
