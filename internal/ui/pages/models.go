// Package pages holds tab page implementations.
package pages

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/modelscanner"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/components"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// pathStatus tracks per-root scan progress shown above the table.
type pathStatus struct {
	state string // "scanning" | "scanned" | "error"
	count int
	err   string
}

// actionOption is one row in the inline action selector overlays.
type actionOption struct{ label, value string }

// actionMenu is the inline modal used for "what should we do with this
// model file?" and the follow-up "which profile do we update?".
// Replaces the earlier huh.NewForm wrapping a single Select that did not
// reliably reach huh.StateCompleted on a single Enter press.
type actionMenu struct {
	title      string
	options    []actionOption
	cursor     int
	targetPath string // GGUF path the menu acts on
	stage      actionStage
}

type actionStage int

const (
	actionStageRoot actionStage = iota
	actionStagePickProfile
)

// ModelsPage browses GGUF files discovered by ModelScanner.
type ModelsPage struct {
	scanner modelscanner.Scanner
	paths   []string
	store   profilestore.Store
	cancel  context.CancelFunc
	scanID  int // epoch; bumped each rescan to discard stale events

	files     []domain.ModelFile
	statusMap map[string]pathStatus

	table      table.Model
	width      int
	height     int
	filter     string
	filterMode bool
	flash      string

	action *actionMenu

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

// WithProfileStore enables the "Use in existing profile" action by
// giving the page access to the profile store. Without it, that option
// is hidden from the action menu.
func (p ModelsPage) WithProfileStore(store profilestore.Store) ModelsPage {
	p.store = store
	return p
}

// IsCapturingInput tells the root model when the page owns global
// keystrokes (Tab/Shift+Tab) — true while the inline action menu is
// open so cursor navigation does not leak into tab cycling.
func (p ModelsPage) IsCapturingInput() bool {
	return p.action != nil
}

// scanStartedMsg delivers the channel + cancel handle from a fresh scan
// start. State mutations happen when this message lands in Update.
// scanID matches ModelsPage.scanID at start; stale messages from a
// previous rescan are discarded by epoch comparison.
type scanStartedMsg struct {
	scanID int
	ch     <-chan domain.ScanEvent
	cancel context.CancelFunc
	err    error
}

// scanEventMsg carries one ScanEvent plus the channel for re-arming.
type scanEventMsg struct {
	scanID int
	ch     <-chan domain.ScanEvent
	evt    domain.ScanEvent
}

// scanChannelClosedMsg signals the scan goroutine finished and closed
// its channel.
type scanChannelClosedMsg struct {
	scanID int
}

func (p ModelsPage) Init() tea.Cmd {
	return startScanCmd(p.scanner, p.paths, p.scanID)
}

// startScanCmd builds a Cmd that creates ctx+cancel, kicks off the
// scanner, and delivers the channel via scanStartedMsg. The Cmd's
// closure owns the cancel until Update captures it.
func startScanCmd(scanner modelscanner.Scanner, paths []string, scanID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		ch, err := scanner.Scan(ctx, paths)
		if err != nil {
			cancel()
			return scanStartedMsg{scanID: scanID, err: err}
		}
		return scanStartedMsg{scanID: scanID, ch: ch, cancel: cancel}
	}
}

func waitForScanEvent(ch <-chan domain.ScanEvent, scanID int) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return scanChannelClosedMsg{scanID: scanID}
		}
		return scanEventMsg{scanID: scanID, ch: ch, evt: evt}
	}
}

func (p ModelsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = msg.Width, msg.Height
		p.table.SetHeight(msg.Height - 8)
		return p, nil
	case scanStartedMsg:
		if msg.scanID != p.scanID {
			if msg.cancel != nil {
				msg.cancel()
			}
			return p, nil
		}
		if msg.err != nil {
			for _, root := range p.paths {
				p.statusMap[root] = pathStatus{state: "error", err: msg.err.Error()}
			}
			return p, nil
		}
		p.cancel = msg.cancel
		return p, waitForScanEvent(msg.ch, msg.scanID)
	case scanEventMsg:
		if msg.scanID != p.scanID {
			return p, waitForScanEvent(msg.ch, msg.scanID)
		}
		updated, _ := p.handleScanEvent(msg.evt)
		next := updated.(ModelsPage)
		return next, waitForScanEvent(msg.ch, msg.scanID)
	case scanChannelClosedMsg:
		return p, nil
	case tea.KeyMsg:
		if p.action != nil {
			return p.updateActionMenu(msg)
		}
		return p.handleKey(msg)
	}
	return p, nil
}

