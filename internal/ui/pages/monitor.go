package pages

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/monitor"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/components"
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

type MonitorPage struct {
	pm               procMgrIface
	mm               monitor.Manager
	ps               profileStoreIface // injected for `r` real restart (slice 6 / Task 4)
	pendingSelectPID int               // set by MonitorSelectPIDMsg, consumed after the next refresh
	tbl              table.Model
	subs             map[int]*subState
	chans            map[int]<-chan monitor.MonitorEvent
	subView          SubViewKind
	paused           bool
	width            int
	height           int
}

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

func (p *MonitorPage) Init() tea.Cmd { return p.refreshInstancesCmd() }

func (p *MonitorPage) refreshInstancesCmd() tea.Cmd {
	return func() tea.Msg { return monitorInstancesRefreshedMsg{insts: p.pm.List()} }
}

type monitorInstancesRefreshedMsg struct {
	insts []domain.RunningInstance
}

// restartCmd kills the instance with pid and relaunches the same profile in
// background mode. Falls back to plain kill when ps is nil (no store wired).
// On any error, the command emits a no-op msg; UI surfaces the error via the
// next instance refresh.
func (p *MonitorPage) restartCmd(pid int) tea.Cmd {
	insts := p.pm.List()
	var inst *domain.RunningInstance
	for i := range insts {
		if insts[i].PID == pid {
			inst = &insts[i]
			break
		}
	}
	if inst == nil {
		return p.refreshInstancesCmd()
	}
	pm := p.pm
	ps := p.ps
	profileID := inst.ProfileID
	bg := inst.Background
	return tea.Batch(
		func() tea.Msg {
			_ = pm.Kill(pid)
			if ps == nil {
				return monitorInstancesRefreshedMsg{insts: pm.List()}
			}
			prof, err := ps.Get(profileID)
			if err != nil {
				return monitorInstancesRefreshedMsg{insts: pm.List()}
			}
			mode := processmgr.LaunchBackground
			if !bg {
				mode = processmgr.LaunchForeground
			}
			_, _ = pm.Launch(prof, mode)
			return monitorInstancesRefreshedMsg{insts: pm.List()}
		},
	)
}

func (p *MonitorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		p.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch {
		case m.Type == tea.KeyTab:
			p.subView = (p.subView + 1) % 3
		case m.Type == tea.KeyRunes && len(m.Runes) == 1 && m.Runes[0] == 'k':
			if pid := p.selectedPID(); pid > 0 {
				_ = p.pm.Kill(pid)
				cmds = append(cmds, p.refreshInstancesCmd())
			}
		case m.Type == tea.KeyRunes && len(m.Runes) == 1 && m.Runes[0] == 'r':
			if pid := p.selectedPID(); pid > 0 {
				cmds = append(cmds, p.restartCmd(pid))
			}
		case m.Type == tea.KeySpace:
			p.paused = !p.paused
		}
	case monitorInstancesRefreshedMsg:
		if c := p.applyInstances(m.insts); c != nil {
			cmds = append(cmds, c)
		}
		if p.pendingSelectPID != 0 {
			p.selectRow(p.pendingSelectPID)
			p.pendingSelectPID = 0
		}
	case MonitorSelectPIDMsg:
		// Defer the cursor move until the next refresh has applied fresh rows,
		// so we never select against a stale row list.
		p.pendingSelectPID = m.PID
		cmds = append(cmds, p.refreshInstancesCmd())
	case monitorEventMsg:
		st, ok := p.subs[m.ev.PID]
		if ok {
			switch m.ev.Source {
			case monitor.SourceLogs:
				if line, ok := m.ev.Data.(monitor.LogLine); ok {
					if !p.paused {
						st.logs = append(st.logs, line.Line)
						if len(st.logs) > 2000 {
							st.logs = st.logs[len(st.logs)-2000:]
						}
					}
				}
			case monitor.SourceSlots:
				if s, ok := m.ev.Data.(monitor.SlotSnapshot); ok {
					st.slots = s
				}
			case monitor.SourceGPU:
				if g, ok := m.ev.Data.(monitor.GPUStats); ok {
					st.gpu = g
				}
			case monitor.SourceHealth:
				if h, ok := m.ev.Data.(monitor.HealthStatus); ok {
					st.health = h
				}
			case monitor.SourceMetrics:
				if mts, ok := m.ev.Data.(monitor.Metrics); ok {
					st.mets = mts
				}
			}
		}
		// Re-arm listener for this PID.
		if ch, ok := p.chans[m.ev.PID]; ok {
			cmds = append(cmds, listenCmd(ch))
		}
	}
	t, tc := p.tbl.Update(msg)
	p.tbl = t
	if tc != nil {
		cmds = append(cmds, tc)
	}
	return p, tea.Batch(cmds...)
}

func (p *MonitorPage) applyInstances(insts []domain.RunningInstance) tea.Cmd {
	rows := make([]table.Row, 0, len(insts))
	for _, ri := range insts {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", ri.PID),
			fmt.Sprintf("%d", ri.Port),
			ri.ProfileID,
			"--", "--", "--",
		})
	}
	p.tbl.SetRows(rows)

	// Add subs for new instances; collect listenCmds.
	var cmds []tea.Cmd
	for _, ri := range insts {
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
	// Cancel orphan subs.
	seen := make(map[int]bool, len(insts))
	for _, ri := range insts {
		seen[ri.PID] = true
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
	header := lipgloss.NewStyle().Bold(true).Render("Running instances")
	if len(p.tbl.Rows()) == 0 {
		return header + "\n  (none)"
	}
	top := header + "\n" + p.tbl.View()
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
			fmt.Fprintf(&b, "tokens/s: %s\n", components.Sparkline(st.mets.TokensPerSec, 40))
			fmt.Fprintf(&b, "req/s   : %s\n", components.Sparkline(st.mets.RequestsPerSec, 40))
			if st.gpu.VRAMTotalMB > 0 {
				fmt.Fprintf(&b, "VRAM    : %d/%d MB  util %.0f%%\n", st.gpu.VRAMUsedMB, st.gpu.VRAMTotalMB, st.gpu.Utilization)
			}
			bottom = b.String()
		}
	}
	return top + "\n\n" + bottom
}

// selectRow positions the table cursor on the row matching pid (no-op if not found).
func (p *MonitorPage) selectRow(pid int) {
	rows := p.tbl.Rows()
	for i, r := range rows {
		var rowPID int
		_, _ = fmt.Sscanf(r[0], "%d", &rowPID)
		if rowPID == pid {
			p.tbl.SetCursor(i)
			return
		}
	}
}

// selectedPID returns the PID of the currently selected row, or 0 if no rows.
func (p *MonitorPage) selectedPID() int {
	if len(p.tbl.Rows()) == 0 {
		return 0
	}
	row := p.tbl.SelectedRow()
	var pid int
	_, _ = fmt.Sscanf(row[0], "%d", &pid)
	return pid
}
