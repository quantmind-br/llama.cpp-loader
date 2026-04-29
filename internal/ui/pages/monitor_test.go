package pages

import (
	"io"
	"strings"
	"sync/atomic"
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
	p := NewMonitorPage(pm, mm, nil)
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
	p := NewMonitorPage(pm, mm, nil)
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

type slowCancelMonMgr struct {
	cancel func() error
}

func (s *slowCancelMonMgr) Subscribe(_, _ int, _ string) (<-chan monitor.MonitorEvent, func() error, error) {
	ch := make(chan monitor.MonitorEvent)
	return ch, s.cancel, nil
}

func updateAs[T tea.Model](p tea.Model, msg tea.Msg) (T, tea.Cmd) {
	out, cmd := p.Update(msg)
	return out.(T), cmd
}

func TestMonitorPage_TabCyclesToSlots(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{{PID: 1, Port: 8080, LogPath: "/tmp/x.log"}}}
	mm := &chanMonMgr{ch: make(chan monitor.MonitorEvent, 8)}
	p := NewMonitorPage(pm, mm, nil)
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

func TestMonitorPage_MetricsViewRendersSparkline(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{{PID: 1, Port: 8080, LogPath: "/tmp/x.log"}}}
	mm := &chanMonMgr{ch: make(chan monitor.MonitorEvent, 8)}
	p := NewMonitorPage(pm, mm, nil)
	p.SetSize(120, 30)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})
	p, _ = updateAs[*MonitorPage](p, monitorEventMsg{ev: monitor.MonitorEvent{
		Source: monitor.SourceMetrics, PID: 1,
		Data: monitor.Metrics{
			TokensPerSec:   []float64{10, 20, 30, 40, 50},
			RequestsPerSec: []float64{0, 0, 1, 2, 3},
			WindowSeconds:  60,
		},
	}})
	// Tab twice -> Metrics.
	p, _ = updateAs[*MonitorPage](p, tea.KeyMsg{Type: tea.KeyTab})
	p, _ = updateAs[*MonitorPage](p, tea.KeyMsg{Type: tea.KeyTab})

	v := p.View()
	if !strings.Contains(v, "tokens/s") {
		t.Fatalf("metrics view missing 'tokens/s':\n%s", v)
	}
	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	found := false
	for _, b := range bars {
		if strings.ContainsRune(v, b) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("metrics view has no sparkline bar")
	}
}

func TestMonitorPage_KKillsSelectedPID(t *testing.T) {
	pm := &killTrackingMgr{fakeProcMgr: fakeProcMgr{insts: []domain.RunningInstance{{PID: 7, Port: 8080, LogPath: "/tmp/x.log"}}}}
	mm := &chanMonMgr{ch: make(chan monitor.MonitorEvent, 8)}
	p := NewMonitorPage(pm, mm, nil)
	p.SetSize(120, 30)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	p, _ = updateAs[*MonitorPage](p, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	if pm.killed != 7 {
		t.Fatalf("expected Kill(7), got Kill(%d)", pm.killed)
	}
}

type killTrackingMgr struct {
	fakeProcMgr
	killed int
}

func (m *killTrackingMgr) Kill(pid int) error { m.killed = pid; return nil }

func TestMonitorPage_KKillCancelsOrphanSub(t *testing.T) {
	var cancelCalled int32
	mm := &countingMonMgr{
		ch:       make(chan monitor.MonitorEvent, 8),
		onCancel: func() { atomic.AddInt32(&cancelCalled, 1) },
	}
	pm := &killTrackingMgr{fakeProcMgr: fakeProcMgr{insts: []domain.RunningInstance{{PID: 7, Port: 8080, LogPath: "/tmp/x.log"}}}}
	p := NewMonitorPage(pm, mm, nil)
	p.SetSize(120, 30)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	// Press k. Expect kill + a refresh cmd.
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if cmd == nil {
		t.Fatal("expected refresh cmd after k, got nil")
	}

	// Simulate the manager removing the killed pid (i.e., List() now empty).
	pm.fakeProcMgr.insts = nil
	// Pump the refresh msg back through Update.
	if msg := cmd(); msg != nil {
		// cmd from k might be tea.Batch — drill in if needed.
		switch tm := msg.(type) {
		case monitorInstancesRefreshedMsg:
			p, _ = updateAs[*MonitorPage](p, tm)
		default:
			// tea.Batch returns a BatchMsg; for this test we manually
			// fire the refresh with the new instance list.
			p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})
		}
	}

	// Cancel runs asynchronously off the UI thread; poll briefly.
	for i := 0; i < 100 && atomic.LoadInt32(&cancelCalled) == 0; i++ {
		time.Sleep(2 * time.Millisecond)
	}
	if got := atomic.LoadInt32(&cancelCalled); got != 1 {
		t.Fatalf("cancel called %d times, want 1 (orphan sub not cancelled)", got)
	}
	if _, ok := p.subs[7]; ok {
		t.Fatal("orphan sub for pid 7 not deleted")
	}
}

type countingMonMgr struct {
	ch       chan monitor.MonitorEvent
	onCancel func()
}

func (m *countingMonMgr) Subscribe(pid, port int, logPath string) (<-chan monitor.MonitorEvent, func() error, error) {
	return m.ch, func() error {
		if m.onCancel != nil {
			m.onCancel()
		}
		return nil
	}, nil
}

func TestMonitorPage_HandlesWindowSize(t *testing.T) {
	pm := &fakeProcMgr{}
	mm := fakeMonMgr{}
	p := NewMonitorPage(pm, mm, nil)
	p.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if p.width != 80 || p.height != 24 {
		t.Fatalf("after WindowSizeMsg: width=%d height=%d, want 80x24", p.width, p.height)
	}
}

func TestMonitorPage_CancelOrphanIsAsync(t *testing.T) {
	// Cancel func will block until release is signaled — simulating a
	// stuck nvidia-smi that takes time to reap. The test asserts that
	// applyInstances returns promptly without blocking on cancel.
	release := make(chan struct{})
	var cancelStarted, cancelFinished int32
	mm := &slowCancelMonMgr{
		cancel: func() error {
			atomic.AddInt32(&cancelStarted, 1)
			<-release
			atomic.AddInt32(&cancelFinished, 1)
			return nil
		},
	}
	pm := &fakeProcMgr{insts: []domain.RunningInstance{{PID: 1, Port: 8080, LogPath: "/tmp/x.log"}}}
	p := NewMonitorPage(pm, mm, nil)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})
	pm.insts = nil // PID 1 vanished -> cancel should be triggered

	done := make(chan struct{})
	go func() {
		p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})
		close(done)
	}()
	select {
	case <-done:
		// applyInstances returned promptly even though cancel still hasn't completed.
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("applyInstances blocked on cancel; cancelStarted=%d", atomic.LoadInt32(&cancelStarted))
	}
	// Now release the cancel.
	close(release)
	for i := 0; i < 100 && atomic.LoadInt32(&cancelFinished) == 0; i++ {
		time.Sleep(2 * time.Millisecond)
	}
	if atomic.LoadInt32(&cancelFinished) != 1 {
		t.Fatalf("cancel never completed after release; finished=%d", atomic.LoadInt32(&cancelFinished))
	}
}
