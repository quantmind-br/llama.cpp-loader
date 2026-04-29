---
title: OPENVINO
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/OPENVINO.md
source: git
fetched_at: 2026-04-28T09:49:11.139272638-03:00
rendered_js: false
word_count: 979
summary: This document explains how to integrate the OpenVINO backend into llama.cpp to enable hardware-accelerated AI inference on Intel CPUs, GPUs, and NPUs.
tags:
    - openvino
    - llama-cpp
    - intel-hardware
    - ai-inference
    - gguf
    - model-optimization
category: guide
optimized: true
optimized_at: 2026-04-28T00:00:00Z
---
# OpenVINO Backend for llama.cpp

> [!NOTE]
> Performance and memory optimizations, accuracy validation, broader quantization coverage, and broader operator/model support are in progress.

[OpenVINO](https://docs.openvino.ai/) is an open-source toolkit for high-performance AI inference on Intel hardware (CPUs, GPUs, NPUs). The [OpenVINO backend for llama.cpp](../../ggml/src/ggml-openvino) enables hardware-accelerated inference on **Intel CPUs, GPUs, and NPUs** while remaining compatible with the **GGUF model ecosystem**. It translates GGML compute graphs into OpenVINO graphs and leverages graph compilation, kernel fusion, and device-specific optimizations.

Implemented in `ggml/src/ggml-openvino`, the backend replaces the standard GGML graph execution path with Intel's OpenVINO inference engine. The same GGUF model file runs on Intel CPUs, GPUs (integrated and discrete), and NPUs without model or code changes. When a `ggml_cgraph` is dispatched to OpenVINO:

1. Walks the GGML graph and identifies inputs, outputs, weights, and KV cache tensors.
2. Translates GGML operations into an `ov::Model` using OpenVINO's frontend API.
3. Compiles and caches the model for the target device.
4. Binds GGML tensor memory to OpenVINO inference tensors and runs inference.

## Supported Devices

- Intel CPUs
- Intel GPUs (integrated and discrete)
- Intel NPUs

Validated specifically on AI PCs with Intel Core Ultra Series 1 and Series 2. See [OpenVINO hardware support](https://docs.openvino.ai/2026/about-openvino/release-notes-openvino/system-requirements.html) for the full range.

## Supported Model Precisions

| Precision | Notes |
|-----------|-------|
| `FP16` | |
| `BF16` | On Intel Xeon |
| `Q8_0` | |
| `Q4_0` | |
| `Q4_1` | |
| `Q4_K` | |
| `Q4_K_M` | |
| `Q5_K` | Converted to `Q8_0_C` at runtime |
| `Q6_K` | Converted to `Q8_0_C` at runtime |

> [!NOTE]
> Accuracy validation and performance optimizations for quantized models are in progress.

## Quantization Support Details

### CPU and GPU

- Supported: `Q4_0`, `Q4_1`, `Q4_K_M`, `Q6_K`
- `Q5_K` and `Q6_K` tensors are converted to `Q8_0_C`

### NPU

- Primary quantization: `Q4_0`
- `Q6_K` tensors are requantized to `Q4_0_128`. Exception: embedding weights use `Q8_0_C` (token embedding matrix is dequantized to fp16)

### Additional Notes

- Both `Q4_0` and `Q4_1` models use `Q6_K` for the token embedding tensor and the final matmul weight tensor (often the same tensor)
- `Q4_0` models may produce `Q4_1` tensors when an imatrix is provided during quantization with `llama-quantize`
- `Q4_K_M` models may include both `Q6_K` and `Q5_K` tensors (observed in Phi-3)

## Validated Models

Tested on Intel Core Ultra Series 1 and Series 2:

- [Llama-3.2-1B-Instruct-GGUF](https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/)
- [Llama-3.1-8B-Instruct](https://huggingface.co/bartowski/Meta-Llama-3.1-8B-Instruct-GGUF)
- [Phi-3-mini-4k-instruct-gguf](https://huggingface.co/microsoft/Phi-3-mini-4k-instruct-gguf)
- [Qwen2.5-1.5B-Instruct-GGUF](https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF)
- [Qwen3-8B](https://huggingface.co/Qwen/Qwen3-8B-GGUF)
- [MiniCPM-1B-sft-bf16](https://huggingface.co/openbmb/MiniCPM-S-1B-sft-gguf)
- [Hunyuan-7B-Instruct](https://huggingface.co/bartowski/tencent_Hunyuan-7B-Instruct-GGUF)
- [Mistral-7B-Instruct-v0.3](https://huggingface.co/bartowski/Mistral-7B-Instruct-v0.3-GGUF)
- [DeepSeek-R1-Distill-Llama-8B-GGUF](https://huggingface.co/bartowski/DeepSeek-R1-Distill-Llama-8B-GGUF)

## Build Instructions

### Prerequisites

Linux or Windows with Intel hardware (CPU, GPU, or NPU).

**For GPU or NPU:** install appropriate drivers — see [Additional Configurations for Hardware Acceleration](https://docs.openvino.ai/2025/get-started/install-openvino/configurations.html).

**Linux:**

```bash
sudo apt-get update
sudo apt-get install -y build-essential libcurl4-openssl-dev libtbb12 cmake ninja-build python3-pip curl wget tar
sudo apt install ocl-icd-opencl-dev opencl-headers opencl-clhpp-headers intel-opencl-icd
```

**Windows:**

1. Install [Microsoft Visual Studio 2022 Build Tools](https://aka.ms/vs/17/release/vs_BuildTools.exe) with **"Desktop development with C++"** workload.

```powershell
winget install Git.Git
winget install GNU.Wget
winget install Ninja-build.Ninja
```

```powershell
cd C:\
git clone https://github.com/microsoft/vcpkg
cd vcpkg
.\bootstrap-vcpkg.bat
.\vcpkg install opencl
.\vcpkg integrate install
```

### 1. Install OpenVINO Runtime

Install from archive: [Linux](https://docs.openvino.ai/2026/get-started/install-openvino/install-openvino-archive-linux.html) | [Windows](https://docs.openvino.ai/2026/get-started/install-openvino/install-openvino-archive-windows.html)

<details>
<summary>📦 OpenVINO installation from archive on Ubuntu</summary>

```bash
wget https://raw.githubusercontent.com/ravi9/misc-scripts/main/openvino/ov-archive-install/install-openvino-from-archive.sh
chmod +x install-openvino-from-archive.sh
./install-openvino-from-archive.sh
```

Verify:

```bash
echo $OpenVINO_DIR
```
</details>

### 2. Build llama.cpp with OpenVINO Backend

```bash
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
```

**Linux:**

```bash
source /opt/intel/openvino/setupvars.sh
cmake -B build/ReleaseOV -G Ninja -DCMAKE_BUILD_TYPE=Release -DGGML_OPENVINO=ON
cmake --build build/ReleaseOV --parallel
```

**Windows** (use `x64 Native Tools Command Prompt for VS 2022`):

```cmd
"C:\Program Files (x86)\Intel\openvino_2026.0\setupvars.bat"
cmake -B build\ReleaseOV -G Ninja -DCMAKE_BUILD_TYPE=Release -DGGML_OPENVINO=ON -DLLAMA_CURL=OFF -DCMAKE_TOOLCHAIN_FILE=C:\vcpkg\scripts\buildsystems\vcpkg.cmake
cmake --build build\ReleaseOV --parallel
```

> [!NOTE]
> Use `x64 Native Tools Command Prompt` for Windows build. After building, either `cmd` or PowerShell can run the OpenVINO backend.

### 3. Download Sample Model

```bash
# Linux
mkdir -p ~/models/
wget https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/main/Llama-3.2-1B-Instruct-Q4_0.gguf \
     -O ~/models/Llama-3.2-1B-Instruct-Q4_0.gguf

# Windows PowerShell
mkdir C:\models
Invoke-WebRequest -Uri https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/main/Llama-3.2-1B-Instruct-Q4_0.gguf -OutFile C:\models\Llama-3.2-1B-Instruct-Q4_0.gguf

# Windows Command Line
mkdir C:\models
curl -L https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/main/Llama-3.2-1B-Instruct-Q4_0.gguf -o C:\models\Llama-3.2-1B-Instruct-Q4_0.gguf
```

### 4. Run Inference

First inference token may have higher latency due to on-the-fly graph conversion. Subsequent tokens are faster.

> [!TIP]
> Default context size equals the model training context (e.g. 131072 for Llama 3.2 1B), which may cause low performance on edge/laptop devices. Use `-c` to limit context (e.g. `-c 512`).

```bash
# Linux
export GGML_OPENVINO_DEVICE=GPU
export GGML_OPENVINO_STATEFUL_EXECUTION=1
./build/ReleaseOV/bin/llama-simple -m ~/models/Llama-3.2-1B-Instruct-Q4_0.gguf -n 50 "The story of AI is "
./build/ReleaseOV/bin/llama-cli -m ~/models/Llama-3.2-1B-Instruct-Q4_0.gguf -c 1024
GGML_OPENVINO_STATEFUL_EXECUTION=1 GGML_OPENVINO_DEVICE=GPU ./build/ReleaseOV/bin/llama-bench -m ~/models/Llama-3.2-1B-Instruct-Q4_0.gguf -fa 1

# NPU (keep context small)
export GGML_OPENVINO_DEVICE=NPU
./build/ReleaseOV/bin/llama-cli -m ~/models/Llama-3.2-1B-Instruct-Q4_0.gguf -c 512

# Windows PowerShell
$env:GGML_OPENVINO_DEVICE = "GPU"
$env:GGML_OPENVINO_STATEFUL_EXECUTION = "1"
build\ReleaseOV\bin\llama-simple.exe -m "C:\models\Llama-3.2-1B-Instruct-Q4_0.gguf" -n 50 "The story of AI is "
build\ReleaseOV\bin\llama-cli.exe -m "C:\models\Llama-3.2-1B-Instruct-Q4_0.gguf" -c 1024
build\ReleaseOV\bin\llama-bench.exe -m "C:\models\Llama-3.2-1B-Instruct-Q4_0.gguf" -fa 1

# Windows NPU
$env:GGML_OPENVINO_DEVICE = "NPU"
build\ReleaseOV\bin\llama-cli.exe -m "C:\models\Llama-3.2-1B-Instruct-Q4_0.gguf" -c 512
```

> [!NOTE]
> On multi-GPU systems, use `GPU.0` or `GPU.1` to target a specific GPU. See [OpenVINO GPU Device](https://docs.openvino.ai/2026/openvino-workflow/running-inference/inference-devices-and-modes/gpu-device.html).

### Known Issues and Workarounds

| Issue | Workaround |
|-------|------------|
| GPU stateless execution has a known issue | Set `GGML_OPENVINO_STATEFUL_EXECUTION=1` with GPU |
| NPU failures with large context size | Set context explicitly (e.g. `-c 1024`). Inspect with `-lv 3`. |
| NPU: no model caching support | — |
| NPU: `llama-server -np > 1` not supported | — |
| NPU: `llama-perplexity` only with `-b 512` or smaller | — |
| `--context-shift` not supported with OpenVINO backend | — |
| Encoder models (embedding, reranking) not supported | — |
| `-fa 1` required for `llama-bench` with OpenVINO | `GGML_OPENVINO_STATEFUL_EXECUTION=1 GGML_OPENVINO_DEVICE=GPU ./llama-bench -fa 1` |
| `llama-server` supports only one chat session when `GGML_OPENVINO_STATEFUL_EXECUTION=1` | — |

> [!NOTE]
> The OpenVINO backend is under active development. This document will be updated as issues are resolved.

### Docker Build

```bash
# Base runtime image
docker build -t llama-openvino:base -f .devops/openvino.Dockerfile .

# Full image with all binaries and tools
docker build --target=full -t llama-openvino:full -f .devops/openvino.Dockerfile .

# Minimal CLI-only image
docker build --target=light -t llama-openvino:light -f .devops/openvino.Dockerfile .

# Server-only image
docker build --target=server -t llama-openvino:server -f .devops/openvino.Dockerfile .

# Behind a proxy
docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy --target=light -t llama-openvino:light -f .devops/openvino.Dockerfile .
```

Save models in `~/models` (mounted in examples below).

```bash
# CPU
docker run --rm -it -v ~/models:/models llama-openvino:light --no-warmup -c 1024 -m /models/Llama-3.2-1B-Instruct-Q4_0.gguf

# Intel GPU
docker run --rm -it -v ~/models:/models \
--device=/dev/dri --group-add=$(stat -c "%g" /dev/dri/render* | head -n 1) -u $(id -u):$(id -g) \
--env=GGML_OPENVINO_DEVICE=GPU --env=GGML_OPENVINO_STATEFUL_EXECUTION=1 \
llama-openvino:light --no-warmup -c 1024 -m /models/Llama-3.2-1B-Instruct-Q4_0.gguf

# Intel NPU
docker run --rm -it -v ~/models:/models \
--device=/dev/accel --group-add=$(stat -c "%g" /dev/dri/render* | head -n 1) -u $(id -u):$(id -g) \
--env=GGML_OPENVINO_DEVICE=NPU \
llama-openvino:light --no-warmup -c 1024 -m /models/Llama-3.2-1B-Instruct-Q4_0.gguf
```

### Docker Server

> [!NOTE]
> `llama-server` supports only one chat session/thread when `GGML_OPENVINO_STATEFUL_EXECUTION=1`.

```bash
docker run --rm -it -p 8080:8080 -v ~/models:/models llama-openvino:server --no-warmup -m /models/Llama-3.2-1B-Instruct-Q4_0.gguf -c 1024
# Or directly:
./build/ReleaseOV/bin/llama-server -m ~/models/Llama-3.2-1B-Instruct-Q4_0.gguf --port 8080 -c 1024

# Behind a proxy, set NO_PROXY for localhost
export NO_PROXY=localhost,127.0.0.1

# Web UI: http://localhost:8080
# Or test with curl:
curl -f http://localhost:8080/health
curl -X POST "http://localhost:8080/v1/chat/completions" -H "Content-Type: application/json" \
 -d '{"messages":[{"role":"user","content":"Write a poem about OpenVINO"}],"max_tokens":100}' | jq .
```

## Runtime Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GGML_OPENVINO_DEVICE` | `CPU` | Target device: `CPU`, `GPU`, `NPU`. Use `GPU.0`/`GPU.1` for multi-GPU. See [OpenVINO GPU Device](https://docs.openvino.ai/2026/openvino-workflow/running-inference/inference-devices-and-modes/gpu-device.html). **NPU** enables static compilation mode. |
| `GGML_OPENVINO_CACHE_DIR` | not set | Model caching directory (recommended: `/tmp/ov_cache`). **Not supported on NPU.** |
| `GGML_OPENVINO_PREFILL_CHUNK_SIZE` | `256` | Token chunk size for **NPU** prefill. |
| `GGML_OPENVINO_STATEFUL_EXECUTION` | `0` | Enable stateful KV cache. Recommended for CPU, GPU. Experimental — validated with llama-simple, llama-cli, llama-bench, llama-run only. Not effective on NPU. Not all models supported. |
| `GGML_OPENVINO_PROFILING` | `0` | Enable execution-time profiling. |
| `GGML_OPENVINO_DUMP_CGRAPH` | `0` | Dump GGML compute graph to `cgraph_ov.txt`. |
| `GGML_OPENVINO_DUMP_IR` | `0` | Serialize OpenVINO IR files with timestamps. |
| `GGML_OPENVINO_DEBUG_INPUT` | `0` | Enable input debugging and print input tensor info. |
| `GGML_OPENVINO_DEBUG_OUTPUT` | `0` | Enable output debugging and print output tensor info. |
| `GGML_OPENVINO_PRINT_CGRAPH_TENSOR_ADDRESS` | `0` | Print tensor address map once. |

### Example: GPU Inference with Profiling

```bash
# Linux
export GGML_OPENVINO_CACHE_DIR=/tmp/ov_cache
export GGML_OPENVINO_PROFILING=1
export GGML_OPENVINO_DEVICE=GPU
export GGML_OPENVINO_STATEFUL_EXECUTION=1
./build/ReleaseOV/bin/llama-simple -m ~/models/Llama-3.2-1B-Instruct-Q4_0.gguf -n 50 "The story of AI is "

# Windows PowerShell
$env:GGML_OPENVINO_CACHE_DIR = "C:\tmp\ov_cache"
$env:GGML_OPENVINO_PROFILING = "1"
$env:GGML_OPENVINO_DEVICE = "GPU"
$env:GGML_OPENVINO_STATEFUL_EXECUTION = "1"
build\ReleaseOV\bin\llama-simple.exe -m "C:\models\Llama-3.2-1B-Instruct-Q4_0.gguf" -n 50 "The story of AI is "
```

## Compatible Tools

Works with OpenVINO backend on CPU, GPU, NPU: llama-bench, llama-cli, llama-completion, llama-perplexity, llama-server, llama-simple.

## Work in Progress

- Performance and memory optimizations
- Accuracy validation
- Broader quantization coverage
- Support for additional model architectures

#openvino #intel-hardware #ai-inference #gguf
