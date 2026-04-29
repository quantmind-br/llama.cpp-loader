---
title: Granitevision
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal/granitevision.md
source: git
fetched_at: 2026-04-28T09:49:33.968368172-03:00
rendered_js: false
word_count: 190
summary: This guide details the process of converting Granite Vision models into the GGUF format for use with llama.cpp, including visual encoder extraction, LLM export, and model quantization.
tags:
    - granite-vision
    - gguf
    - llama-cpp
    - model-conversion
    - quantization
    - multimodal
    - siglip
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# Granite Vision

Download the model and set the `GRANITE_MODEL` env var:

```bash
$ git clone https://huggingface.co/ibm-granite/granite-vision-3.2-2b
$ export GRANITE_MODEL=./granite-vision-3.2-2b
```

## 1. Run llava surgery v2

Split the projector and visual encoder into separate files:

```bash
$ python llava_surgery_v2.py -C -m $GRANITE_MODEL
```

Produces `llava.clip` and `llava.projector` in the model directory:

```bash
$ ls $GRANITE_MODEL | grep -i llava
llava.clip
llava.projector
```

Verify files are non-empty:

```python
import os
import torch

MODEL_PATH = os.getenv("GRANITE_MODEL")
if not MODEL_PATH:
    raise ValueError("env var GRANITE_MODEL is unset!")

encoder_tensors = torch.load(os.path.join(MODEL_PATH, "llava.clip"))
projector_tensors = torch.load(os.path.join(MODEL_PATH, "llava.projector"))

assert len(encoder_tensors) > 0
assert len(projector_tensors) > 0
```

Inspecting `.keys()`: `encoder_tensors` contains `vision_model` tensors; `projector_tensors` contains 5 tensors (`multi_modal_projector.linear_1.bias`, `multi_modal_projector.linear_1.weight`, `multi_modal_projector.linear_2.bias`, `multi_modal_projector.linear_2.weight`, `image_newline`).

## 2. Create the Visual Component GGUF

Create a directory for visual components and copy the llava files:

```bash
$ ENCODER_PATH=$PWD/visual_encoder
$ mkdir $ENCODER_PATH

$ cp $GRANITE_MODEL/llava.clip $ENCODER_PATH/pytorch_model.bin
$ cp $GRANITE_MODEL/llava.projector $ENCODER_PATH/
```

Write a config for the visual encoder. Use the correct `image_grid_pinpoints` from `$GRANITE_MODEL/config.json`:

```json
{
    "_name_or_path": "siglip-model",
    "architectures": [
      "SiglipVisionModel"
    ],
    "image_grid_pinpoints": [
        [384,384],
        [384,768],
        [384,1152],
        [384,1536],
        [384,1920],
        [384,2304],
        [384,2688],
        [384,3072],
        [384,3456],
        [384,3840],
        [768,384],
        [768,768],
        [768,1152],
        [768,1536],
        [768,1920],
        [1152,384],
        [1152,768],
        [1152,1152],
        [1536,384],
        [1536,768],
        [1920,384],
        [1920,768],
        [2304,384],
        [2688,384],
        [3072,384],
        [3456,384],
        [3840,384]
    ],
    "mm_patch_merge_type": "spatial_unpad",
    "hidden_size": 1152,
    "image_size": 384,
    "intermediate_size": 4304,
    "model_type": "siglip_vision_model",
    "num_attention_heads": 16,
    "num_hidden_layers": 27,
    "patch_size": 14,
    "layer_norm_eps": 1e-6,
    "hidden_act": "gelu_pytorch_tanh",
    "projection_dim": 0,
    "vision_feature_layer": [-24, -20, -12, -1]
}
```

Directory should look like:

```bash
$ ls $ENCODER_PATH
config.json             llava.projector         pytorch_model.bin
```

Convert to GGUF. Override image mean/std to `[.5,.5,.5]` (SigLIP encoder; values from `preprocessor_config.json`):

```bash
$ python convert_image_encoder_to_gguf.py \
    -m $ENCODER_PATH \
    --llava-projector $ENCODER_PATH/llava.projector \
    --output-dir $ENCODER_PATH \
    --clip-model-is-vision \
    --clip-model-is-siglip \
    --image-mean 0.5 0.5 0.5 \
    --image-std 0.5 0.5 0.5
```

Output: `$ENCODER_PATH/mmproj-model-f16.gguf`. Reference its absolute path as `$VISUAL_GGUF_PATH`.

## 3. Create the LLM GGUF

Export the LLM from the composite model via `transformers`, then convert normally.

Set the export path:

```bash
$ export LLM_EXPORT_PATH=$PWD/granite_vision_llm
```

Export the LLM:

```python
import os
import transformers

MODEL_PATH = os.getenv("GRANITE_MODEL")
if not MODEL_PATH:
    raise ValueError("env var GRANITE_MODEL is unset!")

LLM_EXPORT_PATH = os.getenv("LLM_EXPORT_PATH")
if not LLM_EXPORT_PATH:
    raise ValueError("env var LLM_EXPORT_PATH is unset!")

tokenizer = transformers.AutoTokenizer.from_pretrained(MODEL_PATH)

# NOTE: granite vision support was added to transformers very recently (4.49);
# if you get size mismatches, your version is too old.
# If you are running with an older version, set `ignore_mismatched_sizes=True`
# as shown below; it won't be loaded correctly, but the LLM part of the model that
# we are exporting will be loaded correctly.
model = transformers.AutoModelForImageTextToText.from_pretrained(MODEL_PATH, ignore_mismatched_sizes=True)

tokenizer.save_pretrained(LLM_EXPORT_PATH)
model.language_model.save_pretrained(LLM_EXPORT_PATH)
```

Convert to GGUF:

```bash
$ LLM_GGUF_PATH=$LLM_EXPORT_PATH/granite_llm.gguf
...
$ python convert_hf_to_gguf.py --outfile $LLM_GGUF_PATH $LLM_EXPORT_PATH
```

## 4. Quantization

Quantize the LLM with `llama-quantize` as any other LLM:

```bash
$ ./build/bin/llama-quantize $LLM_EXPORT_PATH/granite_llm.gguf $LLM_EXPORT_PATH/granite_llm_q4_k_m.gguf Q4_K_M
$ LLM_GGUF_PATH=$LLM_EXPORT_PATH/granite_llm_q4_k_m.gguf
```

> [!warning] You cannot quantize the visual encoder — granite vision uses SigLIP, which has tensor dimensions not divisible by 32.

## 5. Run the Model

Build llama.cpp normally. Use `llama-mtmd-cli` with both model files:

```bash
$ ./build/bin/llama-mtmd-cli -m $LLM_GGUF_PATH \
    --mmproj $VISUAL_GGUF_PATH \
    -c 16384 \
    --temp 0
```

#granite-vision #siglip #gguf #model-conversion #multimodal
