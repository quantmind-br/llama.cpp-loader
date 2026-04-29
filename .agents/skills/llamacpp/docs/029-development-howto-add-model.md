---
title: HOWTO add model
url: https://github.com/ggml-org/llama.cpp/blob/master/docs/development/HOWTO-add-model.md
source: git
fetched_at: 2026-04-28T09:49:22.86636652-03:00
rendered_js: false
word_count: 750
summary: This document provides a comprehensive guide for developers on how to implement support for new machine learning model architectures within the llama.cpp framework.
tags:
    - llama-cpp
    - gguf
    - model-conversion
    - ggml
    - machine-learning
    - inference
    - multimodal
category: guide
optimized: true
optimized_at: '2026-04-28T12:00:00Z'
---
# Add a new model architecture to `llama.cpp`

Adding a model requires these steps:

1. Convert the model to GGUF
2. Define the model architecture in `llama.cpp`
3. Build the GGML graph implementation
4. Optional: Add multimodal encoder implementation

After completion, verify that examples and main backends (CUDA, METAL, CPU) work, especially:
- [cli](/tools/cli/)
- [completion](/tools/completion/)
- [imatrix](/tools/imatrix/)
- [quantize](/tools/quantize/)
- [server](/tools/server/)

## 1. Convert the model to GGUF

