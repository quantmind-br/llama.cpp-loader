package pages

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/monitor"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/components"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

// SubViewKind selects which bottom region the MonitorPage renders.
type SubViewKind int

const (
	SubViewLogs SubViewKind = iota
	SubViewSlots
	SubViewMetrics
)

// monitorEventMsg wraps a monitor.MonitorEvent received from a per-instance
// subscription channel. Re-armed via listenCmd after each delivery.
type monitorEventMsg struct {
	ev monitor.MonitorEvent
}

// listenCmd reads one event from ch and re-arms itself when handled.
func listenCmd(ch <-chan monitor.MonitorEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		return monitorEventMsg{ev: ev}
	}
}

// procMgrIface is the slice of processmgr.Manager that MonitorPage needs.
type procMgrIface interface {
	List() []domain.RunningInstance
	Kill(pid int) error
	Launch(domain.Profile, processmgr.LaunchMode) (domain.RunningInstance, error)
	TailLogs(pid int) (io.ReadCloser, error)
}

// profileStoreIface é o subset de profilestore.Store usado pela MonitorPage
// para implementar `r` (restart real). nil -> `r` cai em modo kill-only.
type profileStoreIface interface {
	Get(id string) (domain.Profile, error)
}

// subState holds per-instance subscription state. Fields populated by later
// tasks (T12-T14); skeleton tracks just the cancel func.
type subState struct {
	cancel func() error
	logs   []string
	slots  monitor.SlotSnapshot
	gpu    monitor.GPUStats
	health monitor.HealthStatus
	mets   monitor.Metrics
}

// Apply mutates the subState according to the concrete type carried by ev.Data.
// When paused is true, log lines are dropped (other event kinds still update).
// The log buffer is capped at 2000 lines.
func (s *subState) Apply(ev monitor.MonitorEvent, paused bool) {
	switch d := ev.Data.(type) {
	case monitor.LogLine:
		if !paused {
			s.logs = append(s.logs, d.Line)
			if len(s.logs) > 2000 {
				s.logs = s.logs[len(s.logs)-2000:]
			}
		}
	case monitor.SlotSnapshot:
		s.slots = d
	case monitor.GPUStats:
		s.gpu = d
	case monitor.HealthStatus:
		s.health = d
	case monitor.Metrics:
		s.mets = d
	}
}

type MonitorPage struct {
	pm                 procMgrIface
	mm                 monitor.Manager
	ps                 profileStoreIface // injected for `r` real restart (slice 6 / Task 4)
	pendingSelectPID   int               // set by MonitorSelectPIDMsg, consumed after the next refresh
	tbl                table.Model
	subs               map[int]*subState
	chans              map[int]<-chan monitor.MonitorEvent
	subView            SubViewKind
	paused             bool
	periodicTickActive bool
	width              int
	height             int

	// Kill confirmation overlay (UIUX-002).
	confirmKillForm     *huh.Form
	confirmKillAnswer   *bool
	confirmKillTargetID int

	// Restart confirmation overlay.
	confirmRestartForm     *huh.Form
	confirmRestartAnswer   *bool
	confirmRestartTargetID int
	confirmRestartProfile  domain.Profile

	flash string
}

// monitorPeriodicTickMsg is delivered every 2s by Init/periodicTickCmd to drive
// background refreshes of the instance list, so crashes detected by the
// processmgr liveness goroutine surface in the UI without user interaction.
type monitorPeriodicTickMsg struct{}

func NewMonitorPage(pm procMgrIface, mm monitor.Manager, ps profileStoreIface) *MonitorPage {
	cols := []table.Column{
		{Title: "PID", Width: 8},
		{Title: "Port", Width: 6},
		{Title: "Profile", Width: 18},
		{Title: "Uptime", Width: 10},
		{Title: "VRAM", Width: 12},
		{Title: "Tokens/s", Width: 10},
	}
	t := table.New(table.WithColumns(cols), table.WithFocused(true), table.WithHeight(8))
	return &MonitorPage{
		pm:    pm,
		mm:    mm,
		ps:    ps,
		tbl:   t,
		subs:  map[int]*subState{},
		chans: map[int]<-chan monitor.MonitorEvent{},
	}
}

