// Package components contains reusable UI building blocks.
package components

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
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

// Update handles keyboard input and streaming scan events.
func (m ModelPicker) Update(msg tea.Msg) (ModelPicker, tea.Cmd) {
	keys := defaultPickerKeys()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case PickerScanStartedMsg:
		if msg.Err != nil {
			m.scanning = false
			return m, nil
		}
		m.cancel = msg.Cancel
		return m, pickerWaitForEvent(msg.Ch)
	case PickerScanEventMsg:
		switch msg.Evt.Type {
		case domain.ScanEventFile:
			if msg.Evt.File != nil {
				m.files = append(m.files, *msg.Evt.File)
				m.applyFilter()
			}
		case domain.ScanEventDone:
			m.scanning = false
		}
		return m, pickerWaitForEvent(msg.Ch)
	case PickerScanClosedMsg:
		m.scanning = false
		return m, nil
	case tea.KeyMsg:
		if key.Matches(msg, keys.Esc) {
			if m.cancel != nil {
				m.cancel()
			}
			return m, func() tea.Msg { return ModelPickerCancelledMsg{} }
		}
		if key.Matches(msg, keys.Enter) {
			if len(m.filtered) == 0 {
				return m, nil
			}
			path := m.filtered[m.cursor].Path
			if m.cancel != nil {
				m.cancel()
			}
			return m, func() tea.Msg { return ModelPickedMsg{Path: path} }
		}
		if key.Matches(msg, keys.Filter) {
			m.filterMode = !m.filterMode
			return m, nil
		}
		if m.filterMode {
			if key.Matches(msg, keys.Backspace) {
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
					m.applyFilter()
				}
				return m, nil
			}
			if len(msg.Runes) == 1 {
				m.filter += string(msg.Runes)
				m.applyFilter()
				return m, nil
			}
		}
		if key.Matches(msg, keys.Up) {
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		}
		if key.Matches(msg, keys.Down) {
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *ModelPicker) applyFilter() {
	if m.filter == "" {
		m.filtered = append(m.filtered[:0], m.files...)
	} else {
		q := strings.ToLower(m.filter)
		m.filtered = m.filtered[:0]
		for _, f := range m.files {
			if strings.Contains(strings.ToLower(f.Name), q) {
				m.filtered = append(m.filtered, f)
			}
		}
	}
	sort.Slice(m.filtered, func(i, j int) bool { return m.filtered[i].Name < m.filtered[j].Name })
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
	}
}

// View renders the picker as a centered overlay box.
func (m ModelPicker) View() string {
	title := theme.Title.Render("Pick a model")
	hint := "[↑↓] move  [/] filter  [enter] select  [esc] cancel"
	if m.filterMode {
		hint = "[type] add  [backspace] del  [/] exit filter  [enter] select"
	}
	statusLine := ""
	if m.scanning {
		statusLine = theme.Subtitle.Render(fmt.Sprintf("scanning… %d found", len(m.files)))
	} else {
		statusLine = theme.Subtitle.Render(fmt.Sprintf("%d models", len(m.files)))
	}
	filterLine := ""
	if m.filterMode || m.filter != "" {
		filterLine = theme.Subtitle.Render(fmt.Sprintf("filter: %q", m.filter))
	}
	rows := make([]string, 0, len(m.filtered))
	for i, f := range m.filtered {
		row := fmt.Sprintf("%-36s  %-8s  %-6s  %s", truncatePath(f.Name, 36), f.Quant, f.Params, truncatePath(f.Path, 60))
		if i == m.cursor {
			row = theme.Selected.Render(row)
		}
		rows = append(rows, row)
	}
	body := strings.Join(rows, "\n")
	help := theme.Subtitle.Render(hint)
	box := lipgloss.JoinVertical(lipgloss.Left, title, statusLine, filterLine, body, help)
	return theme.Pane.Width(80).Render(box)
}

func truncatePath(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
