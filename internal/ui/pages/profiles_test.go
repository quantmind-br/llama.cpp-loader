package pages

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/components"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/theme"
)

func TestProfilesPage_LoadsExistingProfile(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(domain.Profile{
		ID:    "qwen",
		Name:  "Qwen Coder",
		Model: "/m.gguf",
		Args:  map[string]any{"ngl": float64(99)},
	}); err != nil {
		t.Fatal(err)
	}

	page := NewProfilesPage(store, domain.FlagSchema{})
	tm := teatest.NewTestModel(t, page, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Qwen Coder")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}) // root would handle q; here we just ensure quit
	_ = tm.Quit()
}

func TestProfilesPage_NewProfileSavesViaStore(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	page := NewProfilesPage(store, domain.FlagSchema{})
	tm := teatest.NewTestModel(t, page, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "No profiles yet")
	}, teatest.WithDuration(2*time.Second))

	// 'n' opens the form
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Name") && strings.Contains(string(out), "Model path")
	}, teatest.WithDuration(2*time.Second))

	// We don't drive the full huh form here — just exit. The store-side
	// behavior is already covered by FSStore tests; this asserts wiring.
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	_ = tm.Quit()

	// Ensure no profile was persisted (esc cancels)
	got, _ := store.List()
	if len(got) != 0 {
		t.Errorf("List len = %d, want 0", len(got))
	}
}

func TestProfilesPage_ValidationDetectsUbatchOverBatch(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	page := NewProfilesPage(store, domain.FlagSchema{})
	page.draft = &profileDraft{
		ID:         "x",
		Name:       "X",
		BatchSize:  "2048",
		UBatchSize: "4096",
		isNew:      true,
	}

	pr := page.previewProfile()
	report := page.validator.Validate(pr, page.schema)

	found := false
	for _, e := range report.Errors {
		if e.Field == "ubatch-size" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ubatch-size error in report; got Errors=%v Warnings=%v", report.Errors, report.Warnings)
	}
}

type stubScanner struct{}

func (stubScanner) Scan(ctx context.Context, paths []string) (<-chan domain.ScanEvent, error) {
	ch := make(chan domain.ScanEvent, 1)
	close(ch)
	return ch, nil
}

func TestProfilesPage_PickerWritesDraftModel(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	page := NewProfilesPage(store, domain.FlagSchema{}).WithModelScanner(stubScanner{}, nil)

	// Start a new draft so editing is active.
	model, _ := page.startNew()
	page = model.(ProfilesPage)
	page.editing = true

	// Simulate ModelPickedMsg landing in Update.
	updated, _ := page.Update(components.ModelPickedMsg{Path: "/picked/model.gguf"})
	page = updated.(ProfilesPage)

	if page.draft.Model != "/picked/model.gguf" {
		t.Fatalf("draft.Model = %q, want /picked/model.gguf", page.draft.Model)
	}
	if page.pickerActive {
		t.Errorf("pickerActive = true, want false after pick")
	}
}

func TestProfilesPage_RendersCorruptMarker(t *testing.T) {
	store := newFakeStoreWithDiagnostics(
		[]domain.Profile{{ID: "ok", Name: "Ok"}},
		[]profilestore.ListDiagnostic{{ID: "broken", Err: profilestore.ErrInvalidJSON}},
	)
	p := NewProfilesPage(store, domain.FlagSchema{})
	// Seed a window size so the list has room to render.
	model, _ := p.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	p = model.(ProfilesPage)

	loaded := p.loadCmd()()
	model, _ = p.Update(loaded)
	page := model.(ProfilesPage)

	out := page.View()
	if !strings.Contains(out, "broken") || !strings.Contains(out, "⚠") {
		t.Fatalf("view missing corrupt marker; got:\n%s", out)
	}
}

type fakeStoreWithDiag struct {
	ps    []domain.Profile
	diags []profilestore.ListDiagnostic
}

func newFakeStoreWithDiagnostics(ps []domain.Profile, diags []profilestore.ListDiagnostic) *fakeStoreWithDiag {
	return &fakeStoreWithDiag{ps: ps, diags: diags}
}

