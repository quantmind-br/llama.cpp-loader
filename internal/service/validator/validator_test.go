package validator

import (
	"testing"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

func TestValidator_EmptySchemaProducesNoTypeIssues(t *testing.T) {
	v := New()
	p := domain.Profile{
		ID:    "x",
		Name:  "X",
		Model: "/tmp/nonexistent.gguf",
		Args:  map[string]any{},
	}
	rep := v.Validate(p, domain.FlagSchema{Flags: map[string]domain.FlagSpec{}})
	// At this stage, with no schema and no rules wired, Errors should be empty.
	if len(rep.Errors) != 0 {
		t.Errorf("Errors=%v, want empty", rep.Errors)
	}
}

func TestValidator_TypeRule(t *testing.T) {
	schema := domain.FlagSchema{Flags: map[string]domain.FlagSpec{
		"ctx-size":   {Long: "ctx-size", Type: domain.FlagTypeInt},
		"flash-attn": {Long: "flash-attn", Type: domain.FlagTypeEnum, EnumValues: []string{"on", "off", "auto"}},
		"mlock":      {Long: "mlock", Type: domain.FlagTypeBool},
	}}
	cases := []struct {
		name      string
		args      map[string]any
		wantErrs  int
		wantField string
	}{
		{"int ok as float64", map[string]any{"ctx-size": float64(4096)}, 0, ""},
		{"int ok as int", map[string]any{"ctx-size": 4096}, 0, ""},
		{"int rejects string", map[string]any{"ctx-size": "abc"}, 1, "ctx-size"},
		{"enum ok", map[string]any{"flash-attn": "on"}, 0, ""},
		{"enum rejects unknown", map[string]any{"flash-attn": "maybe"}, 1, "flash-attn"},
		{"bool ok", map[string]any{"mlock": true}, 0, ""},
		{"bool rejects string", map[string]any{"mlock": "yes"}, 1, "mlock"},
		{"unknown flag is ignored", map[string]any{"unheard-of": 1}, 0, ""},
	}
	v := New()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := domain.Profile{ID: "x", Args: tc.args}
			rep := v.Validate(p, schema)
			if got := len(rep.Errors); got != tc.wantErrs {
				t.Fatalf("Errors=%d (%v), want %d", got, rep.Errors, tc.wantErrs)
			}
			if tc.wantErrs > 0 && rep.Errors[0].Field != tc.wantField {
				t.Errorf("Errors[0].Field=%q, want %q", rep.Errors[0].Field, tc.wantField)
			}
		})
	}
}
