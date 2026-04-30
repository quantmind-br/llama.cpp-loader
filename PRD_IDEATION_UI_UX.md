# PRD — UI/UX Improvements

**Source:** `IDEATION_UI_UX.md`
**Generated:** 2026-04-29

## Implementation Order

1. **UIUX-008** — Adaptive theme + `NO_COLOR` honoring
2. **UIUX-024** — Modal background adaptive
3. **UIUX-007** — `HintProvider` contract; page footers folded into status bar
4. **UIUX-001** — Populate Monitor row columns from live `subState`
5. **UIUX-002** — Confirmation dialogs on `[k]` kill (Launcher / Monitor) and editor `esc` with unsaved changes
6. **UIUX-003** — `WaitHealthy` waiting indicator with spinner
7. **UIUX-005** — Visible pause state in Monitor logs
8. **UIUX-006a** — Monitor crashed rows colored
9. **UIUX-006b** — Profiles corrupt rows colored via custom `list.ItemDelegate`
10. **UIUX-015** — Picker surfaces scan errors
11. **UIUX-004** — Monitor sub-view tab strip (Logs / Slots / Metrics)
12. **UIUX-018** — Consistent empty states with next-action guidance
13. **UIUX-009** — Stderr notice at quit when background instances remain
14. **UIUX-010** — Document key-case convention in `?` help
15. **UIUX-012** — Tab strip separator
16. **UIUX-013** — Picker responsive width
17. **UIUX-014** — Picker PgUp / PgDn / Home / End / g / G
18. **UIUX-017** — Models status line wrap / truncate
19. **UIUX-020** — Flash auto-clear
20. **UIUX-021** — Inline numeric `huh` validators
21. **UIUX-022** — Distinct sparkline colors per metric
22. **UIUX-023** — "Showing N of M" footer on Monitor logs
23. **UIUX-016** — Reveal path → clipboard copy
24. **UIUX-011** — Pane border on Launcher's left/right split

---

## UIUX-008: Adaptive theme + `NO_COLOR` honoring

### Scope
**In scope**
- Replace every fixed `lipgloss.Color("#…")` in `internal/ui/theme/theme.go` with `lipgloss.AdaptiveColor{Light, Dark}`.
- Honor `NO_COLOR` env var at theme init: emit styles with cleared `Foreground` / `Background`, retain `Bold`.
- Keep style **names** unchanged (`Title`, `Subtitle`, `OK`, `Warn`, `Error`, `Selected`, `TabActive`, `TabInactive`, `Pane`).

**Out of scope**
- New theme switcher UI / runtime toggle.
- `LIGHT_THEME=…` / `DARK_THEME=…` env overrides.
- Restyling individual page bodies (separate items).

### Technical Approach
1. Define a private `paletteEntry` table mapping each color name to `(light, dark)` hex pairs.
2. Build the package-level `Color*` vars by calling `lipgloss.AdaptiveColor{Light, Dark}` from that table.
3. Add `init()` (or `var _ = func()`) that reads `os.Getenv("NO_COLOR")`. When non-empty:
   - Recreate each exported `Style` (`Title`, `Subtitle`, `OK`, `Warn`, `Error`, `Selected`, `TabActive`, `TabInactive`, `Pane`) with `.UnsetForeground().UnsetBackground()` applied.
   - Bold is retained where the original style had it.
4. Audit `internal/ui/components/modal.go` for any fixed colors that should also flow through the palette (`modalBox` border references `theme.ColorAccent` already; only the literal `#1a1a1a` background — handled by UIUX-024).

### Touchpoints
- `internal/ui/theme/theme.go` — full rewrite of color/style definitions.
- (read-only) every file that imports the package — must keep compiling without source changes.

### Contracts
```go
// internal/ui/theme/theme.go

var (
    ColorAccent     lipgloss.AdaptiveColor // Light: "#0969da", Dark: "#58a6ff"
    ColorOK         lipgloss.AdaptiveColor // Light: "#1a7f37", Dark: "#3fb950"
    ColorWarn       lipgloss.AdaptiveColor // Light: "#9a6700", Dark: "#d29922"
    ColorError      lipgloss.AdaptiveColor // Light: "#cf222e", Dark: "#f85149"
    ColorDim        lipgloss.AdaptiveColor // Light: "#57606a", Dark: "#6e7681"
    ColorSelectedBG lipgloss.AdaptiveColor // Light: "#0969da", Dark: "#1f6feb"
    ColorSelectedFG lipgloss.AdaptiveColor // Light: "#ffffff", Dark: "#ffffff"
)

// noColor reports whether the NO_COLOR env var is non-empty.
func noColor() bool { return os.Getenv("NO_COLOR") != "" }
```

