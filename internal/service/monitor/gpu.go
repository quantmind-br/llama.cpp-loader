package monitor

import (
	"context"
	"encoding/csv"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type gpuPoller struct {
	pid           int
	nvidiaSmiPath string
	interval      time.Duration
	out           chan<- MonitorEvent
}

func newGPUPoller(pid int, nvidiaSmi string, interval time.Duration, out chan<- MonitorEvent) *gpuPoller {
	return &gpuPoller{pid: pid, nvidiaSmiPath: nvidiaSmi, interval: interval, out: out}
}

func (p *gpuPoller) run(ctx context.Context) {
	tick := time.NewTicker(p.interval)
	defer tick.Stop()
	p.pollOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			p.pollOnce(ctx)
		}
	}
}

func (p *gpuPoller) pollOnce(ctx context.Context) {
	// gopsutil first (process-aware) — skipped when pid==0 (test mode).
	if p.pid > 0 {
		if stats, ok := p.tryGopsutil(ctx); ok {
			p.emit(stats)
			return
		}
	}
	if stats, ok := p.tryNvidiaSmi(ctx); ok {
		p.emit(stats)
	}
}

// tryGopsutil — placeholder; we keep gopsutil out of the test path because
// it can panic on unsupported drivers. Real implementation:
//   import "github.com/shirou/gopsutil/v3/process"
//   import "github.com/shirou/gopsutil/v3/host"
// On Linux without GPU vendor support, gopsutil will not expose VRAM, so
// in practice this almost always falls through to nvidia-smi.
func (p *gpuPoller) tryGopsutil(ctx context.Context) (GPUStats, bool) {
	return GPUStats{}, false
}

func (p *gpuPoller) tryNvidiaSmi(ctx context.Context) (GPUStats, bool) {
	if p.nvidiaSmiPath == "" {
		return GPUStats{}, false
	}
	cmd := exec.CommandContext(ctx, p.nvidiaSmiPath,
		"--query-gpu=memory.used,memory.total,utilization.gpu",
		"--format=csv,noheader,nounits")
	stdout, err := cmd.Output()
	if err != nil {
		return GPUStats{}, false
	}
	r := csv.NewReader(strings.NewReader(strings.TrimSpace(string(stdout))))
	r.TrimLeadingSpace = true
	rec, err := r.Read()
	if err != nil || len(rec) < 3 {
		return GPUStats{}, false
	}
	used, _ := strconv.ParseUint(strings.TrimSpace(rec[0]), 10, 64)
	total, _ := strconv.ParseUint(strings.TrimSpace(rec[1]), 10, 64)
	util, _ := strconv.ParseFloat(strings.TrimSpace(rec[2]), 64)
	return GPUStats{VRAMUsedMB: used, VRAMTotalMB: total, Utilization: util, Source: "nvidia-smi"}, true
}

func (p *gpuPoller) emit(s GPUStats) {
	ev := MonitorEvent{Timestamp: time.Now(), Source: SourceGPU, Data: s, PID: p.pid}
	select {
	case p.out <- ev:
	default:
	}
}
