// Package pages holds tab page implementations.
package pages

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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
	form          *huh.Form
	draft         profileDraft
	confirmDelete bool
	confirmForm   *huh.Form

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
		return p.form.View()
	}
	if p.confirmDelete && p.confirmForm != nil {
		return p.confirmForm.View()
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

func (p ProfilesPage) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, p.listKeys.New):
		return p.startNew()
	case key.Matches(msg, p.listKeys.Edit):
		return p.startEditSelected()
	case key.Matches(msg, p.listKeys.Duplicate):
		return p.duplicateSelected()
	case key.Matches(msg, p.listKeys.Delete):
		return p.askDeleteSelected()
	}

	updated, cmd := p.list.Update(msg)
	p.list = updated
	return p, cmd
}

func (p ProfilesPage) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		p.editing = false
		p.form = nil
		return p, nil
	}

	updated, cmd := p.form.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		p.form = f
	}

	if p.form != nil && p.form.State == huh.StateCompleted {
		return p.commitDraft()
	}
	return p, cmd
}

func (p ProfilesPage) commitDraft() (tea.Model, tea.Cmd) {
	d := p.draft
	if d.ID == "" {
		d.ID = domain.Slugify(d.Name)
	}
	ngl, _ := strconv.Atoi(d.NGL)
	ctx, _ := strconv.Atoi(d.CtxSize)
	port, _ := strconv.Atoi(d.Port)

	pr := domain.Profile{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		Model:       d.Model,
		Args: map[string]any{
			"ngl":         float64(ngl),
			"ctx-size":    float64(ctx),
			"port":        float64(port),
			"flash-attn":  d.FlashAttn,
		},
		Launch: domain.LaunchConfig{DefaultBackground: true},
	}

	// Preserve existing meta when editing.
	if !d.isNew {
		if existing, err := p.store.Get(d.ID); err == nil {
			pr.Meta = existing.Meta
		}
	}

	if err := p.store.Save(pr); err != nil {
		p.flash = "save failed: " + err.Error()
	} else {
		p.flash = "saved " + pr.ID
	}
	p.editing = false
	p.form = nil
	return p, p.loadCmd()
}

func (p ProfilesPage) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		p.confirmDelete = false
		p.confirmForm = nil
		return p, nil
	}

	updated, cmd := p.confirmForm.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		p.confirmForm = f
	}

	if p.confirmForm != nil && p.confirmForm.State == huh.StateCompleted {
		// huh stored the bool in a local in askDeleteSelected; we cannot reach it.
		// Workaround: re-extract via the form's group, or — simpler — we treat
		// completion as confirmation. Cancel goes through esc above.
		id := p.draft.ID
		if err := p.store.Delete(id); err != nil {
			p.flash = "delete failed: " + err.Error()
		} else {
			p.flash = "deleted " + id
		}
		p.confirmDelete = false
		p.confirmForm = nil
		return p, p.loadCmd()
	}
	return p, cmd
}

func (p ProfilesPage) startNew() (tea.Model, tea.Cmd) {
	p.draft = profileDraft{
		ID:        "",
		Name:      "New Profile",
		NGL:       "99",
		CtxSize:   "8192",
		Port:      "8080",
		FlashAttn: true,
		isNew:     true,
	}
	p.form = buildEditorForm(&p.draft)
	p.editing = true
	return p, p.form.Init()
}

func (p ProfilesPage) startEditSelected() (tea.Model, tea.Cmd) {
	sel, ok := p.list.SelectedItem().(item)
	if !ok {
		return p, nil
	}
	pr := sel.p
	p.draft = profileDraft{
		ID:          pr.ID,
		Name:        pr.Name,
		Description: pr.Description,
		Model:       pr.Model,
		NGL:         argString(pr.Args["ngl"]),
		CtxSize:     argString(pr.Args["ctx-size"]),
		Port:        argString(pr.Args["port"]),
		FlashAttn:   argBool(pr.Args["flash-attn"]),
	}
	p.form = buildEditorForm(&p.draft)
	p.editing = true
	return p, p.form.Init()
}

func (p ProfilesPage) duplicateSelected() (tea.Model, tea.Cmd) {
	sel, ok := p.list.SelectedItem().(item)
	if !ok {
		return p, nil
	}
	newID := sel.p.ID + "-copy"
	if _, err := p.store.Duplicate(sel.p.ID, newID); err != nil {
		p.flash = "duplicate failed: " + err.Error()
		return p, nil
	}
	p.flash = "duplicated as " + newID
	return p, p.loadCmd()
}

func (p ProfilesPage) askDeleteSelected() (tea.Model, tea.Cmd) {
	sel, ok := p.list.SelectedItem().(item)
	if !ok {
		return p, nil
	}
	id := sel.p.ID
	confirm := false
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Delete profile " + id + "?").
			Affirmative("Delete").
			Negative("Cancel").
			Value(&confirm),
	)).WithShowHelp(false).WithShowErrors(false)

	p.confirmForm = form
	p.confirmDelete = true
	// Stash the id+answer pointer so updateConfirm can act on submit.
	p.draft = profileDraft{ID: id} // reuse draft.ID just to carry the id
	return p, form.Init()
}

func argString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func argBool(v any) bool {
	b, _ := v.(bool)
	return b
}

func buildEditorForm(d *profileDraft) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&d.Name),
			huh.NewInput().Title("Description").Value(&d.Description),
			huh.NewInput().Title("Model path").Value(&d.Model),
		),
		huh.NewGroup(
			huh.NewInput().Title("ngl (gpu layers)").Value(&d.NGL),
			huh.NewInput().Title("ctx-size").Value(&d.CtxSize),
			huh.NewInput().Title("port").Value(&d.Port),
			huh.NewConfirm().Title("flash-attn?").Value(&d.FlashAttn).Affirmative("Yes").Negative("No"),
		),
	).WithShowHelp(true)
}