func (p *MonitorPage) SetSize(w, h int) {
	p.width, p.height = w, h
	p.tbl.SetWidth(w)
}

func (p *MonitorPage) Init() tea.Cmd {
	p.periodicTickActive = true
	return tea.Batch(p.refreshInstancesCmd(), p.periodicTickCmd())
}

func (p *MonitorPage) refreshInstancesCmd() tea.Cmd {
	return func() tea.Msg { return monitorInstancesRefreshedMsg{insts: p.pm.List()} }
}

func (p *MonitorPage) periodicTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return monitorPeriodicTickMsg{} })
}

type monitorInstancesRefreshedMsg struct {
	insts []domain.RunningInstance
}

// restartResultMsg carries the outcome of an async kill+launch restart.
type restartResultMsg struct {
	pid int
	err error
}

func (p *MonitorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch m := msg.(type) {

	// ---- Phase 1: Layout ----
	case tea.WindowSizeMsg:
		p.SetSize(m.Width, m.Height)

	// ---- Phase 2: Key routing — confirm overlays first, then actions ----
	case tea.KeyMsg:
		if p.confirmKillForm != nil {
			cmds = append(cmds, p.handleConfirmKillKey(m))
			return p, tea.Batch(cmds...)
		}
		if p.confirmRestartForm != nil {
			cmds = append(cmds, p.handleConfirmRestartKey(m))
			return p, tea.Batch(cmds...)
		}
		switch {
		case m.Type == tea.KeyRunes && len(m.Runes) == 1 && m.Runes[0] == 'v':
			p.subView = (p.subView + 1) % 3
		case m.Type == tea.KeyRunes && len(m.Runes) == 1 && m.Runes[0] == 'k':
			if pid := p.selectedPID(); pid > 0 {
				cmds = append(cmds, p.askConfirmKill(pid))
				// Return early so the original 'k' keypress does NOT
				// reach p.tbl.Update below, where the table's default
				// keymap would interpret it as line-up and move the
				// selection away from the row the user is acting on.
				return p, tea.Batch(cmds...)
			}
		case m.Type == tea.KeyRunes && len(m.Runes) == 1 && m.Runes[0] == 'r':
			if pid := p.selectedPID(); pid > 0 {
				cmds = append(cmds, p.askConfirmRestart(pid))
				return p, tea.Batch(cmds...)
			}
		case m.Type == tea.KeySpace:
			p.paused = !p.paused
		}

	// ---- Phase 3: Instance list refresh and deferred selection ----
	case monitorInstancesRefreshedMsg:
		if c := p.applyInstances(m.insts); c != nil {
			cmds = append(cmds, c)
		}
		if p.pendingSelectPID != 0 {
			p.selectRow(p.pendingSelectPID)
			p.pendingSelectPID = 0
		}
	case restartResultMsg:
		if m.err != nil {
			p.flash = fmt.Sprintf("restart: pid %d failed: %v", m.pid, m.err)
		}
		cmds = append(cmds, p.refreshInstancesCmd())
	case MonitorSelectPIDMsg:
		// Defer the cursor move until the next refresh has applied fresh rows,
		// so we never select against a stale row list.
		p.pendingSelectPID = m.PID
		cmds = append(cmds, p.refreshInstancesCmd())
	case monitorPeriodicTickMsg:
		cmds = append(cmds, p.refreshInstancesCmd(), p.periodicTickCmd())

	// ---- Phase 4: Monitor event routing (logs, slots, GPU, health, metrics) ----
	case monitorEventMsg:
		if st, ok := p.subs[m.ev.PID]; ok {
			st.Apply(m.ev, p.paused)
		}
		// Re-arm listener for this PID.
		if ch, ok := p.chans[m.ev.PID]; ok {
			cmds = append(cmds, listenCmd(ch))
		}
	}

	// ---- Phase 5: Forward non-key messages to confirm forms ----
	// Required so huh internal Cmds (focus init, button styling refresh) fire.
	// Each branch must also check StateCompleted and call finalize, since
	// the form transitions via async nextFieldMsg/nextGroupMsg msgs that
	// arrive here (the original Enter never reaches the form's submit).
	if p.confirmKillForm != nil {
		updated, cmd := p.confirmKillForm.Update(msg)
		if f, ok := updated.(*huh.Form); ok {
			p.confirmKillForm = f
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if p.confirmKillForm != nil && p.confirmKillForm.State == huh.StateCompleted {
			if killCmd := p.finalizeConfirmKill(); killCmd != nil {
				cmds = append(cmds, killCmd)
			}
		}
	}
	if p.confirmRestartForm != nil {
		updated, cmd := p.confirmRestartForm.Update(msg)
		if f, ok := updated.(*huh.Form); ok {
			p.confirmRestartForm = f
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if p.confirmRestartForm != nil && p.confirmRestartForm.State == huh.StateCompleted {
			if restartCmd := p.finalizeConfirmRestart(); restartCmd != nil {
				cmds = append(cmds, restartCmd)
			}
		}
	}

	t, tc := p.tbl.Update(msg)
	p.tbl = t
	if tc != nil {
		cmds = append(cmds, tc)
	}
	return p, tea.Batch(cmds...)
}

// askConfirmKill builds and arms the kill-confirmation huh form. Returns
// the form's Init Cmd so its first focus/render dispatch lands.
func (p *MonitorPage) askConfirmKill(pid int) tea.Cmd {
	answer := false
	p.confirmKillAnswer = &answer
	p.confirmKillTargetID = pid
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(fmt.Sprintf("Kill pid=%d?", pid)).
			Affirmative("Kill").
			Negative("Cancel").
			Value(p.confirmKillAnswer),
	)).WithShowHelp(false).WithShowErrors(false)
	p.confirmKillForm = form
	return form.Init()
}

