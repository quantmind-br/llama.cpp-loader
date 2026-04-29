package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestSubscribe_FansInLogsSlotsAndHealth(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "server.log")
	if err := os.WriteFile(logPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/slots":
			_ = json.NewEncoder(w).Encode([]Slot{{ID: 0, State: "idle", NCtxMax: 4096}})
		}
	}))
	defer srv.Close()

	mgr := New(Config{
		SlotsTickInterval: 50 * time.Millisecond,
		GPUTickInterval:   200 * time.Millisecond,
		LogRingSize:       100,
		MetricsWindow:     time.Second,
	})

	port := mustPort(t, srv.URL)
	ch, cancel, err := mgr.Subscribe(99999, port, logPath)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer cancel()

	go func() {
		time.Sleep(60 * time.Millisecond)
		_ = appendLine(logPath, "hello")
	}()

	gotLog, gotSlots, gotHealth := false, false, false
	deadline := time.After(3 * time.Second)
	for !gotLog || !gotSlots || !gotHealth {
		select {
		case ev := <-ch:
			switch ev.Source {
			case SourceLogs:
				gotLog = true
			case SourceSlots:
				gotSlots = true
			case SourceHealth:
				gotHealth = true
			}
		case <-deadline:
			t.Fatalf("timeout: log=%v slots=%v health=%v", gotLog, gotSlots, gotHealth)
		}
	}
}

func TestSubscribe_RejectsEmptyLogPath(t *testing.T) {
	mgr := New(Config{})
	if _, _, err := mgr.Subscribe(123, 8080, ""); err == nil {
		t.Fatal("expected ErrLogPathEmpty")
	}
}

// helpers
func mustPort(t *testing.T, urlStr string) int {
	t.Helper()
	u, err := url.Parse(urlStr)
	if err != nil {
		t.Fatal(err)
	}
	n, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func appendLine(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}
