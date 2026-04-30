# AGENTS.md — internal/service/processmgr

## OVERVIEW
Process lifecycle service: spawns llama-server as background (detached) or foreground (attached) processes, tracks them in memory, and persists state to instances.json.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| processmgr.go | Manager interface, LaunchMode enum, sentinel errors |
| manager.go | fsManager implementation: Launch, Kill, List, WaitHealthy, TailLogs |
| args.go | BuildArgs(p domain.Profile) → []string for llama-server CLI |
| registry.go | JSON load/save for instances.json |
| recover.go | Reconcile(): drops zombie PIDs, keeps live llama-server processes |
| liveness.go | Background goroutine probing PIDs; marks dead ones Crashed |

## CONVENTIONS

- **LaunchMode**: Background = detached + log file + registry persistence. Foreground = attached TUI stream, max 1 at a time.
- **Logs**: Per-PID files under LogDir (`<pid>.log`), appended via os.OpenFile(O_APPEND).
- **Health check**: TCP dial on the profile's port; timeout is configurable.
- **Flag canonicalization**: `canonicalFlag()` maps user-friendly short keys (e.g. `"ngl"`) to long-form llama-server flags (`"n-gpu-layers"`) via `shortToLong` table.
- **Tests**: Use `fakeBinary(t)` helper for a no-op executable; `freePort(t)` to avoid conflicts.

## ANTI-PATTERNS

- DO NOT spawn multiple foreground instances; manager enforces only one.

## NOTES

- Log files are never rotated by processmgr; external cleanup required.
- `Reconcile` runs at TUI boot after `NewWithCheck`.
