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

// InputCapture is the optional contract a page implements to claim
// global keybindings (Tab/Shift+Tab) while a modal/editor/picker is
// active. When IsCapturingInput returns true the root forwards the
// keystroke to the page instead of cycling tabs.
type InputCapture interface {
	IsCapturingInput() bool
}

// Reloader is the optional contract a page implements to refresh its
// state on demand (e.g. on tab focus, when external files may have
// changed).
type Reloader interface {
	Reload() tea.Cmd
}

// bootBlocker carrega o conteúdo de um modal bloqueante exibido sobre toda a UI.
type bootBlocker struct {
	title string
	body  string
}

// RootModel is the top-level tea.Model.
type RootModel struct {
	pages       [4]tea.Model
	active      Tab
	status      components.StatusBar
	width       int
	height      int
	bootBlocker *bootBlocker
	helpOpen    bool
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
		status: components.StatusBar{Hints: "[1-4] tabs  [tab] next  [q] quit" + components.HelpToken},
	}
}

// WithProfilesPage replaces the placeholder Profiles tab with a real model.
// Used by main.go after services are wired.
func (m RootModel) WithProfilesPage(p tea.Model) RootModel {
	m.pages[TabProfiles] = p
	return m
}

// WithModelsPage replaces the placeholder Models tab with a real model.
func (m RootModel) WithModelsPage(p tea.Model) RootModel {
	m.pages[TabModels] = p
	return m
}

// WithLauncherPage replaces the placeholder Launcher tab with a real model.
func (m RootModel) WithLauncherPage(p tea.Model) RootModel {
	m.pages[TabLauncher] = p
	return m
}

// WithMonitorPage replaces the placeholder Monitor tab with a real model.
func (m RootModel) WithMonitorPage(p tea.Model) RootModel {
	m.pages[TabMonitor] = p
	return m
}

// WithStatusWarn sets a warning message on the status bar (used at boot to
// surface schema fallback notices).
func (m RootModel) WithStatusWarn(msg string) RootModel {
	m.status.SetMessage(components.StatusWarn, msg)
	return m
}

// WithBootBlocker mostra um modal bloqueante sobre toda a UI. Usado quando
// algum recurso crítico falta na boot (e.g. llama-server fora do PATH).
// Apenas `q` / `ctrl+c` continuam respondendo enquanto o blocker está ativo.
func (m RootModel) WithBootBlocker(title, body string) RootModel {
	m.bootBlocker = &bootBlocker{title: title, body: body}
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
	if m.bootBlocker != nil {
		if k, ok := msg.(tea.KeyMsg); ok {
			if k.Type == tea.KeyCtrlC || (k.Type == tea.KeyRunes && len(k.Runes) == 1 && k.Runes[0] == 'q') {
				return m, tea.Quit
			}
		}
		if w, ok := msg.(tea.WindowSizeMsg); ok {
			m.width, m.height = w.Width, w.Height
		}
		return m, nil
	}
	switch msg := msg.(type) {
	case pages.LauncherProfilesLoadedMsg:
		updated, cmd := m.pages[TabLauncher].Update(msg)
		m.pages[TabLauncher] = updated
		return m, cmd

	case pages.UseInNewProfileMsg:
		// Switch to Profiles tab and forward the message so the page
		// opens a pre-filled new draft.
		m.active = TabProfiles
		updated, cmd := m.pages[TabProfiles].Update(msg)
		m.pages[TabProfiles] = updated
		return m, cmd

	case pages.SwitchToMonitorMsg:
		m.active = TabMonitor
		// Forward the PID so MonitorPage can refresh + select that row.
		updated, cmd := m.pages[TabMonitor].Update(pages.MonitorSelectPIDMsg{PID: msg.PID})
		m.pages[TabMonitor] = updated
		return m, cmd

	case pages.LaunchProfileMsg:
		m.active = TabLauncher
		updated, cmd := m.pages[TabLauncher].Update(msg)
		m.pages[TabLauncher] = updated
		return m, cmd

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
		// Help toggle is global and pre-empts page routing.
		if m.helpOpen {
			switch msg.String() {
			case "?", "esc":
				m.helpOpen = false
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil // swallow everything else while help is open
		}
		// ctrl+c is the only unconditional global key — it must work
		// even while an editor/huh form is active so the user can
		// always escape. Every other shortcut (q, 1-4, tab, ?) is
		// gated by IsCapturingInput so printable keys reach the form
		// instead of triggering quit/tab-switch/help.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.activePageCapturesInput() {
			if msg.String() == "?" {
				m.helpOpen = true
				return m, nil
			}
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "1":
				return m.activate(TabProfiles)
			case "2":
				return m.activate(TabLauncher)
			case "3":
				return m.activate(TabMonitor)
			case "4":
				return m.activate(TabModels)
			case "tab":
				return m.activate((m.active + 1) % 4)
			case "shift+tab":
				return m.activate((m.active + 3) % 4)
			}
		}
	}

	// Key/Mouse input goes only to the active page.
	if _, isKey := msg.(tea.KeyMsg); isKey {
		updated, cmd := m.pages[m.active].Update(msg)
		m.pages[m.active] = updated
		return m, cmd
	}
	if _, isMouse := msg.(tea.MouseMsg); isMouse {
		updated, cmd := m.pages[m.active].Update(msg)
		m.pages[m.active] = updated
		return m, cmd
	}
	// Other msgs (cmd→msg returns from background pipelines like the
	// model scanner or the monitor periodic tick) are broadcast to every
	// page, so msgs reach the page that owns them even when it isn't
	// active. Pages that don't recognize the type simply return p, nil.
	var cmds []tea.Cmd
	for i, p := range m.pages {
		updated, cmd := p.Update(msg)
		m.pages[i] = updated
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m RootModel) View() string {
	if m.bootBlocker != nil {
		return components.Modal(m.bootBlocker.title, m.bootBlocker.body+"\n\nPress q to quit.", m.width, m.height)
	}
	if m.helpOpen {
		body, err := components.RenderHelp(m.width - 8) // padding
		if err != nil {
			body = components.HelpMarkdown // fallback raw
		}
		return components.Modal("Keybindings", body, m.width, m.height)
	}
	header := m.renderTabs()
	body := m.pages[m.active].View()
	status := m.status.Render(m.width)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, status)
}

// activate switches to the given tab and triggers Reload on the page if
// it implements the Reloader contract. This keeps Profiles in sync with
// external file changes without requiring a TUI restart.
func (m RootModel) activate(t Tab) (tea.Model, tea.Cmd) {
	m.active = t
	if r, ok := m.pages[t].(Reloader); ok {
		return m, r.Reload()
	}
	return m, nil
}

func (m RootModel) activePageCapturesInput() bool {
	if ic, ok := m.pages[m.active].(InputCapture); ok {
		return ic.IsCapturingInput()
	}
	return false
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