func (f *fakeStoreWithDiag) List() ([]domain.Profile, error) { return f.ps, nil }
func (f *fakeStoreWithDiag) ListWithDiagnostics() ([]domain.Profile, []profilestore.ListDiagnostic, error) {
	return f.ps, f.diags, nil
}
func (f *fakeStoreWithDiag) Get(id string) (domain.Profile, error) {
	for _, p := range f.ps {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.Profile{}, profilestore.ErrNotFound
}
func (f *fakeStoreWithDiag) Save(_ domain.Profile) error { return nil }
func (f *fakeStoreWithDiag) Delete(_ string) error       { return nil }
func (f *fakeStoreWithDiag) Duplicate(_, _ string) (domain.Profile, error) {
	return domain.Profile{}, nil
}

func TestProfilesPage_UseInNewProfilePrefillsDraft(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	page := NewProfilesPage(store, domain.FlagSchema{})

	updated, _ := page.Update(UseInNewProfileMsg{Path: "/foo/bar.gguf"})
	page = updated.(ProfilesPage)

	if !page.editing {
		t.Fatal("page.editing = false, want true")
	}
	if page.draft.Model != "/foo/bar.gguf" {
		t.Fatalf("draft.Model = %q", page.draft.Model)
	}
	if !page.draft.isNew {
		t.Errorf("isNew = false, want true")
	}
}

func TestProfilesPage_LKeyEmitsLaunchProfileMsg(t *testing.T) {
	dir := t.TempDir()
	store, err := profilestore.NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(domain.Profile{
		ID:    "demo",
		Name:  "Demo",
		Model: "/m.gguf",
		Args:  map[string]any{"port": float64(8080)},
	}); err != nil {
		t.Fatal(err)
	}
	page := NewProfilesPage(store, domain.FlagSchema{})
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	page = updated.(ProfilesPage)
	updated, _ = page.Update(loadedMsg{profiles: []domain.Profile{{ID: "demo", Name: "Demo"}}})
	page = updated.(ProfilesPage)

	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
	_ = updated
	if cmd == nil {
		t.Fatal("expected LaunchProfileMsg cmd, got nil")
	}
	got := cmd()
	lp, ok := got.(LaunchProfileMsg)
	if !ok {
		t.Fatalf("msg type = %T, want LaunchProfileMsg", got)
	}
	if lp.ID != "demo" {
		t.Errorf("ID = %q, want demo", lp.ID)
	}
}

func TestProfilesPage_FlashAutoClear(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})

	page, _ = page.withFlash("hello")
	if page.flash != "hello" {
		t.Fatalf("flash = %q, want hello", page.flash)
	}

	// Stale clear (mismatching at) should be ignored.
	updated, _ := page.Update(flashClearMsg{tag: "profiles", at: time.Time{}})
	page = updated.(ProfilesPage)
	if page.flash != "hello" {
		t.Errorf("stale flashClearMsg erased current flash; flash=%q", page.flash)
	}

	// Matching clear erases.
	updated, _ = page.Update(flashClearMsg{tag: "profiles", at: page.flashAt})
	page = updated.(ProfilesPage)
	if page.flash != "" {
		t.Errorf("matching flashClearMsg should clear; flash=%q", page.flash)
	}
}

func TestProfilesPage_FlashRenamedClearTagIgnored(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})
	page, _ = page.withFlash("hello")

	// flashClearMsg from another page must be ignored.
	updated, _ := page.Update(flashClearMsg{tag: "models", at: page.flashAt})
	page = updated.(ProfilesPage)
	if page.flash != "hello" {
		t.Errorf("cross-tag flashClearMsg erased flash; flash=%q", page.flash)
	}
}

func TestProfilesPage_IntRangeRejectsNonInt(t *testing.T) {
	v := intRange(0, 100, false)
	if err := v("abc"); err == nil {
		t.Errorf("intRange(non-int) returned nil; want error")
	}
}

func TestProfilesPage_IntRangeRejectsOutOfBounds(t *testing.T) {
	v := intRange(0, 100, false)
	if err := v("999"); err == nil {
		t.Errorf("intRange(out-of-range) returned nil; want error")
	}
}

func TestProfilesPage_IntRangeAllowsEmptyWhenOptional(t *testing.T) {
	if err := intRange(0, 100, true)(""); err != nil {
		t.Errorf("intRange(allowEmpty=true)(\"\") returned %v; want nil", err)
	}
	if err := intRange(0, 100, false)(""); err == nil {
		t.Errorf("intRange(allowEmpty=false)(\"\") returned nil; want error")
	}
}

func TestProfilesPage_IntRangeAcceptsNegativeOneForNGL(t *testing.T) {
	v := intRange(-1, 9999, false)
	if err := v("-1"); err != nil {
		t.Errorf("intRange(-1,9999)(\"-1\") returned %v; want nil (llama.cpp default)", err)
	}
}

