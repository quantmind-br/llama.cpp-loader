package pages

import (
	"fmt"

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

// SetSchema injects the FlagSchema (used by validator at launch time).
func (p LauncherPage) SetSchema(s domain.FlagSchema) LauncherPage {
	p.schema = s
	return p
}

type launcherProfilesLoadedMsg struct {
	profiles []domain.Profile
	err      error
}

type profileItem struct {
	p domain.Profile
}

func (i profileItem) Title() string       { return i.p.Name }
func (i profileItem) Description() string { return fmt.Sprintf("%s | port %v", i.p.ID, i.p.Args["port"]) }
func (i profileItem) FilterValue() string { return i.p.Name }

func (p LauncherPage) Init() tea.Cmd {
	return loadProfilesCmd(p.store)
}

func loadProfilesCmd(store profilestore.Store) tea.Cmd {
	return func() tea.Msg {
		got, err := store.List()
		return launcherProfilesLoadedMsg{profiles: got, err: err}
	}
}

func (p LauncherPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
		p.plist.SetSize(msg.Width/2, msg.Height-6)
		return p, nil

	case launcherProfilesLoadedMsg:
		if msg.err != nil {
			p.loadErr = msg.err
			return p, nil
		}
		p.profiles = msg.profiles
		items := make([]list.Item, len(msg.profiles))
		for i, pr := range msg.profiles {
			items[i] = profileItem{p: pr}
		}
		p.plist.SetItems(items)
		return p, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			p.background = !p.background
			return p, nil
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

	footer := "[b] mode  [enter] launch  [k] kill  [r] refresh"
	if p.status != "" {
		footer = p.status + "  |  " + footer
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}
