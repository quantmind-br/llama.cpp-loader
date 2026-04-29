package components

import (
	"strings"
	"testing"
)

func TestRenderHelp_ContainsKeybindings(t *testing.T) {
	out, err := RenderHelp(80)
	if err != nil {
		t.Fatalf("RenderHelp: %v", err)
	}
	if out == "" {
		t.Fatal("RenderHelp returned empty string")
	}
	if !strings.Contains(out, "Keybindings") {
		t.Errorf("output missing 'Keybindings' header; got:\n%s", out)
	}
}

func TestRenderHelp_MentionsAllTabs(t *testing.T) {
	out, err := RenderHelp(80)
	if err != nil {
		t.Fatalf("RenderHelp: %v", err)
	}
	for _, tab := range []string{"Profiles", "Launcher", "Monitor", "Models"} {
		if !strings.Contains(out, tab) {
			t.Errorf("output missing %q", tab)
		}
	}
}
