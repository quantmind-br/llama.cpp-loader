---
title: ZenDNN
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/ZenDNN.md
source: git
fetched_at: 2026-04-28T09:49:12.918471226-03:00
rendered_js: false
word_count: 704
summary: This document provides instructions on using AMD's ZenDNN library to accelerate inference performance in llama.cpp on AMD EPYC and Ryzen processors through optimized matrix multiplication operations.
tags:
    - zendnn
    - amd-epyc
    - llama-cpp
    - inference-acceleration
    - matrix-multiplication
    - cpu-optimization
category: guide
optimized: true
optimized_at: '2026-04-28T00:00:00Z'
---
# llama.cpp for AMD ZenDNN

> [!warning]
> **ZenDNN** (this page) is **not** the same as **zDNN** ([IBM zDNN](022-backend-zdnn.md)):
> - **ZenDNN** — AMD's deep learning library for AMD EPYC CPUs
> - **zDNN** — IBM's Deep Neural Network acceleration library for IBM Z & LinuxONE Mainframes

## Background

**ZenDNN** (Zen Deep Neural Network Library) is AMD's high-performance deep learning inference library for AMD EPYC™ CPUs. The llama.cpp ZenDNN backend leverages its **LowOHA (Low Overhead Hardware Accelerated)** MatMul operator for efficient GEMM operations with minimal overhead, built-in weight caching, and direct access to backend libraries (AOCL DLP, LibXSMM, OneDNN).

More information: https://www.amd.com/en/developer/zendnn.html

## OS

| OS      | Status  | Verified                                       |
|:-------:|:-------:|:----------------------------------------------:|
| Linux   | Support | Ubuntu 20.04, 22.04, 24.04                     |

Latest supported OS list: https://github.com/amd/ZenDNN/blob/a18adf8c605fb5f5e52cefd7eda08a7b18febbaf/README.md#15-supported-os

## Hardware

### AMD CPUs

ZenDNN is optimized for AMD EPYC™ and AMD Ryzen™ processors based on "Zen" microarchitecture and newer.

| CPU Family                    | Status  | Notes                              |
|:-----------------------------:|:-------:|:----------------------------------:|
| AMD EPYC™ 9005 Series (Turin) | Support | 5th Gen - Zen 5 architecture       |
| AMD EPYC™ 9004 Series (Genoa) | Support | 4th Gen - Zen 4 architecture       |
| AMD EPYC™ 7003 Series (Milan) | Support | 3rd Gen - Zen 3 architecture       |
| AMD Ryzen™ AI MAX (Strix Halo)| Support | High-performance mobile processors |

- Best performance on EPYC™ processors with high core counts (e.g., EPYC 9005 series).
- Leverages AVX2 and AVX-512 instruction sets.
- Ensure sufficient memory bandwidth for optimal performance.

## Supported Operations

The ZenDNN backend accelerates **MUL_MAT** and **MUL_MAT_ID** operations. All other operations use the standard CPU backend.

| Operation    | Status  | Notes                                          |
|:-------------|:-------:|:----------------------------------------------:|
| MUL_MAT      | Support | Accelerated via ZenDNN LowOHA MatMul           |
| MUL_MAT_ID   | Support | Accelerated via ZenDNN LowOHA MatMul (MoE)     |

Models benefit most when matrix multiplications dominate the workload (typical for transformer-based LLMs and MoE models).

## DataType Supports

| DataType               | Status  | Notes                                         |
|:----------------------:|:-------:|:---------------------------------------------:|
| FP32                   | Support | Full precision floating point                 |
| BF16                   | Support | BFloat16 (best performance on Zen 4/Zen 5)    |

> [!info]
> **BF16** provides best performance on Zen 4 and Zen 5 EPYC™ processors (Genoa, Turin). On older CPUs, operations use FP32.

## Linux

### Setup Environment

#### Option 1: Automatic Download and Build (Recommended)

```sh
# Build llama.cpp - ZenDNN will be automatically downloaded and built
cmake -B build -DGGML_ZENDNN=ON -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release -j $(nproc)
```

No manual ZenDNN installation required — CMake handles everything.