### Acceptance Criteria
- [ ] `go test ./internal/ui/...` passes against the rewritten theme without any test-source change.
- [ ] On a light terminal (programmatic `LIPGLOSS_BACKGROUND=light` test or manual), the Title/Subtitle/Selected styles are legible (foreground contrasts with the bg).
- [ ] Setting `NO_COLOR=1` and starting the TUI yields output where no ANSI color escape codes are present in `View()` strings (verified by snapshot test stripping ANSI then comparing to a colorless golden).
- [ ] No `lipgloss.Color("#…")` literals remain in `theme.go` or `modal.go`.

### Dependencies
- None.

---

## UIUX-024: Modal background adaptive

### Scope
**In scope**
- Drop the hardcoded `Background(lipgloss.Color("#1a1a1a"))` on `modalBox`.

**Out of scope**
- Restyling modal content (title / body) — already use `theme.*`.
- Adding a separate dim-overlay backdrop behind the modal.

### Technical Approach
1. Edit `modalBox` style in `internal/ui/components/modal.go`: remove the `.Background(...)` call. Keep `Border(theme.Border)`, `BorderForeground(theme.ColorAccent)`, `Padding(1, 2)`.

### Touchpoints
- `internal/ui/components/modal.go`.

### Contracts
```go
var modalBox = lipgloss.NewStyle().
    Border(theme.Border).
    BorderForeground(theme.ColorAccent).
    Padding(1, 2)
```

### Acceptance Criteria
- [ ] `Modal("title", "body", 80, 24)` returns a string containing no `\x1b[48;` (background) escape sequence — only foreground escapes for accent border and title.
- [ ] Existing modal tests (`modal_test.go`) pass without changes, or are updated to reflect the no-bg state.

### Dependencies
- UIUX-008.

---

## UIUX-007: `HintProvider` contract collapsing footer / status duplication

### Scope
**In scope**
- Add `HintProvider interface { Hints() string }` to `internal/ui/root.go` next to existing `Page` / `InputCapture` / `Reloader`.
- Implement `Hints()` on `ProfilesPage`, `LauncherPage`, `MonitorPage`, `ModelsPage`.
- `RootModel` recomputes `m.status.Hints` after every successful `activate(t)`, after every `KeyMsg` that mutates page state, and on `tea.WindowSizeMsg`.
- Remove the per-page footer strings (the inline `lipgloss.JoinVertical(..., footer)` blocks) from each page's `View()`.

**Out of scope**
- Animating hint transitions.
- Per-mode hint switching inside a page (e.g., editor vs. list inside Profiles) — pages are free to vary their own `Hints()` return value, but no extra plumbing beyond `HintProvider`.
- Removing `components.HelpToken` itself; just stop appending it twice.

### Technical Approach
1. Add the interface declaration at the top of `internal/ui/root.go`.
2. Add a helper on `RootModel`: `recomputeHints()` reads `m.pages[m.active].(HintProvider)`; if implemented, sets `m.status.Hints = globalHints + " | " + page.Hints()`. Otherwise sets `m.status.Hints = globalHints`. `globalHints` is `"[1-4] tabs  [tab] next  [q] quit" + components.HelpToken`.
3. Call `m.recomputeHints()` at the end of `activate`, `handleKey`, and `handleResize`.
4. Each page implements `Hints() string` returning page-local key hints (no `[?] help` token — global owns it). Examples below.
5. Remove the trailing footer strings from each page's `View()` — just drop the lipgloss join with the footer.
6. Update root tests where `View()` snapshots referenced the old per-page footer text.

### Touchpoints
- `internal/ui/root.go` — interface, `recomputeHints`, hook points.
- `internal/ui/pages/profiles.go` — add `Hints()`, drop footer in `detailView` / `View`.
- `internal/ui/pages/launcher.go` — add `Hints()`, drop footer composition.
- `internal/ui/pages/monitor.go` — add `Hints()`, drop footer.
- `internal/ui/pages/models.go` — add `Hints()`, drop footer.
- `internal/ui/root_test.go` — update affected snapshots.

### Contracts
```go
// internal/ui/root.go
type HintProvider interface { Hints() string }

func (m *RootModel) recomputeHints() {
    const global = "[1-4] tabs  [tab] next  [q] quit" + components.HelpToken
    if h, ok := m.pages[m.active].(HintProvider); ok {
        m.status.Hints = global + " | " + h.Hints()
        return
    }
    m.status.Hints = global
}
```

