# PRD — UI Pages Refactor

**Source**: IDEATION_CODE_QUALITY.md
**Generated**: 2026-04-30

## Implementation Order

1. CQ-008 — Move monitor-event apply onto `subState`
2. CQ-001 — Extract generic `Confirm` dialog component
3. CQ-004 — Decompose `Update` methods across pages into per-message handlers
4. CQ-010 — Split `ModelsPage.handleKey` filter-mode handling
5. CQ-007 — Extract per-page `View()` rendering helpers
6. CQ-003 — Decompose `MonitorPage.Update` and `applyInstances`
7. CQ-002 — Extract `ProfileEditor` sub-model from `ProfilesPage`

---

## CQ-008: Move monitor-event apply onto `subState`

### Scope
**In scope**:
- Add `Apply(ev monitor.MonitorEvent, paused bool)` method on `*subState` in `internal/ui/pages/monitor.go`.
- Replace the 5-branch `case monitor.SourceX` block at `monitor.go:217–261` with a single call to `st.Apply(m.ev, p.paused)`.
- Add a dedicated test for `subState.Apply` covering all five `Data` types and the `paused` log-skip path.

**Out of scope**:
- Changing the `monitorEventMsg` envelope or `monitor.MonitorEvent` shape.
- Renaming `monitor.SourceXxx` constants (they remain used elsewhere).

### Technical Approach
- Replace the type-assertion-per-case pattern with a Go type switch on `ev.Data`. The existing `Source` enum becomes redundant for dispatch but stays in `MonitorEvent` for telemetry/debug.
- Preserve current behavior for the log path: when `paused` is true, do not append to `s.logs`; cap remains 2000 lines.

```go
func (s *subState) Apply(ev monitor.MonitorEvent, paused bool) {
    switch d := ev.Data.(type) {
    case monitor.LogLine:
        if !paused {
            s.logs = append(s.logs, d.Line)
            if len(s.logs) > 2000 {
                s.logs = s.logs[len(s.logs)-2000:]
            }
        }
    case monitor.SlotSnapshot:
        s.slots = d
    case monitor.GPUStats:
        s.gpu = d
    case monitor.HealthStatus:
        s.health = d
    case monitor.Metrics:
        s.mets = d
    }
}
```

### Touchpoints
- `internal/ui/pages/monitor.go` — add `subState.Apply`; replace `case monitorEventMsg` body.
- `internal/ui/pages/monitor_test.go` — add `TestSubState_Apply_*` covering each `Data` variant and the `paused` branch.

### Contracts
```go
// internal/ui/pages/monitor.go
func (s *subState) Apply(ev monitor.MonitorEvent, paused bool)
```

### Acceptance Criteria
- [ ] `MonitorPage.Update`'s `case monitorEventMsg` body is ≤10 lines and contains no `interface{}` type assertions on `ev.Data`.
- [ ] `subState.Apply` is unit-tested for all five `monitor.Source*` variants.
- [ ] Existing `monitor_test.go` tests continue to pass without behavioral changes.
- [ ] `go vet ./...` clean.

### Dependencies
- None.

---

## CQ-001: Extract generic `Confirm` dialog component

### Scope
**In scope**:
- New file `internal/ui/components/confirm.go` exposing a `Confirm` value-type wrapping `*huh.Form` + answer pointer + payload.
- New file `internal/ui/components/confirm_test.go` covering: yes/no resolution, completion-fires-callback, non-`KeyMsg` forwarding, `Active()` gating.
- Migrate the five existing duplicates:
  - `ProfilesPage` — delete-confirm
  - `ProfilesPage` — discard-confirm
  - `LauncherPage` — kill-confirm
  - `MonitorPage` — kill-confirm
  - `MonitorPage` — restart-confirm

**Out of scope**:
- Generalizing to multi-button forms (Confirm is yes/no only).
- Replacing `components.Modal` (different abstraction; stays).

