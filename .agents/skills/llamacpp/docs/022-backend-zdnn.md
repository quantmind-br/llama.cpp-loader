---
title: ZDNN
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/zDNN.md
source: git
fetched_at: 2026-04-28T09:49:19.909763044-03:00
rendered_js: false
word_count: 236
summary: This document provides instructions for compiling and configuring the llama.cpp project to utilize the IBM zDNN hardware acceleration library on supported IBM Z mainframe systems.
tags:
    - ibm-zdnn
    - llama-cpp
    - hardware-acceleration
    - mainframe-optimization
    - neural-network-inference
    - build-instructions
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# llama.cpp for IBM zDNN Accelerator

> [!warning]
> **zDNN ≠ ZenDNN.** **zDNN** (this page): IBM's Deep Neural Network acceleration library for IBM Z & LinuxONE Mainframes. **ZenDNN**: AMD's deep learning library for AMD EPYC CPUs ([[023-backend-zendnn|ZenDNN documentation]]).

## Background

IBM zDNN (Z Deep Neural Network) is a hardware acceleration library for the IBM NNPA (Neural Network Processor Assist) accelerator in IBM Telum I and II processors, providing significant performance improvements for neural network inference.

The llama.cpp zDNN backend enables llama.cpp on IBM z17 and later systems via the zDNN library.

## Software & Hardware Support

| Hardware Level       | Status        | Verified                   |
| -------------------- | ------------- | -------------------------- |
| IBM z17 / LinuxONE 5 | Supported     | RHEL 9.6, IBM z17, 40 IFLs |
| IBM z16 / LinuxONE 4 | Not Supported |                            |

## Data Types Supported

| Data Type | Status    |
| --------- | --------- |
| F32       | Supported |
| F16       | Supported |
| BF16      | Supported |

## CMake Options

| CMake Option | Default Value | Description                         |
| ------------ | ------------- | ----------------------------------- |
| `GGML_ZDNN`  | `OFF`         | Compile llama.cpp with zDNN support |
| `ZDNN_ROOT`  | `""`          | Override zDNN library lookup        |

## 1. Install zDNN Library

> [!warning]
> The `apt`/`yum` zDNN package may not work correctly ([#15772](https://github.com/ggml-org/llama.cpp/issues/15772)). Compile from source.

```sh
git clone --recurse-submodules https://github.com/IBM/zDNN
cd zDNN

autoreconf .
./configure --prefix=/opt/zdnn-libs

make build
sudo make install
```

## 2. Build llama.cpp

```sh
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp

cmake -S . -G Ninja -B build \
    -DCMAKE_BUILD_TYPE=Release \
    -DGGML_ZDNN=ON \
    -DZDNN_ROOT=/opt/zdnn-libs
cmake --build build --config Release -j$(nproc)
```

#ibm-zdnn #hardware-acceleration #mainframe #build-instructions