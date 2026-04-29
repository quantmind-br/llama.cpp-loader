package processmgr

import (
	"errors"
	"syscall"
	"time"
)

// errLivenessUnavailable é retornado quando a plataforma não suporta a probe
// (não é o caso em Linux). Mantido como sentinela para futuras builds.
var errLivenessUnavailable = errors.New("liveness probe unavailable on this platform")

// probePIDAlive retorna true quando syscall.Kill(pid, 0) sucede, o que
// significa que o processo existe (independente da permissão de signaling).
// Em Linux, ESRCH indica que o PID já não existe.
func probePIDAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	if errors.Is(err, syscall.EPERM) {
		// existe mas não temos permissão — ainda assim, vivo.
		return true
	}
	return false
}

// startLivenessWithProbe inicia uma goroutine que polla cada `interval` os
// PIDs trackeados e marca como Crashed os que `probe(pid)` retornar false.
// Retorna função stop() idempotente.
func (m *fsManager) startLivenessWithProbe(interval time.Duration, probe func(int) bool) func() {
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case now := <-t.C:
				m.mu.Lock()
				dirty := false
				for pid, inst := range m.tracked {
					if inst.Crashed {
						continue
					}
					if probe(pid) {
						continue
					}
					t := now.UTC()
					inst.Crashed = true
					inst.ExitedAt = &t
					m.tracked[pid] = inst
					dirty = true
				}
				snapshot := snapshotLocked(m.tracked)
				m.mu.Unlock()
				if dirty {
					_ = saveRegistry(m.registryPath, snapshot)
				}
			}
		}
	}()
	once := false
	return func() {
		if once {
			return
		}
		once = true
		close(stop)
	}
}

// startLiveness usa a probe default (syscall.Kill) e tick de 5 segundos.
func (m *fsManager) startLiveness() func() {
	return m.startLivenessWithProbe(5*time.Second, probePIDAlive)
}
