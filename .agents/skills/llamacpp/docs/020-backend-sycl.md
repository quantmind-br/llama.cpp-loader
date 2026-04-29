---
title: SYCL
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/SYCL.md
source: git
fetched_at: 2026-04-28T09:49:11.153445232-03:00
rendered_js: false
word_count: 1911
summary: This document provides an overview and technical guidelines for using the SYCL backend in llama.cpp, focusing on hardware support and setup requirements for Intel GPUs.
tags:
    - sycl
    - intel-gpu
    - llama-cpp
    - oneapi
    - heterogeneous-computing
    - hardware-acceleration
category: guide
optimized: true
optimized_at: 2026-04-28T00:00:00Z
---
# llama.cpp for SYCL

## Background

**SYCL** is a high-level parallel programming model for heterogeneous computing based on standard C++17, targeting CPUs, GPUs, and FPGAs.

**oneAPI** is an open ecosystem supporting Intel CPUs, GPUs, and FPGAs. Key components:

- **DPCPP** *(Data Parallel C++)*: primary oneAPI SYCL implementation (icpx/icx compilers).
- **oneAPI Libraries**: optimized libraries *(e.g. oneMKL, oneMath, oneDNN)*.
- **oneAPI LevelZero**: low-level interface for Intel iGPUs and dGPUs.

### Llama.cpp + SYCL

The SYCL backend is primarily designed for **Intel GPUs**. Cross-platform capabilities enable support for other vendor GPUs.

## Recommended Release

### Windows

