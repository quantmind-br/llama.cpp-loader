---
title: BLIS
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/BLIS.md
source: git
fetched_at: 2026-04-28T09:49:08.915748088-03:00
rendered_js: false
word_count: 102
summary: This document provides instructions for compiling, installing, and configuring the BLIS high-performance linear algebra framework for use with projects like llama.cpp.
tags:
    - blis-installation
    - linear-algebra
    - high-performance-computing
    - compilation-guide
    - multithreading
    - software-framework
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# BLIS

BLIS is a portable high-performance BLAS-like dense linear algebra framework (2023 James H. Wilkinson Prize, 2020 SIAM Activity Group on Supercomputing Best Paper Prize). It provides a BLAS-like API, typed API, and BLAS/CBLAS compatibility layers.

Project URL: https://github.com/flame/blis

## Prepare

Compile and install BLIS:

```bash
git clone https://github.com/flame/blis
cd blis
./configure --enable-cblas -t openmp,pthreads auto
# will install to /usr/local/ by default.
make -j
sudo make install
```

> [!tip]
> OpenMP is recommended since it's easier to modify the cores being used.

## llama.cpp compilation

```bash
mkdir build
cd build
cmake -DGGML_BLAS=ON -DGGML_BLAS_VENDOR=FLAME ..
make -j
```

## llama.cpp execution

Set OpenMP environment variables to control threading behavior:

```bash
export GOMP_CPU_AFFINITY="0-19"
export BLIS_NUM_THREADS=14
```

Then run binaries normally.

## Intel-specific issue

If you get `libimf.so` cannot be found, follow this [StackOverflow page](https://stackoverflow.com/questions/70687930/intel-oneapi-2022-libimf-so-no-such-file-or-directory-during-openmpi-compila).

## References

- https://github.com/flame/blis#getting-started
- https://github.com/flame/blis/blob/master/docs/Multithreading.md

#blis #linear-algebra #high-performance-computing