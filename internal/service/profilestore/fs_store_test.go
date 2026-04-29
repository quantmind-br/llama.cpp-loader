package profilestore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

func newStore(t *testing.T) (*FSStore, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := NewFSStore(dir)
	if err != nil {
		t.Fatalf("NewFSStore: %v", err)
	}
	return s, dir
}

func sampleProfile(id, name string) domain.Profile {
	return domain.Profile{
		ID:    id,
		Name:  name,
		Model: "/tmp/model.gguf",
		Args: map[string]any{
			"ngl":      float64(99),
			"ctx-size": float64(8192),
			"port":     float64(8080),
		},
		Launch: domain.LaunchConfig{DefaultBackground: true},
	}
}

func TestFSStore_SaveAndGet(t *testing.T) {
	s, _ := newStore(t)
	p := sampleProfile("qwen", "Qwen")

	if err := s.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Get("qwen")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Qwen" {
		t.Errorf("Name = %q, want %q", got.Name, "Qwen")
	}
	if got.SchemaVersion != domain.SchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", got.SchemaVersion, domain.SchemaVersion)
	}
	if got.Meta.CreatedAt.IsZero() || got.Meta.UpdatedAt.IsZero() {
		t.Errorf("Save did not stamp timestamps: %+v", got.Meta)
	}
}

func TestFSStore_GetNotFound(t *testing.T) {
	s, _ := newStore(t)
	_, err := s.Get("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestFSStore_GetInvalidJSON(t *testing.T) {
	s, dir := newStore(t)
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := s.Get("broken")
	if !errors.Is(err, ErrInvalidJSON) {
		t.Errorf("err = %v, want ErrInvalidJSON", err)
	}
}

func TestFSStore_List_SkipsCorrupt(t *testing.T) {
	s, dir := newStore(t)
	if err := s.Save(sampleProfile("a", "Alpha")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{}{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(sampleProfile("b", "Beta")); err != nil {
		t.Fatal(err)
	}

	got, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List len = %d, want 2 (got names: %v)", len(got), names(got))
	}
	if got[0].Name != "Alpha" || got[1].Name != "Beta" {
		t.Errorf("List unsorted: %v", names(got))
	}
}

func TestFSStore_Delete(t *testing.T) {
	s, _ := newStore(t)
	if err := s.Save(sampleProfile("x", "X")); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("x"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := s.Delete("x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Delete err = %v, want ErrNotFound", err)
	}
}

func TestFSStore_Duplicate(t *testing.T) {
	s, _ := newStore(t)
	if err := s.Save(sampleProfile("orig", "Original")); err != nil {
		t.Fatal(err)
	}

	dup, err := s.Duplicate("orig", "orig-copy")
	if err != nil {
		t.Fatalf("Duplicate: %v", err)
	}
	if dup.ID != "orig-copy" {
		t.Errorf("dup.ID = %q, want orig-copy", dup.ID)
	}
	if dup.Name != "Original (copy)" {
		t.Errorf("dup.Name = %q, want %q", dup.Name, "Original (copy)")
	}

	// Existing target -> ErrDuplicateID
	_, err = s.Duplicate("orig", "orig-copy")
	if !errors.Is(err, ErrDuplicateID) {
		t.Errorf("err = %v, want ErrDuplicateID", err)
	}

	// Missing source -> ErrNotFound
	_, err = s.Duplicate("nope", "anywhere")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestFSStore_AtomicWrite_NoLeftoverTmp(t *testing.T) {
	s, dir := newStore(t)
	if err := s.Save(sampleProfile("a", "A")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("found leftover tmp file: %s", e.Name())
		}
	}
}

func TestFSStore_MarkLastUsed(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFSStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(sampleProfile("u", "U")); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	if err := s.MarkLastUsed("u", now); err != nil {
		t.Fatalf("MarkLastUsed: %v", err)
	}
	got, _ := s.Get("u")
	if got.Meta.LastUsedAt == nil || !got.Meta.LastUsedAt.Equal(now) {
		t.Errorf("LastUsedAt = %v, want %v", got.Meta.LastUsedAt, now)
	}
}

func TestFSStore_MarkLastUsed_NotFoundIsNoop(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFSStore(dir)
	if err := s.MarkLastUsed("missing", time.Now()); err != nil {
		t.Errorf("expected nil for missing, got %v", err)
	}
}

func names(ps []domain.Profile) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Name)
	}
	return out
}
