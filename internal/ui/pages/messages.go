package pages

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SwitchToMonitorMsg is emitted by LauncherPage after a launch is healthy.
// root.go consumes this to switch the active tab to Monitor and pre-select
// the new PID.
type SwitchToMonitorMsg struct {
	PID int
}

// MonitorSelectPIDMsg instructs MonitorPage to select the row whose PID
// matches. Sent by root when handling SwitchToMonitorMsg so the Monitor
// page lands focused on the newly-launched instance.
type MonitorSelectPIDMsg struct {
	PID int
}

// flashClearMsg is emitted by setFlashCmd 15s after a flash was set.
// `tag` identifies the owning page so other pages ignore it; `at` is the
// stamp the flash carried at scheduling so a newer flash on the same page
// can detect this clear is stale and skip.
type flashClearMsg struct {
	tag string
	at  time.Time
}

const (
	flashLifetime = 15 * time.Second
	flashDimAfter = 5 * time.Second
)

// scheduleFlashClear returns a tea.Cmd that delivers a flashClearMsg with
// the given tag + stamp once flashLifetime elapses.
func scheduleFlashClear(tag string, at time.Time) tea.Cmd {
	return tea.Tick(flashLifetime, func(time.Time) tea.Msg {
		return flashClearMsg{tag: tag, at: at}
	})
}
