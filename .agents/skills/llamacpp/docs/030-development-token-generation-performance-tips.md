---
title: Token generation performance tips
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/development/token_generation_performance_tips.md
source: git
fetched_at: 2026-04-28T09:49:25.977674464-03:00
rendered_js: false
word_count: 224
summary: This document provides troubleshooting steps to improve token generation performance by verifying GPU offloading via CUDA and optimizing CPU thread count settings.
tags:
    - performance-tuning
    - gpu-acceleration
    - cuda-offloading
    - token-generation
    - inference-speed
    - llama-optimization
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# Token generation performance troubleshooting

## Verify GPU offloading with CUDA

Compile llama.cpp with correct env variables per [[027-build#cuda|this guide]], so llama accepts `-ngl N` (`--n-gpu-layers N`). Set `N` very large to offload the maximum possible layers:

```shell
./llama-cli -m "path/to/model.gguf" -ngl 200000 -p "Please sir, may I have some "
```

Before inference starts, look for these diagnostic lines confirming GPU offload:

```shell
llama_model_load_internal: [cublas] offloading 60 layers to GPU
llama_model_load_internal: [cublas] offloading output layer to GPU
llama_model_load_internal: [cublas] total VRAM used: 17223 MB
... rest of inference
```

If these lines appear, the GPU is being used.

## Avoid CPU oversaturation

The `-t N` (`--threads N`) parameter controls CPU thread count. If token generation is extremely slow, try `-t 1`. If that significantly improves speed, your CPU is oversaturated — set `-t` to the number of **physical CPU cores** (even when using a GPU).

> [!tip]
> Start with `-t 1`, double until you hit a bottleneck, then scale down.

## Benchmark: runtime flags effect on inference speed

**Machine specs:**

| Component | Value |
| --- | --- |
| GPU | A6000 (48GB VRAM) |
| CPU | 7 physical cores |
| RAM | 32GB |

**Model:** `TheBloke_Wizard-Vicuna-30B-Uncensored-GGML/Wizard-Vicuna-30B-Uncensored.q4_0.gguf` (30B params, 4-bit quantization, GGML)

**Run command:**

```shell
./llama-cli -m "path/to/model.gguf" -p "An extremely detailed description of the 10 best ethnic dishes will follow, with recipes: " -n 1000 [additional benchmark flags]
```

**Results:**

| Command | tokens/second (higher is better) |
| - | - |
| -ngl 2000000 | N/A (< 0.1) |
| -t 7 | 1.7 |
| -t 1 -ngl 2000000 | 5.5 |
| -t 7 -ngl 2000000 | 8.7 |
| -t 4 -ngl 2000000 | 9.1 |

#performance-tuning #gpu-acceleration #cuda-offloading #token-generation