```go
// pages/profiles.go (list mode)
func (p ProfilesPage) Hints() string {
    if p.editing { return "[ctrl+t] sub-tab  [ctrl+p] pick model  [esc] cancel" }
    return "[enter] edit  [n] new  [d] dup  [x] del  [L] launch  [/] filter"
}
```

### Acceptance Criteria
- [ ] `HintProvider` is declared in `internal/ui/root.go` exactly once.
- [ ] All four real pages implement `Hints() string`.
- [ ] No page's `View()` returns a string containing the substrings `"[?] help"` or `"[1-4] tabs"`.
- [ ] After switching tabs, the bottom status bar reflects the active page's hints (verified by a teatest assertion).
- [ ] Existing root_test.go snapshots are updated; tests pass.

### Dependencies
- None (independent of UIUX-008/024 but lands after for cleaner sequencing).

---

## UIUX-001: Populate Monitor row columns from live `subState`

### Scope
**In scope**
- Fill `Uptime`, `VRAM`, `Tokens/s` columns in `pages/monitor.go applyInstances`.
- Add `StartedAt time.Time` to `domain.RunningInstance` if not already present.
- Format helpers `humanDuration(d)` and `formatVRAM(used, total)` co-located in `pages/monitor.go`.

**Out of scope**
- Persisting `StartedAt` across TUI restarts (best-effort: re-set on first refresh after reconcile if zero).
- Sub-second refresh rate.
- Per-row sparkline.

### Technical Approach
1. Inspect `domain.RunningInstance`. If missing `StartedAt`, add it (pure additive struct field). Have `processmgr.Launch(...)` set it to `time.Now()` at process start; preserve from `instances.json` on `Reconcile`.
2. Modify `MonitorPage.applyInstances`:
   - For each `ri`, look up `p.subs[ri.PID]`.
   - Compute `uptime`, `vram`, `toks` strings (defaulting to `"--"` when no data).
   - Append the values into `table.Row{...}` in place of the three `"--"` literals currently hardcoded.
3. Add `humanDuration(time.Duration) string` (e.g., `"3m12s"` / `"1h04m"`).
4. Add `formatVRAM(used, total int) string` returning `"%d/%dMB"` when total > 0, else `"--"`.
5. Add a periodic re-render path for uptime: the existing 2 s `monitorPeriodicTickMsg` already triggers `refreshInstancesCmd()`, so uptime updates without new ticks.

### Touchpoints
- `internal/domain/process.go` (or wherever `RunningInstance` lives) — add `StartedAt`.
- `internal/service/processmgr/...` — set `StartedAt` on Launch + preserve across `Reconcile`.
- `internal/ui/pages/monitor.go` — `applyInstances` rewrite, helper funcs.
- `internal/ui/pages/monitor_test.go` — extend coverage.

### Contracts
```go
// domain
type RunningInstance struct {
    PID        int
    Port       int
    ProfileID  string
    Background bool
    Crashed    bool
    LogPath    string
    StartedAt  time.Time // new
}

// pages/monitor.go
func humanDuration(d time.Duration) string
func formatVRAM(used, total int) string
```

### Acceptance Criteria
- [ ] `RunningInstance.StartedAt` is non-zero for every instance returned by `processmgr.List()` after a `Launch`.
- [ ] After launching one instance with metrics flowing, the Monitor table row reads `Uptime` as e.g. `5s`, `VRAM` as `1234/24576MB`, `Tokens/s` as a non-`--` value within 5 seconds.
- [ ] When `subState.gpu.VRAMTotalMB == 0`, the cell renders `"--"`.
- [ ] When `subState.mets.TokensPerSec` is empty, the cell renders `"--"`.
- [ ] `monitor_test.go` adds at least one assertion for non-`"--"` row values when stub `subState` carries metrics.

### Dependencies
- None (independent of UIUX-007 but better after to avoid edit conflict on monitor `View`).

---

## UIUX-002: Confirmation on destructive actions

### Scope
**In scope**
- `LauncherPage`: `[k]` kills last-launched. Wrap in a `huh.NewConfirm` modal (page-local, mirrors `profiles.askDeleteSelected`).
- `MonitorPage`: `[k]` kills selected. Same pattern.
- `ProfilesPage` editor: `esc` while editing. If the current `*p.draft` differs from the snapshot taken at form-open, show a confirm `Discard unsaved changes?`. Otherwise close immediately as today.

**Out of scope**
- Adding "are you sure" to non-destructive ops (`refresh`, `pause`, etc.).
- Adding a global confirm-suppression flag for power users.

