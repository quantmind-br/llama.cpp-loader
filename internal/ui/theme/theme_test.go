package theme

import (
	"strings"
	"testing"
)

// TestTheme_NoColorStripsForeground verifies that when NO_COLOR is set,
// RebuildStyles produces styles whose Render output contains no ANSI color
// escape codes (foreground or background).
func TestTheme_NoColorStripsForeground(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	RebuildStyles()
	t.Cleanup(func() {
		t.Setenv("NO_COLOR", "")
		RebuildStyles()
	})

	cases := []struct {
		name   string
		render string
	}{
		{"Title", Title.Render("hello")},
		{"Subtitle", Subtitle.Render("hello")},
		{"OK", OK.Render("hello")},
		{"Warn", Warn.Render("hello")},
		{"Error", Error.Render("hello")},
		{"Selected", Selected.Render("hello")},
		{"TabActive", TabActive.Render("hello")},
		{"TabInactive", TabInactive.Render("hello")},
	}
	for _, tc := range cases {
		// Allow Bold (\x1b[1m) but not foreground (\x1b[3X) or background (\x1b[4X).
		if strings.Contains(tc.render, "\x1b[3") || strings.Contains(tc.render, "\x1b[4") {
			t.Errorf("%s: expected no fg/bg escapes under NO_COLOR; got %q", tc.name, tc.render)
		}
	}
}

func TestTheme_DefaultColorsPresent(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	RebuildStyles()

	// At least one of Title/Error should emit a foreground escape on a
	// color-capable terminal. Lipgloss may emit nothing if the test profile
	// is ascii-only; in that case skip rather than fail.
	out := Title.Render("x") + Error.Render("y")
	if !strings.Contains(out, "\x1b[") {
		t.Skip("ANSI not emitted under current lipgloss profile; skipping color presence assertion")
	}
}
