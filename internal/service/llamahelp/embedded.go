package llamahelp

import "github.com/quantmind-br/llama-cpp-loader/internal/domain"

// EmbeddedSchema returns the compile-time fallback schema covering the curated
// essentials. Used when llama-server --help cannot be invoked. Version label
// reflects the upstream build the embedded data was pinned against.
func EmbeddedSchema() domain.FlagSchema {
	flags := map[string]domain.FlagSpec{
		"model": {
			Long:     "model",
			Short:    "m",
			Type:     domain.FlagTypeString,
			HelpText: "model path (.gguf)",
			Group:    "embedded",
		},
		"n-gpu-layers": {
			Long:     "n-gpu-layers",
			Short:    "ngl",
			Aliases:  []string{"gpu-layers"},
			Type:     domain.FlagTypeInt,
			Default:  -1,
			HelpText: "max number of layers to store in VRAM",
			Group:    "embedded",
		},
		"ctx-size": {
			Long:     "ctx-size",
			Short:    "c",
			Type:     domain.FlagTypeInt,
			Default:  4096,
			HelpText: "size of the prompt context",
			Group:    "embedded",
		},
		"batch-size": {
			Long:     "batch-size",
			Short:    "b",
			Type:     domain.FlagTypeInt,
			Default:  2048,
			HelpText: "logical maximum batch size",
			Group:    "embedded",
		},
		"ubatch-size": {
			Long:     "ubatch-size",
			Short:    "ub",
			Type:     domain.FlagTypeInt,
			Default:  512,
			HelpText: "physical maximum batch size",
			Group:    "embedded",
		},
		"flash-attn": {
			Long:       "flash-attn",
			Short:      "fa",
			Type:       domain.FlagTypeEnum,
			EnumValues: []string{"on", "off", "auto"},
			Default:    "auto",
			HelpText:   "Flash Attention mode",
			Group:      "embedded",
		},
		"threads": {
			Long:     "threads",
			Short:    "t",
			Type:     domain.FlagTypeInt,
			Default:  -1,
			HelpText: "CPU threads",
			Group:    "embedded",
		},
		"parallel": {
			Long:     "parallel",
			Short:    "np",
			Type:     domain.FlagTypeInt,
			Default:  1,
			HelpText: "number of parallel sequences to decode",
			Group:    "embedded",
		},
		"mlock": {
			Long:     "mlock",
			Type:     domain.FlagTypeBool,
			HelpText: "lock model in RAM (no swap)",
			Group:    "embedded",
		},
		"cache-type-k": {
			Long:       "cache-type-k",
			Short:      "ctk",
			Type:       domain.FlagTypeEnum,
			EnumValues: cacheTypeEnum,
			HelpText:   "KV cache data type for K",
			Group:      "embedded",
		},
		"cache-type-v": {
			Long:       "cache-type-v",
			Short:      "ctv",
			Type:       domain.FlagTypeEnum,
			EnumValues: cacheTypeEnum,
			HelpText:   "KV cache data type for V",
			Group:      "embedded",
		},
		"split-mode": {
			Long:       "split-mode",
			Short:      "sm",
			Type:       domain.FlagTypeEnum,
			EnumValues: []string{"none", "layer", "row"},
			HelpText:   "how to split model across multiple GPUs",
			Group:      "embedded",
		},
		"tensor-split": {
			Long:     "tensor-split",
			Short:    "ts",
			Type:     domain.FlagTypeString,
			HelpText: "fraction of model offloaded to each GPU (comma-separated)",
			Group:    "embedded",
		},
	}
	return domain.FlagSchema{
		Version: "embedded-v7376",
		Flags:   flags,
	}
}
