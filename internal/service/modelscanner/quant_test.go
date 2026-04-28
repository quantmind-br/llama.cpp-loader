package modelscanner

import "testing"

func TestParseQuant(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"Qwen2.5-Coder-32B-Instruct-Q5_K_M.gguf", "Q5_K_M"},
		{"llama-3.1-8b-instruct-q4_k_m.gguf", "Q4_K_M"},
		{"Mistral-7B-Instruct-v0.3-Q8_0.gguf", "Q8_0"},
		{"some-model-IQ4_NL.gguf", "IQ4_NL"},
		{"deepseek-coder-6.7B-instruct-Q4_0.gguf", "Q4_0"},
		{"f16-only.gguf", "F16"},
		{"plain-name.gguf", ""},
		{"model-q5_k_s.gguf", "Q5_K_S"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseQuant(tc.name)
			if got != tc.want {
				t.Fatalf("parseQuant(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestParseParams(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"Qwen2.5-Coder-32B-Instruct-Q5_K_M.gguf", "32B"},
		{"llama-3.1-8b-instruct-q4_k_m.gguf", "8B"},
		{"deepseek-coder-6.7B-instruct-Q4_0.gguf", "6.7B"},
		{"Mixtral-8x7B-Instruct-v0.1-Q4_K_M.gguf", "8x7B"},
		{"phi-3.5-mini-3.8B-instruct-Q4_K_M.gguf", "3.8B"},
		{"plain-name.gguf", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseParams(tc.name)
			if got != tc.want {
				t.Fatalf("parseParams(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}