#### Option 2: Use Custom ZenDNN Installation

**Step 1:** Build ZenDNN from source:

```sh
git clone https://github.com/amd/ZenDNN.git
cd ZenDNN

# Build and install (requires CMake >= 3.25)
mkdir build && cd build
cmake ..
cmake --build . --target all
```

Default installation path: `ZenDNN/build/install`. Full instructions: https://github.com/amd/ZenDNN/blob/a18adf8c605fb5f5e52cefd7eda08a7b18febbaf/README.md

**Step 2:** Build llama.cpp with custom ZenDNN path:

```sh
# Using environment variable
export ZENDNN_ROOT=/path/to/ZenDNN/build/install
cmake -B build -DGGML_ZENDNN=ON -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release -j $(nproc)

# OR specify path directly in CMake
cmake -B build -DGGML_ZENDNN=ON -DZENDNN_ROOT=/path/to/ZenDNN/build/install -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release -j $(nproc)
```

### Run the Server

#### 1. Download Model

```sh
huggingface-cli download meta-llama/Llama-3.1-8B-Instruct-GGUF --local-dir models/
```

#### 2. Start Server

```sh
# Set optimal configuration
export ZENDNNL_MATMUL_ALGO=1    # Blocked AOCL DLP algo for best performance

# Start server
./build/bin/llama-server \
    -m models/Llama-3.1-8B-Instruct.BF16.gguf \
    --host 0.0.0.0 \
    --port 8080 \
    -t 64
```

Access at `http://localhost:8080`.

- Use `ZENDNNL_MATMUL_ALGO=1` for optimal performance.
- For NUMA systems: `numactl --cpunodebind=0 --membind=0 ./build/bin/llama-server ...`

## Environment Variables

For ZenDNN environment variables: https://github.com/amd/ZenDNN/blob/a18adf8c605fb5f5e52cefd7eda08a7b18febbaf/docs/runtime_env.md

### Performance Optimization

ZenDNN's LowOHA MatMul supports multiple backend algorithms. **Best performance** uses the Blocked AOCL DLP algorithm:

```sh
export ZENDNNL_MATMUL_ALGO=1    # Blocked AOCL DLP algo (recommended)
```

Algorithm details: https://github.com/amd/ZenDNN/blob/a18adf8c605fb5f5e52cefd7eda08a7b18febbaf/docs/runtime_env.md#algorithm-details

### Profiling and Debugging

Logging options: https://github.com/amd/ZenDNN/blob/a18adf8c605fb5f5e52cefd7eda08a7b18febbaf/docs/logging.md

## Known Issues

- **Limited operation support**: Only MUL_MAT and MUL_MAT_ID are accelerated. Other operations fall back to the standard CPU backend.
- **BF16 support**: Requires AMD Zen 4 or Zen 5 (EPYC 9004/9005). Older CPUs use FP32.
- **NUMA awareness**: Multi-socket systems may require manual NUMA binding for optimal performance.

## Q&A

**Q: How do I verify ZenDNN backend is being used?**
Check the log output — you should see messages indicating the ZenDNN backend is initialized.

**Q: What performance improvement can I expect?**
On AMD EPYC processors, typically 1.1x–2x speedup for matrix multiplication operations. Varies by model size, batch size, and CPU architecture.

**Q: Can I use ZenDNN on non-AMD processors?**
ZenDNN is optimized for AMD processors. Performance benefits are only guaranteed on AMD Zen-based architectures.

**Q: Does ZenDNN support quantized models?**
Currently only FP32 and BF16 are supported.

**Q: Why is my inference not faster with ZenDNN?**
Ensure:
1. Using an AMD EPYC or Ryzen processor (Zen 2 or newer)
2. `ZENDNNL_MATMUL_ALGO=1` is set
3. Model is sufficiently large (small models may not benefit)
4. Enable profiling to verify ZenDNN MatMul is being called

> [!tip]
> Add the **[ZenDNN]** prefix in issues/PRs titles for faster triage by the ZenDNN team.

## TODO

- Expand operation support beyond MUL_MAT and MUL_MAT_ID (attention operations, activations, etc.)
