// Package pages holds tab page implementations.
package pages

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// ProfilesPage is the master-detail page for managing profiles.
type ProfilesPage struct {
	store profilestore.Store

	list      list.Model
	listKeys  profilesKeyMap
	width     int
	height    int

	// Detail/edit state.
	editing       bool
	form          interface{} // Placeholder for *huh.Form (task 1.6)
	draft         profileDraft
	confirmDelete bool
	confirmForm   interface{} // Placeholder for *huh.Form (task 1.7)

	// Status feedback.
	flash string
}

type profilesKeyMap struct {
	New       key.Binding
	Save      key.Binding
	Duplicate key.Binding
	Delete    key.Binding
	Edit      key.Binding
	Cancel    key.Binding
}

func defaultProfilesKeys() profilesKeyMap {
	return profilesKeyMap{
		New:       key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Save:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "save")),
		Duplicate: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "dup")),
		Delete:    key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "del")),
		Edit:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "edit")),
		Cancel:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}

// profileDraft is the editor state, mapped from huh form back to a Profile on save.
type profileDraft struct {
	ID          string // immutable once created
	Name        string
	Description string
	Model       string
	NGL         string // strings — converted on save
	CtxSize     string
	Port        string
	FlashAttn   bool
	isNew       bool
}

// item adapts domain.Profile to bubbles/list.
type item struct {
	p domain.Profile
}

func (i item) Title() string       { return i.p.Name }
func (i item) Description() string { return i.p.ID }
func (i item) FilterValue() string { return i.p.Name + " " + i.p.ID }

// NewProfilesPage constructs the page wired to a Store.
func NewProfilesPage(store profilestore.Store) ProfilesPage {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Profiles"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return ProfilesPage{
		store:    store,
		list:     l,
		listKeys: defaultProfilesKeys(),
	}
}

// loadedMsg is emitted by the load command.
type loadedMsg struct {
	profiles []domain.Profile
	err      error
}

func (p ProfilesPage) Init() tea.Cmd {
	return p.loadCmd()
}

func (p ProfilesPage) loadCmd() tea.Cmd {
	return func() tea.Msg {
		ps, err := p.store.List()
		return loadedMsg{profiles: ps, err: err}
	}
}

func (p ProfilesPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
		p.list.SetSize(msg.Width/3, msg.Height-2)
		return p, nil

	case loadedMsg:
		if msg.err != nil {
			p.flash = "load error: " + msg.err.Error()
			return p, nil
		}
		items := make([]list.Item, 0, len(msg.profiles))
		for _, pr := range msg.profiles {
			items = append(items, item{p: pr})
		}
		p.list.SetItems(items)
		return p, nil

	case tea.KeyMsg:
		if p.editing {
			return p.updateForm(msg)
		}
		if p.confirmDelete {
			return p.updateConfirm(msg)
		}
		return p.updateList(msg)
	}

	return p, nil
}

func (p ProfilesPage) View() string {
	if p.editing && p.form != nil {
		// Task 1.6 will cast form to *huh.Form and call View()
		return fmt.Sprintf("Form editing: %v", p.form)
	}
	if p.confirmDelete && p.confirmForm != nil {
		// Task 1.7 will cast confirmForm to *huh.Form and call View()
		return fmt.Sprintf("Confirm delete: %v", p.confirmForm)
	}

	left := theme.Pane.Width(p.width / 3).Render(p.list.View())
	right := theme.Pane.Width((p.width*2)/3 - 2).Render(p.detailView())
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	if p.flash != "" {
		body = lipgloss.JoinVertical(lipgloss.Left, body, theme.Subtitle.Render(p.flash))
	}
	return body
}

func (p ProfilesPage) detailView() string {
	if len(p.list.Items()) == 0 {
		return theme.Subtitle.Render("No profiles yet. Press [n] to create one.")
	}
	sel, ok := p.list.SelectedItem().(item)
	if !ok {
		return ""
	}
	pr := sel.p
	return fmt.Sprintf(
		"%s\n%s\n\nID:    %s\nModel: %s\nArgs:  ngl=%v ctx=%v port=%v flash-attn=%v\n\n%s",
		theme.Title.Render(pr.Name),
		theme.Subtitle.Render(pr.Description),
		pr.ID,
		pr.Model,
		pr.Args["ngl"], pr.Args["ctx-size"], pr.Args["port"], pr.Args["flash-attn"],
		theme.Subtitle.Render("[enter] edit  [n] new  [d] dup  [x] del"),
	)
}

// Stubs to be filled in by tasks 1.6, 1.7, 1.8.

func (p ProfilesPage) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd)    { return p, nil }
func (p ProfilesPage) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd)    { return p, nil }
func (p ProfilesPage) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) { return p, nil }