func (p *MonitorPage) handleConfirmKillKey(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		p.confirmKillForm = nil
		p.confirmKillAnswer = nil
		p.confirmKillTargetID = 0
		return nil
	}
	updated, cmd := p.confirmKillForm.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		p.confirmKillForm = f
	}
	if p.confirmKillForm != nil && p.confirmKillForm.State == huh.StateCompleted {
		if killCmd := p.finalizeConfirmKill(); killCmd != nil {
			return tea.Batch(cmd, killCmd)
		}
	}
	return cmd
}

// finalizeConfirmKill consumes the current confirmation state, clears it,
// and returns a refresh Cmd when the user picked the affirmative button.
// Exposed package-internal so the affirmative branch can be exercised in
// tests without driving huh's keymap end-to-end.
func (p *MonitorPage) finalizeConfirmKill() tea.Cmd {
	pid := p.confirmKillTargetID
	affirmative := p.confirmKillAnswer != nil && *p.confirmKillAnswer
	p.confirmKillForm = nil
	p.confirmKillAnswer = nil
	p.confirmKillTargetID = 0
	if !affirmative {
		return nil
	}
	_ = p.pm.Kill(pid)
	return p.refreshInstancesCmd()
}

// askConfirmRestart preloads the profile for the selected PID and arms a
// confirmation form. If the instance or profile is missing, it surfaces the
// error via p.flash instead of opening the form.
func (p *MonitorPage) askConfirmRestart(pid int) tea.Cmd {
	insts := p.pm.List()
	var inst *domain.RunningInstance
	for i := range insts {
		if insts[i].PID == pid {
			inst = &insts[i]
			break
		}
	}
	if inst == nil {
		p.flash = fmt.Sprintf("restart: pid %d not found", pid)
		return nil
	}
	if p.ps == nil {
		p.flash = "restart: profile store not available"
		return nil
	}
	prof, err := p.ps.Get(inst.ProfileID)
	if err != nil {
		p.flash = fmt.Sprintf("restart: profile %q not found", inst.ProfileID)
		return nil
	}
	answer := false
	p.confirmRestartAnswer = &answer
	p.confirmRestartTargetID = pid
	p.confirmRestartProfile = prof
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(fmt.Sprintf("Restart pid=%d (%s)?", pid, prof.Name)).
			Affirmative("Restart").
			Negative("Cancel").
			Value(p.confirmRestartAnswer),
	)).WithShowHelp(false).WithShowErrors(false)
	p.confirmRestartForm = form
	return form.Init()
}