### Technical Approach
1. Add private state to each page:
   - `LauncherPage.confirmKillForm *huh.Form`, `confirmKillAnswer *bool`, `confirmKillTargetPID int`.
   - `MonitorPage.confirmKillForm *huh.Form`, `confirmKillAnswer *bool`, `confirmKillTargetPID int`.
   - `ProfilesPage.editorOpenSnapshot profileDraft` (value, not pointer; copied at form open).
   - `ProfilesPage.confirmDiscardForm *huh.Form`, `confirmDiscardAnswer *bool`.
2. Modify `IsCapturingInput()` on Launcher and Monitor to also return `true` when `confirmKillForm != nil`.
3. Update each `Update`:
   - On `[k]`: capture the target PID, build `huh.NewConfirm`, set `confirmKillForm`, return `form.Init()`.
   - When `confirmKillForm.State == huh.StateCompleted`: read the bool answer, call `manager.Kill(pid)` if affirmative, clear confirm state.
4. Profile editor `esc`: compare `*p.draft` to `editorOpenSnapshot`. If equal, close as today; otherwise build the discard confirm. Mark `editorOpenSnapshot` set on every `startNew`/`startEditSelected`.
5. View paths: when `confirmKillForm != nil` (or `confirmDiscardForm != nil`), render the form view as the page body.

### Touchpoints
- `internal/ui/pages/launcher.go`.
- `internal/ui/pages/monitor.go`.
- `internal/ui/pages/profiles.go`.
- `internal/ui/pages/launcher_test.go`, `monitor_test.go`, `profiles_test.go`.

### Contracts
```go
// shared pattern
form := huh.NewForm(huh.NewGroup(
    huh.NewConfirm().
        Title(fmt.Sprintf("Kill pid=%d?", pid)).
        Affirmative("Kill").
        Negative("Cancel").
        Value(&answer),
)).WithShowHelp(false).WithShowErrors(false)
```

### Acceptance Criteria
- [ ] In Launcher with one instance, pressing `k` does NOT kill immediately; a confirm form is rendered.
- [ ] Pressing the Negative button leaves the instance running; pressing Affirmative kills it.
- [ ] In Monitor, same behavior.
- [ ] In Profiles editor, pressing `esc` with an unmodified draft closes the editor without prompt.
- [ ] In Profiles editor, pressing `esc` with a modified draft renders a `Discard unsaved changes?` confirm; Negative returns to the editor preserving edits, Affirmative drops the draft.
- [ ] All three paths have at least one test asserting the confirm path.

### Dependencies
- None.

---

## UIUX-003: `WaitHealthy` waiting indicator

### Scope
**In scope**
- Animated spinner in the Launcher status line during the `WaitHealthy` window.
- Status text updated to indicate "waiting for /health".

**Out of scope**
- Cancelable `WaitHealthy`.
- Per-second progress estimate.

### Technical Approach
1. Use `bubbles/spinner` (`spinner.New(spinner.WithSpinner(spinner.Dot))`).
2. Add fields to `LauncherPage`: `spin spinner.Model`, `waitingPID int`.
3. On `launchedMsg`:
   - `p.waitingPID = msg.inst.PID`
   - `p.status = fmt.Sprintf("pid=%d port=%d — waiting for /health…", msg.inst.PID, msg.inst.Port)`
   - Return a batch of `(spin.Tick, waitHealthyCmd)`.
4. Forward `spinner.TickMsg` to `p.spin.Update`.
5. In `View()`, when `waitingPID != 0`, render `spinner.View() + " " + p.status` instead of plain `p.status`.
6. On `healthyMsg` / `launchErrMsg`: clear `waitingPID`, stop forwarding ticks.

### Touchpoints
- `internal/ui/pages/launcher.go`.
- `internal/ui/pages/launcher_test.go`.

### Contracts
```go
import "github.com/charmbracelet/bubbles/spinner"

type LauncherPage struct {
    // existing…
    spin       spinner.Model
    waitingPID int
}
```

### Acceptance Criteria
- [ ] After `launchedMsg` lands and before `healthyMsg`, the page `View()` contains a spinner glyph.
- [ ] After `healthyMsg`, the spinner glyph is gone and the status reflects `healthy pid=N`.
- [ ] After `launchErrMsg`, the spinner glyph is gone and the status reflects the friendly error.

### Dependencies
- None.

---

## UIUX-005: Visible pause state in Monitor logs

### Scope
**In scope**
- Render a `(PAUSED — Space to resume)` line above the log region when paused.

**Out of scope**
- Pausing other sub-views.

