---
title: Docker
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/docker.md
source: git
fetched_at: 2026-04-28T09:49:27.293470443-03:00
rendered_js: false
word_count: 284
summary: This document outlines the available Docker images for llama.cpp, providing instructions on how to run model conversion, inference, and server deployment using different hardware acceleration configurations.
tags:
    - docker
    - llama-cpp
    - containerization
    - machine-learning
    - gpu-acceleration
    - inference
category: guide
optimized: true
optimized_at: '2026-04-28T12:00:00Z'
---
# Docker

## Prerequisites

- Docker installed and running.
- A folder for big models & intermediate files (e.g. `/llama/models`).

## Images

Three base images (platforms: `linux/amd64`, `linux/arm64`, `linux/s390x`):

| Image | Contents |
|---|---|
| `ghcr.io/ggml-org/llama.cpp:full` | `llama-cli`, `llama-completion`, conversion tools, 4-bit quantization |
| `ghcr.io/ggml-org/llama.cpp:light` | `llama-cli`, `llama-completion` only |
| `ghcr.io/ggml-org/llama.cpp:server` | `llama-server` only |

GPU-accelerated variants (append the suffix to the base image name):

| Suffix | Backend | Platforms |
|---|---|---|
| `-cuda` | CUDA 12 | `linux/amd64`, `linux/arm64` |
| `-cuda13` | CUDA 13 | `linux/amd64`, `linux/arm64` |
| `-rocm` | ROCm | `linux/amd64` |
| `-musa` | MUSA | `linux/amd64` |
| `-intel` | SYCL | `linux/amd64` |
| `-vulkan` | Vulkan | `linux/amd64`, `linux/arm64` |
| `-openvino` | OpenVino | `linux/amd64` |
| `-s390x` | s390x alias | `linux/s390x` |

All GPU images: `full-*`, `light-*`, `server-*`.

GPU images are built from the same Dockerfiles in `.devops/` and GitHub Action in `.github/workflows/docker.yml` without variation. For different settings (e.g. different CUDA/ROCm/MUSA library versions), build locally.

## Usage

### All-in-one (full image)

Download, convert, and optimize models in one command:

```bash
docker run -v /path/to/models:/models ghcr.io/ggml-org/llama.cpp:full --all-in-one "/models/" 7B
```

Then run:

```bash
docker run -v /path/to/models:/models ghcr.io/ggml-org/llama.cpp:full --run -m /models/7B/ggml-model-q4_0.gguf
docker run -v /path/to/models:/models ghcr.io/ggml-org/llama.cpp:full --run-legacy -m /models/32B/ggml-model-q8_0.gguf -no-cnv -p "Building a mobile app can be done in 15 steps:" -n 512
```

### Light image

```bash
docker run -v /path/to/models:/models --entrypoint /app/llama-cli ghcr.io/ggml-org/llama.cpp:light -m /models/7B/ggml-model-q4_0.gguf
docker run -v /path/to/models:/models --entrypoint /app/llama-completion ghcr.io/ggml-org/llama.cpp:light -m /models/32B/ggml-model-q8_0.gguf -no-cnv -p "Building a mobile app can be done in 15 steps:" -n 512
```

### Server image

```bash
docker run -v /path/to/models:/models -p 8080:8080 ghcr.io/ggml-org/llama.cpp:server -m /models/7B/ggml-model-q4_0.gguf --port 8080 --host 0.0.0.0 -n 512
```

> [!tip]
> `--entrypoint /app/llama-cli` is the default entrypoint and can be omitted.

## Docker with CUDA

Requires [nvidia-container-toolkit](https://github.com/NVIDIA/nvidia-container-toolkit) installed on Linux (or a GPU-enabled cloud). `cuBLAS` is then accessible inside the container.

### Build locally

```bash
docker build -t local/llama.cpp:full-cuda --target full -f .devops/cuda.Dockerfile .
docker build -t local/llama.cpp:light-cuda --target light -f .devops/cuda.Dockerfile .
docker build -t local/llama.cpp:server-cuda --target server -f .devops/cuda.Dockerfile .
```

Defaults: `CUDA_VERSION=12.8.1`, `CUDA_DOCKER_ARCH` = cmake default (all supported architectures).

### Run

```bash
docker run --gpus all -v /path/to/models:/models local/llama.cpp:full-cuda --run -m /models/7B/ggml-model-q4_0.gguf -p "Building a website can be done in 10 simple steps:" -n 512 --n-gpu-layers 1
docker run --gpus all -v /path/to/models:/models local/llama.cpp:light-cuda -m /models/7B/ggml-model-q4_0.gguf -p "Building a website can be done in 10 simple steps:" -n 512 --n-gpu-layers 1
docker run --gpus all -v /path/to/models:/models local/llama.cpp:server-cuda -m /models/7B/ggml-model-q4_0.gguf --port 8080 --host 0.0.0.0 -n 512 --n-gpu-layers 1
```

## Docker with MUSA

Requires [mt-container-toolkit](https://developer.mthreads.com/musa/native) installed on Linux. `muBLAS` is then accessible inside the container.

### Build locally

```bash
docker build -t local/llama.cpp:full-musa --target full -f .devops/musa.Dockerfile .
docker build -t local/llama.cpp:light-musa --target light -f .devops/musa.Dockerfile .
docker build -t local/llama.cpp:server-musa --target server -f .devops/musa.Dockerfile .
```

Default: `MUSA_VERSION=rc4.3.0`.

### Run

Set `mthreads` as default Docker runtime: `(cd /usr/bin/musa && sudo ./docker setup $PWD)` and verify with `docker info | grep mthreads`.

```bash
docker run -v /path/to/models:/models local/llama.cpp:full-musa --run -m /models/7B/ggml-model-q4_0.gguf -p "Building a website can be done in 10 simple steps:" -n 512 --n-gpu-layers 1
docker run -v /path/to/models:/models local/llama.cpp:light-musa -m /models/7B/ggml-model-q4_0.gguf -p "Building a website can be done in 10 simple steps:" -n 512 --n-gpu-layers 1
docker run -v /path/to/models:/models local/llama.cpp:server-musa -m /models/7B/ggml-model-q4_0.gguf --port 8080 --host 0.0.0.0 -n 512 --n-gpu-layers 1
```
