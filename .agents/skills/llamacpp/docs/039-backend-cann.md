---
title: CANN
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/CANN.md
source: git
fetched_at: 2026-04-28T09:49:08.915765911-03:00
rendered_js: false
word_count: 1298
summary: This document provides an overview of the CANN backend for llama.cpp, detailing hardware support for Ascend NPUs, compatible AI models, and setup requirements for running the environment in Linux or Docker.
tags:
    - llama-cpp
    - ascend-npu
    - cann
    - ai-inference
    - hardware-acceleration
    - docker
category: reference
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# llama.cpp for CANN

The CANN backend enables llama.cpp inference on Huawei **Ascend NPU** hardware via the CANN (Compute Architecture for Neural Networks) toolkit, using AscendC and ACLNN APIs.

## News

| Date | Update |
|------|--------|
| 2024.11 | F16 and F32 support for Ascend 310P |
| 2024.8 | `Q4_0` and `Q8_0` support for Ascend NPU |
| 2024.7 | Initial CANN backend created |

## OS Support

| OS | Status | Verified |
|:--:|:------:|:--------:|
| Linux | Support | Ubuntu 22.04, OpenEuler22.03 |

## Hardware

Retrieve Ascend device IDs:

```sh
lspci -n | grep -Eo '19e5:d[0-9a-f]{3}' | cut -d: -f2
```

### Supported Devices

| Device ID | Product Series | Product Models | Chip Model | Verified |
|:---------:|----------------|----------------|:----------:|:--------:|
| d803 | Atlas A3 Train | | 910C | |
| d803 | Atlas A3 Infer | | 910C | |
| d802 | Atlas A2 Train | | 910B | |
| d802 | Atlas A2 Infer | Atlas 300I A2 | 910B | Support |
| d801 | Atlas Train | | 910 | |
| d500 | Atlas Infer | Atlas 300I Duo | 310P | Support |

> [!info]
> - File issues with **[CANN]** prefix for Ascend NPU problems.
> - If your device works, please update the table above.

## Model Supports

<details>
<summary>Text-only</summary>

| Model Name | FP16 | Q4_0 | Q8_0 |
|:-----------|:----:|:----:|:----:|
| Llama-2 | √ | √ | √ |
| Llama-3 | √ | √ | √ |
| Mistral-7B | √ | √ | √ |
| Mistral MOE | √ | √ | √ |
| DBRX | - | - | - |
| Falcon | √ | √ | √ |
| Chinese LLaMA/Alpaca | √ | √ | √ |
| Vigogne(French) | √ | √ | √ |
| BERT | x | x | x |
| Koala | √ | √ | √ |
| Baichuan | √ | √ | √ |
| Aquila 1 & 2 | √ | √ | √ |
| Starcoder models | √ | √ | √ |
| Refact | √ | √ | √ |
| MPT | √ | √ | √ |
| Bloom | √ | √ | √ |
| Yi models | √ | √ | √ |
| stablelm models | √ | √ | √ |
| DeepSeek models | x | x | x |
| Qwen models | √ | √ | √ |
| PLaMo-13B | √ | √ | √ |
| Phi models | √ | √ | √ |
| PhiMoE | √ | √ | √ |
| GPT-2 | √ | √ | √ |
| Orion | √ | √ | √ |
| InternlLM2 | √ | √ | √ |
| CodeShell | √ | √ | √ |
| Gemma | √ | √ | √ |
| Mamba | √ | √ | √ |
| Xverse | √ | √ | √ |
| command-r models | √ | √ | √ |
| Grok-1 | - | - | - |
| SEA-LION | √ | √ | √ |
| GritLM-7B | √ | √ | √ |
| OLMo | √ | √ | √ |
| OLMo 2 | √ | √ | √ |
| OLMoE | √ | √ | √ |
| Granite models | √ | √ | √ |
| GPT-NeoX | √ | √ | √ |
| Pythia | √ | √ | √ |
| Snowflake-Arctic MoE | - | - | - |
| Smaug | √ | √ | √ |
| Poro 34B | √ | √ | √ |
| Bitnet b1.58 models | √ | x | x |
| Flan-T5 | √ | √ | √ |
| Open Elm models | x | √ | √ |
| chatGLM3-6B + ChatGLM4-9b + GLMEdge-1.5b + GLMEdge-4b | √ | √ | √ |
| GLM-4-0414 | √ | √ | √ |
| SmolLM | √ | √ | √ |
| EXAONE-3.0-7.8B-Instruct | √ | √ | √ |
| FalconMamba Models | √ | √ | √ |
| Jais Models | - | x | x |
| Bielik-11B-v2.3 | √ | √ | √ |
| RWKV-6 | - | √ | √ |
| QRWKV-6 | √ | √ | √ |
| GigaChat-20B-A3B | x | x | x |
| Trillion-7B-preview | √ | √ | √ |
| Ling models | √ | √ | √ |

