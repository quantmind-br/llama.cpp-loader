---
title: MiniCPM-V 2.6
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal/minicpmv2.6.md
source: git
fetched_at: 2026-04-28T09:49:40.941019833-03:00
rendered_js: false
word_count: 59
summary: This document provides instructions for setting up, converting, and running the MiniCPM-V 2.6 vision model using the llama.cpp framework.
tags:
    - minicpm-v
    - llama-cpp
    - gguf
    - model-conversion
    - quantization
    - multimodal-ai
category: guide
optimized: true
optimized_at: 2026-04-28T12:00:00Z
---
## MiniCPM-V 2.6

### Prepare models and code

Download [MiniCPM-V-2_6](https://huggingface.co/openbmb/MiniCPM-V-2_6) PyTorch model from Hugging Face to `MiniCPM-V-2_6` folder.

### Build llama.cpp

Readme modification time: 20250206

For differences in usage refer to the official [[027-build#|build documentation]].

```bash
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
cmake -B build
cmake --build build --config Release
```

### Usage of MiniCPM-V 2.6

Convert PyTorch model to GGUF files (pre-converted [gguf](https://huggingface.co/openbmb/MiniCPM-V-2_6-gguf) also available):

```bash
python ./tools/mtmd/legacy-models/minicpmv-surgery.py -m ../MiniCPM-V-2_6
python ./tools/mtmd/legacy-models/minicpmv-convert-image-encoder-to-gguf.py -m ../MiniCPM-V-2_6 --minicpmv-projector ../MiniCPM-V-2_6/minicpmv.projector --output-dir ../MiniCPM-V-2_6/ --minicpmv_version 3
python ./convert_hf_to_gguf.py ../MiniCPM-V-2_6/model

# quantize int4 version
./build/bin/llama-quantize ../MiniCPM-V-2_6/model/ggml-model-f16.gguf ../MiniCPM-V-2_6/model/ggml-model-Q4_K_M.gguf Q4_K_M
```

Inference on Linux or Mac:

```bash
# run in single-turn mode
./build/bin/llama-mtmd-cli -m ../MiniCPM-V-2_6/model/ggml-model-f16.gguf --mmproj ../MiniCPM-V-2_6/mmproj-model-f16.gguf -c 4096 --temp 0.7 --top-p 0.8 --top-k 100 --repeat-penalty 1.05 --image xx.jpg -p "What is in the image?"

# run in conversation mode
./build/bin/llama-mtmd-cli -m ../MiniCPM-V-2_6/model/ggml-model-Q4_K_M.gguf --mmproj ../MiniCPM-V-2_6/mmproj-model-f16.gguf
```

#minicpm-v #llama-cpp #gguf #multimodal-ai