Done in Python using the [gguf](https://pypi.org/project/gguf/) library. Use [convert_hf_to_gguf.py](/convert_hf_to_gguf.py) or [examples/convert_legacy_llama.py](/examples/convert_legacy_llama.py) (for `llama/llama2` models in `.pth` format). The script reads model config, tokenizer, tensor names+data and converts them to GGUF metadata and tensors.

Steps for an HF model:

1. Define the model `ModelBase.register` annotation in a new `TextModel` or `MmprojModel` subclass:

```python
@ModelBase.register("MyModelForCausalLM")
class MyModel(TextModel):
    model_arch = gguf.MODEL_ARCH.MYMODEL
```

or

```python
@ModelBase.register("MyModelForConditionalGeneration")
class MyModel(MmprojModel):
    model_arch = gguf.MODEL_ARCH.MYMODEL
```

2. Define GGUF tensor layout in [constants.py](/gguf-py/gguf/constants.py): add an enum entry in `MODEL_ARCH`, the human-friendly name in `MODEL_ARCH_NAMES`, and tensor names in `MODEL_TENSORS`.

Example for `falcon` model:
```python
    MODEL_ARCH.FALCON: [
        MODEL_TENSOR.TOKEN_EMBD,
        MODEL_TENSOR.OUTPUT_NORM,
        MODEL_TENSOR.OUTPUT,
        MODEL_TENSOR.ATTN_NORM,
        MODEL_TENSOR.ATTN_NORM_2,
        MODEL_TENSOR.ATTN_QKV,
        MODEL_TENSOR.ATTN_OUT,
        MODEL_TENSOR.FFN_DOWN,
        MODEL_TENSOR.FFN_UP,
    ]
```

3. Map original tensor names to GGUF equivalents in [tensor_mapping.py](/gguf-py/gguf/tensor_mapping.py). Verify the equivalent naming doesn't already exist before adding new names. The keyword `bid` substitutes repetitive layer/block indices.

Example for attention normalization:
```python
block_mappings_cfg: dict[MODEL_TENSOR, tuple[str, ...]] = {
        # Attention norm
        MODEL_TENSOR.ATTN_NORM: (
            "gpt_neox.layers.{bid}.input_layernorm",                # gptneox
            "transformer.h.{bid}.ln_1",                             # gpt2 gpt-j refact qwen
            "transformer.blocks.{bid}.norm_1",                      # mpt
            ...
        )
}
```

`transformer.blocks.{bid}.norm_1` maps to `blk.{bid}.attn_norm` in GGUF.

Depending on model configuration, you may need to override:
- `TextModel#set_gguf_parameters`
- `MmprojModel#set_gguf_parameters`
- `ModelBase#set_vocab`
- `ModelBase#modify_tensors`

> [!warning]
> Tensor names must end with `.weight` or `.bias` suffixes — tools like `quantize` expect this.

## 2. Define the model architecture in `llama.cpp`

Define params and tensor layout in source files:

1. New `llm_arch` enum value in `src/llama-arch.h`.
2. In `src/llama-arch.cpp`:
    - Architecture name in `LLM_ARCH_NAMES` map.
    - Tensor list in `llm_get_tensor_names` (may also need `LLM_TENSOR_NAMES` updates).
3. Non-standard metadata loading in `llama_model_loader` constructor in `src/llama-model-loader.cpp`.
4. If the model uses RoPE, add a case in `llama_model_rope_type` function in `src/llama-model.cpp`.

> [!info]
> Dimensions in `ggml` are typically in reverse order of `pytorch` dimensions.

## 3. Build the GGML graph implementation

Provide the inference graph implementation in `src/llama-model.cpp`:

1. Create a new struct inheriting from `llm_graph_context` with graph-building logic in its constructor. Reference existing implementations: `llm_build_llama`, `llm_build_dbrx`, `llm_build_bert`.
2. In `llama_model::build_graph`, add a case to instantiate your new struct.

Some `ggml` backends don't support all operations — backend implementations can be added in a separate PR.

To debug the inference graph: use [llama-eval-callback](/examples/eval-callback/).

## 4. Optional: Add multimodal encoder implementation

For multimodal inputs, add a new encoder definition in `libmtmd`. See [[038-multimodal|the docs]] and `tools/mtmd` source directory.

1. In the conversion script, add a subclass extending `MmprojModel` or equivalent.
2. Add encoder definition in `clip.cpp`.
3. Implement preprocessor in `mtmd.cpp` (reuse existing preprocessor when possible).
4. Implement the encoder GGML graph — either in a dedicated file or by reusing existing implementations (siglip, pixtral, qwen) with a model-specific projector.

> [!tip]
> - Read existing encoder definitions in `tools/mtmd/models` before adding new ones. Extend existing models rather than duplicating code.
> - Debug with [llama-mtmd-debug](tools/mtmd/debug/mtmd-debug.cpp).
> - Adding model-specific API or CLI is an anti-pattern in `libmtmd`. The library provides a model-agnostic multimodal pipeline.
> - In most cases, `llama-mtmd-cli` should not be modified. Model-specific prompts should be provided by the user or baked into the Jinja chat template.

## Tips and tricks

### Working with ggml_rope_ext

PyTorch typically computes `freq_cis`/`sin`/`cos` explicitly. In llama.cpp, most RoPE operations use `ggml_rope_ext` which doesn't require a sin/cos matrix — saving memory and enabling kernel fusion with other ops.

Since `ggml_rope_ext` only provides a subset of RoPE implementations, converting from PyTorch may require creative adaptations. See in-code docs in `ggml.h`.

Examples:
- `libmtmd` implements 2D RoPE with `GGML_ROPE_TYPE_NORMAL` by splitting input, applying `ggml_rope_ext` to each half, then joining via `ggml_concat`.
- [Kimi-K2.5](https://github.com/ggml-org/llama.cpp/pull/19170) vision encoder uses interleaved-frequency vision RoPE — weights must be permuted during conversion to reuse `build_rope_2d()`.
- [Gemma 4](https://github.com/ggml-org/llama.cpp/pull/21309) uses "proportional" RoPE — `rope_freqs` is set to a very large value in last dimensions to prevent rotation. See `Gemma4Model` in `convert_hf_to_gguf.py`.
- Some models require position scaling (e.g. `[0, 1, 2, ...]` → `[0, 0.5, 1, ...]`) — use `freq_scale = 0.5f`.
- Some models use learned RoPE frequencies instead of `powf(freq_base, -2.0 * i / n_dims)` — provide via `rope_freqs` tensor (`c` argument in `ggml_rope_ext`), set `freq_base = 1.0f`. Note: `rope_freqs` in GGML is the **inverse** (`theta = pos[i] / rope_freqs`), so invert during conversion.

## GGUF Specification

https://github.com/ggml-org/ggml/blob/master/docs/gguf.md

## Resources

- YaRN RoPE scaling: https://github.com/ggml-org/llama.cpp/pull/2268
- Baichuan serial models: https://github.com/ggml-org/llama.cpp/pull/3009
- Attention bias support: https://github.com/ggml-org/llama.cpp/pull/4283
- Mixtral support: https://github.com/ggml-org/llama.cpp/pull/4406
- BERT embeddings: https://github.com/ggml-org/llama.cpp/pull/5423
- Grok-1 support: https://github.com/ggml-org/llama.cpp/pull/6204
- Command R Plus support: https://github.com/ggml-org/llama.cpp/pull/6491
- DBRX arch support: https://github.com/ggml-org/llama.cpp/pull/6515
- HuggingFace to GGUF conversion: https://github.com/ggml-org/llama.cpp/discussions/2948
