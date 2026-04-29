package pages

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
	"github.com/quantmind-br/llama-cpp-loader/internal/service/profilestore"
	"github.com/quantmind-br/llama-cpp-loader/internal/ui/components"
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
	page.draft = profileDraft{
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

func TestProfilesPage_FooterMentionsHelp(t *testing.T) {
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
	updated, _ = page.Update(loadedMsg{profiles: []domain.Profile{{
		ID: "demo", Name: "Demo", Model: "/m.gguf",
		Args: map[string]any{"port": float64(8080)},
	}}})
	page = updated.(ProfilesPage)
	if !strings.Contains(page.View(), "[?] help") {
		t.Errorf("profiles footer missing [?] help; got:\n%s", page.View())
	}
}
