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
