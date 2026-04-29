package pages

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/validator"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// LauncherPage is the Tab 2 page: pick a profile, choose mode, launch.
type LauncherPage struct {
	store     profilestore.Store
	manager   processmgr.Manager
	validator validator.Validator
	schema    domain.FlagSchema

	profiles []domain.Profile
	plist    list.Model

	background bool
	status     string
	running    []domain.RunningInstance

	width, height int
	loadErr       error
}

// NewLauncherPage builds a LauncherPage. manager/validator may be nil for
// smoke tests (UI degrades gracefully and the launch action is disabled).
func NewLauncherPage(store profilestore.Store, manager processmgr.Manager, val validator.Validator) LauncherPage {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 40, 20)
	l.Title = "Profiles"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	return LauncherPage{
		store:      store,
		manager:    manager,
		validator:  val,
		plist:      l,
		background: true,
	}
}

// friendlyLaunchError translates sentinel manager errors into actionable
// hints shown in the page status line.
func friendlyLaunchError(err error) string {
	switch {
	case errors.Is(err, processmgr.ErrPortBusy):
		return "error: port in use — change the profile port or kill the running PID"
	case errors.Is(err, processmgr.ErrModelNotFound):
		return "error: model file not found — fix the profile's Model path"
	case errors.Is(err, processmgr.ErrForegroundBusy):
		return "error: a foreground instance is already running — toggle [b] to background mode"
	case errors.Is(err, processmgr.ErrHealthCheckTimeout):
		return "error: server did not become healthy within timeout — check logs"
	default:
		return "error: " + err.Error()
	}
}

// SetSchema injects the FlagSchema (used by validator at launch time).
func (p LauncherPage) SetSchema(s domain.FlagSchema) LauncherPage {
	p.schema = s
	return p
}

type LauncherProfilesLoadedMsg struct {
	Profiles []domain.Profile
	Err      error
}

// launchedMsg is emitted after a successful Launch + WaitHealthy.
type launchedMsg struct {
	inst domain.RunningInstance
}

// launchErrMsg is emitted when validation or Launch itself fails.
type launchErrMsg struct {
	err error
}

type healthyMsg struct{ pid int }

type profileItem struct {
	p domain.Profile
}

func (i profileItem) Title() string { return i.p.Name }
func (i profileItem) Description() string {
	return fmt.Sprintf("%s | port %v", i.p.ID, i.p.Args["port"])
}
func (i profileItem) FilterValue() string { return i.p.Name }

func (p LauncherPage) Init() tea.Cmd {
	return loadProfilesCmd(p.store)
}

func loadProfilesCmd(store profilestore.Store) tea.Cmd {
	return func() tea.Msg {
		got, err := store.List()
		return LauncherProfilesLoadedMsg{Profiles: got, Err: err}
	}
}

func (p LauncherPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
		p.plist.SetSize(msg.Width/2, msg.Height-6)
		return p, nil

	case LauncherProfilesLoadedMsg:
		if msg.Err != nil {
			p.loadErr = msg.Err
			return p, nil
		}
		p.profiles = msg.Profiles
		items := make([]list.Item, len(msg.Profiles))
		for i, pr := range msg.Profiles {
			items[i] = profileItem{p: pr}
		}
		p.plist.SetItems(items)
		return p, nil

	case launchedMsg:
		p.running = append(p.running, msg.inst)
		p.status = fmt.Sprintf("launched %s pid=%d port=%d", msg.inst.ProfileID, msg.inst.PID, msg.inst.Port)
		mgr := p.manager
		port := msg.inst.Port
		pid := msg.inst.PID
		return p, func() tea.Msg {
			if err := mgr.WaitHealthy(pid, port, 30*time.Second); err != nil {
				return launchErrMsg{err: fmt.Errorf("pid %d not healthy: %w", pid, err)}
			}
			return healthyMsg{pid: pid}
		}

	case healthyMsg:
		p.status = fmt.Sprintf("healthy pid=%d", msg.pid)
		pid := msg.pid
		return p, func() tea.Msg { return SwitchToMonitorMsg{PID: pid} }

	case launchErrMsg:
		p.status = friendlyLaunchError(msg.err)
		return p, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			p.background = !p.background
			return p, nil
		case "k":
			if len(p.running) == 0 || p.manager == nil {
				return p, nil
			}
			pid := p.running[len(p.running)-1].PID
			if err := p.manager.Kill(pid); err != nil {
				p.status = "error: " + err.Error()
				return p, nil
			}
			out := p.running[:0]
			for _, ri := range p.running {
				if ri.PID != pid {
					out = append(out, ri)
				}
			}
			p.running = out
			p.status = fmt.Sprintf("killed pid=%d", pid)
			return p, nil
		case "r":
			return p, loadProfilesCmd(p.store)
		case "enter":
			it, ok := p.plist.SelectedItem().(profileItem)
			if !ok || p.manager == nil {
				return p, nil
			}
			selected := it.p
			val := p.validator
			schema := p.schema
			mgr := p.manager
			mode := processmgr.LaunchBackground
			if !p.background {
				mode = processmgr.LaunchForeground
			}
			return p, func() tea.Msg {
				if val != nil {
					rep := val.Validate(selected, schema)
					if rep.HasBlockingErrors() {
						return launchErrMsg{err: fmt.Errorf("validation failed: %d errors", len(rep.Errors))}
					}
				}
				inst, err := mgr.Launch(selected, mode)
				if err != nil {
					return launchErrMsg{err: err}
				}
				return launchedMsg{inst: inst}
			}
		}
	}

	updatedList, cmd := p.plist.Update(msg)
	p.plist = updatedList
	return p, cmd
}

func (p LauncherPage) View() string {
	if p.loadErr != nil {
		return theme.Subtitle.Render(fmt.Sprintf("load profiles: %v", p.loadErr))
	}
	left := p.plist.View()

	var right string
	if it, ok := p.plist.SelectedItem().(profileItem); ok {
		mode := "Foreground"
		if p.background {
			mode = "Background"
		}
		right = lipgloss.JoinVertical(lipgloss.Left,
			theme.Subtitle.Render(it.p.Name),
			fmt.Sprintf("ID:    %s", it.p.ID),
			fmt.Sprintf("Model: %s", it.p.Model),
			fmt.Sprintf("Port:  %v", it.p.Args["port"]),
			fmt.Sprintf("Mode:  [%s]   (b to toggle)", mode),
		)
	} else {
		right = theme.Subtitle.Render("No profile selected")
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)

	footer := "[b] mode  [enter] launch  [k] kill last  [r] refresh"
	if p.status != "" {
		footer = p.status + "  |  " + footer
	}
	runningView := "Running: (none)"
	if len(p.running) > 0 {
		lines := []string{theme.Subtitle.Render("Running")}
		for _, ri := range p.running {
			tag := "fg"
			if ri.Background {
				tag = "bg"
			}
			lines = append(lines, fmt.Sprintf("  %s pid=%d port=%d %s", ri.ProfileID, ri.PID, ri.Port, tag))
		}
		runningView = strings.Join(lines, "\n")
	}

	return lipgloss.JoinVertical(lipgloss.Left, body, "", runningView, footer)
}
