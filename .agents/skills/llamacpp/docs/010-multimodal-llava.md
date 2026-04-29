---
title: Llava
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal/llava.md
source: git
fetched_at: 2026-04-28T09:49:36.528511741-03:00
rendered_js: false
word_count: 276
summary: This document provides instructions on how to convert, configure, and run LLaVA multimodal models within a GGUF-compatible environment.
tags:
    - llava
    - gguf
    - model-conversion
    - multimodal
    - llm
    - llama-cpp
    - image-encoder
category: guide
optimized: true
optimized_at: "2026-04-28T12:00:00Z"
---
# LLaVA

Supports [llava-v1.5](https://huggingface.co/liuhaotian/llava-v1.5-7b) and llava-1.6 [llava-v1.6](https://huggingface.co/collections/liuhaotian/llava-16-65b9e40155f60fd046a5ccf2) variants.

Pre-converted models: [7b](https://huggingface.co/mys/ggml_llava-v1.5-7b), [13b](https://huggingface.co/mys/ggml_llava-v1.5-13b). For llava-1.6: [7b-34b](https://huggingface.co/cmp-nct/llava-1.6-gguf).

## Usage

Build the `llama-mtmd-cli` binary, then run:

```sh
./llama-mtmd-cli -m ../llava-v1.5-7b/ggml-model-f16.gguf \
    --mmproj ../llava-v1.5-7b/mmproj-model-f16.gguf \
    --chat-template vicuna
```

> [!tip] Use a lower temperature (e.g., `--temp 0.1`) for better quality. For GPU offloading, use the `-ngl` flag.

## LLaVA 1.5

1. Clone a LLaVA and a CLIP model ([available options](https://github.com/haotian-liu/LLaVA/blob/main/docs/MODEL_ZOO.md)):

```sh
git clone https://huggingface.co/liuhaotian/llava-v1.5-7b

git clone https://huggingface.co/openai/clip-vit-large-patch14-336
```

2. Install required Python packages:

```sh
pip install -r tools/mtmd/requirements.txt
```

3. Split the LLaVA model with `llava_surgery.py`:

```sh
python ./tools/mtmd/llava_surgery.py -m ../llava-v1.5-7b
```

4. Convert the image encoder to GGUF:

```sh
python ./tools/mtmd/convert_image_encoder_to_gguf.py -m ../clip-vit-large-patch14-336 --llava-projector ../llava-v1.5-7b/llava.projector --output-dir ../llava-v1.5-7b
```

5. Convert the LLaMA part to GGUF:

```sh
python ./examples/convert_legacy_llama.py ../llava-v1.5-7b --skip-unknown
```

Both the LLaMA part and image encoder will be in the `llava-v1.5-7b` directory.

## LLaVA 1.6 GGUF Conversion

1. Clone a LLaVA 1.6 model:

```console
git clone https://huggingface.co/liuhaotian/llava-v1.6-vicuna-7b
```

2. Install required Python packages:

```sh
pip install -r tools/mtmd/requirements.txt
```

3. Use `llava_surgery_v2.py` (supports both pytorch and safetensor models):

```console
python tools/mtmd/llava_surgery_v2.py -C -m ../llava-v1.6-vicuna-7b/
```

Produces `llava.projector` and `llava.clip` in the model directory.

4. Set up the visual encoder directory:

```console
mkdir vit
cp ../llava-v1.6-vicuna-7b/llava.clip vit/pytorch_model.bin
cp ../llava-v1.6-vicuna-7b/llava.projector vit/
curl -s -q https://huggingface.co/cmp-nct/llava-1.6-gguf/raw/main/config_vit.json -o vit/config.json
```

5. Create the visual GGUF model (uses `--clip-model-is-vision` for the pure vision part):

```console
python ./tools/mtmd/convert_image_encoder_to_gguf.py -m vit --llava-projector vit/llava.projector --output-dir vit --clip-model-is-vision
```

6. Convert the model to GGUF:

```console
python ./examples/convert_legacy_llama.py ../llava-v1.6-vicuna-7b/ --skip-unknown
```

7. Run the CLI:

```console
./llama-mtmd-cli -m ../llava-v1.6-vicuna-7b/ggml-model-f16.gguf --mmproj vit/mmproj-model-f16.gguf
```

> [!note] llava-1.6 needs at least 3000 context tokens (use `-c 4096`). It greatly benefits from batched prompt processing (defaults work).

> [!tip] If the language model in step 6 is incompatible with the legacy conversion script, load the model in `transformers` and export only the LLM:

```python
import os
import transformers

model_path = ...
llm_export_path = ...

tokenizer = transformers.AutoTokenizer.from_pretrained(model_path)
model = transformers.AutoModelForImageTextToText.from_pretrained(model_path)

tokenizer.save_pretrained(llm_export_path)
model.language_model.save_pretrained(llm_export_path)
```

Then convert using `convert_hf_to_gguf.py` which handles more LLM architectures.

## Chat Template

Both llava-1.5 and llava-1.6 require the `vicuna` chat template. Add `--chat-template vicuna` to activate.

## Detecting LLaVA Version

Check the image embedding token count at runtime:

- **Llava-1.5**: `encode_image_with_clip: image embedding created: 576 tokens`
- **Llava-1.6** (anything above 576): `encode_image_with_clip: image embedding created: 2880 tokens`

Prompt token counts will also show 1000+ tokens for llava-1.6.

#llava #multimodal #gguf #model-conversion #image-encoder
