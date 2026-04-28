package modelscanner

import (
	"regexp"
	"strings"
)

// quantRe matches well-known llama.cpp quant labels in filenames.
// Order matters: longer (more specific) patterns come first so partial
// matches (e.g. Q4 vs Q4_K_M) yield the more specific result.
var quantRe = regexp.MustCompile(`(?i)(IQ[1-4](?:_[A-Z]+)?|Q[1-8]_K_[MS]|Q[1-8]_K|Q[1-8]_[01]|F16|F32|BF16)`)

// parseQuant extracts the quant tag from a GGUF filename. Returns "" if
// no recognized pattern is found.
func parseQuant(name string) string {
	m := quantRe.FindString(name)
	if m == "" {
		return ""
	}
	return strings.ToUpper(m)
}

// paramsRe matches parameter-count labels in filenames:
// "32B", "7B", "6.7B", "3.8B", "8x7B" (mixture-of-experts).
// Captures the number+B with optional decimal and optional NxM prefix.
var paramsRe = regexp.MustCompile(`(?i)(?:^|[-_])(\d+(?:x\d+)?(?:\.\d+)?B)(?:[-_.]|$)`)

// parseParams extracts the parameter-count tag from a GGUF filename.
// Returns "" if no recognizable size label is present.
func parseParams(name string) string {
	m := paramsRe.FindStringSubmatch(name)
	if len(m) < 2 {
		return ""
	}
	// Preserve mixed-case labels like "8x7B"; only normalise the trailing B.
	s := m[1]
	return s[:len(s)-1] + strings.ToUpper(string(s[len(s)-1]))
}
