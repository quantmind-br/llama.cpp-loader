package processmgr

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func fakeBinary(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../../../testdata/fake-llama-server.sh")
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func newTestManager(t *testing.T) (*fsManager, string) {
	t.Helper()
	dir := t.TempDir()
	mgr := New(Config{
		Binary:       fakeBinary(t),
		LogDir:       filepath.Join(dir, "logs"),
		RegistryPath: filepath.Join(dir, "instances.json"),
	})
	return mgr, dir
}

func TestManager_LaunchBackground_WaitsHealthyAndPersists(t *testing.T) {
	mgr, _ := newTestManager(t)
	port := freePort(t)
	p := domain.Profile{
		ID:    "smoke",
		Name:  "Smoke",
		Model: "/dev/null",
		Args:  map[string]any{"port": float64(port)},
	}
	inst, err := mgr.Launch(p, LaunchBackground)
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	defer mgr.Kill(inst.PID)

	if inst.PID <= 0 || inst.Port != port || !inst.Background {
		t.Fatalf("inst = %+v", inst)
	}

	if err := mgr.WaitHealthy(inst.PID, port, 5*time.Second); err != nil {
		t.Fatalf("WaitHealthy: %v", err)
	}

	loaded, err := loadRegistry(mgr.registryPath)
	if err != nil {
		t.Fatalf("loadRegistry: %v", err)
	}
	found := false
	for _, ri := range loaded {
		if ri.PID == inst.PID {
			found = true
		}
	}
	if !found {
		t.Errorf("registry missing pid %d; got %+v", inst.PID, loaded)
	}
}

func TestManager_LaunchBackground_PortBusy(t *testing.T) {
	mgr, _ := newTestManager(t)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	p := domain.Profile{
		ID:    "busy",
		Model: "/dev/null",
		Args:  map[string]any{"port": float64(port)},
	}
	_, err = mgr.Launch(p, LaunchBackground)
	if err == nil {
		t.Fatal("expected ErrPortBusy, got nil")
	}
	if !errorsIs(err, ErrPortBusy) {
		t.Fatalf("err = %v, want ErrPortBusy", err)
	}
}

// errorsIs is a small wrapper to keep imports tidy across test files.
func errorsIs(err, target error) bool {
	for e := err; e != nil; {
		if e == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

func TestManager_WaitHealthy_TimesOut(t *testing.T) {
	mgr, _ := newTestManager(t)
	port := freePort(t)
	err := mgr.WaitHealthy(99999, port, 300*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errorsIs(err, ErrHealthCheckTimeout) {
		t.Fatalf("err = %v, want ErrHealthCheckTimeout", err)
	}
}

// helper used by future tests
func mustExtractPort(t *testing.T, p int) string {
	t.Helper()
	return fmt.Sprintf("%d", p)
}