### Technical Approach
- `Confirm` is a value type so it can be embedded by-value in pages with value receivers (`ProfilesPage`, `LauncherPage`). The `*huh.Form` and `*bool` answer pointer are heap-allocated, matching the existing `confirmAnswer *bool` rationale documented in `profiles.go:24–28`.
- `Update` returns a new `Confirm` value (not `*Confirm`) so usage matches bubbletea's value-copy idiom.
- On `huh.StateCompleted`, `Update` invokes `onYes(payload)` if the answer is `true`, and clears the form (sets `c.form = nil`). Caller checks `c.Active()` to know when the dialog finished.
- Non-`tea.KeyMsg` messages must be forwarded to the inner form so huh's internal Cmd→Msg handshake (focus init, button styling refresh) completes — mirrors the project invariant in `CLAUDE.md`.

### Touchpoints
- `internal/ui/components/confirm.go` — new.
- `internal/ui/components/confirm_test.go` — new.
- `internal/ui/pages/profiles.go` — replace `confirmDelete`/`confirmForm`/`confirmAnswer` and `confirmDiscardForm`/`confirmDiscardAnswer` with two `Confirm` values; remove `askConfirm`/`updateConfirm`/`finalizeConfirm` boilerplate.
- `internal/ui/pages/launcher.go` — replace `confirmKillForm`/`confirmKillAnswer`/`confirmKillTargetID` with one `Confirm`; remove `askConfirmKill`/`updateConfirmKill`/`finalizeConfirmKill`.
- `internal/ui/pages/monitor.go` — same replacement for kill-confirm and restart-confirm.
- Tests: update `profiles_test.go`, `launcher_test.go`, `monitor_test.go` to invoke through the new component.

### Contracts
```go
// internal/ui/components/confirm.go
package components

type Confirm struct {
    form    *huh.Form
    answer  *bool
    payload any
    onYes   func(any) tea.Cmd
}

// NewConfirm builds a yes/no dialog. onYes is invoked with payload when the
// user confirms; nil onYes is allowed (caller polls Active()).
func NewConfirm(title string, payload any, onYes func(any) tea.Cmd) Confirm

func (c Confirm) Active() bool                          // form != nil
func (c Confirm) View() string                          // delegates to form
func (c Confirm) Init() tea.Cmd                         // delegates to form
func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) // forwards + handles completion
```

### Acceptance Criteria
- [ ] `internal/ui/components/confirm.go` exists and compiles.
- [ ] `confirm_test.go` covers: yes resolution invokes `onYes(payload)`, no resolution does not, non-`KeyMsg` is forwarded, `Active()` returns `false` after completion.
- [ ] Zero references to `huh.NewConfirm` remain in `internal/ui/pages/`.
- [ ] `IsCapturingInput()` in `ProfilesPage`, `LauncherPage`, `MonitorPage` reads exclusively from `Confirm.Active()` (plus `editor.Active()`/`pickerActive` where applicable).
- [ ] All existing `profiles_test.go`, `launcher_test.go`, `monitor_test.go` cases pass.
- [ ] No references to the old field names (`confirmKillForm`, `confirmRestartForm`, `confirmDeleteForm`, `confirmDiscardForm`, `confirm*Answer`, `confirm*TargetID`) remain in the tree.

### Dependencies
- None.

---

## CQ-004: Decompose `Update` methods across pages into per-message handlers

### Scope
**In scope**:
- `ProfilesPage.Update` (`profiles.go:182`, 141 lines)
- `LauncherPage.Update` (`launcher.go:138`, 141 lines)
- `MonitorPage.Update` (`monitor.go:153`, 139 lines)
- `ModelPicker.Update` (`components/picker.go:124`, 108 lines)

**Out of scope**:
- Changing message types or external behavior.
- Touching `models.go` (covered by CQ-010).

### Technical Approach
- Adopt the pattern already in `internal/ui/root.go:181–266`: each `case` in the type-switch becomes a method (`handleResize`, `handleKey`, `handleEventX`, …).
- `Update` becomes a thin dispatcher (≤20 lines).
- Forwarding rules unchanged — non-`KeyMsg` messages still flow to live forms (now via `Confirm.Update`).