### Technical Approach
1. In `MonitorPage.View()`, inside the `SubViewLogs` branch: if `p.paused`, prepend `theme.Warn.Render("Logs (PAUSED — Space to resume)") + "\n"` to `bottom`.

### Touchpoints
- `internal/ui/pages/monitor.go`.

### Acceptance Criteria
- [ ] After pressing `Space`, the rendered Monitor `View()` contains the substring `(PAUSED`.
- [ ] After pressing `Space` again, the substring is gone.

### Dependencies
- None.

---

## UIUX-006a: Color-code crashed Monitor rows

### Scope
**In scope**
- Pre-style each cell of a crashed row with `theme.Error.Render(...)` before `tbl.SetRows(rows)` in `pages/monitor.go applyInstances`.

**Out of scope**
- Custom row delegate; bubbles `table.Row` is `[]string` and accepts ANSI.

### Technical Approach
1. In the per-instance loop of `applyInstances`, when `ri.Crashed`:
   - Apply `theme.Error.Render(...)` to each cell string before the `append`.
2. Verify bubbles' table renderer doesn't double-strip ANSI; if it does, wrap only PID + Profile cells.

### Touchpoints
- `internal/ui/pages/monitor.go`.
- `internal/ui/pages/monitor_test.go`.

### Acceptance Criteria
- [ ] When an instance is crashed, the corresponding `table.Row` contains ANSI escape sequences for `theme.Error`.
- [ ] When an instance is healthy, the row contains no error-color escapes.

### Dependencies
- UIUX-001 (touches the same loop; coordinate edits).

---

## UIUX-006b: Color-code corrupt rows in Profiles list

### Scope
**In scope**
- Custom `list.ItemDelegate` that renders `corruptItem` titles with `theme.Warn` / `theme.Error` and a different bullet glyph.

**Out of scope**
- Replacing the delegate for `item` (healthy profile rows). They keep `list.NewDefaultDelegate`.

### Technical Approach
1. Add a `profileItemDelegate` type in `pages/profiles.go` that wraps `list.NewDefaultDelegate`.
2. Override `Render(...)` to detect `corruptItem` via type assertion. For corrupt rows, build the row string manually with `theme.Warn`/`Error`.
3. For non-corrupt rows, call into the wrapped default delegate.
4. Pass the new delegate to `list.New` in `NewProfilesPage`.

### Touchpoints
- `internal/ui/pages/profiles.go`.
- `internal/ui/pages/profiles_test.go`.

### Contracts
```go
type profileItemDelegate struct{ list.DefaultDelegate }

func newProfileItemDelegate() profileItemDelegate {
    return profileItemDelegate{DefaultDelegate: list.NewDefaultDelegate()}
}

func (d profileItemDelegate) Render(w io.Writer, m list.Model, idx int, item list.Item)
```

### Acceptance Criteria
- [ ] A `corruptItem` rendered through the delegate contains `theme.Warn` ANSI escapes around the title.
- [ ] A regular `item` rendered through the delegate matches the previous default-delegate output.

### Dependencies
- None.

---

## UIUX-015: Picker surfaces scan errors

### Scope
**In scope**
- Capture `PickerScanStartedMsg.Err` into `m.err string`.
- Render an error line styled `theme.Error` under the picker title when `m.err != ""`.

**Out of scope**
- Retry button.

### Technical Approach
1. Add `err string` to `ModelPicker`.
2. In `Update`, on `PickerScanStartedMsg.Err != nil`, set `m.err = msg.Err.Error()` (in addition to `m.scanning = false`).
3. In `View`, prepend `theme.Error.Render("scan error: " + m.err) + "\n"` to the box body when `m.err != ""`.

### Touchpoints
- `internal/ui/components/picker.go`.
- `internal/ui/components/picker_test.go`.

### Acceptance Criteria
- [ ] When `scanner.Scan` returns an error, the picker `View()` contains the substring `scan error:`.
- [ ] When the scan succeeds, the substring is absent.

### Dependencies
- None.

---

## UIUX-004: Monitor sub-view tab strip

### Scope
**In scope**
- Render a horizontal tab strip above the bottom region in `pages/monitor.go View()`: `Logs / Slots / Metrics`.
- Active sub-view styled `theme.TabActive`, inactive `theme.TabInactive`.
- `[v]` continues to cycle.

**Out of scope**
- Direct sub-view selection by number key.

### Technical Approach
1. Define a helper `renderSubViewTabs(active SubViewKind) string`.
2. Build it with `lipgloss.JoinHorizontal(lipgloss.Top, ...)` styling each label per active state.
3. Insert it into `MonitorPage.View()` between the table and the bottom region: `top + "\n" + subviewTabs + "\n" + bottom + "\n\n" + footer`.

