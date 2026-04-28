package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_CreatesDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	cfg, err := LoadFrom(cfgPath)
	if err != nil {
		t.Fatalf("LoadFrom returned error: %v", err)
	}

	if !strings.HasSuffix(cfg.Paths.ProfilesDir, "profiles") {
		t.Errorf("default ProfilesDir should end with 'profiles', got %q", cfg.Paths.ProfilesDir)
	}
	if cfg.UI.DefaultTab != "profiles" {
		t.Errorf("default UI.DefaultTab = %q, want %q", cfg.UI.DefaultTab, "profiles")
	}
	if len(cfg.Models.SearchPaths) == 0 {
		t.Errorf("default SearchPaths must not be empty")
	}

	// File should have been written to disk.
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file was not created at %s: %v", cfgPath, err)
	}
}

func TestLoad_RoundtripExisting(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	contents := `
[paths]
profiles_dir = "/tmp/p"
log_dir      = "/tmp/l"
state_dir    = "/tmp/s"

[models]
search_paths = ["/tmp/m"]

[ui]
default_tab = "monitor"
keybindings = "default"
`
	if err := os.WriteFile(cfgPath, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(cfgPath)
	if err != nil {
		t.Fatalf("LoadFrom returned error: %v", err)
	}
	if cfg.Paths.ProfilesDir != "/tmp/p" {
		t.Errorf("ProfilesDir = %q, want /tmp/p", cfg.Paths.ProfilesDir)
	}
	if cfg.UI.DefaultTab != "monitor" {
		t.Errorf("UI.DefaultTab = %q, want monitor", cfg.UI.DefaultTab)
	}
	if len(cfg.Models.SearchPaths) != 1 || cfg.Models.SearchPaths[0] != "/tmp/m" {
		t.Errorf("SearchPaths = %v, want [/tmp/m]", cfg.Models.SearchPaths)
	}
}
