// Package pages holds tab page implementations.
package pages

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/validator"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/components"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// ProfilesPage is the master-detail page for managing profiles.
type ProfilesPage struct {
	store     profilestore.Store
	schema    domain.FlagSchema
	validator validator.Validator

	list     list.Model
	listKeys profilesKeyMap
	width    int
	height   int

	// Detail/edit state.
	editing bool
	subTab  subTab
	form    *huh.Form
	// draft is a heap pointer so &draft.Field bindings inside the huh form
	// stay valid across the value-receiver copies of ProfilesPage that
	// bubbletea makes on every Update. A value field would escape to a
	// different address each copy and the form would write to a stale draft.
	draft         *profileDraft
	confirmDelete bool
	confirmForm   *huh.Form
	confirmAnswer *bool // heap-allocated so address remains valid across Update copies

	// Discard-unsaved-changes confirmation overlay (UIUX-002).
	editorOpenSnapshot   profileDraft
	confirmDiscardForm   *huh.Form
	confirmDiscardAnswer *bool

	// Advanced sub-tab state.
	advanced       table.Model
	advancedAll    []table.Row
	advancedFilter string
	filterMode     bool

	// Status feedback.
	flash   string
	flashAt time.Time

	// Picker overlay (slice 3).
	pickerActive bool
	picker       components.ModelPicker
	scanner      components.ModelScanner
	scanPaths    []string
}

type profilesKeyMap struct {
	New, Save, Duplicate, Delete, Edit, Cancel, Tab, Launch key.Binding
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
		Launch:    key.NewBinding(key.WithKeys("L"), key.WithHelp("L", "launch")),
	}
}

// item adapts domain.Profile to bubbles/list.
type item struct {
	p domain.Profile
}

func (i item) Title() string       { return i.p.Name }
func (i item) Description() string { return i.p.ID }
func (i item) FilterValue() string { return i.p.Name + " " + i.p.ID }

// corruptItem is a list row representing a profile JSON entry that failed
// to parse. Edit/duplicate are no-ops; delete is allowed so the user can
// remove the bad file. Implementa design § 8 — "marca entry com ⚠,
// exclui de operações até user fix/delete".
type corruptItem struct {
	id  string
	err error
}

func (c corruptItem) Title() string       { return "⚠ " + c.id }
func (c corruptItem) Description() string { return "corrupt: " + c.err.Error() }
func (c corruptItem) FilterValue() string { return c.id }

// profileItemDelegate wraps list.DefaultDelegate to paint corruptItem rows
// in warn/error theme colors so they stand out from healthy entries. Healthy
// rows fall through to the default delegate's rendering.
type profileItemDelegate struct {
	list.DefaultDelegate
}

func newProfileItemDelegate() profileItemDelegate {
	return profileItemDelegate{DefaultDelegate: list.NewDefaultDelegate()}
}

func (d profileItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	c, ok := item.(corruptItem)
	if !ok {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}
	title := theme.Error.Render("⚠ " + c.id)
	desc := theme.Warn.Render("corrupt: " + c.err.Error())
	if index == m.Index() {
		// Mirror the default delegate's "selected" indent ("> ") to keep the
		// cursor position visible even on corrupt rows.
		fmt.Fprintf(w, "> %s\n  %s", title, desc)
		return
	}
	fmt.Fprintf(w, "  %s\n  %s", title, desc)
}

