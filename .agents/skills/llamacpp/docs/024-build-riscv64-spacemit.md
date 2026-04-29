---
title: Build riscv64 spacemit
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/build-riscv64-spacemit.md
source: git
fetched_at: 2026-04-28T09:49:20.263217178-03:00
rendered_js: false
word_count: 487
summary: This document provides instructions for building and simulating llama.cpp specifically for RISC-V SpacemiT SoC architectures, including build configuration requirements and performance benchmarking results.
tags:
    - risc-v
    - spacemit
    - llama-cpp
    - cross-compilation
    - qemu
    - soc-optimization
category: guide
optimized: true
optimized_at: '2026-04-28T00:00:00Z'
---
# Build riscv64 spacemit

> [!important]
> This build documentation is specific only to RISC-V SpacemiT SOCs.

## Build llama.cpp locally (for riscv64)

### 1. Prepare Toolchain

~~~
wget https://archive.spacemit.com/toolchain/spacemit-toolchain-linux-glibc-x86_64-v1.1.2.tar.xz
~~~

### 2. Build

The build requires RISC-V vector instructions for acceleration. Enable `GGML_CPU_RISCV64_SPACEMIT` with the `RISCV64_SPACEMIT_IME1` optimization version. Ensure the RISC-V compiler is installed and `RISCV_ROOT_PATH` is set via `export RISCV_ROOT_PATH={your_compiler_path}`. Compiler configurations are in `riscv64-spacemit-linux-gnu-gcc.cmake`.

```bash

cmake -B build \
    -DCMAKE_BUILD_TYPE=Release \
    -DGGML_CPU_RISCV64_SPACEMIT=ON \
    -DLLAMA_OPENSSL=OFF \
    -DGGML_RVV=ON \
    -DGGML_RV_ZFH=ON \
    -DGGML_RV_ZICBOP=ON \
    -DGGML_RV_ZIHINTPAUSE=ON \
    -DRISCV64_SPACEMIT_IME_SPEC=RISCV64_SPACEMIT_IME1 \
    -DCMAKE_TOOLCHAIN_FILE=${PWD}/cmake/riscv64-spacemit-linux-gnu-gcc.cmake \
    -DCMAKE_INSTALL_PREFIX=build/installed

cmake --build build --parallel $(nproc) --config Release

pushd build
make install
popd
```

## Simulation

Use QEMU to emulate on non-RISC-V architectures.

### 1. Download QEMU

~~~
wget https://archive.spacemit.com/spacemit-ai/qemu/jdsk-qemu-v0.0.14.tar.gz
~~~

### 2. Run Simulation

After building llama.cpp, run via QEMU:

~~~
export QEMU_ROOT_PATH={your QEMU file path}
export RISCV_ROOT_PATH_IME1={your RISC-V compiler path}

${QEMU_ROOT_PATH}/bin/qemu-riscv64 -L ${RISCV_ROOT_PATH_IME1}/sysroot -cpu max,vlen=256,elen=64,vext_spec=v1.0 ${PWD}/build/bin/llama-cli -m ${PWD}/models/Qwen2.5-0.5B-Instruct-Q4_0.gguf -t 1
~~~

## Performance

Benchmarked on Spacemit(R) X60:

~~~
model name      : Spacemit(R) X60
isa             : rv64imafdcv_zicbom_zicboz_zicntr_zicond_zicsr_zifencei_zihintpause_zihpm_zfh_zfhmin_zca_zcd_zba_zbb_zbc_zbs_zkt_zve32f_zve32x_zve64d_zve64f_zve64x_zvfh_zvfhmin_zvkt_sscofpmf_sstc_svinval_svnapot_svpbmt
mmu             : sv39
uarch           : spacemit,x60
mvendorid       : 0x710
marchid         : 0x8000000058000001
~~~

### Q4_0

|   Model    |   Size   | Params | backend | threads | test | t/s |
| -----------| -------- | ------ | ------- | ------- | ---- |------|
Qwen2.5 0.5B |403.20 MiB|630.17 M|   cpu   |    4    | pp512|64.12 ± 0.26|
Qwen2.5 0.5B |403.20 MiB|630.17 M|   cpu   |    4    | tg128|10.03 ± 0.01|
Qwen2.5 1.5B |1011.16 MiB| 1.78 B |   cpu   |    4    | pp512|24.16 ± 0.02|
Qwen2.5 1.5B |1011.16 MiB| 1.78 B |   cpu   |    4    | tg128|3.83 ± 0.06|
Qwen2.5 3B   | 1.86 GiB  | 3.40 B |   cpu   |    4    | pp512|12.08 ± 0.02|
Qwen2.5 3B   | 1.86 GiB  | 3.40 B |   cpu   |    4    | tg128|2.23 ± 0.02|

### Q4_1

|   Model    |   Size   | Params | backend | threads | test | t/s |
| -----------| -------- | ------ | ------- | ------- | ---- |------|
Qwen2.5 0.5B |351.50 MiB|494.03 M|   cpu   |    4    | pp512|62.07 ± 0.12|
Qwen2.5 0.5B |351.50 MiB|494.03 M|   cpu   |    4    | tg128|9.91 ± 0.01|
Qwen2.5 1.5B |964.06 MiB| 1.54 B |   cpu   |    4    | pp512|22.95 ± 0.25|
Qwen2.5 1.5B |964.06 MiB| 1.54 B |   cpu   |    4    | tg128|4.01 ± 0.15|
Qwen2.5 3B   | 1.85 GiB | 3.09 B |   cpu   |    4    | pp512|11.55 ± 0.16|
Qwen2.5 3B   | 1.85 GiB | 3.09 B |   cpu   |    4    | tg128|2.25 ± 0.04|

### Q4_K

|   Model    |   Size   | Params | backend | threads | test | t/s |
| -----------| -------- | ------ | ------- | ------- | ---- |------|
Qwen2.5 0.5B |462.96 MiB|630.17 M|   cpu   |    4    | pp512|9.29 ± 0.05|
Qwen2.5 0.5B |462.96 MiB|630.17 M|   cpu   |    4    | tg128|5.67 ± 0.04|
Qwen2.5 1.5B | 1.04 GiB | 1.78 B |   cpu   |    4    | pp512|10.38 ± 0.10|
Qwen2.5 1.5B | 1.04 GiB | 1.78 B |   cpu   |    4    | tg128|3.17 ± 0.08|
Qwen2.5 3B   | 1.95 GiB | 3.40 B |   cpu   |    4    | pp512|4.23 ± 0.04|
Qwen2.5 3B   | 1.95 GiB | 3.40 B |   cpu   |    4    | tg128|1.73 ± 0.00|
