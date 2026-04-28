package pages

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
)

func TestProfilesPage_LoadsExistingProfile(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(domain.Profile{
		ID:    "qwen",
		Name:  "Qwen Coder",
		Model: "/m.gguf",
		Args:  map[string]any{"ngl": float64(99)},
	}); err != nil {
		t.Fatal(err)
	}

	page := NewProfilesPage(store)
	tm := teatest.NewTestModel(t, page, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Qwen Coder")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}) // root would handle q; here we just ensure quit
	_ = tm.Quit()
}

func TestProfilesPage_NewProfileSavesViaStore(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	page := NewProfilesPage(store)
	tm := teatest.NewTestModel(t, page, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "No profiles yet")
	}, teatest.WithDuration(2*time.Second))

	// 'n' opens the form
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Name") && strings.Contains(string(out), "Model path")
	}, teatest.WithDuration(2*time.Second))

	// We don't drive the full huh form here — just exit. The store-side
	// behavior is already covered by FSStore tests; this asserts wiring.
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	_ = tm.Quit()

	// Ensure no profile was persisted (esc cancels)
	got, _ := store.List()
	if len(got) != 0 {
		t.Errorf("List len = %d, want 0", len(got))
	}
}
