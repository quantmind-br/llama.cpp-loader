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

func TestParseFlagLine_LongOnlyPlaceholder(t *testing.T) {
	cases := []struct {
		name string
		line string
		want domain.FlagSpec
	}{
		{
			name: "port long-only",
			line: "--port PORT                             port to listen (default: 8080)",
			want: domain.FlagSpec{
				Long:     "port",
				Type:     domain.FlagTypeInt,
				Default:  8080,
				HelpText: "port to listen (default: 8080)",
			},
		},
		{
			name: "keep long-only",
			line: "--keep N                                number of tokens to keep from the initial prompt (default: 0, -1 = all)",
			want: domain.FlagSpec{
				Long:     "keep",
				Type:     domain.FlagTypeInt,
				Default:  0,
				HelpText: "number of tokens to keep from the initial prompt (default: 0, -1 = all)",
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

func TestParseFlagLine_BoolNoPlaceholder(t *testing.T) {
	cases := []struct {
		name string
		line string
		want domain.FlagSpec
	}{
		{
			name: "mlock",
			line: "--mlock                                 force system to keep model in RAM rather than swapping or compressing",
			want: domain.FlagSpec{
				Long:     "mlock",
				Type:     domain.FlagTypeBool,
				HelpText: "force system to keep model in RAM rather than swapping or compressing",
			},
		},
		{
			name: "swa-full bool",
			line: "--swa-full                              use full-size SWA cache (default: false)",
			want: domain.FlagSpec{
				Long:     "swa-full",
				Type:     domain.FlagTypeBool,
				Default:  false,
				HelpText: "use full-size SWA cache (default: false)",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseFlagLine(tc.line)
			if !ok {
				t.Fatalf("!ok for %q", tc.line)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