// NewProfilesPage constructs the page wired to a Store and FlagSchema.
func NewProfilesPage(store profilestore.Store, schema domain.FlagSchema) ProfilesPage {
	delegate := newProfileItemDelegate()
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

// WithModelScanner enables the ctrl+p model picker overlay in the editor.
func (p ProfilesPage) WithModelScanner(scanner components.ModelScanner, paths []string) ProfilesPage {
	p.scanner = scanner
	p.scanPaths = paths
	return p
}

// loadedMsg is emitted by the load command.
type loadedMsg struct {
	profiles []domain.Profile
	diags    []profilestore.ListDiagnostic
	err      error
}

func (p ProfilesPage) Init() tea.Cmd {
	return p.loadCmd()
}

func (p ProfilesPage) loadCmd() tea.Cmd {
	return func() tea.Msg {
		ps, diags, err := p.store.ListWithDiagnostics()
		return loadedMsg{profiles: ps, diags: diags, err: err}
	}
}

func (p ProfilesPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ---- Phase 1: Layout and lifecycle messages ----
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
		p.list.SetSize(msg.Width/3, msg.Height-2)
		return p, nil

	case flashClearMsg:
		if msg.tag == "profiles" && msg.at.Equal(p.flashAt) {
			p.flash = ""
			p.flashAt = time.Time{}
		}
		return p, nil

	case loadedMsg:
		if msg.err != nil {
			p, fc := p.withFlash("load error: " + msg.err.Error())
			return p, fc
		}
		items := make([]list.Item, 0, len(msg.profiles)+len(msg.diags))
		for _, pr := range msg.profiles {
			items = append(items, item{p: pr})
		}
		for _, d := range msg.diags {
			items = append(items, corruptItem{id: d.ID, err: d.Err})
		}
		p.list.SetItems(items)
		return p, nil

	// ---- Phase 2: Model picker overlay messages ----
	case components.PickerScanStartedMsg, components.PickerScanEventMsg, components.PickerScanClosedMsg:
		if p.pickerActive {
			return p.updatePicker(msg)
		}
		return p, nil

	case UseInNewProfileMsg:
		// Open a new draft pre-filled with the selected model path.
		p.draft = &profileDraft{
			ID:         "",
			Name:       "New Profile",
			Model:      msg.Path,
			NGL:        "99",
			CtxSize:    "8192",
			BatchSize:  "2048",
			UBatchSize: "512",
			Port:       "8080",
			FlashAttn:  "auto",
			CacheTypeK: "q8_0",
			CacheTypeV: "q8_0",
			isNew:      true,
		}
		p.editorOpenSnapshot = *p.draft
		p.form = buildEditorForm(p.draft, p.schema)
		p.editing = true
		p.subTab = subTabEssentials
		p.advancedFilter = ""
		p.filterMode = false
		p.advanced.SetRows(p.advancedAll)
		p, fc := p.withFlash("new profile prefilled with picked model")
		return p, tea.Batch(p.form.Init(), fc)

	// ---- Phase 3: Model picker result messages ----
	case components.ModelPickedMsg:
		if p.draft != nil {
			p.draft.Model = msg.Path
		}
		p.pickerActive = false
		if c := p.picker.Cancel(); c != nil {
			c()
		}
		// Rebuild form so the new Model value is shown if user is on
		// essentials sub-tab.
		p.form = buildEditorForm(p.draft, p.schema)
		return p, p.form.Init()

	case components.ModelPickerCancelledMsg:
		p.pickerActive = false
		return p, nil

	// ---- Phase 4: Key routing — dispatch to the active sub-mode ----
	// Priority: discard confirm > editor form > delete confirm > list nav.
	case tea.KeyMsg:
		if p.confirmDiscardForm != nil {
			return p.updateConfirmDiscard(msg)
		}
		if p.editing {
			return p.updateForm(msg)
		}
		if p.confirmDelete {
			return p.updateConfirm(msg)
		}
		return p.updateList(msg)
	}

	// ---- Phase 5: Forward non-key messages to active huh surfaces ----
	// This is required so internal Cmd→Msg loops (focus init, async
	// validation, button styling refresh) actually fire. Without this
	// the form never completes its Init() handshake.
	//
	// Priority must match Phase 4 key routing: discard confirm > editor
	// form > delete confirm.  If discard is open p.editing is still true,
	// so we check discard first to avoid stealing its Init handshake.
	// Non-key forwarding also drives huh forms to StateCompleted via
	// async nextFieldMsg / nextGroupMsg msgs. Each branch must check for
	// completion and run the finalize path; otherwise the form stays
	// referenced (View() blanks, IsCapturingInput stays true).
	if p.confirmDiscardForm != nil {
		updated, cmd := p.confirmDiscardForm.Update(msg)
		if f, ok := updated.(*huh.Form); ok {
			p.confirmDiscardForm = f
		}
		if p.confirmDiscardForm != nil && p.confirmDiscardForm.State == huh.StateCompleted {
			p = p.finalizeConfirmDiscard()
		}
		return p, cmd
	}
	if p.editing && p.form != nil {
		updated, cmd := p.form.Update(msg)
		if f, ok := updated.(*huh.Form); ok {
			p.form = f
		}
		if p.form != nil && p.form.State == huh.StateCompleted {
			return p.commitDraft()
		}
		return p, cmd
	}
	if p.confirmDelete && p.confirmForm != nil {
		updated, cmd := p.confirmForm.Update(msg)
		if f, ok := updated.(*huh.Form); ok {
			p.confirmForm = f
		}
		if p.confirmForm != nil && p.confirmForm.State == huh.StateCompleted {
			return p.finalizeConfirmDelete(cmd)
		}
		return p, cmd
	}

	return p, nil
}