### Touchpoints
- `internal/ui/pages/monitor.go`.

### Acceptance Criteria
- [ ] Monitor `View()` contains the substring `Logs` styled as the active tab when `p.subView == SubViewLogs`.
- [ ] Cycling with `v` flips which label is rendered as active.

### Dependencies
- UIUX-008 (uses `theme.TabActive` / `TabInactive`).

---

## UIUX-018: Consistent empty states

### Scope
**In scope**
- Monitor: when `len(p.tbl.Rows()) == 0`, render `(no instances running — switch to Launcher [2] to start one)` styled `theme.Subtitle`.
- Launcher: when `len(p.profiles) == 0`, render `(no profiles yet — switch to Profiles [1] to create one)` in place of the list.
- Models: when `len(p.files) == 0` and at least one root has `state == "scanned"`, render `(no .gguf files in configured search paths — edit ~/.config/llama-cpp-loader/config.toml)`.

**Out of scope**
- Reworking Profiles' existing empty state.

### Technical Approach
1. Add empty-state branches in each page's `View()`. Use `theme.Subtitle.Render(...)`.

### Touchpoints
- `internal/ui/pages/monitor.go`.
- `internal/ui/pages/launcher.go`.
- `internal/ui/pages/models.go`.

### Acceptance Criteria
- [ ] Monitor with zero instances renders the empty-state line.
- [ ] Launcher with zero profiles renders the empty-state line.
- [ ] Models after scan with zero files renders the empty-state line.

### Dependencies
- None.

---

## UIUX-009: Stderr notice at quit when background instances remain

### Scope
**In scope**
- After `tea.Quit` tears down the TUI, if `processmgr.Manager.List()` reports `len > 0`, print one line to `os.Stderr`.

**Out of scope**
- In-TUI confirm modal.

### Technical Approach
1. In `cmd/llama-cpp-loader/main.go`, immediately after `program.Run()` returns, query the manager and `fmt.Fprintln(os.Stderr, ...)` when applicable.
2. Phrasing: `"N background llama-server instance(s) still running — use 'pkill llama-server' or restart the TUI to manage them."`

### Touchpoints
- `cmd/llama-cpp-loader/main.go`.

### Acceptance Criteria
- [ ] Quitting with one running instance writes the notice to stderr.
- [ ] Quitting with zero running instances writes nothing.

### Dependencies
- None.

---

## UIUX-010: Document key-case convention in `?` help

### Scope
**In scope**
- Edit the `HelpMarkdown` constant in `internal/ui/components/help.go` to add one sentence under the Global section.

**Out of scope**
- Renaming any existing keybindings.

### Technical Approach
1. Append to the Global section of `HelpMarkdown`:
   `\n\n_Convention: lowercase keys are light/cheap actions; uppercase keys are heavy or destructive (e.g. ` + "`R`" + ` rescan walks the filesystem, ` + "`L`" + ` launches a process)._`

### Touchpoints
- `internal/ui/components/help.go`.

### Acceptance Criteria
- [ ] Rendering the help modal contains the substring `Convention: lowercase`.

### Dependencies
- None.

---

## UIUX-012: Tab strip separator

### Scope
**In scope**
- Insert a dim `│` between tabs in `RootModel.renderTabs()`.

**Out of scope**
- Restyling `TabActive`/`TabInactive` themselves.

### Technical Approach
1. Build an interleaved `[]string` of tab parts and separators in `renderTabs`. Separator: `theme.Subtitle.Render(" │ ")`.

### Touchpoints
- `internal/ui/root.go`.

### Acceptance Criteria
- [ ] The rendered top tab strip contains `│` between adjacent tab labels.

### Dependencies
- UIUX-008.

---

## UIUX-013: Picker responsive width

### Scope
**In scope**
- Replace hardcoded `theme.Pane.Width(80)` in `components/picker.go View()` with `min(m.width - 4, 120)`.
- Recompute row column widths proportionally to the new total inside `View()` row formatter.

**Out of scope**
- Rewriting the row format from `fmt.Sprintf` to a styled table.

### Technical Approach
1. Compute `boxW := min(m.width-4, 120)` (with a `60` floor).
2. Derive: `nameW := boxW / 3`, `quantW := 8`, `paramsW := 6`, `pathW := boxW - nameW - quantW - paramsW - 8`.
3. Replace the fixed-width `%-36s %-8s %-6s %s` with the computed widths.

