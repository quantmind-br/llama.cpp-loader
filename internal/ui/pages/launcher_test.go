package pages

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/validator"
)

func TestLauncherPage_ListsProfiles(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(domain.Profile{
		ID: "qwen", Name: "Qwen Coder", Model: "/m.gguf",
		Args: map[string]any{"port": float64(8080)},
	}); err != nil {
		t.Fatal(err)
	}

	page := NewLauncherPage(store, nil, nil) // no manager / no validator yet
	tm := teatest.NewTestModel(t, page, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Qwen Coder")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = tm.Quit()
}

type fakeManager struct {
	launched []domain.Profile
	mode     processmgr.LaunchMode
	nextErr  error
}

func (f *fakeManager) Launch(p domain.Profile, mode processmgr.LaunchMode) (domain.RunningInstance, error) {
	if f.nextErr != nil {
		err := f.nextErr
		f.nextErr = nil
		return domain.RunningInstance{}, err
	}
	f.launched = append(f.launched, p)
	f.mode = mode
	return domain.RunningInstance{ProfileID: p.ID, PID: 4242, Port: 8080, Background: mode == processmgr.LaunchBackground}, nil
}
func (f *fakeManager) Kill(pid int) error                          { return nil }
func (f *fakeManager) List() []domain.RunningInstance              { return nil }
func (f *fakeManager) WaitHealthy(_, _ int, _ time.Duration) error { return nil }
func (f *fakeManager) TailLogs(_ int) (io.ReadCloser, error)       { return nil, processmgr.ErrUnknownPID }

func TestLauncherPage_EnterLaunchesSelected(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	_ = store.Save(domain.Profile{
		ID: "alpha", Name: "Alpha", Model: "/m.gguf",
		Args: map[string]any{"port": float64(8080)},
	})

	mgr := &fakeManager{}
	page := NewLauncherPage(store, mgr, nil)

	model, _ := page.Update(LauncherProfilesLoadedMsg{Profiles: []domain.Profile{{
		ID: "alpha", Name: "Alpha", Model: "/m.gguf",
		Args: map[string]any{"port": float64(8080)},
	}}})
	page = model.(LauncherPage)

	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(LauncherPage)

	if cmd == nil {
		t.Fatal("Enter did not produce a tea.Cmd")
	}
	msg := cmd()
	switch m := msg.(type) {
	case launchedMsg:
		if m.inst.ProfileID != "alpha" {
			t.Errorf("launched ProfileID = %s, want alpha", m.inst.ProfileID)
		}
	case launchErrMsg:
		t.Fatalf("got launchErrMsg: %v", m.err)
	default:
		t.Fatalf("unexpected msg type: %T", msg)
	}

	if len(mgr.launched) != 1 {
		t.Errorf("manager.launched len = %d, want 1", len(mgr.launched))
	}
	if mgr.mode != processmgr.LaunchBackground {
		t.Errorf("mode = %v, want LaunchBackground", mgr.mode)
	}
}

func TestLauncherPage_ValidationBlocksLaunch(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)

	bad := domain.Profile{
		ID: "bad", Name: "Bad", Model: "/m.gguf",
		Args: map[string]any{
			"port":        float64(8080),
			"batch-size":  float64(1024),
			"ubatch-size": float64(2048), // > batch-size -> error
		},
	}
	mgr := &fakeManager{}
	page := NewLauncherPage(store, mgr, validator.New())
	model, _ := page.Update(LauncherProfilesLoadedMsg{Profiles: []domain.Profile{bad}})
	page = model.(LauncherPage)

	_, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a cmd (with launchErrMsg), got nil")
	}
	msg := cmd()
	if _, ok := msg.(launchErrMsg); !ok {
		t.Fatalf("expected launchErrMsg from validation, got %T", msg)
	}
	if len(mgr.launched) != 0 {
		t.Errorf("manager.launched len = %d, want 0 (validation should have blocked)", len(mgr.launched))
	}
}

