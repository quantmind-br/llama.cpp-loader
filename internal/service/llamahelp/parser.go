package llamahelp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

var sectionHeaderRe = regexp.MustCompile(`^-{5}\s+(.+?)\s+params\s+-{5}$`)

// parseSectionHeader returns the section name (e.g., "common") for a header
// line, or "" if the line is not a section header.
func parseSectionHeader(line string) string {
	trimmed := strings.TrimSpace(line)
	m := sectionHeaderRe.FindStringSubmatch(trimmed)
	if m == nil {
		return ""
	}
	return m[1]
}

// flagLineRe matches the canonical "<aliases>  <description>" layout.
// Group 1 = alias chunk (left), Group 2 = description chunk (right).
// Two or more spaces separate the alias chunk from the description.
var flagLineRe = regexp.MustCompile(`^(.*[^ ])\s{2,}(\S.*)$`)

// defaultRe extracts "(default: X)" — first occurrence wins.
var defaultRe = regexp.MustCompile(`\(default:\s*([^,)]+)`)

// parseFlagLine parses a single help line into a FlagSpec. Returns false when
// the line is not a flag definition (header, blank, continuation).
func parseFlagLine(line string) (domain.FlagSpec, bool) {
	if strings.HasPrefix(strings.TrimSpace(line), "(env:") {
		return domain.FlagSpec{}, false
	}
	m := flagLineRe.FindStringSubmatch(strings.TrimRight(line, " "))
	if m == nil {
		return domain.FlagSpec{}, false
	}
	aliasChunk, descChunk := m[1], strings.TrimSpace(m[2])

	short, longs, placeholder := splitAliases(aliasChunk)
	if len(longs) == 0 {
		return domain.FlagSpec{}, false
	}

	spec := domain.FlagSpec{
		Long:     longs[len(longs)-1],
		Short:    short,
		HelpText: descChunk,
	}
	if len(longs) > 1 {
		spec.Aliases = longs[:len(longs)-1]
	}
	spec.Type = inferType(placeholder)
	if d := extractDefault(descChunk); d != nil {
		spec.Default = coerceDefault(spec.Type, d)
	}
	return spec, true
}

// splitAliases parses "-c, --ctx-size N" → short="c", longs=["ctx-size"], placeholder="N".
// Multi-alias: "-ngl, --gpu-layers, --n-gpu-layers N" → short="ngl",
// longs=["gpu-layers","n-gpu-layers"], placeholder="N".
// The last whitespace-delimited token in the alias chunk may be a placeholder
// such as "N", "TYPE", "[on|off|auto]", or "{a,b,c}". If the last token starts
// with '-', there is no placeholder.
func splitAliases(chunk string) (short string, longs []string, placeholder string) {
	parts := strings.Fields(chunk)
	if len(parts) == 0 {
		return "", nil, ""
	}
	// Detect placeholder: last token without leading '-'.
	last := parts[len(parts)-1]
	if !strings.HasPrefix(last, "-") {
		placeholder = last
		parts = parts[:len(parts)-1]
	}
	for _, p := range parts {
		p = strings.TrimSuffix(p, ",")
		if strings.HasPrefix(p, "--") {
			longs = append(longs, strings.TrimPrefix(p, "--"))
		} else if strings.HasPrefix(p, "-") && short == "" {
			short = strings.TrimPrefix(p, "-")
		}
	}
	return short, longs, placeholder
}

// inferType maps the placeholder token to a FlagType. Defaults to FlagTypeBool
// when the placeholder is empty (no value expected).
func inferType(placeholder string) domain.FlagType {
	if placeholder == "" {
		return domain.FlagTypeBool
	}
	// Will be enriched in later tasks (enums, floats, special tokens).
	return domain.FlagTypeInt
}

// extractDefault pulls the first "(default: X)" payload from the description.
// Returns nil if absent.
func extractDefault(desc string) any {
	m := defaultRe.FindStringSubmatch(desc)
	if m == nil {
		return nil
	}
	return strings.TrimSpace(m[1])
}

// coerceDefault converts the raw default string to the FlagSpec's typed value.
// Falls back to the original string on parse failure.
func coerceDefault(t domain.FlagType, raw any) any {
	s, ok := raw.(string)
	if !ok {
		return raw
	}
	switch t {
	case domain.FlagTypeInt:
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err == nil {
			return n
		}
	case domain.FlagTypeBool:
		switch strings.ToLower(s) {
		case "true", "yes", "1":
			return true
		case "false", "no", "0":
			return false
		}
	}
	return s
}
