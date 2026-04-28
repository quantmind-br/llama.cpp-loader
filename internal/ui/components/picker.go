// Package components contains reusable UI building blocks.
package components

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

// ModelScanner is the minimal Scanner contract the picker depends on.
// Mirrors modelscanner.Scanner; redeclared locally to avoid the
// components package importing service/.
type ModelScanner interface {
	Scan(ctx context.Context, paths []string) (<-chan domain.ScanEvent, error)
}

// ModelPickedMsg is emitted when the user selects a model.
type ModelPickedMsg struct {
	Path string
}

// ModelPickerCancelledMsg is emitted on Esc / cancel.
type ModelPickerCancelledMsg struct{}

// ModelPicker is a modal overlay listing GGUF files.
// Streaming pattern mirrors ModelsPage: Init returns a Cmd that creates
// ctx+channel+cancel and delivers them via PickerScanStartedMsg.
type ModelPicker struct {
	scanner ModelScanner
	paths   []string
	cancel  context.CancelFunc

	files      []domain.ModelFile
	filtered   []domain.ModelFile
	cursor     int
	filter     string
	filterMode bool
	scanning   bool

	width  int
	height int
}

// NewModelPicker constructs a picker bound to a Scanner and search paths.
func NewModelPicker(scanner ModelScanner, paths []string) ModelPicker {
	return ModelPicker{scanner: scanner, paths: paths, scanning: true}
}

// PickerScanStartedMsg delivers the scan channel + cancel handle.
type PickerScanStartedMsg struct {
	Ch     <-chan domain.ScanEvent
	Cancel context.CancelFunc
	Err    error
}

// PickerScanEventMsg carries one ScanEvent + the channel for re-arm.
type PickerScanEventMsg struct {
	Ch  <-chan domain.ScanEvent
	Evt domain.ScanEvent
}

// PickerScanClosedMsg signals the scan channel closed.
type PickerScanClosedMsg struct{}

// Init returns a Cmd that starts the scan goroutine and emits
// PickerScanStartedMsg.
func (m ModelPicker) Init() tea.Cmd {
	scanner := m.scanner
	paths := m.paths
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		ch, err := scanner.Scan(ctx, paths)
		if err != nil {
			cancel()
			return PickerScanStartedMsg{Err: err}
		}
		return PickerScanStartedMsg{Ch: ch, Cancel: cancel}
	}
}

func pickerWaitForEvent(ch <-chan domain.ScanEvent) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return PickerScanClosedMsg{}
		}
		return PickerScanEventMsg{Ch: ch, Evt: evt}
	}
}

// pickerKeys for the modal.
type pickerKeys struct {
	Up, Down, Enter, Esc, Filter, Backspace key.Binding
}

func defaultPickerKeys() pickerKeys {
	return pickerKeys{
		Up:        key.NewBinding(key.WithKeys("up", "k")),
		Down:      key.NewBinding(key.WithKeys("down", "j")),
		Enter:     key.NewBinding(key.WithKeys("enter")),
		Esc:       key.NewBinding(key.WithKeys("esc")),
		Filter:    key.NewBinding(key.WithKeys("/")),
		Backspace: key.NewBinding(key.WithKeys("backspace")),
	}
}