// askConfirmDiscard arms the "discard unsaved changes?" overlay.
func (p ProfilesPage) askConfirmDiscard() (tea.Model, tea.Cmd) {
	answer := false
	p.confirmDiscardAnswer = &answer
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Discard unsaved changes?").
			Affirmative("Discard").
			Negative("Keep editing").
			Value(p.confirmDiscardAnswer),
	)).WithShowHelp(false).WithShowErrors(false)
	p.confirmDiscardForm = form
	return p, form.Init()
}

func (p ProfilesPage) updateConfirmDiscard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		p.confirmDiscardForm = nil
		p.confirmDiscardAnswer = nil
		return p, nil
	}
	updated, cmd := p.confirmDiscardForm.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		p.confirmDiscardForm = f
	}
	if p.confirmDiscardForm != nil && p.confirmDiscardForm.State == huh.StateCompleted {
		p = p.finalizeConfirmDiscard()
	}
	return p, cmd
}

// finalizeConfirmDiscard consumes the discard-confirmation state. Affirmative
// drops the editor draft; Negative keeps the user in the editor.
// Exposed for tests that drive the affirmative/negative paths without
// relying on huh's internal keymap.
func (p ProfilesPage) finalizeConfirmDiscard() ProfilesPage {
	affirmative := p.confirmDiscardAnswer != nil && *p.confirmDiscardAnswer
	p.confirmDiscardForm = nil
	p.confirmDiscardAnswer = nil
	if affirmative {
		p.editing = false
		p.form = nil
		p.draft = nil
	}
	return p
}

func (p ProfilesPage) View() string {
	if p.confirmDiscardForm != nil {
		return p.confirmDiscardForm.View()
	}
	if p.pickerActive {
		return p.picker.View()
	}
	if p.editing && p.form != nil {
		header := theme.Title.Render(fmt.Sprintf("Editor — [%s]   ctrl+t to switch  ctrl+p to pick model", p.subTab))
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
		style := theme.Subtitle
		if !p.flashAt.IsZero() && time.Since(p.flashAt) >= flashDimAfter {
			style = style.Faint(true)
		}
		body = lipgloss.JoinVertical(lipgloss.Left, body, style.Render(p.flash))
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
		"%s\n%s\n\nID:    %s\nModel: %s\nArgs:  ngl=%v ctx=%v port=%v flash-attn=%v",
		theme.Title.Render(pr.Name),
		theme.Subtitle.Render(pr.Description),
		pr.ID,
		pr.Model,
		pr.Args["ngl"], pr.Args["ctx-size"], pr.Args["port"], pr.Args["flash-attn"],
	)
}

