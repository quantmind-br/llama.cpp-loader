package processmgr

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

func TestReconcile_DropsZombiePIDs(t *testing.T) {
	mgr, dir := newTestManager(t)

	// Write a registry with two entries: one fake (PID=1, init — alive but
	// name != llama-server) and one with our own PID (alive, name != llama-server).
	// Both should be dropped because comm doesn't contain the binary.
	entries := []domain.RunningInstance{
		{ProfileID: "ghost", PID: 1, Port: 9001, LogPath: filepath.Join(dir, "logs/ghost.log"), StartedAt: time.Now(), Background: true},
		{ProfileID: "self", PID: os.Getpid(), Port: 9002, LogPath: filepath.Join(dir, "logs/self.log"), StartedAt: time.Now(), Background: true},
	}
	if err := saveRegistry(mgr.registryPath, entries); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Reconcile(); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if got := mgr.List(); len(got) != 0 {
		t.Errorf("List after Reconcile = %v, want empty", got)
	}

	loaded, _ := loadRegistry(mgr.registryPath)
	if len(loaded) != 0 {
		t.Errorf("on-disk registry = %v, want empty", loaded)
	}
}

func TestReconcile_KeepsLiveLlamaServer(t *testing.T) {
	mgr, _ := newTestManager(t)
	port := freePort(t)
	p := domain.Profile{ID: "alive", Model: "/dev/null", Args: map[string]any{"port": float64(port)}}
	inst, err := mgr.Launch(p, LaunchBackground)
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	defer mgr.Kill(inst.PID)

	// Wait for the fake server to actually serve /health — this guarantees
	// the bash→python3 exec transition has completed and /proc/<pid>/comm
	// reads "python3" deterministically.
	if err := mgr.WaitHealthy(inst.PID, port, 5*time.Second); err != nil {
		t.Fatalf("WaitHealthy: %v", err)
	}

	// Forge a fresh manager pointing at the same registry — simulates restart.
	dir := filepath.Dir(mgr.registryPath)
	freshMgr := New(Config{
		Binary:       "python3", // matches /proc/<pid>/comm of fake-llama-server.sh's exec'd interpreter
		LogDir:       filepath.Join(dir, "logs"),
		RegistryPath: mgr.registryPath,
	})
	if err := freshMgr.Reconcile(); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got := freshMgr.List()
	if len(got) != 1 || got[0].PID != inst.PID {
		t.Errorf("expected 1 entry pid=%d, got %+v", inst.PID, got)
	}

	// Cleanup via fresh manager.
	_ = freshMgr.Kill(inst.PID)
}
