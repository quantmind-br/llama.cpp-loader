package pages

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/monitor"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
)

// procMgrIface is the slice of processmgr.Manager that MonitorPage needs.
type procMgrIface interface {
	List() []domain.RunningInstance
	Kill(pid int) error
	Launch(domain.Profile, processmgr.LaunchMode) (domain.RunningInstance, error)
	TailLogs(pid int) (io.ReadCloser, error)
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
	pm     procMgrIface
	mm     monitor.Manager
	tbl    table.Model
	subs   map[int]*subState
	width  int
	height int
}

func NewMonitorPage(pm procMgrIface, mm monitor.Manager) *MonitorPage {
	cols := []table.Column{
		{Title: "PID", Width: 8},
		{Title: "Port", Width: 6},
		{Title: "Profile", Width: 18},
		{Title: "Uptime", Width: 10},
		{Title: "VRAM", Width: 12},
		{Title: "Tokens/s", Width: 10},
	}
	t := table.New(table.WithColumns(cols), table.WithFocused(true), table.WithHeight(8))
	return &MonitorPage{pm: pm, mm: mm, tbl: t, subs: map[int]*subState{}}
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

func (p *MonitorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case monitorInstancesRefreshedMsg:
		p.applyInstances(m.insts)
	}
	t, cmd := p.tbl.Update(msg)
	p.tbl = t
	return p, cmd
}

func (p *MonitorPage) applyInstances(insts []domain.RunningInstance) {
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
}

func (p *MonitorPage) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("Running instances")
	if len(p.tbl.Rows()) == 0 {
		return header + "\n  (none)"
	}
	return header + "\n" + p.tbl.View()
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
