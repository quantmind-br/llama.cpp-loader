// Package domain holds shared types with zero external dependencies.
package domain

import "time"

// SchemaVersion is the current Profile JSON schema version.
const SchemaVersion = 1

// Profile represents a llama-server load profile persisted on disk.
// Full field definitions are added in slice 1, task 1.
type Profile struct {
	SchemaVersion int               `json:"schemaVersion"`
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Model         string            `json:"model"`
	Args          map[string]any    `json:"args"`
	ExtraArgs     []string          `json:"extraArgs,omitempty"`
	Launch        LaunchConfig      `json:"launch"`
	Meta          ProfileMeta       `json:"meta"`
}

// LaunchConfig holds per-profile launcher defaults.
type LaunchConfig struct {
	DefaultBackground bool   `json:"defaultBackground"`
	LogFilePath       string `json:"logFilePath,omitempty"`
}

// ProfileMeta holds timestamps and bookkeeping.
type ProfileMeta struct {
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}
