---
title: Configuration
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/VirtGPU/configuration.md
source: git
fetched_at: 2026-04-28T09:49:11.396077135-03:00
rendered_js: false
word_count: 454
summary: This document provides a comprehensive reference for the environment variables used to configure the ggml-virtgpu backend system across guest, hypervisor, and host components.
tags:
    - ggml
    - virtgpu
    - configuration
    - environment-variables
    - virglrenderer
    - backend-integration
category: reference
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# GGML-VirtGPU Backend Configuration

Environment variables for the ggml-virtgpu backend across three components: **Frontend (Guest)**, **Hypervisor (Virglrenderer/APIR)**, and **Backend (Host)**.

> [!info]
> Hypervisor and host variables below are transitional—they will be replaced by hypervisor-side APIR Configuration Keys in the future.

## Frontend (Guest)

### `GGML_REMOTING_USE_APIR_CAPSET`

| Property | Value |
|----------|-------|
| Location | `ggml/src/ggml-virtgpu/virtgpu.cpp` |
| Type | Boolean (presence-based) |
| Default | Unset (Venus capset) |

Controls which virtio-gpu capability set to use: **set** → APIR capset (long-term), **unset** → Venus capset (easier testing).

```bash
export GGML_REMOTING_USE_APIR_CAPSET=1  # APIR capset
# or leave unset for Venus capset
```

## Hypervisor (Virglrenderer/APIR)

### `VIRGL_APIR_BACKEND_LIBRARY`

| Property | Value |
|----------|-------|
| Location | `virglrenderer/src/apir/apir-context.c` |
| Config Key | `apir.load_library.path` |
| Type | File path |
| Required | Yes |

Path to the APIR backend library for virglrenderer to dynamically load.

```bash
export VIRGL_APIR_BACKEND_LIBRARY="/path/to/libggml-remotingbackend.so"
```

### `VIRGL_ROUTE_VENUS_TO_APIR`

| Property | Value |
|----------|-------|
| Location | `virglrenderer/src/apir/apir-renderer.h` |
| Type | Boolean (presence-based) |
| Status | **Temporary** — will be removed when hypervisors support APIR natively |

Routes Venus capset calls to APIR for unmodified hypervisors.

> [!warning]
> Breaks normal Vulkan/Venus functionality.

```bash
export VIRGL_ROUTE_VENUS_TO_APIR=1  # Testing only
```

### `VIRGL_APIR_LOG_TO_FILE`

| Property | Value |
|----------|-------|
| Location | `virglrenderer/src/apir/apir-renderer.c` |
| Type | File path |
| Default | stderr |

Enable debug logging from VirglRenderer APIR to a file.

```bash
export VIRGL_APIR_LOG_TO_FILE="/tmp/apir-debug.log"
```

## Backend (Host)

### `APIR_LLAMA_CPP_GGML_LIBRARY_PATH`

| Property | Value |
|----------|-------|
| Location | `ggml/src/ggml-virtgpu/backend/backend.cpp` |
| Config Key | `ggml.library.path` |
| Type | File path |
| Required | **Yes** — fails without it |

Path to the actual GGML backend library (Metal, CUDA, Vulkan, etc.).

```bash
# macOS (Metal)
export APIR_LLAMA_CPP_GGML_LIBRARY_PATH="/opt/llama.cpp/lib/libggml-metal.dylib"

# Linux (CUDA)
export APIR_LLAMA_CPP_GGML_LIBRARY_PATH="/opt/llama.cpp/lib/libggml-cuda.so"

# Vulkan
export APIR_LLAMA_CPP_GGML_LIBRARY_PATH="/opt/llama.cpp/lib/libggml-vulkan.so"
```

### `APIR_LLAMA_CPP_GGML_LIBRARY_REG`

| Property | Value |
|----------|-------|
| Location | `ggml/src/ggml-virtgpu/backend/backend.cpp` |
| Config Key | `ggml.library.reg` |
| Type | Function symbol name |
| Default | `ggml_backend_init` |

Backend registration function to call after loading the library.

```bash
export APIR_LLAMA_CPP_GGML_LIBRARY_REG="ggml_backend_metal_reg"   # Metal
export APIR_LLAMA_CPP_GGML_LIBRARY_REG="ggml_backend_cuda_reg"    # CUDA
export APIR_LLAMA_CPP_GGML_LIBRARY_REG="ggml_backend_vulkan_reg"  # Vulkan
```

### `APIR_LLAMA_CPP_LOG_TO_FILE`

| Property | Value |
|----------|-------|
| Location | `ggml/src/ggml-virtgpu/backend/backend.cpp:62` |
| Type | File path |
| Default | stderr |

Enable debug logging from the GGML backend to a file.

```bash
export APIR_LLAMA_CPP_LOG_TO_FILE="/tmp/ggml-backend-debug.log"
```

## Configuration Flow

1. **Hypervisor setup**: Virglrenderer loads the APIR backend library via `VIRGL_APIR_BACKEND_LIBRARY`.
2. **Context creation**: APIR context populates a config table:
   - `apir.load_library.path` ← `VIRGL_APIR_BACKEND_LIBRARY`
   - `ggml.library.path` ← `APIR_LLAMA_CPP_GGML_LIBRARY_PATH`
   - `ggml.library.reg` ← `APIR_LLAMA_CPP_GGML_LIBRARY_REG`
   - *(future: hypervisor will configure via command-line args instead of env vars)*
3. **Backend init**: Queries config via `virgl_cbs->get_config(ctx_id, key)`.
4. **Library loading**: Dynamically loads and initializes the specified GGML library.

## Error Messages

| Scenario | Message |
|----------|---------|
| Missing library path | `"cannot open the GGML library: env var 'APIR_LLAMA_CPP_GGML_LIBRARY_PATH' not defined"` |
| Missing registration function | `"cannot register the GGML library: env var 'APIR_LLAMA_CPP_GGML_LIBRARY_REG' not defined"` |

## Complete Example (macOS + Metal)

```bash
# Hypervisor
export VIRGL_APIR_BACKEND_LIBRARY="/opt/llama.cpp/lib/libggml-virtgpu-backend.dylib"

# Backend
export APIR_LLAMA_CPP_GGML_LIBRARY_PATH="/opt/llama.cpp/lib/libggml-metal.dylib"
export APIR_LLAMA_CPP_GGML_LIBRARY_REG="ggml_backend_metal_reg"

# Optional logging
export VIRGL_APIR_LOG_TO_FILE="/tmp/apir.log"
export APIR_LLAMA_CPP_LOG_TO_FILE="/tmp/ggml.log"

# Guest
export GGML_REMOTING_USE_APIR_CAPSET=1
```

#virtgpu #configuration #environment-variables #virglrenderer #backend-integration