// Hints implements ui.HintProvider — returns page-local key reminders for
// the status bar. Varies by editor / picker / confirm / list mode.
func (p ProfilesPage) Hints() string {
	switch {
	case p.confirmDiscardForm != nil:
		return "[←→] choose  [enter] confirm  [esc] cancel"
	case p.pickerActive:
		return "[↑↓] move  [enter] pick  [esc] cancel"
	case p.confirmDelete:
		return "[←→] choose  [enter] confirm"
	case p.editing:
		return "[ctrl+t] sub-tab  [ctrl+p] pick model  [esc] cancel"
	default:
		return "[enter] edit  [n] new  [d] dup  [x] del  [L] launch  [/] filter"
	}
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
	case key.Matches(msg, p.listKeys.Launch):
		return p.launchSelected()
	}

	updated, cmd := p.list.Update(msg)
	p.list = updated
	return p, cmd
}

// launchSelected emits a LaunchProfileMsg for the currently selected
// profile so the root model can switch to the Launcher tab and run it.
func (p ProfilesPage) launchSelected() (tea.Model, tea.Cmd) {
	if _, isCorrupt := p.list.SelectedItem().(corruptItem); isCorrupt {
		p, fc := p.withFlash("corrupt entry — fix the JSON file or delete it")
		return p, fc
	}
	sel, ok := p.list.SelectedItem().(item)
	if !ok {
		return p, nil
	}
	id := sel.p.ID
	return p, func() tea.Msg { return LaunchProfileMsg{ID: id} }
}

// IsCapturingInput tells the root model when the page owns Tab/Shift+Tab.
// True whenever an editor, picker overlay, or any confirm dialog is on
// screen — these states need Tab/arrows/enter to navigate huh forms.
func (p ProfilesPage) IsCapturingInput() bool {
	return p.editing || p.pickerActive || p.confirmDelete || p.confirmDiscardForm != nil
}

// Reload triggers a fresh load from the underlying store. Called by the
// root when the user navigates back to the Profiles tab.
func (p ProfilesPage) Reload() tea.Cmd {
	return p.loadCmd()
}

// withFlash sets the flash message, stamps the time, and returns the
// page plus the Cmd that schedules the lifetime clear. Centralizing this
// keeps every flash site honest about the timer.
func (p ProfilesPage) withFlash(msg string) (ProfilesPage, tea.Cmd) {
	p.flash = msg
	p.flashAt = time.Now()
	return p, scheduleFlashClear("profiles", p.flashAt)
}

func (p ProfilesPage) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	picker, cmd := p.picker.Update(msg)
	p.picker = picker
	return p, cmd
}

func (p ProfilesPage) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if p.pickerActive {
		return p.updatePicker(msg)
	}
	if msg.String() == "ctrl+p" && p.scanner != nil {
		p.picker = components.NewModelPicker(p.scanner, p.scanPaths)
		p.pickerActive = true
		return p, p.picker.Init()
	}
	if msg.String() == "esc" {
		if p.draft != nil && *p.draft != p.editorOpenSnapshot {
			return p.askConfirmDiscard()
		}
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
	if p.draft == nil {
		return p, nil
	}
	d := p.draft
	if d.ID == "" {
		d.ID = domain.Slugify(d.Name)
	}
	pr := d.toProfile()

	// Preserve existing meta when editing.
	if !d.isNew {
		if existing, err := p.store.Get(d.ID); err == nil {
			pr.Meta = existing.Meta
		}
	}

	var fc tea.Cmd
	if err := p.store.Save(pr); err != nil {
		p, fc = p.withFlash("save failed: " + err.Error())
	} else {
		p, fc = p.withFlash("saved " + pr.ID)
	}
	p.editing = false
	p.form = nil
	return p, tea.Batch(p.loadCmd(), fc)
}

// previewProfile builds a Profile from the current draft (without saving)
// for the validator.
func (p ProfilesPage) previewProfile() domain.Profile {
	return p.draft.toProfile()
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
		return p.finalizeConfirmDelete(cmd)
	}
	return p, cmd
}