</details>

<details>
<summary>Multimodal</summary>

| Model Name | FP16 | Q4_0 | Q8_0 |
|:-----------|:----:|:----:|:----:|
| LLaVA 1.5 / 1.6 models | x | x | x |
| BakLLaVA | √ | √ | √ |
| Obsidian | √ | - | - |
| ShareGPT4V | x | - | - |
| MobileVLM 1.7B/3B | - | - | - |
| Yi-VL | - | - | - |
| Mini CPM | √ | √ | √ |
| Moondream | √ | √ | √ |
| Bunny | √ | - | - |
| GLM-EDGE | √ | √ | √ |
| Qwen2-VL | √ | √ | √ |

</details>

## DataType Supports

| DataType | 910B | 310P |
|:--------:|:----:|:----:|
| FP16 | Support | Support |
| Q8_0 | Support | Partial |
| Q4_0 | Support | Partial |
| BF16 | Support | |

> [!warning]
> **310P limitations:**
> - `Q8_0`: data transform/buffer path + `GET_ROWS` supported, but quantized `MUL_MAT`/`MUL_MAT_ID` not supported.
> - `Q4_0`: data transform/buffer path implemented, but quantized `MUL_MAT`/`MUL_MAT_ID` not supported.

## Docker

### Build Image

```sh
docker build -t llama-cpp-cann -f .devops/llama-cli-cann.Dockerfile .
```

### Run Container

```sh
npu-smi info  # List cards

docker run --name llamacpp \
  --device /dev/davinci0 \
  --device /dev/davinci_manager \
  --device /dev/devmm_svm \
  --device /dev/hisi_hdc \
  -v /usr/local/dcmi:/usr/local/dcmi \
  -v /usr/local/bin/npu-smi:/usr/local/bin/npu-smi \
  -v /usr/local/Ascend/driver/lib64/:/usr/local/Ascend/driver/lib64/ \
  -v /usr/local/Ascend/driver/version.info:/usr/local/Ascend/driver/version.info \
  -v /PATH_TO_YOUR_MODELS/:/app/models \
  -it llama-cpp-cann \
  -m /app/models/MODEL_PATH \
  -ngl 32 \
  -p "Building a website can be done in 10 simple steps:"
```

