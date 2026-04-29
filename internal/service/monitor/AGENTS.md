# AGENTS.md — internal/service/monitor

## OVERVIEW

Streams logs, slot snapshots, health status, GPU stats, and rolling metrics from a running llama-server via a single `Subscribe` call. Six goroutines per subscription feed a unified event channel.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `monitor.go` | `Manager` interface, `MonitorEvent` types, `Config`, sentinel errors |
| `subscribe.go` | `Subscribe` spawns 6 goroutines: log tail, slots poll, GPU poll, log pump, slots pump, metrics ticker |
| `slots.go` | Polls `GET /health` and `GET /slots` on the server port |
| `gpu.go` | `nvidia-smi` fallback; `gopsutil` is a no-op placeholder |
| `logs.go` | `fsnotify`-based log file tailer |
| `metrics.go` | Rolling window aggregator: tokens/s from log regex, requests/s from slot `n_decoded` diffs |
| `ring.go` | Fixed-capacity ring buffer for latest log lines |

## CONVENTIONS

- All pollers drop events on backpressure (non-blocking send with `default`)
- Consumers type-assert `MonitorEvent.Data` based on `EventSource`
- `HTTPDoer` interface swaps in a mock client for tests
- `New(Config)` applies defaults for all zero values

## ANTI-PATTERNS

- DO NOT block the event channel: Subscribe uses a 256-buffered channel and drops on overflow
- DO NOT expect `gopsutil` GPU data: it always returns false; rely on `nvidia-smi`

## NOTES

- `Subscribe` requires a non-empty `logPath`; it validates before spawning goroutines
- Cancel func returned by `Subscribe` waits on `sync.WaitGroup` before closing the channel
- `Metrics` window defaults to 60s; `LogRing` capacity defaults to 2000 lines
