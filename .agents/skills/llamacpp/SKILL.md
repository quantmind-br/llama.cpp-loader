---
name: llamacpp
description: |
  llama.cpp - C/C++ framework for LLM inference with GGUF models.
  Use when running, building, or optimizing LLM inference locally with llama.cpp.
  Keywords: llama.cpp, gguf, llm, inference, quantization, cuda, metal, vulkan, rocm, hip, opencl, multimodal, function-calling, speculative-decoding, build, cmake.
compatibility: C/C++, CMake, GGUF models
metadata:
  source: https://github.com/ggml-org/llama.cpp/
  total_docs: 43
  generated: 2026-04-28
---

# llama.cpp

> C/C++ framework for LLM inference using GGUF format. Supports CPU, GPU (CUDA, Metal, Vulkan, ROCm, OpenCL), and NPU backends with multimodal and function calling capabilities.

## Quick Start

```bash
# Clone and build (CPU only)
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
cmake -B build
cmake --build build --config Release

# Run inference
./build/bin/llama-cli -m <model.gguf> -p "Hello, how are you?" -n 50

# Start OpenAI-compatible server
./build/bin/llama-server -m <model.gguf> --host 0.0.0.0 --port 8080
```

## Installation

### Pre-built packages

```bash
# Windows
winget install llama.cpp

# macOS / Linux
brew install llama.cpp

# macOS (MacPorts)
sudo port install llama.cpp

# Nix
nix profile install nixpkgs#llama-cpp
```

### From source

```bash
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
cmake -B build
cmake --build build --config Release
```

## Documentation

Documentação completa em `docs/`. Consulte `docs/000-index.md` para navegação detalhada.

### By Topic

| Topic | Files | Description |
|-------|-------|-------------|
| Build & Install | 002, 024, 025, 026, 027 | Build from source, platform-specific builds, Docker |
| GPU Backends | 014-023, 039, 040 | CUDA, Metal, Vulkan, ROCm, OpenCL, SYCL, CANN, etc. |
| Multimodal | 003-013, 038 | Vision and audio model support (LLaVA, Qwen VL, Gemma, etc.) |
| Function Calling | 036 | OpenAI-compatible tool use with Jinja templates |
| Speculative Decoding | 043 | Draft models, n-gram strategies for faster generation |
| Configuration | 032, 037, 041 | Presets, llguidance, parsing |
| Development | 028-031 | Debugging, adding models, performance tips |

### By Keyword

| Keyword | File |
|---------|------|
| build, cmake | 027-build.md |
| installation | 002-install.md |
| cuda, nvidia | 015-backend-cuda-fedora.md |
| metal, macos | 027-build.md |
| vulkan | 027-build.md |
| rocm, amd, hip | 027-build.md |
| opencl, adreno | 016-backend-opencl.md |
| snapdragon, hexagon, npu | 001-backend-snapdragon-readme.md |
| android | 025-android.md |
| docker | 031-docker.md |
| multimodal, vision, llava | 038-multimodal.md |
| function-calling, tools | 036-function-calling.md |
| speculative, draft | 043-speculative.md |
| quantization, gguf | 027-build.md |
| performance, optimization | 030-development-token-generation-performance-tips.md |
| debugging, testing | 028-development-debugging-tests.md |
| model-architecture | 029-development-howto-add-model.md |
| openvino, intel | 017-backend-openvino.md |
| sycl, intel-gpu | 020-backend-sycl.md |
| cann, ascend | 039-backend-cann.md |

### Learning Path

1. **Start**: `docs/002-install.md` - Install llama.cpp
2. **Build**: `docs/027-build.md` - Build from source with GPU support
3. **Run**: Use `llama-cli` or `llama-server` with a GGUF model
4. **Multimodal**: `docs/038-multimodal.md` - Run vision/audio models
5. **Function Calling**: `docs/036-function-calling.md` - Enable tool use
6. **Optimize**: `docs/030-development-token-generation-performance-tips.md` - Performance tuning
7. **Speculative**: `docs/043-speculative.md` - Faster generation with draft models

## Common Tasks

### Build with CUDA support
→ `docs/027-build.md` (CUDA section)

```bash
cmake -B build -DGGML_CUDA=ON
cmake --build build --config Release
```

### Run multimodal (vision) model
→ `docs/038-multimodal.md`

```bash
llama-server -hf ggml-org/gemma-3-4b-it-GGUF
# or
llama-mtmd-cli -hf ggml-org/gemma-3-4b-it-GGUF
```

### Enable function calling
→ `docs/036-function-calling.md`

```bash
llama-server --jinja -fa -hf bartowski/Qwen2.5-7B-Instruct-GGUF:Q4_K_M
```

### Build for Android
→ `docs/025-android.md`, `docs/001-backend-snapdragon-readme.md`

```bash
docker run -it -u $(id -u):$(id -g) --volume $(pwd):/workspace --platform linux/amd64 ghcr.io/snapdragon-toolchain/arm64-android:v0.3
```

### Run with Docker
→ `docs/031-docker.md`

```bash
docker build -t llama-cpp-vulkan --target light -f .devops/vulkan.Dockerfile .
docker run -it --rm -v "$(pwd):/app:Z" --device /dev/dri/renderD128 llama-cpp-vulkan -m "/app/models/model.gguf"
```

### Enable speculative decoding
→ `docs/043-speculative.md`

```bash
# n-gram based (no draft model needed)
llama-server --spec-type ngram-simple --draft-max 64 -m model.gguf

# With draft model
llama-server -m target.gguf --draft draft.gguf
```

### Build with Vulkan (cross-platform GPU)
→ `docs/027-build.md` (Vulkan section)

```bash
# Linux: install libvulkan-dev glslc spirv-headers
cmake -B build -DGGML_VULKAN=ON
cmake --build build --config Release
```

### Performance tuning
→ `docs/030-development-token-generation-performance-tips.md`

Key options: `-ngl` (GPU offload layers), `-t` (threads), `--mlock`, `-c` (context size)

### Add support for a new model
→ `docs/029-development-howto-add-model.md`

1. Convert model to GGUF with `convert_hf_to_gguf.py`
2. Define architecture in `llama-arch.h/cpp`
3. Build GGML graph in `llama-model.cpp`

### Debug and test
→ `docs/028-development-debugging-tests.md`

```bash
cmake -B build -DCMAKE_BUILD_TYPE=Debug
cmake --build build
./build/bin/llama-cli -m model.gguf -p "test" --verbose
```
