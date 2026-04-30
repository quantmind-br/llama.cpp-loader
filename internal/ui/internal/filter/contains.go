// Package filter holds shared slice predicates used across UI pages.
package filter

import "strings"

// ContainsFold returns items whose key contains q (case-insensitive).
// q == "" returns the input slice unchanged (no allocation).
func ContainsFold[T any](items []T, q string, key func(T) string) []T {
	if q == "" {
		return items
	}
	q = strings.ToLower(q)
	out := make([]T, 0, len(items))
	for _, it := range items {
		if strings.Contains(strings.ToLower(key(it)), q) {
			out = append(out, it)
		}
	}
	return out
}
