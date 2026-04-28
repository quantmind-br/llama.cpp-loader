// Package pages holds tab page implementations.
package pages

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/validator"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// ProfilesPage is the master-detail page for managing profiles.
type ProfilesPage struct {
	store     profilestore.Store
	schema    domain.FlagSchema
	validator validator.Validator

	list      list.Model
	listKeys  profilesKeyMap
	width     int
	height    int

	// Detail/edit state.
	editing       bool
	subTab        subTab
	form          *huh.Form
	draft         profileDraft
	confirmDelete bool
	confirmForm   *huh.Form
	confirmAnswer *bool // heap-allocated so address remains valid across Update copies

	// Advanced sub-tab state.
	advanced       table.Model
	advancedAll    []table.Row
	advancedFilter string
	filterMode     bool

	// Status feedback.
	flash string
}

type profilesKeyMap struct {
	New, Save, Duplicate, Delete, Edit, Cancel, Tab key.Binding
}

func defaultProfilesKeys() profilesKeyMap {
	return profilesKeyMap{
		New:       key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Save:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "save")),
		Duplicate: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "dup")),
		Delete:    key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "del")),
		Edit:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "edit")),
		Cancel:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Tab:       key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "tab editor")),
	}
}

// item adapts domain.Profile to bubbles/list.
type item struct {
	p domain.Profile
}

func (i item) Title() string       { return i.p.Name }
func (i item) Description() string { return i.p.ID }
func (i item) FilterValue() string { return i.p.Name + " " + i.p.ID }