> [!note]
> Ascend Driver and firmware must be installed on the **host** machine (see [Linux setup](#linux)).

## Linux Setup

### I. Setup Environment

1. **Configure Ascend user/group:**

    ```sh
    sudo groupadd HwHiAiUser
    sudo useradd -g HwHiAiUser -d /home/HwHiAiUser -m HwHiAiUser -s /bin/bash
    sudo usermod -aG HwHiAiUser $USER
    ```

2. **Install dependencies:**

    **Ubuntu/Debian:**
    ```sh
    sudo apt-get update
    sudo apt-get install -y gcc python3 python3-pip linux-headers-$(uname -r)
    ```

    **RHEL/CentOS:**
    ```sh
    sudo yum makecache
    sudo yum install -y gcc python3 python3-pip kernel-headers-$(uname -r) kernel-devel-$(uname -r)
    ```

3. **Install CANN (driver + toolkit):**

    `$ARCH` = `x86_64` or `aarch64`. `$CHIP` = `910b` or `310p`.

    ```sh
    wget https://ascend-repo.obs.cn-east-2.myhuaweicloud.com/CANN/CANN%208.5.T63/Ascend-cann_8.5.0_linux-$ARCH.run
    sudo bash ./Ascend-cann_8.5.0_linux-$ARCH.run --install

    wget https://ascend-repo.obs.cn-east-2.myhuaweicloud.com/CANN/CANN%208.5.T63/Ascend-cann-$CHIP-ops_8.5.0_linux-$ARCH.run
    sudo bash ./Ascend-cann-$CHIP-ops_8.5.0_linux-$ARCH.run --install
    ```

4. **Verify installation:**

    ```sh
    npu-smi info  # Should display device info

    source /usr/local/Ascend/cann/set_env.sh
    python3 -c "import acl; print(acl.get_soc_name())"  # Should print chip model
    ```

### II. Build llama.cpp

```sh
cmake -B build -DGGML_CANN=on -DCMAKE_BUILD_TYPE=release
cmake --build build --config release
```

### III. Run Inference

1. **Prepare model:** Refer to [*Obtaining and quantizing models*](../../README.md#obtaining-and-quantizing-models). CANN backend only supports FP16/Q4_0/Q8_0.

2. **Launch inference** with device selection:

    | Mode | Parameter |
    |------|-----------|
    | Single device | `--split-mode none --main-gpu DEVICE_ID` |
    | Multiple devices | `--split-mode layer` (default) |

    ```sh
    # Single device (device 0)
    ./build/bin/llama-cli -m path_to_model -p "Building a website can be done in 10 simple steps:" -n 400 -e -ngl 33 -sm none -mg 0

    # Multiple devices
    ./build/bin/llama-cli -m path_to_model -p "Building a website can be done in 10 simple steps:" -n 400 -e -ngl 33 -sm layer
    ```

> [!tip]
> Add the **[CANN]** prefix/tag in issues/PRs to help the CANN team respond promptly.

## Flash Attention

Basic FA kernel added in `aclnn_ops.cpp` via aclnnops. Currently supports **FP16 KV tensors only**, no logit softcap. Quantized version planned (aclnn FA interface doesn't support logit softcap).

Authors (Peking University): Bizhao Shi (bshi@pku.edu.cn), Yuxin Yang (xyxang@pku.edu.cn), Ruiyang Ma (ruiyang@stu.pku.edu.cn), Guojie Luo (gluo@pku.edu.cn). Thanks to Tuo Dai, Shanni Li, and Huawei project maintainers.

## Environment Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `GGML_CANN_MEM_POOL` | Enum | `vmm` | Memory pool strategy: `vmm` (virtual memory, falls back to `leg`), `prio` (priority queue), `leg` (fixed-size buffer) |
| `GGML_CANN_DISABLE_BUF_POOL_CLEAN` | Boolean | — | Disable automatic memory pool cleanup. Only effective with `prio` or `leg`. |
| `GGML_CANN_WEIGHT_NZ` | Boolean | Enabled | Convert matmul weight format ND→NZ for performance. |
| `GGML_CANN_ACL_GRAPH` | Boolean | Enabled | Use ACL graph execution instead of op-by-op. Requires `USE_ACL_GRAPH=ON` at compile time. |
| `GGML_CANN_GRAPH_CACHE_CAPACITY` | Integer | 12 | Max compiled CANN graphs in LRU cache. |
| `GGML_CANN_PREFILL_USE_GRAPH` | Boolean | false | Enable ACL graph during prefill. Requires FA enabled. |
| `GGML_CANN_OPERATOR_FUSION` | Boolean | false | Fuse compatible operators (e.g., ADD + RMS_NORM) to reduce overhead. |

To enable ACL graph at compile time:

```sh
cmake -B build -DGGML_CANN=on -DCMAKE_BUILD_TYPE=release -DUSE_ACL_GRAPH=ON
cmake --build build --config release
```

#cann #ascend-npu #hardware-acceleration #docker #ai-inference
