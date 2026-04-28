// Package profilestore persists Profile JSON files on disk.
package profilestore

import (
	"errors"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

// Store is the interface for profile persistence.
type Store interface {
	List() ([]domain.Profile, error)
	Get(id string) (domain.Profile, error)
	Save(p domain.Profile) error
	Delete(id string) error
	Duplicate(srcID, newID string) (domain.Profile, error)
}

// Sentinel errors returned by Store implementations.
var (
	ErrNotFound     = errors.New("profile not found")
	ErrInvalidJSON  = errors.New("profile json is invalid")
	ErrDuplicateID  = errors.New("profile id already exists")
	ErrInvalidID    = errors.New("profile id is invalid")
)
