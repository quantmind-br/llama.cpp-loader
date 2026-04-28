package pages

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

// fakeScanner emits a fixed sequence of events for tests.
type fakeScanner struct {
	events []domain.ScanEvent
}

func (f *fakeScanner) Scan(ctx context.Context, paths []string) (<-chan domain.ScanEvent, error) {
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

func TestModelsPage_LoadsFilesIntoTable(t *testing.T) {
	mf := domain.ModelFile{
		Path:      "/tmp/models/qwen-32b.gguf",
		SizeBytes: 16_000_000_000,
		Name:      "qwen-32b.gguf",
		Quant:     "Q4_K_M",
		Params:    "32B",
	}
	scanner := &fakeScanner{events: []domain.ScanEvent{
		{Type: domain.ScanEventFile, Root: "/tmp/models", File: &mf},
		{Type: domain.ScanEventProgress, Root: "/tmp/models", Count: 1},
		{Type: domain.ScanEventDone},
	}}

	page := NewModelsPage(scanner, []string{"/tmp/models"})
	model := tea.Model(page)

	cmd := model.Init()
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		var c tea.Cmd
		model, c = model.Update(msg)
		cmd = c
	}

	mp := model.(ModelsPage)
	if len(mp.files) != 1 {
		t.Fatalf("files = %d, want 1", len(mp.files))
	}
	if mp.files[0].Name != "qwen-32b.gguf" {
		t.Errorf("Name = %q", mp.files[0].Name)
	}
	if mp.statusMap["/tmp/models"].state != "scanned" {
		t.Errorf("status state = %q, want scanned", mp.statusMap["/tmp/models"].state)
	}
	if mp.statusMap["/tmp/models"].count != 1 {
		t.Errorf("status count = %d, want 1", mp.statusMap["/tmp/models"].count)
	}
}

func TestModelsPage_FilterModeReducesRows(t *testing.T) {
	files := []domain.ModelFile{
		{Path: "/m/a.gguf", Name: "alpha.gguf", Quant: "Q4_K_M"},
		{Path: "/m/b.gguf", Name: "beta.gguf", Quant: "Q5_K_M"},
		{Path: "/m/c.gguf", Name: "gamma.gguf", Quant: "Q8_0"},
	}
	scanner := &fakeScanner{}
	page := NewModelsPage(scanner, nil)
	page.files = files
	page.refreshRows()
	if got := len(page.table.Rows()); got != 3 {
		t.Fatalf("rows pre-filter = %d, want 3", got)
	}

	page.filter = "alpha"
	page.refreshRows()
	if got := len(page.table.Rows()); got != 1 {
		t.Fatalf("rows post-filter = %d, want 1", got)
	}
}