func (p *MonitorPage) handleConfirmRestartKey(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		p.confirmRestartForm = nil
		p.confirmRestartAnswer = nil
		p.confirmRestartTargetID = 0
		p.confirmRestartProfile = domain.Profile{}
		return nil
	}
	updated, cmd := p.confirmRestartForm.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		p.confirmRestartForm = f
	}
	if p.confirmRestartForm != nil && p.confirmRestartForm.State == huh.StateCompleted {
		if restartCmd := p.finalizeConfirmRestart(); restartCmd != nil {
			return tea.Batch(cmd, restartCmd)
		}
	}
	return cmd
}

// finalizeConfirmRestart consumes the confirmation state and, when affirmative,
// returns a tea.Cmd that performs kill+launch asynchronously so the Bubble Tea
// update loop never blocks on process teardown.
func (p *MonitorPage) finalizeConfirmRestart() tea.Cmd {
	pid := p.confirmRestartTargetID
	prof := p.confirmRestartProfile
	bg := true // default to background; look up from current instance
	insts := p.pm.List()
	for _, ri := range insts {
		if ri.PID == pid {
			bg = ri.Background
			break
		}
	}
	affirmative := p.confirmRestartAnswer != nil && *p.confirmRestartAnswer
	p.confirmRestartForm = nil
	p.confirmRestartAnswer = nil
	p.confirmRestartTargetID = 0
	p.confirmRestartProfile = domain.Profile{}
	if !affirmative {
		return nil
	}
	return restartCmd(p.pm, pid, prof, bg)
}

// restartCmd performs Kill then Launch off the UI thread and delivers the
// result wrapped in a restartResultMsg.
func restartCmd(pm procMgrIface, pid int, prof domain.Profile, bg bool) tea.Cmd {
	return func() tea.Msg {
		if err := pm.Kill(pid); err != nil {
			return restartResultMsg{pid: pid, err: fmt.Errorf("kill: %w", err)}
		}
		mode := processmgr.LaunchBackground
		if !bg {
			mode = processmgr.LaunchForeground
		}
		if _, err := pm.Launch(prof, mode); err != nil {
			return restartResultMsg{pid: pid, err: fmt.Errorf("launch: %w", err)}
		}
		return restartResultMsg{pid: pid}
	}
}

// IsCapturingInput tells the root model when the page owns global keys.
func (p *MonitorPage) IsCapturingInput() bool {
	return p.confirmKillForm != nil || p.confirmRestartForm != nil
}

