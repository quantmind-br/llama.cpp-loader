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
	slotsRaw := make(chan MonitorEvent, 256)
	agg := newMetricsAgg(m.cfg.MetricsWindow)

	var wg sync.WaitGroup
	if err := startLogFollower(ctx, &wg, logPath, logLines); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("log follower: %w", err)
	}
	startSlotsPoller(ctx, &wg, port, m.cfg, slotsRaw)
	startGPUPoller(ctx, &wg, pid, m.cfg, out)
	runLogPump(ctx, &wg, pid, m.cfg.LogRingSize, agg, logLines, out)
	runSlotsPump(ctx, &wg, pid, agg, slotsRaw, out)
	runMetricsTick(ctx, &wg, pid, m.cfg.SlotsTickInterval, agg, out)

	sub := &subscription{cancel: cancel, doneCh: make(chan struct{})}
	closeOnDone(&wg, out, sub.doneCh)

	return out, makeCancelFn(sub.cancel, sub.doneCh), nil
}

// startLogFollower opens the log file and spawns a tailing goroutine that
// pushes every appended line into logLines. Returns an error iff the file
// cannot be opened. logLines is closed when the goroutine exits so the log
// pump terminates cleanly on cancel.
func startLogFollower(ctx context.Context, wg *sync.WaitGroup, logPath string, logLines chan<- string) error {
	follower, err := newLogFollower(logPath, logLines)
	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(logLines)
		follower.run(ctx)
	}()
	return nil
}

// startSlotsPoller emits SourceSlots/SourceHealth events into slotsRaw at
// cfg.SlotsTickInterval. Closes slotsRaw on exit so the slots pump
// terminates.
func startSlotsPoller(ctx context.Context, wg *sync.WaitGroup, port int, cfg Config, slotsRaw chan<- MonitorEvent) {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	sp := newSlotsPoller(baseURL, cfg.HTTPClient, cfg.SlotsTickInterval, slotsRaw)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(slotsRaw)
		sp.run(ctx)
	}()
}

// startGPUPoller emits SourceGPU events directly to out at
// cfg.GPUTickInterval.
func startGPUPoller(ctx context.Context, wg *sync.WaitGroup, pid int, cfg Config, out chan<- MonitorEvent) {
	g := newGPUPoller(pid, cfg.NvidiaSMIPath, cfg.GPUTickInterval, out)
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.run(ctx)
	}()
}

// runLogPump consumes logLines, maintains the ring buffer for crash dumps,
// feeds the metrics aggregator, and forwards each line as SourceLogs.
func runLogPump(ctx context.Context, wg *sync.WaitGroup, pid, ringSize int, agg *metricsAgg, logLines <-chan string, out chan<- MonitorEvent) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ring := newLogRing(ringSize)
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
}

// runSlotsPump feeds slot snapshots into the metrics aggregator and forwards
// every slotsRaw event to out, stamping the PID first.
func runSlotsPump(ctx context.Context, wg *sync.WaitGroup, pid int, agg *metricsAgg, slotsRaw <-chan MonitorEvent, out chan<- MonitorEvent) {
	wg.Add(1)
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
}

// runMetricsTick emits a SourceMetrics snapshot every interval.
func runMetricsTick(ctx context.Context, wg *sync.WaitGroup, pid int, interval time.Duration, agg *metricsAgg, out chan<- MonitorEvent) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(interval)
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
}

// closeOnDone waits for every poller goroutine to finish, then closes the
// output channel and signals subscription teardown via doneCh.
func closeOnDone(wg *sync.WaitGroup, out chan<- MonitorEvent, doneCh chan struct{}) {
	go func() {
		wg.Wait()
		close(out)
		close(doneCh)
	}()
}

// makeCancelFn wraps the context cancel and a wait on doneCh so the caller's
// teardown blocks until every poller has exited.
func makeCancelFn(cancel context.CancelFunc, doneCh <-chan struct{}) func() error {
	return func() error {
		cancel()
		<-doneCh
		return nil
	}
}
