package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type subscription struct {
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

	slotsRaw := make(chan MonitorEvent, 256)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	slots := newSlotsPoller(baseURL, m.cfg.HTTPClient, m.cfg.SlotsTickInterval, slotsRaw)
	gpu := newGPUPoller(pid, m.cfg.NvidiaSMIPath, m.cfg.GPUTickInterval, out)
	agg := newMetricsAgg(m.cfg.MetricsWindow)

	var wg sync.WaitGroup
	wg.Add(6)

	// follower: tail log file, push lines into logLines.
	go func() {
		defer wg.Done()
		defer close(logLines)
		follower.run(ctx)
	}()

	// slots-poller: emit SourceSlots/SourceHealth into slotsRaw.
	go func() {
		defer wg.Done()
		defer close(slotsRaw)
		slots.run(ctx)
	}()

	// gpu-poller: emit SourceGPU directly to out.
	go func() {
		defer wg.Done()
		gpu.run(ctx)
	}()

	// log-pump: ring buffer + metrics observe + forward as SourceLogs.
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

	// slots-pump: feed observeSlots, forward all slotsRaw events to out.
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-slotsRaw:
				if !ok {
					return
				}
				if ev.Source == SourceSlots {
					if snap, ok := ev.Data.(SlotSnapshot); ok {
						agg.observeSlots(time.Now(), snap)
					}
				}
				ev.PID = pid
				select {
				case out <- ev:
				default:
				}
			}
		}
	}()

	// metrics-tick: periodic snapshot emission.
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

	sub := &subscription{cancel: cancel, doneCh: make(chan struct{})}
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
