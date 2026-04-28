package domain

import "time"

// RunningInstance describes a live llama-server process tracked by ProcessManager.
type RunningInstance struct {
	ProfileID  string    `json:"profileId"`
	PID        int       `json:"pid"`
	Port       int       `json:"port"`
	LogPath    string    `json:"logPath"`
	StartedAt  time.Time `json:"startedAt"`
	Background bool      `json:"background"`
}

// LogLine is a single line of llama-server output.
type LogLine struct {
	Timestamp time.Time
	Level     string // INFO | WARN | ERROR | "" if unparseable
	Text      string
}