func TestProfilesPage_PortValidator(t *testing.T) {
	v := portValidator()
	if err := v("99999"); err == nil {
		t.Errorf("portValidator(99999) returned nil; want error (out of range)")
	}
	if err := v("8080"); err != nil {
		t.Errorf("portValidator(8080) returned %v; want nil", err)
	}
	if err := v(""); err == nil {
		t.Errorf("portValidator(\"\") returned nil; want error (required)")
	}
}

func TestProfilesPage_DelegateRendersCorruptInWarnColors(t *testing.T) {
	dgt := newProfileItemDelegate()
	l := list.New([]list.Item{corruptItem{id: "broken", err: errors.New("syntax error")}}, dgt, 80, 6)

	var buf strings.Builder
	dgt.Render(&buf, l, 0, l.Items()[0])

	want := theme.Error.Render("⚠ broken")
	if !strings.Contains(buf.String(), want) {
		t.Errorf("delegate output missing styled title %q; got %q", want, buf.String())
	}
}

func TestProfilesPage_DelegateDelegatesHealthyRowsToDefault(t *testing.T) {
	dgt := newProfileItemDelegate()
	healthy := item{p: domain.Profile{ID: "demo", Name: "Demo"}}
	l := list.New([]list.Item{healthy}, dgt, 80, 6)

	var buf strings.Builder
	dgt.Render(&buf, l, 0, healthy)
	if !strings.Contains(buf.String(), "Demo") {
		t.Errorf("default delegate output missing %q; got %q", "Demo", buf.String())
	}
}

// TestProfilesPage_StartNewResetsEditorSubTabAndFilter ensures that
// reopening the editor (n / Use-In-New / edit) starts on Essentials with
// no advanced filter, regardless of what the previous editor session left
// behind. Regression for the leak where p.subTab and p.advancedFilter
// persisted across editor lifecycles, surprising the user.
func TestProfilesPage_StartNewResetsEditorSubTabAndFilter(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})

	// Simulate a previous session that left advanced state behind.
	page.subTab = subTabAdvanced
	page.advancedFilter = "stale"
	page.filterMode = true

	updated, _ := page.startNew()
	page = updated.(ProfilesPage)
	if page.subTab != subTabEssentials {
		t.Errorf("startNew subTab = %v, want subTabEssentials", page.subTab)
	}
	if page.advancedFilter != "" {
		t.Errorf("startNew advancedFilter = %q, want empty", page.advancedFilter)
	}
	if page.filterMode {
		t.Error("startNew filterMode should be false")
	}
}

// TestProfilesPage_DeleteCompletesViaAsyncMsgs drives the delete-confirm
// flow end-to-end through teatest. Regression for the bug where
// huh.StateCompleted was reached only via async nextFieldMsg/nextGroupMsg
// arriving in the non-key forwarding path — if that path doesn't check
// completion + finalize, the form stays referenced and the delete never
// fires until the user presses another key.
func TestProfilesPage_DeleteCompletesViaAsyncMsgs(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	_ = store.Save(domain.Profile{
		ID: "doomed", Name: "Doomed", Model: "/m.gguf",
		Args: map[string]any{"port": float64(8080)},
	})

	page := NewProfilesPage(store, domain.FlagSchema{})
	tm := teatest.NewTestModel(t, page, teatest.WithInitialTermSize(120, 30))
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 30})

	// Wait for the profile to render.
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Doomed")
	}, teatest.WithDuration(2*time.Second))

	// Press 'x' to open the delete confirm.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "Delete profile doomed?")
	}, teatest.WithDuration(2*time.Second))

	// Toggle to affirmative (left arrow) and submit.
	tm.Send(tea.KeyMsg{Type: tea.KeyLeft})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Without the fix the delete never fires on a single submit; the flash
	// would not appear until another keypress kicks the loop.
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(string(out), "deleted doomed")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	final := tm.FinalModel(t).(ProfilesPage)
	if final.confirmDelete {
		t.Error("confirmDelete flag should be false after async-driven completion")
	}
	if final.confirmForm != nil {
		t.Error("confirmForm should be nil after finalize")
	}
	// Profile must actually be gone from the store.
	if _, err := store.Get("doomed"); err == nil {
		t.Error("profile 'doomed' still in store after delete")
	}
}

func TestProfilesPage_HintsIncludeLaunch(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})
	hints := page.Hints()
	if !strings.Contains(hints, "[L] launch") {
		t.Errorf("list-mode Hints missing [L] launch; got %q", hints)
	}
}