| Commit ID | Tag | Release | Verified Platform | Update date |
|-|-|-|-|-|
| 24e86cae7219b0f3ede1d5abdf5bf3ad515cccb8 | b5377 | [llama-b5377-bin-win-sycl-x64.zip](https://github.com/ggml-org/llama.cpp/releases/download/b5377/llama-b5377-bin-win-sycl-x64.zip) | Arc B580/Linux/oneAPI 2025.1; LNL Arc GPU/Windows 11/oneAPI 2025.1.1 | 2025-05-15 |
| 3bcd40b3c593d14261fb2abfabad3c0fb5b9e318 | b4040 | [llama-b4040-bin-win-sycl-x64.zip](https://github.com/ggml-org/llama.cpp/releases/download/b4040/llama-b4040-bin-win-sycl-x64.zip) | Arc A770/Linux/oneAPI 2024.1; MTL Arc GPU/Windows 11/oneAPI 2024.1 | 2024-11-19 |
| fb76ec31a9914b7761c1727303ab30380fd4f05c | b3038 | [llama-b3038-bin-win-sycl-x64.zip](https://github.com/ggml-org/llama.cpp/releases/download/b3038/llama-b3038-bin-win-sycl-x64.zip) | Arc A770/Linux/oneAPI 2024.1; MTL Arc GPU/Windows 11/oneAPI 2024.1 | |

### Ubuntu 24.04

Ubuntu 24.04 x64 release packages (FP32/FP16) include only SYCL backend binaries. They require pre-installed Intel GPU drivers and oneAPI packages matching the build package version (see release.yml: ubuntu-24-sycl). Recommended with Intel Docker.

FP32 and FP16 differ in accuracy and performance. Choose based on test results.

## News

- **2026.04** — Optimize mul_mat reorder for Q4_K, Q5_K, Q_K, Q8_0. Fused MoE. Upgrade CI/package for oneAPI 2025.3.3, Ubuntu 24.04 build package.
- **2026.03** — Flash-Attention support (less memory usage; performance impact depends on LLM).
- **2026.02** — Removed Nvidia & AMD GPU support (oneAPI plugin channels unavailable).
- **2025.11** — Support malloc device memory >4GB.
- **2025.2** — Optimize MUL_MAT Q4_0 for all dGPUs and built-in GPUs since MTL. 21%-87% improvement on llama-2-7b.Q4_0:

  | GPU | Base tokens/s | Increased tokens/s | Percent |
  |-|-|-|-|
  | PVC 1550 | 39 | 73 | +87% |
  | Flex 170 | 39 | 50 | +28% |
  | Arc A770 | 42 | 55 | +30% |
  | MTL | 13 | 16 | +23% |
  | ARL-H | 14 | 17 | +21% |

- **2024.11** — Use syclcompat for performance improvement (requires oneAPI 2025.0+).
- **2024.8** — oneDNN as default GEMM library; improved compatibility for new Intel GPUs.
- **2024.5** — Performance: 34→37 tokens/s llama-2-7b.Q4_0 on Arc A770. Arch Linux verified.
- **2024.4** — New data types: GGML_TYPE_IQ4_NL, IQ4_XS, IQ3_XXS, IQ3_S, IQ2_XXS, IQ2_XS, IQ2_S, IQ1_S, IQ1_M.
- **2024.3** — Windows binary release. Blog: [intel.com](https://www.intel.com/content/www/us/en/developer/articles/technical/run-llm-on-all-gpus-using-llama-cpp-artical.html) | [medium.com](https://medium.com/@jianyu_neo/run-llm-on-all-intel-gpus-using-llama-cpp-fd2e2dcbd9bd). Baseline: [tag b2437](https://github.com/ggml-org/llama.cpp/tree/b2437). Multi-card support: `--split-mode [none|layer]` (row not supported). `--main-gpu` replaces `$GGML_SYCL_DEVICE`. Detect all GPUs with level-zero and same top Max compute units. New OPs: hardsigmoid, hardswish, pool2d.
- **2024.1** — SYCL backend created. Windows build support.

## OS

| OS | Status | Verified |
|-|-|-|
| Linux | Support | Ubuntu 22.04, Fedora Silverblue 39, Arch Linux |
| Windows | Support | Windows 11 |

## Hardware

### Intel GPU

Supported Intel GPU families:

- Intel Data Center Max Series
- Intel Flex Series, Arc Series
- Intel Built-in Arc GPU
- Intel iGPU in Core CPU (11th Gen+, see [oneAPI supported GPU](https://www.intel.com/content/www/us/en/developer/articles/system-requirements/intel-oneapi-base-toolkit-system-requirements.html#inpage-nav-1-1))

Older Intel GPUs may work with [[016-backend-opencl|OpenCL]], though performance is suboptimal and some GPUs lack GPGPU support.

#### Verified devices

| Intel GPU | Status | Verified Model |
|-|-|-|
| Intel Data Center Max Series | Support | Max 1550, 1100 |
| Intel Data Center Flex Series | Support | Flex 170 |
| Intel Arc A-Series | Support | Arc A770, Arc A730M, Arc A750 |
| Intel Arc B-Series | Support | Arc B580 |
| Intel built-in Arc GPU | Support | Built-in Arc GPU in Meteor Lake, Arrow Lake, Lunar Lake |
| Intel iGPU | Support | iGPU in 13700k, 13400, i5-1250P, i7-1260P, i7-1165G7 |

> [!note]
> **Memory**: Device memory limits large model loading. Loaded model size (`llm_load_tensors: buffer_size`) appears in `./bin/llama-completion` logs. GPU shared memory must accommodate the model: e.g. llama-2-7b.Q4_0 requires ≥8.0GB (iGPU) or ≥4.0GB (dGPU).
>
> **Execution Unit (EU)**: iGPUs with <80 EUs will likely be too slow for practical inference.

### Other Vendor GPU

NA

## Docker

Docker build is limited to *Intel GPU* targets.

### Build image

```sh
# Using FP32
docker build -t llama-cpp-sycl --build-arg="GGML_SYCL_F16=OFF" --target light -f .devops/intel.Dockerfile .

# Using FP16
docker build -t llama-cpp-sycl --build-arg="GGML_SYCL_F16=ON" --target light -f .devops/intel.Dockerfile .
```

> [!tip]
> Use `.devops/llama-server-intel.Dockerfile` for the "server" alternative. See [[031-docker|documentation for Docker]] for available images.

### Run container

```sh
# First, find all the DRI cards
ls -la /dev/dri
# Then, pick the card that you want to use (here for e.g. /dev/dri/card1).
docker run -it --rm -v "/path/to/models:/models" --device /dev/dri/renderD128:/dev/dri/renderD128 --device /dev/dri/card0:/dev/dri/card0 llama-cpp-sycl -m /models/7B/ggml-model-q4_0.gguf -p "Building a website can be done in 10 simple steps:" -n 400 -e -ngl 33 -c 4096 -s 0
```

> [!warning]
> - Docker tested on native Linux only. WSL not verified.
> - You may need Intel GPU driver on the **host** machine (see [[#Linux]] below).

## Linux

### I. Setup Environment

#### 1. Install GPU drivers

- **Intel data center GPUs**: [Get Intel dGPU Drivers](https://dgpu-docs.intel.com/driver/installation.html#ubuntu-install-steps)
- **Client GPUs (iGPU & Arc A-Series)**: [client iGPU driver](https://dgpu-docs.intel.com/driver/client/overview.html)

After install, add user to `video` and `render` groups, then logout/re-login:

```sh
sudo usermod -aG render $USER
sudo usermod -aG video $USER
```

Verify with `clinfo`:

```sh
sudo apt install clinfo
sudo clinfo -l
```

Sample output:

```sh
Platform #0: Intel(R) OpenCL Graphics
 `-- Device #0: Intel(R) Arc(TM) A770 Graphics

Platform #0: Intel(R) OpenCL HD Graphics
 `-- Device #0: Intel(R) Iris(R) Xe Graphics [0x9a49]
```

#### 2. Install Intel® oneAPI Base toolkit

SYCL depends on: oneAPI DPC++/C++ compiler/runtime, oneDPL, oneDNN, oneMKL.

All included in **Intel® oneAPI Base toolkit** and **Intel® Deep Learning Essentials** (recommended — smaller install). Download from [Intel® oneAPI Base Toolkit](https://www.intel.com/content/www/us/en/developer/tools/oneapi/base-toolkit.html). Keep default install path (`/opt/intel/oneapi`).

Verified releases:

| Release |
|-|
| 2025.3.3 |
| 2025.2.1 |
| 2025.1 |
| 2024.1 |

#### 3. Verify installation

```sh
source /opt/intel/oneapi/setvars.sh
sycl-ls
```

Expected: at least one `[level_zero:gpu]` device. Example output:

```
[level_zero:gpu][level_zero:0] Intel(R) oneAPI Unified Runtime over Level-Zero, Intel(R) Arc(TM) A770 Graphics 12.55.8 [1.3.29735+27]
[level_zero:gpu][level_zero:1] Intel(R) oneAPI Unified Runtime over Level-Zero, Intel(R) UHD Graphics 730 12.2.0 [1.3.29735+27]
[opencl:cpu][opencl:0] Intel(R) OpenCL, 13th Gen Intel(R) Core(TM) i5-13400 OpenCL 3.0 (Build 0) [2025.20.8.0.06_160000]
[opencl:gpu][opencl:1] Intel(R) OpenCL Graphics, Intel(R) Arc(TM) A770 Graphics OpenCL 3.0 NEO  [24.39.31294]
[opencl:gpu][opencl:2] Intel(R) OpenCL Graphics, Intel(R) UHD Graphics 730 OpenCL 3.0 NEO  [24.39.31294]
```

### II. Build llama.cpp

#### Intel GPU

```sh
./examples/sycl/build.sh
```

or manually:

```sh
source /opt/intel/oneapi/setvars.sh

# Option 1: FP32 (recommended for most cases)
cmake -B build -DGGML_SYCL=ON -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx

# Option 2: FP16
cmake -B build -DGGML_SYCL=ON -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx -DGGML_SYCL_F16=ON

cmake --build build --config Release -j -v
```

> [!tip]
> Precision issues during tests can be resolved with:
> `export SYCL_PROGRAM_COMPILE_OPTIONS="-cl-fp32-correctly-rounded-divide-sqrt"`

### III. Run the inference

#### Retrieve and prepare model

See the general [*Obtaining and quantizing models*](../../README.md#obtaining-and-quantizing-models) guide, or download quantized models like [llama-2-7b.Q4_0.gguf](https://huggingface.co/TheBloke/Llama-2-7B-GGUF/resolve/main/llama-2-7b.Q4_0.gguf?download=true) or [Meta-Llama-3-8B-Instruct-Q4_0.gguf](https://huggingface.co/aptha/Meta-Llama-3-8B-Instruct-Q4_0-GGUF/resolve/main/Meta-Llama-3-8B-Instruct-Q4_0.gguf).

#### Check device

1. Enable oneAPI:

```sh
source /opt/intel/oneapi/setvars.sh
```

2. List SYCL devices:

```sh
./build/bin/llama-ls-sycl-device
```

Default backend is level_zero. Example output with 2 Intel GPUs:

```
found 2 SYCL devices:

|  |                  |                                             |Compute   |Max compute|Max work|Max sub|               |
|ID|       Device Type|                                         Name|capability|units      |group   |group  |Global mem size|
|--|------------------|---------------------------------------------|----------|-----------|--------|-------|---------------|
| 0|[level_zero:gpu:0]|               Intel(R) Arc(TM) A770 Graphics|       1.3|        512|    1024|     32|    16225243136|
| 1|[level_zero:gpu:1]|                    Intel(R) UHD Graphics 770|       1.3|         32|     512|     32|    53651849216|
```

#### Choose level-zero devices

| Device ID | Setting |
|-|-|
| 0 | `export ONEAPI_DEVICE_SELECTOR="level_zero:0"` or no action |
| 1 | `export ONEAPI_DEVICE_SELECTOR="level_zero:1"` |
| 0 & 1 | `export ONEAPI_DEVICE_SELECTOR="level_zero:0;level_zero:1"` |

#### Execute

**Script method:**

```sh
# Single device 0
./examples/sycl/test.sh -mg 0

# Multiple devices
./examples/sycl/test.sh

# Run llama-server
./examples/sycl/start-svr.sh -m PATH/MODEL_FILE
```

**Command line method:**

| Device selection | Parameter |
|------------------|-----------|
| Single device | `--split-mode none --main-gpu DEVICE_ID` |
| Multiple devices | `--split-mode layer` (default) |

Device 0:

```sh
ZES_ENABLE_SYSMAN=1 ./build/bin/llama-completion -no-cnv -m models/llama-2-7b.Q4_0.gguf -p "Building a website can be done in 10 simple steps:" -n 400 -e -ngl 99 -sm none -mg 0 --mmap
```

Multiple devices:

```sh
ZES_ENABLE_SYSMAN=1 ./build/bin/llama-completion -no-cnv -m models/llama-2-7b.Q4_0.gguf -p "Building a website can be done in 10 simple steps:" -n 400 -e -ngl 99 -sm layer --mmap
```

> [!info]
> Verify selected device(s) in output log:
> ```
> detect 1 SYCL GPUs: [0] with top Max compute units:512
> ```
> or
> ```
> use 1 SYCL GPUs: [0] with Max compute units:512
> ```

## Windows

### Install GPU driver

[Get Intel GPU Drivers](https://www.intel.com/content/www/us/en/products/docs/discrete-gpus/arc/software/drivers.html).

### Option 1: Download binary package

Download from [releases](https://github.com/ggml-org/llama.cpp/releases). Extract and run directly. Package includes SYCL runtime and all required DLLs — no oneAPI install needed.

### Option 2: Build from source

#### I. Setup environment

1. **Install Visual Studio** — skip if recent version already installed. See [Microsoft Visual Studio](https://visualstudio.microsoft.com/).

2. **Install Intel® oneAPI Base toolkit** — same dependencies as Linux (oneAPI DPC++/C++ compiler/runtime, oneDPL, oneDNN, oneMKL). Recommended: **Intel® Deep Learning Essentials**. Default install path: `C:\Program Files (x86)\Intel\oneAPI`.

Enable runtime:

- **CMD**: `"C:\Program Files (x86)\Intel\oneAPI\setvars.bat" intel64`
- **PowerShell**: `cmd.exe "/K" '"C:\Program Files (x86)\Intel\oneAPI\setvars.bat" && powershell'`

Verify: run `sycl-ls.exe`. Expect `[ext_oneapi_level_zero:gpu]` device(s). Example:

```
[opencl:acc:0] Intel(R) FPGA Emulation Platform for OpenCL(TM), Intel(R) FPGA Emulation Device OpenCL 1.2  [2023.16.10.0.17_160000]
[opencl:cpu:1] Intel(R) OpenCL, 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz OpenCL 3.0 (Build 0) [2023.16.10.0.17_160000]
[opencl:gpu:2] Intel(R) OpenCL Graphics, Intel(R) Iris(R) Xe Graphics OpenCL 3.0 NEO  [31.0.101.5186]
[ext_oneapi_level_zero:gpu:0] Intel(R) Level-Zero, Intel(R) Iris(R) Xe Graphics 1.3 [1.3.28044]
```

3. **Install build tools** — [CMake](https://cmake.org/download/) (also in VS Installer). Ninja comes with Visual Studio; install manually from [ninja-build.org](https://ninja-build.org/) if needed.

#### II. Build llama.cpp

##### Option 1: Script

```sh
.\examples\sycl\win-build-sycl.bat
```

##### Option 2: CMake

```bat
@call "C:\Program Files (x86)\Intel\oneAPI\setvars.bat" intel64 --force

# FP32 (recommended)
cmake -B build -G "Ninja" -DGGML_SYCL=ON -DCMAKE_C_COMPILER=cl -DCMAKE_CXX_COMPILER=icx  -DCMAKE_BUILD_TYPE=Release

# FP16
cmake -B build -G "Ninja" -DGGML_SYCL=ON -DCMAKE_C_COMPILER=cl -DCMAKE_CXX_COMPILER=icx  -DCMAKE_BUILD_TYPE=Release -DGGML_SYCL_F16=ON

cmake --build build --config Release -j
```

Or use CMake presets:

```sh
cmake --preset x64-windows-sycl-release
cmake --build build-x64-windows-sycl-release -j --target llama-completion

cmake -DGGML_SYCL_F16=ON --preset x64-windows-sycl-release
cmake --build build-x64-windows-sycl-release -j --target llama-completion

cmake --preset x64-windows-sycl-debug
cmake --build build-x64-windows-sycl-debug -j --target llama-completion
```

##### Option 3: Visual Studio

**As CMake Project:** Open `llama.cpp` folder directly. Select preset `x64-windows-sycl-release` or `x64-windows-sycl-debug`. Minimal build:

```Powershell
cmake --build build --config Release -j --target llama-completion
```

**As Visual Studio Solution:** Convert CMake to `.sln`:

```Powershell
# Full project with Intel compiler
cmake -B build -G "Visual Studio 17 2022" -T "Intel C++ Compiler 2025" -A x64 -DGGML_SYCL=ON -DCMAKE_BUILD_TYPE=Release

# Or Intel compiler only for ggml-sycl (requires -DBUILD_SHARED_LIBRARIES=ON, which is default)
cmake -B build -G "Visual Studio 17 2022" -A x64 -DGGML_SYCL=ON -DCMAKE_BUILD_TYPE=Release \
      -DSYCL_INCLUDE_DIR="C:\Program Files (x86)\Intel\oneAPI\compiler\latest\include" \
      -DSYCL_LIBRARY_DIR="C:\Program Files (x86)\Intel\oneAPI\compiler\latest\lib"
```

Open `build/llama.cpp.sln` in Visual Studio, then:

1. Right-click `ggml-sycl` → **Properties**
2. Expand **C/C++** → select **DPC++**
3. Set **Enable SYCL Offload** to `Yes`
4. Apply and save

Build from menu: `Build -> Build Solution`. Output in `build/Release/bin`.

> [!tip]
> Set environment variables `SYCL_INCLUDE_DIR_HINT` and `SYCL_LIBRARY_DIR_HINT` to avoid specifying `SYCL_INCLUDE_DIR`/`SYCL_LIBRARY_DIR`. Tested with VS 17 Community + oneAPI 2025.0.

### III. Run the inference

#### Retrieve and prepare model

Same as Linux. Download quantized models: [llama-2-7b.Q4_0.gguf](https://huggingface.co/TheBloke/Llama-2-7B-GGUF/blob/main/llama-2-7b.Q4_0.gguf) or [Meta-Llama-3-8B-Instruct-Q4_0.gguf](https://huggingface.co/aptha/Meta-Llama-3-8B-Instruct-Q4_0-GGUF/resolve/main/Meta-Llama-3-8B-Instruct-Q4_0.gguf).

#### Check device

1. Enable oneAPI: `"C:\Program Files (x86)\Intel\oneAPI\setvars.bat" intel64`
2. List devices: `build\bin\llama-ls-sycl-device.exe`

#### Choose level-zero devices

| Device ID | Setting |
|-|-|
| 0 | Default. Optionally `set ONEAPI_DEVICE_SELECTOR="level_zero:0"` |
| 1 | `set ONEAPI_DEVICE_SELECTOR="level_zero:1"` |
| 0 & 1 | `set ONEAPI_DEVICE_SELECTOR="level_zero:0;level_zero:1"` or `set ONEAPI_DEVICE_SELECTOR="level_zero:*"` |

#### Execute

**Script:**

```bat
examples\sycl\win-test.bat
examples\sycl\win-start-svr.bat -m PATH\MODEL_FILE
```

**Command line** — same device selection modes as Linux:

| Device selection | Parameter |
|------------------|-----------|
| Single device | `--split-mode none --main-gpu DEVICE_ID` |
| Multiple devices | `--split-mode layer` (default) |

Device 0:

```bat
build\bin\llama-completion.exe -no-cnv -m models\llama-2-7b.Q4_0.gguf -p "Building a website can be done in 10 simple steps:\nStep 1:" -n 400 -e -ngl 99 -sm none -mg 0 --mmap
```

Multiple devices:

```bat
build\bin\llama-completion.exe -no-cnv -m models\llama-2-7b.Q4_0.gguf -p "Building a website can be done in 10 simple steps:\nStep 1:" -n 400 -e -ngl 99 -sm layer --mmap
```

## Environment Variable

### Build

| Name | Value | Function |
|-|-|-|
| GGML_SYCL | ON (mandatory) | Enable SYCL code path. |
| GGML_SYCL_TARGET | INTEL *(default)* | SYCL target device type. |
| GGML_SYCL_DEVICE_ARCH | Optional | SYCL device architecture. Improves performance. See [offload-arch list](https://github.com/intel/llvm/blob/sycl/sycl/doc/design/OffloadDesign.md#--offload-arch). |
| GGML_SYCL_F16 | OFF *(default)* \| ON | Enable FP16 build. Different performance impact per LLM — test both. Requires rebuild. |
| GGML_SYCL_GRAPH | OFF *(default)* \| ON | Enable [SYCL Graph extension](https://github.com/intel/llvm/blob/sycl/sycl/doc/extensions/experimental/sycl_ext_oneapi_graph.asciidoc). |
| GGML_SYCL_DNN | ON *(default)* \| OFF | Enable oneDNN. |
| GGML_SYCL_HOST_MEM_FALLBACK | ON *(default)* \| OFF | Host memory fallback when device memory full during quantized weight reorder. Continues at reduced speed (PCIe) instead of failing. Requires Linux kernel 6.8+. |
| CMAKE_C_COMPILER | `icx` (Linux), `icx/cl` (Windows) | Compiler for SYCL. |
| CMAKE_CXX_COMPILER | `icpx` (Linux), `icx` (Windows) | Compiler for SYCL. |

### Runtime

| Name | Value | Function |
|-|-|-|
| GGML_SYCL_DEBUG | 0 (default) or 1 | Enable debug logging. |
| GGML_SYCL_ENABLE_FLASH_ATTN | 1 (default) or 0 | Flash-Attention. Reduces memory usage. Performance impact varies by LLM. |
| GGML_SYCL_DISABLE_OPT | 0 (default) or 1 | Disable Intel GPU optimizations. Set to 1 for pre-Gen 10 devices. |
| GGML_SYCL_DISABLE_GRAPH | 0 or 1 (default) | Disable SYCL Graphs. Default off (still in development). |
| GGML_SYCL_DISABLE_DNN | 0 (default) or 1 | Disable oneDNN; always use oneMKL. |
| ZES_ENABLE_SYSMAN | 0 (default) or 1 | Get free GPU memory via `sycl::aspect::ext_intel_free_memory`. Recommended with `--split-mode layer`. |
| UR_L0_ENABLE_RELAXED_ALLOCATION_LIMITS | 0 (default) or 1 | Support malloc device memory >4GB. |

## Design Rule

- All code changes must benefit users: fix bugs, add features, improve performance/usage, or improve maintainability.
- Reject: breaking legacy features, reducing default performance of legacy cases, incomplete/undemonstrated work.
- Prefer environment variables for feature toggles — no rebuild needed for user evaluation.
- Design based on published official oneAPI releases (compiler, library, driver, OS kernel).
- Developers maintain their submitted code.

## Known Issues

- `Split-mode:[row]` not supported.
- No AOT (Ahead-of-Time) compilation: fast build + small binary, but slow first startup (JIT). Subsequent runs unaffected.

## Q&A

- **`error while loading shared libraries: libsycl.so`** — Install oneAPI and run `source /opt/intel/oneapi/setvars.sh`.
- **General compiler error** — Remove `build` folder or do a clean build.
- **`[ext_oneapi_level_zero:gpu]` not visible after driver install on Linux** — Run `sudo sycl-ls`. If present, add `video`/`render` groups and logout/login or restart. Otherwise re-check driver install.
- **Ollama issues on Intel GPU** — Cannot support directly. Reproduce on llama.cpp and report there.
- **`Native API returns: 39 (UR_RESULT_ERROR_OUT_OF_DEVICE_MEMORY)` / `failed to allocate SYCL0 buffer`** — Out of device memory.

  | Reason | Solution |
  |-|-|
  | Default context too big (excessive memory usage) | Set `-c 8192` or smaller |
  | Model too large for available memory | Use smaller model or quantization (Q5→Q4); use multiple devices |

- **`can't allocate 5000000000 Bytes`** — Enable >4GB support:
  ```
  export UR_L0_ENABLE_RELAXED_ALLOCATION_LIMITS=1
  set UR_L0_ENABLE_RELAXED_ALLOCATION_LIMITS=1
  ```

> [!tip]
> **GitHub contributions**: Add `[SYCL]` prefix in issues/PR titles for faster triage.

## TODO

- Review ZES_ENABLE_SYSMAN: https://github.com/intel/compute-runtime/blob/master/programmers-guide/SYSMAN.md#support-and-limitations

#sycl #intel-gpu #oneapi #hardware-acceleration
