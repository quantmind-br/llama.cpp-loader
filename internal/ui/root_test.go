package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/pages"
)

func TestRoot_StartsOnProfilesAndQuitsOnQ(t *testing.T) {
	tm := teatest.NewTestModel(t, NewRoot(TabProfiles), teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Profiles")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if err := tm.Quit(); err != nil {
		t.Fatalf("Quit returned err: %v", err)
	}
}

func TestRoot_TabSwitchByNumber(t *testing.T) {
	tm := teatest.NewTestModel(t, NewRoot(TabProfiles), teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Monitor")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if err := tm.Quit(); err != nil {
		t.Fatalf("Quit returned err: %v", err)
	}
}

func TestRoot_UseInNewProfileSwitchesTab(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	root := NewRoot(TabModels).
		WithProfilesPage(pages.NewProfilesPage(store, domain.FlagSchema{}))

	updated, _ := root.Update(pages.UseInNewProfileMsg{Path: "/x.gguf"})
	r := updated.(RootModel)

	if r.active != TabProfiles {
		t.Errorf("active = %d, want TabProfiles=%d", r.active, TabProfiles)
	}
}

func TestRoot_TabSwitchToLauncherShowsPage(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(domain.Profile{
		ID: "alpha", Name: "AlphaProfile", Model: "/m.gguf",
		Args: map[string]any{"port": float64(8080)},
	}); err != nil {
		t.Fatal(err)
	}

	root := NewRoot(TabProfiles).
		WithLauncherPage(pages.NewLauncherPage(store, nil, nil))

	tm := teatest.NewTestModel(t, root, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "AlphaProfile")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = tm.Quit()
}

func TestRoot_WithMonitorPageReplacesPlaceholder(t *testing.T) {
	root := NewRoot(TabMonitor).WithMonitorPage(pages.Placeholder{TabName: "MONITOR_REPLACED"})

	tm := teatest.NewTestModel(t, root, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "MONITOR_REPLACED")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = tm.Quit()
}

func TestRoot_RoutesSwitchToMonitorMsg(t *testing.T) {
	root := NewRoot(TabProfiles).WithMonitorPage(pages.Placeholder{TabName: "MONITOR"})

	updated, _ := root.Update(pages.SwitchToMonitorMsg{PID: 999})
	r := updated.(RootModel)

	if r.active != TabMonitor {
		t.Fatalf("active = %d, want TabMonitor=%d", r.active, TabMonitor)
	}
}

func TestRoot_ForwardsSwitchPIDToMonitor(t *testing.T) {
	rec := &recordingMonitor{}
	r := NewRoot(TabProfiles).
		WithProfilesPage(pages.Placeholder{TabName: "P"}).
		WithLauncherPage(pages.Placeholder{TabName: "L"}).
		WithMonitorPage(rec).
		WithModelsPage(pages.Placeholder{TabName: "M"})
	updated, _ := r.Update(pages.SwitchToMonitorMsg{PID: 4321})
	rm := updated.(RootModel)
	if rm.active != TabMonitor {
		t.Errorf("active = %v, want TabMonitor", rm.active)
	}
	if rec.lastSelectPID != 4321 {
		t.Errorf("rec.lastSelectPID = %d, want 4321", rec.lastSelectPID)
	}
}

type recordingMonitor struct {
	lastSelectPID int
}

func (r *recordingMonitor) Init() tea.Cmd { return nil }
func (r *recordingMonitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m, ok := msg.(pages.MonitorSelectPIDMsg); ok {
		r.lastSelectPID = m.PID
	}
	return r, nil
}
func (r *recordingMonitor) View() string { return "" }