### Touchpoints
- `internal/ui/components/picker.go`.

### Acceptance Criteria
- [ ] On `tea.WindowSizeMsg{Width: 200}`, the picker box width is `120` (capped) and rows fill the box.
- [ ] On `tea.WindowSizeMsg{Width: 70}`, the picker box width is `66` and rows do not visually overflow.

### Dependencies
- None.

---

## UIUX-014: Picker navigation keys

### Scope
**In scope**
- Add bindings for `pgup` / `pgdown`, `home` / `end`, `g` / `G` to `defaultPickerKeys` and route them in `Update`.

**Out of scope**
- Configurable page size.

### Technical Approach
1. Add `PgUp`, `PgDn`, `Home`, `End` fields to `pickerKeys`.
2. In `Update tea.KeyMsg`, handle each:
   - PgUp: `m.cursor = max(0, m.cursor - 10)`.
   - PgDn: `m.cursor = min(len(m.filtered)-1, m.cursor + 10)`.
   - Home / `g`: `m.cursor = 0`.
   - End / `G`: `m.cursor = len(m.filtered) - 1`.

### Touchpoints
- `internal/ui/components/picker.go`.

### Contracts
```go
type pickerKeys struct {
    Up, Down, Enter, Esc, Filter, Backspace, PgUp, PgDn, Home, End key.Binding
}
```

### Acceptance Criteria
- [ ] With 100 entries loaded and cursor at 0, pressing PgDn moves cursor to 10.
- [ ] Pressing End moves cursor to 99.

### Dependencies
- None.

---

## UIUX-017: Models status line wrap / truncate

### Scope
**In scope**
- In `pages/models.go renderStatus`, when the joined width exceeds `p.width - 2`, render one label per line.
- Truncate root path with leading ellipsis (`…/big`) when the path itself exceeds `40` cols.

**Out of scope**
- Hideable status section.

### Technical Approach
1. Build the per-path label list (already done).
2. Sum `lipgloss.Width(label)` + `2` per separator and check against `p.width - 2`.
3. If over, return `strings.Join(labels, "\n")`. Otherwise the existing two-space join.
4. Add `truncFront(s, n) string` returning `"…" + s[len(s)-(n-1):]` when over.

### Touchpoints
- `internal/ui/pages/models.go`.

### Acceptance Criteria
- [ ] On `p.width = 60` with three roots, `renderStatus` returns a string with two `\n` separators (one per pair of labels).
- [ ] A 100-char path renders truncated to `40` chars with leading `…`.

### Dependencies
- None.

---

## UIUX-020: Flash auto-clear

### Scope
**In scope**
- Stamp `flashAt time.Time` when `flash` is set in `pages/profiles.go`, `pages/launcher.go`, `pages/models.go`.
- Render `flash` dimmed via `theme.Subtitle` after 5 s elapsed.
- Clear `flash` after 15 s by a `tea.Tick` cmd dispatched at flash-set time.

**Out of scope**
- Reusing the global `components.StatusBar` for these flashes.

### Technical Approach
1. Define `flashClearMsg struct{}`; pages embed an `at time.Time` field next to `flash`.
2. Helper `setFlash(p, msg string) tea.Cmd` returning `tea.Tick(15*time.Second, ...)` that emits `flashClearMsg`.
3. In each page's `Update`, on `flashClearMsg`, only clear when `time.Since(p.flashAt) >= 15*time.Second` (guards against races with newer flash).
4. In each `View`, when `time.Since(p.flashAt) >= 5*time.Second`, render with `theme.Subtitle.Faint(true)`.

### Touchpoints
- `internal/ui/pages/profiles.go`.
- `internal/ui/pages/launcher.go`.
- `internal/ui/pages/models.go`.

### Acceptance Criteria
- [ ] After `setFlash`, the page's `View()` contains the message immediately.
- [ ] After a simulated 15 s tick, the flash is cleared from `View()`.
- [ ] A new `setFlash` within 5 s of an old one resets the timer (the clear from the old timer does not erase the new flash).

### Dependencies
- None.

---

## UIUX-021: Inline `huh` numeric validators

### Scope
**In scope**
- Wire `huh.Input.Validate(...)` for `NGL`, `CtxSize`, `BatchSize`, `UBatchSize`, `Port` in `pages/profiles_editor.go buildEditorForm`.

**Out of scope**
- Cross-field validators (kept in the existing `validator.Validate` footer block).

