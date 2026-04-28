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