// updateActionMenu owns the inline action selector while it is on
// screen. Up/Down move the cursor, Enter commits the highlighted option,
// Esc cancels.
func (p ModelsPage) updateActionMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		p.action = nil
		return p, nil
	case "up", "k":
		if p.action.cursor > 0 {
			p.action.cursor--
		}
		return p, nil
	case "down", "j":
		if p.action.cursor < len(p.action.options)-1 {
			p.action.cursor++
		}
		return p, nil
	case "enter":
		opt := p.action.options[p.action.cursor]
		switch p.action.stage {
		case actionStageRoot:
			return p.commitRootAction(opt.value, p.action.targetPath)
		case actionStagePickProfile:
			return p.commitProfileTarget(opt.value, p.action.targetPath)
		}
	}
	return p, nil
}

// commitRootAction handles the choice from the first action menu.
func (p ModelsPage) commitRootAction(choice, path string) (tea.Model, tea.Cmd) {
	switch choice {
	case "new":
		p.action = nil
		return p, func() tea.Msg { return UseInNewProfileMsg{Path: path} }
	case "reveal":
		p.flash = path
		p.action = nil
		return p, nil
	case "existing":
		if p.store == nil {
			p.flash = "profile store not wired"
			p.action = nil
			return p, nil
		}
		profiles, err := p.store.List()
		if err != nil {
			p.flash = "load profiles: " + err.Error()
			p.action = nil
			return p, nil
		}
		if len(profiles) == 0 {
			p.flash = "no existing profiles to update"
			p.action = nil
			return p, nil
		}
		opts := make([]actionOption, 0, len(profiles))
		for _, pr := range profiles {
			opts = append(opts, actionOption{label: pr.Name + " (" + pr.ID + ")", value: pr.ID})
		}
		p.action = &actionMenu{
			title:      "Update which profile?",
			options:    opts,
			targetPath: path,
			stage:      actionStagePickProfile,
		}
		return p, nil
	}
	p.action = nil
	return p, nil
}

// commitProfileTarget applies the selected GGUF path to the chosen
// existing profile and persists.
func (p ModelsPage) commitProfileTarget(profileID, path string) (tea.Model, tea.Cmd) {
	p.action = nil
	if p.store == nil {
		p.flash = "profile store not wired"
		return p, nil
	}
	pr, err := p.store.Get(profileID)
	if err != nil {
		p.flash = "load profile: " + err.Error()
		return p, nil
	}
	pr.Model = path
	if err := p.store.Save(pr); err != nil {
		p.flash = "save profile: " + err.Error()
		return p, nil
	}
	p.flash = "updated " + profileID
	return p, nil
}

func (p ModelsPage) handleScanEvent(evt domain.ScanEvent) (tea.Model, tea.Cmd) {
	switch evt.Type {
	case domain.ScanEventFile:
		if evt.File != nil {
			p.files = append(p.files, *evt.File)
			p.refreshRows()
		}
	case domain.ScanEventProgress:
		st := p.statusMap[evt.Root]
		st.count = evt.Count
		st.state = "scanned"
		p.statusMap[evt.Root] = st
	case domain.ScanEventError:
		st := p.statusMap[evt.Root]
		st.state = "error"
		if evt.Error != nil {
			st.err = evt.Error.Error()
		}
		p.statusMap[evt.Root] = st
	case domain.ScanEventDone:
		// Channel will close right after; nothing to do.
	}
	return p, nil
}