### Technical Approach
1. Define small validator funcs:
```go
func intRange(min, max int) func(string) error
func portValidator() func(string) error
```
2. Apply them to the relevant `huh.NewInput()`.
3. Allow empty strings for `BatchSize` / `UBatchSize` (current code only sets the field when parseable, so empty stays empty — preserve that semantics).

### Touchpoints
- `internal/ui/pages/profiles_editor.go`.
- `internal/ui/pages/profiles_test.go`.

### Acceptance Criteria
- [ ] Typing `abc` into NGL surfaces an inline huh error message.
- [ ] Typing `99999` into Port surfaces an inline huh error message (out of range).
- [ ] Typing nothing into `BatchSize` validates clean.

### Dependencies
- None.

---

## UIUX-022: Distinct sparkline colors per metric

### Scope
**In scope**
- Wrap tokens/s sparkline in `theme.OK`, req/s in `theme.Warn` in `pages/monitor.go SubViewMetrics`.

**Out of scope**
- Per-instance configurable colors.

### Technical Approach
1. In `SubViewMetrics` rendering, change:
```go
fmt.Fprintf(&b, "tokens/s: %s\n", theme.OK.Render(components.Sparkline(st.mets.TokensPerSec, 40)))
fmt.Fprintf(&b, "req/s   : %s\n", theme.Warn.Render(components.Sparkline(st.mets.RequestsPerSec, 40)))
```

### Touchpoints
- `internal/ui/pages/monitor.go`.

### Acceptance Criteria
- [ ] `SubViewMetrics` rendered output contains an ANSI escape for `theme.OK` around the tokens/s line and `theme.Warn` around the req/s line.

### Dependencies
- UIUX-008.

---

## UIUX-023: "Showing N of M" footer on Monitor logs

### Scope
**In scope**
- Append a dim `— showing last 10 of N (Space pauses, buffer 2000)` line below the log slice when `len(st.logs) > 10` in `SubViewLogs`.

**Out of scope**
- Configurable visible-line count.

### Technical Approach
1. Add to the `SubViewLogs` branch after building `bottom`:
```go
if len(st.logs) > 10 {
    bottom += "\n" + theme.Subtitle.Render(fmt.Sprintf("— showing last 10 of %d (Space pauses, buffer 2000)", len(st.logs)))
}
```

### Touchpoints
- `internal/ui/pages/monitor.go`.

### Acceptance Criteria
- [ ] With 11 simulated log lines, the rendered Monitor `View()` contains `showing last 10 of 11`.
- [ ] With 5 lines, the substring is absent.

### Dependencies
- None.

---

## UIUX-016: Reveal path → clipboard copy

### Scope
**In scope**
- Replace the `case "reveal":` action in `ModelsPage.commitRootAction` with a clipboard write using `atotto/clipboard`.
- Update the action menu label to `Copy path to clipboard`.

**Out of scope**
- `xdg-open` / `open -R` integration.

### Technical Approach
1. Import `github.com/atotto/clipboard`.
2. In `commitRootAction("reveal", path)`:
```go
if err := clipboard.WriteAll(path); err != nil {
    p.flash = "clipboard error: " + err.Error()
} else {
    p.flash = "path copied to clipboard"
}
p.action = nil
return p, nil
```
3. Update the option label in `handleKey` from `"Reveal path"` → `"Copy path to clipboard"`. Keep the value `"reveal"` for stable JSON test golden files (or change to `"copy"`; document the choice in the diff).

### Touchpoints
- `internal/ui/pages/models.go`.
- `go.mod` — promote `atotto/clipboard` from indirect to direct.
- `internal/ui/pages/models_test.go`.

### Acceptance Criteria
- [ ] After committing the action, `p.flash` reads `path copied to clipboard`.
- [ ] When `clipboard.WriteAll` fails, `p.flash` reads `clipboard error: …`.

### Dependencies
- None.

---

## UIUX-011: Pane border on Launcher's split

### Scope
**In scope**
- Wrap the left list and right detail panels of `LauncherPage.View()` in `theme.Pane`, mirroring `ProfilesPage`'s split.

**Out of scope**
- Pane-wrapping Monitor or Models.

### Technical Approach
1. Replace:
```go
body := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
```
   with:
```go
left := theme.Pane.Width(p.width / 2).Render(p.plist.View())
right := theme.Pane.Width(p.width/2 - 2).Render(rightContent)
body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
```

### Touchpoints
- `internal/ui/pages/launcher.go`.

### Acceptance Criteria
- [ ] Launcher `View()` output contains the rounded-border characters used by `theme.Border` around both panels.
- [ ] No visual regression on the running list / footer regions (still rendered below the split).

### Dependencies
- UIUX-008.
