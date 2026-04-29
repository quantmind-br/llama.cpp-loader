package pages

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/monitor"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
)

type fakeProcMgr struct {
	insts []domain.RunningInstance
}

func (f *fakeProcMgr) Launch(p domain.Profile, m processmgr.LaunchMode) (domain.RunningInstance, error) {
	return domain.RunningInstance{}, nil
}
func (f *fakeProcMgr) Kill(pid int) error                               { return nil }
func (f *fakeProcMgr) List() []domain.RunningInstance                   { return f.insts }
func (f *fakeProcMgr) WaitHealthy(pid, port int, t time.Duration) error { return nil }
func (f *fakeProcMgr) TailLogs(pid int) (io.ReadCloser, error)          { return nil, nil }

type fakeMonMgr struct{}

func (fakeMonMgr) Subscribe(pid, port int, logPath string) (<-chan monitor.MonitorEvent, func() error, error) {
	ch := make(chan monitor.MonitorEvent)
	return ch, func() error { close(ch); return nil }, nil
}

func TestMonitorPage_RendersInstanceRows(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{
		{PID: 1234, Port: 8080, ProfileID: "p1", LogPath: "/tmp/x.log"},
		{PID: 5678, Port: 8081, ProfileID: "p2", LogPath: "/tmp/y.log"},
	}}
	mm := fakeMonMgr{}
	p := NewMonitorPage(pm, mm)
	p.SetSize(120, 30)

	if cmd := p.Init(); cmd != nil {
		if msg := cmd(); msg != nil {
			p.Update(msg)
		}
	}

	view := p.View()
	if !strings.Contains(view, "1234") {
		t.Fatalf("view missing pid 1234:\n%s", view)
	}
	if !strings.Contains(view, "5678") {
		t.Fatalf("view missing pid 5678:\n%s", view)
	}
}

func TestMonitorPage_LogsSubViewShowsLines(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{{PID: 1, Port: 8080, LogPath: "/tmp/x.log"}}}

	mm := &chanMonMgr{ch: make(chan monitor.MonitorEvent, 8)}
	p := NewMonitorPage(pm, mm)
	p.SetSize(120, 30)
	if cmd := p.Init(); cmd != nil {
		if msg := cmd(); msg != nil {
			p.Update(msg)
		}
	}

	mm.ch <- monitor.MonitorEvent{Source: monitor.SourceLogs, PID: 1, Data: monitor.LogLine{Line: "boot complete"}}
	p, _ = updateAs[*MonitorPage](p, monitorEventMsg{ev: <-mm.ch})

	v := p.View()
	if !strings.Contains(v, "boot complete") {
		t.Fatalf("logs view missing 'boot complete':\n%s", v)
	}
}

type chanMonMgr struct{ ch chan monitor.MonitorEvent }

func (m *chanMonMgr) Subscribe(pid, port int, logPath string) (<-chan monitor.MonitorEvent, func() error, error) {
	return m.ch, func() error { return nil }, nil
}

func updateAs[T tea.Model](p tea.Model, msg tea.Msg) (T, tea.Cmd) {
	out, cmd := p.Update(msg)
	return out.(T), cmd
}

func TestMonitorPage_TabCyclesToSlots(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{{PID: 1, Port: 8080, LogPath: "/tmp/x.log"}}}
	mm := &chanMonMgr{ch: make(chan monitor.MonitorEvent, 8)}
	p := NewMonitorPage(pm, mm)
	p.SetSize(120, 30)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	// Inject a slot snapshot.
	p, _ = updateAs[*MonitorPage](p, monitorEventMsg{ev: monitor.MonitorEvent{
		Source: monitor.SourceSlots, PID: 1,
		Data: monitor.SlotSnapshot{Slots: []monitor.Slot{{ID: 0, State: "idle", NCtxMax: 4096}}},
	}})

	// Press Tab -> sub-view becomes Slots.
	p, _ = updateAs[*MonitorPage](p, tea.KeyMsg{Type: tea.KeyTab})

	v := p.View()
	if !strings.Contains(v, "idle") {
		t.Fatalf("after Tab, slots view missing 'idle':\n%s", v)
	}
}
