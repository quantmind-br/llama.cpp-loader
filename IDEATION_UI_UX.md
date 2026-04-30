# UI/UX Improvements — Decisions of Record

**Project:** llama-cpp-loader (Go TUI · bubbletea / lipgloss / huh)
**Generated:** 2026-04-29 (initial analysis), updated post-critical-review.
**Scope:** `internal/ui/**` (root, pages, components, theme).

> This file is the **decisions log**. Each entry below is cleared for
> implementation in the indicated phase. Detailed scope, contracts,
> touchpoints, and acceptance criteria live in
> [`PRD_IDEATION_UI_UX.md`](./PRD_IDEATION_UI_UX.md).

---

## Decision Summary

| Status | Count | IDs |
|--------|-------|-----|
| ✅ Kept | 18 | 001, 002, 003, 004, 005, 007, 008, 012, 013, 014, 015, 017, 018, 020, 021, 022, 023, 024 |
| ⚠️ Adjusted (scope refined) | 5 | 006, 009, 010, 011, 016 |
| ❌ Dropped | 3 | 019, 025, 026 |
| 🕓 Deferred | 0 | — |

---

## Phase 1 — Foundation (theme + structural)

Touch shared style and the page→root contract. Land first; later phases
build on them.

### UIUX-008 — Adaptive theme + `NO_COLOR`
**Priority:** High **Effort:** Medium
Replace every hardcoded hex color in `internal/ui/theme/theme.go` with
`lipgloss.AdaptiveColor{Light: …, Dark: …}`. Honor the `NO_COLOR` env var at
theme init: when set, build styles with cleared `Foreground` / `Background`
and rely on `Bold` + glyphs for emphasis. Light-terminal users and
screen-reader users on `NO_COLOR=1` get a usable theme.

### UIUX-024 — Modal background adaptive
**Priority:** Medium **Effort:** Trivial
Drop the hardcoded `Background("#1a1a1a")` on `modalBox` in
`internal/ui/components/modal.go`. Let the terminal background show
through; the accent border + padding remain the visual demarcation.

### UIUX-007 — `HintProvider` contract collapses footer / status duplication
**Priority:** High **Effort:** Medium
Add an optional `HintProvider interface { Hints() string }` to the page
contract. Root reads `m.pages[m.active].(HintProvider)` after each tab
activation / key event and updates `m.status.Hints` to
`globalHints + " | " + pageHints`. Pages stop rendering their own footer
strings. Auto-resolves the duplicate `[?] help` token previously tracked
as UIUX-025 (now dropped).

---

## Phase 2 — Safety + Information Surfacing

### UIUX-001 — Populate Monitor row columns (Uptime / VRAM / Tokens-per-sec)
**Priority:** High **Effort:** Small
Wire `applyInstances` in `internal/ui/pages/monitor.go` to read live
`subState.gpu` and `subState.mets` plus a per-instance start timestamp
and fill the three `"--"` columns on each refresh. Render `"--"` only
when there is no sample yet. Requires a `StartedAt time.Time` field on
`domain.RunningInstance` if not already present.

### UIUX-002 — Confirmation on destructive actions
**Priority:** High **Effort:** Small
Add `huh.NewConfirm` (matching the existing pattern in
`profiles.askDeleteSelected`) to:
- `pages/launcher.go` — `[k]` kill last-launched.
- `pages/monitor.go` — `[k]` kill selected.
- Profile editor — `esc` when the draft differs from a snapshot taken at
  form open.

### UIUX-003 — `WaitHealthy` waiting indicator
**Priority:** High **Effort:** Small
On `launchedMsg`, set status to
`"pid=N port=P — waiting for /health…"` and append an animated spinner
glyph (use `bubbles/spinner` driven by `tea.Tick`). Clear on
`healthyMsg` / `launchErrMsg`.

### UIUX-005 — Pause indicator visible in Monitor logs
**Priority:** Medium **Effort:** Trivial
Render `Logs (PAUSED — Space to resume)` styled `theme.Warn` above the
log region when `p.paused && p.subView == SubViewLogs`.

### UIUX-006 — Color-code crashed / corrupt rows ⚠️ Adjusted
**Priority:** High **Effort:** Medium
**Adjusted scope:** the original bundled monitor + profiles into one
item. Split into two implementable sub-tasks:
- **006a — Monitor crashed rows.** Pre-style each row cell with
  `theme.Error.Render(...)` before `tbl.SetRows`. Bubbles' table renders
  ANSI on cell strings. Trivial.
- **006b — Profiles corrupt rows.** Bubbles' `list.DefaultDelegate`
  strips ANSI from the title returned by `Item.Title()`. Implement a
  custom `list.ItemDelegate` that detects `corruptItem` and renders
  with `theme.Warn` / `Error` and a different bullet. Small but real
  piece of code.

### UIUX-015 — Picker surfaces scan errors
**Priority:** Medium **Effort:** Trivial
Capture `PickerScanStartedMsg.Err` into a new `m.err string` and render
an error line styled `theme.Error` under the picker title when set.

---

## Phase 3 — Discoverability

### UIUX-004 — Monitor sub-view tab strip
**Priority:** High **Effort:** Small
Render a tab strip above the bottom region in `pages/monitor.go`
mirroring the top tab strip: active sub-view styled `theme.TabActive`,
inactive `TabInactive`, separator a dim pipe. `[v]` continues to cycle.

