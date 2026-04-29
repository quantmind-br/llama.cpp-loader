---
title: OPENCL
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/OPENCL.md
source: git
fetched_at: 2026-04-28T09:49:10.075226565-03:00
rendered_js: false
word_count: 546
summary: This document provides a guide for setting up and building the OpenCL backend for llama.cpp, focusing on hardware support, model quantization requirements, and compilation instructions for Android and Windows Arm64 platforms.
tags:
    - llama-cpp
    - opencl
    - gpu-acceleration
    - adreno
    - model-quantization
    - cross-compilation
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# llama.cpp for OpenCL

## Background

OpenCL (Open Computing Language) is an open, royalty-free standard for parallel programming of diverse accelerators (GPUs, CPUs, FPGAs) across supercomputers, cloud servers, PCs, mobile, and embedded platforms. It specifies a C99-based programming language and APIs. Like CUDA, it is widely used for GPU programming and supported by most GPU vendors.

### Llama.cpp + OpenCL

The OpenCL backend primarily targets **Qualcomm Adreno GPUs**. It can also run on certain Intel GPUs without [[020-backend-sycl|SYCL]] support, though performance is not optimal.

## OS Support

| OS | Status | Verified |
|---------|---------|------------------------------------------------|
| Android | Support | Snapdragon 8 Gen 3, Snapdragon 8 Elite |
| Windows | Support | Windows 11 Arm64 with Snapdragon X Elite |
| Linux | Support | Ubuntu 22.04 WSL2 with Intel 12700H |

## Hardware

### Adreno GPU

| Adreno GPU | Status |
|:------------------------------------:|:-------:|
| Adreno 750 (Snapdragon 8 Gen 3) | Support |
| Adreno 830 (Snapdragon 8 Elite) | Support |
| Adreno X85 (Snapdragon X Elite) | Support |

> A6x GPUs with a recent driver and compiler are supported (usually IoT platforms). A6x GPUs in phones are likely unsupported due to outdated drivers/compilers.

## DataType Support

| DataType | Status |
|:----------------------:|:--------------------------:|
| Q4_0 | Support |
| Q6_K | Support, but not optimized |
| Q8_0 | Support |
| MXFP4 | Support |

## Model Preparation

See the [llama-quantize tool](/tools/quantize/README.md) for converting HuggingFace safetensor models to GGUF.

`Q4_0` is the optimized quantization. For best Adreno GPU performance, add `--pure` (all weights in `Q4_0`):

```sh
./llama-quantize --pure ggml-model-qwen2.5-3b-f16.gguf ggml-model-qwen-3b-Q4_0.gguf Q4_0
```

`Q4_0` without `--pure` also works (uses mixed `Q4_0`/`Q6_K`) but with worse performance.

### MXFP4 MoE Models

