---
title: Build s390x
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/build-s390x.md
source: git
fetched_at: 2026-04-28T09:49:21.242413789-03:00
rendered_js: false
word_count: 820
summary: This guide provides instructions for building llama.cpp on IBM Z and LinuxONE mainframes, including configuration for BLAS, hardware accelerators, and required model conversion for big-endian architecture.
tags:
    - s390x
    - llama-cpp
    - ibm-z
    - linuxone
    - model-conversion
    - cmake
    - zdnn
category: guide
optimized: true
optimized_at: '2026-04-28T12:00:00Z'
---
> [!IMPORTANT]
> This build documentation is specific to IBM Z & LinuxONE mainframes (s390x). For other architectures, see [[027-build|build.md]].

# Build llama.cpp locally (for s390x)

The main product is the `llama` library with a C-style interface in [include/llama.h](../include/llama.h). The project also includes example programs and tools ranging from minimal code snippets to an OpenAI-compatible HTTP server.

```bash
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
```

## CPU Build with BLAS

BLAS support is highly recommended for performance. Requires OpenBLAS installed.

```bash
cmake -S . -B build             \
    -DCMAKE_BUILD_TYPE=Release  \
    -DGGML_BLAS=ON              \
    -DGGML_BLAS_VENDOR=OpenBLAS

cmake --build build --config Release -j $(nproc)
```

Notes:

- For faster repeated compilation, install [ccache](https://ccache.dev/).
- VXE/VXE2 is enabled by default. To disable (not recommended):

    ```bash
    cmake -S . -B build             \
        -DCMAKE_BUILD_TYPE=Release  \
        -DGGML_BLAS=ON              \
        -DGGML_BLAS_VENDOR=OpenBLAS \
        -DGGML_VXE=OFF

    cmake --build build --config Release -j $(nproc)
    ```

- Debug builds:

    ```bash
    cmake -S . -B build             \
        -DCMAKE_BUILD_TYPE=Debug    \
        -DGGML_BLAS=ON              \
        -DGGML_BLAS_VENDOR=OpenBLAS
    cmake --build build --config Debug -j $(nproc)
    ```

- Static builds: add `-DBUILD_SHARED_LIBS=OFF`:

    ```bash
    cmake -S . -B build             \
        -DCMAKE_BUILD_TYPE=Release  \
        -DGGML_BLAS=ON              \
        -DGGML_BLAS_VENDOR=OpenBLAS \
        -DBUILD_SHARED_LIBS=OFF

    cmake --build build --config Release -j $(nproc)
    ```

## IBM zDNN Accelerator

Acceleration via the IBM zAIU co-processor in Telum I/II processors. Requires the [IBM zDNN library](https://github.com/IBM/zDNN). Build instructions: [Building and Installing zDNN](https://github.com/IBM/zDNN?tab=readme-ov-file#building-and-installing-zdnn).

```bash
cmake -S . -B build             \
    -DCMAKE_BUILD_TYPE=Release  \
    -DGGML_ZDNN=ON
cmake --build build --config Release -j$(nproc)
```

## Getting GGUF Models

All models must be converted to Big-Endian. Three options:

### 1. Pre-converted models (easiest)

Use models pre-converted and verified at [s390x Verified Models](https://huggingface.co/collections/taronaeo/s390x-verified-models-672765393af438d0ccb72a08) or [s390x Runnable Models](https://huggingface.co/collections/taronaeo/s390x-runnable-models-686e951824198df12416017e). These are converted from `safetensors` to GGUF Big-Endian with tokenizers verified on IBM z15+.

### 2. Convert safetensors to GGUF Big-Endian (recommended)

Model must be in `safetensors` format (e.g. [IBM Granite 3.3 2B](https://huggingface.co/ibm-granite/granite-3.3-2b-instruct)). Download the model repository first.

```bash
pip3 install -r requirements.txt

python3 convert_hf_to_gguf.py \
    --outfile model-name-be.f16.gguf \
    --outtype f16 \
    --bigendian \
    model-directory/
```

Example:

```bash
python3 convert_hf_to_gguf.py \
    --outfile granite-3.3-2b-instruct-be.f16.gguf \
    --outtype f16 \
    --bigendian \
    granite-3.3-2b-instruct/
```

### 3. Convert existing GGUF Little-Endian to Big-Endian

Model must be in `gguf` format (e.g. [IBM Granite 3.3 2B GGUF](https://huggingface.co/ibm-granite/granite-3.3-2b-instruct-GGUF)). Download the model file first.

```bash
python3 gguf-py/gguf/scripts/gguf_convert_endian.py model-name.f16.gguf BIG
```

Example:

```bash
python3 gguf-py/gguf/scripts/gguf_convert_endian.py granite-3.3-2b-instruct-le.f16.gguf BIG
mv granite-3.3-2b-instruct-le.f16.gguf granite-3.3-2b-instruct-be.f16.gguf
```

> [!warning]
> The GGUF endian conversion script may not support all data types. If it fails, convert via Step 2 (safetensors → GGUF Big-Endian).

## IBM Accelerators

| Accelerator | Minimum System | Compile Flag | Status |
|---|---|---|---|
| SIMD (VXE/VXE2) | IBM z15/LinuxONE 3+ | `-DGGML_VXE=ON` (default) | Available. Older systems (z14/arch12) use scalar fallback. |
| zDNN | IBM z17/LinuxONE 5+ | `-DGGML_ZDNN=ON` | WIP. Older systems (z15/arch13) default to CPU. |
| Spyre | IBM z17/LinuxONE 5+ | — | No support currently available. |

## Performance Tuning

1. **Virtualization**: Use LPAR (Type-1) virtualization. Type-2 is not supported and yields poor performance.
2. **IFL count**: Minimum 8 shared IFLs recommended. Increasing past 8 improves prompt processing only, not token generation. IFL count ≠ vCPU count.
3. **SMT**: Disable SMT via kernel boot parameters — it negatively affects performance.
4. **BLAS**: Strongly recommended; IBM VXE/VXE2 SIMD acceleration depends on the BLAS implementation.

## FAQ

1. **Error: `gguf_init_from_file_impl: failed to load model: this GGUF file version 50331648 is extremely large, is there a mismatch between the host and model endianness?`**
    Ensure the model is GGUFv3 Big-Endian (denoted with `-be` suffix). See [Getting GGUF Models](#getting-gguf-models) to convert manually.

2. **Extremely poor performance**
    Check [Appendix B: SIMD Support Matrix](#appendix-b-simd-support-matrix) for your quantization's SIMD support.

3. **Error: `invalid switch -march=z17` on IBM z17**
    Requires minimum GCC 15.1.0 and latest `binutils`.

4. **`sentencepiece` install fails with GCC 15+**
    Known issue: [sentencepiece#1108](https://github.com/google/sentencepiece/issues/1108). Workaround:

    ```bash
    CXXFLAGS="-include cstdint" pip3 install -r requirements.txt
    ```

## Getting Help

- **Bugs/Feature Requests**: File an issue in llama.cpp with "s390x" in the title.
- **Other Questions**: Contact [aionz@us.ibm.com](mailto:aionz@us.ibm.com).

## Appendix A: Hardware Support Matrix

| System | Support | Minimum Compiler |
|---|---|---|
| IBM z15 | ✅ | — |
| IBM z16 | ✅ | — |
| IBM z17 | ✅ | GCC 15.1.0 |
| IBM zDNN | ✅ | — |

Legend: ✅ = supported and verified, 🚫 = unsupported.

## Appendix B: SIMD Support Matrix

| Type | VX/VXE/VXE2 | zDNN | Spyre |
|---|---|---|---|
| FP32 | ✅ | ✅ | ❓ |
| FP16 | ✅ | ✅ | ❓ |
| BF16 | ✅ | ✅ | ❓ |
| Q4_0 | ✅ | ❓ | ❓ |
| Q4_1 | ✅ | ❓ | ❓ |
| MXFP4 | ✅ | ❓ | ❓ |
| Q5_0 | ✅ | ❓ | ❓ |
| Q5_1 | ✅ | ❓ | ❓ |
| Q8_0 | ✅ | ❓ | ❓ |
| Q2_K | 🚫 | ❓ | ❓ |
| Q3_K | ✅ | ❓ | ❓ |
| Q4_K | ✅ | ❓ | ❓ |
| Q5_K | ✅ | ❓ | ❓ |
| Q6_K | ✅ | ❓ | ❓ |
| TQ1_0 | 🚫 | ❓ | ❓ |
| TQ2_0 | 🚫 | ❓ | ❓ |
| IQ2_XXS | 🚫 | ❓ | ❓ |
| IQ2_XS | 🚫 | ❓ | ❓ |
| IQ2_S | 🚫 | ❓ | ❓ |
| IQ3_XXS | 🚫 | ❓ | ❓ |
| IQ3_S | 🚫 | ❓ | ❓ |
| IQ1_S | 🚫 | ❓ | ❓ |
| IQ1_M | 🚫 | ❓ | ❓ |
| IQ4_NL | ✅ | ❓ | ❓ |
| IQ4_XS | ✅ | ❓ | ❓ |
| FP32→FP16 | 🚫 | ❓ | ❓ |
| FP16→FP32 | 🚫 | ❓ | ❓ |

Legend: ✅ = acceleration available, 🚫 = unavailable (scalar fallback), ❓ = unknown.

Last Updated by **Aaron Teo (aaron.teo1@ibm.com)** on Feb 15, 2026.