func TestLauncherPage_KillRemovesInstance(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	_ = store.Save(domain.Profile{
		ID: "alpha", Name: "Alpha", Model: "/m.gguf",
		Args: map[string]any{"port": float64(8080)},
	})

	mgr := &fakeManager{}
	page := NewLauncherPage(store, mgr, nil)

	// Inject one running instance via launchedMsg.
	model, _ := page.Update(launchedMsg{inst: domain.RunningInstance{ProfileID: "alpha", PID: 4242, Port: 8080, Background: true}})
	page = model.(LauncherPage)
	if len(page.running) != 1 {
		t.Fatalf("running len = %d, want 1", len(page.running))
	}

	// Press 'k' — should call mgr.Kill(4242) and drop from page.running.
	updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	page = updated.(LauncherPage)
	if len(page.running) != 0 {
		t.Errorf("running len after kill = %d, want 0", len(page.running))
	}
}

func TestLauncherPage_RefreshReloadsProfiles(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewLauncherPage(store, nil, nil)

	// 0 profiles initially.
	model, _ := page.Update(LauncherProfilesLoadedMsg{Profiles: nil})
	page = model.(LauncherPage)
	if len(page.profiles) != 0 {
		t.Fatalf("initial profiles = %d, want 0", len(page.profiles))
	}

	// Add one to disk.
	if err := store.Save(domain.Profile{ID: "x", Name: "X", Model: "/m.gguf", Args: map[string]any{"port": float64(8080)}}); err != nil {
		t.Fatal(err)
	}

	// Press 'r' -> cmd that yields a fresh launcherProfilesLoadedMsg.
	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	page = updated.(LauncherPage)
	if cmd == nil {
		t.Fatal("'r' did not produce reload cmd")
	}
	msg := cmd()
	loaded, ok := msg.(LauncherProfilesLoadedMsg)
	if !ok {
		t.Fatalf("got %T, want LauncherProfilesLoadedMsg", msg)
	}
	if len(loaded.Profiles) != 1 || loaded.Profiles[0].ID != "x" {
		t.Errorf("reloaded profiles = %v", loaded.Profiles)
	}
}

func TestLauncherPage_PortBusyHint(t *testing.T) {
	page := NewLauncherPage(nil, nil, nil)
	wrapped := fmt.Errorf("port 8080: %w", processmgr.ErrPortBusy)
	updated, _ := page.Update(launchErrMsg{err: wrapped})
	out := updated.(LauncherPage).View()
	if !strings.Contains(out, "port") || !strings.Contains(out, "in use") {
		t.Fatalf("view missing port-busy hint; got:\n%s", out)
	}
}

func TestLauncherPage_ModelMissingHint(t *testing.T) {
	page := NewLauncherPage(nil, nil, nil)
	wrapped := fmt.Errorf("foo: %w", processmgr.ErrModelNotFound)
	updated, _ := page.Update(launchErrMsg{err: wrapped})
	out := updated.(LauncherPage).View()
	if !strings.Contains(out, "model file not found") {
		t.Fatalf("view missing model-not-found hint; got:\n%s", out)
	}
}

func TestLauncherPage_ForegroundBusyHint(t *testing.T) {
	page := NewLauncherPage(nil, nil, nil)
	updated, _ := page.Update(launchErrMsg{err: processmgr.ErrForegroundBusy})
	out := updated.(LauncherPage).View()
	if !strings.Contains(out, "foreground") || !strings.Contains(out, "[b]") {
		t.Fatalf("view missing foreground-busy hint; got:\n%s", out)
	}
}

func TestLauncherPage_HealthCheckTimeoutHint(t *testing.T) {
	page := NewLauncherPage(nil, nil, nil)
	wrapped := fmt.Errorf("pid 1: %w", processmgr.ErrHealthCheckTimeout)
	updated, _ := page.Update(launchErrMsg{err: wrapped})
	out := updated.(LauncherPage).View()
	if !strings.Contains(out, "healthy") || !strings.Contains(out, "check logs") {
		t.Fatalf("view missing health-timeout hint; got:\n%s", out)
	}
}

func TestLauncherPage_HealthyEmitsSwitchToMonitor(t *testing.T) {
	page := LauncherPage{}
	model, cmd := page.Update(healthyMsg{pid: 4242})
	if cmd == nil {
		t.Fatal("expected SwitchToMonitorMsg cmd")
	}
	msg := cmd()
	sw, ok := msg.(SwitchToMonitorMsg)
	if !ok {
		t.Fatalf("msg = %T, want SwitchToMonitorMsg", msg)
	}
	if sw.PID != 4242 {
		t.Fatalf("SwitchToMonitorMsg.PID = %d, want 4242", sw.PID)
	}
	_ = model
}
