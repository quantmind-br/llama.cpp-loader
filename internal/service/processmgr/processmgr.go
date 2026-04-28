// Package processmgr launches and tracks llama-server processes.
package processmgr

import (
	"errors"
	"time"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

// LaunchMode selects how the spawned process is attached.
type LaunchMode int

const (
	// LaunchBackground spawns the process detached (Setsid), redirects
	// stdout/stderr to a log file, and persists the entry in instances.json.
	LaunchBackground LaunchMode = iota
	// LaunchForeground spawns the process attached for in-TUI streaming.
	// Only one foreground instance is allowed at a time.
	LaunchForeground
)

// Manager owns the lifecycle of llama-server processes.
type Manager interface {
	Launch(p domain.Profile, mode LaunchMode) (domain.RunningInstance, error)
	Kill(pid int) error
	List() []domain.RunningInstance
	WaitHealthy(pid, port int, timeout time.Duration) error
}

// Sentinel errors. UI maps these to status bar messages.
var (
	ErrPortBusy           = errors.New("port already in use")
	ErrModelNotFound      = errors.New("model file not found") // reserved; explicit os.Stat check on Profile.Model deferred
	ErrForegroundBusy     = errors.New("a foreground instance is already running")
	ErrUnknownPID         = errors.New("pid is not tracked by this manager")
	ErrHealthCheckTimeout = errors.New("llama-server did not become healthy within timeout")
)
