// Package pages holds tab page implementations.
package pages

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/modelscanner"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// pathStatus tracks per-root scan progress shown above the table.
type pathStatus struct {
	state string // "scanning" | "scanned" | "error"
	count int
	err   string
}

// ModelsPage browses GGUF files discovered by ModelScanner.
type ModelsPage struct {
	scanner modelscanner.Scanner
	paths   []string
	cancel  context.CancelFunc

	files     []domain.ModelFile
	statusMap map[string]pathStatus

	table      table.Model
	width      int
	height     int
	filter     string
	filterMode bool
	flash      string

	keys modelsKeyMap
}

type modelsKeyMap struct {
	Filter, Rescan, Enter, Cancel key.Binding
}

func defaultModelsKeys() modelsKeyMap {
	return modelsKeyMap{
		Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Rescan: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "rescan")),
		Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "actions")),
		Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear filter")),
	}
}

// NewModelsPage builds a page wired to a Scanner and configured search paths.
func NewModelsPage(scanner modelscanner.Scanner, paths []string) ModelsPage {
	cols := []table.Column{
		{Title: "Name", Width: 36},
		{Title: "Size", Width: 10},
		{Title: "Quant", Width: 10},
		{Title: "Params", Width: 8},
		{Title: "Path", Width: 40},
	}
	t := table.New(table.WithColumns(cols), table.WithFocused(true), table.WithHeight(12))

	statusMap := make(map[string]pathStatus, len(paths))
	for _, p := range paths {
		statusMap[p] = pathStatus{state: "scanning"}
	}
	return ModelsPage{
		scanner:   scanner,
		paths:     paths,
		statusMap: statusMap,
		table:     t,
		keys:      defaultModelsKeys(),
	}
}

// scanStartedMsg delivers the channel + cancel handle from a fresh scan
// start. State mutations happen when this message lands in Update.
type scanStartedMsg struct {
	ch     <-chan domain.ScanEvent
	cancel context.CancelFunc
	err    error
}

// scanEventMsg carries one ScanEvent plus the channel for re-arming.
type scanEventMsg struct {
	ch  <-chan domain.ScanEvent
	evt domain.ScanEvent
}

// scanChannelClosedMsg signals the scan goroutine finished and closed
// its channel.
type scanChannelClosedMsg struct{}

func (p ModelsPage) Init() tea.Cmd {
	return startScanCmd(p.scanner, p.paths)
}

// startScanCmd builds a Cmd that creates ctx+cancel, kicks off the
// scanner, and delivers the channel via scanStartedMsg. The Cmd's
// closure owns the cancel until Update captures it.
func startScanCmd(scanner modelscanner.Scanner, paths []string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		ch, err := scanner.Scan(ctx, paths)
		if err != nil {
			cancel()
			return scanStartedMsg{err: err}
		}
		return scanStartedMsg{ch: ch, cancel: cancel}
	}
}

func waitForScanEvent(ch <-chan domain.ScanEvent) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return scanChannelClosedMsg{}
		}
		return scanEventMsg{ch: ch, evt: evt}
	}
}

func (p ModelsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
		p.table.SetHeight(msg.Height - 8)
		return p, nil
	case scanStartedMsg:
		if msg.err != nil {
			for _, root := range p.paths {
				p.statusMap[root] = pathStatus{state: "error", err: msg.err.Error()}
			}
			return p, nil
		}
		p.cancel = msg.cancel
		return p, waitForScanEvent(msg.ch)
	case scanEventMsg:
		updated, _ := p.handleScanEvent(msg.evt)
		next := updated.(ModelsPage)
		return next, waitForScanEvent(msg.ch)
	case scanChannelClosedMsg:
		return p, nil
	case tea.KeyMsg:
		return p.handleKey(msg)
	}
	return p, nil
}

func (p ModelsPage) handleScanEvent(evt domain.ScanEvent) (tea.Model, tea.Cmd) {
	// Filled in Task 14.
	return p, nil
}

func (p ModelsPage) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Filled in Task 16.
	return p, nil
}

func (p ModelsPage) View() string {
	header := theme.Title.Render("Models")
	return lipgloss.JoinVertical(lipgloss.Left, header, p.table.View())
}