func (p ModelsPage) visibleFiles() []domain.ModelFile {
	files := p.files
	if p.filter != "" {
		q := strings.ToLower(p.filter)
		filtered := make([]domain.ModelFile, 0, len(files))
		for _, f := range files {
			if strings.Contains(strings.ToLower(f.Name), q) {
				filtered = append(filtered, f)
			}
		}
		files = filtered
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return files
}

// refreshRows rebuilds table rows from p.files honoring the current
// filter. Sorted by name for stable display.
func (p *ModelsPage) refreshRows() {
	files := p.visibleFiles()
	rows := make([]table.Row, 0, len(files))
	for _, f := range files {
		rows = append(rows, table.Row{
			truncate(f.Name, 36),
			humanSize(f.SizeBytes),
			f.Quant,
			f.Params,
			truncate(f.Path, 40),
		})
	}
	p.table.SetRows(rows)
}

// humanSize formats bytes as "X.YG" / "X.YM".
func humanSize(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(n)/float64(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(n)/float64(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(n)/float64(1<<10))
	default:
		return fmt.Sprintf("%dB", n)
	}
}

func (p ModelsPage) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, p.keys.Filter):
		p.filterMode = !p.filterMode
		return p, nil
	case key.Matches(msg, p.keys.Cancel):
		if p.filterMode || p.filter != "" {
			p.filterMode = false
			p.filter = ""
			p.refreshRows()
			return p, nil
		}
	case key.Matches(msg, p.keys.Rescan):
		if p.cancel != nil {
			p.cancel()
			p.cancel = nil
		}
		p.scanID++
		p.flash = "rescan started"
		p.files = nil
		for _, root := range p.paths {
			p.statusMap[root] = pathStatus{state: "scanning"}
		}
		p.refreshRows()
		return p, startScanCmd(p.scanner, p.paths, p.scanID)
	case key.Matches(msg, p.keys.Enter):
		if len(p.table.Rows()) == 0 {
			return p, nil
		}
		idx := p.table.Cursor()
		if idx < 0 {
			return p, nil
		}
		// Map row idx to file via filtered ordering. Recompute filtered
		// list to match what's displayed.
		visible := p.visibleFiles()
		if idx >= len(visible) {
			return p, nil
		}
		selected := visible[idx]
		opts := []actionOption{
			{label: "Use in new profile", value: "new"},
		}
		if p.store != nil {
			opts = append(opts, actionOption{label: "Use in existing profile", value: "existing"})
		}
		opts = append(opts, actionOption{label: "Reveal path", value: "reveal"})
		p.action = &actionMenu{
			title:      "Action for " + selected.Name,
			options:    opts,
			targetPath: selected.Path,
			stage:      actionStageRoot,
		}
		return p, nil
	}

	if p.filterMode {
		switch msg.String() {
		case "backspace":
			if len(p.filter) > 0 {
				p.filter = p.filter[:len(p.filter)-1]
				p.refreshRows()
			}
			return p, nil
		}
		if len(msg.Runes) == 1 {
			p.filter += string(msg.Runes)
			p.refreshRows()
			return p, nil
		}
	}

	t, cmd := p.table.Update(msg)
	p.table = t
	return p, cmd
}

func (p ModelsPage) View() string {
	if p.action != nil {
		return p.renderActionMenu()
	}
	header := theme.Title.Render("Models")
	statusLine := p.renderStatus()
	filterLine := ""
	if p.filterMode || p.filter != "" {
		filterLine = theme.Subtitle.Render(fmt.Sprintf("filter: %q", p.filter))
	}
	help := theme.Subtitle.Render("[/] filter  [R] rescan  [enter] actions  [esc] clear" + components.HelpToken)
	footer := ""
	if p.flash != "" {
		footer = theme.Subtitle.Render(p.flash)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, statusLine, p.table.View(), filterLine, help, footer)
}

// renderActionMenu draws the inline modal-ish overlay used for both the
// root action menu and the follow-up profile picker.
func (p ModelsPage) renderActionMenu() string {
	lines := []string{theme.Title.Render(p.action.title)}
	for i, opt := range p.action.options {
		prefix := "  "
		label := opt.label
		if i == p.action.cursor {
			prefix = "> "
			label = theme.OK.Render(label)
		}
		lines = append(lines, prefix+label)
	}
	lines = append(lines, "", theme.Subtitle.Render("[↑/↓] move  [enter] select  [esc] cancel"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (p ModelsPage) renderStatus() string {
	if len(p.paths) == 0 {
		return theme.Subtitle.Render("no search paths configured")
	}
	parts := make([]string, 0, len(p.paths))
	for _, root := range p.paths {
		st := p.statusMap[root]
		var label string
		switch st.state {
		case "scanning":
			label = theme.Subtitle.Render(fmt.Sprintf("%s [scanning]", root))
		case "scanned":
			label = theme.OK.Render(fmt.Sprintf("%s [%d]", root, st.count))
		case "error":
			label = theme.Error.Render(fmt.Sprintf("%s [error: %s]", root, truncate(st.err, 30)))
		default:
			label = theme.Subtitle.Render(root)
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, "  ")
}

// UseInNewProfileMsg requests creating a new profile pre-filled with Path.
// Root catches this message, switches to the Profiles tab, and forwards
// it so ProfilesPage starts a new draft.
type UseInNewProfileMsg struct {
	Path string
}
