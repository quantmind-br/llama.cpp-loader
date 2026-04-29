package components

import (
	"strings"
	"testing"
)

func TestModal_RendersTitleAndBody(t *testing.T) {
	out := Modal("Help", "Body content here", 80, 24)
	if !strings.Contains(out, "Help") {
		t.Errorf("missing title in output:\n%s", out)
	}
	if !strings.Contains(out, "Body content here") {
		t.Errorf("missing body in output:\n%s", out)
	}
}

func TestModal_FillsViewport(t *testing.T) {
	out := Modal("T", "B", 40, 10)
	// Lipgloss.Place fills with the given dimensions; the rendered string
	// should have at least 9 newlines (for 10 lines).
	if strings.Count(out, "\n") < 9 {
		t.Errorf("expected viewport-filling output; got %d newlines:\n%s", strings.Count(out, "\n"), out)
	}
}

func TestModal_NoSizePassesThrough(t *testing.T) {
	// width/height = 0 -> raw box, no Place wrapper.
	out := Modal("T", "B", 0, 0)
	if strings.Count(out, "\n") > 6 {
		t.Errorf("expected compact box; got %d newlines:\n%s", strings.Count(out, "\n"), out)
	}
}
