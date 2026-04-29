package pages

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/monitor"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/processmgr"
)

type fakeProcMgr struct {
	insts []domain.RunningInstance
}

func (f *fakeProcMgr) Launch(p domain.Profile, m processmgr.LaunchMode) (domain.RunningInstance, error) {
	return domain.RunningInstance{}, nil
}
func (f *fakeProcMgr) Kill(pid int) error                               { return nil }
func (f *fakeProcMgr) List() []domain.RunningInstance                   { return f.insts }
func (f *fakeProcMgr) WaitHealthy(pid, port int, t time.Duration) error { return nil }
func (f *fakeProcMgr) TailLogs(pid int) (io.ReadCloser, error)          { return nil, nil }

type fakeMonMgr struct{}

func (fakeMonMgr) Subscribe(pid, port int, logPath string) (<-chan monitor.MonitorEvent, func() error, error) {
	ch := make(chan monitor.MonitorEvent)
	return ch, func() error { close(ch); return nil }, nil
}

func TestMonitorPage_RendersInstanceRows(t *testing.T) {
	pm := &fakeProcMgr{insts: []domain.RunningInstance{
		{PID: 1234, Port: 8080, ProfileID: "p1", LogPath: "/tmp/x.log"},
		{PID: 5678, Port: 8081, ProfileID: "p2", LogPath: "/tmp/y.log"},
	}}
	mm := fakeMonMgr{}
	p := NewMonitorPage(pm, mm)
	p.SetSize(120, 30)

	if cmd := p.Init(); cmd != nil {
		if msg := cmd(); msg != nil {
			p.Update(msg)
		}
	}

	view := p.View()
	if !strings.Contains(view, "1234") {
		t.Fatalf("view missing pid 1234:\n%s", view)
	}
	if !strings.Contains(view, "5678") {
		t.Fatalf("view missing pid 5678:\n%s", view)
	}
}