// finalizeConfirmDelete consumes the delete confirmation. It is called by
// both the key path (updateConfirm) and the non-key forwarding path, since
// huh transitions to StateCompleted via async nextFieldMsg/nextGroupMsg
// after the user's Enter rather than during the original keypress.
func (p ProfilesPage) finalizeConfirmDelete(formCmd tea.Cmd) (tea.Model, tea.Cmd) {
	var id string
	if p.draft != nil {
		id = p.draft.ID
	}
	affirmative := p.confirmAnswer != nil && *p.confirmAnswer
	p.confirmDelete = false
	p.confirmForm = nil
	p.confirmAnswer = nil
	if !affirmative {
		p, fc := p.withFlash("delete cancelled")
		return p, tea.Batch(formCmd, fc)
	}
	var fc tea.Cmd
	if err := p.store.Delete(id); err != nil {
		p, fc = p.withFlash("delete failed: " + err.Error())
	} else {
		p, fc = p.withFlash("deleted " + id)
	}
	return p, tea.Batch(formCmd, p.loadCmd(), fc)
}

func (p ProfilesPage) startNew() (tea.Model, tea.Cmd) {
	p.draft = &profileDraft{
		ID:         "",
		Name:       "New Profile",
		NGL:        "99",
		CtxSize:    "8192",
		BatchSize:  "2048",
		UBatchSize: "512",
		Port:       "8080",
		FlashAttn:  "auto",
		CacheTypeK: "q8_0",
		CacheTypeV: "q8_0",
		isNew:      true,
	}
	p.editorOpenSnapshot = *p.draft
	p.form = buildEditorForm(p.draft, p.schema)
	p.editing = true
	p.subTab = subTabEssentials
	p.advancedFilter = ""
	p.filterMode = false
	p.advanced.SetRows(p.advancedAll)
	return p, p.form.Init()
}

func (p ProfilesPage) startEditSelected() (tea.Model, tea.Cmd) {
	if _, isCorrupt := p.list.SelectedItem().(corruptItem); isCorrupt {
		p, fc := p.withFlash("selected entry is corrupt — delete it (x) or fix the JSON file")
		return p, fc
	}
	sel, ok := p.list.SelectedItem().(item)
	if !ok {
		return p, nil
	}
	pr := sel.p
	p.draft = &profileDraft{
		ID:          pr.ID,
		Name:        pr.Name,
		Description: pr.Description,
		Model:       pr.Model,
		NGL:         argString(pr.Args["ngl"]),
		CtxSize:     argString(pr.Args["ctx-size"]),
		BatchSize:   argString(pr.Args["batch-size"]),
		UBatchSize:  argString(pr.Args["ubatch-size"]),
		Port:        argString(pr.Args["port"]),
		FlashAttn:   flashAttnToString(pr.Args["flash-attn"]),
		CacheTypeK:  argString(pr.Args["cache-type-k"]),
		CacheTypeV:  argString(pr.Args["cache-type-v"]),
	}
	p.editorOpenSnapshot = *p.draft
	p.form = buildEditorForm(p.draft, p.schema)
	p.editing = true
	p.subTab = subTabEssentials
	p.advancedFilter = ""
	p.filterMode = false
	p.advanced.SetRows(p.advancedAll)
	return p, p.form.Init()
}

func (p ProfilesPage) duplicateSelected() (tea.Model, tea.Cmd) {
	if _, isCorrupt := p.list.SelectedItem().(corruptItem); isCorrupt {
		p, fc := p.withFlash("selected entry is corrupt — delete it (x) or fix the JSON file")
		return p, fc
	}
	sel, ok := p.list.SelectedItem().(item)
	if !ok {
		return p, nil
	}
	newID := sel.p.ID + "-copy"
	if _, err := p.store.Duplicate(sel.p.ID, newID); err != nil {
		p, fc := p.withFlash("duplicate failed: " + err.Error())
		return p, fc
	}
	p, fc := p.withFlash("duplicated as " + newID)
	return p, tea.Batch(p.loadCmd(), fc)
}

func (p ProfilesPage) askDeleteSelected() (tea.Model, tea.Cmd) {
	var id string
	switch sel := p.list.SelectedItem().(type) {
	case item:
		id = sel.p.ID
	case corruptItem:
		id = sel.id
	default:
		return p, nil
	}
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
	p.draft = &profileDraft{ID: id} // reuse draft.ID just to carry the id
	return p, form.Init()
}
