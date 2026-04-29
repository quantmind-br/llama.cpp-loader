// Package monitor streams logs, slot snapshots, GPU stats, and aggregated
// metrics from a running llama-server instance.
package monitor

import (
	"errors"
	"io"
	"net/http"
	"time"
)

// EventSource identifies which producer emitted a MonitorEvent.
type EventSource int

const (
	SourceLogs EventSource = iota
	SourceSlots
	SourceHealth
	SourceGPU
	SourceMetrics
)

// MonitorEvent is the unit pushed on the channel returned by Subscribe.
// Data type varies per Source — consumers type-assert.
type MonitorEvent struct {
	Timestamp time.Time
	Source    EventSource
	PID       int
	Data      any
}

// LogLine wraps one line tailed from the log file.
type LogLine struct {
	Line string
}

// SlotSnapshot is the decoded payload of GET /slots.
type SlotSnapshot struct {
	Slots []Slot
}

type Slot struct {
	ID         int    `json:"id"`
	State      string `json:"state"`         // "idle" | "processing"
	NCtxUsed   int    `json:"n_ctx"`         // current ctx tokens
	NCtxMax    int    `json:"n_ctx_total"`   // ctx capacity
	NDecoded   int    `json:"n_decoded"`     // cumulative tokens decoded
	NPrompt    int    `json:"n_prompt"`      // cumulative prompt tokens
	Client     string `json:"id_task"`       // best-effort client id
}

// HealthStatus is the decoded payload of GET /health.
type HealthStatus struct {
	OK     bool
	Status string
}

// GPUStats is one snapshot of GPU usage for the process pid.
type GPUStats struct {
	VRAMUsedMB  uint64
	VRAMTotalMB uint64
	Utilization float64 // 0..100
	Source      string  // "gopsutil" | "nvidia-smi"
}

// Metrics is the rolling 60s aggregate.
type Metrics struct {
	TokensPerSec   []float64 // last 60 samples (1/s)
	RequestsPerSec []float64
	WindowSeconds  int
}

// Manager is the public service interface.
type Manager interface {
	// Subscribe starts a goroutine that emits events for the given pid/port.
	// cancel() releases all goroutines and closes the returned channel.
	Subscribe(pid int, port int, logPath string) (<-chan MonitorEvent, func() error, error)
}

// Config tunes pollers and buffers.
type Config struct {
	HTTPClient        HTTPDoer      // pluggable for tests
	NvidiaSMIPath     string        // override binary path; "" = look up "nvidia-smi" in PATH
	SlotsTickInterval time.Duration // default 1s
	GPUTickInterval   time.Duration // default 2s
	LogRingSize       int           // default 2000
	MetricsWindow     time.Duration // default 60s
}

// HTTPDoer abstracts http.Client for tests.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Sentinel errors. UI maps these to status bar messages.
var (
	ErrUnknownPID    = errors.New("pid is not tracked")
	ErrLogPathEmpty  = errors.New("log path is empty")
	ErrSubscribeBusy = errors.New("subscription already active for pid")
)

// New returns a Manager backed by the default in-process implementation.
func New(cfg Config) Manager {
	if cfg.SlotsTickInterval == 0 {
		cfg.SlotsTickInterval = time.Second
	}
	if cfg.GPUTickInterval == 0 {
		cfg.GPUTickInterval = 2 * time.Second
	}
	if cfg.LogRingSize == 0 {
		cfg.LogRingSize = 2000
	}
	if cfg.MetricsWindow == 0 {
		cfg.MetricsWindow = 60 * time.Second
	}
	return &fsMonitor{cfg: cfg}
}

// fsMonitor — Subscribe será implementado em subscribe.go (T7). Stub aqui
// só para satisfazer a interface durante T1.
type fsMonitor struct {
	cfg Config
}

func (m *fsMonitor) Subscribe(pid int, port int, logPath string) (<-chan MonitorEvent, func() error, error) {
	return nil, nil, errors.New("monitor: Subscribe not yet implemented (T7)")
}

// Garante imports usados depois.
var _ io.ReadCloser = (*emptyReader)(nil)

type emptyReader struct{}

func (emptyReader) Read(p []byte) (int, error) { return 0, io.EOF }
func (emptyReader) Close() error               { return nil }
