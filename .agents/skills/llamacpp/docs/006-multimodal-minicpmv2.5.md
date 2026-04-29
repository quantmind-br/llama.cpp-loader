---
title: MiniCPM-Llama3-V 2.5
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal/minicpmv2.5.md
source: git
fetched_at: 2026-04-28T09:49:38.933691373-03:00
rendered_js: false
word_count: 59
summary: This document provides instructions for building llama.cpp and converting, quantizing, and running inference with the MiniCPM-Llama3-V 2.5 multimodal model.
tags:
    - minicpm
    - llama-cpp
    - model-conversion
    - quantization
    - gguf
    - multimodal-ai
    - inference
category: guide
optimized: true
optimized_at: 2026-04-28T12:00:00Z
---
## MiniCPM-Llama3-V 2.5

### Prepare models and code

Download [MiniCPM-Llama3-V-2_5](https://huggingface.co/openbmb/MiniCPM-Llama3-V-2_5) PyTorch model from Hugging Face to `MiniCPM-Llama3-V-2_5` folder.

### Build llama.cpp

Readme modification time: 20250206

For differences in usage refer to the official [[027-build#|build documentation]].

```bash
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
cmake -B build
cmake --build build --config Release
```

### Usage of MiniCPM-Llama3-V 2.5

Convert PyTorch model to GGUF files (pre-converted [gguf](https://huggingface.co/openbmb/MiniCPM-Llama3-V-2_5-gguf) also available):

```bash
python ./tools/mtmd/legacy-models/minicpmv-surgery.py -m ../MiniCPM-Llama3-V-2_5
python ./tools/mtmd/legacy-models/minicpmv-convert-image-encoder-to-gguf.py -m ../MiniCPM-Llama3-V-2_5 --minicpmv-projector ../MiniCPM-Llama3-V-2_5/minicpmv.projector --output-dir ../MiniCPM-Llama3-V-2_5/ --minicpmv_version 2
python ./convert_hf_to_gguf.py ../MiniCPM-Llama3-V-2_5/model

# quantize int4 version
./build/bin/llama-quantize ../MiniCPM-Llama3-V-2_5/model/model-8B-F16.gguf ../MiniCPM-Llama3-V-2_5/model/ggml-model-Q4_K_M.gguf Q4_K_M
```

Inference on Linux or Mac:

```bash
# run in single-turn mode
./build/bin/llama-mtmd-cli -m ../MiniCPM-Llama3-V-2_5/model/model-8B-F16.gguf --mmproj ../MiniCPM-Llama3-V-2_5/mmproj-model-f16.gguf -c 4096 --temp 0.7 --top-p 0.8 --top-k 100 --repeat-penalty 1.05 --image xx.jpg -p "What is in the image?"

# run in conversation mode
./build/bin/llama-mtmd-cli -m ../MiniCPM-Llama3-V-2_5/model/ggml-model-Q4_K_M.gguf --mmproj ../MiniCPM-Llama3-V-2_5/mmproj-model-f16.gguf
```

#minicpm #llama-cpp #model-conversion #multimodal-ai