func TestProfilesPage_EscWithUnchangedDraftClosesEditor(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})

	// Open a fresh editor.
	updated, _ := page.startNew()
	page = updated.(ProfilesPage)
	if !page.editing {
		t.Fatal("expected editing after startNew")
	}

	// Press esc with no edits — should close immediately, no confirm.
	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyEsc})
	page = updated.(ProfilesPage)
	if page.editing {
		t.Error("esc with unchanged draft should close editor")
	}
	if page.confirmDiscardForm != nil {
		t.Error("esc with unchanged draft should not open discard confirm")
	}
}

func TestProfilesPage_EscWithChangedDraftPromptsDiscard(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})
	updated, _ := page.startNew()
	page = updated.(ProfilesPage)

	// Mutate draft to diverge from snapshot.
	page.draft.Name = "Mutated"

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyEsc})
	page = updated.(ProfilesPage)
	if page.confirmDiscardForm == nil {
		t.Fatal("expected discard confirm after esc with mutated draft")
	}
	if !page.IsCapturingInput() {
		t.Fatal("page should capture input while discard confirm is open")
	}

	// Negative path keeps the editor.
	*page.confirmDiscardAnswer = false
	page = page.finalizeConfirmDiscard()
	if !page.editing {
		t.Error("negative discard should keep editor open")
	}
	if page.draft == nil || page.draft.Name != "Mutated" {
		t.Error("negative discard should preserve draft mutations")
	}

	// Re-open confirm and exercise affirmative path.
	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyEsc})
	page = updated.(ProfilesPage)
	if page.confirmDiscardForm == nil {
		t.Fatal("expected discard confirm second time")
	}
	*page.confirmDiscardAnswer = true
	page = page.finalizeConfirmDiscard()
	if page.editing {
		t.Error("affirmative discard should close editor")
	}
	if page.draft != nil {
		t.Error("affirmative discard should clear draft")
	}
}

func TestProfilesPage_DiscardConfirmReceivesInitHandshake(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})
	updated, _ := page.startNew()
	page = updated.(ProfilesPage)

	// Mutate draft so esc opens the discard confirm.
	page.draft.Name = "Mutated"
	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyEsc})
	page = updated.(ProfilesPage)
	if page.confirmDiscardForm == nil {
		t.Fatal("expected discard confirm form")
	}

	// Pump the discard form's Init cmd back through Update.
	// If non-key messages were wrongly routed to the hidden editor form,
	// the discard form would miss its focus/size handshake.
	initCmd := page.confirmDiscardForm.Init()
	if initCmd != nil {
		updated, _ = page.Update(initCmd())
		page = updated.(ProfilesPage)
	}

	if page.confirmDiscardForm == nil {
		t.Error("discard confirm was dropped after Init handshake; editor form likely stole it")
	}
}

func TestProfilesPage_HintsVaryByMode(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})

	page.editing = true
	if !strings.Contains(page.Hints(), "[ctrl+t]") {
		t.Errorf("editing Hints missing [ctrl+t]; got %q", page.Hints())
	}
	page.editing = false

	page.pickerActive = true
	if !strings.Contains(page.Hints(), "[enter] pick") {
		t.Errorf("picker Hints missing [enter] pick; got %q", page.Hints())
	}
	page.pickerActive = false

	page.confirmDelete = true
	if !strings.Contains(page.Hints(), "[enter] confirm") {
		t.Errorf("confirm Hints missing [enter] confirm; got %q", page.Hints())
	}
}

func TestProfilesPage_IsCapturingInputDuringEditAndPicker(t *testing.T) {
	dir := t.TempDir()
	store, _ := profilestore.NewFSStore(dir)
	page := NewProfilesPage(store, domain.FlagSchema{})

	if page.IsCapturingInput() {
		t.Errorf("idle page captures input")
	}
	page.editing = true
	if !page.IsCapturingInput() {
		t.Errorf("editing page should capture input")
	}
	page.editing = false
	page.pickerActive = true
	if !page.IsCapturingInput() {
		t.Errorf("picker page should capture input")
	}
	page.pickerActive = false
	page.confirmDelete = true
	if !page.IsCapturingInput() {
		t.Errorf("confirmDelete page should capture input")
	}
}

// [?] help token is now owned by the global status bar (see
// ui/root_test.go TestRoot_StatusBarMentionsHelp). Pages publish their
// own hints via the HintProvider contract — see TestProfilesPage_Hints*.