```go
func (p MonitorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m := msg.(type) {
    case tea.WindowSizeMsg:            return p.handleResize(m)
    case tea.KeyMsg:                    return p.handleKey(m)
    case monitorInstancesRefreshedMsg: return p.handleInstancesRefreshed(m)
    case monitorEventMsg:               return p.handleMonitorEvent(m)
    case restartResultMsg:              return p.handleRestartResult(m)
    case monitorPeriodicTickMsg:        return p.handlePeriodicTick()
    }
    return p.forwardToConfirms(msg)
}
```

For each handler, target ≤30 lines.

### Touchpoints
- `internal/ui/pages/profiles.go`
- `internal/ui/pages/launcher.go`
- `internal/ui/pages/monitor.go`
- `internal/ui/components/picker.go`

### Contracts
Each page exposes private helpers named `handle<MsgType>(m <MsgType>) (tea.Model, tea.Cmd)`. Naming convention matches `root.go`.

### Acceptance Criteria
- [ ] No `Update` method in `internal/ui/pages/` or `internal/ui/components/` exceeds 25 lines.
- [ ] Each handler method is ≤40 lines.
- [ ] All existing page-level tests pass without modification.
- [ ] `go vet ./...` clean.

### Dependencies
- CQ-001 (confirm forwarding goes through `Confirm.Update`, so handler bodies are simpler).

---

## CQ-010: Split `ModelsPage.handleKey` filter-mode handling

### Scope
**In scope**:
- Extract the first 30 lines of `ModelsPage.handleKey` (`models.go:400–490`, 91 lines total) into `handleFilterKey(msg tea.KeyMsg) (handled bool, model tea.Model, cmd tea.Cmd)`.
- `handleKey` calls it first; if `handled`, returns; otherwise falls through to command keys.

**Out of scope**:
- Changing filter semantics or key bindings.

### Technical Approach
- The filter branch is currently gated by `if p.filterMode { ... }`. Extract that block as a separate method that returns whether it consumed the event.
- The remainder of `handleKey` (the command-key switch) becomes the post-filter path.

### Touchpoints
- `internal/ui/pages/models.go` — split `handleKey` into `handleFilterKey` + reduced `handleKey`.
- `internal/ui/pages/models_test.go` — no changes if behavior is preserved.

### Contracts
```go
// internal/ui/pages/models.go
func (p ModelsPage) handleFilterKey(msg tea.KeyMsg) (handled bool, m tea.Model, cmd tea.Cmd)
func (p ModelsPage) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) // ≤55 lines after split
```

### Acceptance Criteria
- [ ] `handleKey` is ≤55 lines.
- [ ] `handleFilterKey` is ≤40 lines.
- [ ] `models_test.go` filter-mode tests pass unchanged.

### Dependencies
- None (independent of CQ-004 but follows the same spirit).

---

## CQ-007: Extract per-page `View()` rendering helpers

### Scope
**In scope**:
- `LauncherPage.View` (`launcher.go:389–455`, 66 lines)
- `MonitorPage.View` (`monitor.go:541–602`, 61 lines)

Per page, extract three helpers:
- `renderProfileDetail() string` (or equivalent for the right pane)
- `renderRunningList() string` (Launcher) or `renderTable() string` (Monitor)
- `renderStatusLine() string`

**Out of scope**:
- Changing layout, styling, or output text.
- Touching `ProfilesPage.View` or `ModelsPage.View` (already short).

### Technical Approach
- Pure mechanical extraction. Each helper takes no args (reads from receiver) and returns its rendered slice. `View` composes via `lipgloss.JoinVertical`.

### Touchpoints
- `internal/ui/pages/launcher.go`
- `internal/ui/pages/monitor.go`

### Contracts
```go
func (p LauncherPage) renderProfileDetail() string
func (p LauncherPage) renderRunningList() string
func (p LauncherPage) renderStatusLine() string

func (p *MonitorPage) renderTable() string
func (p *MonitorPage) renderSubViewBody() string
func (p *MonitorPage) renderStatusLine() string
```

### Acceptance Criteria
- [ ] `LauncherPage.View` and `MonitorPage.View` are each ≤25 lines.
- [ ] Golden/snapshot output (where tests assert rendered strings) is byte-identical to before.

### Dependencies
- None.

---

## CQ-003: Decompose `MonitorPage.Update` and `applyInstances`