func (p *MonitorPage) applyInstances(insts []domain.RunningInstance) tea.Cmd {
	rows := make([]table.Row, 0, len(insts))
	for _, ri := range insts {
		pidCol := fmt.Sprintf("%d", ri.PID)
		profileCol := ri.ProfileID
		if ri.Crashed {
			pidCol = "✗ " + pidCol
			profileCol = ri.ProfileID + " (crashed)"
		}
		uptime, vram, toks := "--", "--", "--"
		if !ri.StartedAt.IsZero() {
			uptime = humanDuration(time.Since(ri.StartedAt))
		}
		if st, ok := p.subs[ri.PID]; ok && st != nil {
			vram = formatVRAM(st.gpu.VRAMUsedMB, st.gpu.VRAMTotalMB)
			toks = formatTokensPerSec(st.mets.TokensPerSec)
		}
		portCol := fmt.Sprintf("%d", ri.Port)
		if ri.Crashed {
			pidCol = theme.Error.Render(pidCol)
			portCol = theme.Error.Render(portCol)
			profileCol = theme.Error.Render(profileCol)
			uptime = theme.Error.Render(uptime)
			vram = theme.Error.Render(vram)
			toks = theme.Error.Render(toks)
		}
		rows = append(rows, table.Row{
			pidCol,
			portCol,
			profileCol,
			uptime, vram, toks,
		})
	}
	p.tbl.SetRows(rows)
	// Bubbles' table doesn't auto-clamp the cursor when rows shrink, so
	// SelectedRow can later return an empty Row and panic on row[0].
	if cur := p.tbl.Cursor(); cur >= len(rows) {
		if len(rows) == 0 {
			p.tbl.SetCursor(0)
		} else {
			p.tbl.SetCursor(len(rows) - 1)
		}
	}

	// Add subs for new (non-crashed) instances; collect listenCmds.
	var cmds []tea.Cmd
	for _, ri := range insts {
		if ri.Crashed {
			continue
		}
		if _, ok := p.subs[ri.PID]; ok {
			continue
		}
		ch, cancel, err := p.mm.Subscribe(ri.PID, ri.Port, ri.LogPath)
		if err != nil {
			continue
		}
		p.subs[ri.PID] = &subState{cancel: cancel}
		p.chans[ri.PID] = ch
		cmds = append(cmds, listenCmd(ch))
	}
	// Cancel subs for crashed insts (data sources are dead) and for orphans
	// (PID no longer in the list at all). Async cancel preserves UI thread.
	seen := make(map[int]bool, len(insts))
	for _, ri := range insts {
		seen[ri.PID] = true
		if ri.Crashed {
			if st, ok := p.subs[ri.PID]; ok {
				cancel := st.cancel
				go func() { _ = cancel() }()
				delete(p.subs, ri.PID)
				delete(p.chans, ri.PID)
			}
		}
	}
	for pid, st := range p.subs {
		if !seen[pid] {
			cancel := st.cancel
			go func() { _ = cancel() }() // do not block UI on subscription teardown
			delete(p.subs, pid)
			delete(p.chans, pid)
		}
	}
	return tea.Batch(cmds...)
}

func (p *MonitorPage) View() string {
	if p.confirmKillForm != nil {
		return p.confirmKillForm.View()
	}
	if p.confirmRestartForm != nil {
		return p.confirmRestartForm.View()
	}
	header := lipgloss.NewStyle().Bold(true).Render("Running instances")
	if p.flash != "" {
		header = theme.Error.Render(p.flash) + "\n" + header
	}
	if len(p.tbl.Rows()) == 0 {
		return header + "\n" + theme.Subtitle.Render("(no instances running — switch to Launcher [2] to start one)")
	}
	top := header + "\n" + p.tbl.View()
	subviewTabs := renderSubViewTabs(p.subView)
	pid := p.selectedPID()
	st := p.subs[pid]
	bottom := "no subscription"
	if st != nil {
		switch p.subView {
		case SubViewLogs:
			start := len(st.logs) - 10
			if start < 0 {
				start = 0
			}
			bottom = strings.Join(st.logs[start:], "\n")
			if bottom == "" {
				bottom = "(no log lines yet)"
			}
			if p.paused {
				bottom = theme.Warn.Render("Logs (PAUSED — Space to resume)") + "\n" + bottom
			}
			if len(st.logs) > 10 {
				bottom += "\n" + theme.Subtitle.Render(fmt.Sprintf("— showing last 10 of %d (Space pauses, buffer 2000)", len(st.logs)))
			}
		case SubViewSlots:
			var b strings.Builder
			b.WriteString("idx | state      | ctx used/max | client\n")
			for _, s := range st.slots.Slots {
				fmt.Fprintf(&b, "%-3d | %-10s | %5d/%-5d | %s\n", s.ID, s.State, s.NCtxUsed, s.NCtxMax, s.Client)
			}
			bottom = b.String()
			if bottom == "idx | state      | ctx used/max | client\n" {
				bottom = "(no slot data yet)"
			}
		case SubViewMetrics:
			if len(st.mets.TokensPerSec) == 0 && len(st.mets.RequestsPerSec) == 0 {
				bottom = "(no metrics yet — first sample arrives after the slots tick)"
				break
			}
			var b strings.Builder
			fmt.Fprintf(&b, "tokens/s: %s\n", theme.OK.Render(components.Sparkline(st.mets.TokensPerSec, 40)))
			fmt.Fprintf(&b, "req/s   : %s\n", theme.Warn.Render(components.Sparkline(st.mets.RequestsPerSec, 40)))
			if st.gpu.VRAMTotalMB > 0 {
				fmt.Fprintf(&b, "VRAM    : %d/%d MB  util %.0f%%\n", st.gpu.VRAMUsedMB, st.gpu.VRAMTotalMB, st.gpu.Utilization)
			}
			bottom = b.String()
		}
	}
	return top + "\n\n" + subviewTabs + "\n" + bottom
}

