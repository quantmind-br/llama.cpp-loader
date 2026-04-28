// Package ui hosts the root Bubbletea model and tab routing.
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/ui/components"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/pages"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// Tab identifies a top-level section.
type Tab int

const (
	TabProfiles Tab = iota
	TabLauncher
	TabMonitor
	TabModels
)

func (t Tab) Title() string {
	switch t {
	case TabProfiles:
		return "Profiles"
	case TabLauncher:
		return "Launcher"
	case TabMonitor:
		return "Monitor"
	case TabModels:
		return "Models"
	default:
		return "?"
	}
}

// Page is the contract every tab page implements.
type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

// RootModel is the top-level tea.Model.
type RootModel struct {
	pages   [4]tea.Model
	active  Tab
	status  components.StatusBar
	width   int
	height  int
}

// NewRoot constructs a RootModel with placeholder pages.
// Slice 1 swaps the Profiles slot with the real implementation in main.go.
func NewRoot(initial Tab) RootModel {
	return RootModel{
		pages: [4]tea.Model{
			pages.Placeholder{TabName: TabProfiles.Title()},
			pages.Placeholder{TabName: TabLauncher.Title()},
			pages.Placeholder{TabName: TabMonitor.Title()},
			pages.Placeholder{TabName: TabModels.Title()},
		},
		active: initial,
		status: components.StatusBar{Hints: "[1-4] tabs  [tab] next  [q] quit"},
	}
}

// WithProfilesPage replaces the placeholder Profiles tab with a real model.
// Used by main.go after services are wired.
func (m RootModel) WithProfilesPage(p tea.Model) RootModel {
	m.pages[TabProfiles] = p
	return m
}

// WithStatusWarn sets a warning message on the status bar (used at boot to
// surface schema fallback notices).
func (m RootModel) WithStatusWarn(msg string) RootModel {
	m.status.SetMessage(components.StatusWarn, msg)
	return m
}

func (m RootModel) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.pages))
	for _, p := range m.pages {
		if c := p.Init(); c != nil {
			cmds = append(cmds, c)
		}
	}
	return tea.Batch(cmds...)
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// forward sized message to all pages so their internal state knows
		var cmds []tea.Cmd
		for i, p := range m.pages {
			updated, cmd := p.Update(tea.WindowSizeMsg{
				Width:  msg.Width,
				Height: msg.Height - 2, // header + status
			})
			m.pages[i] = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.active = TabProfiles
			return m, nil
		case "2":
			m.active = TabLauncher
			return m, nil
		case "3":
			m.active = TabMonitor
			return m, nil
		case "4":
			m.active = TabModels
			return m, nil
		case "tab":
			m.active = (m.active + 1) % 4
			return m, nil
		case "shift+tab":
			m.active = (m.active + 3) % 4
			return m, nil
		}
	}

	// route remaining messages to the active page
	updated, cmd := m.pages[m.active].Update(msg)
	m.pages[m.active] = updated
	return m, cmd
}

func (m RootModel) View() string {
	header := m.renderTabs()
	body := m.pages[m.active].View()
	status := m.status.Render(m.width)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, status)
}

func (m RootModel) renderTabs() string {
	parts := make([]string, 0, 4)
	for i := Tab(0); i < 4; i++ {
		title := fmt.Sprintf("%d %s", int(i)+1, i.Title())
		if i == m.active {
			parts = append(parts, theme.TabActive.Render(title))
		} else {
			parts = append(parts, theme.TabInactive.Render(title))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