### Scope
**In scope**:
- `MonitorPage.Update` event-routing block (post-CQ-004) — already shrinks to a dispatcher.
- `MonitorPage.applyInstances` (`monitor.go:455–537`, 84 lines) split into three methods.
- Collapse the two near-identical cancel-loops at `monitor.go:521–537` into one.

**Out of scope**:
- Changing `monitor.Manager.Subscribe` contract.
- Modifying subscription teardown semantics (still async via goroutines).

### Technical Approach

Split `applyInstances` into three single-purpose methods:

```go
func (p *MonitorPage) applyInstances(insts []domain.RunningInstance) tea.Cmd {
    p.tbl.SetRows(p.renderRows(insts))
    p.clampCursor(len(insts))
    cmds := p.ensureSubscriptions(insts)
    p.reapDeadSubscriptions(insts)
    return tea.Batch(cmds...)
}

func (p *MonitorPage) renderRows(insts []domain.RunningInstance) []table.Row
func (p *MonitorPage) ensureSubscriptions(insts []domain.RunningInstance) []tea.Cmd
func (p *MonitorPage) reapDeadSubscriptions(insts []domain.RunningInstance)
```

`reapDeadSubscriptions` collapses the two cancel-loops:

```go
byPID := make(map[int]domain.RunningInstance, len(insts))
for _, ri := range insts { byPID[ri.PID] = ri }
for pid, st := range p.subs {
    inst, present := byPID[pid]
    if present && !inst.Crashed { continue }
    cancel := st.cancel
    go func() { _ = cancel() }()
    delete(p.subs, pid)
    delete(p.chans, pid)
}
```

### Touchpoints
- `internal/ui/pages/monitor.go` — split `applyInstances`; integrate with CQ-004's `handleInstancesRefreshed`.
- `internal/ui/pages/monitor_test.go` — add per-helper tests for `renderRows`, `ensureSubscriptions`, `reapDeadSubscriptions`.

### Contracts
```go
func (p *MonitorPage) renderRows(insts []domain.RunningInstance) []table.Row
func (p *MonitorPage) ensureSubscriptions(insts []domain.RunningInstance) []tea.Cmd
func (p *MonitorPage) reapDeadSubscriptions(insts []domain.RunningInstance)
func (p *MonitorPage) clampCursor(rowCount int)
```

### Acceptance Criteria
- [ ] `applyInstances` is ≤15 lines.
- [ ] No method in `MonitorPage` exceeds 40 lines.
- [ ] Subscription-leak test: launching N instances then crashing all of them results in zero entries in `p.subs` and zero entries in `p.chans` after one `applyInstances` call.
- [ ] Orphan-cleanup test: removing a PID from the input list (but not crashed) drops its sub.
- [ ] Existing `monitor_test.go` tests pass.

### Dependencies
- CQ-001, CQ-004, CQ-008.

---

## CQ-002: Extract `ProfileEditor` sub-model from `ProfilesPage`

### Scope
**In scope**:
- New package `internal/ui/pages/profile_editor/` with type `Editor` implementing `tea.Model`.
- Move into `Editor`: `huh.Form`, `draft *profileDraft`, `editorOpenSnapshot profileDraft`, `subTab`, `advanced table.Model`, `advancedAll []table.Row`, `advancedFilter string`, `filterMode bool`, and the discard-confirm (as a `components.Confirm` from CQ-001).
- `ProfilesPage` retains: master list, delete-confirm (`components.Confirm`), picker overlay, status flash, `width`/`height`.
- `ProfilesPage.IsCapturingInput()` aggregates: `p.editor.Active() || p.deleteConfirm.Active() || p.pickerActive`.
- `ProfilesPage.Update` dispatches to `Editor.Update` while `Editor.Active()` is true; commit/cancel paths bubble back via dedicated message types (`EditorCommittedMsg{draft profileDraft}`, `EditorCancelledMsg{}`).

**Out of scope**:
- Splitting list or picker into separate models (deferred to a future iteration once Editor extraction is proven).
- Changing the user-visible flow (master/detail layout, key bindings, validation messages).
- Touching `profiles_editor.go`'s `buildEditorForm` signature beyond its move into the new package.

