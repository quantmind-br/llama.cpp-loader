package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type subscription struct {
	pid    int
	cancel context.CancelFunc
	doneCh chan struct{}
}

func (m *fsMonitor) Subscribe(pid, port int, logPath string) (<-chan MonitorEvent, func() error, error) {
	if logPath == "" {
		return nil, nil, ErrLogPathEmpty
	}
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan MonitorEvent, 256)

	logLines := make(chan string, 256)
	follower, err := newLogFollower(logPath, logLines)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("log follower: %w", err)
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	slots := newSlotsPoller(baseURL, m.cfg.HTTPClient, m.cfg.SlotsTickInterval, out)
	gpu := newGPUPoller(pid, m.cfg.NvidiaSMIPath, m.cfg.GPUTickInterval, out)
	agg := newMetricsAgg(m.cfg.MetricsWindow)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); follower.run(ctx) }()
	go func() { defer wg.Done(); slots.run(ctx) }()
	go func() { defer wg.Done(); gpu.run(ctx) }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ring := newLogRing(m.cfg.LogRingSize)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-logLines:
				if !ok {
					return
				}
				ring.push(line)
				agg.observeLog(time.Now(), line)
				ev := MonitorEvent{
					Timestamp: time.Now(),
					Source:    SourceLogs,
					PID:       pid,
					Data:      LogLine{Line: line},
				}
				select {
				case out <- ev:
				default:
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(m.cfg.SlotsTickInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-t.C:
				snap := agg.snapshot(now)
				select {
				case out <- MonitorEvent{Timestamp: now, Source: SourceMetrics, PID: pid, Data: snap}:
				default:
				}
			}
		}
	}()

	sub := &subscription{pid: pid, cancel: cancel, doneCh: make(chan struct{})}
	go func() {
		wg.Wait()
		close(out)
		close(sub.doneCh)
	}()

	return out, func() error {
		sub.cancel()
		<-sub.doneCh
		return nil
	}, nil
}
