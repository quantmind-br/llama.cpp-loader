package pages

import (
	"errors"
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

	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

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
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

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

func TestMonitorPage_MetricsPlaceholderWhenEmpty(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{{PID: 42, Port: 8080, LogPath: "/tmp/x.log"}}}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, nil)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})
	// Switch to metrics sub-view.
	p, _ = updateAs[*MonitorPage](p, tea.KeyMsg{Type: tea.KeyTab})
	p, _ = updateAs[*MonitorPage](p, tea.KeyMsg{Type: tea.KeyTab})
	out := p.View()
	if !strings.Contains(out, "(no metrics yet") {
		t.Fatalf("metrics view missing placeholder; got:\n%s", out)
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

func TestMonitorPage_SelectsRowByPID(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{
		{PID: 100, Port: 8080, LogPath: "/tmp/a.log"},
		{PID: 200, Port: 8081, LogPath: "/tmp/b.log"},
		{PID: 300, Port: 8082, LogPath: "/tmp/c.log"},
	}}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, nil)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	// Public handler queues a refresh; pendingSelectPID is consumed by the
	// refresh handler when it lands. We drain the single refresh cmd here.
	p, cmd := updateAs[*MonitorPage](p, MonitorSelectPIDMsg{PID: 200})
	if cmd == nil {
		t.Fatalf("MonitorSelectPIDMsg should issue refreshInstancesCmd")
	}
	p, _ = updateAs[*MonitorPage](p, cmd())

	if got := p.selectedPID(); got != 200 {
		t.Fatalf("selectedPID = %d, want 200", got)
	}
}

func TestMonitorPage_PendingSelectAppliesAfterFirstRefresh(t *testing.T) {
	// Simulates: SwitchToMonitorMsg fires before MonitorPage has any rows
	// (e.g., very first instance ever launched). pendingSelectPID must
	// wait for the refresh to land, then position the cursor.
	pm := &fakeProcMgr{insts: nil} // initially empty
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, nil)

	// Public msg arrives while rows are still empty.
	p, cmd := updateAs[*MonitorPage](p, MonitorSelectPIDMsg{PID: 200})
	if cmd == nil {
		t.Fatalf("MonitorSelectPIDMsg should issue refreshInstancesCmd")
	}
	// selectedPID is 0 right now (no rows), even though MonitorSelectPIDMsg arrived.
	if got := p.selectedPID(); got != 0 {
		t.Fatalf("selectedPID before refresh = %d, want 0", got)
	}
	// Now the registry gets the instance and the refresh cmd lands.
	pm.insts = []domain.RunningInstance{
		{PID: 100, Port: 8080, LogPath: "/tmp/a.log"},
		{PID: 200, Port: 8081, LogPath: "/tmp/b.log"},
	}
	p, _ = updateAs[*MonitorPage](p, cmd())
	if got := p.selectedPID(); got != 200 {
		t.Fatalf("selectedPID after refresh = %d, want 200 (pendingSelectPID should have been consumed)", got)
	}
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

func TestMonitorPage_RTriggersRestart(t *testing.T) {
	prof := domain.Profile{
		ID:    "qwen",
		Name:  "Qwen",
		Model: "/tmp/x.gguf",
		Args:  map[string]any{"port": 8080.0},
	}
	psk := &fakeProfileStore{p: prof}
	pm := &restartTrackingMgr{
		insts:   []domain.RunningInstance{{ProfileID: "qwen", PID: 100, Port: 8080, LogPath: "/tmp/a.log", Background: true}},
		newPID:  200,
		newPort: 8080,
	}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, psk)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	// Press 'r' on the selected (only) row.
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r did not produce a Cmd")
	}
	// Run the cmd: kill + launch should happen synchronously inside it.
	_ = cmd()
	if pm.killedPID != 100 {
		t.Errorf("killedPID = %d, want 100", pm.killedPID)
	}
	if pm.launchedID != "qwen" {
		t.Errorf("launchedID = %q, want qwen", pm.launchedID)
	}
	if pm.launchMode != processmgr.LaunchBackground {
		t.Errorf("launchMode = %v, want LaunchBackground", pm.launchMode)
	}
}