// renderSubViewTabs draws the Logs / Slots / Metrics tab strip with the
// active sub-view styled via theme.TabActive. Cycled by the [v] key.
func renderSubViewTabs(active SubViewKind) string {
	render := func(k SubViewKind, label string) string {
		if k == active {
			return theme.TabActive.Render(label)
		}
		return theme.TabInactive.Render(label)
	}
	parts := []string{
		render(SubViewLogs, "Logs"),
		theme.Subtitle.Render(" │ "),
		render(SubViewSlots, "Slots"),
		theme.Subtitle.Render(" │ "),
		render(SubViewMetrics, "Metrics"),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// Hints implements ui.HintProvider for the Monitor tab.
func (p *MonitorPage) Hints() string {
	if p.confirmKillForm != nil || p.confirmRestartForm != nil {
		return "[←→] choose  [enter] confirm  [esc] cancel"
	}
	return "[v] cycle view  [Space] pause  [k] kill  [r] restart"
}

// humanDuration formats a Duration as a compact uptime string (e.g. "5s",
// "3m12s", "1h04m"). Negative or zero returns "--".
func humanDuration(d time.Duration) string {
	if d <= 0 {
		return "--"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d / time.Minute)
		s := int((d % time.Minute) / time.Second)
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)
	return fmt.Sprintf("%dh%02dm", h, m)
}

// formatVRAM renders the per-instance VRAM cell. Returns "--" when no
// totalsample yet.
func formatVRAM(usedMB, totalMB uint64) string {
	if totalMB == 0 {
		return "--"
	}
	return fmt.Sprintf("%d/%dMB", usedMB, totalMB)
}

// formatTokensPerSec returns the latest tokens-per-second sample formatted
// to one decimal place, or "--" when the metric series is empty.
func formatTokensPerSec(samples []float64) string {
	if len(samples) == 0 {
		return "--"
	}
	return fmt.Sprintf("%.1f", samples[len(samples)-1])
}

// selectRow positions the table cursor on the row matching pid (no-op if not found).
func (p *MonitorPage) selectRow(pid int) {
	rows := p.tbl.Rows()
	for i, r := range rows {
		pidCol := strings.TrimPrefix(stripANSI(r[0]), "✗ ")
		var rowPID int
		_, _ = fmt.Sscanf(pidCol, "%d", &rowPID)
		if rowPID == pid {
			p.tbl.SetCursor(i)
			return
		}
	}
}

// stripANSI removes ANSI SGR escape sequences from s. Used when parsing
// PIDs out of table rows that may have been styled (e.g. crashed rows
// rendered in theme.Error).
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			// Skip until 'm' (or end).
			i++
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// selectedPID returns the PID of the currently selected row, or 0 if no rows.
// Guarded against `table.Model.SelectedRow()` returning an empty Row when the
// cursor is out of range (e.g. rows shrunk after a row was killed) — without
// the length check, row[0] panics with index out of range.
func (p *MonitorPage) selectedPID() int {
	if len(p.tbl.Rows()) == 0 {
		return 0
	}
	row := p.tbl.SelectedRow()
	if len(row) == 0 {
		return 0
	}
	pidCol := strings.TrimPrefix(stripANSI(row[0]), "✗ ")
	var pid int
	_, _ = fmt.Sscanf(pidCol, "%d", &pid)
	return pid
}
