package processmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

// Reconcile reads the on-disk registry, validates each entry against the
// running process table, drops zombies, and updates the registry.
//
// An entry survives only if BOTH:
//   - the PID is alive (signal 0 succeeds), AND
//   - /proc/<pid>/comm contains the basename of the binary the manager
//     was configured with (default: "llama-server"). This avoids
//     mistaking a recycled PID for a live server.
//
// Linux-only. Other platforms: this method drops every entry (safe default).
func (m *fsManager) Reconcile() error {
	loaded, err := loadRegistry(m.registryPath)
	if err != nil {
		return err
	}
	expectedComm := filepath.Base(m.binary)

	survivors := make([]domain.RunningInstance, 0, len(loaded))
	tracked := make(map[int]domain.RunningInstance, len(loaded))
	for _, ri := range loaded {
		if !pidAliveAndNameMatches(ri.PID, expectedComm) {
			continue
		}
		survivors = append(survivors, ri)
		tracked[ri.PID] = ri
	}

	m.mu.Lock()
	m.tracked = tracked
	m.mu.Unlock()

	if err := saveRegistry(m.registryPath, survivors); err != nil {
		return fmt.Errorf("rewrite registry: %w", err)
	}
	return nil
}

// pidAliveAndNameMatches returns true iff pid is alive AND the basename of
// /proc/<pid>/comm contains expectedComm. Reads /proc directly (Linux).
func pidAliveAndNameMatches(pid int, expectedComm string) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	commBytes, err := os.ReadFile(filepath.Join("/proc", fmt.Sprintf("%d", pid), "comm"))
	if err != nil {
		return false
	}
	comm := strings.TrimSpace(string(commBytes))
	return strings.Contains(comm, expectedComm)
}