func TestMonitorPage_RTriggersRestartForeground(t *testing.T) {
	prof := domain.Profile{ID: "qwen", Name: "Qwen", Model: "/tmp/x.gguf", Args: map[string]any{"port": 8080.0}}
	psk := &fakeProfileStore{p: prof}
	pm := &restartTrackingMgr{
		// Background: false (zero-value) → restartCmd should pick LaunchForeground.
		insts:   []domain.RunningInstance{{ProfileID: "qwen", PID: 100, Port: 8080, LogPath: "/tmp/a.log"}},
		newPID:  200,
		newPort: 8080,
	}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, psk)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r did not produce a Cmd")
	}
	_ = cmd()
	if pm.launchMode != processmgr.LaunchForeground {
		t.Errorf("launchMode = %v, want LaunchForeground", pm.launchMode)
	}
}

func TestMonitorPage_RFallsBackToKillOnlyWhenStoreNil(t *testing.T) {
	pm := &restartTrackingMgr{
		insts:   []domain.RunningInstance{{ProfileID: "qwen", PID: 100, Port: 8080, LogPath: "/tmp/a.log", Background: true}},
		newPID:  200,
		newPort: 8080,
	}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, nil) // store nil
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r did not produce a Cmd")
	}
	_ = cmd()
	if pm.killedPID != 100 {
		t.Errorf("killedPID = %d, want 100", pm.killedPID)
	}
	if pm.launchedID != "" {
		t.Errorf("launchedID = %q, want empty (no Launch should run with nil store)", pm.launchedID)
	}
}

func TestMonitorPage_RFallsBackWhenProfileGetErrors(t *testing.T) {
	psk := &fakeProfileStore{err: errors.New("profile vanished")}
	pm := &restartTrackingMgr{
		insts:   []domain.RunningInstance{{ProfileID: "qwen", PID: 100, Port: 8080, LogPath: "/tmp/a.log", Background: true}},
		newPID:  200,
		newPort: 8080,
	}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, psk)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r did not produce a Cmd")
	}
	_ = cmd()
	if pm.killedPID != 100 {
		t.Errorf("killedPID = %d, want 100", pm.killedPID)
	}
	if pm.launchedID != "" {
		t.Errorf("launchedID = %q, want empty (no Launch should run when Get errors)", pm.launchedID)
	}
}

type fakeProfileStore struct {
	p   domain.Profile
	err error
}

func (f *fakeProfileStore) Get(id string) (domain.Profile, error) {
	if f.err != nil {
		return domain.Profile{}, f.err
	}
	return f.p, nil
}

type restartTrackingMgr struct {
	insts      []domain.RunningInstance
	killedPID  int
	launchedID string
	launchMode processmgr.LaunchMode
	newPID     int
	newPort    int
}

func (r *restartTrackingMgr) List() []domain.RunningInstance { return r.insts }
func (r *restartTrackingMgr) Kill(pid int) error             { r.killedPID = pid; return nil }
func (r *restartTrackingMgr) TailLogs(_ int) (io.ReadCloser, error) {
	return nil, processmgr.ErrUnknownPID
}
func (r *restartTrackingMgr) Launch(p domain.Profile, mode processmgr.LaunchMode) (domain.RunningInstance, error) {
	r.launchedID = p.ID
	r.launchMode = mode
	return domain.RunningInstance{ProfileID: p.ID, PID: r.newPID, Port: r.newPort, Background: true}, nil
}

func TestMonitorPage_CrashedRowShowsMarker(t *testing.T) {
	exit := time.Now().UTC()
	pm := &fakeProcMgr{insts: []domain.RunningInstance{{
		PID:       777,
		Port:      8080,
		ProfileID: "qwen",
		LogPath:   "/tmp/x.log",
		Crashed:   true,
		ExitedAt:  &exit,
	}}}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, nil)
	p, _ = updateAs[*MonitorPage](p, monitorInstancesRefreshedMsg{insts: pm.List()})
	out := p.View()
	if !strings.Contains(out, "✗") && !strings.Contains(out, "crashed") {
		t.Fatalf("crashed row missing badge; got:\n%s", out)
	}
}

func TestMonitorPage_PeriodicRefreshTickEmitsRefreshCmd(t *testing.T) {
	pm := &fakeProcMgr{}
	mm := &fakeMonMgr{}
	p := NewMonitorPage(pm, mm, nil)
	cmd := p.Init()
	if cmd == nil {
		t.Fatal("Init returned nil")
	}
	if !p.periodicTickActive {
		t.Errorf("periodicTickActive = false; want true")
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