// NewProfilesPage constructs the page wired to a Store and FlagSchema.
func NewProfilesPage(store profilestore.Store, schema domain.FlagSchema) ProfilesPage {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Profiles"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	tbl := newAdvancedTable(schema, 100, 12)
	return ProfilesPage{
		store:       store,
		schema:      schema,
		validator:   validator.New(),
		advanced:    tbl,
		advancedAll: tbl.Rows(),
		list:        l,
		listKeys:    defaultProfilesKeys(),
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
		header := theme.Title.Render(fmt.Sprintf("Editor — [%s]   ctrl+t to switch", p.subTab))
		var body string
		if p.subTab == subTabEssentials {
			body = p.form.View()
		} else {
			body = p.advanced.View()
		}
		report := p.validator.Validate(p.previewProfile(), p.schema)
		var lines []string
		for _, e := range report.Errors {
			lines = append(lines, theme.Error.Render("✗ "+e.Field+": "+e.Message))
		}
		for _, w := range report.Warnings {
			lines = append(lines, theme.Warn.Render("! "+w.Field+": "+w.Message))
		}
		filterLine := ""
		if p.subTab == subTabAdvanced {
			filterLine = theme.Subtitle.Render(fmt.Sprintf("filter: %q", p.advancedFilter))
		}
		footer := strings.Join(lines, "\n")
		return lipgloss.JoinVertical(lipgloss.Left, header, body, filterLine, footer)
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
	if key.Matches(msg, p.listKeys.Tab) {
		if p.subTab == subTabEssentials {
			p.subTab = subTabAdvanced
		} else {
			p.subTab = subTabEssentials
		}
		return p, nil
	}

	if p.subTab == subTabAdvanced {
		switch msg.String() {
		case "/":
			p.filterMode = !p.filterMode
			return p, nil
		case "backspace":
			if p.filterMode && len(p.advancedFilter) > 0 {
				p.advancedFilter = p.advancedFilter[:len(p.advancedFilter)-1]
				p.advanced.SetRows(filterRows(p.advancedAll, p.advancedFilter))
			}
			return p, nil
		}
		if p.filterMode && len(msg.Runes) == 1 {
			p.advancedFilter += string(msg.Runes)
			p.advanced.SetRows(filterRows(p.advancedAll, p.advancedFilter))
			return p, nil
		}
		t, cmd := p.advanced.Update(msg)
		p.advanced = t
		return p, cmd
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

	args := map[string]any{
		"ngl":        float64(ngl),
		"ctx-size":   float64(ctx),
		"port":       float64(port),
		"flash-attn": d.FlashAttn,
	}
	if v, err := strconv.Atoi(d.BatchSize); err == nil {
		args["batch-size"] = float64(v)
	}
	if v, err := strconv.Atoi(d.UBatchSize); err == nil {
		args["ubatch-size"] = float64(v)
	}
	if d.CacheTypeK != "" {
		args["cache-type-k"] = d.CacheTypeK
	}
	if d.CacheTypeV != "" {
		args["cache-type-v"] = d.CacheTypeV
	}
	pr := domain.Profile{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		Model:       d.Model,
		Args:        args,
		Launch:      domain.LaunchConfig{DefaultBackground: true},
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

// previewProfile builds a Profile from the current draft (without saving) for
// the validator. Mirrors commitDraft's mapping but is allocation-only.
func (p ProfilesPage) previewProfile() domain.Profile {
	d := p.draft
	args := map[string]any{
		"flash-attn": d.FlashAttn,
	}
	if v, err := strconv.Atoi(d.NGL); err == nil {
		args["ngl"] = float64(v)
	}
	if v, err := strconv.Atoi(d.CtxSize); err == nil {
		args["ctx-size"] = float64(v)
	}
	if v, err := strconv.Atoi(d.BatchSize); err == nil {
		args["batch-size"] = float64(v)
	}
	if v, err := strconv.Atoi(d.UBatchSize); err == nil {
		args["ubatch-size"] = float64(v)
	}
	if v, err := strconv.Atoi(d.Port); err == nil {
		args["port"] = float64(v)
	}
	if d.CacheTypeK != "" {
		args["cache-type-k"] = d.CacheTypeK
	}
	if d.CacheTypeV != "" {
		args["cache-type-v"] = d.CacheTypeV
	}
	return domain.Profile{
		ID:    p.draft.ID,
		Model: d.Model,
		Args:  args,
	}
}

func (p ProfilesPage) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		p.confirmDelete = false
		p.confirmForm = nil
		p.confirmAnswer = nil
		return p, nil
	}

	updated, cmd := p.confirmForm.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		p.confirmForm = f
	}

	if p.confirmForm != nil && p.confirmForm.State == huh.StateCompleted {
		id := p.draft.ID
		affirmative := p.confirmAnswer != nil && *p.confirmAnswer
		p.confirmDelete = false
		p.confirmForm = nil
		p.confirmAnswer = nil
		if !affirmative {
			p.flash = "delete cancelled"
			return p, nil
		}
		if err := p.store.Delete(id); err != nil {
			p.flash = "delete failed: " + err.Error()
		} else {
			p.flash = "deleted " + id
		}
		return p, p.loadCmd()
	}
	return p, cmd
}

func (p ProfilesPage) startNew() (tea.Model, tea.Cmd) {
	p.draft = profileDraft{
		ID:         "",
		Name:       "New Profile",
		NGL:        "99",
		CtxSize:    "8192",
		BatchSize:  "2048",
		UBatchSize: "512",
		Port:       "8080",
		FlashAttn:  true,
		CacheTypeK: "q8_0",
		CacheTypeV: "q8_0",
		isNew:      true,
	}
	p.form = buildEditorForm(&p.draft, p.schema)
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
		BatchSize:   argString(pr.Args["batch-size"]),
		UBatchSize:  argString(pr.Args["ubatch-size"]),
		Port:        argString(pr.Args["port"]),
		FlashAttn:   argBool(pr.Args["flash-attn"]),
		CacheTypeK:  argString(pr.Args["cache-type-k"]),
		CacheTypeV:  argString(pr.Args["cache-type-v"]),
	}
	p.form = buildEditorForm(&p.draft, p.schema)
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
	answer := false
	p.confirmAnswer = &answer
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Delete profile " + id + "?").
			Affirmative("Delete").
			Negative("Cancel").
			Value(p.confirmAnswer),
	)).WithShowHelp(false).WithShowErrors(false)

	p.confirmForm = form
	p.confirmDelete = true
	// Stash the id so updateConfirm can act on submit.
	p.draft = profileDraft{ID: id} // reuse draft.ID just to carry the id
	return p, form.Init()
}
