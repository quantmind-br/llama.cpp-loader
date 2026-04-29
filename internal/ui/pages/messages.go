package pages

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
