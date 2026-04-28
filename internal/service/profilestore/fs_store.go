package profilestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

// FSStore persists profiles as one JSON file per profile under a directory.
type FSStore struct {
	dir string
}

// NewFSStore returns a Store rooted at dir. The directory is created if missing.
func NewFSStore(dir string) (*FSStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir profiles dir: %w", err)
	}
	return &FSStore{dir: dir}, nil
}

func (s *FSStore) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

func (s *FSStore) List() ([]domain.Profile, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read profiles dir: %w", err)
	}

	profiles := make([]domain.Profile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		p, err := s.Get(id)
		if err != nil {
			// Skip corrupt entries — slice 1 surfaces this in UI later via marker.
			continue
		}
		profiles = append(profiles, p)
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
	return profiles, nil
}

func (s *FSStore) Get(id string) (domain.Profile, error) {
	if id == "" {
		return domain.Profile{}, ErrInvalidID
	}
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return domain.Profile{}, ErrNotFound
		}
		return domain.Profile{}, fmt.Errorf("read profile: %w", err)
	}
	var p domain.Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return domain.Profile{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return p, nil
}

func (s *FSStore) Save(p domain.Profile) error {
	if p.ID == "" {
		return ErrInvalidID
	}
	if p.SchemaVersion == 0 {
		p.SchemaVersion = domain.SchemaVersion
	}
	now := time.Now().UTC()
	if p.Meta.CreatedAt.IsZero() {
		p.Meta.CreatedAt = now
	}
	p.Meta.UpdatedAt = now

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}

	tmp := s.path(p.ID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path(p.ID)); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func (s *FSStore) Delete(id string) error {
	if id == "" {
		return ErrInvalidID
	}
	if err := os.Remove(s.path(id)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNotFound
		}
		return fmt.Errorf("remove profile: %w", err)
	}
	return nil
}

func (s *FSStore) Duplicate(srcID, newID string) (domain.Profile, error) {
	if newID == "" {
		return domain.Profile{}, ErrInvalidID
	}
	if _, err := os.Stat(s.path(newID)); err == nil {
		return domain.Profile{}, ErrDuplicateID
	}
	src, err := s.Get(srcID)
	if err != nil {
		return domain.Profile{}, err
	}

	dup := src
	dup.ID = newID
	dup.Name = src.Name + " (copy)"
	dup.Meta = domain.ProfileMeta{} // reset timestamps; Save fills them

	if err := s.Save(dup); err != nil {
		return domain.Profile{}, err
	}
	return dup, nil
}
