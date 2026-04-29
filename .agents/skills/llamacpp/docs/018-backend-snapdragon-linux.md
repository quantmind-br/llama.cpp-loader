---
title: Linux
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/snapdragon/linux.md
source: git
fetched_at: 2026-04-28T09:49:17.155324368-03:00
rendered_js: false
word_count: 134
summary: This document provides instructions for cross-compiling, building, and deploying the llama.cpp project for Qualcomm Snapdragon-based Linux devices using a specialized Docker toolchain.
tags:
    - snapdragon
    - llama-cpp
    - cross-compilation
    - docker
    - hexagon-sdk
    - arm64-linux
    - opencl
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# Snapdragon-based Linux devices

## Docker Setup

The easiest way to build llama.cpp for Snapdragon-based Linux is using the toolchain Docker image ([github.com/snapdragon-toolchain](https://github.com/snapdragon-toolchain)), which includes OpenCL SDK, Hexagon SDK, CMake, and the ARM64 Linux cross-compilation toolchain.

Cross-compilation runs on **Linux X86** hosts. Binaries deploy to and run on **Qualcomm Snapdragon ARM64 Linux** devices.

```bash
~/src/llama.cpp$ docker run -it -u $(id -u):$(id -g) --volume $(pwd):/workspace --platform linux/amd64 ghcr.io/snapdragon-toolchain/arm64-linux:v0.1
[d]/> cd /workspace
```

> [!note]
> The rest of the Linux build process assumes you're running inside the toolchain container.

## How to Build

Build llama.cpp with CPU, OpenCL, and Hexagon backends via CMake presets:

```bash
cp docs/backend/snapdragon/CMakeUserPresets.json .
cmake --preset arm64-linux-snapdragon-release -B build-snapdragon
cmake --build build-snapdragon -j $(nproc)
```

To create an installable package, use `cmake --install` and zip:

```bash
cmake --install build-snapdragon --prefix pkg-snapdragon
zip -r pkg-snapdragon.zip pkg-snapdragon
```

## How to Install

Transfer `pkg-snapdragon.zip` to the target device, then unzip and set environment variables:

```bash
unzip pkg-snapdragon.zip
cd pkg-snapdragon
export LD_LIBRARY_PATH=./lib
export ADSP_LIBRARY_PATH=./lib
```

Download a model onto the device:

```bash
wget https://huggingface.co/bartowski/Llama-3.2-3B-Instruct-GGUF/resolve/main/Llama-3.2-3B-Instruct-Q4_0.gguf
```

## How to Run

With environment variables set, run llama-cli with Hexagon backends:

```bash
./bin/llama-cli -m Llama-3.2-3B-Instruct-Q4_0.gguf --device HTP0 -ngl 99 -p "what is the most popular cookie in the world?"
```

#snapdragon #cross-compilation #docker #hexagon