package components

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

type fakePickerScanner struct {
	events []domain.ScanEvent
}

func (f *fakePickerScanner) Scan(ctx context.Context, paths []string) (<-chan domain.ScanEvent, error) {
	ch := make(chan domain.ScanEvent, len(f.events)+1)
	go func() {
		defer close(ch)
		for _, e := range f.events {
			select {
			case ch <- e:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

// drainScan walks the picker through Init → ScanStarted → all events
// until the channel closes, returning the final picker state.
func drainScan(t *testing.T, p ModelPicker) ModelPicker {
	t.Helper()
	cmd := p.Init()
	if cmd == nil {
		return p
	}
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		var c tea.Cmd
		p, c = p.Update(msg)
		cmd = c
	}
	return p
}

func TestModelPicker_SelectEmitsModelPickedMsg(t *testing.T) {
	mf := domain.ModelFile{Path: "/m/a.gguf", Name: "a.gguf"}
	scanner := &fakePickerScanner{events: []domain.ScanEvent{
		{Type: domain.ScanEventFile, File: &mf},
		{Type: domain.ScanEventDone},
	}}

	p := NewModelPicker(scanner, []string{"/m"})
	p = drainScan(t, p)

	// Trigger Enter and capture the resulting Cmd's message.
	_, c := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if c == nil {
		t.Fatal("expected ModelPickedMsg cmd, got nil")
	}
	got := c()
	picked, ok := got.(ModelPickedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want ModelPickedMsg", got)
	}
	if picked.Path != "/m/a.gguf" {
		t.Fatalf("path = %q", picked.Path)
	}
}

func TestModelPicker_EscEmitsCancel(t *testing.T) {
	scanner := &fakePickerScanner{events: []domain.ScanEvent{{Type: domain.ScanEventDone}}}
	p := NewModelPicker(scanner, nil)
	p = drainScan(t, p)

	_, c := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if c == nil {
		t.Fatal("expected ModelPickerCancelledMsg cmd")
	}
	got := c()
	if _, ok := got.(ModelPickerCancelledMsg); !ok {
		t.Fatalf("msg type = %T, want ModelPickerCancelledMsg", got)
	}
}

func TestModelPicker_FilterReducesList(t *testing.T) {
	files := []domain.ModelFile{
		{Path: "/m/qwen.gguf", Name: "qwen.gguf"},
		{Path: "/m/llama.gguf", Name: "llama.gguf"},
	}
	scanner := &fakePickerScanner{}
	p := NewModelPicker(scanner, nil)
	p.files = files
	p.applyFilter()
	if len(p.filtered) != 2 {
		t.Fatalf("filtered = %d", len(p.filtered))
	}
	p.filter = "qwen"
	p.applyFilter()
	if len(p.filtered) != 1 {
		t.Fatalf("filtered after 'qwen' = %d", len(p.filtered))
	}
	if p.filtered[0].Name != "qwen.gguf" {
		t.Fatalf("filtered[0] = %q", p.filtered[0].Name)
	}
}
