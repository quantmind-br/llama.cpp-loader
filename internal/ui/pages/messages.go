package pages

// SwitchToMonitorMsg is emitted by LauncherPage after a launch is healthy.
// root.go consumes this to switch the active tab to Monitor and pre-select
// the new PID.
type SwitchToMonitorMsg struct {
	PID int
}