### Technical Approach
- Receiver convention for `Editor` mirrors `ProfilesPage`: **value receivers** + heap-allocated `*profileDraft`. Documented in `CLAUDE.md` per CQ-012.
- `Editor.Active()` returns `true` while a form is being edited (the current `editing` bool, now private to Editor).
- Editor exposes `Open(draft profileDraft, schema domain.FlagSchema) (Editor, tea.Cmd)` and `Cancel() Editor` for state transitions; `Update(msg tea.Msg) (Editor, tea.Cmd)` for routing.
- Editor emits `EditorCommittedMsg` (via a returned `tea.Cmd`) when the user saves; `ProfilesPage` handles the message by persisting through `store.Save` and refreshing the list.
- `profileDraft` and `buildEditorForm` move from `internal/ui/pages/profiles_editor.go` to `internal/ui/pages/profile_editor/` as exported (`Draft`, `BuildForm`) or stay package-private if `ProfilesPage` only needs the message types. Prefer package-private; expose only `Editor`, `Open`, `Active`, `View`, `Update`, `EditorCommittedMsg`, `EditorCancelledMsg`.
- Picker integration: `ProfilesPage` opens the picker; on `ModelPickedMsg`, it forwards the chosen path into the editor via a new `Editor.SetModelPath(path string) Editor` method (or by re-opening the editor with an updated draft).

### Touchpoints
- `internal/ui/pages/profile_editor/editor.go` — new (Editor type, Update, View, Open, Active, message types).
- `internal/ui/pages/profile_editor/draft.go` — new (move `profileDraft`, `buildEditorForm`, `argString`, `flashAttnToString` from `profiles_editor.go`).
- `internal/ui/pages/profile_editor/editor_test.go` — new (cover open → save, open → discard, advanced filter typing).
- `internal/ui/pages/profiles.go` — replace 9+ editor-related fields with `editor profile_editor.Editor`; `Update` delegates to `editor.Update` when active.
- `internal/ui/pages/profiles_editor.go` — delete (contents moved).
- `internal/ui/pages/profiles_test.go` — adjust to new field layout; behavioral assertions unchanged.

### Contracts
```go
// internal/ui/pages/profile_editor/editor.go
package profile_editor

type Editor struct { /* form, draft *Draft, openSnapshot Draft, subTab, advanced table.Model,
                       advancedAll []table.Row, advancedFilter string, filterMode bool,
                       discardConfirm components.Confirm */ }

func New(schema domain.FlagSchema) Editor

func (e Editor) Active() bool
func (e Editor) Init() tea.Cmd
func (e Editor) View() string
func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd)

// Open starts editing the given draft. Returns the form's Init Cmd.
func (e Editor) Open(d Draft) (Editor, tea.Cmd)

// SetModelPath updates the in-flight draft's Model field (for picker integration).
func (e Editor) SetModelPath(path string) (Editor, tea.Cmd)

// Result messages bubbled out via tea.Cmd.
type EditorCommittedMsg struct { Draft Draft }
type EditorCancelledMsg struct{}
```

### Acceptance Criteria
- [ ] `internal/ui/pages/profile_editor/` package exists, compiles, has tests covering open/save/cancel/discard-confirm/advanced-filter.
- [ ] `ProfilesPage` struct has ≤12 fields (down from 22+).
- [ ] `ProfilesPage.IsCapturingInput()` body is ≤3 lines and reads exclusively from sub-model `Active()` calls plus `pickerActive`.
- [ ] `ProfilesPage.Update` routes to `editor.Update` while `editor.Active()` is true.
- [ ] `profiles.go` line count drops below 500.
- [ ] All existing `profiles_test.go` cases pass (with field-name updates only, no logic changes).
- [ ] No stale references to removed fields (`editing`, `subTab`, `form`, `draft`, `editorOpenSnapshot`, `advanced`, `advancedAll`, `advancedFilter`, `filterMode`, `confirmDiscardForm`, `confirmDiscardAnswer`) remain in `profiles.go`.

### Dependencies
- CQ-001 (Editor's discard-confirm uses `components.Confirm`).
- CQ-004 recommended (smaller diff: `ProfilesPage.Update` is already decomposed).
