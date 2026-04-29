---
title: Build
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/build.md
source: git
fetched_at: 2026-04-28T09:49:22.034382563-03:00
rendered_js: false
word_count: 1458
summary: This document provides comprehensive instructions for compiling and building the llama.cpp project from source across various operating systems and hardware backends.
tags:
    - llama-cpp
    - build-instructions
    - cmake
    - gpu-acceleration
    - compiler-configuration
category: guide
optimized: true
optimized_at: 2026-04-28T00:00:00Z
---
# Build llama.cpp locally

The main product is the `llama` library (C interface in [include/llama.h](../include/llama.h)). The project also includes example programs and tools ranging from minimal snippets to an OpenAI-compatible HTTP server.

```bash
git clone https://github.com/ggml-org/llama.cpp
cd llama.cpp
```

## CPU Build

```bash
cmake -B build
cmake --build build --config Release
```

Notes:

- Add `-j` for parallel jobs, or use Ninja generator. Example: `cmake --build build --config Release -j 8`.
- Install [ccache](https://ccache.dev/) for faster repeated compilation.
- **Debug builds**:
  1. Single-config generators (default Unix Makefiles):
     ```bash
     cmake -B build -DCMAKE_BUILD_TYPE=Debug
     cmake --build build
     ```
  2. Multi-config generators (Visual Studio, XCode):
     ```bash
     cmake -B build -G "Xcode"
     cmake --build build --config Debug
     ```
  See [CMake documentation](https://cmake.org/cmake/help/latest/manual/cmake-generators.7.html).
- **Static builds**: add `-DBUILD_SHARED_LIBS=OFF`.
- **Windows (x86, x64, arm64) with MSVC/Clang**:
  - Install [Visual Studio 2022 Community](https://visualstudio.microsoft.com/vs/community/) with: Desktop-development with C++ workload; C++ CMake Tools, Git for Windows, C++ Clang Compiler, MS-Build Support for LLVM-Toolset components.
  - Always use Developer Command Prompt / PowerShell for VS2022.
  - Windows ARM (arm64):
    ```bash
    cmake --preset arm64-windows-llvm-release -D GGML_OPENMP=OFF
    cmake --build build-arm64-windows-llvm-release
    ```
  - x64 with Ninja + Clang:
    ```bash
    cmake --preset x64-windows-llvm-release
    cmake --build build-x64-windows-llvm-release
    ```
- **HTTPS/TLS** (optional): Install OpenSSL dev libraries. Without it, the project builds without SSL.
  - Debian/Ubuntu: `sudo apt-get install libssl-dev`
  - Fedora/RHEL/Rocky/Alma: `sudo dnf install openssl-devel`
  - Arch/Manjaro: `sudo pacman -S openssl`

## BLAS Build

BLAS improves prompt processing performance for batch sizes >32 (default 512). Does not affect generation performance.

### Accelerate Framework

Mac only, enabled by default. Build normally.

### OpenBLAS

```bash
cmake -B build -DGGML_BLAS=ON -DGGML_BLAS_VENDOR=OpenBLAS
cmake --build build --config Release
```

### BLIS

See [[014-backend-blis|BLIS backend]].

### Intel oneMKL

Enables `avx_vnni` for Intel processors without avx512/avx512_vnni. **Does not support Intel GPU** — use [[020-backend-sycl|SYCL]] for Intel GPU.

```bash
source /opt/intel/oneapi/setvars.sh  # skip in oneapi-basekit docker image
cmake -B build -DGGML_BLAS=ON -DGGML_BLAS_VENDOR=Intel10_64lp -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx -DGGML_NATIVE=ON
cmake --build build --config Release
```

Alternatively, build using [oneAPI-basekit](https://hub.docker.com/r/intel/oneapi-basekit) docker image.

See [Optimizing and Running LLaMA2 on Intel® CPU](https://builders.intel.com/solutionslibrary/optimizing-and-running-llama2-on-intel-cpu) for more info.

### Other BLAS libraries

Set `GGML_BLAS_VENDOR` option. See [CMake BLAS vendors](https://cmake.org/cmake/help/latest/module/FindBLAS.html#blas-lapack-vendors).

## Metal Build

Metal is enabled by default on macOS (GPU computation). Disable with `-DGGML_METAL=OFF`. At runtime, use `--n-gpu-layers 0` to disable GPU inference.

## SYCL

SYCL targets **Intel GPU** (Data Center Max, Flex, Arc, Built-in GPU, iGPU). See [[020-backend-sycl|llama.cpp for SYCL]].

## CUDA

Provides NVIDIA GPU acceleration. Requires [CUDA toolkit](https://developer.nvidia.com/cuda-toolkit).

- Download from [NVIDIA developer site](https://developer.nvidia.com/cuda-downloads).
- For Fedora toolbox container setup, see [[015-backend-cuda-fedora|CUDA on Fedora guide]].
  - **Necessary** for [Atomic Desktops for Fedora](https://fedoraproject.org/atomic-desktops/) (Silverblue, Kinoite) — no supported CUDA packages.
  - **Necessary** for non-[supported CUDA platforms](https://developer.nvidia.com/cuda-downloads) (e.g. Fedora 42 Beta).
  - **Convenient** for Fedora Workstation/KDE wanting a clean host.
  - Also available for Arch Linux, RHEL ≥8.5, Ubuntu.

### Compilation

```bash
cmake -B build -DGGML_CUDA=ON
cmake --build build --config Release
```

Read CPU build notes for compilation speed tips.

### Non-Native Builds

Default builds for connected hardware. For universal CUDA GPU support, disable `GGML_NATIVE`:

```bash
cmake -B build -DGGML_CUDA=ON -DGGML_NATIVE=OFF
```

### Override Compute Capability

If `nvcc` can't detect your GPU:

```text
nvcc warning : Cannot find valid GPU for '-arch=native', default arch is used
```

Options: non-native build (large binary), or explicitly specify architectures. See `ggml/src/ggml-cuda/CMakeLists.txt` for logic.

1. Find your GPU's [Compute Capability](https://developer.nvidia.com/cuda-gpus):

   ```text
   GeForce RTX 4090      8.9
   GeForce RTX 3080 Ti   8.6
   GeForce RTX 3070      8.6
   ```

2. Set architectures:

   ```bash
   cmake -B build -DGGML_CUDA=ON -DCMAKE_CUDA_ARCHITECTURES="86;89"
   ```

### Overriding the CUDA Version

For specific CUDA version (e.g. CUDA 11.7 at `/opt/cuda-11.7`):

```bash
cmake -B build -DGGML_CUDA=ON -DCMAKE_CUDA_COMPILER=/opt/cuda-11.7/bin/nvcc -DCMAKE_INSTALL_RPATH="/opt/cuda-11.7/lib64;\$ORIGIN" -DCMAKE_BUILD_WITH_INSTALL_RPATH=ON
```

#### Fixing Old CUDA + New glibc Compatibility

Old CUDA (e.g. v11.7) with new glibc produces errors like `exception specification incompatible`. Patch `/path/to/cuda/targets/x86_64-linux/include/crt/math_functions.h`:

```C++
// original lines
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ double                 cospi(double x);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ float                  cospif(float x);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ double                 sinpi(double x);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ float                  sinpif(float x);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ double                 rsqrt(double x);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ float                  rsqrtf(float x);

// edited lines
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ double                 cospi(double x) noexcept (true);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ float                  cospif(float x) noexcept (true);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ double                 sinpi(double x) noexcept (true);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ float                  sinpif(float x) noexcept (true);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ double                 rsqrt(double x) noexcept (true);
extern __DEVICE_FUNCTIONS_DECL__ __device_builtin__ float                  rsqrtf(float x) noexcept (true);
```

### Runtime CUDA Environment Variables

```bash
# Use `CUDA_VISIBLE_DEVICES` to hide the first compute device.
CUDA_VISIBLE_DEVICES="-0" ./build/bin/llama-server --model /srv/models/llama.gguf
```

#### CUDA_SCALE_LAUNCH_QUEUES

[`CUDA_SCALE_LAUNCH_QUEUES`](https://docs.nvidia.com/cuda/cuda-programming-guide/05-appendices/environment-variables.html#cuda-scale-launch-queues) controls CUDA command buffer size. Set `CUDA_SCALE_LAUNCH_QUEUES=4x` for 4× default — especially beneficial for **multi-GPU pipeline parallelism** prompt processing throughput.

#### GGML_CUDA_FORCE_CUBLAS_COMPUTE_32F

Use FP32 compute type in FP16 cuBLAS to prevent numerical overflows (slower prompt processing; small impact on RTX PRO/Datacenter, significant on GeForce).

#### GGML_CUDA_FORCE_CUBLAS_COMPUTE_16F

Force FP16 compute type in FP16 cuBLAS for V100, CDNA, and RDNA4.

### Unified Memory

`GGML_CUDA_ENABLE_UNIFIED_MEMORY=1` enables swapping to system RAM when VRAM exhausted (Linux). On Windows, available as `System Memory Fallback` in NVIDIA control panel.

### Peer Access

`GGML_CUDA_P2P` enables peer-to-peer data transfer between GPUs (requires driver support, usually workstation/datacenter). May cause crashes/corruption on some motherboards with IOMMU.

### Performance Tuning

| Option | Legal values | Default | Description |
|-|-|-|-|
| GGML_CUDA_FORCE_MMQ | Boolean | false | Force custom matrix multiplication kernels for quantized models instead of FP16 cuBLAS (affects V100, CDNA, RDNA3+). Lower VRAM but worse large-batch speed. |
| GGML_CUDA_FORCE_CUBLAS | Boolean | false | Force FP16 cuBLAS over custom kernels. Possible numerical overflows, higher memory. May be faster on recent datacenter GPUs. |
| GGML_CUDA_PEER_MAX_BATCH_SIZE | Positive integer | 128 | Max batch size for multi-GPU peer access. With NVLink, may benefit larger batches. |
| GGML_CUDA_FA_ALL_QUANTS | Boolean | false | Compile FlashAttention support for all KV cache quantization types. Longer compile time. |

## MUSA

Moore Threads GPU acceleration. Requires [MUSA SDK](https://developer.mthreads.com/musa/musa-sdk). Download from [Moore Threads developer site](https://developer.mthreads.com/sdk/download/musa).

### Compilation

```bash
cmake -B build -DGGML_MUSA=ON
cmake --build build --config Release
```

Override compute capabilities:

```bash
cmake -B build -DGGML_MUSA=ON -DMUSA_ARCHITECTURES="21"
cmake --build build --config Release
```

`21` = MTT S80 (compute capability 2.1). Most CUDA compilation options also apply to MUSA.

Static builds:

```bash
cmake -B build -DGGML_MUSA=ON \
  -DBUILD_SHARED_LIBS=OFF -DCMAKE_POSITION_INDEPENDENT_CODE=ON
cmake --build build --config Release
```

### Runtime MUSA Environment Variables

```bash
# Use `MUSA_VISIBLE_DEVICES` to hide the first compute device.
MUSA_VISIBLE_DEVICES="-0" ./build/bin/llama-server --model /srv/models/llama.gguf
```

### Unified Memory

`GGML_CUDA_ENABLE_UNIFIED_MEMORY=1` enables unified memory on Linux (swap to system RAM). Hurts performance on non-integrated GPUs but enables integrated GPU usage.

## HIP

AMD GPU acceleration via HIP/ROCm. Install from distro packages or [ROCm Quick Start (Linux)](https://rocm.docs.amd.com/projects/install-on-linux/en/latest/tutorial/quick-start.html#rocm-install-quick).

### Linux

```bash
HIPCXX="$(hipconfig -l)/clang" HIP_PATH="$(hipconfig -R)" \
    cmake -S . -B build -DGGML_HIP=ON -DGPU_TARGETS=gfx1030 -DCMAKE_BUILD_TYPE=Release \
    && cmake --build build --config Release -- -j 16
```

`GPU_TARGETS` is optional (builds for current system GPUs if omitted).

#### Flash Attention with rocWMMA

For RDNA3+ or CDNA, enable `-DGGML_HIP_ROCWMMA_FATTN=ON`. Requires rocWMMA headers. Included by default with the `rocm` meta package, or install `rocwmma-dev`/`rocwmma-devel` separately. Alternatively, clone from [GitHub](https://github.com/ROCm/rocWMMA), checkout tag (e.g. `rocm-6.2.4`), and set `-DCMAKE_CXX_FLAGS="-I<path/to/rocwmma>/library/include/"`. Works on Windows too (unofficially AMD-supported).

#### Fixing `cannot find ROCm device library` Error

```bash
HIPCXX="$(hipconfig -l)/clang" HIP_PATH="$(hipconfig -p)" \
HIP_DEVICE_LIB_PATH=<directory-containing-oclc_abi_version_400.bc> \
    cmake -S . -B build -DGGML_HIP=ON -DGPU_TARGETS=gfx1030 -DCMAKE_BUILD_TYPE=Release \
    && cmake --build build -- -j 16
```

### Windows

```bash
set PATH=%HIP_PATH%\bin;%PATH%
cmake -S . -B build -G Ninja -DGPU_TARGETS=gfx1100 -DGGML_HIP=ON -DCMAKE_C_COMPILER=clang -DCMAKE_CXX_COMPILER=clang++ -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

`gfx1100` = Radeon RX 7900XTX/XT/GRE. Find your GPU arch: `rocminfo | grep gfx | head -1 | awk '{print $2}'`. See [AMDGPU targets](https://llvm.org/docs/AMDGPUUsage.html#processors).

`HIP_VISIBLE_DEVICES` selects GPUs. For unsupported GPUs, set `HSA_OVERRIDE_GFX_VERSION` (e.g. `10.3.0` for RDNA2 gfx1030/1031/1035, `11.0.0` for RDNA3). Not [supported on Windows](https://github.com/ROCm/ROCm/issues/2654).

### Unified Memory

`GGML_CUDA_ENABLE_UNIFIED_MEMORY=1` shares main memory with integrated GPU on Linux. Hurts non-integrated GPU performance.

## Vulkan

### Windows

**w64devkit**: Download [`w64devkit`](https://github.com/skeeto/w64devkit/releases) and [Vulkan SDK](https://vulkan.lunarg.com/sdk/home#windows). Copy Vulkan dependencies:

```sh
SDK_VERSION=1.3.283.0
cp /VulkanSDK/$SDK_VERSION/Bin/glslc.exe $W64DEVKIT_HOME/bin/
cp /VulkanSDK/$SDK_VERSION/Lib/vulkan-1.lib $W64DEVKIT_HOME/x86_64-w64-mingw32/lib/
cp -r /VulkanSDK/$SDK_VERSION/Include/* $W64DEVKIT_HOME/x86_64-w64-mingw32/include/
cat > $W64DEVKIT_HOME/x86_64-w64-mingw32/lib/pkgconfig/vulkan.pc <<EOF
Name: Vulkan-Loader
Description: Vulkan Loader
Version: $SDK_VERSION
Libs: -lvulkan-1
EOF
```

Build:

```sh
cmake -B build -DGGML_VULKAN=ON
cmake --build build --config Release
```

**Git Bash MINGW64**: Install [Git-SCM](https://git-scm.com/downloads/win), [Visual Studio Community](https://visualstudio.microsoft.com/) with C++, [CMake](https://cmake.org/download/), [Vulkan SDK](https://vulkan.lunarg.com/sdk/home#windows). Build in Git Bash:

```
cmake -B build -DGGML_VULKAN=ON
cmake --build build --config Release
```

Test:

```sh
build/bin/Release/llama-cli -m "[PATH TO MODEL]" -ngl 100 -c 16384 -t 10 -n -2 -cnv
```

**MSYS2**: Install [MSYS2](https://www.msys2.org/) and dependencies in UCRT terminal:

```sh
pacman -S git \
    mingw-w64-ucrt-x86_64-gcc \
    mingw-w64-ucrt-x86_64-cmake \
    mingw-w64-ucrt-x86_64-vulkan-devel \
    mingw-w64-ucrt-x86_64-shaderc \
    mingw-w64-ucrt-x86_64-spirv-headers
```

Build:

```sh
cmake -B build -DGGML_VULKAN=ON
cmake --build build --config Release
```

### Docker

Vulkan SDK installs inside the container automatically.

```sh
# Build the image
docker build -t llama-cpp-vulkan --target light -f .devops/vulkan.Dockerfile .

# Then, use it:
docker run -it --rm -v "$(pwd):/app:Z" --device /dev/dri/renderD128:/dev/dri/renderD128 --device /dev/dri/card1:/dev/dri/card1 llama-cpp-vulkan -m "/app/models/YOUR_MODEL_FILE" -p "Building a website can be done in 10 simple steps:" -n 400 -e -ngl 33
```

### Linux

#### LunarG Vulkan SDK

Follow [Getting Started with the Linux Tarball Vulkan SDK](https://vulkan.lunarg.com/doc/sdk/latest/linux/getting_started.html).

> [!warning]
> You **must** `source setup_env.sh` from the Vulkan SDK in each terminal session. Without this, the build won't work.

#### System packages

Debian/Ubuntu:

```sh
sudo apt-get install libvulkan-dev glslc spirv-headers
```

SPIRV-Headers (`spirv/unified1/spirv.hpp`) may not be included with the Vulkan loader dev package. Other distros: `spirv-headers` (Arch), `spirv-headers-devel` (Fedora/openSUSE). Windows: included in LunarG SDK `Include` directory.

#### Build

Verify setup: `vulkaninfo`. Then:

```bash
cmake -B build -DGGML_VULKAN=1
cmake --build build --config Release
```

Test: `./build/bin/llama-cli -m "PATH_TO_MODEL" -p "Hi you how are you" -ngl 99`. Output should show `ggml_vulkan: Using <GPU>`.

### macOS

Follow [Getting Started with the MacOS Vulkan SDK](https://vulkan.lunarg.com/doc/sdk/latest/mac/getting_started.html). Two Vulkan drivers (both translate to Metal); swap via `VK_ICD_FILENAMES`.

Check "KosmicKrisp" during LunarG SDK install. Set environment:

```bash
source /path/to/vulkan-sdk/setup-env.sh
```

#### MoltenVK

Default driver installed with LunarG SDK. Use as-is.

#### KosmicKrisp

```bash
export VK_ICD_FILENAMES=$VULKAN_SDK/share/vulkan/icd.d/libkosmickrisp_icd.json
export VK_DRIVER_FILES=$VULKAN_SDK/share/vulkan/icd.d/libkosmickrisp_icd.json
```

#### Build

```bash
cmake -B build -DGGML_VULKAN=1 -DGGML_METAL=OFF
cmake --build build --config Release
```

## CANN

NPU acceleration via Ascend NPU AI cores. Requires [CANN toolkit](https://www.hiascend.com/developer/download/community/result?module=cann). See [Ascend Community](https://www.hiascend.com/en/).

```bash
cmake -B build -DGGML_CANN=on -DCMAKE_BUILD_TYPE=release
cmake --build build --config release
```

Test: `./build/bin/llama-cli -m PATH_TO_MODEL -p "Building a website can be done in 10 steps:" -ngl 32`. Successful CANN backend output includes `CANN model buffer size` and `CANN compute buffer size` lines.

See [[039-backend-cann|llama.cpp for CANN]] for model/device support and install details.

## ZenDNN

Optimized deep learning primitives for AMD EPYC™ CPUs (matrix multiplication acceleration).

### Compilation

```bash
cmake -B build -DGGML_ZENDNN=ON
cmake --build build --config Release
```

First build auto-downloads and builds ZenDNN (5-10 min). Subsequent builds faster.

Custom ZenDNN path:

```bash
cmake -B build -DGGML_ZENDNN=ON -DZENDNN_ROOT=/path/to/zendnn/install
cmake --build build --config Release
```

### Testing

```bash
./build/bin/llama-cli -m PATH_TO_MODEL -p "Building a website can be done in 10 steps:" -n 50
```

See [[023-backend-zendnn|llama.cpp for ZenDNN]] for hardware support, setup, and performance details.

## Arm® KleidiAI™

Optimized microkernels for Arm CPUs (dotprod, int8mm, SVE, SME). Enable:

```bash
cmake -B build -DGGML_CPU_KLEIDIAI=ON
cmake --build build --config Release
```

Verify: run `./build/bin/llama-cli -m PATH_TO_MODEL -p "What is a car?"`. Output should contain `load_tensors: CPU_KLEIDIAI model buffer size = ...`.

SME microkernels auto-enable via runtime detection on supported CPUs. Environment variable `GGML_KLEIDIAI_SME`:

| Value | Behavior |
|-|-|
| Not set | Auto-enable SME if supported |
| 0 | Disable SME |
| `<n>` > 0 | Enable SME with `<n>` assumed units (override auto-detect) |

If SME unsupported by CPU, always disabled.

> [!warning]
> Higher-priority backends may override CPU backend. Disable them at compile time (e.g. `-DGGML_METAL=OFF`) or at runtime (`--device none`).

## OpenCL

GPU acceleration via OpenCL on recent Adreno GPUs. See [[016-backend-opencl|OPENCL.md]].

### Android (OpenCL)

Assumes NDK at `$ANDROID_NDK`. Install OpenCL headers and ICD loader:

```sh
mkdir -p ~/dev/llm
cd ~/dev/llm

git clone https://github.com/KhronosGroup/OpenCL-Headers && \
cd OpenCL-Headers && \
cp -r CL $ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include

cd ~/dev/llm

git clone https://github.com/KhronosGroup/OpenCL-ICD-Loader && \
cd OpenCL-ICD-Loader && \
mkdir build_ndk && cd build_ndk && \
cmake .. -G Ninja -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_TOOLCHAIN_FILE=$ANDROID_NDK/build/cmake/android.toolchain.cmake \
  -DOPENCL_ICD_LOADER_HEADERS_DIR=$ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=24 \
  -DANDROID_STL=c++_shared && \
ninja && \
cp libOpenCL.so $ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/lib/aarch64-linux-android
```

Build llama.cpp with OpenCL:

```sh
cd ~/dev/llm

git clone https://github.com/ggml-org/llama.cpp && \
cd llama.cpp && \
mkdir build-android && cd build-android

cmake .. -G Ninja \
  -DCMAKE_TOOLCHAIN_FILE=$ANDROID_NDK/build/cmake/android.toolchain.cmake \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-28 \
  -DBUILD_SHARED_LIBS=OFF \
  -DGGML_OPENCL=ON

ninja
```

### Windows Arm64 (OpenCL)

Install OpenCL headers and ICD loader:

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

Build llama.cpp with OpenCL:

```powershell
cmake .. -G Ninja `
  -DCMAKE_TOOLCHAIN_FILE="$HOME/dev/llm/llama.cpp/cmake/arm64-windows-llvm.cmake" `
  -DCMAKE_BUILD_TYPE=Release `
  -DCMAKE_PREFIX_PATH="$HOME/dev/llm/opencl" `
  -DBUILD_SHARED_LIBS=OFF `
  -DGGML_OPENCL=ON
ninja
```

## Android

See [[025-android|Android]] for build instructions.

## WebGPU [In Progress]

Relies on [Dawn](https://dawn.googlesource.com/dawn). Install per [Dawn CMake quickstart](https://dawn.googlesource.com/dawn/+/refs/heads/main/docs/quickstart-cmake.md). Current implementation: Dawn commit `18eb229`.

```bash
cmake -B build -DGGML_WEBGPU=ON
cmake --build build --config Release
```

### Browser Support

Uses [Emscripten](https://emscripten.org/) to compile to WebAssembly. Dawn maintains `emdawnwebgpu` bindings. Build locally to stay in sync with your Dawn version. Set `EMDAWNWEBGPU_DIR` CMake flag to the emdawnwebgpu port file path.

## IBM Z & LinuxONE

See [[026-build-s390x|Build llama.cpp locally (for s390x)]].

## OpenVINO

[OpenVINO](https://docs.openvino.ai/) optimizes and deploys AI inference on Intel hardware (CPUs, GPUs, NPUs). See [[017-backend-openvino|OPENVINO.md]].

## Notes about GPU-accelerated backends

- `-ngl 0` may still use GPU for some computation. Use `--device none` to fully disable.
- Multiple backends can coexist (e.g. `-DGGML_CUDA=ON -DGGML_VULKAN=ON`). Use `--device` at runtime; `--list-devices` shows available devices.
- Backends can be built as dynamic libraries via `GGML_BACKEND_DL` for runtime loading across different machines.

#build #cmake #gpu-acceleration #compiler-configuration
