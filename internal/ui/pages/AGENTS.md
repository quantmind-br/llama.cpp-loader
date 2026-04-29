# internal/ui/pages

## OVERVIEW
Tab page implementations. Each page is a `tea.Model` with isolated state and lifecycle.

## WHERE TO LOOK
| File | Purpose |
|------|---------|
| `profiles.go` + `profiles_editor.go` | Profile CRUD master-detail with `huh` forms |
| `models.go` | GGUF model browser and scanner integration |
| `launcher.go` | Profile selection, validation, launch orchestration |
| `monitor.go` | Live GPU metrics, logs, slots per running instance |
| `messages.go` | Cross-tab messages consumed by `root.go` |
| `placeholder.go` | Generic stub for unimplemented tabs |

## CONVENTIONS
- Pages receive service dependencies via constructors, never globals
- `huh` forms for all interactive input
- `bubbles/list` and `bubbles/table` for collections
- Test files are side-by-side (`*_test.go`) with table-driven tests

## ANTI-PATTERNS
- DO NOT import `internal/ui/root` from pages (pages are leaf nodes)
- DO NOT handle global keybinds here (`root.go` owns tab switching, quit, etc.)

## NOTES
- `SwitchToMonitorMsg` and `MonitorSelectPIDMsg` are the only cross-tab coordination mechanism
- `Placeholder` can be reused for new tabs before full implementation
