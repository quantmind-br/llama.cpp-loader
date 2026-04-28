package llamahelp

import (
	"reflect"
	"testing"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

func TestParseSectionHeader(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"----- common params -----", "common"},
		{"----- sampling params -----", "sampling"},
		{"----- example-specific params -----", "example-specific"},
		{"   ----- weird params -----   ", "weird"},
		{"-c,    --ctx-size N", ""},
		{"", ""},
		{"-----", ""},
	}
	for _, tc := range cases {
		t.Run(tc.line, func(t *testing.T) {
			got := parseSectionHeader(tc.line)
			if got != tc.want {
				t.Fatalf("parseSectionHeader(%q) = %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}

func TestParseFlagLine_ShortLongPlaceholder(t *testing.T) {
	cases := []struct {
		name string
		line string
		want domain.FlagSpec
	}{
		{
			name: "ctx-size with N placeholder and default",
			line: "-c,    --ctx-size N                     size of the prompt context (default: 4096, 0 = loaded from model)",
			want: domain.FlagSpec{
				Long:     "ctx-size",
				Short:    "c",
				Type:     domain.FlagTypeInt,
				Default:  4096,
				HelpText: "size of the prompt context (default: 4096, 0 = loaded from model)",
			},
		},
		{
			name: "batch-size",
			line: "-b,    --batch-size N                   logical maximum batch size (default: 2048)",
			want: domain.FlagSpec{
				Long:     "batch-size",
				Short:    "b",
				Type:     domain.FlagTypeInt,
				Default:  2048,
				HelpText: "logical maximum batch size (default: 2048)",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseFlagLine(tc.line)
			if !ok {
				t.Fatalf("parseFlagLine returned !ok for %q", tc.line)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