### UIUX-018 — Consistent empty states with next-action guidance
**Priority:** Medium **Effort:** Trivial
Add empty states modeled on the existing Profiles one:
- Monitor: `(no instances running — switch to Launcher [2] to start one)`
- Launcher (when no profiles): `(no profiles yet — switch to Profiles [1] to create one)`
- Models (when scan complete with zero results):
  `(no .gguf files in configured search paths — edit ~/.config/llama-cpp-loader/config.toml)`

### UIUX-009 — Stderr notice on quit when bg instances survive ⚠️ Adjusted
**Priority:** Medium **Effort:** Small
**Adjusted scope:** drop the in-TUI confirm-modal option from the
original. On `tea.Quit`, when the process manager reports
`len(running) > 0`, print a single-line message to `os.Stderr` after
the TUI tears down. No prompt. First-time users get the mental model
post-exit; power users keep fast quit.

### UIUX-010 — Document letter-case convention in `?` help ⚠️ Adjusted
**Priority:** Low **Effort:** Trivial
**Adjusted scope:** no binding changes. The original proposed
unification of `R` vs `r`. Instead, add one sentence to
`components/help.go` HelpMarkdown:
*"Convention: lowercase keys are light/cheap actions; uppercase keys are
heavy or destructive (e.g., `R` rescan walks the filesystem, `L`
launches a process)."*

---

## Phase 4 — Polish

### UIUX-012 — Tab strip separator
**Priority:** Low **Effort:** Trivial
Insert a dim `│` separator between tabs in `RootModel.renderTabs()` so
inactive tabs don't run together visually.

### UIUX-013 — Picker responsive width
**Priority:** Medium **Effort:** Small
In `components/picker.go View()`, replace `theme.Pane.Width(80)` with
`min(m.width - 4, 120)`. Recompute the row column widths proportionally
from the new total.

### UIUX-014 — Picker navigation keys
**Priority:** Medium **Effort:** Trivial
Add PgUp / PgDn (page = 10 rows), Home / End, vim-style `g` / `G` to
the picker's `defaultPickerKeys` and route them in `Update`.

### UIUX-017 — ModelsPage status line wrap / truncate
**Priority:** Medium **Effort:** Small
When the joined per-path label width exceeds `p.width - 2`, switch to
one label per line. Truncate each root path with leading ellipsis
(`…/big`) — preserve the tail, which is the discriminator.

### UIUX-020 — Flash auto-clear
**Priority:** Medium **Effort:** Small
Stamp `flashAt time.Time` when `flash` is set on `pages/profiles.go`,
`pages/launcher.go`, `pages/models.go`. Render dimmed after 5 s; clear
after 15 s via a `tea.Tick` cmd issued at the moment flash is set.

### UIUX-021 — Inline `huh` validators for numeric editor fields
**Priority:** Medium **Effort:** Small
Wire `huh.Input.Validate(func(string) error { … })` for `NGL`,
`CtxSize`, `BatchSize`, `UBatchSize`, `Port` in
`pages/profiles_editor.go buildEditorForm`. Keep the cross-field
validator footer for rules `huh` can't express per-field.

### UIUX-022 — Distinct sparkline colors per metric
**Priority:** Low **Effort:** Trivial
Wrap tokens/s sparkline in `theme.OK`, req/s in `theme.Warn`, in
`pages/monitor.go SubViewMetrics`.

### UIUX-023 — "Showing N of M" footer on Monitor logs
**Priority:** Low **Effort:** Trivial
Append a dimmed `— showing last 10 of N (Space pauses, buffer 2000)`
line below the log slice when `len(st.logs) > 10`.

### UIUX-016 — Reveal path → clipboard copy ⚠️ Adjusted
**Priority:** Low **Effort:** Small
**Adjusted scope:** drop the `xdg-open` / `open -R` cross-platform
branch from the original. Implement clipboard copy only, via
`atotto/clipboard` (already an indirect dependency in go.sum). Flash:
`path copied to clipboard`. Action menu label: `Copy path to clipboard`.

### UIUX-011 — Pane border on Launcher split ⚠️ Adjusted
**Priority:** Low **Effort:** Small
**Adjusted scope:** the original asked for Pane wrapping on every page.
Apply only to Launcher's existing left/right split (parity with
Profiles). Monitor and Models stay full-bleed — they have no split that
benefits from a divider, and Pane padding eats columns on narrow
terminals.

---

## Deferred

(none)

---

## Dropped on Critical Review

| ID | Name | Reason |
|----|------|--------|
| UIUX-019 | `/` filter collision in ProfilesPage | The two `/` filter UIs live on mutually-exclusive screens (list mode vs. editor advanced sub-tab gated by `p.editing`). The "collision" is theoretical, not actual. |
| UIUX-025 | Duplicate `[?] help` token | Auto-resolved when UIUX-007 (`HintProvider`) lands — page footers go away entirely, so the duplicated token disappears. Standalone item would be redundant. |
| UIUX-026 | Placeholder "coming soon" diagnostic | Developer-facing concern. Production path always wires all four pages in `cmd/llama-cpp-loader/main.go`; the failure mode would surface in CI/code review long before reaching a user. |