OpenAI gpt-oss models are MoE in `MXFP4`. The quantized format is `MXFP4_MOE` (mixture of `MXFP4` and `Q8_0`). No `--pure` needed. For gpt-oss-20b, download the pre-quantized GGUF directly from [Hugging Face](https://huggingface.co/ggml-org/gpt-oss-20b-GGUF).

> [!note] Pure `Q4_0` quantization of gpt-oss-20b is possible but not recommended — `MXFP4` is optimized for MoE while `Q4_0` is not, and accuracy degrades. The `Q4_0` variant found [here](https://huggingface.co/unsloth/gpt-oss-20b-GGUF/blob/main/gpt-oss-20b-Q4_0.gguf) is actually a mixture of `Q4_0`, `Q8_0`, and `MXFP4` and outperforms `MXFP4_MOE`.

## CMake Options

| CMake option | Default | Description |
|:---------------------------------:|:-------:|:------------------------------------------|
| `GGML_OPENCL_EMBED_KERNELS` | `ON` | Embed OpenCL kernels into the executable. |
| `GGML_OPENCL_USE_ADRENO_KERNELS` | `ON` | Use Adreno-optimized kernels. |

## Android Build

Requires: Git, CMake 3.29, Ninja, Python3 on Ubuntu 22.04.

### I. Setup Environment

1. **Install NDK**

```sh
cd ~
wget https://dl.google.com/android/repository/commandlinetools-linux-8512546_latest.zip && \
unzip commandlinetools-linux-8512546_latest.zip && \
mkdir -p ~/android-sdk/cmdline-tools && \
mv cmdline-tools latest && \
mv latest ~/android-sdk/cmdline-tools/ && \
rm -rf commandlinetools-linux-8512546_latest.zip

yes | ~/android-sdk/cmdline-tools/latest/bin/sdkmanager "ndk;26.3.11579264"
```

2. **Install OpenCL Headers and Library**

```sh
mkdir -p ~/dev/llm
cd ~/dev/llm

git clone https://github.com/KhronosGroup/OpenCL-Headers && \
cd OpenCL-Headers && \
cp -r CL ~/android-sdk/ndk/26.3.11579264/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include

cd ~/dev/llm

git clone https://github.com/KhronosGroup/OpenCL-ICD-Loader && \
cd OpenCL-ICD-Loader && \
mkdir build_ndk26 && cd build_ndk26 && \
cmake .. -G Ninja -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_TOOLCHAIN_FILE=$HOME/android-sdk/ndk/26.3.11579264/build/cmake/android.toolchain.cmake \
  -DOPENCL_ICD_LOADER_HEADERS_DIR=$HOME/android-sdk/ndk/26.3.11579264/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=24 \
  -DANDROID_STL=c++_shared && \
ninja && \
cp libOpenCL.so ~/android-sdk/ndk/26.3.11579264/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/lib/aarch64-linux-android
```

### II. Build llama.cpp

```sh
cd ~/dev/llm

git clone https://github.com/ggml-org/llama.cpp && \
cd llama.cpp && \
mkdir build-android && cd build-android

cmake .. -G Ninja \
  -DCMAKE_TOOLCHAIN_FILE=$HOME/android-sdk/ndk/26.3.11579264/build/cmake/android.toolchain.cmake \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-28 \
  -DBUILD_SHARED_LIBS=OFF \
  -DGGML_OPENCL=ON

ninja
```

## Windows 11 Arm64 Build

Requires: Git, CMake 3.29, Clang 19, Ninja, Visual Studio 2022, Powershell 7, Python. Visual Studio provides headers/libraries; Clang must be used (not `cl`). Alternatively install Visual Studio Build Tools. Powershell 7 is required for the commands below.

### I. Setup Environment

1. **Install OpenCL Headers and Library**

```powershell
mkdir -p ~/dev/llm

cd ~/dev/llm
git clone https://github.com/KhronosGroup/OpenCL-Headers && cd OpenCL-Headers
mkdir build && cd build
cmake .. -G Ninja `
  -DBUILD_TESTING=OFF `
  -DOPENCL_HEADERS_BUILD_TESTING=OFF `
  -DOPENCL_HEADERS_BUILD_CXX_TESTS=OFF `
  -DCMAKE_INSTALL_PREFIX="$HOME/dev/llm/opencl"
cmake --build . --target install

cd ~/dev/llm
git clone https://github.com/KhronosGroup/OpenCL-ICD-Loader && cd OpenCL-ICD-Loader
mkdir build && cd build
cmake .. -G Ninja `
  -DCMAKE_BUILD_TYPE=Release `
  -DCMAKE_PREFIX_PATH="$HOME/dev/llm/opencl" `
  -DCMAKE_INSTALL_PREFIX="$HOME/dev/llm/opencl"
cmake --build . --target install
```

### II. Build llama.cpp

```powershell

mkdir -p ~/dev/llm
cd ~/dev/llm

git clone https://github.com/ggml-org/llama.cpp && cd llama.cpp
mkdir build && cd build

cmake .. -G Ninja `
  -DCMAKE_TOOLCHAIN_FILE="$HOME/dev/llm/llama.cpp/cmake/arm64-windows-llvm.cmake" `
  -DCMAKE_BUILD_TYPE=Release `
  -DCMAKE_PREFIX_PATH="$HOME/dev/llm/opencl" `
  -DBUILD_SHARED_LIBS=OFF `
  -DGGML_OPENCL=ON
ninja
```

## Linux Build

Same two steps as Windows Arm64, but without `-DCMAKE_TOOLCHAIN_FILE` and with backslashes instead of backticks. Requires Git, CMake, Clang, Ninja, Python.

### I. Setup Environment

1. **Install OpenCL Headers and Library**

```bash
mkdir -p ~/dev/llm

cd ~/dev/llm
git clone https://github.com/KhronosGroup/OpenCL-Headers && cd OpenCL-Headers
mkdir build && cd build
cmake .. -G Ninja \
  -DBUILD_TESTING=OFF \
  -DOPENCL_HEADERS_BUILD_TESTING=OFF \
  -DOPENCL_HEADERS_BUILD_CXX_TESTS=OFF \
  -DCMAKE_INSTALL_PREFIX="$HOME/dev/llm/opencl"
cmake --build . --target install

cd ~/dev/llm
git clone https://github.com/KhronosGroup/OpenCL-ICD-Loader && cd OpenCL-ICD-Loader
mkdir build && cd build
cmake .. -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_PREFIX_PATH="$HOME/dev/llm/opencl" \
  -DCMAKE_INSTALL_PREFIX="$HOME/dev/llm/opencl"
cmake --build . --target install
```

### II. Build llama.cpp

```bash
mkdir -p ~/dev/llm
cd ~/dev/llm

git clone https://github.com/ggml-org/llama.cpp && cd llama.cpp
mkdir build && cd build

cmake .. -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_PREFIX_PATH="$HOME/dev/llm/opencl" \
  -DBUILD_SHARED_LIBS=OFF \
  -DGGML_OPENCL=ON
ninja
```

## Known Issues

- Flash attention does not always improve performance.
- A6xx GPUs with old drivers/compilers (phones) are not supported.

## TODO

- Optimization for Q6_K
- Support and optimization for Q4_K
- Improve flash attention

#opencl #gpu-acceleration #adreno #cross-compilation #model-quantization
